package main

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	setLang(LangEN)
	os.Exit(m.Run())
}

func TestParseLang(t *testing.T) {
	cases := map[string]Lang{
		"en":    LangEN,
		"en-US": LangEN,
		"zh":    LangZHCN,
		"zh-CN": LangZHCN,
		"zh-TW": LangZHTW,
		"zh-HK": LangZHTW,
		"ja":    LangJA,
		"fr-FR": LangFR,
		"ru":    LangRU,
		"vi-VN": LangVI,
		"nope":  LangEN,
	}
	for in, want := range cases {
		if got := ParseLang(in); got != want {
			t.Fatalf("%s: got %s want %s", in, got, want)
		}
	}
}

func TestCatalogKeys(t *testing.T) {
	en := catalogs[LangEN]
	for _, lang := range allLangs {
		setLang(lang)
		for key := range en {
			if got := lookup(lang, key); got == "" {
				t.Errorf("%s missing %s", lang, key)
			}
		}
	}
	setLang(LangEN)
}

func TestCatalogFormatVerbs(t *testing.T) {
	en := catalogs[LangEN]
	for _, lang := range allLangs {
		for key, enVal := range en {
			got := lookup(lang, key)
			if formatVerbCount(got) != formatVerbCount(enVal) {
				t.Errorf("%s %s verbs: en=%d lang=%d\nen %q\ngot %q", lang, key, formatVerbCount(enVal), formatVerbCount(got), enVal, got)
			}
			args := dummyFormatArgs(got)
			out := got
			if len(args) > 0 {
				out = fmt.Sprintf(got, args...)
			}
			if strings.Contains(out, "%!") {
				t.Errorf("%s %s sprintf leftover: %q", lang, key, out)
			}
		}
	}
	setLang(LangEN)
}

func TestChineseAppName(t *testing.T) {
	setLang(LangZHCN)
	t.Cleanup(func() { setLang(LangEN) })
	if T("CopyButton") != "复制文件" {
		t.Fatalf("got %q", T("CopyButton"))
	}
	if T("JobPreview") == catalogs[LangEN]["JobPreview"] {
		t.Fatal("Chinese preview heading should not fall back to English")
	}
	if T("SourceHelp") == catalogs[LangEN]["SourceHelp"] {
		t.Fatal("Chinese source help should not fall back to English")
	}
	if strings.Contains(T("VolumeLabel"), "label") {
		t.Fatalf("Chinese volume line still has English label: %q", T("VolumeLabel"))
	}
	if !strings.Contains(T("SourceHelp"), "追加") {
		t.Fatalf("Chinese source help should say browse appends: %q", T("SourceHelp"))
	}
}

func TestDestDriveInfoLocalized(t *testing.T) {
	setLang(LangEN)
	if got := destDriveInfo(""); got != catalogs[LangEN]["NoDestSelected"] {
		t.Fatalf("en: %q", got)
	}
	setLang(LangZHCN)
	t.Cleanup(func() { setLang(LangEN) })
	if got := destDriveInfo(""); got != catalogs[LangZHCN]["NoDestSelected"] {
		t.Fatalf("zh-CN: %q", got)
	}
}

func TestIsAffirmative(t *testing.T) {
	if !isAffirmative("yes") || !isAffirmative("Y") || !isAffirmative("是") {
		t.Fatal("expected yes words to be accepted")
	}
}

func TestParseLangAuto(t *testing.T) {
	got := ParseLang("auto")
	if got == "" {
		t.Fatal("auto must resolve to a real language")
	}
	if ParseLang("") != LangEN {
		t.Fatal("empty language must be English")
	}
}

func formatVerbCount(s string) int {
	return len(formatVerbs(s))
}

func formatVerbs(s string) []byte {
	var verbs []byte
	for i := 0; i < len(s); i++ {
		if s[i] != '%' {
			continue
		}
		if i+1 < len(s) && s[i+1] == '%' {
			i++
			continue
		}
		for j := i + 1; j < len(s); j++ {
			c := s[j]
			if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') {
				verbs = append(verbs, c)
				i = j
				break
			}
		}
	}
	return verbs
}

func dummyFormatArgs(format string) []any {
	verbs := formatVerbs(format)
	args := make([]any, len(verbs))
	for i, v := range verbs {
		switch v {
		case 'd', 'v', 'w', 'f', 'x', 'o', 'b', 'c', 'U':
			args[i] = 1
		default:
			args[i] = "x"
		}
	}
	return args
}
