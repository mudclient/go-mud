package mud

import (
	"bytes"
	"io"
	"net"
	"time"
)

type Message interface {
	IsMessage()
}

type CSIMessage struct {
	Parameter    bytes.Buffer
	Intermediate bytes.Buffer
	Command      byte
}

type Line struct{ *bytes.Buffer }
type IncompleteLine struct{ *bytes.Buffer }
type EOF bool

func (CSIMessage) IsMessage()     {}
func (Line) IsMessage()           {}
func (IncompleteLine) IsMessage() {}
func (EOF) IsMessage()            {}

type ReaderWithDeadline interface {
	io.Reader
	SetReadDeadline(t time.Time) error
}

type Scanner struct {
	r     ReaderWithDeadline
	buf   bytes.Buffer
	state ScannerStatus
	msg   Message
	done  bool
}

type ScannerStatus int

const (
	stText ScannerStatus = iota
	stIACCommand
	stANSICodes
)

func NewScanner(r ReaderWithDeadline) *Scanner {
	return &Scanner{
		r: r,
	}
}

func (s *Scanner) Scan() Message {
	if s.done {
		return EOF(true)
	}

	iacCmd := NewIACMessage()
	line := new(bytes.Buffer)

	for {
		b, err := s.readByte()
		if err == io.EOF {
			s.done = true
			return EOF(true)
		} else if err != nil {
			if line.Len() == 0 {
				continue
			} else {
				return IncompleteLine{line}
			}
		}

		switch s.state {
		case stText:
			switch b {
			case IAC:
				s.state = stIACCommand
				if line.Len() > 0 {
					return IncompleteLine{line}
				}
			case '\r': // 忽略
			case '\n':
				return Line{line}
			default:
				line.WriteByte(b)
			}
		case stIACCommand:
			if b == IAC {
				return *iacCmd
			} else if iacCmd.Scan(b) {
				s.state = stText
				return *iacCmd
			}
		}
	}
}

// readByte 努力读取一个字节，并返回成功(nil)或两种错误之一：
//     timeout:    超时
//     io.EOF:     连接已经不可用
// 优先从 s.buf 中读取，如果 s.buf 为空，则从 s.r 中读取
func (s *Scanner) readByte() (byte, error) {
	b, err := s.buf.ReadByte()
	if err != io.EOF {
		return b, err
	}

	s.r.SetReadDeadline(time.Now().Add(1 * time.Second))
	bytes := make([]byte, 1024)
	n, err := s.r.Read(bytes)
	if err == nil && n > 0 {
		s.buf.Write(bytes[:n])
		return s.buf.ReadByte()
	}

	e, ok := err.(net.Error)
	if ok && (e.Timeout() || e.Temporary()) {
		return 0, err
	}

	return 0, io.EOF
}
