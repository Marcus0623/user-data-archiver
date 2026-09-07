package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

const (
	reportFileName     = "_archive-report.txt"
	failListFileName   = "_failed-files.csv"
	writeProbeFileName = "_write-test.tmp"
	defaultDestParent  = `D:\offboarding-archive`
)

func MapDest(srcAbs, archiveRoot string) (string, error) {
	srcAbs = normalizeAbs(srcAbs)
	archiveRoot = normalizeAbs(archiveRoot)
	vol, rel := volumeAndRel(srcAbs)
	if vol == "" {
		return "", fmt.Errorf("cannot determine drive for path: %s", srcAbs)
	}
	letter := driveFolderName(vol)
	if rel == "" {
		return filepath.Join(archiveRoot, letter), nil
	}
	return filepath.Join(archiveRoot, letter, rel), nil
}

func sanitizeEmployeeName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ""
	}
	replacer := strings.NewReplacer(
		`<`, "_", `>`, "_", `:`, "_", `"`, "_",
		`/`, "_", `\`, "_", `|`, "_", `?`, "_", `*`, "_",
	)
	name = replacer.Replace(name)
	name = strings.Trim(name, " .")
	if name == "" || strings.EqualFold(name, "CON") || strings.EqualFold(name, "NUL") {
		return "archive"
	}
	return name
}

func probeDestWritable(dest string) error {
	vol := filepath.VolumeName(dest)
	if vol == "" {
		return fmt.Errorf("cannot determine destination drive: %s", dest)
	}
	root := vol + `\`
	kind := driveKind(root)
	if kind == "missing" {
		return fmt.Errorf("destination drive %s is missing or not ready (USB unplugged, empty optical drive, or BitLocker locked)", vol)
	}
	if kind == "cd-rom" {
		return fmt.Errorf("destination drive %s is a CD/DVD drive; use a hard disk or USB drive", vol)
	}
	if _, err := os.Stat(root); err != nil {
		return fmt.Errorf("cannot access destination drive %s: %s", vol, explainIOError(err))
	}
	if err := os.MkdirAll(longPath(dest), 0o755); err != nil {
		return fmt.Errorf("cannot create destination folder %s: %s", dest, explainIOError(err))
	}
	probe := filepath.Join(dest, writeProbeFileName)
	payload := []byte("user-data-archiver-write-probe")
	if err := os.WriteFile(longPath(probe), payload, 0o644); err != nil {
		return fmt.Errorf("destination is not writable %s: %s", dest, explainIOError(err))
	}
	got, err := os.ReadFile(longPath(probe))
	removeErr := os.Remove(longPath(probe))
	if err != nil {
		return fmt.Errorf("cannot read back the probe file (read permission missing): %s", explainIOError(err))
	}
	if string(got) != string(payload) {
		return fmt.Errorf("destination read/write check failed: content mismatch")
	}
	if removeErr != nil {
		return fmt.Errorf("probe file was written but cannot be deleted %s: %s", probe, explainIOError(removeErr))
	}
	return nil
}

func explainIOError(err error) string {
	if err == nil {
		return ""
	}
	if errors.Is(err, os.ErrPermission) {
		return "access denied"
	}
	var errno syscall.Errno
	if errors.As(err, &errno) {
		switch errno {
		case 5:
			return "access denied"
		case 19:
			return "disk is write-protected (USB switch or policy)"
		case 21:
			return "device not ready (locked or media missing)"
		case 82:
			return "cannot create file (read-only or insufficient permission)"
		case 112:
			return "not enough disk space"
		case 3, 2:
			return "path not found"
		}
	}
	return err.Error()
}

func destDriveInfo(dest string) string {
	dest = strings.TrimSpace(dest)
	if dest == "" {
		return "No destination selected."
	}
	dest = normalizeAbs(dest)
	vol := filepath.VolumeName(dest)
	if vol == "" {
		return dest
	}
	root := vol + `\`
	kind := driveKind(root)
	label := volumeLabel(root)
	freeStr := "free space unknown"
	if free, _, err := diskFreeBytes(root); err == nil {
		freeStr = "free " + formatBytes(int64(free))
	}
	if label != "" {
		return fmt.Sprintf("%s  %s  label %q  %s", vol, kind, label, freeStr)
	}
	return fmt.Sprintf("%s  %s  %s", vol, kind, freeStr)
}
