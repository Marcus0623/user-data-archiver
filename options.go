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
		return T("ModePersonal")
	case ModeWithAppData:
		return T("ModeAppData")
	case ModeAllNonSystem:
		return T("ModeAll")
	default:
		return ModeWithAppData.Title()
	}
}

func ParseMode(s string) Mode {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "personal", "p":
		return ModePersonal
	case "3", "all", "a":
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
	Cut                 bool
}

type TransferAction int

const (
	TransferCopy TransferAction = iota
	TransferCut
)

func (a TransferAction) Title() string {
	if a == TransferCut {
		return T("TransferCut")
	}
	return T("TransferCopy")
}
