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
		return true, isDir, "归档目标目录"
	}

	_, rel := volumeAndRel(absPath)
	parts := pathComponents(rel)
	if len(parts) == 0 {
		return false, false, ""
	}

	for _, p := range parts {
		if hasName(skipAnywhereDirNames, p) {
			return true, isDir, "系统卷/回收站"
		}
		for _, ex := range opt.ExtraExclude {
			if strings.EqualFold(p, strings.TrimSpace(ex)) && strings.TrimSpace(ex) != "" {
				return true, isDir, "自定义排除"
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
			return true, isDir, "已安装程序/ProgramData"
		}
		if hasName(oemRootDirNames, rootName) && opt.Mode == ModeAllNonSystem && opt.IncludeProgramFiles {
			return false, false, ""
		}
		return true, isDir, "系统/OEM 初始目录"
	}

	if !isDir && len(logical) == 1 && hasName(systemRootFileNames, logical[0]) {
		return true, false, "系统页面/休眠文件"
	}

	if len(logical) >= 2 && strings.EqualFold(logical[0], "Users") {
		userName := logical[1]
		if hasName(skipUserProfileNames, userName) {
			return true, isDir, "系统默认用户配置"
		}
		if len(logical) >= 3 {
			leaf := logical[len(logical)-1]
			if !isDir && isHiveFile(leaf) && len(logical) == 3 {
				return true, false, "用户注册表配置单元"
			}
			if opt.Mode == ModePersonal && underAppData(logical[2:]) {
				return true, isDir, "AppData（个人资料模式已排除）"
			}
			if opt.Mode != ModePersonal && underAppData(logical[2:]) && isDir && hasName(cacheDirNames, leaf) {
				return true, true, "软件缓存目录"
			}
		}
	}

	if opt.SkipRegeneratable && isDir && hasName(regeneratableDirNames, logical[len(logical)-1]) {
		return true, true, "可再生成目录"
	}

	return false, false, ""
}

func isHiveFile(name string) bool {
	n := strings.ToLower(name)
	return strings.HasPrefix(n, "ntuser.dat") || n == "ntuser.ini"
}
