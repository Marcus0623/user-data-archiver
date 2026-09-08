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

var (
	uiFont       = d.Font{Family: "Segoe UI", PointSize: 10}
	uiTitleFont  = d.Font{Family: "Segoe UI", PointSize: 16, Bold: true}
	uiButtonFont = d.Font{Family: "Segoe UI", PointSize: 11, Bold: true}
	uiMonoFont   = d.Font{Family: "Consolas", PointSize: 9}
	uiPage       = walk.RGB(244, 247, 251)
	uiHeader1    = walk.RGB(15, 76, 129)
	uiHeader2    = walk.RGB(30, 111, 176)
	uiWhite      = walk.RGB(255, 255, 255)
	uiSubHead    = walk.RGB(210, 230, 250)
	uiMuted      = walk.RGB(71, 85, 105)
	uiCopyAccent = walk.RGB(21, 128, 61)
	uiCutAccent  = walk.RGB(185, 28, 28)
	uiLogFg      = walk.RGB(15, 23, 42)
	uiLogBg      = walk.RGB(248, 250, 252)
)

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

func uiBanner(title, subtitle string) d.GradientComposite {
	return d.GradientComposite{
		Color1:   uiHeader1,
		Color2:   uiHeader2,
		Vertical: true,
		MinSize:  d.Size{Height: 92},
		Layout:   d.VBox{Margins: d.Margins{Left: 20, Top: 14, Right: 20, Bottom: 14}, Spacing: 4},
		Children: []d.Widget{
			d.Label{Text: title, Font: uiTitleFont, TextColor: uiWhite, Background: d.TransparentBrush{}},
			d.TextLabel{Text: subtitle, TextColor: uiSubHead, Background: d.TransparentBrush{}, MinSize: d.Size{Width: 520}},
		},
	}
}

func runForm(spec d.MainWindow, quitApp bool, onClosing, afterCreate func()) error {
	if spec.AssignTo == nil {
		var mw *walk.MainWindow
		spec.AssignTo = &mw
	}
	if spec.Font.Family == "" {
		spec.Font = uiFont
	}
	if spec.Background == nil {
		spec.Background = d.SolidColorBrush{Color: uiPage}
	}
	if spec.Icon == nil {
		spec.Icon = walk.IconInformation()
	}
	if err := spec.Create(); err != nil {
		return err
	}
	if afterCreate != nil {
		afterCreate()
	}
	w := *spec.AssignTo
	w.Closing().Attach(func(canceled *bool, reason walk.CloseReason) {
		if uiRestart {
			return
		}
		if onClosing != nil {
			onClosing()
		}
		if quitApp {
			walk.App().Exit(0)
		}
	})
	w.Run()
	return nil
}

func chooseTransferAction() (cut bool, ok bool) {
	for {
		uiRestart = false
		cut, ok = showChooser()
		if uiRestart {
			continue
		}
		return cut, ok
	}
}

