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
	messages        []ChatMessage
	inputBuffer     string
	scrollOffset    int
	roster          []string
	activeTyping    = make(map[string]time.Time)
	activeTransfers = make(map[string]int)
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
				// Handle Enter Press
				if inputBuffer != "" {
					if strings.HasPrefix(inputBuffer, "/send ") {
						parts := strings.SplitN(inputBuffer, " ", 3)
						if len(parts) == 3 {
							SendFileToPeer(parts[1], parts[2])
						} else {
							AddSystemMessage("Usage: /send <username> <filepath>")
						}
					} else if strings.HasPrefix(inputBuffer, "/msg ") {
						parts := strings.SplitN(inputBuffer, " ", 3)
						if len(parts) == 3 {
							SendPrivateMessage(parts[1], parts[2])
						} else {
							AddSystemMessage("Usage: /msg <username> <message>")
						}
					} else {
						// Normal chat message
						onSend(inputBuffer)
						AddLocalMessage(username, inputBuffer)
					}
					inputBuffer = ""
					scrollOffset = 0 // Snap to bottom on send
				}
			} else if ev.Key == termbox.KeyBackspace || ev.Key == termbox.KeyBackspace2 {
				// Handle Backspace
				if len(inputBuffer) > 0 {
					inputBuffer = inputBuffer[:len(inputBuffer)-1]
					triggerTypingSafe(onType)
				}
			} else if ev.Key == termbox.KeySpace {
				// Handle Spacebar
				inputBuffer += " "
				triggerTypingSafe(onType)
			} else if ev.Key == termbox.KeyPgup {
				// Handle Scroll Up
				scrollOffset += h / 2 
			} else if ev.Key == termbox.KeyPgdn {
				// Handle Scroll Down
				scrollOffset -= h / 2
				if scrollOffset < 0 {
					scrollOffset = 0
				}
			} else if ev.Ch != 0 {
				// Handle Regular Typing
				inputBuffer += string(ev.Ch)
				triggerTypingSafe(onType)
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

	// --- 1. PRE-CALCULATE ALL WRAPPED LINES ---
	var displayLines []renderLine
	for _, msg := range messages {
		timeStr := msg.Timestamp.Format("15:04")

		if msg.IsSystem {
			formatted := fmt.Sprintf("%s [*] %s", timeStr, msg.Text)
			wrapped := wrapText(formatted, chatWidth-4)
			for _, line := range wrapped {
				displayLines = append(displayLines, renderLine{
					text: line, isSys: true, offset: 2, sysCol: termbox.ColorYellow,
				})
			}
		} else {
			userStr := "<" + msg.Username + ">"
			textOffset := 8 + len(msg.Username) + 3
			maxWidth := chatWidth - textOffset - 2

			wrapped := wrapText(msg.Text, maxWidth)

			for i, line := range wrapped {
				if i == 0 {
					// First line gets the timestamp and username
					displayLines = append(displayLines, renderLine{
						timeStr: timeStr, userStr: userStr, userCol: getUserColor(msg.Username),
						text: line, isSys: false, offset: textOffset,
					})
				} else {
					// Wrapped lines are blank on the left for a clean indent
					displayLines = append(displayLines, renderLine{
						timeStr: "", userStr: "", userCol: termbox.ColorDefault,
						text: line, isSys: false, offset: textOffset,
					})
				}
			}
		}
	}

	// --- 2. CALCULATE SCROLLING BASED ON TOTAL LINES ---
	msgAreaHeight := h - 4
	totalLines := len(displayLines)

	if scrollOffset > totalLines-msgAreaHeight {
		scrollOffset = totalLines - msgAreaHeight
	}
	if scrollOffset < 0 {
		scrollOffset = 0
	}

	start := 0
	if totalLines > msgAreaHeight {
		start = totalLines - msgAreaHeight - scrollOffset
	}
	if start < 0 {
		start = 0
	}

	// --- 3. DRAW THE VISIBLE LINES ---
	drawY := 1
	for i := start; i < totalLines && drawY <= msgAreaHeight; i++ {
		dl := displayLines[i]

		if dl.isSys {
			drawText(dl.offset, drawY, dl.text, dl.sysCol)
		} else {
			if dl.timeStr != "" {
				drawText(2, drawY, dl.timeStr, termbox.ColorDarkGray)
				drawText(8, drawY, dl.userStr, dl.userCol)
			}
			drawText(dl.offset, drawY, dl.text, termbox.ColorWhite)
		}
		drawY++
	}

	// Draw Typing Indicator
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
	drawText(x+1, y+1, strings.Repeat("─", width-2), termbox.ColorCyan)

	for i, user := range roster {
		if i >= height-10 {
			break
		}
		drawText(x+2, y+2+i, user, getUserColor(user))
	}

	// Draw Active Transfers at the bottom of the roster panel
	if len(activeTransfers) > 0 {
		transferStartY := y + height - 2 - (len(activeTransfers) * 2)
		drawText(x+1, transferStartY-1, "TRANSFERS", termbox.ColorYellow|termbox.AttrBold)
		drawText(x+1, transferStartY, strings.Repeat("─", width-2), termbox.ColorCyan)

		i := 0
		for name, pct := range activeTransfers {
			// Ensure filename fits in panel
			display := name
			if len(display) > width-4 {
				display = display[:width-7] + "..."
			}

			// Draw name and percentage
			drawText(x+2, transferStartY+1+(i*2), fmt.Sprintf("%s (%d%%)", display, pct), termbox.ColorWhite)

			// Draw ASCII Progress Bar
			barWidth := width - 4
			filled := int((float32(pct) / 100.0) * float32(barWidth))
			bar := strings.Repeat("█", filled) + strings.Repeat("░", barWidth-filled)
			drawText(x+2, transferStartY+2+(i*2), bar, termbox.ColorCyan)
			i++
		}
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

func UpdateTransferUI(filename string, progress int, isUpload bool) {
	direction := "DL"
	if isUpload {
		direction = "UL"
	}
	key := fmt.Sprintf("[%s] %s", direction, filename)

	if progress >= 100 {
		delete(activeTransfers, key)
	} else {
		activeTransfers[key] = progress
	}
	redrawAll("")
}


type renderLine struct {
	timeStr string
	userStr string
	userCol termbox.Attribute
	text    string
	isSys   bool
	offset  int
	sysCol  termbox.Attribute
}

func wrapText(text string, maxWidth int) []string {
	if maxWidth <= 0 {
		return []string{text}
	}
	var lines []string
	words := strings.Split(text, " ")

	currentLine := ""
	for _, word := range words {
		// Handle massive single words (like long URLs) that exceed maxWidth
		for len(word) > maxWidth {
			if currentLine != "" {
				lines = append(lines, currentLine)
				currentLine = ""
			}
			lines = append(lines, word[:maxWidth])
			word = word[maxWidth:]
		}

		if currentLine == "" {
			currentLine = word
		} else if len(currentLine)+1+len(word) <= maxWidth {
			currentLine += " " + word
		} else {
			lines = append(lines, currentLine)
			currentLine = word
		}
	}
	if currentLine != "" {
		lines = append(lines, currentLine)
	}
	return lines
}

// Helper to prevent typing indicators from leaking during private commands
func triggerTypingSafe(onType func()) {
	// Only trigger the network broadcast if we are NOT typing a command
	if !strings.HasPrefix(inputBuffer, "/") {
		onType()
	}
}