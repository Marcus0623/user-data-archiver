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
	title, err := syscall.UTF16PtrFromString("User Data Archiver")
	if err == nil {
		_, _, _ = setTitle.Call(uintptr(unsafe.Pointer(title)))
	}
}

func moveFileReplace(from, to string) error {
	fromP, err := syscall.UTF16PtrFromString(longPath(from))
	if err != nil {
		return err
	}
	toP, err := syscall.UTF16PtrFromString(longPath(to))
	if err != nil {
		return err
	}
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	proc := kernel32.NewProc("MoveFileExW")
	const (
		moveFileReplaceExisting = 0x1
		moveFileWriteThrough    = 0x8
	)
	r, _, callErr := proc.Call(
		uintptr(unsafe.Pointer(fromP)),
		uintptr(unsafe.Pointer(toP)),
		uintptr(moveFileReplaceExisting|moveFileWriteThrough),
	)
	if r != 0 {
		return nil
	}
	if callErr != nil && callErr != syscall.Errno(0) {
		return callErr
	}
	return fmt.Errorf("MoveFileEx failed")
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
		return 0, 0, fmt.Errorf("GetDiskFreeSpaceEx failed")
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

func hideConsoleWindow() {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getConsole := kernel32.NewProc("GetConsoleWindow")
	hwnd, _, _ := getConsole.Call()
	if hwnd == 0 {
		return
	}
	user32 := syscall.NewLazyDLL("user32.dll")
	show := user32.NewProc("ShowWindow")
	_, _, _ = show.Call(hwnd, 0)
}

func consoleProcessCount() int {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	proc := kernel32.NewProc("GetConsoleProcessList")
	var pids [8]uint32
	n, _, _ := proc.Call(uintptr(unsafe.Pointer(&pids[0])), uintptr(len(pids)))
	return int(n)
}

func computerName() string {
	name, err := os.Hostname()
	if err != nil {
		return "unknown"
	}
	return name
}

func driveKind(root string) string {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	proc := kernel32.NewProc("GetDriveTypeW")
	p, err := syscall.UTF16PtrFromString(root)
	if err != nil {
		return "unknown"
	}
	n, _, _ := proc.Call(uintptr(unsafe.Pointer(p)))
	switch n {
	case 1:
		return "missing"
	case 2:
		return "removable"
	case 3:
		return "local disk"
	case 4:
		return "network"
	case 5:
		return "cd-rom"
	case 6:
		return "ram disk"
	default:
		return "unknown"
	}
}

func volumeLabel(root string) string {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	proc := kernel32.NewProc("GetVolumeInformationW")
	p, err := syscall.UTF16PtrFromString(root)
	if err != nil {
		return ""
	}
	var label [261]uint16
	var serial, maxComp, flags uint32
	var fsName [261]uint16
	r, _, _ := proc.Call(
		uintptr(unsafe.Pointer(p)),
		uintptr(unsafe.Pointer(&label[0])),
		uintptr(len(label)),
		uintptr(unsafe.Pointer(&serial)),
		uintptr(unsafe.Pointer(&maxComp)),
		uintptr(unsafe.Pointer(&flags)),
		uintptr(unsafe.Pointer(&fsName[0])),
		uintptr(len(fsName)),
	)
	if r == 0 {
		return ""
	}
	return syscall.UTF16ToString(label[:])
}