func showChooser() (cut bool, ok bool) {
	var mw *walk.MainWindow
	var langBox *walk.ComboBox
	langReady := false
	if err := runForm(d.MainWindow{
		AssignTo:        &mw,
		Title:           T("AppName"),
		Icon:            walk.IconQuestion(),
		MinSize:         d.Size{Width: 600, Height: 720},
		Size:            d.Size{Width: 640, Height: 760},
		Layout:          d.VBox{MarginsZero: true, SpacingZero: true},
		DoubleBuffering: true,
		Children: []d.Widget{
			uiBanner(T("AppName"), T("ChooserSubtitle")),
			d.Composite{
				Layout: d.VBox{Margins: d.Margins{Left: 20, Top: 16, Right: 20, Bottom: 16}, Spacing: 12},
				Children: []d.Widget{
					d.Composite{
						Layout: d.HBox{MarginsZero: true, Spacing: 8},
						Children: []d.Widget{
							d.Label{Text: T("Language"), TextColor: uiMuted},
							d.ComboBox{
								AssignTo:     &langBox,
								Model:        langNames(),
								CurrentIndex: langIndex(currentLangCode()),
								MinSize:      d.Size{Width: 160},
								OnCurrentIndexChanged: func() {
									if !langReady || langBox == nil {
										return
									}
									i := langBox.CurrentIndex()
									if i < 0 || i >= len(allLangs) || allLangs[i] == currentLangCode() {
										return
									}
									setLang(allLangs[i])
									uiRestart = true
									mw.Close()
								},
							},
							d.HSpacer{},
						},
					},
					d.GroupBox{
						Title:  T("CopyGroup"),
						Font:   uiFont,
						Layout: d.VBox{Margins: d.Margins{Left: 12, Top: 10, Right: 12, Bottom: 12}, Spacing: 8},
						Children: []d.Widget{
							d.TextLabel{
								Text:      T("CopyHelp"),
								TextColor: uiMuted,
								MinSize:   d.Size{Width: 520},
							},
							d.PushButton{
								Text:    T("CopyButton"),
								Font:    uiButtonFont,
								MinSize: d.Size{Height: 40},
								OnClicked: func() {
									cut = false
									ok = true
									mw.Close()
								},
							},
						},
					},
					d.GroupBox{
						Title:  T("CutGroup"),
						Font:   uiFont,
						Layout: d.VBox{Margins: d.Margins{Left: 12, Top: 10, Right: 12, Bottom: 12}, Spacing: 8},
						Children: []d.Widget{
							d.TextLabel{
								Text:      T("CutHelp"),
								TextColor: uiCutAccent,
								MinSize:   d.Size{Width: 520},
							},
							d.PushButton{
								Text:    T("CutButton"),
								Font:    uiButtonFont,
								MinSize: d.Size{Height: 40},
								OnClicked: func() {
									cut = true
									ok = true
									mw.Close()
								},
							},
						},
					},
					d.Composite{
						Layout: d.HBox{MarginsZero: true},
						Children: []d.Widget{
							d.HSpacer{},
							d.PushButton{Text: T("Cancel"), MinSize: d.Size{Width: 100, Height: 32}, OnClicked: func() {
								ok = false
								mw.Close()
							}},
						},
					},
				},
			},
		},
	}, false, nil, func() {
		if langBox != nil {
			_ = langBox.SetCurrentIndex(langIndex(currentLangCode()))
		}
		langReady = true
	}); err != nil {
		showErrorDialog(T("CouldNotOpenWindow", err.Error()))
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
	for {
		uiRestart = false
		prefill = runMainWindowOnce(prefill)
		if !uiRestart {
			return
		}
	}
}

func runMainWindowOnce(prefill guiPrefill) guiPrefill {
	var (
		mw            *walk.MainWindow
		nameEdit      *walk.LineEdit
		srcEdit       *walk.LineEdit
		destEdit      *walk.LineEdit
		excludeEdit   *walk.LineEdit
		driveLabel    *walk.TextLabel
		testLabel     *walk.Label
		modeBox       *walk.ComboBox
		pfCheck       *walk.CheckBox
		regenCheck    *walk.CheckBox
		previewChk    *walk.CheckBox
		logEdit       *walk.TextEdit
		progressBar   *walk.ProgressBar
		progressLbl   *walk.Label
		testBtn       *walk.PushButton
		previewBtn    *walk.PushButton
		startBtn      *walk.PushButton
		stopBtn       *walk.PushButton
		srcBrowseBtn  *walk.PushButton
		destBrowseBtn *walk.PushButton
		statusJob     *walk.StatusBarItem
		langBox       *walk.ComboBox
	)
	langReady := false

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
		mu          sync.Mutex
		cancel      context.CancelFunc
		jobBusy     bool
		autoDest    = dest == filepath.Join(defaultDestParent, sanitizeEmployeeName(name))
		syncingDest bool
	)
	alive := func() bool {
		return mw != nil && !mw.IsDisposed()
	}
	stopJob := func() {
		mu.Lock()
		defer mu.Unlock()
		if cancel != nil {
			cancel()
		}
	}
	setBusy := func(busy bool) {
		jobBusy = busy
		testBtn.SetEnabled(!busy)
		previewBtn.SetEnabled(!busy)
		if previewChk != nil {
			startBtn.SetEnabled(!busy && !previewChk.Checked())
		} else {
			startBtn.SetEnabled(!busy)
		}
		stopBtn.SetEnabled(busy)
		nameEdit.SetEnabled(!busy)
		srcEdit.SetEnabled(!busy)
		destEdit.SetEnabled(!busy)
		if excludeEdit != nil {
			excludeEdit.SetEnabled(!busy)
		}
		modeBox.SetEnabled(!busy)
		pfCheck.SetEnabled(!busy)
		regenCheck.SetEnabled(!busy)
		previewChk.SetEnabled(!busy)
		if langBox != nil {
			langBox.SetEnabled(!busy)
		}
		if srcBrowseBtn != nil {
			srcBrowseBtn.SetEnabled(!busy)
		}
		if destBrowseBtn != nil {
			destBrowseBtn.SetEnabled(!busy)
		}
		if statusJob != nil {
			if busy {
				statusJob.SetText(T("Working"))
			} else {
				statusJob.SetText(T("Ready"))
			}
		}
		if progressBar != nil && !busy {
			_ = progressBar.SetMarqueeMode(false)
			progressBar.SetValue(0)
			if progressLbl != nil {
				_ = progressLbl.SetText(T("Idle"))
			}
		}
		if progressBar != nil && busy {
			_ = progressBar.SetMarqueeMode(true)
			if progressLbl != nil {
				_ = progressLbl.SetText(T("Scanning"))
			}
		}
	}
	appendLog := func(format string, args ...any) {
		line := fmt.Sprintf(format, args...)
		if !alive() {
			return
		}
		mw.Synchronize(func() {
			if !alive() || logEdit == nil {
				return
			}
			logEdit.AppendText(line + "\r\n")
		})
	}
	refreshDrive := func() {
		if destEdit == nil || driveLabel == nil {
			return
		}
		_ = driveLabel.SetText(destDriveInfo(destEdit.Text()))
	}
	browseFolder := func(title, start string) (string, bool) {
		dlg := new(walk.FileDialog)
		dlg.Title = title
		dlg.FilePath = start
		ok, err := dlg.ShowBrowseFolder(mw)
		if err != nil || !ok {
			return "", false
		}
		return dlg.FilePath, true
	}
	runTest := func() {
		path := strings.TrimSpace(destEdit.Text())
		if path == "" {
			_ = testLabel.SetText(T("ChooseDestFirst"))
			return
		}
		path = normalizeAbs(path)
		_ = destEdit.SetText(path)
		refreshDrive()
		_ = testLabel.SetText(T("TestingRW"))
		if err := probeDestWritable(path); err != nil {
			_ = testLabel.SetText(T("TestFailed", err.Error()))
			appendLog("%s", T("LogTestFailed", err))
			walk.MsgBox(mw, T("ReadWriteTest"), err.Error(), walk.MsgBoxIconError)
			return
		}
		_ = testLabel.SetText(T("TestOKWrote", writeProbeFileName))
		appendLog("%s", T("LogTestOK", path, destDriveInfo(path)))
		walk.MsgBox(mw, T("ReadWriteTest"), T("DestWritableMsg", path, destDriveInfo(path)), walk.MsgBoxIconInformation)
	}
	startJob := func(dryRun bool) {
		employee := sanitizeEmployeeName(strings.TrimSpace(nameEdit.Text()))
		if employee == "" {
			employee = "archive"
			_ = nameEdit.SetText(employee)
		}
		dest := strings.TrimSpace(destEdit.Text())
		if dest == "" {
			walk.MsgBox(mw, T("AppName"), T("ChooseDest"), walk.MsgBoxIconWarning)
			return
		}
		dest = normalizeAbs(dest)
		_ = destEdit.SetText(dest)
		srcText := strings.TrimSpace(srcEdit.Text())
		roots := splitRoots(srcText)
		if len(roots) == 0 {
			walk.MsgBox(mw, T("AppName"), T("EnterSource"), walk.MsgBoxIconWarning)
			return
		}
		if err := probeDestWritable(dest); err != nil {
			_ = testLabel.SetText(T("TestFailed", err.Error()))
			walk.MsgBox(mw, T("ReadWriteTest"), T("FixDestThenTest", err.Error()), walk.MsgBoxIconError)
			return
		}
		_ = testLabel.SetText(T("TestOKWritable"))

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
			ExtraExclude:        splitRoots(excludeEdit.Text()),
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
				if alive() {
					mw.Synchronize(func() {
						if alive() {
							setBusy(false)
						}
					})
				}
			}()
			err := runArchiveJob(ctx, roots, dest, employee, opt, dryRun, false, jobUI{
				log: appendLog,
				progress: func(doneFiles, totalFiles, doneBytes, totalBytes int64, src string) {
					pct := ""
					if totalBytes > 0 {
						pct = fmt.Sprintf(" %.1f%%", float64(doneBytes)*100/float64(totalBytes))
					}
					appendLog("%s", T("LogProgress", doneFiles, totalFiles, pct, formatBytes(doneBytes), shortenPath(src, 80)))
					if !alive() {
						return
					}
					mw.Synchronize(func() {
						if !alive() || progressBar == nil {
							return
						}
						_ = progressBar.SetMarqueeMode(false)
						if totalBytes > 0 {
							progressBar.SetRange(0, 1000)
							progressBar.SetValue(int(doneBytes * 1000 / totalBytes))
						}
						if progressLbl != nil {
							_ = progressLbl.SetText(T("ProgressFiles", doneFiles, totalFiles, formatBytes(doneBytes)))
						}
						if statusJob != nil {
							statusJob.SetText(T("CopyingStatus"))
						}
					})
				},
				ask: func(question string, defYes bool) bool {
					if !alive() {
						return false
					}
					ch := make(chan bool, 1)
					mw.Synchronize(func() {
						if !alive() {
							ch <- false
							return
						}
						btn := walk.MsgBoxYesNo | walk.MsgBoxIconQuestion
						if defYes {
							btn |= walk.MsgBoxDefButton1
						} else {
							btn |= walk.MsgBoxDefButton2
						}
						ch <- walk.MsgBox(mw, T("AppName"), question, btn) == walk.DlgCmdYes
					})
					return <-ch
				},
			})
			if err != nil {
				appendLog("%s", T("ErrorPrefix", err))
				if alive() {
					mw.Synchronize(func() {
						if alive() {
							walk.MsgBox(mw, T("AppName"), err.Error(), walk.MsgBoxIconError)
						}
					})
				}
			}
		}()
	}

	rights := T("RightsUser")
	if isAdmin() {
		rights = T("RightsAdmin")
	}
	xferColor := uiCopyAccent
	transferLine := TransferCopy.Title()
	winTitle := T("WinTitleCopy")
	startLabel := T("StartCopy")
	previewHint := T("PreviewOnlyCopy")
	bannerSub := T("BannerCopy", computerName(), rights)
	headerIcon := walk.IconInformation()
	if prefill.Cut {
		xferColor = uiCutAccent
		transferLine = TransferCut.Title()
		winTitle = T("WinTitleCut")
		startLabel = T("StartCut")
		previewHint = T("PreviewOnlyCut")
		bannerSub = T("BannerCut", computerName(), rights)
		headerIcon = walk.IconWarning()
	}

	savePrefill := func() {
		if nameEdit != nil {
			prefill.Name = nameEdit.Text()
		}
		if srcEdit != nil {
			prefill.Src = srcEdit.Text()
		}
		if destEdit != nil {
			prefill.Dest = destEdit.Text()
		}
		if excludeEdit != nil {
			prefill.Exclude = excludeEdit.Text()
		}
		if pfCheck != nil {
			prefill.IncludeProgramFiles = pfCheck.Checked()
		}
		if regenCheck != nil {
			prefill.IncludeRegen = regenCheck.Checked()
		}
		if previewChk != nil {
			prefill.DryRun = previewChk.Checked()
		}
		if modeBox != nil {
			switch modeBox.CurrentIndex() {
			case 0:
				prefill.Mode = ModePersonal.String()
			case 2:
				prefill.Mode = ModeAllNonSystem.String()
			default:
				prefill.Mode = ModeWithAppData.String()
			}
		}
		prefill.ActionChosen = true
	}

	if err := runForm(d.MainWindow{
		AssignTo:        &mw,
		Title:           winTitle,
		Icon:            headerIcon,
		MinSize:         d.Size{Width: 820, Height: 680},
		Size:            d.Size{Width: 880, Height: 740},
		Layout:          d.VBox{MarginsZero: true, Spacing: 0},
		DoubleBuffering: true,
		StatusBarItems: []d.StatusBarItem{
			{Text: T("StatusPC", computerName()), Width: 220},
			{Text: rights, Width: 320},
			{AssignTo: &statusJob, Text: T("Ready"), Width: 160},
		},
		Children: []d.Widget{
			uiBanner(winTitle, bannerSub),
			d.Composite{
				Layout: d.VBox{Margins: d.Margins{Left: 16, Top: 12, Right: 16, Bottom: 12}, Spacing: 10},
				Children: []d.Widget{
					d.GroupBox{
						Title:  T("Paths"),
						Layout: d.Grid{Columns: 3, Spacing: 8},
						Children: []d.Widget{
							d.Label{Text: T("PersonName"), TextColor: uiMuted},
							d.LineEdit{
								AssignTo:   &nameEdit,
								Text:       name,
								CueBanner:  T("PersonCue"),
								ColumnSpan: 2,
								OnTextChanged: func() {
									if !autoDest || destEdit == nil {
										return
									}
									n := sanitizeEmployeeName(strings.TrimSpace(nameEdit.Text()))
									if n == "" {
										n = "archive"
									}
									syncingDest = true
									_ = destEdit.SetText(filepath.Join(defaultDestParent, n))
									syncingDest = false
								},
							},
							d.Label{Text: T("Source"), TextColor: uiMuted},
							d.LineEdit{AssignTo: &srcEdit, Text: src, CueBanner: T("SourceCue")},
							d.PushButton{AssignTo: &srcBrowseBtn, Text: T("Browse"), MinSize: d.Size{Width: 96}, OnClicked: func() {
								picked, ok := browseFolder(T("BrowseSource"), initialBrowsePath(srcEdit.Text()))
								if !ok {
									return
								}
								_ = srcEdit.SetText(mergePickedRoot(srcEdit.Text(), picked))
							}},
							d.TextLabel{
								Text:       "ⓘ  " + T("SourceHelp"),
								TextColor:  uiMuted,
								ColumnSpan: 3,
								MinSize:    d.Size{Width: 640},
							},
							d.Label{Text: T("SaveTo"), TextColor: uiMuted},
							d.LineEdit{
								AssignTo:  &destEdit,
								Text:      dest,
								CueBanner: T("DestCue"),
								OnTextChanged: func() {
									if !syncingDest {
										autoDest = false
									}
									refreshDrive()
								},
							},
							d.PushButton{AssignTo: &destBrowseBtn, Text: T("Browse"), MinSize: d.Size{Width: 96}, OnClicked: func() {
								picked, ok := browseFolder(T("BrowseDest"), destEdit.Text())
								if !ok {
									return
								}
								autoDest = false
								_ = destEdit.SetText(picked)
								refreshDrive()
							}},
							d.Label{Text: T("Exclude"), TextColor: uiMuted},
							d.LineEdit{
								AssignTo:    &excludeEdit,
								Text:        prefill.Exclude,
								CueBanner:   T("ExcludeCue"),
								ColumnSpan:  2,
								ToolTipText: T("ExcludeTip"),
							},
							d.Label{Text: T("Drive"), TextColor: uiMuted},
							d.TextLabel{AssignTo: &driveLabel, Text: destDriveInfo(dest), ColumnSpan: 2, MinSize: d.Size{Width: 480}, TextColor: uiMuted},
							d.PushButton{AssignTo: &testBtn, Text: T("ReadWriteTest"), MinSize: d.Size{Height: 32}, OnClicked: runTest},
							d.Label{AssignTo: &testLabel, Text: T("TestNotYet"), ColumnSpan: 2, TextColor: uiMuted},
						},
					},
					d.GroupBox{
						Title:  T("Options"),
						Layout: d.VBox{Margins: d.Margins{Left: 10, Top: 8, Right: 10, Bottom: 10}, Spacing: 6},
						Children: []d.Widget{
							d.Composite{
								Layout: d.HBox{MarginsZero: true, Spacing: 8},
								Children: []d.Widget{
									d.Label{Text: T("Language"), TextColor: uiMuted},
									d.ComboBox{
										AssignTo:     &langBox,
										Model:        langNames(),
										CurrentIndex: langIndex(currentLangCode()),
										MinSize:      d.Size{Width: 160},
										OnCurrentIndexChanged: func() {
											if !langReady || langBox == nil {
												return
											}
											i := langBox.CurrentIndex()
											if i < 0 || i >= len(allLangs) || allLangs[i] == currentLangCode() {
												return
											}
											savePrefill()
											setLang(allLangs[i])
											uiRestart = true
											mw.Close()
										},
									},
									d.HSpacer{},
								},
							},
							d.Label{Text: T("TransferPrefix", transferLine), Font: d.Font{Family: "Segoe UI", PointSize: 10, Bold: true}, TextColor: xferColor},
							d.ComboBox{
								AssignTo:     &modeBox,
								Model:        []string{ModePersonal.Title(), ModeWithAppData.Title(), ModeAllNonSystem.Title()},
								CurrentIndex: modeIndex,
								MinSize:      d.Size{Height: 26},
							},
							d.CheckBox{AssignTo: &pfCheck, Text: T("IncludePF"), Checked: prefill.IncludeProgramFiles},
							d.CheckBox{AssignTo: &regenCheck, Text: T("IncludeRegen"), Checked: prefill.IncludeRegen},
							d.CheckBox{
								AssignTo: &previewChk,
								Text:     previewHint,
								Checked:  prefill.DryRun,
								OnCheckedChanged: func() {
									if startBtn == nil || previewChk == nil {
										return
									}
									startBtn.SetEnabled(!jobBusy && !previewChk.Checked())
								},
							},
						},
					},
					d.GroupBox{
						Title:         T("Activity"),
						StretchFactor: 1,
						Layout:        d.VBox{Margins: d.Margins{Left: 10, Top: 8, Right: 10, Bottom: 10}, Spacing: 6},
						Children: []d.Widget{
							d.Label{AssignTo: &progressLbl, Text: T("Idle"), TextColor: uiMuted},
							d.ProgressBar{AssignTo: &progressBar, MinSize: d.Size{Height: 18}, MaxValue: 1000},
							d.TextEdit{
								AssignTo:      &logEdit,
								ReadOnly:      true,
								VScroll:       true,
								HScroll:       true,
								StretchFactor: 1,
								Font:          uiMonoFont,
								TextColor:     uiLogFg,
								Background:    d.SolidColorBrush{Color: uiLogBg},
								MinSize:       d.Size{Height: 160},
							},
						},
					},
					d.Composite{
						Layout: d.HBox{Spacing: 8},
						Children: []d.Widget{
							d.PushButton{AssignTo: &previewBtn, Text: T("Preview"), MinSize: d.Size{Width: 110, Height: 36}, OnClicked: func() { startJob(true) }},
							d.PushButton{AssignTo: &startBtn, Text: startLabel, Font: uiButtonFont, MinSize: d.Size{Width: 130, Height: 36}, Enabled: !prefill.DryRun, OnClicked: func() {
								if previewChk != nil && previewChk.Checked() {
									startJob(true)
									return
								}
								startJob(false)
							}},
							d.PushButton{AssignTo: &stopBtn, Text: T("Stop"), MinSize: d.Size{Width: 90, Height: 36}, Enabled: false, OnClicked: stopJob},
							d.HSpacer{},
						},
					},
				},
			},
		},
	}, true, stopJob, func() {
		if langBox != nil {
			_ = langBox.SetCurrentIndex(langIndex(currentLangCode()))
		}
		langReady = true
	}); err != nil {
		showErrorDialog(T("CouldNotOpenWindow", err.Error()))
	}
	return prefill
}
