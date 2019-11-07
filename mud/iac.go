package mud

import (
	"bytes"
	"fmt"
	"log"
)

// IANA 管理的 Telnet 选项分配情况：
// https://www.iana.org/assignments/telnet-options/telnet-options.xhtml

// Standard Telnet Commands ，参见 https://tools.ietf.org/html/rfc854
const (
	IAC   = 255 // 0xFF	Interpret as Command
	DONT  = 254 // 0xFE	Don't do something
	DO    = 253 // 0xFD	Do something
	WONT  = 252 // 0xFC	Won't do something
	WILL  = 251 // 0xFB	Will do something
	SB    = 250 // 0xFA	Subnegotiation Begin
	GA    = 249 // 0xF9	Go Ahead
	EL    = 248 // 0xF8	Erase Line
	EC    = 247 // 0xF7	Erase Character
	AYT   = 246 // 0xF6	Are You Here?
	AO    = 245 // 0xF5	Abort Output
	IP    = 244 // 0xF4	Interrupt Process
	BREAK = 243 // 0xF3	NVT character BRK
	DM    = 242 // 0xF2	Data Mark
	NOP   = 241 // 0xF1	No Operation
	SE    = 240 // 0xF0	Subnegotiation End
)

// Standard Telnet Commands: https://tools.ietf.org/html/rfc854
const (
	EOR   = 239 // 0xEF	End Of Record
	O_EOR = 25  // 0x19	Negotiate About EOR
)

// Telnet Linemode Option: https://tools.ietf.org/html/rfc1116
const (
	LMABORT = 238 // 0xEE	(Linemode) Abort
	LMSUSP  = 237 // 0xED	(Linemode) Suspend
	LMEOF   = 236 // 0xEC	(Linemode) End Of File

	O_LINEMODE = 34 // 0x22	Linemode Option
)

const (
	O_BINARY    = 0  // 0x00	[RFC856]  Binary Transmission
	O_ECHO      = 1  // 0x01	[RFC857]  Echo
	O_RCP       = 2  // 0x02	[NIC5005] Telnet Reconnection Option
	O_NAOL      = 8  // 0x08	[NIC5005] Negotiate About Output Line Width
	O_NAOP      = 9  // 0x09	[NIC5005] Negotiate About Output Page Size
	O_SGA       = 3  // 0x03	[RFC858]  Suppress GA (Go Ahead)
	O_NAMS      = 4  // 0x04	[      ]  Negotiate About Message Size
	O_STATUS    = 5  // 0x05	[RFC859]  Status
	O_TM        = 6  // 0x06	[RFC860]  Timing Mark
	O_RCTE      = 7  // 0x07	[RFC726]  Remote Controlled Transmssion and Echoing
	O_NAOCRD    = 10 // 0x0A	[RFC652]  Negotiate About Output Carriage-Return Disposition
	O_NAOHTS    = 11 // 0x0B	[RFC653]  Negotiate About Output Horizontal Tab Stops
	O_NAOHTD    = 12 // 0x0C	[RFC654]  Negotiate About Output Horizontal Tab Disposition
	O_NAOFFD    = 13 // 0x0D	[RFC655]  Negotiate About Output Formfeed Disposition
	O_NAOVTS    = 14 // 0x0E	[RFC656]  Negotiate About Output Vertical Tabstops
	O_NAOVTD    = 15 // 0x0F	[RFC657]  Negotiate About Output Vertical Tab Disposition
	O_NAOLFD    = 16 // 0x10	[RFC658]  Negotiate About Output Linefeed Disposition
	O_XASCII    = 17 // 0x11	[RFC698]  Extended ASCII
	O_LOGOUT    = 18 // 0x12	[RFC727]  Logout
	O_BM        = 19 // 0x13	[RFC735]  Byte Macro
	O_DET       = 20 // 0x14	[RFC1043] Data Entry Terminal
	O_SUPDUP    = 21 // 0x15	[RFC736]  SUPDUP Display Protocol
	O_SUPDUPOUT = 22 // 0x16	[RFC749]  SUPDUP OUTPUT
	O_SNDLOC    = 23 // 0x17	[RFC779]  Send Location
	O_TTYPE     = 24 // 0x18	[RFC1091] Terminal Type
	O_TUID      = 26 // 0x1A	[RFC927]  TACACS User Identification
	O_OUTMRK    = 27 // 0x1B	[RFC933]  Output Marking
	O_TTYLOC    = 28 // 0x1C	[RFC946]  Terminal Location Number
	O_3270      = 29 // 0x1D	[RFC1041] Telnet 3270 Regime
	O_X3PAD     = 30 // 0x1E	[RFC1053] X.3 PAD
	O_NAWS      = 31 // 0x1F	[RFC1073] Negotiate About Window Size
	O_TSPEED    = 32 // 0x20	[RFC1079] Terminal Speed
	O_LFLOW     = 33 // 0x21	[RFC1372] Remote Flow Control
	O_XDISPLOC  = 35 // 0x23	[RFC1096] X Display Location
	O_ENVIRON   = 36 // 0x24	[RFC1408] Environment Option
	O_AUTH      = 37 // 0x25	[RFC2941] Authentication Option
	O_ENCRYPT   = 38 // 0x26	[RFC2946] Encryption Option
	O_NENV      = 39 // 0x27	[RFC1572] New Environment
	O_TN3270E   = 40 // 0x28	[RFC2355] TN3270 Enhancements
	O_XAUTH     = 41 // 0x29
	O_CHARSET   = 42 // 0x30	[RFC2066] Charset Option
	O_COMPORT   = 44 // 0x32	[RFC2217] Com Port Control Option
	O_KERMIT    = 47 // 0x35	[RFC2840] KERMIT Option

	O_MSSP  = 70  // 0x46	MUD Server Status Protocol
	O_MCCP  = 85  // 0x55	MUD Client Compression Protocol
	O_MCCP2 = 86  // 0x56	MUD Client Compression Protocol 2.0
	O_MXP   = 91  // 0x5B	MUD eXtension Protocol
	O_ZMP   = 93  // 0x5D	Zenith MUD Protocol
	O_GMCP  = 201 // 0xC9	Generic MUD Communication Protocol
	O_EXOPL = 255 // 0xFF	Extended Options List
)

