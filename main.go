package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	lua "github.com/yuin/gopher-lua"

	"github.com/axgle/mahonia"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"golang.org/x/text/width"
)

var (
	decoder mahonia.Decoder
	encoder mahonia.Encoder

	ansiRe     *regexp.Regexp
	app        *tview.Application
	luaRobot   *LuaRobot
	mainWindow *tview.TextView
)

func init() {
	decoder = mahonia.NewDecoder("GB18030")
	encoder = mahonia.NewEncoder("GB18030")

	ansiRe = regexp.MustCompile("\x1b" + `\[\d*(?:;\d*(?:;\d*)?)?(?:A|D|K|m)`)
}

type LuaRobot struct {
	lstate *lua.LState
	timer  sync.Map
	mud    io.Writer

	onReceive lua.P
	onSend    lua.P
}

type Timer struct {
	id       string
	code     string
	delay    int
	maxTimes int
	times    int
	quit     chan<- bool
}

func (t *Timer) Emit(l *LuaRobot) {
	err := l.lstate.DoString(`call_timer_actions("` + t.id + `")`)
	if err != nil {
		l.Logf("Lua Error: %s", err)
	}
}

func initLua() {
	luaRobot = &LuaRobot{}
	luaRobot.Reload()
}

func (l *LuaRobot) Reload() {
	if luaRobot.lstate != nil {
		luaRobot.lstate.Close()
		l.Logf("Lua 环境已关闭。")
	}

	l.Logf("初始化 Lua 环境...")

	luaRobot.lstate = lua.NewState()

	luaRobot.lstate.SetGlobal("RegEx", luaRobot.lstate.NewFunction(luaRobot.RegEx))
	luaRobot.lstate.SetGlobal("Print", luaRobot.lstate.NewFunction(luaRobot.Print))
	luaRobot.lstate.SetGlobal("Show", luaRobot.lstate.NewFunction(luaRobot.Show))
	luaRobot.lstate.SetGlobal("Run", luaRobot.lstate.NewFunction(luaRobot.Run))
	luaRobot.lstate.SetGlobal("Echo", luaRobot.lstate.NewFunction(luaRobot.Echo))
	luaRobot.lstate.SetGlobal("Send", luaRobot.lstate.NewFunction(luaRobot.Send))
	luaRobot.lstate.SetGlobal("AddTimer", luaRobot.lstate.NewFunction(luaRobot.AddTimer))
	luaRobot.lstate.SetGlobal("AddMSTimer", luaRobot.lstate.NewFunction(luaRobot.AddTimer))
	luaRobot.lstate.SetGlobal("DelTimer", luaRobot.lstate.NewFunction(luaRobot.DelTimer))
	luaRobot.lstate.SetGlobal("DelMSTimer", luaRobot.lstate.NewFunction(luaRobot.DelTimer))

	luaRobot.lstate.Panic = func(*lua.LState) {
		luaRobot.Panic(errors.New("LUA Panic"))
		return
	}

	os.Setenv(lua.LuaPath, "lua/?.lua;;")
	if err := luaRobot.lstate.DoFile("lua/main.lua"); err != nil {
		luaRobot.Panic(err)
		return
	}

	luaRobot.onReceive = lua.P{
		Fn:      luaRobot.lstate.GetGlobal("OnReceive"),
		NRet:    0,
		Protect: true,
	}

	luaRobot.onSend = lua.P{
		Fn:      luaRobot.lstate.GetGlobal("OnSend"),
		NRet:    1,
		Protect: true,
	}

	l.Logf("Lua 环境初始化完成。")
}

func (l *LuaRobot) OnReceive(raw, input string) {
	L := l.lstate
	err := L.CallByParam(l.onReceive, lua.LString(raw), lua.LString(input))
	if err != nil {
		luaRobot.Panic(err)
	}
}

func (l *LuaRobot) OnSend(cmd string) bool {
	L := l.lstate
	err := L.CallByParam(l.onSend, lua.LString(cmd))
	if err != nil {
		l.Panic(err)
	}

	ret := L.Get(-1)
	L.Pop(1)

	if ret == lua.LFalse {
		return false
	} else {
		return true
	}
}

func (l *LuaRobot) Panic(err error) {
	l.Logf("Lua error: [%s]", err)
}

func (l *LuaRobot) RegEx(L *lua.LState) int {
	text := L.ToString(1)
	regex := L.ToString(2)

	re, err := regexp.Compile(regex)
	if err != nil {
		L.Push(lua.LString("0"))
		return 1
	}

	matchs := re.FindAllStringSubmatch(text, -1)
	if matchs == nil {
		L.Push(lua.LString("0"))
		return 1
	}

	subs := matchs[0]
	length := len(subs)
	if length == 1 {
		L.Push(lua.LString("-1"))
		return 1
	}

	L.Push(lua.LString(fmt.Sprintf("%d", length-1)))

	for i := 1; i < length; i++ {
		L.Push(lua.LString(subs[i]))
	}

	return length
}

func (l *LuaRobot) Print(L *lua.LState) int {
	text := L.ToString(1)
	l.Logf("Lua.Print: %s", text)
	return 0
}

func (l *LuaRobot) Echo(L *lua.LState) int {
	text := L.ToString(1)
	l.Logf("Lua.Echo: %s", text)
	return 0
}

