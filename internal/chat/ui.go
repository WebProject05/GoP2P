// internal/chat/ui.go
package chat

import (
	"fmt"
	"strings"
	"time"

	"github.com/nsf/termbox-go"
)

type ChatMessage struct {
	Timestamp time.Time
	Username  string
	Text      string
	IsSystem  bool
}

var (
	messages     []ChatMessage
	inputBuffer  string
	scrollOffset int
	roster       []string
	activeTyping = make(map[string]time.Time)
)

func StartUI(username string, onSend func(string), onType func()) {
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
					scrollOffset = 0 // Snap to bottom on send
				}
			} else if ev.Key == termbox.KeyBackspace || ev.Key == termbox.KeyBackspace2 {
				if len(inputBuffer) > 0 {
					inputBuffer = inputBuffer[:len(inputBuffer)-1]
					onType()
				}
			} else if ev.Key == termbox.KeySpace {
				inputBuffer += " "
				onType()
			} else if ev.Key == termbox.KeyPgup {
				scrollOffset += h / 2 // Scroll up half a screen
			} else if ev.Key == termbox.KeyPgdn {
				scrollOffset -= h / 2
				if scrollOffset < 0 {
					scrollOffset = 0
				}
			} else if ev.Ch != 0 {
				inputBuffer += string(ev.Ch)
				onType()
			}
		case termbox.EventResize:
			// Handle resize safely
		case termbox.EventError:
			return
		}
		redrawAll(username)
	}
}

// Data Updaters
func AddLocalMessage(user, msg string) {
	messages = append(messages, ChatMessage{time.Now(), user, msg, false})
	redrawAll(user)
}

func AddRemoteMessage(user, msg string) {
	messages = append(messages, ChatMessage{time.Now(), user, msg, false})
	redrawAll("")
}

func AddSystemMessage(msg string) {
	messages = append(messages, ChatMessage{time.Now(), "System", msg, true})
	redrawAll("")
}

func UpdateRoster(users []string) {
	roster = users
	redrawAll("")
}

func SetTyping(user string) {
	activeTyping[user] = time.Now()
	redrawAll("")
}

// Drawing Logic
var w, h int

func redrawAll(username string) {
	w, h = termbox.Size()
	if w < 30 || h < 10 {
		return
	}

	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	rosterWidth := 20
	chatWidth := w - rosterWidth - 1

	drawBorders(w, h, chatWidth)
	drawRoster(chatWidth+1, 1, rosterWidth-1, h-2)

	// Draw Messages with Scrollback
	msgAreaHeight := h - 4
	totalMsgs := len(messages)
	
	// Enforce scroll limits
	if scrollOffset > totalMsgs-msgAreaHeight {
		scrollOffset = totalMsgs - msgAreaHeight
	}
	if scrollOffset < 0 {
		scrollOffset = 0
	}

	start := 0
	if totalMsgs > msgAreaHeight {
		start = totalMsgs - msgAreaHeight - scrollOffset
	}
	if start < 0 {
		start = 0
	}

	drawY := 1
	for i := start; i < totalMsgs && drawY <= msgAreaHeight; i++ {
		msg := messages[i]
		timeStr := msg.Timestamp.Format("15:04") // HH:MM 24hr format
		
		if msg.IsSystem {
			formatted := fmt.Sprintf("%s [*] %s", timeStr, msg.Text)
			if len(formatted) > chatWidth-4 {
				formatted = formatted[:chatWidth-4] + "..."
			}
			drawText(2, drawY, formatted, termbox.ColorYellow)
		} else {
			// Draw time
			drawText(2, drawY, timeStr, termbox.ColorDarkGray)
			
			// Draw colored username
			userColor := getUserColor(msg.Username)
			drawText(8, drawY, "<"+msg.Username+">", userColor)
			
			// Draw message text
			textOffset := 8 + len(msg.Username) + 3
			textContent := msg.Text
			if len(textContent) > chatWidth-textOffset-2 {
				textContent = textContent[:chatWidth-textOffset-2] + "..."
			}
			drawText(textOffset, drawY, textContent, termbox.ColorWhite)
		}
		drawY++
	}

	// Draw Typing Indicator (Just above input line)
	typingStr := getTypingString()
	if typingStr != "" {
		drawText(2, h-3, typingStr, termbox.ColorCyan)
	}

	// Draw Input Box
	prompt := fmt.Sprintf("[%s] > %s", username, inputBuffer)
	if len(prompt) > chatWidth-4 {
		prompt = prompt[len(prompt)-(chatWidth-4):]
	}
	drawText(2, h-2, prompt, termbox.ColorGreen)

	cursorX := 2 + len(prompt)
	if cursorX >= chatWidth {
		cursorX = chatWidth - 1
	}
	termbox.SetCursor(cursorX, h-2)
	termbox.Flush()
}

