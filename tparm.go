// Copyright 2022 The TCell Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use file except in compliance with the License.
// You may obtain a copy of the license at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rtk

import (
	"bytes"
	"fmt"
	"strconv"
)

// tparm takes a terminfo parameterized string, such as setaf or cup, and
// evaluates the string, and returns the result with the parameter
// applied.
func tparm(s string, p ...interface{}) string {
	var stk stack
	var a string
	var ai, bi int
	var dvars [26]string
	var params [9]interface{}
	pb := &paramsBuffer{}

	pb.Start(s)

	// make sure we always have 9 parameters -- makes it easier
	// later to skip checks
	for i := 0; i < len(params) && i < len(p); i++ {
		params[i] = p[i]
	}

	const (
		emit = iota
		toEnd
		toElse
	)

	skip := emit

	for {

		ch, err := pb.NextCh()
		if err != nil {
			break
		}

		if ch != '%' {
			if skip == emit {
				pb.PutCh(ch)
			}
			continue
		}

		ch, err = pb.NextCh()
		if err != nil {
			// XXX Error
			break
		}
		if skip == toEnd {
			if ch == ';' {
				skip = emit
			}
			continue
		} else if skip == toElse {
			if ch == 'e' || ch == ';' {
				skip = emit
			}
			continue
		}

		switch ch {
		case '%': // quoted %
			pb.PutCh(ch)

		case 'i': // increment both parameters (ANSI cup support)
			if i, ok := params[0].(int); ok {
				params[0] = i + 1
			}
			if i, ok := params[1].(int); ok {
				params[1] = i + 1
			}

		case 's':
			// NB: 's', 'c', and 'd' below are special cased for
			// efficiency.  They could be handled by the richer
			// format support below, less efficiently.
			a, stk = stk.PopString()
			pb.PutString(a)

		case 'c':
			// Integer as special character.
			ai, stk = stk.PopInt()
			pb.PutCh(byte(ai))

		case 'd':
			ai, stk = stk.PopInt()
			pb.PutString(strconv.Itoa(ai))

		case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'x', 'X', 'o', ':':
			// This is pretty suboptimal, but this is rarely used.
			// None of the mainstream terminals use any of this,
			// and it would surprise me if this code is ever
			// executed outside test cases.
			f := "%"
			if ch == ':' {
				ch, _ = pb.NextCh()
			}
			f += string(ch)
			for ch == '+' || ch == '-' || ch == '#' || ch == ' ' {
				ch, _ = pb.NextCh()
				f += string(ch)
			}
			for (ch >= '0' && ch <= '9') || ch == '.' {
				ch, _ = pb.NextCh()
				f += string(ch)
			}
			switch ch {
			case 'd', 'x', 'X', 'o':
				ai, stk = stk.PopInt()
				pb.PutString(fmt.Sprintf(f, ai))
			case 's':
				a, stk = stk.PopString()
				pb.PutString(fmt.Sprintf(f, a))
			case 'c':
				ai, stk = stk.PopInt()
				pb.PutString(fmt.Sprintf(f, ai))
			}

		case 'p': // push parameter
			ch, _ = pb.NextCh()
			ai = int(ch - '1')
			if ai >= 0 && ai < len(params) {
				stk = stk.Push(params[ai])
			} else {
				stk = stk.Push(0)
			}

		case 'P': // pop & store variable
			ch, _ = pb.NextCh()
			if ch >= 'A' && ch <= 'Z' {
				svars[int(ch-'A')], stk = stk.PopString()
			} else if ch >= 'a' && ch <= 'z' {
				dvars[int(ch-'a')], stk = stk.PopString()
			}

		case 'g': // recall & push variable
			ch, _ = pb.NextCh()
			if ch >= 'A' && ch <= 'Z' {
				stk = stk.Push(svars[int(ch-'A')])
			} else if ch >= 'a' && ch <= 'z' {
				stk = stk.Push(dvars[int(ch-'a')])
			}

		case '\'': // push(char) - the integer value of it
			ch, _ = pb.NextCh()
			_, _ = pb.NextCh() // must be ' but we don't check
			stk = stk.Push(int(ch))

		case '{': // push(int)
			ai = 0
			ch, _ = pb.NextCh()
			for ch >= '0' && ch <= '9' {
				ai *= 10
				ai += int(ch - '0')
				ch, _ = pb.NextCh()
			}
			// ch must be '}' but no verification
			stk = stk.Push(ai)

		case 'l': // push(strlen(pop))
			a, stk = stk.PopString()
			stk = stk.Push(len(a))

		case '+':
			bi, stk = stk.PopInt()
			ai, stk = stk.PopInt()
			stk = stk.Push(ai + bi)

		case '-':
			bi, stk = stk.PopInt()
			ai, stk = stk.PopInt()
			stk = stk.Push(ai - bi)

		case '*':
			bi, stk = stk.PopInt()
			ai, stk = stk.PopInt()
			stk = stk.Push(ai * bi)

		case '/':
			bi, stk = stk.PopInt()
			ai, stk = stk.PopInt()
			if bi != 0 {
				stk = stk.Push(ai / bi)
			} else {
				stk = stk.Push(0)
			}

		case 'm': // push(pop mod pop)
			bi, stk = stk.PopInt()
			ai, stk = stk.PopInt()
			if bi != 0 {
				stk = stk.Push(ai % bi)
			} else {
				stk = stk.Push(0)
			}

		case '&': // AND
			bi, stk = stk.PopInt()
			ai, stk = stk.PopInt()
			stk = stk.Push(ai & bi)

		case '|': // OR
			bi, stk = stk.PopInt()
			ai, stk = stk.PopInt()
			stk = stk.Push(ai | bi)

		case '^': // XOR
			bi, stk = stk.PopInt()
			ai, stk = stk.PopInt()
			stk = stk.Push(ai ^ bi)

		case '~': // bit complement
			ai, stk = stk.PopInt()
			stk = stk.Push(ai ^ -1)

		case '!': // logical NOT
			ai, stk = stk.PopInt()
			stk = stk.Push(ai == 0)

		case '=': // numeric compare
			bi, stk = stk.PopInt()
			ai, stk = stk.PopInt()
			stk = stk.Push(ai == bi)

		case '>': // greater than, numeric
			bi, stk = stk.PopInt()
			ai, stk = stk.PopInt()
			stk = stk.Push(ai > bi)

		case '<': // less than, numeric
			bi, stk = stk.PopInt()
			ai, stk = stk.PopInt()
			stk = stk.Push(ai < bi)

		case '?': // start conditional

		case ';':
			skip = emit

		case 't':
			ai, stk = stk.PopInt()
			if ai == 0 {
				skip = toElse
			}

		case 'e':
			skip = toEnd

		default:
			pb.PutString("%" + string(ch))
		}
	}

	return pb.End()
}

