package main

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/lxn/walk"
	d "github.com/lxn/walk/declarative"
)

func init() {
	runtime.LockOSThread()
}

type guiPrefill struct {
	Name                string
	Dest                string
	Src                 string
	Mode                string
	Exclude             string
	IncludeProgramFiles bool
	IncludeRegen        bool
	DryRun              bool
	Cut                 bool
	ActionChosen        bool
}

func chooseTransferAction() (cut bool, ok bool) {
	var mw *walk.MainWindow
	if _, err := (d.MainWindow{
		AssignTo: &mw,
		Title:    "User Data Archiver",
		MinSize:  d.Size{Width: 480, Height: 300},
		Size:     d.Size{Width: 520, Height: 320},
		Layout:   d.VBox{Margins: d.Margins{Left: 16, Top: 16, Right: 16, Bottom: 16}, Spacing: 10},
		Children: []d.Widget{
			d.Label{
				Text: "How should files be transferred?",
				Font: d.Font{PointSize: 12, Bold: true},
			},
			d.Label{Text: "This choice applies to the next window. You can still preview before any files are changed."},
			d.VSpacer{Size: 8},
			d.PushButton{
				Text:    "Copy files",
				MinSize: d.Size{Height: 40},
				OnClicked: func() {
					cut = false
					ok = true
					mw.Close()
				},
			},
			d.Label{Text: "Save to the destination and keep the originals on the source disk."},
			d.VSpacer{Size: 4},
			d.PushButton{
				Text:    "Cut files (move)",
				MinSize: d.Size{Height: 40},
				OnClicked: func() {
					cut = true
					ok = true
					mw.Close()
				},
			},
			d.Label{Text: "Save to the destination, then delete each original only after the destination is flushed and the contents match. This cannot be undone."},
			d.VSpacer{},
			d.Composite{
				Layout: d.HBox{},
				Children: []d.Widget{
					d.HSpacer{},
					d.PushButton{Text: "Cancel", OnClicked: func() {
						ok = false
						mw.Close()
					}},
				},
			},
		},
	}.Run()); err != nil {
		return false, false
	}
	return cut, ok
}

func runGUI(prefill guiPrefill) {
	if !prefill.ActionChosen {
		cut, ok := chooseTransferAction()
		if !ok {
			return
		}
		prefill.Cut = cut
	}
	runMainWindow(prefill)
}