func drawRoster(x, y, width, height int) {
	drawText(x+1, y, "ONLINE PEERS", termbox.ColorGreen|termbox.AttrBold)
	drawText(x+1, y+1, strings.Repeat("-", width-2), termbox.ColorCyan)
	
	for i, user := range roster {
		if i >= height-3 {
			break
		}
		drawText(x+2, y+2+i, user, getUserColor(user))
	}
}

func drawBorders(w, h, splitX int) {
	color := termbox.ColorCyan

	// Outer Box
	termbox.SetCell(0, 0, '┌', color, termbox.ColorDefault)
	termbox.SetCell(w-1, 0, '┐', color, termbox.ColorDefault)
	termbox.SetCell(0, h-1, '└', color, termbox.ColorDefault)
	termbox.SetCell(w-1, h-1, '┘', color, termbox.ColorDefault)

	// Intersections
	termbox.SetCell(splitX, 0, '┬', color, termbox.ColorDefault)
	termbox.SetCell(splitX, h-1, '┴', color, termbox.ColorDefault)
	termbox.SetCell(0, h-4, '├', color, termbox.ColorDefault)
	termbox.SetCell(splitX, h-4, '┤', color, termbox.ColorDefault)

	// Horizontal lines
	for x := 1; x < w-1; x++ {
		termbox.SetCell(x, 0, '─', color, termbox.ColorDefault)
		termbox.SetCell(x, h-1, '─', color, termbox.ColorDefault)
		if x < splitX {
			termbox.SetCell(x, h-4, '─', color, termbox.ColorDefault)
		}
	}

	// Vertical lines
	for y := 1; y < h-1; y++ {
		termbox.SetCell(0, y, '│', color, termbox.ColorDefault)
		termbox.SetCell(w-1, y, '│', color, termbox.ColorDefault)
		termbox.SetCell(splitX, y, '│', color, termbox.ColorDefault)
	}
}

func drawText(x, y int, text string, fg termbox.Attribute) {
	for i, ch := range text {
		if x+i >= 0 {
			termbox.SetCell(x+i, y, ch, fg, termbox.ColorDefault)
		}
	}
}

// Helper: Consistently hashes a username to a specific color
func getUserColor(name string) termbox.Attribute {
	colors := []termbox.Attribute{
		termbox.ColorRed, termbox.ColorGreen, termbox.ColorYellow,
		termbox.ColorBlue, termbox.ColorMagenta, termbox.ColorCyan,
	}
	sum := 0
	for _, c := range name {
		sum += int(c)
	}
	return colors[sum%len(colors)]
}

// Helper: Formats "Alice, Bob are typing..."
func getTypingString() string {
	var typers []string
	now := time.Now()
	for user, t := range activeTyping {
		if now.Sub(t) < 3*time.Second {
			typers = append(typers, user)
		} else {
			delete(activeTyping, user)
		}
	}
	
	if len(typers) == 0 {
		return ""
	} else if len(typers) == 1 {
		return typers[0] + " is typing..."
	}
	return strings.Join(typers, ", ") + " are typing..."
}

func ClearTyping(user string) {
	delete(activeTyping, user)
	redrawAll("")
}