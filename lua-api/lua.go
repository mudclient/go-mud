package lua

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"regexp"
	"sync"
	"time"

	"github.com/flw-cn/printer"
	lua "github.com/yuin/gopher-lua"
)

type LuaRobotConfig struct {
	Enable bool   `flag:"|true|是否加载 Lua 机器人"`
	Path   string `flag:"p|lua|Lua 插件路径 {path}"`
}

type LuaRobot struct {
	config LuaRobotConfig

	screen printer.Printer
	mud    io.Writer

	lstate    *lua.LState
	onReceive lua.P
	onSend    lua.P

	timer sync.Map
}

func NewLuaRobot(config LuaRobotConfig) *LuaRobot {
	return &LuaRobot{
		config: config,
		screen: printer.NewSimplePrinter(os.Stdout),
	}
}

func (l *LuaRobot) Init() {
	if !l.config.Enable {
		return
	}

	if err := l.Reload(); err != nil {
		l.screen.Println("Lua 初始化失败。")
		return
	}
}

func (l *LuaRobot) SetScreen(w printer.Printer) {
	l.screen = w
}

func (l *LuaRobot) SetMud(w io.Writer) {
	l.mud = w
}

func (l *LuaRobot) Reload() error {
	mainFile := path.Join(l.config.Path, "main.lua")
	if _, err := os.Open(mainFile); err != nil {
		l.screen.Printf("Load error: %s\n", err)
		l.screen.Println("无法打开 lua 主程序，请检查你的配置。")
		return err
	}

	if l.lstate != nil {
		l.lstate.Close()
		l.screen.Println("Lua 环境已关闭。")
	}

	l.screen.Println("初始化 Lua 环境...")

	luaPath := path.Join(l.config.Path, "?.lua")
	os.Setenv(lua.LuaPath, luaPath+";;")

	l.lstate = lua.NewState()

	l.lstate.SetGlobal("RegEx", l.lstate.NewFunction(l.LuaRegEx))
	l.lstate.SetGlobal("Echo", l.lstate.NewFunction(l.LuaEcho))
	l.lstate.SetGlobal("Print", l.lstate.NewFunction(l.LuaPrint))
	l.lstate.SetGlobal("Run", l.lstate.NewFunction(l.LuaRun))
	l.lstate.SetGlobal("Send", l.lstate.NewFunction(l.LuaSend))
	l.lstate.SetGlobal("AddTimer", l.lstate.NewFunction(l.LuaAddTimer))
	l.lstate.SetGlobal("AddMSTimer", l.lstate.NewFunction(l.LuaAddTimer))
	l.lstate.SetGlobal("DelTimer", l.lstate.NewFunction(l.LuaDelTimer))
	l.lstate.SetGlobal("DelMSTimer", l.lstate.NewFunction(l.LuaDelTimer))

	l.lstate.Panic = func(*lua.LState) {
		l.Panic(errors.New("LUA Panic"))
		return
	}

	if err := l.lstate.DoFile(mainFile); err != nil {
		l.lstate.Close()
		l.lstate = nil
		return err
	}

	l.onReceive = lua.P{
		Fn:      l.lstate.GetGlobal("OnReceive"),
		NRet:    0,
		Protect: true,
	}

	l.onSend = lua.P{
		Fn:      l.lstate.GetGlobal("OnSend"),
		NRet:    1,
		Protect: true,
	}

	l.screen.Println("Lua 环境初始化完成。")

	return nil
}

func (l *LuaRobot) OnReceive(raw, input string) {
	if l.lstate == nil {
		return
	}

	L := l.lstate
	err := L.CallByParam(l.onReceive, lua.LString(raw), lua.LString(input))
	if err != nil {
		l.Panic(err)
	}
}

func (l *LuaRobot) OnSend(cmd string) bool {
	if l.lstate == nil {
		return true
	}

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
	l.screen.Printf("Lua error: [%v]\n", err)
}

func (l *LuaRobot) LuaRegEx(L *lua.LState) int {
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

func (l *LuaRobot) LuaPrint(L *lua.LState) int {
	text := L.ToString(1)
	l.screen.Println(text)
	return 0
}

func (l *LuaRobot) LuaEcho(L *lua.LState) int {
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
			l.screen.Printf("Find Unknown Color Code: %s\n", code)
		}
		return ""
	})

	l.screen.Println(text)

	// TODO: 这里暂时不支持 ANSI 到 PLAIN 的转换
	l.OnReceive(text, text)

	return 0
}

func (l *LuaRobot) LuaRun(L *lua.LState) int {
	text := L.ToString(1)
	l.screen.Println(text)
	return 0
}

func (l *LuaRobot) LuaSend(L *lua.LState) int {
	text := L.ToString(1)
	fmt.Fprintln(l.mud, text)
	return 0
}

func (l *LuaRobot) LuaAddTimer(L *lua.LState) int {
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

func (l *LuaRobot) LuaDelTimer(L *lua.LState) int {
	id := L.ToString(1)
	v, ok := l.timer.Load(id)
	if ok {
		v.(Timer).quit <- true
	}
	l.timer.Delete(id)
	return 0
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
		l.screen.Printf("Lua Error: %s\n", err)
	}
}
