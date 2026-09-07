package main

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestMapDestDriveLayout(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows 路径布局")
	}
	archive := `D:\离职归档\张三`
	got, err := MapDest(`C:\Downloads\报告.pdf`, archive)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(archive, `C`, `Downloads`, `报告.pdf`)
	if got != want {
		t.Fatalf("got %s want %s", got, want)
	}
}

func TestMapDestUsersProfile(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows 路径布局")
	}
	archive := `E:\handoff\lisi`
	got, err := MapDest(`C:\Users\lisi\Documents\合同.pdf`, archive)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(archive, `C`, `Users`, `lisi`, `Documents`, `合同.pdf`)
	if got != want {
		t.Fatalf("got %s want %s", got, want)
	}
}

func TestSanitizeEmployeeName(t *testing.T) {
	if got := sanitizeEmployeeName(`张三:离职`); got != `张三_离职` {
		t.Fatalf("got %q", got)
	}
	if got := sanitizeEmployeeName(`  `); got != "" {
		t.Fatalf("got %q", got)
	}
}

func TestExpandRootDriveLetter(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows 盘符")
	}
	got, err := expandRoot("C:")
	if err != nil {
		t.Fatal(err)
	}
	if got != `C:\` {
		t.Fatalf("C: 应展开为 C:\\，得到 %s（不能展开成当前工作目录）", got)
	}
}

func TestIsBaseOf(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip()
	}
	base := `D:\离职归档\张三`
	if !isBaseOf(base, filepath.Join(base, `C`, `Users`)) {
		t.Fatal("子目录应判定为位于目标内")
	}
	if isBaseOf(base, `D:\离职归档`) {
		t.Fatal("父目录不应判定为位于目标内")
	}
}
