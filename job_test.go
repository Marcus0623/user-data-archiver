package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestJobPreviewLogsChinese(t *testing.T) {
	setLang(LangZHCN)
	t.Cleanup(func() { setLang(LangEN) })

	src := t.TempDir()
	dest := t.TempDir()
	if err := os.MkdirAll(filepath.Join(src, "Downloads"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "Downloads", "a.txt"), []byte("hi"), 0o644); err != nil {
		t.Fatal(err)
	}

	var logs []string
	err := runArchiveJob(context.Background(), []string{src}, dest, "tester", Options{
		Mode:         ModeWithAppData,
		DestAbs:      dest,
		EmployeeName: "tester",
	}, true, true, jobUI{
		log: func(format string, args ...any) {
			logs = append(logs, fmt.Sprintf(format, args...))
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(logs, "\n")
	for _, want := range []string{"正在扫描", "预览", "预览结束", "个文件"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("missing %q in:\n%s", want, joined)
		}
	}
	if strings.Contains(joined, "Scanning (this may take") {
		t.Fatalf("still English scan line:\n%s", joined)
	}
}

func TestJobWarnsWhenDestInsideSource(t *testing.T) {
	setLang(LangEN)
	t.Cleanup(func() { setLang(LangEN) })

	disk := t.TempDir()
	if err := os.MkdirAll(filepath.Join(disk, "keep"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(disk, "keep", "a.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	dest := filepath.Join(disk, "offboarding-archive", "alice")
	var logs []string
	err := runArchiveJob(context.Background(), []string{disk}, dest, "alice", Options{
		Mode:         ModeWithAppData,
		DestAbs:      dest,
		EmployeeName: "alice",
	}, true, true, jobUI{
		log: func(format string, args ...any) {
			logs = append(logs, fmt.Sprintf(format, args...))
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(logs, "\n")
	if !strings.Contains(joined, "destination is inside the source path") {
		t.Fatalf("expected dest-inside-source warning:\n%s", joined)
	}
}
