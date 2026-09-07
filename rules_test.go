package main

import (
	"runtime"
	"testing"
)

func TestSkipReason(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows 路径")
	}
	opt := Options{Mode: ModeWithAppData, SkipRegeneratable: true, DestAbs: `D:\离职归档\张三`}
	personal := Options{Mode: ModePersonal, SkipRegeneratable: true}
	all := Options{Mode: ModeAllNonSystem, IncludeProgramFiles: false}

	tests := []struct {
		path   string
		isDir  bool
		opt    Options
		skip   bool
		reason string
	}{
		{`C:\Windows`, true, opt, true, "系统/OEM 初始目录"},
		{`C:\$Recycle.Bin`, true, opt, true, "系统卷/回收站"},
		{`C:\pagefile.sys`, false, opt, true, "系统页面/休眠文件"},
		{`C:\Program Files`, true, opt, true, "已安装程序/ProgramData"},
		{`C:\Program Files`, true, Options{IncludeProgramFiles: true}, false, ""},
		{`C:\Downloads\资料.zip`, false, opt, false, ""},
		{`C:\Users\zhang\Downloads\a.pdf`, false, opt, false, ""},
		{`C:\Users\zhang\Documents\WeChat Files\x`, true, opt, false, ""},
		{`C:\Users\zhang\AppData\Roaming\Tencent`, true, opt, false, ""},
		{`C:\Users\zhang\AppData\Roaming\Tencent`, true, personal, true, "AppData（个人资料模式已排除）"},
		{`C:\Users\zhang\AppData\Local\Temp`, true, opt, true, "软件缓存目录"},
		{`C:\Users\zhang\Documents\Temp\notes.txt`, false, opt, false, ""},
		{`C:\Users\Default`, true, opt, true, "系统默认用户配置"},
		{`C:\Users\zhang\NTUSER.DAT`, false, opt, true, "用户注册表配置单元"},
		{`C:\Windows.old\Windows`, true, opt, true, "系统/OEM 初始目录"},
		{`C:\Windows.old\Users\zhang\Documents\a.txt`, false, opt, false, ""},
		{`C:\work\proj\node_modules`, true, opt, true, "可再生成目录"},
		{`C:\work\proj\node_modules`, true, Options{Mode: ModeWithAppData, SkipRegeneratable: false}, false, ""},
		{`D:\离职归档\张三\C\Users`, true, opt, true, "归档目标目录"},
		{`C:\games`, true, Options{ExtraExclude: []string{"games"}}, true, "自定义排除"},
		{`C:\Intel`, true, all, true, "系统/OEM 初始目录"},
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
