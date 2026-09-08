package main

import (
	"fmt"
	"strings"
	"sync"
	"syscall"
	"unsafe"
)

type Lang string

const (
	LangEN   Lang = "en"
	LangZHCN Lang = "zh-CN"
	LangZHTW Lang = "zh-TW"
	LangJA   Lang = "ja"
	LangFR   Lang = "fr"
	LangRU   Lang = "ru"
	LangVI   Lang = "vi"
)

var allLangs = []Lang{LangEN, LangZHCN, LangZHTW, LangJA, LangFR, LangRU, LangVI}

var (
	langMu      sync.RWMutex
	currentLang = LangEN
	catalogs    map[Lang]map[string]string
	uiRestart   bool
)

func init() {
	catalogs = buildCatalogs()
	setLang(detectWindowsLang())
}

func currentLangCode() Lang {
	langMu.RLock()
	defer langMu.RUnlock()
	return currentLang
}

func setLang(lang Lang) {
	lang = normalizeLang(lang)
	langMu.Lock()
	currentLang = lang
	langMu.Unlock()
}

func T(key string, args ...any) string {
	langMu.RLock()
	lang := currentLang
	langMu.RUnlock()
	s := lookup(lang, key)
	if len(args) == 0 {
		return s
	}
	return fmt.Sprintf(s, args...)
}

func lookup(lang Lang, key string) string {
	if m := catalogs[lang]; m != nil {
		if s, ok := m[key]; ok && s != "" {
			return s
		}
	}
	if m := catalogs[LangEN]; m != nil {
		if s, ok := m[key]; ok && s != "" {
			return s
		}
	}
	return key
}

func ParseLang(s string) Lang {
	return normalizeLang(Lang(strings.TrimSpace(s)))
}

func normalizeLang(lang Lang) Lang {
	s := strings.ToLower(strings.TrimSpace(string(lang)))
	s = strings.ReplaceAll(s, "_", "-")
	switch {
	case s == "":
		return LangEN
	case s == "auto":
		return detectWindowsLang()
	case s == "en" || strings.HasPrefix(s, "en-"):
		return LangEN
	case s == "zh-tw" || s == "zh-hk" || s == "zh-mo" || s == "zh-hant" || s == "hant":
		return LangZHTW
	case s == "zh" || s == "zh-cn" || s == "zh-sg" || s == "zh-hans" || s == "hans" || s == "cn":
		return LangZHCN
	case s == "ja" || s == "jp" || strings.HasPrefix(s, "ja-"):
		return LangJA
	case s == "fr" || strings.HasPrefix(s, "fr-"):
		return LangFR
	case s == "ru" || strings.HasPrefix(s, "ru-"):
		return LangRU
	case s == "vi" || s == "vn" || strings.HasPrefix(s, "vi-"):
		return LangVI
	default:
		return LangEN
	}
}

func langNativeName(lang Lang) string {
	switch lang {
	case LangZHCN:
		return "简体中文"
	case LangZHTW:
		return "繁體中文"
	case LangJA:
		return "日本語"
	case LangFR:
		return "Français"
	case LangRU:
		return "Русский"
	case LangVI:
		return "Tiếng Việt"
	default:
		return "English"
	}
}

func langNames() []string {
	out := make([]string, len(allLangs))
	for i, lang := range allLangs {
		out[i] = langNativeName(lang)
	}
	return out
}

func langIndex(lang Lang) int {
	for i, item := range allLangs {
		if item == lang {
			return i
		}
	}
	return 0
}

func detectWindowsLang() Lang {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	proc := kernel32.NewProc("GetUserDefaultLocaleName")
	buf := make([]uint16, 85)
	r, _, _ := proc.Call(uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	if r == 0 {
		return LangEN
	}
	name := strings.TrimSpace(syscall.UTF16ToString(buf))
	if name == "" || strings.EqualFold(name, "auto") {
		return LangEN
	}
	return ParseLang(name)
}

func isAffirmative(text string) bool {
	t := strings.TrimSpace(strings.ToLower(text))
	if t == "" {
		return false
	}
	switch t {
	case "y", "yes", "1", "true", "o", "oui", "да", "д", "はい", "是", "好", "有", "có", "ok":
		return true
	}
	return t == strings.ToLower(strings.TrimSpace(T("WordYes")))
}

func mergeCatalog(base, over map[string]string) map[string]string {
	out := make(map[string]string, len(base)+len(over))
	for k, v := range base {
		out[k] = v
	}
	for k, v := range over {
		if v != "" {
			out[k] = v
		}
	}
	return out
}

func buildCatalogs() map[Lang]map[string]string {
	en := catalogEN()
	return map[Lang]map[string]string{
		LangEN:   en,
		LangZHCN: mergeCatalog(en, catalogZHCN()),
		LangZHTW: mergeCatalog(en, catalogZHTW()),
		LangJA:   mergeCatalog(en, catalogJA()),
		LangFR:   mergeCatalog(en, catalogFR()),
		LangRU:   mergeCatalog(en, catalogRU()),
		LangVI:   mergeCatalog(en, catalogVI()),
	}
}
