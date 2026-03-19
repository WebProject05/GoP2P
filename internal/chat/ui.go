// internal/chat/ui.go
package chat

import (
	"fmt"
	"github.com/nsf/termbox-go"
)

var (
	messages    []string
	inputBuffer string
)

func StartUI(username string, onSend func(string)) {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	termbox.SetInputMode(termbox.InputEsc)
	redrawAll(username)

	for {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			if ev.Key == termbox.KeyEsc || ev.Key == termbox.KeyCtrlC {
				return
			} else if ev.Key == termbox.KeyEnter {
				if inputBuffer != "" {
					onSend(inputBuffer)
					AddLocalMessage(username, inputBuffer)
					inputBuffer = ""
				}
			} else if ev.Key == termbox.KeyBackspace || ev.Key == termbox.KeyBackspace2 {
				if len(inputBuffer) > 0 {
					inputBuffer = inputBuffer[:len(inputBuffer)-1]
				}
			} else if ev.Key == termbox.KeySpace {
				inputBuffer += " "
			} else if ev.Ch != 0 {
				inputBuffer += string(ev.Ch)
			}
		case termbox.EventResize:
			// Safely catch window resizing so it redraws instead of crashing
		case termbox.EventError:
			return
		}
		redrawAll(username)
	}
}

func AddLocalMessage(user, msg string) {
	messages = append(messages, fmt.Sprintf("%s: %s", user, msg))
	redrawAll(user)
}

func AddRemoteMessage(user, msg string) {
	messages = append(messages, fmt.Sprintf("%s: %s", user, msg))
	redrawAll("")
}

func AddSystemMessage(msg string) {
	messages = append(messages, fmt.Sprintf("[System] %s", msg))
	redrawAll("")
}

func redrawAll(username string) {
	w, h := termbox.Size()
	
	// CRITICAL FIX 1: Prevent Windows panic if the terminal is too small
	if w < 10 || h < 10 {
		return 
	}

	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	// 1. Draw the Outline Box
	drawBorders(w, h)

	// 2. Draw Messages (Top Part)
	msgAreaHeight := h - 4
	start := 0
	if len(messages) > msgAreaHeight {
		start = len(messages) - msgAreaHeight
	}
	
	for i, msg := range messages[start:] {
		if len(msg) > w-4 {
			msg = msg[:w-4] + "..."
		}
		
		if len(msg) > 8 && msg[:8] == "[System]" {
			drawText(2, i+1, msg, termbox.ColorYellow)
		} else {
			drawText(2, i+1, msg, termbox.ColorWhite)
		}
	}

	// 3. Draw Input Prompt (Bottom Part)
	var prompt string
	if username != "" {
		prompt = fmt.Sprintf("[%s] > %s", username, inputBuffer)
	} else {
		prompt = fmt.Sprintf("> %s", inputBuffer)
	}
	
	if len(prompt) > w-4 {
		prompt = prompt[len(prompt)-(w-4):]
	}
	
	drawText(2, h-2, prompt, termbox.ColorGreen)
	
	// CRITICAL FIX 2: Safely clamp the cursor position
	cursorX := 2 + len(prompt)
	if cursorX >= w {
		cursorX = w - 1
	}
	cursorY := h - 2
	if cursorY >= h {
		cursorY = h - 1
	}
	
	termbox.SetCursor(cursorX, cursorY)
	termbox.Flush()
}

func drawBorders(w, h int) {
	color := termbox.ColorCyan

	termbox.SetCell(0, 0, '┌', color, termbox.ColorDefault)
	termbox.SetCell(w-1, 0, '┐', color, termbox.ColorDefault)
	termbox.SetCell(0, h-1, '└', color, termbox.ColorDefault)
	termbox.SetCell(w-1, h-1, '┘', color, termbox.ColorDefault)

	termbox.SetCell(0, h-3, '├', color, termbox.ColorDefault)
	termbox.SetCell(w-1, h-3, '┤', color, termbox.ColorDefault)

	for x := 1; x < w-1; x++ {
		termbox.SetCell(x, 0, '─', color, termbox.ColorDefault)
		termbox.SetCell(x, h-1, '─', color, termbox.ColorDefault)
		termbox.SetCell(x, h-3, '─', color, termbox.ColorDefault)
	}

	for y := 1; y < h-1; y++ {
		if y == h-3 {
			continue
		}
		termbox.SetCell(0, y, '│', color, termbox.ColorDefault)
		termbox.SetCell(w-1, y, '│', color, termbox.ColorDefault)
	}
}

func drawText(x, y int, text string, fg termbox.Attribute) {
	for i, ch := range text {
		// Extra safety check to keep characters inside the terminal
		if x+i >= 0 { 
			termbox.SetCell(x+i, y, ch, fg, termbox.ColorDefault)
		}
	}
}