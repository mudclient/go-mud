package mud

import (
	"bytes"
	"fmt"
	"log"
)

const IAC = 255  // 0xFF	Interpret as Command
const WILL = 251 // 0xFB	Will do something
const WONT = 252 // 0xFC	Won't do something
const DO = 253   // 0xFD	Do something
const DONT = 254 // 0xFE	Don't do something
const SB = 250   // 0xFA	Subnegotiation Begin
const SE = 240   // 0xF0	Subnegotiation End
const GA = 249   // 0xF9	Go Ahead

const EL = 248   // 0xF8	Erase Line
const EC = 247   // 0xF7	Erase Character
const AYT = 246  // 0xF6	Are You Here?
const TTYPE = 24 // 0x18	Terminal Type
const NAWS = 31  // 0x1F	Negotiate About Window Size
const NENV = 39  // 0x27	New Environment
const MXP = 91   // 0x5B	MUD eXtension Protocol
const MSSP = 70  // 0x46	MUD Server Status Protocol
const ZMP = 93   // 0x5D	Zenith MUD Protocol
const GMCP = 201 // 0xC9	Generic MUD Communication Protocol
const NOP = 241  // 0xF1	No operation
const ECHO = 1   // 0x01	Echo

var codeName = map[byte]string{
	IAC:   "IAC",
	WILL:  "WILL",
	WONT:  "WONT",
	DO:    "DO",
	DONT:  "DONT",
	SB:    "SB",
	SE:    "SE",
	GA:    "GA",
	EL:    "EL",
	EC:    "EC",
	AYT:   "AYT",
	TTYPE: "TTYPE",
	NAWS:  "NAWS",
	NENV:  "NEW-ENV",
	MXP:   "MXP",
	MSSP:  "MSSP",
	ZMP:   "ZMP",
	GMCP:  "GMCP",
	NOP:   "NOP",
	ECHO:  "ECHO",
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
