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
		return "Personal files (user documents, no AppData)"
	case ModeWithAppData:
		return "Personal files + app data (recommended; includes chat/browser data, excludes caches)"
	case ModeAllNonSystem:
		return "All non-system files (excludes Windows / system volumes, not installed programs)"
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
		return "Cut (move: delete originals only after the destination is verified)"
	}
	return "Copy (keep originals)"
}
