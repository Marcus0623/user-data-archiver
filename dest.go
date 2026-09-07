package main

import (
	"fmt"
	"path/filepath"
	"strings"
)

func MapDest(srcAbs, archiveRoot string) (string, error) {
	srcAbs = normalizeAbs(srcAbs)
	archiveRoot = normalizeAbs(archiveRoot)
	vol, rel := volumeAndRel(srcAbs)
	if vol == "" {
		return "", fmt.Errorf("无法识别路径所在驱动器: %s", srcAbs)
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
