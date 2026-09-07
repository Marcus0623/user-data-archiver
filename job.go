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

	ui.log("Files keep their original folders. Drive colons become folder names (Windows does not allow D:\\C:\\...):")
	ui.log("  C:\\Users\\%s\\Downloads\\file.pdf", employee)
	ui.log("    -> %s", filepath.Join(dest, "C", "Users", employee, "Downloads", "file.pdf"))
	ui.log("Mode: %s", opt.Mode.Title())
	ui.log("Transfer: %s", transferTitle(opt.Cut))
	ui.log("Source: %s", joinComma(roots))
	ui.log("Destination: %s", dest)
	for _, r := range roots {
		abs, err := expandRoot(r)
		if err != nil {
			continue
		}
		if isBaseOf(abs, dest) {
			ui.log("Warning: destination is inside the source path %s. Files already in the destination folder are skipped.", abs)
		}
	}
	ui.log("Note: Windows, Recycle Bin, pagefile, and other OS files are skipped. Program Files is skipped by default.")
	ui.log("Tip: set OneDrive files to Always keep on this device before archiving.")
	ui.log("Scanning (this may take a few minutes)...")

	started := time.Now()
	var scanErrs int
	sum, err := Scan(ctx, roots, opt, func(src string, e error) {
		scanErrs++
		if scanErrs <= 20 {
			ui.log("[scan skip] %s: %v", src, e)
		}
	})
	if err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("interrupted")
		}
		return fmt.Errorf("scan failed: %w", err)
	}

	ui.log("Preview (by top-level folder):")
	for _, folder := range sum.sortedFolders() {
		if folder.Files == 0 && folder.Bytes == 0 {
			continue
		}
		ui.log("  %-28s %8d files  %10s", folder.Key, folder.Files, formatBytes(folder.Bytes))
	}
	totalLine := fmt.Sprintf("Total: %d files, %s", sum.Files, formatBytes(sum.Bytes))
	if sum.Errors > 0 {
		totalLine += fmt.Sprintf(" (%d paths unreadable during scan)", sum.Errors)
	}
	ui.log("%s", totalLine)

	if vol := filepath.VolumeName(dest); vol != "" {
		if free, _, err := diskFreeBytes(vol + `\`); err == nil {
			ui.log("Free space on destination drive: %s", formatBytes(int64(free)))
			need := uint64(sum.Bytes) + uint64(sum.Bytes/20) + 64*1024*1024
			if free < need {
				ui.log("Warning: free space may not be enough.")
				if !skipConfirm && !ui.ask("Free space may not be enough. Continue anyway?", false) {
					ui.log("Cancelled.")
					return nil
				}
			}
		}
	}

	if dryRun {
		if err := os.MkdirAll(longPath(dest), 0o755); err == nil {
			_ = writeReport(dest, computerName(), employee, roots, opt, sum, nil, "", started)
			ui.log("Preview report written: %s", filepath.Join(dest, reportFileName))
		}
		ui.log("dry-run finished; no files copied or moved.")
		return nil
	}

	confirm := "Copy files to the destination?"
	defYes := true
	if opt.Cut {
		confirm = "MOVE files to the destination and DELETE each original only after that file is flushed and verified? This cannot be undone."
		defYes = false
	}
	if !skipConfirm && !ui.ask(confirm, defYes) {
		ui.log("Cancelled.")
		return nil
	}

	if err := os.MkdirAll(longPath(dest), 0o755); err != nil {
		return fmt.Errorf("cannot create destination folder: %w", err)
	}

	ui.log("Copying (you can stop and run again on the same destination to resume)...")
	if opt.Cut {
		ui.log("Cut mode: a source file is deleted only after the destination is flushed and the contents match.")
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
		ui.log("Interrupted. Run again with the same destination to resume.")
	} else if err != nil {
		ui.log("Copy error: %v", err)
	}

	failPath := ""
	if len(fails) > 0 {
		failPath = filepath.Join(dest, failListFileName)
		if werr := writeFailCSV(failPath, fails); werr != nil {
			ui.log("failed to write failure list: %v", werr)
			failPath = ""
		}
	}
	if werr := writeReport(dest, computerName(), employee, roots, opt, sum, res, failPath, started); werr != nil {
		ui.log("failed to write report: %v", werr)
	}

	if res != nil {
		ui.log("Done: copied %d files (%s), resumed/skipped %d, failed %d",
			res.CopiedFiles, formatBytes(res.CopiedBytes), res.SkippedSame, res.Failed)
		if opt.Cut {
			ui.log("Originals deleted: %d; delete failed: %d", res.CutFiles, res.CutFailed)
		}
	}
	ui.log("Destination: %s", dest)
	ui.log("Report: %s", filepath.Join(dest, reportFileName))
	if failPath != "" {
		ui.log("Failure list: %s", failPath)
	}
	return nil
}
