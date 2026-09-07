package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestCopyFileAndResume(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "src.txt")
	dst := filepath.Join(dir, "sub", "dst.txt")
	content := []byte("离职归档测试内容")
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
		t.Fatal("复制后应判定为相同文件以便续传跳过")
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
}