type stack []interface{}

func (st stack) Push(v interface{}) stack {
	if b, ok := v.(bool); ok {
		if b {
			return append(st, 1)
		} else {
			return append(st, 0)
		}
	}
	return append(st, v)
}

func (st stack) PopString() (string, stack) {
	if len(st) > 0 {
		e := st[len(st)-1]
		var s string
		switch v := e.(type) {
		case int:
			s = strconv.Itoa(v)
		case string:
			s = v
		}
		return s, st[:len(st)-1]
	}
	return "", st
}

func (st stack) PopInt() (int, stack) {
	if len(st) > 0 {
		e := st[len(st)-1]
		var i int
		switch v := e.(type) {
		case int:
			i = v
		case string:
			i, _ = strconv.Atoi(v)
		}
		return i, st[:len(st)-1]
	}
	return 0, st
}

type paramsBuffer struct {
	out bytes.Buffer
	buf bytes.Buffer
}

// Start initializes the params buffer with the initial string data.
// It also locks the paramsBuffer.  The caller must call End() when
// finished.
func (pb *paramsBuffer) Start(s string) {
	pb.out.Reset()
	pb.buf.Reset()
	pb.buf.WriteString(s)
}

// End returns the final output from TParam, but it also releases the lock.
func (pb *paramsBuffer) End() string {
	s := pb.out.String()
	return s
}

// NextCh returns the next input character to the expander.
func (pb *paramsBuffer) NextCh() (byte, error) {
	return pb.buf.ReadByte()
}

// PutCh "emits" (rather schedules for output) a single byte character.
func (pb *paramsBuffer) PutCh(ch byte) {
	pb.out.WriteByte(ch)
}

// PutString schedules a string for output.
func (pb *paramsBuffer) PutString(s string) {
	pb.out.WriteString(s)
}

// static vars
var svars [26]string
