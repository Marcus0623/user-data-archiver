package main

import (
	"fmt"
	"os"
	"strings"
	"syscall"
	"unsafe"
)

func setupConsole() {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	setOut := kernel32.NewProc("SetConsoleOutputCP")
	setIn := kernel32.NewProc("SetConsoleCP")
	_, _, _ = setOut.Call(65001)
	_, _, _ = setIn.Call(65001)
	setTitle := kernel32.NewProc("SetConsoleTitleW")
	title, err := syscall.UTF16PtrFromString("离职资料归档工具")
	if err == nil {
		_, _, _ = setTitle.Call(uintptr(unsafe.Pointer(title)))
	}
}

func longPath(p string) string {
	p = normalizeAbs(p)
	if strings.HasPrefix(p, `\\?\`) {
		return p
	}
	if len(p) >= 2 && p[0] == '\\' && p[1] == '\\' {
		return `\\?\UNC\` + p[2:]
	}
	return `\\?\` + p
}

func diskFreeBytes(path string) (free, total uint64, err error) {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	proc := kernel32.NewProc("GetDiskFreeSpaceExW")
	p, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return 0, 0, err
	}
	var totalFree uint64
	r1, _, callErr := proc.Call(
		uintptr(unsafe.Pointer(p)),
		uintptr(unsafe.Pointer(&free)),
		uintptr(unsafe.Pointer(&total)),
		uintptr(unsafe.Pointer(&totalFree)),
	)
	if r1 == 0 {
		if callErr != nil {
			return 0, 0, callErr
		}
		return 0, 0, fmt.Errorf("GetDiskFreeSpaceEx 失败")
	}
	return free, total, nil
}

func getLongPathName(p string) string {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	proc := kernel32.NewProc("GetLongPathNameW")
	src, err := syscall.UTF16PtrFromString(p)
	if err != nil {
		return p
	}
	buf := make([]uint16, 32768)
	n, _, _ := proc.Call(uintptr(unsafe.Pointer(src)), uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	if n == 0 {
		return p
	}
	return syscall.UTF16ToString(buf[:n])
}

func isAdmin() bool {
	shell32 := syscall.NewLazyDLL("shell32.dll")
	proc := shell32.NewProc("IsUserAnAdmin")
	r, _, _ := proc.Call()
	return r != 0
}

func computerName() string {
	name, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return name
}