var codeName = map[byte]string{
	IAC:         "IAC",
	DONT:        "DONT",
	DO:          "DO",
	WONT:        "WONT",
	WILL:        "WILL",
	SB:          "SB",
	GA:          "GA",
	EL:          "EL",
	EC:          "EC",
	AYT:         "AYT",
	AO:          "AO",
	IP:          "IP",
	BREAK:       "BREAK",
	DM:          "DM",
	NOP:         "NOP",
	SE:          "SE",
	EOR:         "EOR",
	LMABORT:     "ABORT",
	LMSUSP:      "SUSP",
	LMEOF:       "EOF",
	O_BINARY:    "BINARY",
	O_ECHO:      "ECHO",
	O_RCP:       "RCP",
	O_SGA:       "SGA",
	O_NAMS:      "NAMS",
	O_STATUS:    "STATUS",
	O_TM:        "TM",
	O_RCTE:      "RCTE",
	O_NAOL:      "NAOL",
	O_NAOP:      "NAOP",
	O_NAOCRD:    "NAOCRD",
	O_NAOHTS:    "NAOHTS",
	O_NAOHTD:    "NAOHTD",
	O_NAOFFD:    "NAOFFD",
	O_NAOVTS:    "NAOVTS",
	O_NAOVTD:    "NAOVTD",
	O_NAOLFD:    "NAOLFD",
	O_XASCII:    "XASCII",
	O_LOGOUT:    "LOGOUT",
	O_BM:        "BM",
	O_DET:       "DET",
	O_SUPDUP:    "SUP",
	O_SUPDUPOUT: "SUPOUT",
	O_SNDLOC:    "SNDLOC",
	O_TTYPE:     "TTYPE",
	O_EOR:       "EOR",
	O_TUID:      "TUID",
	O_OUTMRK:    "OUTMRK",
	O_TTYLOC:    "TTYLOC",
	O_3270:      "3270",
	O_X3PAD:     "X3PAD",
	O_NAWS:      "NAWS",
	O_TSPEED:    "TSPEED",
	O_LFLOW:     "LFLOW",
	O_LINEMODE:  "LINEMODE",
	O_XDISPLOC:  "XDISPLOC",
	O_ENVIRON:   "ENVIRON",
	O_AUTH:      "AUTH",
	O_ENCRYPT:   "ENCRYPT",
	O_NENV:      "NENV",
	O_TN3270E:   "TN3270E",
	O_XAUTH:     "XAUTH",
	O_CHARSET:   "CHARSET",
	O_COMPORT:   "COMPORT",
	O_KERMIT:    "KERMIT",
	O_MSSP:      "MSSP",
	O_MCCP:      "MCCP",
	O_MCCP2:     "MCCP2",
	O_MXP:       "MXP",
	O_ZMP:       "ZMP",
	O_GMCP:      "GMCP",
}

type iacStage int

const (
	stCmd iacStage = iota
	stArg
	stSuboption
	stDone
)

type IACMessage struct {
	state   iacStage
	Command byte
	Args    []byte
}

func NewIACMessage() *IACMessage {
	iac := &IACMessage{}
	iac.Reset()
	return iac
}

func (IACMessage) IsMessage() {}

func (iac *IACMessage) Reset() {
	iac.state = stCmd
	iac.Args = make([]byte, 0, 128)
}

func (iac IACMessage) String() string {
	cmdName := codeName[iac.Command]
	if cmdName == "" {
		cmdName = fmt.Sprintf("%d", iac.Command)
	}
	argName := fmt.Sprintf("%v", iac.Args[:len(iac.Args)])
	if (iac.Command == WILL ||
		iac.Command == WONT ||
		iac.Command == DO ||
		iac.Command == DONT ||
		iac.Command == SB) &&
		codeName[iac.Args[0]] != "" {
		argName = codeName[iac.Args[0]]
	}
	return fmt.Sprintf("IAC %s %s", cmdName, argName)
}

func (iac IACMessage) Eq(command byte, args ...byte) bool {
	if iac.Command != command {
		return false
	}

	return bytes.Equal(iac.Args, args)
}

func (iac *IACMessage) Scan(b byte) bool {
	switch iac.state {
	case stCmd:
		switch b {
		case WILL, WONT, DO, DONT:
			iac.Command = b
			iac.state = stArg
			return false
		case SB:
			iac.Command = SB
			iac.state = stSuboption
			return false
		case SE:
			iac.Command = SE
			iac.state = stDone
			return true
		case GA:
			iac.Command = GA
			iac.state = stDone
			return true
		default:
			// TODO: 在这里处理所有的 IAC 指令
			log.Printf("----未知指令: IAC %d", b)
			iac.state = stDone
			return true
		}
	case stArg:
		iac.Args = append(iac.Args, b)
		iac.state = stDone
		return true
	case stSuboption:
		iac.Args = append(iac.Args, b)
		return false
	default:
		iac.state = stDone
		log.Printf("未知指令: IAC %d", b)
		return true
	}
}
