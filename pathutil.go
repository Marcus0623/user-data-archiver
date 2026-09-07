package main

import (
	"path/filepath"
	"strings"
)

func stripLongPath(p string) string {
	if strings.HasPrefix(p, `\\?\UNC\`) {
		return `\\` + strings.TrimPrefix(p, `\\?\UNC\`)
	}
	if strings.HasPrefix(p, `\\?\`) {
		return strings.TrimPrefix(p, `\\?\`)
	}
	return p
}

func normalizeAbs(p string) string {
	p = stripLongPath(p)
	abs, err := filepath.Abs(p)
	if err != nil {
		return filepath.Clean(p)
	}
	abs = filepath.Clean(abs)
	if long := getLongPathName(abs); long != "" {
		return long
	}
	return abs
}

func volumeAndRel(abs string) (vol, rel string) {
	abs = filepath.Clean(stripLongPath(abs))
	vol = filepath.VolumeName(abs)
	rest := abs
	if vol != "" {
		rest = strings.TrimPrefix(abs, vol)
	}
	rest = strings.TrimPrefix(rest, string(filepath.Separator))
	return vol, rest
}

func driveFolderName(vol string) string {
	vol = strings.TrimSpace(vol)
	if len(vol) >= 2 && vol[1] == ':' {
		return strings.ToUpper(vol[:1])
	}
	s := strings.TrimPrefix(vol, `\\`)
	s = strings.ReplaceAll(s, `\`, `_`)
	s = strings.ReplaceAll(s, ":", "_")
	if s == "" {
		return "UNKNOWN"
	}
	return "UNC_" + s
}

func pathComponents(rel string) []string {
	if rel == "" {
		return nil
	}
	parts := strings.Split(rel, string(filepath.Separator))
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p == "" || p == "." {
			continue
		}
		out = append(out, p)
	}
	return out
}

func lowerComponents(rel string) []string {
	parts := pathComponents(rel)
	for i, p := range parts {
		parts[i] = strings.ToLower(p)
	}
	return parts
}

func isBaseOf(base, target string) bool {
	base = normalizeAbs(base)
	target = normalizeAbs(target)
	rel, err := filepath.Rel(base, target)
	if err != nil {
		return false
	}
	if rel == "." {
		return true
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func samePath(a, b string) bool {
	return strings.EqualFold(normalizeAbs(a), normalizeAbs(b))
}

func stripWindowsOldPrefix(parts []string) []string {
	if len(parts) >= 2 && strings.EqualFold(parts[0], "Windows.old") {
		return parts[1:]
	}
	return parts
}

func topLevelKey(vol, rel string) string {
	parts := pathComponents(rel)
	if len(parts) == 0 {
		return driveFolderName(vol) + `\`
	}
	return filepath.Join(driveFolderName(vol), parts[0])
}
