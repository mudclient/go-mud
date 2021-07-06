package ui

import (
	"strings"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

const defaultHistorySize = 10000

type Readline struct {
	*tview.InputField

	history     []string
	curSel      int
	historySize int

	repeat   bool
	autoTrim bool

	cmds      []string
	split     bool
	separator string
}

func NewReadline() *Readline {
	return &Readline{
		InputField:  tview.NewInputField(),
		history:     make([]string, 0, 32),
		curSel:      0,
		historySize: defaultHistorySize,
	}
}

func (r *Readline) SetSeparator(s string) *Readline {
	r.split = true
	r.separator = s
	return r
}

func (r *Readline) SetRepeat(b bool) *Readline {
	r.repeat = b
	return r
}

func (r *Readline) SetAutoTrim(b bool) *Readline {
	r.autoTrim = b
	return r
}

func (r *Readline) InputCapture(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyCtrlC:
		r.InputField.SetText("")
		return nil
	case tcell.KeyUp:
		if r.curSel > 0 {
			r.curSel--
			r.InputField.SetText(r.history[r.curSel])
		}
		return nil
	case tcell.KeyDown:
		if r.curSel == len(r.history)-1 {
			r.curSel++
			r.InputField.SetText("")
		}
		if r.curSel < len(r.history)-1 {
			r.curSel++
			r.InputField.SetText(r.history[r.curSel])
		}
		return nil
	default:
	}

	return event
}

func (r *Readline) Enter() string {
	text := r.InputField.GetText()

	r.cmds = nil

	if text != "" && r.autoTrim {
		text = strings.TrimSpace(text)
		// 如果 trim 之后变成了空串，则至少保留一个空格，以免用户发不出空格
		if text == "" {
			text = " "
		}else if r.split && "" != r.separator {
			//命令分割
			r.cmds = strings.Split(text,r.separator)
			if len(r.cmds) < 2 {
				r.cmds = nil
			}
		}
	}



	last := ""
	if len(r.history) > 0 {
		last = r.history[len(r.history)-1]
	}

	if text == "" && r.repeat && last != "" {
		text = last
	} else if text != " " && text != last {
		if len(r.history) >= r.historySize {
			r.history = r.history[1 : len(r.history)-1]
		}
		r.history = append(r.history, text)
		r.curSel = len(r.history)
	}

	r.InputField.SetText("")

	return text
}
