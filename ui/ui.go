package ui

import (
	"fmt"
	"io"
	"runtime"
	"strings"
	"sync"

	"github.com/flw-cn/printer"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

type Config struct {
	AmbiguousWidth string `flag:"|auto|二义性字符宽度，可选值: auto/single/double/space"`
	HistoryLines   int    `flag:"|100000|历史记录保留行数"`
	RTTVHeight     int    `flag:"|10|历史查看模式下实时文本区域高度"`
}

type UI struct {
	printer.Printer
	sync.Mutex

	config Config
	app    *tview.Application

	ansiWriter io.Writer
	pages      *tview.Pages
	historyTV  *tview.TextView
	sepLine    *tview.TextView
	realtimeTV *tview.TextView
	cmdLine    *Readline

	buffer    []string
	unformed  bool
	scrolling bool
	offset    int

	input chan string
}

func init() {
	tcell.ColorValues[tcell.ColorYellow] = 0xC7C400
	tcell.ColorValues[tcell.ColorWhite] = 0xC7C7C7
	tcell.ColorValues[tcell.ColorGreen] = 0x00C200
}

func NewUI(config Config) *UI {
	return &UI{
		config: config,
		input:  make(chan string, 10),
	}
}

func (ui *UI) Create(title string) {
	InitConsole(title)

	ui.app = tview.NewApplication()
	ui.historyTV = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetChangedFunc(func() {
			ui.app.Draw()
		})

	ui.realtimeTV = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(false).
		SetChangedFunc(func() {
			ui.app.Draw()
		})

	ui.ansiWriter = tview.ANSIWriter(ui.realtimeTV)

	ui.cmdLine = NewReadline()
	ui.cmdLine.SetRepeat(true).
		SetAutoTrim(true).
		SetFieldBackgroundColor(tcell.ColorBlack).
		SetLabelColor(tcell.ColorWhite).
		SetLabel("命令: ")

	ui.cmdLine.SetChangedFunc(ui.cmdLineTextChanged)

	ui.sepLine = tview.NewTextView().
		SetTextAlign(tview.AlignCenter)

	ui.sepLine.SetBackgroundColor(tcell.ColorBlue)

	historyView := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(ui.historyTV, 0, 1, false).
		AddItem(ui.sepLine, 1, 1, false).
		AddItem(ui.realtimeTV, ui.config.RTTVHeight, 1, false)

	ui.pages = tview.NewPages().
		AddPage("historyView", historyView, true, false).
		AddPage("mainView", ui.realtimeTV, true, true)

	mainView := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(ui.pages, 0, 1, false).
		AddItem(ui.cmdLine, 1, 1, false)

	if runtime.GOOS == "windows" {
		imStatusLine := tview.NewBox()
		mainView.AddItem(imStatusLine, 1, 1, false)
	}

	ui.app.SetRoot(mainView, true).
		SetFocus(ui.cmdLine).
		SetInputCapture(ui.InputCapture)
}

func (ui *UI) InputCapture(event *tcell.EventKey) *tcell.EventKey {
	key := event.Key()

	if ui.isScrolling() {
		if key == tcell.KeyCtrlC {
			ui.stopScrolling()
			ui.app.SetFocus(ui.cmdLine)
		} else {
			ui.historyInputCapture(event)
		}
		return nil
	}

	if key == tcell.KeyCtrlB || key == tcell.KeyPgUp {
		ui.app.SetFocus(ui.historyTV)
		ui.startScrolling()
		ui.pageUp(10)
		return nil
	}

	if key == tcell.KeyEnter {
		cmd := ui.cmdLine.Enter()
		ui.input <- cmd
		return nil
	}

	return ui.cmdLine.InputCapture(event)
}

func (ui *UI) cmdLineTextChanged(text string) {
	if len(text) == 0 {
		return
	}

	switch text[0] {
	case '"':
		ui.cmdLine.SetLabel("闲聊: ").
			SetLabelColor(tcell.ColorLightCyan).
			SetFieldTextColor(tcell.ColorLightCyan)
	case '*':
		ui.cmdLine.SetLabel("表情: ").
			SetLabelColor(tcell.ColorLime).
			SetFieldTextColor(tcell.ColorLime)
	case '\'':
		ui.cmdLine.SetLabel("说话: ").
			SetLabelColor(tcell.ColorDarkCyan).
			SetFieldTextColor(tcell.ColorDarkCyan)
	case ';':
		ui.cmdLine.SetLabel("谣言: ").
			SetLabelColor(tcell.ColorPink).
			SetFieldTextColor(tcell.ColorPink)
	default:
		ui.cmdLine.SetLabel("命令: ").
			SetLabelColor(tcell.ColorWhite).
			SetFieldTextColor(tcell.ColorLightGrey)
	}
}

func (ui *UI) historyInputCapture(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyCtrlB, tcell.KeyPgUp:
		ui.pageUp(10)
	case tcell.KeyCtrlF, tcell.KeyPgDn:
		ui.pageDown(10)
	case tcell.KeyRune:
		switch event.Rune() {
		case 'k':
			ui.pageUp(1)
		case 'j':
			ui.pageDown(1)
		case 'g':
			ui.pageHome()
		case 'G':
			ui.pageEnd()
		}
	default:
	}

	return nil
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

func (ui *UI) startScrolling() {
	ui.Lock()
	defer ui.Unlock()

	if ui.scrolling {
		return
	}

	ui.scrolling = true
	_, _, _, height := ui.pages.GetInnerRect()
	ui.pages.SwitchToPage("historyView")
	ui.offset = len(ui.buffer) - height + 1
	ui.app.Draw()
}

func (ui *UI) stopScrolling() {
	ui.Lock()
	defer ui.Unlock()

	if !ui.scrolling {
		return
	}

	ui.pages.SwitchToPage("mainView")
	end := len(ui.buffer)
	_, _, _, height := ui.pages.GetRect()
	ui.offset = end - height
	ui.scrolling = false
	text := strings.Join(ui.buffer[ui.offset:end], "\n")
	text = tview.TranslateANSI(text + "\n")
	ui.realtimeTV.SetText(text)
}

func (ui *UI) isScrolling() bool {
	ui.Lock()
	defer ui.Unlock()

	scrolling := ui.scrolling
	return scrolling
}

func (ui *UI) pageUp(pageSize int) {
	ui.Lock()
	defer ui.Unlock()

	if !ui.scrolling {
		return
	}

	ui.offset -= pageSize
	ui.drawHistory()
}

func (ui *UI) pageDown(pageSize int) {
	ui.Lock()
	defer ui.Unlock()

	if !ui.scrolling {
		return
	}

	ui.offset += pageSize
	ui.drawHistory()
}

func (ui *UI) pageHome() {
	ui.Lock()
	defer ui.Unlock()

	if !ui.scrolling {
		return
	}

	ui.offset = 0
	ui.drawHistory()
}

func (ui *UI) pageEnd() {
	ui.Lock()
	defer ui.Unlock()

	if !ui.scrolling {
		return
	}

	ui.offset = len(ui.buffer)
	ui.drawHistory()
}

func (ui *UI) drawHistory() {
	if ui.offset < 0 {
		ui.offset = 0
	}

	_, _, _, height := ui.historyTV.GetInnerRect()
	end := ui.offset + height
	stopLine := len(ui.buffer) - ui.config.RTTVHeight
	if end > stopLine {
		end = stopLine
		ui.offset = end - height
		if ui.offset < 0 {
			ui.offset = 0
		}
	}

	hint := "PageUp/PageDown/Ctrl+B/F 向上/下翻屏, k/j 向上/下滚动, g/G 滚到头/尾, Ctrl+C 结束翻屏"
	status := fmt.Sprintf("%d~%d/%d(%d%%)", ui.offset, end, stopLine, ui.offset*100/stopLine)
	ui.sepLine.SetText(fmt.Sprintf("%s %25s", hint, status))
	text := strings.Join(ui.buffer[ui.offset:end], "\n")
	text = tview.TranslateANSI(text)
	ui.historyTV.SetText(text)
}

func (ui *UI) SetOutput(w io.Writer) {
}

func (ui *UI) Print(a ...interface{}) (n int, err error) {
	str := fmt.Sprint(a...)

	if len(str) == 0 {
		return 0, nil
	}

	lines := strings.Split(str, "\n")

	count := len(lines)
	// 根据 strings.Split 的定义，如果 str 以换行符结束，则 lines 的最后一行为空串
	unformed := len(lines[count-1]) > 0
	if !unformed {
		count--
	}

	i := 0
	ui.Lock()
	defer ui.Unlock()
	l := len(ui.buffer)
	if ui.unformed {
		ui.buffer[l-1] += lines[0]
		i++
	}

	for ; i < count; i++ {
		ui.buffer = append(ui.buffer, lines[i])
	}

	if len(ui.buffer) > ui.config.HistoryLines {
		offset := len(ui.buffer) - ui.config.HistoryLines
		ui.buffer = ui.buffer[offset:len(ui.buffer)]
	}

	ui.unformed = unformed

	fmt.Fprint(ui.ansiWriter, str)

	return len(str), nil
}

func (ui *UI) Println(a ...interface{}) (n int, err error) {
	str := fmt.Sprintln(a...)
	return ui.Print(str)
}

func (ui *UI) Printf(format string, a ...interface{}) (n int, err error) {
	str := fmt.Sprintf(format, a...)
	return ui.Print(str)
}
