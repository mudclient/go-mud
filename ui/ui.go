package ui

import (
	"runtime"
	"strings"

	"github.com/flw-cn/printer"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

type UIConfig struct {
	AmbiguousWidth string `flag:"|auto|二义性字符宽度"`
}

type UI struct {
	printer.SimplePrinter

	config     UIConfig
	app        *tview.Application
	mainWindow *tview.TextView
	input      chan string
}

func NewUI(config UIConfig) *UI {
	return &UI{
		input: make(chan string, 10),
	}
}

func (ui *UI) Create(title string) {
	InitConsole(title)

	ui.app = tview.NewApplication()
	ui.mainWindow = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetChangedFunc(func() {
			ui.app.Draw()
		})

	ui.SetOutput(tview.ANSIWriter(ui.mainWindow))

	cmdLine := tview.NewInputField().
		SetFieldBackgroundColor(tcell.ColorBlack).
		SetLabelColor(tcell.ColorWhite).
		SetLabel("命令: ")

	cmdLine.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			cmd := cmdLine.GetText()
			if cmd != "" {
				ui.input <- cmd
				cmdLine.SetText("")
			}
			ui.mainWindow.ScrollToEnd()
		}
	})

	cmdLine.SetChangedFunc(func(text string) {
		if strings.HasPrefix(text, `"`) {
			cmdLine.SetLabel("闲聊: ").
				SetLabelColor(tcell.ColorLightCyan).
				SetFieldTextColor(tcell.ColorLightCyan)
		} else if strings.HasPrefix(text, `*`) {
			cmdLine.SetLabel("表情: ").
				SetLabelColor(tcell.ColorLime).
				SetFieldTextColor(tcell.ColorLime)
		} else if strings.HasPrefix(text, `'`) {
			cmdLine.SetLabel("说话: ").
				SetLabelColor(tcell.ColorDarkCyan).
				SetFieldTextColor(tcell.ColorDarkCyan)
		} else if strings.HasPrefix(text, `;`) {
			cmdLine.SetLabel("谣言: ").
				SetLabelColor(tcell.ColorPink).
				SetFieldTextColor(tcell.ColorPink)
		} else {
			cmdLine.SetLabel("命令: ").
				SetLabelColor(tcell.ColorWhite).
				SetFieldTextColor(tcell.ColorLightGrey)
		}
	})

	mainFrame := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(ui.mainWindow, 0, 1, false).
		AddItem(cmdLine, 1, 1, false)

	if runtime.GOOS == "windows" {
		imStatusLine := tview.NewBox()
		mainFrame.AddItem(imStatusLine, 1, 1, false)
	}

	ui.app.SetRoot(mainFrame, true).
		SetFocus(cmdLine).
		SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyCtrlC {
				if ui.mainWindow.HasFocus() {
					ui.mainWindow.ScrollToEnd()
					ui.app.SetFocus(cmdLine)
				} else {
					cmdLine.SetText("")
				}
				return nil
			} else if event.Key() == tcell.KeyCtrlB {
				ui.app.SetFocus(ui.mainWindow)
				row, _ := ui.mainWindow.GetScrollOffset()
				row -= 10
				if row < 0 {
					row = 0
				}
				ui.mainWindow.ScrollTo(row, 0)
			} else if event.Key() == tcell.KeyCtrlF {
				ui.app.SetFocus(ui.mainWindow)
				row, _ := ui.mainWindow.GetScrollOffset()
				row += 10
				ui.mainWindow.ScrollTo(row, 0)
			}
			return event
		})
}

func (ui *UI) Run() {
	defer ui.app.Stop()

	if err := ui.app.Run(); err != nil {
		panic(err)
	}
}

func (ui *UI) Stop() {
	ui.app.Stop()
	close(ui.input)
}

func (ui *UI) Input() <-chan string {
	return ui.input
}
