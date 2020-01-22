package main

import (
	"fmt"
	"regexp"
	"runtime"
	"strings"
	"time"

	lua "github.com/dzpao/go-mud/lua-api"
	"github.com/dzpao/go-mud/mud"
	"github.com/dzpao/go-mud/ui"
	smartConfig "github.com/flw-cn/go-smartConfig"
	"github.com/mattn/go-runewidth"
	"github.com/rivo/tview"
	"github.com/spf13/cobra"
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
	cobra.MousetrapHelpText = "" // 允许在 Windows(R) 下直接双击运行
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

	title := fmt.Sprintf("GoMud v0.6.1 beta, server = %s:%d",
		c.config.Mud.Host, c.config.Mud.Port)
	c.ui.Create(title)
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
				if debug {
					line := showLine
					line = strings.Replace(line, "\x1b[", "<OSI>", -1)
					line = strings.Replace(line, "\t", "<TAB>", -1)
					c.ui.Println(line)
					line = tview.TranslateANSI(showLine)
					line = tview.Escape(line)
					c.ui.Println(line)
				}
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

var debug bool

func (c *Client) DoCmd(cmd string) {
	if cmd == "exit" || cmd == "quit" {
		c.quit <- true
		return
	} else if cmd == "/reload-lua" {
		c.lua.Reload()
		return
	} else if cmd == "/debug" {
		debug = !debug
		return
	} else if cmd == "/lines" {
		for i := 0; i < 100000; i++ {
			c.ui.Printf("%d %s\n", i, time.Now())
		}
		c.ui.Println("测试内容填充完毕")
		return
	} else if strings.HasPrefix(cmd, `'`) {
		cmd = "say " + cmd[1:]
	} else if strings.HasPrefix(cmd, `"`) {
		cmd = "chat " + cmd[1:]
	} else if strings.HasPrefix(cmd, `*`) {
		cmd = "chat* " + cmd[1:]
	} else if strings.HasPrefix(cmd, `;`) {
		cmd = "rumor " + cmd[1:]
	} else if cmd == "debug" {
		debug = !debug
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
	spaceAmbiguousWidth := func(str string) string {
		newStr := ""
		for _, c := range str {
			newStr += string(c)
			p := width.LookupRune(c)
			if p.Kind() == width.EastAsianAmbiguous {
				newStr += " "
			}
		}
		return newStr
	}
	option = strings.ToLower(option)
	switch option {
	case "double":
		return doubleAmbiguousWidth
	case "single":
		return singleAmbiguousWidth
	case "space":
		return spaceAmbiguousWidth
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
		case '┌', '┎', '└', '┖', '─',
			'┬', '┭', '┰', '┱', '├', '┞', '┟', '┠', '┴', '┵', '┸', '┹',
			'┼', '╁', '╀', '╂', '┽', '╃', '╅', '╉',
			'╓', '╙', '╥', '╟', '╨', '╫', '╭', '╰':
			newStr += "─"
		case '┏', '┍', '┗', '┕', '━',
			'┳', '┲', '┯', '┮', '┣', '┢', '┡', '┝', '┻', '┺', '┷', '┶',
			'╋', '╇', '╈', '┿', '╊', '╆', '╄', '┾':
			newStr += "━"
		case '╔', '╦', '╠', '╬', '╚', '╩', '═',
			'╒', '╤', '╞', '╪', '╘', '╧':
			newStr += "═"
			// 上述三类字符的共同点就是右侧有水平线线头，因此以相应的线条来延伸它们。
		case '█', '▇', '▆', '▅', '▄', '▃', '▂', '▁', '▀',
			'▔', '┄', '┅', '┈', '┉':
			// 这几个字符从语义上讲宽度是含糊的，
			// 且实际显示效果占一个拉丁字母宽度，并且充斥了整个宽度，因此双写以延伸它们
			newStr += string(c)
		case '▕', '▒', '▓':
			// 这几个字符虽然从语义上讲宽度是含糊的，
			// 但在某些字体中显示效果已经是两个拉丁字母宽度了，
			// 因此仅用空格来调整宽度，不再重复，以免重叠
			newStr += " "
		case '╌', '╍', '╶', '╺', '╾', '╼', '░', '▗', '▙', '▚', '▜', '▟', '▝', '▛', '▞', '▐':
			// 这些字符的 East_Assia_Width 属性都是 single，语义上就只占一个拉丁字母的宽度，因此什么也不做
		default:
			// U+2500 ~ U+259F 区间除了本函数列出来的字符之外，其余字符从外观上看只能是通过空格来扩展
			p := width.LookupRune(c)
			if p.Kind() == width.EastAsianAmbiguous && runewidth.RuneWidth(c) == 1 {
				newStr += " "
			}
		}
	}

	return newStr
}
