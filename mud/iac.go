package mud

import (
	"bytes"
	"fmt"
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
	EOR    = 239 // 0xEF	End Of Record
	OptEOR = 25  // 0x19	Negotiate About EOR
)

// Telnet Linemode Option: https://tools.ietf.org/html/rfc1116
const (
	LMABORT = 238 // 0xEE	(Linemode) Abort
	LMSUSP  = 237 // 0xED	(Linemode) Suspend
	LMEOF   = 236 // 0xEC	(Linemode) End Of File

	OptLINEMODE = 34 // 0x22	Linemode Option
)

const (
	OptBINARY    = 0  // 0x00	[RFC856]  Binary Transmission
	OptECHO      = 1  // 0x01	[RFC857]  Echo
	OptRCP       = 2  // 0x02	[NIC5005] Telnet Reconnection Option
	OptNAOL      = 8  // 0x08	[NIC5005] Negotiate About Output Line Width
	OptNAOP      = 9  // 0x09	[NIC5005] Negotiate About Output Page Size
	OptSGA       = 3  // 0x03	[RFC858]  Suppress GA (Go Ahead)
	OptNAMS      = 4  // 0x04	[      ]  Negotiate About Message Size
	OptSTATUS    = 5  // 0x05	[RFC859]  Status
	OptTM        = 6  // 0x06	[RFC860]  Timing Mark
	OptRCTE      = 7  // 0x07	[RFC726]  Remote Controlled Transmssion and Echoing
	OptNAOCRD    = 10 // 0x0A	[RFC652]  Negotiate About Output Carriage-Return Disposition
	OptNAOHTS    = 11 // 0x0B	[RFC653]  Negotiate About Output Horizontal Tab Stops
	OptNAOHTD    = 12 // 0x0C	[RFC654]  Negotiate About Output Horizontal Tab Disposition
	OptNAOFFD    = 13 // 0x0D	[RFC655]  Negotiate About Output Formfeed Disposition
	OptNAOVTS    = 14 // 0x0E	[RFC656]  Negotiate About Output Vertical Tabstops
	OptNAOVTD    = 15 // 0x0F	[RFC657]  Negotiate About Output Vertical Tab Disposition
	OptNAOLFD    = 16 // 0x10	[RFC658]  Negotiate About Output Linefeed Disposition
	OptXASCII    = 17 // 0x11	[RFC698]  Extended ASCII
	OptLOGOUT    = 18 // 0x12	[RFC727]  Logout
	OptBM        = 19 // 0x13	[RFC735]  Byte Macro
	OptDET       = 20 // 0x14	[RFC1043] Data Entry Terminal
	OptSUPDUP    = 21 // 0x15	[RFC736]  SUPDUP Display Protocol
	OptSUPDUPOUT = 22 // 0x16	[RFC749]  SUPDUP OUTPUT
	OptSNDLOC    = 23 // 0x17	[RFC779]  Send Location
	OptTTYPE     = 24 // 0x18	[RFC1091] Terminal Type
	OptTUID      = 26 // 0x1A	[RFC927]  TACACS User Identification
	OptOUTMRK    = 27 // 0x1B	[RFC933]  Output Marking
	OptTTYLOC    = 28 // 0x1C	[RFC946]  Terminal Location Number
	Opt3270      = 29 // 0x1D	[RFC1041] Telnet 3270 Regime
	OptX3PAD     = 30 // 0x1E	[RFC1053] X.3 PAD
	OptNAWS      = 31 // 0x1F	[RFC1073] Negotiate About Window Size
	OptTSPEED    = 32 // 0x20	[RFC1079] Terminal Speed
	OptLFLOW     = 33 // 0x21	[RFC1372] Remote Flow Control
	OptXDISPLOC  = 35 // 0x23	[RFC1096] X Display Location
	OptENVIRON   = 36 // 0x24	[RFC1408] Environment Option
	OptAUTH      = 37 // 0x25	[RFC2941] Authentication Option
	OptENCRYPT   = 38 // 0x26	[RFC2946] Encryption Option
	OptNENV      = 39 // 0x27	[RFC1572] New Environment
	OptTN3270E   = 40 // 0x28	[RFC2355] TN3270 Enhancements
	OptXAUTH     = 41 // 0x29
	OptCHARSET   = 42 // 0x30	[RFC2066] Charset Option
	OptCOMPORT   = 44 // 0x32	[RFC2217] Com Port Control Option
	OptKERMIT    = 47 // 0x35	[RFC2840] KERMIT Option

	OptMSSP  = 70  // 0x46	MUD Server Status Protocol
	OptMCCP  = 85  // 0x55	MUD Client Compression Protocol
	OptMCCP2 = 86  // 0x56	MUD Client Compression Protocol 2.0
	OptMXP   = 91  // 0x5B	MUD eXtension Protocol
	OptZMP   = 93  // 0x5D	Zenith MUD Protocol
	OptGMCP  = 201 // 0xC9	Generic MUD Communication Protocol
	OptEXOPL = 255 // 0xFF	Extended Options List
)

