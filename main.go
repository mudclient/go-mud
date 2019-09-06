package main

import (
	"regexp"
	"runtime"
	"strings"
	"time"

	lua "github.com/dzpao/go-mud/lua-api"
	"github.com/dzpao/go-mud/mud"
	"github.com/dzpao/go-mud/ui"
	smartConfig "github.com/flw-cn/go-smartConfig"
	"golang.org/x/text/width"
)

type ClientConfig struct {
	UI  ui.UIConfig
	Mud mud.MudConfig
	Lua lua.LuaRobotConfig
}

type Client struct {
	config ClientConfig
	ui     *ui.UI
	lua    *lua.LuaRobot
	mud    *mud.MudServer
	quit   chan bool
}

func main() {
	config := ClientConfig{}
	smartConfig.LoadConfig("go-mud", "0.6", &config)
	client := NewClient(config)
	client.Run()
}

func NewClient(config ClientConfig) *Client {
	return &Client{
		config: config,
		ui:     ui.NewUI(config.UI),
		lua:    lua.NewLuaRobot(config.Lua),
		mud:    mud.NewMudServer(config.Mud),
		quit:   make(chan bool, 1),
	}
}

func (c *Client) Run() {
	ansiRe := regexp.MustCompile("\x1b" + `\[\d*(?:;\d*(?:;\d*)?)?(?:A|D|K|m)`)

	c.ui.Create()
	go c.ui.Run()
	c.lua.SetScreen(c.ui)
	c.lua.SetMud(c.mud)
	c.lua.Init()
	c.mud.SetScreen(c.ui)
	go c.mud.Run()

	beautify := ambiWidthAdjuster(c.config.UI.AmbiguousWidth)

LOOP:
	for {
		select {
		case <-c.quit:
			break LOOP
		case rawLine, ok := <-c.mud.Input():
			if ok {
				showLine := beautify(rawLine)
				plainLine := ansiRe.ReplaceAllString(rawLine, "")
				c.ui.Println(showLine)
				c.lua.OnReceive(rawLine, plainLine)
			} else {
				c.ui.Println("程序即将退出。")
				time.Sleep(3 * time.Second)
				break LOOP
			}
		case cmd := <-c.ui.Input():
			c.DoCmd(cmd)
		}
	}

	c.ui.Stop()
	c.mud.Stop()
}

func (c *Client) DoCmd(cmd string) {
	if cmd == "exit" || cmd == "quit" {
		c.quit <- true
		return
	} else if cmd == "lua.reload" {
		c.lua.Reload()
		return
	} else if strings.HasPrefix(cmd, `'`) {
		// 北侠默认支持单引号自动变成 say 命令效果
	} else if strings.HasPrefix(cmd, `"`) {
		cmd = "chat " + cmd[1:]
	} else if strings.HasPrefix(cmd, `*`) {
		cmd = "chat* " + cmd[1:]
	} else if strings.HasPrefix(cmd, `;`) {
		cmd = "rumor " + cmd[1:]
	}

	c.ui.Println(cmd)
	needSend := c.lua.OnSend(cmd)
	if needSend {
		c.mud.Println(cmd)
	}
}

func ambiWidthAdjuster(option string) func(string) string {
	singleAmbiguousWidth := func(str string) string {
		return str
	}
	option = strings.ToLower(option)
	switch option {
	case "double":
		return doubleAmbiguousWidth
	case "single":
		return singleAmbiguousWidth
	case "auto":
		if runtime.GOOS == "windows" {
			return singleAmbiguousWidth
		} else {
			return doubleAmbiguousWidth
		}
	default:
		return singleAmbiguousWidth
	}
}

func doubleAmbiguousWidth(str string) string {
	newStr := ""
	for _, c := range str {
		newStr += string(c)
		switch c {
		case '┌', '┬', '├', '┼', '└', '┴', '─',
			'╓', '╥', '╟', '╫', '╙', '╨',
			'╭', '╰':
			newStr += "─"
		case '╔', '╦', '╠', '╬', '╚', '╩', '═',
			'╒', '╤', '╞', '╪', '╘', '╧':
			newStr += "═"
		case '█', '▇', '▆', '▅', '▄', '▃', '▂', '▁', '▀':
			newStr += string(c)
		default:
			p := width.LookupRune(c)
			if p.Kind() == width.EastAsianAmbiguous {
				newStr += " "
			}
		}
	}

	return newStr
}
