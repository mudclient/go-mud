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

var errPanic = errors.New("LUA Panic")

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

	_ = api.Reload()
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
		api.screen.Printf("Load error: %v\n", err)
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

	// 为 Lua 环境提供 API
	api.register()

	l := api.lstate

	l.Panic = func(*lua.LState) {
		api.Panic(errPanic)
	}

	if err := l.DoFile(mainFile); err != nil {
		l.Close()
		api.screen.Printf("Lua 初始化失败：%v\n", err)
		api.lstate = nil
		return err
	}

	// 和 Lua 环境中的钩子相连接
	api.hookOn()

	api.screen.Println("Lua 环境初始化完成。")

	return nil
}

func (api *API) register() {
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
}

func (api *API) hookOn() {
	l := api.lstate

	if v := l.GetGlobal("OnReceive"); v.Type() == lua.LTFunction {
		api.onReceive = lua.P{
			Fn:      v,
			NRet:    0,
			Protect: true,
		}
	} else {
		api.screen.Println("Lua 环境中未定义 OnReceive 函数，将无法接收游戏数据。")
	}

	if v := l.GetGlobal("OnSend"); v.Type() == lua.LTFunction {
		api.onSend = lua.P{
			Fn:      v,
			NRet:    1,
			Protect: true,
		}
	} else {
		api.screen.Println("Lua 环境中未定义 OnSend 函数，将无法获知向游戏发送的数据。")
	}
}

func (api *API) OnReceive(raw, input string) {
	if api.lstate == nil ||
		api.onReceive.Fn == nil ||
		api.onReceive.Fn.Type() != lua.LTFunction {
		return
	}

	l := api.lstate
	err := l.CallByParam(api.onReceive, lua.LString(raw), lua.LString(input))
	if err != nil {
		api.Panic(err)
	}
}

func (api *API) OnSend(cmd string) bool {
	if api.lstate == nil ||
		api.onSend.Fn == nil ||
		api.onSend.Fn.Type() != lua.LTFunction {
		return true
	}

	l := api.lstate
	err := l.CallByParam(api.onSend, lua.LString(cmd))
	if err != nil {
		api.Panic(err)
	}

	ret := l.Get(-1)
	l.Pop(1)

	return ret != lua.LFalse
}

func (api *API) Panic(err error) {
	api.screen.Printf("Lua error: %v\n", err)
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

	codes := map[string]string{
		"$BLK$": "[black::]",
		"$NOR$": "[-:-:-]",
		"$RED$": "[red::]",
		"$HIR$": "[red::b]",
		"$GRN$": "[green::]",
		"$HIG$": "[green::b]",
		"$YEL$": "[yellow::]",
		"$HIY$": "[yellow::b]",
		"$BLU$": "[blue::]",
		"$HIB$": "[blue::b]",
		"$MAG$": "[darkmagenta::]",
		"$HIM$": "[#ff00ff::]",
		"$CYN$": "[dardcyan::]",
		"$HIC$": "[#00ffff::]",
		"$WHT$": "[white::]",
		"$HIW$": "[#ffffff::]",
		"$BNK$": "[::l]",
		"$REV$": "[::7]",
		"$U$":   "[::u]",
	}

	re := regexp.MustCompile(`\$(BLK|NOR|RED|HIR|GRN|HIG|YEL|HIY|BLU|HIB|MAG|HIM|CYN|HIC|WHT|HIW|BNK|REV|U)\$`)
	text = re.ReplaceAllStringFunc(text, func(code string) string {
		code, ok := codes[code]
		if ok {
			return code
		}
		api.screen.Printf("Find Unknown Color Code: %s\n", code)
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
		l.screen.Printf("Lua Error: %v\n", err)
	}
}
