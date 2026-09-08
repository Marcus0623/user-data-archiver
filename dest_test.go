package main

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestMapDestDriveLayout(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows path layout")
	}
	archive := `D:\offboarding-archive\alice`
	got, err := MapDest(`C:\Downloads\report.pdf`, archive)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(archive, `C`, `Downloads`, `report.pdf`)
	if got != want {
		t.Fatalf("got %s want %s", got, want)
	}
}

func TestMapDestUsersProfile(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows path layout")
	}
	archive := `E:\handoff\lisi`
	got, err := MapDest(`C:\Users\lisi\Documents\contract.pdf`, archive)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(archive, `C`, `Users`, `lisi`, `Documents`, `contract.pdf`)
	if got != want {
		t.Fatalf("got %s want %s", got, want)
	}
}

func TestSanitizeEmployeeName(t *testing.T) {
	if got := sanitizeEmployeeName(`alice:offboard`); got != `alice_offboard` {
		t.Fatalf("got %q", got)
	}
	if got := sanitizeEmployeeName(`  `); got != "" {
		t.Fatalf("got %q", got)
	}
	if got := sanitizeEmployeeName("PRN"); got != "archive" {
		t.Fatalf("got %q", got)
	}
	if got := sanitizeEmployeeName("COM1"); got != "archive" {
		t.Fatalf("got %q", got)
	}
}

func TestExpandRootDriveLetter(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows drive letter")
	}
	got, err := expandRoot("C:")
	if err != nil {
		t.Fatal(err)
	}
	if got != `C:\` {
		t.Fatalf("C: should expand to C:\\, got %s (must not expand to the working directory)", got)
	}
}

func TestIsBaseOf(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip()
	}
	base := `D:\offboarding-archive\alice`
	if !isBaseOf(base, filepath.Join(base, `C`, `Users`)) {
		t.Fatal("child path should be inside destination")
	}
	if isBaseOf(base, `D:\offboarding-archive`) {
		t.Fatal("parent path should not be treated as inside destination")
	}
	if isBaseOf(base, `D:\offboarding-archive\alice2`) {
		t.Fatal("a similarly named folder must not be treated as inside destination")
	}
	if !samePath(base, base) {
		t.Fatal("samePath should match a path to itself")
	}
}

func TestIsBaseOfWhenDestDoesNotExistYet(t *testing.T) {
	disk := t.TempDir()
	dest := filepath.Join(disk, "offboarding-archive", "alice")
	if !isBaseOf(disk, dest) {
		t.Fatal("a destination folder inside the source must be detected before it is created")
	}
	if isBaseOf(disk, filepath.Join(disk+"2", "x")) {
		t.Fatal("a similarly named sibling folder must not count as inside the source")
	}
}

func TestIsVolumeRoot(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip()
	}
	if !isVolumeRoot(`C:\`) || !isVolumeRoot(`C:`) || !isVolumeRoot(`D:\`) {
		t.Fatal("drive roots should be detected")
	}
	if isVolumeRoot(`C:\Users`) || isVolumeRoot(`D:\offboarding-archive\alice`) {
		t.Fatal("a folder on a drive is not a drive root")
	}
}

func TestProbeDestRejectsDriveRoot(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip()
	}
	setLang(LangEN)
	err := probeDestWritable(`C:\`)
	if err == nil {
		t.Fatal("drive root must be rejected")
	}
	if !strings.Contains(err.Error(), "cannot be a drive root") {
		t.Fatalf("got %v", err)
	}
	if _, statErr := os.Stat(`C:\_write-test.tmp`); statErr == nil {
		t.Fatal("probe must not write _write-test.tmp to the drive root")
	}
}

func TestProbeDestWritable(t *testing.T) {
	dir := t.TempDir()
	if err := probeDestWritable(dir); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, writeProbeFileName)); !os.IsNotExist(err) {
		t.Fatalf("probe file should be removed afterwards: %v", err)
	}
}

func TestDestDriveInfoEmpty(t *testing.T) {
	setLang(LangEN)
	if got := destDriveInfo(""); got != "No destination selected." {
		t.Fatalf("got %q", got)
	}
}

func TestExplainIOErrorPermission(t *testing.T) {
	setLang(LangEN)
	if got := explainIOError(os.ErrPermission); got != "access denied" {
		t.Fatalf("got %q", got)
	}
}
