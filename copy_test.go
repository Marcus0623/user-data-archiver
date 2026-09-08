package main

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestCopyFileAndResume(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "sub", "dst.txt")
	content := []byte("archive test payload")
	if err := os.WriteFile(src, content, 0o644); err != nil {
		t.Fatal(err)
	}
	mtime := time.Date(2024, 5, 1, 12, 0, 0, 0, time.Local)
	if err := os.Chtimes(src, mtime, mtime); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(src)
	if err != nil {
		t.Fatal(err)
	}
	if err := copyFileWithRetry(src, dst, info); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(content) {
		t.Fatalf("copied %q", got)
	}
	if !sameFile(info, dst) {
		t.Fatal("copied file should match source so resume can skip it")
	}
}

func TestFormatBytes(t *testing.T) {
	if formatBytes(500) != "500 B" {
		t.Fatalf("got %s", formatBytes(500))
	}
	if formatBytes(1024) != "1.00 KB" {
		t.Fatalf("got %s", formatBytes(1024))
	}
}

func TestArchivePreservesLayout(t *testing.T) {
	src := t.TempDir()
	sub := filepath.Join(src, "Downloads")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sub, "a.txt"), []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	dest := t.TempDir()
	opt := Options{Mode: ModeWithAppData, DestAbs: dest}
	if _, err := Archive(context.Background(), []string{src}, opt, dest, nil, nil); err != nil {
		t.Fatal(err)
	}
	mapped, err := MapDest(filepath.Join(sub, "a.txt"), dest)
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(mapped)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "hello" {
		t.Fatalf("got %q", got)
	}
	if _, err := os.Stat(filepath.Join(sub, "a.txt")); err != nil {
		t.Fatalf("copy must keep the source file: %v", err)
	}
}

