package mud

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"time"

	"github.com/axgle/mahonia"
	"github.com/flw-cn/printer"
)

var (
	decoder mahonia.Decoder
	encoder mahonia.Encoder
)

func init() {
	decoder = mahonia.NewDecoder("GB18030")
	encoder = mahonia.NewEncoder("GB18030")
}

type Config struct {
	IACDebug bool
	Host     string `flag:"H|mud.pkuxkx.net|服务器 {IP/Domain}"`
	Port     int    `flag:"P|8080|服务器 {Port}"`
}

type Server struct {
	printer.SimplePrinter

	config Config

	screen printer.Printer
	server printer.WritePrinter

	conn  net.Conn
	input chan string
}

func NewServer(config Config) *Server {
	mud := &Server{
		config: config,
		screen: printer.NewSimplePrinter(os.Stdout),
		server: printer.NewSimplePrinter(ioutil.Discard),
		input:  make(chan string, 1024),
	}

	mud.SetOutput(mud.server)

	return mud
}

func (mud *Server) SetScreen(w printer.Printer) {
	mud.screen = w
}

func (mud *Server) Run() {
	serverAddress := fmt.Sprintf("%s:%d", mud.config.Host, mud.config.Port)
	mud.screen.Printf("连接到服务器 %s...", serverAddress)

	var err error
	mud.conn, err = net.DialTimeout("tcp", serverAddress, 4*time.Second)

	if err != nil {
		mud.screen.Println("连接失败。")
		mud.screen.Printf("失败原因: %v\n", err)
		close(mud.input)
		return
	}

	mud.screen.Println("连接成功。")

	netWriter := encoder.NewWriter(mud.conn)
	mud.server.SetOutput(netWriter)

	scanner := NewScanner(mud.conn)

	mud.conn.Write([]byte{IAC, DONT, OptSGA})

LOOP:
	for {
		msg := scanner.Scan()

		switch m := msg.(type) {
		case EOF:
			break LOOP
		case IncompleteLine:
			r := decoder.NewReader(m)
			buf, _ := ioutil.ReadAll(r)
			mud.input <- string(buf)
		case Line:
			r := decoder.NewReader(m)
			buf, _ := ioutil.ReadAll(r)
			mud.input <- string(buf)
		case IACMessage:
			mud.telnetNegotiate(m)
		}
	}

	mud.server.SetOutput(ioutil.Discard)

	mud.screen.Println("连接已断开。")
	mud.screen.Println("TODO: 这里需要实现自动重连。")

	close(mud.input)
}

func (mud *Server) telnetNegotiate(m IACMessage) {
	if m.Eq(WILL, OptZMP) {
		mud.conn.Write([]byte{IAC, DO, OptZMP})
		go func() {
			for {
				time.Sleep(10 * time.Second)
				mud.conn.Write([]byte{IAC, SB, OptZMP})
				mud.conn.Write([]byte("zmp.ping"))
				mud.conn.Write([]byte{0, IAC, SE})
			}
		}()
	} else if m.Eq(DO, OptTTYPE) {
		mud.conn.Write([]byte{IAC, WILL, OptTTYPE})
	} else if m.Eq(SB, OptTTYPE, 0x01) {
		mud.conn.Write(append([]byte{IAC, SB, OptTTYPE, 0x00}, []byte("GoMud")...))
		mud.conn.Write([]byte{IAC, SE})
	} else if m.Command == WILL {
		mud.conn.Write([]byte{IAC, DONT, m.Args[0]})
	} else if m.Command == DO {
		mud.conn.Write([]byte{IAC, WONT, m.Args[0]})
	} else if m.Command == GA {
		// FIXME: 接收到 GA 后，应当强制完成当前的不完整的行。
		// TODO: 更进一步地，应当在 GA 收到前，阻止用户发送命令。
		//       为了不影响用户体验，可以允许输入，但不允许回车发送，等到收到 GA 后再发送。
		// TODO: 此功能应当仅当 GA 可用时打开，且允许用户通过配置文件关闭。
	}
	// TODO: IAC 不继续传递给 UI
	if mud.config.IACDebug {
		mud.input <- m.String()
	}
}

func (mud *Server) Stop() {
	if mud.conn != nil {
		mud.conn.Close()
	}
}

func (mud *Server) Input() <-chan string {
	return mud.input
}
