package main

import "strings"

type Mode int

const (
	ModePersonal Mode = iota
	ModeWithAppData
	ModeAllNonSystem
)

func (m Mode) String() string {
	switch m {
	case ModePersonal:
		return "personal"
	case ModeWithAppData:
		return "appdata"
	case ModeAllNonSystem:
		return "all"
	default:
		return "appdata"
	}
}

func (m Mode) Title() string {
	switch m {
	case ModePersonal:
		return "个人资料（用户文档，不含 AppData）"
	case ModeWithAppData:
		return "个人资料 + 软件配置（推荐，含微信/浏览器等，排除缓存）"
	case ModeAllNonSystem:
		return "非系统全量（排除 Windows / 系统卷，不含已安装程序）"
	default:
		return ModeWithAppData.Title()
	}
}

func ParseMode(s string) Mode {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "personal", "p", "文档":
		return ModePersonal
	case "3", "all", "a", "全量":
		return ModeAllNonSystem
	default:
		return ModeWithAppData
	}
}

type Options struct {
	Mode                Mode
	IncludeProgramFiles bool
	SkipRegeneratable   bool
	DestAbs             string
	EmployeeName        string
	ExtraExclude        []string
}