func runMainWindow(prefill guiPrefill) {
	var (
		mw         *walk.MainWindow
		nameEdit   *walk.LineEdit
		srcEdit    *walk.LineEdit
		destEdit   *walk.LineEdit
		driveLabel *walk.Label
		testLabel  *walk.Label
		modeBox    *walk.ComboBox
		pfCheck    *walk.CheckBox
		regenCheck *walk.CheckBox
		previewChk *walk.CheckBox
		logEdit    *walk.TextEdit
		testBtn    *walk.PushButton
		previewBtn *walk.PushButton
		startBtn   *walk.PushButton
		stopBtn    *walk.PushButton
	)

	name := prefill.Name
	if name == "" {
		name = defaultName()
	}
	dest := prefill.Dest
	if dest == "" {
		dest = filepath.Join(defaultDestParent, sanitizeEmployeeName(name))
	}
	src := prefill.Src
	if src == "" {
		src = "C:"
	}
	modeIndex := 1
	switch ParseMode(prefill.Mode) {
	case ModePersonal:
		modeIndex = 0
	case ModeAllNonSystem:
		modeIndex = 2
	}

	var (
		mu     sync.Mutex
		cancel context.CancelFunc
	)
	setBusy := func(busy bool) {
		testBtn.SetEnabled(!busy)
		previewBtn.SetEnabled(!busy)
		startBtn.SetEnabled(!busy)
		stopBtn.SetEnabled(busy)
		nameEdit.SetEnabled(!busy)
		srcEdit.SetEnabled(!busy)
		destEdit.SetEnabled(!busy)
		modeBox.SetEnabled(!busy)
		pfCheck.SetEnabled(!busy)
		regenCheck.SetEnabled(!busy)
		previewChk.SetEnabled(!busy)
	}
	appendLog := func(format string, args ...any) {
		line := fmt.Sprintf(format, args...)
		mw.Synchronize(func() {
			logEdit.AppendText(line + "\r\n")
		})
	}
	refreshDrive := func() {
		if destEdit == nil || driveLabel == nil {
			return
		}
		_ = driveLabel.SetText(destDriveInfo(destEdit.Text()))
	}
	browseFolder := func(title, current string) string {
		dlg := new(walk.FileDialog)
		dlg.Title = title
		dlg.FilePath = current
		ok, err := dlg.ShowBrowseFolder(mw)
		if err != nil || !ok {
			return current
		}
		return dlg.FilePath
	}
	runTest := func() {
		path := strings.TrimSpace(destEdit.Text())
		if path == "" {
			_ = testLabel.SetText("Choose a destination folder first.")
			return
		}
		path = normalizeAbs(path)
		_ = destEdit.SetText(path)
		refreshDrive()
		_ = testLabel.SetText("Testing read/write...")
		if err := probeDestWritable(path); err != nil {
			_ = testLabel.SetText("FAILED: " + err.Error())
			appendLog("Read/write test FAILED: %v", err)
			walk.MsgBox(mw, "Read/Write Test", err.Error(), walk.MsgBoxIconError)
			return
		}
		_ = testLabel.SetText("OK: created, wrote, read, and deleted " + writeProbeFileName)
		appendLog("Read/write test OK for %s (%s)", path, destDriveInfo(path))
		walk.MsgBox(mw, "Read/Write Test", "Destination is writable:\n"+path+"\n\n"+destDriveInfo(path), walk.MsgBoxIconInformation)
	}
	startJob := func(dryRun bool) {
		employee := sanitizeEmployeeName(strings.TrimSpace(nameEdit.Text()))
		if employee == "" {
			employee = "archive"
			_ = nameEdit.SetText(employee)
		}
		dest := strings.TrimSpace(destEdit.Text())
		if dest == "" {
			walk.MsgBox(mw, "User Data Archiver", "Choose a destination folder.", walk.MsgBoxIconWarning)
			return
		}
		dest = normalizeAbs(dest)
		_ = destEdit.SetText(dest)
		srcText := strings.TrimSpace(srcEdit.Text())
		roots := splitRoots(srcText)
		if len(roots) == 0 {
			walk.MsgBox(mw, "User Data Archiver", "Enter a source path such as C:.", walk.MsgBoxIconWarning)
			return
		}
		if err := probeDestWritable(dest); err != nil {
			_ = testLabel.SetText("FAILED: " + err.Error())
			walk.MsgBox(mw, "Read/Write Test", err.Error()+"\n\nFix the destination, then click Read/Write Test.", walk.MsgBoxIconError)
			return
		}
		_ = testLabel.SetText("OK: destination is writable")

		mode := ModeWithAppData
		switch modeBox.CurrentIndex() {
		case 0:
			mode = ModePersonal
		case 2:
			mode = ModeAllNonSystem
		}
		opt := Options{
			Mode:                mode,
			IncludeProgramFiles: pfCheck.Checked(),
			SkipRegeneratable:   !regenCheck.Checked(),
			DestAbs:             dest,
			EmployeeName:        employee,
			ExtraExclude:        splitRoots(prefill.Exclude),
			Cut:                 prefill.Cut,
		}

		ctx, jobCancel := context.WithCancel(context.Background())
		mu.Lock()
		cancel = jobCancel
		mu.Unlock()
		setBusy(true)
		appendLog("----")
		go func() {
			defer func() {
				mu.Lock()
				cancel = nil
				mu.Unlock()
				mw.Synchronize(func() { setBusy(false) })
			}()
			err := runArchiveJob(ctx, roots, dest, employee, opt, dryRun, false, jobUI{
				log: appendLog,
				progress: func(doneFiles, totalFiles, doneBytes, totalBytes int64, src string) {
					pct := ""
					if totalBytes > 0 {
						pct = fmt.Sprintf(" %.1f%%", float64(doneBytes)*100/float64(totalBytes))
					}
					appendLog("  %d/%d files%s  %s  %s", doneFiles, totalFiles, pct, formatBytes(doneBytes), shortenPath(src, 80))
				},
				ask: func(question string, defYes bool) bool {
					ch := make(chan bool, 1)
					mw.Synchronize(func() {
						btn := walk.MsgBoxYesNo | walk.MsgBoxIconQuestion
						if defYes {
							btn |= walk.MsgBoxDefButton1
						} else {
							btn |= walk.MsgBoxDefButton2
						}
						ch <- walk.MsgBox(mw, "User Data Archiver", question, btn) == walk.DlgCmdYes
					})
					return <-ch
				},
			})
			if err != nil {
				appendLog("Error: %v", err)
				mw.Synchronize(func() {
					walk.MsgBox(mw, "User Data Archiver", err.Error(), walk.MsgBoxIconError)
				})
			}
		}()
	}

	rights := "not administrator (cannot read other user profiles unless you re-run as administrator)"
	if isAdmin() {
		rights = "administrator (can read other user profiles)"
	}
	transferLine := TransferCopy.Title()
	winTitle := "User Data Archiver — Copy"
	startLabel := "Start Copy"
	previewHint := "Preview only (scan, do not copy)"
	if prefill.Cut {
		transferLine = TransferCut.Title()
		winTitle = "User Data Archiver — Cut"
		startLabel = "Start Cut"
		previewHint = "Preview only (scan, do not cut)"
	}

	if _, err := (d.MainWindow{
		AssignTo: &mw,
		Title:    winTitle,
		MinSize:  d.Size{Width: 740, Height: 580},
		Size:     d.Size{Width: 780, Height: 640},
		Layout:   d.VBox{MarginsZero: false},
		Children: []d.Widget{
			d.Label{Text: "Computer: " + computerName() + "    Rights: " + rights},
			d.Label{Text: "Transfer: " + transferLine},
			d.Composite{
				Layout: d.Grid{Columns: 3, Spacing: 6},
				Children: []d.Widget{
					d.Label{Text: "Person name:"},
					d.LineEdit{AssignTo: &nameEdit, Text: name, ColumnSpan: 2},
					d.Label{Text: "Source:"},
					d.LineEdit{AssignTo: &srcEdit, Text: src},
					d.PushButton{Text: "Browse...", OnClicked: func() {
						_ = srcEdit.SetText(browseFolder("Select source folder", srcEdit.Text()))
					}},
					d.Label{Text: "Save to:"},
					d.LineEdit{
						AssignTo: &destEdit,
						Text:     dest,
						OnTextChanged: func() {
							refreshDrive()
						},
					},
					d.PushButton{Text: "Browse...", OnClicked: func() {
						_ = destEdit.SetText(browseFolder("Select destination folder", destEdit.Text()))
						refreshDrive()
					}},
				},
			},
			d.Label{AssignTo: &driveLabel, Text: destDriveInfo(dest)},
			d.Composite{
				Layout: d.HBox{},
				Children: []d.Widget{
					d.PushButton{AssignTo: &testBtn, Text: "Read/Write Test", OnClicked: runTest},
					d.Label{AssignTo: &testLabel, Text: "Not tested yet. Choose Save to, then test."},
					d.HSpacer{},
				},
			},
			d.Label{Text: "Mode:"},
			d.ComboBox{
				AssignTo:     &modeBox,
				Model:        []string{ModePersonal.Title(), ModeWithAppData.Title(), ModeAllNonSystem.Title()},
				CurrentIndex: modeIndex,
			},
			d.CheckBox{AssignTo: &pfCheck, Text: "Include Program Files / ProgramData", Checked: prefill.IncludeProgramFiles},
			d.CheckBox{AssignTo: &regenCheck, Text: "Include node_modules and other regeneratable folders", Checked: prefill.IncludeRegen},
			d.CheckBox{AssignTo: &previewChk, Text: previewHint, Checked: prefill.DryRun},
			d.Label{Text: "Log:"},
			d.TextEdit{AssignTo: &logEdit, ReadOnly: true, VScroll: true, MaxSize: d.Size{Height: 220}},
			d.Composite{
				Layout: d.HBox{},
				Children: []d.Widget{
					d.PushButton{AssignTo: &previewBtn, Text: "Preview", OnClicked: func() { startJob(true) }},
					d.PushButton{AssignTo: &startBtn, Text: startLabel, OnClicked: func() {
						startJob(previewChk.Checked())
					}},
					d.PushButton{AssignTo: &stopBtn, Text: "Stop", Enabled: false, OnClicked: func() {
						mu.Lock()
						defer mu.Unlock()
						if cancel != nil {
							cancel()
						}
					}},
					d.HSpacer{},
				},
			},
		},
	}.Run()); err != nil {
		fatal("window failed: " + err.Error())
	}
}