func TestArchiveCutRemovesSource(t *testing.T) {
	src := t.TempDir()
	sub := filepath.Join(src, "Downloads")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	srcFile := filepath.Join(sub, "a.txt")
	if err := os.WriteFile(srcFile, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	dest := t.TempDir()
	opt := Options{Mode: ModeWithAppData, DestAbs: dest, Cut: true}
	if _, err := Archive(context.Background(), []string{src}, opt, dest, nil, nil); err != nil {
		t.Fatal(err)
	}
	mapped, err := MapDest(srcFile, dest)
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(mapped)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "hello" {
		t.Fatalf("got %q", got)
	}
	if _, err := os.Stat(srcFile); !os.IsNotExist(err) {
		t.Fatalf("source file should be removed after cut: %v", err)
	}
}

func TestMayRemoveEmptiedDir(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows paths")
	}
	if mayRemoveEmptiedDir(`C:\`) {
		t.Fatal("must not remove drive root")
	}
	if mayRemoveEmptiedDir(`C:\Users`) {
		t.Fatal("must not remove Users")
	}
	if mayRemoveEmptiedDir(`C:\Users\alice`) {
		t.Fatal("must not remove profile root")
	}
	if !mayRemoveEmptiedDir(`C:\Users\alice\Downloads`) {
		t.Fatal("empty Downloads under a profile may be removed")
	}
}

func TestContentsEqual(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.bin")
	b := filepath.Join(dir, "b.bin")
	c := filepath.Join(dir, "c.bin")
	if err := os.WriteFile(a, []byte("payload-one"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(b, []byte("payload-one"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(c, []byte("payload-TWO"), 0o644); err != nil {
		t.Fatal(err)
	}
	ok, err := contentsEqual(a, b)
	if err != nil || !ok {
		t.Fatalf("equal files: ok=%v err=%v", ok, err)
	}
	ok, err = contentsEqual(a, c)
	if err != nil || ok {
		t.Fatalf("different files: ok=%v err=%v", ok, err)
	}
	emptyA := filepath.Join(dir, "empty-a")
	emptyB := filepath.Join(dir, "empty-b")
	if err := os.WriteFile(emptyA, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(emptyB, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	ok, err = contentsEqual(emptyA, emptyB)
	if err != nil || !ok {
		t.Fatalf("empty files: ok=%v err=%v", ok, err)
	}
}

func TestCopyDoesNotDeleteSource(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "dst.txt")
	if err := os.WriteFile(src, []byte("keep-source"), 0o644); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(src)
	if err != nil {
		t.Fatal(err)
	}
	copied, destOK, cutOK, err := saveThenMaybeCut(src, dst, info, false, false)
	if err != nil {
		t.Fatal(err)
	}
	if !copied || !destOK || cutOK {
		t.Fatalf("copied=%v destOK=%v cutOK=%v", copied, destOK, cutOK)
	}
	if _, err := os.Stat(src); err != nil {
		t.Fatalf("source must remain after copy: %v", err)
	}
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "keep-source" {
		t.Fatalf("got %q", got)
	}
}

func TestCutKeepsSourceWhenDestinationCannotBeWritten(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	if err := os.WriteFile(src, []byte("must-remain"), 0o644); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(src)
	if err != nil {
		t.Fatal(err)
	}
	blocker := filepath.Join(dir, "blocker")
	if err := os.WriteFile(blocker, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(blocker, "dst.txt")
	copied, destOK, cutOK, err := saveThenMaybeCut(src, dst, info, true, false)
	if err == nil {
		t.Fatal("expected write failure")
	}
	if copied || destOK || cutOK {
		t.Fatalf("copied=%v destOK=%v cutOK=%v", copied, destOK, cutOK)
	}
	if _, err := os.Stat(src); err != nil {
		t.Fatalf("source must remain when destination write fails: %v", err)
	}
}

func TestCutRecopiesCorruptResumeThenDeletesSource(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "out", "dst.txt")
	mtime := time.Date(2024, 6, 2, 9, 0, 0, 0, time.Local)
	if err := os.WriteFile(src, []byte("GOOD-DATA"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(src, mtime, mtime); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dst, []byte("BAD!-DATA"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(dst, mtime, mtime); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(src)
	if err != nil {
		t.Fatal(err)
	}
	if !sameFile(info, dst) {
		t.Fatal("test setup: size and mtime should match so resume would skip")
	}
	copied, destOK, cutOK, err := saveThenMaybeCut(src, dst, info, true, true)
	if err != nil {
		t.Fatal(err)
	}
	if !copied || !destOK || !cutOK {
		t.Fatalf("copied=%v destOK=%v cutOK=%v", copied, destOK, cutOK)
	}
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "GOOD-DATA" {
		t.Fatalf("destination should be repaired, got %q", got)
	}
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Fatalf("source should be removed only after verified copy: %v", err)
	}
}

func TestCutKeepsSourceWhenDestPathIsDirectory(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	if err := os.WriteFile(src, []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(src)
	if err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(dir, "dst-as-dir")
	if err := os.MkdirAll(dst, 0o755); err != nil {
		t.Fatal(err)
	}
	copied, destOK, cutOK, err := saveThenMaybeCut(src, dst, info, true, false)
	if err == nil {
		t.Fatal("expected failure when destination is a directory")
	}
	if copied || destOK || cutOK {
		t.Fatalf("copied=%v destOK=%v cutOK=%v", copied, destOK, cutOK)
	}
	if _, err := os.Stat(src); err != nil {
		t.Fatalf("source must remain: %v", err)
	}
}

func TestDestReadyForCutRequiresMatchingBytes(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "dst.txt")
	if err := os.WriteFile(src, []byte("abc"), 0o644); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(src)
	if err != nil {
		t.Fatal(err)
	}
	if err := destReadyForCut(src, dst, info); err == nil {
		t.Fatal("missing destination must not be ready for cut")
	}
	if err := os.WriteFile(dst, []byte("abx"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := destReadyForCut(src, dst, info); err == nil {
		t.Fatal("different content must not be ready for cut")
	}
	if err := os.WriteFile(dst, []byte("abc"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := destReadyForCut(src, dst, info); err != nil {
		t.Fatal(err)
	}
}

func TestDryRunDoesNotCopyOrCut(t *testing.T) {
	src := t.TempDir()
	sub := filepath.Join(src, "Downloads")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	srcFile := filepath.Join(sub, "a.txt")
	if err := os.WriteFile(srcFile, []byte("dry-run"), 0o644); err != nil {
		t.Fatal(err)
	}
	dest := t.TempDir()
	opt := Options{Mode: ModeWithAppData, DestAbs: dest, Cut: true}
	if err := runArchiveJob(context.Background(), []string{src}, dest, "tester", opt, true, true, jobUI{}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(srcFile); err != nil {
		t.Fatalf("dry-run must keep the source: %v", err)
	}
	mapped, err := MapDest(srcFile, dest)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(mapped); !os.IsNotExist(err) {
		t.Fatalf("dry-run must not write the archived file: %v", err)
	}
}

func TestArchiveCopyKeepsSourceEmptyFile(t *testing.T) {
	src := t.TempDir()
	sub := filepath.Join(src, "Downloads")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	srcFile := filepath.Join(sub, "empty.txt")
	if err := os.WriteFile(srcFile, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	dest := t.TempDir()
	opt := Options{Mode: ModeWithAppData, DestAbs: dest, Cut: true}
	res, err := Archive(context.Background(), []string{src}, opt, dest, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.CutFiles != 1 || res.Failed != 0 || res.CutFailed != 0 {
		t.Fatalf("result: %+v", res)
	}
	mapped, err := MapDest(srcFile, dest)
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(mapped)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("got %q", got)
	}
	if _, err := os.Stat(srcFile); !os.IsNotExist(err) {
		t.Fatalf("empty file should be cut after verify: %v", err)
	}
}

func TestCutRefusesWhenSourceIsDestination(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "only-copy.txt")
	if err := os.WriteFile(path, []byte("unique"), 0o644); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	copied, destOK, cutOK, err := saveThenMaybeCut(path, path, info, true, false)
	if err == nil {
		t.Fatal("expected refusal when source and destination are the same")
	}
	if copied || destOK || cutOK {
		t.Fatalf("copied=%v destOK=%v cutOK=%v", copied, destOK, cutOK)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "unique" {
		t.Fatalf("the only copy must remain, got %q", got)
	}
}

func TestCutRemovesReadOnlySourceAfterVerify(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "ro.txt")
	dst := filepath.Join(dir, "out", "ro.txt")
	if err := os.WriteFile(src, []byte("readonly-payload"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(src, 0o444); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(src)
	if err != nil {
		t.Fatal(err)
	}
	copied, destOK, cutOK, err := saveThenMaybeCut(src, dst, info, true, false)
	if err != nil {
		t.Fatal(err)
	}
	if !copied || !destOK || !cutOK {
		t.Fatalf("copied=%v destOK=%v cutOK=%v", copied, destOK, cutOK)
	}
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "readonly-payload" {
		t.Fatalf("got %q", got)
	}
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Fatalf("read-only source should be deleted after verified save: %v", err)
	}
}

func TestCutWhenDestinationAlreadyMatches(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "dst.txt")
	payload := []byte("already-there")
	if err := os.WriteFile(src, payload, 0o644); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(src)
	if err != nil {
		t.Fatal(err)
	}
	if err := copyFileWithRetry(src, dst, info); err != nil {
		t.Fatal(err)
	}
	info, err = os.Stat(src)
	if err != nil {
		t.Fatal(err)
	}
	if !sameFile(info, dst) {
		t.Fatal("destination should already match")
	}
	copied, destOK, cutOK, err := saveThenMaybeCut(src, dst, info, true, true)
	if err != nil {
		t.Fatal(err)
	}
	if copied || !destOK || !cutOK {
		t.Fatalf("copied=%v destOK=%v cutOK=%v", copied, destOK, cutOK)
	}
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(payload) {
		t.Fatalf("got %q", got)
	}
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Fatalf("source should be removed: %v", err)
	}
}

func TestCancelledArchiveKeepsSources(t *testing.T) {
	src := t.TempDir()
	sub := filepath.Join(src, "Downloads")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	srcFile := filepath.Join(sub, "a.txt")
	if err := os.WriteFile(srcFile, []byte("stay"), 0o644); err != nil {
		t.Fatal(err)
	}
	dest := t.TempDir()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	opt := Options{Mode: ModeWithAppData, DestAbs: dest, Cut: true}
	_, err := Archive(ctx, []string{src}, opt, dest, nil, nil)
	if err == nil {
		t.Fatal("expected interrupted archive")
	}
	if _, err := os.Stat(srcFile); err != nil {
		t.Fatalf("cancelled cut must keep the source: %v", err)
	}
}

func TestArchiveSkipsFilesAlreadyInDestination(t *testing.T) {
	src := t.TempDir()
	keepDir := filepath.Join(src, "Downloads")
	if err := os.MkdirAll(keepDir, 0o755); err != nil {
		t.Fatal(err)
	}
	keep := filepath.Join(keepDir, "keep.txt")
	if err := os.WriteFile(keep, []byte("outside"), 0o644); err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(src, "archive-dest")
	if err := os.MkdirAll(dest, 0o755); err != nil {
		t.Fatal(err)
	}
	inside := filepath.Join(dest, "inside.txt")
	if err := os.WriteFile(inside, []byte("already-in-dest"), 0o644); err != nil {
		t.Fatal(err)
	}
	opt := Options{Mode: ModeWithAppData, DestAbs: dest, Cut: true}
	if _, err := Archive(context.Background(), []string{src}, opt, dest, nil, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(inside); err != nil {
		t.Fatalf("file already in destination must not be deleted: %v", err)
	}
	got, err := os.ReadFile(inside)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "already-in-dest" {
		t.Fatalf("got %q", got)
	}
	mapped, err := MapDest(keep, dest)
	if err != nil {
		t.Fatal(err)
	}
	got, err = os.ReadFile(mapped)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "outside" {
		t.Fatalf("got %q", got)
	}
	if _, err := os.Stat(keep); !os.IsNotExist(err) {
		t.Fatalf("file outside destination should be cut: %v", err)
	}
}

func TestArchiveUnicodeName(t *testing.T) {
	src := t.TempDir()
	sub := filepath.Join(src, "下载")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	srcFile := filepath.Join(sub, "说明.txt")
	if err := os.WriteFile(srcFile, []byte("中文内容"), 0o644); err != nil {
		t.Fatal(err)
	}
	dest := t.TempDir()
	opt := Options{Mode: ModeWithAppData, DestAbs: dest}
	if _, err := Archive(context.Background(), []string{src}, opt, dest, nil, nil); err != nil {
		t.Fatal(err)
	}
	mapped, err := MapDest(srcFile, dest)
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(mapped)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "中文内容" {
		t.Fatalf("got %q", got)
	}
	if _, err := os.Stat(srcFile); err != nil {
		t.Fatalf("copy must keep unicode source: %v", err)
	}
}

func TestContentsEqualLargeFile(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.bin")
	b := filepath.Join(dir, "b.bin")
	c := filepath.Join(dir, "c.bin")
	payload := make([]byte, 300*1024)
	for i := range payload {
		payload[i] = byte(i)
	}
	if err := os.WriteFile(a, payload, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(b, payload, 0o644); err != nil {
		t.Fatal(err)
	}
	changed := append([]byte(nil), payload...)
	changed[len(changed)-1] ^= 0xFF
	if err := os.WriteFile(c, changed, 0o644); err != nil {
		t.Fatal(err)
	}
	ok, err := contentsEqual(a, b)
	if err != nil || !ok {
		t.Fatalf("equal large files: ok=%v err=%v", ok, err)
	}
	ok, err = contentsEqual(a, c)
	if err != nil || ok {
		t.Fatalf("different large files: ok=%v err=%v", ok, err)
	}
}

func TestCopyLeavesNoTempSidecar(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "dst.txt")
	if err := os.WriteFile(src, []byte("sidecar"), 0o644); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(src)
	if err != nil {
		t.Fatal(err)
	}
	if err := copyFileWithRetry(src, dst, info); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(dst + ".archiving"); !os.IsNotExist(err) {
		t.Fatalf("leftover .archiving: %v", err)
	}
	if _, err := os.Stat(dst + ".replacing"); !os.IsNotExist(err) {
		t.Fatalf("leftover .replacing: %v", err)
	}
}

func TestExtraExcludeSkipsNamedFolder(t *testing.T) {
	src := t.TempDir()
	keepDir := filepath.Join(src, "Downloads")
	skipDir := filepath.Join(src, "games")
	if err := os.MkdirAll(keepDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(skipDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(keepDir, "a.txt"), []byte("keep"), 0o644); err != nil {
		t.Fatal(err)
	}
	skipFile := filepath.Join(skipDir, "b.txt")
	if err := os.WriteFile(skipFile, []byte("skip-me"), 0o644); err != nil {
		t.Fatal(err)
	}
	dest := t.TempDir()
	opt := Options{Mode: ModeWithAppData, DestAbs: dest, ExtraExclude: []string{"games"}, Cut: true}
	if _, err := Archive(context.Background(), []string{src}, opt, dest, nil, nil); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(skipFile); err != nil {
		t.Fatalf("excluded folder must not be cut: %v", err)
	}
	mappedSkip, err := MapDest(skipFile, dest)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(mappedSkip); !os.IsNotExist(err) {
		t.Fatal("excluded folder must not be copied")
	}
}

func TestParseModeAndSplitRoots(t *testing.T) {
	if ParseMode("personal") != ModePersonal || ParseMode("all") != ModeAllNonSystem || ParseMode("") != ModeWithAppData {
		t.Fatalf("ParseMode: personal=%v all=%v default=%v", ParseMode("personal"), ParseMode("all"), ParseMode(""))
	}
	got := splitRoots("C:, C:\\Users; D:")
	if len(got) != 3 || got[0] != "C:" || got[1] != `C:\Users` || got[2] != "D:" {
		t.Fatalf("got %#v", got)
	}
	dedup := splitRoots(`C:, C:\`)
	if len(dedup) != 1 || dedup[0] != "C:" {
		t.Fatalf("drive letter aliases should collapse: %#v", dedup)
	}
}

func TestMergePickedRootAndBrowseStart(t *testing.T) {
	if got := initialBrowsePath("C:, D:"); got != `D:\` {
		t.Fatalf("initialBrowsePath last drive: %q", got)
	}
	if got := initialBrowsePath(`C:\Users;E:\data`); got != `E:\data` {
		t.Fatalf("initialBrowsePath last folder: %q", got)
	}
	if got := mergePickedRoot("C:", `D:\data`); got != `C:, D:\data` {
		t.Fatalf("append other drive: %q", got)
	}
	if got := mergePickedRoot("C:", `C:\`); got != "C:" {
		t.Fatalf("duplicate drive root: %q", got)
	}
	if got := mergePickedRoot(`C:, D:`, `D:\`); got != `C:, D:` {
		t.Fatalf("duplicate in list: %q", got)
	}
	if got := mergePickedRoot("", `E:\usb`); got != `E:\usb` {
		t.Fatalf("empty current: %q", got)
	}
}

func TestWriteReportChineseFolderLine(t *testing.T) {
	setLang(LangZHCN)
	t.Cleanup(func() { setLang(LangEN) })

	dest := t.TempDir()
	sum := &Summary{
		Files: 1,
		Bytes: 2,
		ByFolder: map[string]*FolderStat{
			`C:\Downloads`: {Key: `C:\Downloads`, Files: 1, Bytes: 2},
		},
	}
	if err := writeReport(dest, "PC", "alice", []string{`C:\`}, Options{Mode: ModePersonal}, sum, nil, "", time.Now()); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(dest, reportFileName))
	if err != nil {
		t.Fatal(err)
	}
	text := string(b)
	if !strings.Contains(text, "用户数据") {
		t.Fatalf("missing Chinese title:\n%s", text)
	}
	if !strings.Contains(text, "个文件") {
		t.Fatalf("folder line should be Chinese:\n%s", text)
	}
	if strings.Contains(text, " files ") {
		t.Fatalf("leftover English 'files' in report:\n%s", text)
	}
}

func TestWriteFailCSVHeadersChinese(t *testing.T) {
	setLang(LangZHCN)
	t.Cleanup(func() { setLang(LangEN) })

	p := filepath.Join(t.TempDir(), failListFileName)
	if err := writeFailCSV(p, [][]string{{"a", "b", "c"}}); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	text := string(b)
	if !strings.Contains(text, "源路径") || !strings.Contains(text, "目标路径") {
		t.Fatalf("CSV headers:\n%s", text)
	}
}