func (l *LuaRobot) Show(L *lua.LState) int {
	text := L.ToString(1)

	re := regexp.MustCompile(`\$(BLK|NOR|RED|HIR|GRN|HIG|YEL|HIY|BLU|HIB|MAG|HIM|CYN|HIC|WHT|HIW|BNK|REV|U)\$`)
	text = re.ReplaceAllStringFunc(text, func(code string) string {
		switch code {
		case "$BLK$":
			return "[black::]"
		case "$NOR$":
			return "[-:-:-]"
		case "$RED$":
			return "[red::]"
		case "$HIR$":
			return "[red::b]"
		case "$GRN$":
			return "[green::]"
		case "$HIG$":
			return "[green::b]"
		case "$YEL$":
			return "[yellow::]"
		case "$HIY$":
			return "[yellow::b]"
		case "$BLU$":
			return "[blue::]"
		case "$HIB$":
			return "[blue::b]"
		case "$MAG$":
			return "[darkmagenta::]"
		case "$HIM$":
			return "[#ff00ff::]"
		case "$CYN$":
			return "[dardcyan::]"
		case "$HIC$":
			return "[#00ffff::]"
		case "$WHT$":
			return "[white::]"
		case "$HIW$":
			return "[#ffffff::]"
		case "$BNK$":
			return "[::l]"
		case "$REV$":
			return "[::7]"
		case "$U$":
			return "[::u]"
		default:
			l.Logf("Find Unknown Color Code: %s", code)
		}
		return ""
	})

	l.Logf("%s", text)

	return 0
}

func (l *LuaRobot) Run(L *lua.LState) int {
	text := L.ToString(1)
	l.Logf("Lua.Run: %s", text)
	return 0
}

func (l *LuaRobot) Send(L *lua.LState) int {
	text := L.ToString(1)
	text = UTF8_TO_GBK(text)
	fmt.Fprintln(l.mud, text)
	return 0
}

func (l *LuaRobot) AddTimer(L *lua.LState) int {
	id := L.ToString(1)
	code := L.ToString(2)
	delay := L.ToInt(3)
	times := L.ToInt(4)

	go func() {
		count := 0
		quit := make(chan bool, 1)
		timer := Timer{
			id:       id,
			code:     code,
			delay:    delay,
			maxTimes: times,
			times:    0,
			quit:     quit,
		}
		v, exists := l.timer.LoadOrStore(id, timer)
		if exists {
			v.(Timer).quit <- true
			l.timer.Store(id, timer)
		}

		for {
			select {
			case <-quit:
				return
			case <-time.After(time.Millisecond * time.Duration(delay)):
				timer.Emit(l)
				count++
				if times > 0 && times >= count {
					return
				}
			}
		}
	}()

	return 0
}

func (l *LuaRobot) DelTimer(L *lua.LState) int {
	id := L.ToString(1)
	v, ok := l.timer.Load(id)
	if ok {
		v.(Timer).quit <- true
	}
	l.timer.Delete(id)
	return 0
}

func (l *LuaRobot) Logf(format string, a ...interface{}) {
	if mainWindow == nil {
		log.Printf(format, a...)
	} else {
		fmt.Fprintf(mainWindow, format+"\n", a...)
	}
	return
}

func main() {
	initLua()

	app = tview.NewApplication()
	mainWindow = tview.NewTextView().
		SetDynamicColors(true).
		SetScrollable(true).
		SetChangedFunc(func() {
			app.Draw()
		})

	mudInput, output := mudServer()
	go func() {
		w := tview.ANSIWriter(mainWindow)
		for raw := range mudInput {
			fmt.Fprintln(w, raw)
			input := ansiRe.ReplaceAllString(raw, "")
			luaRobot.OnReceive(raw, input)
		}
	}()

	luaRobot.mud = output

	cmdLine := tview.NewInputField().
		SetFieldBackgroundColor(tcell.ColorBlack).
		SetLabelColor(tcell.ColorWhite).
		SetLabel("命令: ")

	cmdLine.SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEnter {
			cmd := cmdLine.GetText()
			if cmd == "exit" || cmd == "quit" {
				app.Stop()
				log.Print("程序已退出。")
				return
			} else if cmd == "lua.reload" {
				luaRobot.Reload()
			} else if strings.HasPrefix(cmd, `'`) {
				// 北侠默认支持单引号自动变成 say 命令效果
			} else if strings.HasPrefix(cmd, `"`) {
				cmd = "chat " + cmd[1:]
			} else if strings.HasPrefix(cmd, `*`) {
				cmd = "chat* " + cmd[1:]
			} else if strings.HasPrefix(cmd, `;`) {
				cmd = "rumor " + cmd[1:]
			}

			if cmd != "" {
				cmdLine.SetText("")
				fmt.Fprintln(mainWindow, cmd)
				cmd = UTF8_TO_GBK(cmd)
				needSend := luaRobot.OnSend(cmd)
				if needSend {
					fmt.Fprintln(output, cmd)
				}
			}
			mainWindow.ScrollToEnd()
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
		AddItem(mainWindow, 0, 1, false).
		AddItem(cmdLine, 1, 1, false)

	app.SetRoot(mainFrame, true).
		SetFocus(cmdLine).
		SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			if event.Key() == tcell.KeyCtrlC {
				if mainWindow.HasFocus() {
					mainWindow.ScrollToEnd()
					app.SetFocus(cmdLine)
				} else {
					cmdLine.SetText("")
				}
				return nil
			} else if event.Key() == tcell.KeyCtrlB {
				app.SetFocus(mainWindow)
				row, _ := mainWindow.GetScrollOffset()
				row -= 10
				if row < 0 {
					row = 0
				}
				mainWindow.ScrollTo(row, 0)
			} else if event.Key() == tcell.KeyCtrlF {
				app.SetFocus(mainWindow)
				row, _ := mainWindow.GetScrollOffset()
				row += 10
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

	log.Printf("连接到服务器 %s...", serverAddress)
	conn, _ := net.Dial("tcp", serverAddress)
	rd := bufio.NewReader(conn)
	log.Print("连接成功。")

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
