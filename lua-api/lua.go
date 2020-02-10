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

type Config struct {
	Enable bool   `flag:"|true|是否加载 Lua 机器人"`
	Path   string `flag:"p|lua|Lua 插件路径 {path}"`
}

type API struct {
	config Config

	screen printer.Printer
	mud    io.Writer

	lstate    *lua.LState
	onReceive lua.P
	onSend    lua.P

	timer sync.Map
}

func NewAPI(config Config) *API {
	return &API{
		config: config,
		screen: printer.NewSimplePrinter(os.Stdout),
	}
}

func (api *API) Init() {
	if !api.config.Enable {
		return
	}

	if err := api.Reload(); err != nil {
		api.screen.Println("Lua 初始化失败。")
		return
	}
}

func (api *API) SetScreen(w printer.Printer) {
	api.screen = w
}

func (api *API) SetMud(w io.Writer) {
	api.mud = w
}

func (api *API) Reload() error {
	mainFile := path.Join(api.config.Path, "main.lua")
	if _, err := os.Open(mainFile); err != nil {
		api.screen.Printf("Load error: %s\n", err)
		api.screen.Println("无法打开 lua 主程序，请检查你的配置。")
		return err
	}

	if api.lstate != nil {
		api.lstate.Close()
		api.screen.Println("Lua 环境已关闭。")
	}

	api.screen.Println("初始化 Lua 环境...")

	luaPath := path.Join(api.config.Path, "?.lua")
	os.Setenv(lua.LuaPath, luaPath+";;")

	api.lstate = lua.NewState()

	l := api.lstate

	l.SetGlobal("RegEx", l.NewFunction(api.LuaRegEx))
	l.SetGlobal("Echo", l.NewFunction(api.LuaEcho))
	l.SetGlobal("Print", l.NewFunction(api.LuaPrint))
	l.SetGlobal("Run", l.NewFunction(api.LuaRun))
	l.SetGlobal("Send", l.NewFunction(api.LuaSend))
	l.SetGlobal("AddTimer", l.NewFunction(api.LuaAddTimer))
	l.SetGlobal("AddMSTimer", l.NewFunction(api.LuaAddTimer))
	l.SetGlobal("DelTimer", l.NewFunction(api.LuaDelTimer))
	l.SetGlobal("DelMSTimer", l.NewFunction(api.LuaDelTimer))

	l.Panic = func(*lua.LState) {
		api.Panic(errors.New("LUA Panic"))
		return
	}

	if err := l.DoFile(mainFile); err != nil {
		l.Close()
		api.lstate = nil
		return err
	}

	api.onReceive = lua.P{
		Fn:      l.GetGlobal("OnReceive"),
		NRet:    0,
		Protect: true,
	}

	api.onSend = lua.P{
		Fn:      l.GetGlobal("OnSend"),
		NRet:    1,
		Protect: true,
	}

	api.screen.Println("Lua 环境初始化完成。")

	return nil
}

func (api *API) OnReceive(raw, input string) {
	if api.lstate == nil {
		return
	}

	l := api.lstate
	err := l.CallByParam(api.onReceive, lua.LString(raw), lua.LString(input))
	if err != nil {
		api.Panic(err)
	}
}

func (api *API) OnSend(cmd string) bool {
	if api.lstate == nil {
		return true
	}

	l := api.lstate
	err := l.CallByParam(api.onSend, lua.LString(cmd))
	if err != nil {
		api.Panic(err)
	}

	ret := l.Get(-1)
	l.Pop(1)

	if ret == lua.LFalse {
		return false
	} else {
		return true
	}
}

func (api *API) Panic(err error) {
	api.screen.Printf("Lua error: [%v]\n", err)
}

func (api *API) LuaRegEx(l *lua.LState) int {
	text := l.ToString(1)
	regex := l.ToString(2)

	re, err := regexp.Compile(regex)
	if err != nil {
		l.Push(lua.LString("0"))
		return 1
	}

	matchs := re.FindAllStringSubmatch(text, -1)
	if matchs == nil {
		l.Push(lua.LString("0"))
		return 1
	}

	subs := matchs[0]
	length := len(subs)
	if length == 1 {
		l.Push(lua.LString("-1"))
		return 1
	}

	l.Push(lua.LString(fmt.Sprintf("%d", length-1)))

	for i := 1; i < length; i++ {
		l.Push(lua.LString(subs[i]))
	}

	return length
}

func (api *API) LuaPrint(l *lua.LState) int {
	text := l.ToString(1)
	api.screen.Println(text)
	return 0
}

func (api *API) LuaEcho(l *lua.LState) int {
	text := l.ToString(1)

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
			api.screen.Printf("Find Unknown Color Code: %s\n", code)
		}
		return ""
	})

	api.screen.Println(text)

	// TODO: 这里暂时不支持 ANSI 到 PLAIN 的转换
	api.OnReceive(text, text)

	return 0
}

func (api *API) LuaRun(l *lua.LState) int {
	text := l.ToString(1)
	api.screen.Println(text)
	return 0
}

func (api *API) LuaSend(l *lua.LState) int {
	text := l.ToString(1)
	fmt.Fprintln(api.mud, text)
	return 0
}

func (api *API) LuaAddTimer(l *lua.LState) int {
	id := l.ToString(1)
	code := l.ToString(2)
	delay := l.ToInt(3)
	times := l.ToInt(4)

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
		v, exists := api.timer.LoadOrStore(id, timer)
		if exists {
			v.(Timer).quit <- true
			api.timer.Store(id, timer)
		}

		for {
			select {
			case <-quit:
				return
			case <-time.After(time.Millisecond * time.Duration(delay)):
				timer.Emit(api)
				count++
				if times > 0 && times >= count {
					return
				}
			}
		}
	}()

	return 0
}

func (api *API) LuaDelTimer(l *lua.LState) int {
	id := l.ToString(1)
	v, ok := api.timer.Load(id)
	if ok {
		v.(Timer).quit <- true
	}
	api.timer.Delete(id)
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

func (t *Timer) Emit(l *API) {
	err := l.lstate.DoString(`call_timer_actions("` + t.id + `")`)
	if err != nil {
		l.screen.Printf("Lua Error: %s\n", err)
	}
}
