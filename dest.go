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
		return "", fmt.Errorf("%s", T("ErrNoDrive", srcAbs))
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
	if name == "" || isReservedWindowsName(name) {
		return "archive"
	}
	return name
}

func isReservedWindowsName(name string) bool {
	n := strings.ToUpper(strings.TrimSpace(name))
	switch n {
	case "CON", "PRN", "AUX", "NUL", "CLOCK$":
		return true
	}
	if len(n) == 4 && (strings.HasPrefix(n, "COM") || strings.HasPrefix(n, "LPT")) {
		d := n[3]
		return d >= '1' && d <= '9'
	}
	return false
}

func isVolumeRoot(p string) bool {
	p = strings.TrimSpace(p)
	if p == "" {
		return false
	}
	p = normalizeAbs(p)
	vol := filepath.VolumeName(p)
	if vol == "" {
		return false
	}
	rest := strings.Trim(strings.TrimPrefix(p, vol), `\/`)
	return rest == ""
}

func probeDestWritable(dest string) error {
	vol := filepath.VolumeName(dest)
	if vol == "" {
		return fmt.Errorf("%s", T("ErrDestDrive", dest))
	}
	if isVolumeRoot(dest) {
		return fmt.Errorf("%s", T("ErrDestDriveRoot"))
	}
	root := vol + `\`
	kind := driveKind(root)
	if kind == "missing" {
		return fmt.Errorf("%s", T("ErrDriveMissing", vol))
	}
	if kind == "cd-rom" {
		return fmt.Errorf("%s", T("ErrDriveCD", vol))
	}
	if _, err := os.Stat(root); err != nil {
		return fmt.Errorf("%s", T("ErrAccessDrive", vol, explainIOError(err)))
	}
	if err := os.MkdirAll(longPath(dest), 0o755); err != nil {
		return fmt.Errorf("%s", T("ErrCreateFolder", dest, explainIOError(err)))
	}
	probe := filepath.Join(dest, writeProbeFileName)
	payload := []byte("user-data-archiver-write-probe")
	if err := os.WriteFile(longPath(probe), payload, 0o644); err != nil {
		return fmt.Errorf("%s", T("ErrNotWritable", dest, explainIOError(err)))
	}
	got, err := os.ReadFile(longPath(probe))
	removeErr := os.Remove(longPath(probe))
	if err != nil {
		return fmt.Errorf("%s", T("ErrReadProbe", explainIOError(err)))
	}
	if string(got) != string(payload) {
		return fmt.Errorf("%s", T("ErrMismatch"))
	}
	if removeErr != nil {
		return fmt.Errorf("%s", T("ErrDeleteProbe", probe, explainIOError(removeErr)))
	}
	return nil
}

func explainIOError(err error) string {
	if err == nil {
		return ""
	}
	if errors.Is(err, os.ErrPermission) {
		return T("ErrAccessDenied")
	}
	var errno syscall.Errno
	if errors.As(err, &errno) {
		switch errno {
		case 5:
			return T("ErrAccessDenied")
		case 19:
			return T("ErrWriteProtect")
		case 21:
			return T("ErrNotReady")
		case 82:
			return T("ErrCannotCreate")
		case 112:
			return T("ErrNoSpace")
		case 3, 2:
			return T("ErrPathNotFound")
		}
	}
	return err.Error()
}

func destDriveInfo(dest string) string {
	dest = strings.TrimSpace(dest)
	if dest == "" {
		return T("NoDestSelected")
	}
	dest = normalizeAbs(dest)
	vol := filepath.VolumeName(dest)
	if vol == "" {
		return dest
	}
	root := vol + `\`
	kind := translateDriveKind(driveKind(root))
	label := volumeLabel(root)
	freeStr := T("FreeUnknown")
	if free, _, err := diskFreeBytes(root); err == nil {
		freeStr = T("FreeBytes", formatBytes(int64(free)))
	}
	if label != "" {
		return T("VolumeLabel", vol, kind, label, freeStr)
	}
	return T("VolumePlain", vol, kind, freeStr)
}

func translateDriveKind(kind string) string {
	switch kind {
	case "missing":
		return T("DriveMissingKind")
	case "removable":
		return T("DriveRemovable")
	case "local disk":
		return T("DriveLocal")
	case "network":
		return T("DriveNetwork")
	case "cd-rom":
		return T("DriveCD")
	case "ram disk":
		return T("DriveRAM")
	default:
		return T("DriveUnknown")
	}
}