var codeName = map[byte]string{
	IAC:          "IAC",
	DONT:         "DONT",
	DO:           "DO",
	WONT:         "WONT",
	WILL:         "WILL",
	SB:           "SB",
	GA:           "GA",
	EL:           "EL",
	EC:           "EC",
	AYT:          "AYT",
	AO:           "AO",
	IP:           "IP",
	BREAK:        "BREAK",
	DM:           "DM",
	NOP:          "NOP",
	SE:           "SE",
	EOR:          "EOR",
	LMABORT:      "ABORT",
	LMSUSP:       "SUSP",
	LMEOF:        "EOF",
	OptBINARY:    "BINARY",
	OptECHO:      "ECHO",
	OptRCP:       "RCP",
	OptSGA:       "SGA",
	OptNAMS:      "NAMS",
	OptSTATUS:    "STATUS",
	OptTM:        "TM",
	OptRCTE:      "RCTE",
	OptNAOL:      "NAOL",
	OptNAOP:      "NAOP",
	OptNAOCRD:    "NAOCRD",
	OptNAOHTS:    "NAOHTS",
	OptNAOHTD:    "NAOHTD",
	OptNAOFFD:    "NAOFFD",
	OptNAOVTS:    "NAOVTS",
	OptNAOVTD:    "NAOVTD",
	OptNAOLFD:    "NAOLFD",
	OptXASCII:    "XASCII",
	OptLOGOUT:    "LOGOUT",
	OptBM:        "BM",
	OptDET:       "DET",
	OptSUPDUP:    "SUP",
	OptSUPDUPOUT: "SUPOUT",
	OptSNDLOC:    "SNDLOC",
	OptTTYPE:     "TTYPE",
	OptEOR:       "EOR",
	OptTUID:      "TUID",
	OptOUTMRK:    "OUTMRK",
	OptTTYLOC:    "TTYLOC",
	Opt3270:      "3270",
	OptX3PAD:     "X3PAD",
	OptNAWS:      "NAWS",
	OptTSPEED:    "TSPEED",
	OptLFLOW:     "LFLOW",
	OptLINEMODE:  "LINEMODE",
	OptXDISPLOC:  "XDISPLOC",
	OptENVIRON:   "ENVIRON",
	OptAUTH:      "AUTH",
	OptENCRYPT:   "ENCRYPT",
	OptNENV:      "NENV",
	OptTN3270E:   "TN3270E",
	OptXAUTH:     "XAUTH",
	OptCHARSET:   "CHARSET",
	OptCOMPORT:   "COMPORT",
	OptKERMIT:    "KERMIT",
	OptMSSP:      "MSSP",
	OptMCCP:      "MCCP",
	OptMCCP2:     "MCCP2",
	OptMXP:       "MXP",
	OptZMP:       "ZMP",
	OptGMCP:      "GMCP",
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

func (iac *IACMessage) Scan(b byte) (completed bool) {
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
			// TODO: 需要处理未知 IAC 指令
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
		// TODO: 需要处理未知 IAC 指令
		return true
	}
}
