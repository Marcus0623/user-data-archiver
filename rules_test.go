package main

import (
	"runtime"
	"testing"
)

func TestSkipReason(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows paths")
	}
	opt := Options{Mode: ModeWithAppData, SkipRegeneratable: true, DestAbs: `D:\offboarding-archive\alice`}
	personal := Options{Mode: ModePersonal, SkipRegeneratable: true}
	all := Options{Mode: ModeAllNonSystem, IncludeProgramFiles: false}

	tests := []struct {
		path   string
		isDir  bool
		opt    Options
		skip   bool
		reason string
	}{
		{`C:\Windows`, true, opt, true, "system/OEM directory"},
		{`C:\$Recycle.Bin`, true, opt, true, "recycle bin / system volume"},
		{`C:\pagefile.sys`, false, opt, true, "paging/hiber file"},
		{`C:\Program Files`, true, opt, true, "installed programs / ProgramData"},
		{`C:\Program Files`, true, Options{IncludeProgramFiles: true}, false, ""},
		{`C:\Downloads\notes.zip`, false, opt, false, ""},
		{`C:\Users\zhang\Downloads\a.pdf`, false, opt, false, ""},
		{`C:\Users\zhang\Documents\WeChat Files\x`, true, opt, false, ""},
		{`C:\Users\zhang\AppData\Roaming\Tencent`, true, opt, false, ""},
		{`C:\Users\zhang\AppData\Roaming\Tencent`, true, personal, true, "AppData (excluded in personal mode)"},
		{`C:\Users\zhang\AppData\Local\Temp`, true, opt, true, "app cache"},
		{`C:\Users\zhang\Documents\Temp\notes.txt`, false, opt, false, ""},
		{`C:\Users\Default`, true, opt, true, "default user profile"},
		{`C:\Users\zhang\NTUSER.DAT`, false, opt, true, "user registry hive"},
		{`C:\Windows.old\Windows`, true, opt, true, "system/OEM directory"},
		{`C:\Windows.old\Users\zhang\Documents\a.txt`, false, opt, false, ""},
		{`C:\work\proj\node_modules`, true, opt, true, "regeneratable directory"},
		{`C:\work\proj\node_modules`, true, Options{Mode: ModeWithAppData, SkipRegeneratable: false}, false, ""},
		{`D:\offboarding-archive\alice\C\Users`, true, opt, true, "archive destination"},
		{`D:\offboarding-archive\alice`, true, opt, true, "archive destination"},
		{`D:\offboarding-archive\alice\_archive-report.txt`, false, opt, true, "archive destination"},
		{`D:\offboarding-archive\alice2`, true, opt, false, ""},
		{`C:\games`, true, Options{ExtraExclude: []string{"games"}}, true, "custom exclude"},
		{`C:\Intel`, true, all, true, "system/OEM directory"},
	}
	for _, tt := range tests {
		skip, _, reason := skipReason(tt.path, tt.isDir, tt.opt)
		if skip != tt.skip || (tt.skip && tt.reason != "" && reason != tt.reason) {
			t.Errorf("%s isDir=%v: skip=%v reason=%q, want skip=%v reason=%q",
				tt.path, tt.isDir, skip, reason, tt.skip, tt.reason)
		}
		if !tt.skip && reason != "" && skip {
			t.Errorf("%s unexpectedly skipped: %s", tt.path, reason)
		}
	}
}
