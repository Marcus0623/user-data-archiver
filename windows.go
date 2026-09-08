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
	title, err := syscall.UTF16PtrFromString(T("AppName"))
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

func showErrorDialog(msg string) {
	user32 := syscall.NewLazyDLL("user32.dll")
	proc := user32.NewProc("MessageBoxW")
	text, err1 := syscall.UTF16PtrFromString(msg)
	title, err2 := syscall.UTF16PtrFromString(T("AppName"))
	if err1 != nil || err2 != nil {
		return
	}
	const (
		mbIconError     = 0x10
		mbSetForeground = 0x10000
	)
	_, _, _ = proc.Call(0, uintptr(unsafe.Pointer(text)), uintptr(unsafe.Pointer(title)), mbIconError|mbSetForeground)
}

func attachParentConsole() {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	attach := kernel32.NewProc("AttachConsole")
	alloc := kernel32.NewProc("AllocConsole")
	const attachParentProcess = uintptr(^uint32(0))
	r, _, _ := attach.Call(attachParentProcess)
	if r == 0 {
		_, _, _ = alloc.Call()
	}
	bindStdioToConsole()
}

func bindStdioToConsole() {
	in, errIn := os.OpenFile("CONIN$", os.O_RDONLY, 0)
	if errIn == nil {
		os.Stdin = in
	}
	out, errOut := os.OpenFile("CONOUT$", os.O_WRONLY, 0)
	if errOut == nil {
		os.Stdout = out
		os.Stderr = out
	}
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
