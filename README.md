# Vaxis

```
Know now there is no time, space between the Well & Unknowing.
Our story starts there.
Well into our future, yet far beyond our past.
In a romance between a pair of Unheavenly Creatures.

[...]

It begins with them, but ends with me. Their son, Vaxis
```

## Usage

### Minimal example

```go
package main

import "git.sr.ht/~rockorager/vaxis"

func main() {
	err := vaxis.Init(vaxis.Options{})
	if err != nil {
		panic(err)
	}
	for {
		switch msg := vaxis.PollMsg().(type) {
		case vaxis.Resize:
			win := vaxis.Window{Width: -1, Height: -1}
			vaxis.Clear(win)
			vaxis.Print(win, vaxis.Text{Content: "Hello, World!"})
			vaxis.Render()
		case vaxis.Key:
			switch msg.String() {
			case "Ctrl+c":
				vaxis.Close()
				return
			}
		}
	}
}
```

## TUI Library Roundup

Notcurses is included because it's the most advanced, most efficient,
most dank TUI library

| Feature                        | Vaxis | tcell | bubbletea | notcurses |
| ------------------------------ | :---: | :---: | :-------: | :-------: |
| RGB                            |  ✅   |  ✅   |    ✅     |    ✅     |
| Hyperlinks                     |  ✅   |  ✅   |    ❌     |    ❌     |
| Bracketed Paste                |  ✅   |  ✅   |    ❌     |    ❌     |
| Kitty Keyboard                 |  ✅   |  ❌   |    ❌     |    ✅     |
| Styled Underlines              |  ✅   |  ❌   |    ❌     |    ✅     |
| Mouse Shapes (OSC 22)          |  ✅   |  ❌   |    ❌     |    ❌     |
| System Clipboard (OSC 52)      |  ✅   |  ❌   |    ❌     |    ❌     |
| System Notifications (OSC 9)   |  ✅   |  ❌   |    ❌     |    ❌     |
| System Notifications (OSC 777) |  ✅   |  ❌   |    ❌     |    ❌     |
| Synchronized Output            |  ✅   |  ❌   |    ❌     |    ✅     |
| Images (sixel)                 |  ✅   |  ✅   |    ❌     |    ✅     |
| Images (kitty)                 |  ✅   |  ❌   |    ❌     |    ✅     |
| Images (iterm2)                |  ❌   |  ❌   |    ❌     |    ✅     |
| Video                          |  ❌   |  ❌   |    ❌     |    ✅     |
| Dank                           |  🆗   |  ❌   |    ❌     |    ✅     |
