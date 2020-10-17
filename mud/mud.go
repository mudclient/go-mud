package mud

import (
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"strings"
	"time"

	"github.com/flw-cn/printer"
	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/transform"
)

type Config struct {
	IACDebug bool
	Host     string `flag:"H|mud.pkuxkx.net|服务器 {IP/Domain}"`
	Port     int    `flag:"P|8080|服务器 {Port}"`
	Encoding string `flag:"|UTF-8|服务器输出文本的 {Encoding}"`
}

type Server struct {
	printer.SimplePrinter

	config Config

	screen printer.Printer
	server printer.WritePrinter

	conn  net.Conn
	input chan string

	decoder *encoding.Decoder
	encoder *encoding.Encoder
}

func NewServer(config Config) *Server {
	mud := &Server{
		config: config,
		screen: printer.NewSimplePrinter(os.Stdout),
		server: printer.NewSimplePrinter(ioutil.Discard),
		input:  make(chan string, 1024),
	}

	mud.decoder = resolveEncoding(config.Encoding).NewDecoder()
	mud.encoder = resolveEncoding(config.Encoding).NewEncoder()

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

	netWriter := transform.NewWriter(mud.conn, mud.encoder)
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
			r := transform.NewReader(m, mud.decoder)
			buf, _ := ioutil.ReadAll(r)
			mud.input <- string(buf)
		case Line:
			r := transform.NewReader(m, mud.decoder)
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
	switch {
	case m.Eq(WILL, OptZMP):
		mud.conn.Write([]byte{IAC, DO, OptZMP})
		go func() {
			for {
				time.Sleep(10 * time.Second)
				mud.conn.Write([]byte{IAC, SB, OptZMP})
				mud.conn.Write([]byte("zmp.ping"))
				mud.conn.Write([]byte{0, IAC, SE})
			}
		}()
	case m.Eq(DO, OptTTYPE):
		mud.conn.Write([]byte{IAC, WILL, OptTTYPE})
	case m.Eq(SB, OptTTYPE, 0x01):
		mud.conn.Write(append([]byte{IAC, SB, OptTTYPE, 0x00}, []byte("GoMud")...))
		mud.conn.Write([]byte{IAC, SE})
	case m.Eq(WILL):
		mud.conn.Write([]byte{IAC, DONT, m.Args[0]})
	case m.Eq(DO):
		mud.conn.Write([]byte{IAC, WONT, m.Args[0]})
	case m.Eq(GA):
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

func resolveEncoding(e string) encoding.Encoding {
	e = strings.ToUpper(e)
	switch e {
	case "GB2312", "HZ-GB-2312", "HZGB2312", "EUC-CN", "EUCCN":
		return simplifiedchinese.HZGB2312
	case "GBK", "CP936":
		return simplifiedchinese.GBK
	case "GB18030":
		return simplifiedchinese.GB18030
	case "BIG5", "BIG-5", "BIG-FIVE":
		return traditionalchinese.Big5
	case "UTF8", "UTF-8":
		return encoding.Nop
	}

	return encoding.Nop
}
