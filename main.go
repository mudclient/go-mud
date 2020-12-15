package main

import (
	"fmt"
	"log"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/flw-cn/go-smartConfig"
	"github.com/mattn/go-runewidth"
	"github.com/rivo/tview"
	"github.com/spf13/cobra"
	"golang.org/x/text/width"

	"github.com/mudclient/go-mud/app"
	"github.com/mudclient/go-mud/lua-api"
	"github.com/mudclient/go-mud/mud"
	"github.com/mudclient/go-mud/ui"
)

type ClientConfig struct {
	UI  ui.Config
	Mud mud.Config
	Lua lua.Config
}

type Client struct {
	config ClientConfig
	ui     *ui.UI
	lua    *lua.API
	mud    *mud.Server
	quit   chan bool

	debug bool
}

func main() {
	cobra.MousetrapHelpText = "" // 允许在 Windows(R) 下直接双击运行
	config := ClientConfig{}
	smartConfig.VersionDetail = app.VersionDetail()
	smartConfig.LoadConfig(app.AppName, app.Version, &config)

	client := NewClient(config)
	client.Run()
}

func NewClient(config ClientConfig) *Client {
	return &Client{
		config: config,
		ui:     ui.NewUI(config.UI),
		lua:    lua.NewAPI(config.Lua),
		mud:    mud.NewServer(config.Mud),
		quit:   make(chan bool, 1),
	}
}

func (c *Client) Run() {
	ansiRe := regexp.MustCompile("\x1b" + `\[\d*(?:;\d*(?:;\d*)?)?(?:A|D|K|m)`)

	title := fmt.Sprintf("%s(%s), server = %s:%d",
		app.AppName, app.Version,
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
				if c.debug {
					line := showLine
					line = strings.ReplaceAll(line, "\x1b[", "<OSI>")
					line = strings.ReplaceAll(line, "\t", "<TAB>")
					c.ui.Println(line)
					line = tview.TranslateANSI(showLine)
					line = tview.Escape(line)
					c.ui.Println(line)
				}
				c.ui.Println(showLine)
				c.lua.OnReceive(rawLine, plainLine)
			} else {
				defer log.Printf("连接已断开。")
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
	switch cmd {
	case "exit", "quit":
		c.quit <- true
		return
	case "/version":
		c.ui.Print(app.VersionDetail())
		return
	case "/reload-lua":
		_ = c.lua.Reload()
		return
	case "/debug":
		c.debug = !c.debug
		return
	case "/lines":
		for i := 0; i < 100000; i++ {
			c.ui.Printf("%d %s\n", i, time.Now())
		}
		c.ui.Println("测试内容填充完毕")
		return
	}

	if len(cmd) > 0 {
		switch cmd[0] {
		case '\'':
			cmd = "say " + cmd[1:]
		case '"':
			cmd = "chat " + cmd[1:]
		case '*':
			cmd = "chat* " + cmd[1:]
		case ';':
			cmd = "rumor " + cmd[1:]
		}
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
		}
		return doubleAmbiguousWidth
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
