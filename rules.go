package main

import (
	"strings"
)

var systemRootDirNames = map[string]struct{}{
	"windows":                   {},
	"$recycle.bin":              {},
	"system volume information": {},
	"recovery":                  {},
	"boot":                      {},
	"efi":                       {},
	"perflogs":                  {},
	"documents and settings":    {},
	"config.msi":                {},
	"msocache":                  {},
	"$winreagent":               {},
	"onedrivetemp":              {},
	"windowsapps":               {},
	"program files":             {},
	"program files (x86)":       {},
	"program files (arm)":       {},
	"programdata":               {},
	"intel":                     {},
	"amd":                       {},
	"nvidia":                    {},
	"nvidia corporation":        {},
	"dell":                      {},
	"hp":                        {},
	"lenovo":                    {},
	"asus":                      {},
	"acer":                      {},
	"msi":                       {},
	"gigabyte":                  {},
	"realtek":                   {},
	"inetpub":                   {},
	"sysprep":                   {},
	"$windows.~bt":              {},
	"$windows.~ws":              {},
	"wdac":                      {},
}

var programRootDirNames = map[string]struct{}{
	"program files":       {},
	"program files (x86)": {},
	"program files (arm)": {},
	"programdata":         {},
	"windowsapps":         {},
}

var oemRootDirNames = map[string]struct{}{
	"intel":              {},
	"amd":                {},
	"nvidia":             {},
	"nvidia corporation": {},
	"dell":               {},
	"hp":                 {},
	"lenovo":             {},
	"asus":               {},
	"acer":               {},
	"msi":                {},
	"gigabyte":           {},
	"realtek":            {},
}

var skipAnywhereDirNames = map[string]struct{}{
	"$recycle.bin":              {},
	"system volume information": {},
	"$winreagent":               {},
}

var skipUserProfileNames = map[string]struct{}{
	"default":            {},
	"default user":       {},
	"all users":          {},
	"wdagutilityaccount": {},
	"defaultapppool":     {},
}

var systemRootFileNames = map[string]struct{}{
	"pagefile.sys":      {},
	"hiberfil.sys":      {},
	"swapfile.sys":      {},
	"dumpstack.log.tmp": {},
	"bootmgr":           {},
	"bootnxt":           {},
	"bootsect.bak":      {},
}

var cacheDirNames = map[string]struct{}{
	"cache":                    {},
	"caches":                   {},
	"code cache":               {},
	"gpucache":                 {},
	"shadercache":              {},
	"dawncache":                {},
	"grshadercache":            {},
	"temp":                     {},
	"tmp":                      {},
	"inetcache":                {},
	"webcache":                 {},
	"crashdumps":               {},
	"package cache":            {},
	"temporary internet files": {},
	"d3dscache":                {},
	"pip":                      {},
	"npm-cache":                {},
	"yarn-cache":               {},
	"yarn cache":               {},
	".cache":                   {},
	"service worker":           {},
	"serviceworker":            {},
}

var regeneratableDirNames = map[string]struct{}{
	"node_modules":     {},
	"__pycache__":      {},
	".pytest_cache":    {},
	".mypy_cache":      {},
	".npm":             {},
	".yarn":            {},
	"bower_components": {},
	".gradle":          {},
	".nuget":           {},
	".pnpm-store":      {},
	"pnpm-store":       {},
}

func hasName(set map[string]struct{}, name string) bool {
	_, ok := set[strings.ToLower(strings.TrimSpace(name))]
	return ok
}

func underAppData(parts []string) bool {
	for _, p := range parts {
		n := strings.ToLower(p)
		if n == "appdata" || n == "application data" {
			return true
		}
	}
	return false
}

func skipReason(absPath string, isDir bool, opt Options) (skip bool, skipWhole bool, reason string) {
	absPath = normalizeAbs(absPath)
	if opt.DestAbs != "" && isBaseOf(opt.DestAbs, absPath) {
		return true, isDir, "archive destination"
	}

	_, rel := volumeAndRel(absPath)
	parts := pathComponents(rel)
	if len(parts) == 0 {
		return false, false, ""
	}

	for _, p := range parts {
		if hasName(skipAnywhereDirNames, p) {
			return true, isDir, "recycle bin / system volume"
		}
		for _, ex := range opt.ExtraExclude {
			if strings.EqualFold(p, strings.TrimSpace(ex)) && strings.TrimSpace(ex) != "" {
				return true, isDir, "custom exclude"
			}
		}
	}

	logical := stripWindowsOldPrefix(parts)
	if len(logical) == 0 {
		return false, false, ""
	}

	rootName := logical[0]
	if hasName(systemRootDirNames, rootName) {
		if hasName(programRootDirNames, rootName) {
			if opt.IncludeProgramFiles {
				return false, false, ""
			}
			return true, isDir, "installed programs / ProgramData"
		}
		if hasName(oemRootDirNames, rootName) && opt.Mode == ModeAllNonSystem && opt.IncludeProgramFiles {
			return false, false, ""
		}
		return true, isDir, "system/OEM directory"
	}

	if !isDir && len(logical) == 1 && hasName(systemRootFileNames, logical[0]) {
		return true, false, "paging/hiber file"
	}

	if len(logical) >= 2 && strings.EqualFold(logical[0], "Users") {
		userName := logical[1]
		if hasName(skipUserProfileNames, userName) {
			return true, isDir, "default user profile"
		}
		if len(logical) >= 3 {
			leaf := logical[len(logical)-1]
			if !isDir && isHiveFile(leaf) && len(logical) == 3 {
				return true, false, "user registry hive"
			}
			if opt.Mode == ModePersonal && underAppData(logical[2:]) {
				return true, isDir, "AppData (excluded in personal mode)"
			}
			if opt.Mode != ModePersonal && underAppData(logical[2:]) && isDir && hasName(cacheDirNames, leaf) {
				return true, true, "app cache"
			}
		}
	}

	if opt.SkipRegeneratable && isDir && hasName(regeneratableDirNames, logical[len(logical)-1]) {
		return true, true, "regeneratable directory"
	}

	return false, false, ""
}

func isHiveFile(name string) bool {
	n := strings.ToLower(name)
	return strings.HasPrefix(n, "ntuser.dat") || n == "ntuser.ini"
}
