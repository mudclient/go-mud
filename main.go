package main

import (
	"bufio"
	"fmt"
	"io"
	"net"

	"github.com/axgle/mahonia"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"golang.org/x/text/width"
)

var (
	decoder mahonia.Decoder
	encoder mahonia.Encoder
)

func init() {
	decoder = mahonia.NewDecoder("GB18030")
	encoder = mahonia.NewEncoder("GB18030")
}

func main() {
	app := tview.NewApplication()
	mainWindow := tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetChangedFunc(func() {
			app.Draw()
		})

	mudInput, output := mudServer()
	go func() {
		w := tview.ANSIWriter(mainWindow)
		for input := range mudInput {
			fmt.Fprintln(w, input)
		}
	}()

	cmdLine := tview.NewInputField().
		SetFieldBackgroundColor(tcell.ColorBlack).
		SetLabel("MUD: ")

	cmdLine.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			cmd := cmdLine.GetText()
			if cmd == "exit" || cmd == "quit" {
				app.Stop()
			}
			cmdLine.SetText("")
			fmt.Fprintln(mainWindow, cmd)
			mainWindow.ScrollToEnd()
			cmd = UTF8_TO_GBK(cmd)
			fmt.Fprintln(output, cmd)
		}
	})

	mainFrame := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(mainWindow, 0, 1, false).
		AddItem(cmdLine, 1, 1, false)

	app.SetRoot(mainFrame, true).
		SetFocus(cmdLine).
		SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyCtrlC {
				cmdLine.SetText("")
				return nil
			} else if event.Key() == tcell.KeyCtrlB {
				row, _ := mainWindow.GetScrollOffset()
				row -= 50
				if row < 0 {
					row = 0
				}
				mainWindow.ScrollTo(row, 0)
			}
			return event
		})

	if err := app.Run(); err != nil {
		panic(err)
	}
}

func mudServer() (input <-chan string, output io.Writer) {
	serverAddress := "mud.pkuxkx.net:8080"

	fmt.Printf("连接到服务器 %s...", serverAddress)
	conn, _ := net.Dial("tcp", serverAddress)
	rd := bufio.NewReader(conn)
	fmt.Println(" 连接成功。")

	mudInput := make(chan string, 1024)
	go func(ch chan<- string) {
		for {
			lineBuf, _, err := rd.ReadLine()
			if err != nil {
				break
			}
			lineStr := GBK_TO_UTF8(string(lineBuf))
			newLineStr := ""
			for _, c := range lineStr {
				newLineStr += string(c)
				switch c {
				case '┌', '┬', '├', '┼', '└', '┴', '─',
					'╓', '╥', '╟', '╫', '╙', '╨',
					'╭', '╰':
					newLineStr += "─"
				case '╔', '╦', '╠', '╬', '╚', '╩', '═',
					'╒', '╤', '╞', '╪', '╘', '╧':
					newLineStr += "═"
				case '█', '▇', '▆', '▅', '▄', '▃', '▂', '▁', '▀':
					newLineStr += string(c)
				default:
					p := width.LookupRune(c)
					if p.Kind() == width.EastAsianAmbiguous {
						newLineStr += " "
					}
				}
			}

			ch <- newLineStr
		}
		close(ch)
	}(mudInput)

	return mudInput, conn
}

func UTF8_TO_GBK(input string) (output string) {
	output = encoder.ConvertString(input)
	return
}

func GBK_TO_UTF8(input string) (output string) {
	output = decoder.ConvertString(input)
	return
}
