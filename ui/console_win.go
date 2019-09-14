// +build windows

package ui

import (
	"syscall"
	"unsafe"
)

var k32 = syscall.NewLazyDLL("kernel32.dll")
var u32 = syscall.NewLazyDLL("user32.dll")

var (
	consoleStdout                  syscall.Handle
	procGetConsoleScreenBufferInfo = k32.NewProc("GetConsoleScreenBufferInfo")
	procSetConsoleScreenBufferSize = k32.NewProc("SetConsoleScreenBufferSize")
	procSetConsoleWindowInfo       = k32.NewProc("SetConsoleWindowInfo")
	procGetConsoleWindow           = k32.NewProc("GetConsoleWindow")
	procShowWindow                 = u32.NewProc("ShowWindow")
	procSetWindowPos               = u32.NewProc("SetWindowPos")
	procSetConsoleTitle            = k32.NewProc("SetConsoleTitleW")
)

const (
	SW_MAXIMIZE = 3
	SW_RESTORE  = 9
)

func InitConsole(title string) {
	maximizeConsole()
	setConsoleTitle(title)
}

type coord struct {
	x int16
	y int16
}

func (c coord) uintptr() uintptr {
	// little endian, put x first
	return uintptr(c.x) | (uintptr(c.y) << 16)
}

type rect struct {
	left   int16
	top    int16
	right  int16
	bottom int16
}

type consoleInfo struct {
	size  coord
	pos   coord
	attrs uint16
	win   rect
	maxsz coord
}

func getConsoleInfo(info *consoleInfo) {
	procGetConsoleScreenBufferInfo.Call(
		uintptr(consoleStdout),
		uintptr(unsafe.Pointer(info)))
}

func setBufferSize(x, y int) {
	procSetConsoleScreenBufferSize.Call(
		uintptr(consoleStdout),
		coord{int16(x), int16(y)}.uintptr())
}

func setConsoleInfo(rc *rect) {
	procSetConsoleWindowInfo.Call(
		uintptr(consoleStdout),
		uintptr(1),
		uintptr(unsafe.Pointer(rc)))
}

func getConsoleWindow() uintptr {
	consoleWindow, _, _ := procGetConsoleWindow.Call()
	return consoleWindow
}

func setConsoleTitle(title string) {
	procSetConsoleTitle.Call(uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(title))))
}

func maximizeConsole() {
	consoleStdout, _ = syscall.Open("CONOUT$", syscall.O_RDWR, 0)
	info := consoleInfo{}
	getConsoleInfo(&info)
	setBufferSize(10000, 1000)
	win := getConsoleWindow()
	procShowWindow.Call(uintptr(win), uintptr(SW_MAXIMIZE))
	getConsoleInfo(&info)
	procShowWindow.Call(uintptr(win), uintptr(SW_RESTORE))
	rc := rect{top: 0, left: 0, right: info.win.right + 1, bottom: info.win.bottom + 1}
	setConsoleInfo(&rc)
	procSetWindowPos.Call(
		uintptr(win),
		uintptr(0), // HWND_TOP,
		uintptr(1), // x,
		uintptr(1), // y,
		uintptr(0), // cx (ignore by SWP_NOSIZE),
		uintptr(0), // cy (ignore by SWP_NOSIZE),
		uintptr(1), // SWP_NOSIZE
	)
}
