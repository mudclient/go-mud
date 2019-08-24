package mud

import (
	"fmt"
	"io"
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

type MudConfig struct {
	IACDebug bool
	Host     string `flag:"H|mud.pkuxkx.net|服务器 {IP/Domain}"`
	Port     int    `flag:"P|8080|服务器 {Port}"`
}

type MudServer struct {
	printer.SimplePrinter

	config MudConfig

	screen printer.Printer
	server printer.WritePrinter

	conn  net.Conn
	input chan string
}

func NewMudServer(config MudConfig) *MudServer {
	mud := &MudServer{
		config: config,
		screen: printer.NewSimplePrinter(os.Stdout),
		server: printer.NewSimplePrinter(ioutil.Discard),
		input:  make(chan string, 1024),
	}

	mud.SetOutput(mud.server)

	return mud
}

func (mud *MudServer) SetScreen(w io.Writer) {
	mud.screen.SetOutput(w)
}

func (mud *MudServer) Run() {
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

	mud.Println("连接成功。")

	netWriter := encoder.NewWriter(mud.conn)
	mud.server.SetOutput(netWriter)

	scanner := NewScanner(mud.conn)

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

func (mud *MudServer) telnetNegotiate(m IACMessage) {
	if m.Eq(WILL, ZMP) {
		mud.conn.Write([]byte{IAC, DO, ZMP})
		go func() {
			for {
				time.Sleep(10 * time.Second)
				mud.conn.Write([]byte{IAC, SB, ZMP})
				mud.conn.Write([]byte("zmp.ping"))
				mud.conn.Write([]byte{0, IAC, SE})
			}
		}()
	} else if m.Eq(DO, TTYPE) {
		mud.conn.Write([]byte{IAC, WILL, TTYPE})
	} else if m.Eq(SB, TTYPE, 0x01) {
		mud.conn.Write(append([]byte{IAC, SB, TTYPE, 0x00}, []byte("GoMud")...))
		mud.conn.Write([]byte{IAC, SE})
	} else if m.Command == WILL {
		mud.conn.Write([]byte{IAC, DONT, m.Args[0]})
	} else if m.Command == DO {
		mud.conn.Write([]byte{IAC, WONT, m.Args[0]})
	} else if m.Command == GA {
		mud.input <- "IAC GA"
	}
	// TODO: IAC 不继续传递给 UI
	if mud.config.IACDebug {
		mud.input <- m.String()
	}
}

func (mud *MudServer) Stop() {
	if mud.conn != nil {
		mud.conn.Close()
	}
}

func (mud *MudServer) Input() <-chan string {
	return mud.input
}
