package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type jobUI struct {
	log      func(format string, args ...any)
	progress func(doneFiles, totalFiles, doneBytes, totalBytes int64, src string)
	ask      func(question string, defYes bool) bool
}

func runArchiveJob(ctx context.Context, roots []string, dest, employee string, opt Options, dryRun bool, skipConfirm bool, ui jobUI) error {
	if ui.log == nil {
		ui.log = func(string, ...any) {}
	}
	if ui.ask == nil {
		ui.ask = func(string, bool) bool { return true }
	}

	ui.log("%s", T("JobKeepFolders"))
	ui.log("  C:\\Users\\%s\\Downloads\\file.pdf", employee)
	ui.log("    -> %s", filepath.Join(dest, "C", "Users", employee, "Downloads", "file.pdf"))
	ui.log("%s", T("JobMode", opt.Mode.Title()))
	ui.log("%s", T("JobTransfer", transferTitle(opt.Cut)))
	ui.log("%s", T("JobSource", joinComma(roots)))
	ui.log("%s", T("JobDest", dest))
	for _, r := range roots {
		abs, err := expandRoot(r)
		if err != nil {
			continue
		}
		if isBaseOf(abs, dest) {
			ui.log("%s", T("JobDestInsideSource", abs))
		}
	}
	ui.log("%s", T("JobNoteSkip"))
	ui.log("%s", T("JobTipOneDrive"))
	ui.log("%s", T("JobScanning"))

	started := time.Now()
	var scanErrs int
	sum, err := Scan(ctx, roots, opt, func(src string, e error) {
		scanErrs++
		if scanErrs <= 20 {
			ui.log("%s", T("JobScanSkip", src, e))
		}
	})
	if err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("%s", T("JobInterrupted"))
		}
		return fmt.Errorf("%s: %w", T("JobScanFailed"), err)
	}

	ui.log("%s", T("JobPreview"))
	for _, folder := range sum.sortedFolders() {
		if folder.Files == 0 && folder.Bytes == 0 {
			continue
		}
		ui.log("%s", T("JobFolderLine", folder.Key, folder.Files, formatBytes(folder.Bytes)))
	}
	totalLine := T("JobTotal", sum.Files, formatBytes(sum.Bytes))
	if sum.Errors > 0 {
		totalLine += T("JobUnreadable", sum.Errors)
	}
	ui.log("%s", totalLine)

	if vol := filepath.VolumeName(dest); vol != "" {
		if free, _, err := diskFreeBytes(vol + `\`); err == nil {
			ui.log("%s", T("JobFreeSpace", formatBytes(int64(free))))
			need := uint64(sum.Bytes) + uint64(sum.Bytes/20) + 64*1024*1024
			if free < need {
				ui.log("%s", T("JobLowSpace"))
				if !skipConfirm && !ui.ask(T("JobLowSpaceAsk"), false) {
					ui.log("%s", T("JobCancelled"))
					return nil
				}
			}
		}
	}

	if dryRun {
		if err := os.MkdirAll(longPath(dest), 0o755); err == nil {
			_ = writeReport(dest, computerName(), employee, roots, opt, sum, nil, "", started)
			ui.log("%s", T("JobPreviewReport", filepath.Join(dest, reportFileName)))
		}
		ui.log("%s", T("JobDryRunDone"))
		return nil
	}

	confirm := T("JobConfirmCopy")
	defYes := true
	if opt.Cut {
		confirm = T("JobConfirmCut")
		defYes = false
	}
	if !skipConfirm && !ui.ask(confirm, defYes) {
		ui.log("%s", T("JobCancelled"))
		return nil
	}

	if err := os.MkdirAll(longPath(dest), 0o755); err != nil {
		return fmt.Errorf("%s: %w", T("JobCreateDest"), err)
	}

	ui.log("%s", T("JobCopying"))
	if opt.Cut {
		ui.log("%s", T("JobCutMode"))
	}
	lastPrint := time.Now()
	fails := make([][]string, 0, 64)
	res, err := Archive(ctx, roots, opt, dest, func(doneFiles, doneBytes int64, src string) {
		if ui.progress != nil && time.Since(lastPrint) >= 200*time.Millisecond {
			lastPrint = time.Now()
			ui.progress(doneFiles, sum.Files, doneBytes, sum.Bytes, src)
		}
	}, func(src, dst string, e error) {
		if len(fails) < 5000 {
			fails = append(fails, []string{src, dst, e.Error()})
		}
	})
	if err != nil && ctx.Err() != nil {
		ui.log("%s", T("JobInterruptedResume"))
	} else if err != nil {
		ui.log("%s", T("JobCopyError", err))
	}

	failPath := ""
	if len(fails) > 0 {
		failPath = filepath.Join(dest, failListFileName)
		if werr := writeFailCSV(failPath, fails); werr != nil {
			ui.log("%s", T("JobFailList", werr))
			failPath = ""
		}
	}
	if werr := writeReport(dest, computerName(), employee, roots, opt, sum, res, failPath, started); werr != nil {
		ui.log("%s", T("JobReportFail", werr))
	}

	if res != nil {
		ui.log("%s", T("JobDone",
			res.CopiedFiles, formatBytes(res.CopiedBytes), res.SkippedSame, res.Failed))
		if opt.Cut {
			ui.log("%s", T("JobDeleted", res.CutFiles, res.CutFailed))
		}
	}
	ui.log("%s", T("JobDest", dest))
	ui.log("%s", T("JobReport", filepath.Join(dest, reportFileName)))
	if failPath != "" {
		ui.log("%s", T("JobFailPath", failPath))
	}
	return nil
}
