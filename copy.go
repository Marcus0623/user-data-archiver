package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unicode/utf8"
)

var errOnlinePlaceholder = errors.New("not deleting online-only/cloud placeholder")

type CopyResult struct {
	CopiedFiles int64
	CopiedBytes int64
	SkippedSame int64
	Failed      int64
	Dirs        int64
	CutFiles    int64
	CutFailed   int64
}

func sameFile(srcInfo os.FileInfo, dst string) bool {
	st, err := os.Stat(longPath(dst))
	if err != nil {
		return false
	}
	if st.Size() != srcInfo.Size() {
		return false
	}
	return st.ModTime().Equal(srcInfo.ModTime())
}

func copyFileWithRetry(src, dst string, srcInfo os.FileInfo) error {
	if err := os.MkdirAll(longPath(filepath.Dir(dst)), 0o755); err != nil {
		return err
	}
	var last error
	for attempt := 0; attempt < 2; attempt++ {
		last = copyFileOnce(src, dst, srcInfo)
		if last == nil {
			return nil
		}
		if attempt == 0 {
			time.Sleep(400 * time.Millisecond)
		}
	}
	return last
}

func copyFileOnce(src, dst string, srcInfo os.FileInfo) error {
	in, err := os.Open(longPath(src))
	if err != nil {
		return err
	}
	defer in.Close()

	tmp := dst + ".archiving"
	out, err := os.OpenFile(longPath(tmp), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	ok := false
	defer func() {
		_ = out.Close()
		if !ok {
			_ = os.Remove(longPath(tmp))
		}
	}()

	live, err := in.Stat()
	if err != nil {
		return err
	}
	want := live.Size()
	n, err := io.Copy(out, in)
	if err != nil {
		return err
	}
	if n != want {
		return fmt.Errorf("copied %d bytes, source size is %d", n, want)
	}
	if err := out.Sync(); err != nil {
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	st, err := os.Stat(longPath(tmp))
	if err != nil {
		return err
	}
	if st.Size() != want {
		return fmt.Errorf("temp file size %d does not match source %d", st.Size(), want)
	}
	_ = os.Chtimes(longPath(tmp), srcInfo.ModTime(), srcInfo.ModTime())
	if err := replaceFile(tmp, dst); err != nil {
		return err
	}
	ok = true
	if err := verifyDestComplete(dst, want); err != nil {
		return err
	}
	_ = os.Chtimes(longPath(dst), srcInfo.ModTime(), srcInfo.ModTime())
	return nil
}

func replaceFile(tmp, dst string) error {
	clearReadOnly(dst)
	if err := moveFileReplace(tmp, dst); err == nil {
		return nil
	}
	return replaceByCopy(tmp, dst)
}

func replaceByCopy(tmp, dst string) error {
	in, err := os.Open(longPath(tmp))
	if err != nil {
		return err
	}
	defer in.Close()
	srcInfo, err := in.Stat()
	if err != nil {
		return err
	}
	sidecar := dst + ".replacing"
	out, err := os.OpenFile(longPath(sidecar), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	n, copyErr := io.Copy(out, in)
	syncErr := out.Sync()
	closeErr := out.Close()
	if copyErr != nil || syncErr != nil || closeErr != nil || n != srcInfo.Size() {
		_ = os.Remove(longPath(sidecar))
		if copyErr != nil {
			return copyErr
		}
		if n != srcInfo.Size() {
			return fmt.Errorf("copied %d bytes, temp size is %d", n, srcInfo.Size())
		}
		if syncErr != nil {
			return syncErr
		}
		return closeErr
	}
	if err := moveFileReplace(sidecar, dst); err != nil {
		_ = os.Remove(longPath(sidecar))
		return err
	}
	_ = os.Remove(longPath(tmp))
	return nil
}

func verifyDestComplete(dst string, wantSize int64) error {
	st, err := os.Stat(longPath(dst))
	if err != nil {
		return fmt.Errorf("destination missing after save: %w", err)
	}
	if st.IsDir() {
		return fmt.Errorf("destination is a directory, not a file: %s", dst)
	}
	if st.Size() != wantSize {
		return fmt.Errorf("destination size %d does not match source %d", st.Size(), wantSize)
	}
	f, err := os.Open(longPath(dst))
	if err != nil {
		return fmt.Errorf("destination not readable after save: %w", err)
	}
	_ = f.Close()
	return nil
}

func contentsEqual(src, dst string) (bool, error) {
	fa, err := os.Open(longPath(src))
	if err != nil {
		return false, err
	}
	defer fa.Close()
	fb, err := os.Open(longPath(dst))
	if err != nil {
		return false, err
	}
	defer fb.Close()
	sa, err := fa.Stat()
	if err != nil {
		return false, err
	}
	sb, err := fb.Stat()
	if err != nil {
		return false, err
	}
	if sa.Size() != sb.Size() {
		return false, nil
	}
	bufa := make([]byte, 256*1024)
	bufb := make([]byte, 256*1024)
	for {
		na, ea := fa.Read(bufa)
		nb, eb := fb.Read(bufb)
		if !bytes.Equal(bufa[:na], bufb[:nb]) {
			return false, nil
		}
		aDone := ea == io.EOF
		bDone := eb == io.EOF
		if ea != nil && !aDone {
			return false, ea
		}
		if eb != nil && !bDone {
			return false, eb
		}
		if aDone && bDone {
			return true, nil
		}
		if aDone != bDone {
			return false, nil
		}
		if na == 0 && nb == 0 {
			return false, errors.New("empty read without EOF")
		}
	}
}

func destReadyForCut(src, dst string, _ os.FileInfo) error {
	if samePath(src, dst) {
		return fmt.Errorf("source and destination are the same path")
	}
	st, err := os.Stat(longPath(src))
	if err != nil {
		return err
	}
	if err := verifyDestComplete(dst, st.Size()); err != nil {
		return err
	}
	ok, err := contentsEqual(src, dst)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("destination content does not match source")
	}
	return nil
}

func clearReadOnly(path string) {
	st, err := os.Stat(longPath(path))
	if err != nil {
		return
	}
	if st.Mode()&0200 == 0 {
		_ = os.Chmod(longPath(path), st.Mode()|0200)
	}
}

func isOnlinePlaceholder(info os.FileInfo) bool {
	if info == nil {
		return false
	}
	st, ok := info.Sys().(*syscall.Win32FileAttributeData)
	if !ok {
		return false
	}
	const (
		fileAttributeRecallOnDataAccess = 0x00400000
		fileAttributeRecallOnOpen       = 0x00040000
	)
	return st.FileAttributes&(fileAttributeRecallOnDataAccess|fileAttributeRecallOnOpen) != 0
}

func saveThenMaybeCut(src, dst string, info os.FileInfo, cut bool, already bool) (copied, destOK, cutOK bool, err error) {
	if samePath(src, dst) {
		return false, false, false, fmt.Errorf("refusing to archive a file onto itself")
	}
	needCopy := !already
	if cut && already {
		if destReadyForCut(src, dst, info) != nil {
			needCopy = true
		}
	}
	if needCopy {
		if err := copyFileWithRetry(src, dst, info); err != nil {
			return false, false, false, err
		}
		copied = true
	}
	if !cut {
		return copied, true, false, nil
	}
	if isOnlinePlaceholder(info) {
		ready := destReadyForCut(src, dst, info) == nil
		return copied, ready, false, errOnlinePlaceholder
	}
	if err := destReadyForCut(src, dst, info); err != nil {
		return copied, false, false, err
	}
	clearReadOnly(src)
	if err := os.Remove(longPath(src)); err != nil {
		return copied, true, false, fmt.Errorf("saved but could not delete source: %w", err)
	}
	removeEmptyParents(src)
	return copied, true, true, nil
}

func Archive(ctx context.Context, roots []string, opt Options, dest string, onProgress func(doneFiles, doneBytes int64, src string), onErr func(src, dst string, err error)) (*CopyResult, error) {
	res := &CopyResult{}
	err := walkSources(ctx, roots, opt, func(src string, e error) {
		res.Failed++
		if onErr != nil {
			onErr(src, "", e)
		}
	}, func(src string, info os.FileInfo) error {
		dst, err := MapDest(src, dest)
		if err != nil {
			res.Failed++
			if onErr != nil {
				onErr(src, "", err)
			}
			return nil
		}
		if info.IsDir() {
			if err := os.MkdirAll(longPath(dst), 0o755); err != nil {
				res.Failed++
				if onErr != nil {
					onErr(src, dst, err)
				}
				return nil
			}
			res.Dirs++
			return nil
		}
		already := sameFile(info, dst)
		copied, destOK, cutOK, err := saveThenMaybeCut(src, dst, info, opt.Cut, already)
		if err != nil {
			if destOK {
				if copied {
					res.CopiedFiles++
					res.CopiedBytes += info.Size()
				} else {
					res.SkippedSame++
				}
				if opt.Cut {
					res.CutFailed++
				}
			} else {
				res.Failed++
			}
			if onErr != nil {
				onErr(src, dst, err)
			}
			return nil
		}
		if copied {
			res.CopiedFiles++
			res.CopiedBytes += info.Size()
		} else {
			res.SkippedSame++
		}
		if opt.Cut && cutOK {
			res.CutFiles++
		}
		if onProgress != nil {
			onProgress(res.CopiedFiles+res.SkippedSame, res.CopiedBytes, src)
		}
		return nil
	})
	return res, err
}

func removeEmptyParents(filePath string) {
	dir := filepath.Dir(normalizeAbs(filePath))
	root := filepath.VolumeName(dir) + `\`
	for {
		cleanRoot := strings.TrimSuffix(root, `\`)
		if strings.EqualFold(dir, root) || strings.EqualFold(dir, cleanRoot) {
			return
		}
		if !mayRemoveEmptiedDir(dir) {
			return
		}
		if err := os.Remove(longPath(dir)); err != nil {
			return
		}
		parent := filepath.Dir(dir)
		if strings.EqualFold(parent, dir) {
			return
		}
		dir = parent
	}
}

func mayRemoveEmptiedDir(path string) bool {
	_, rel := volumeAndRel(path)
	parts := pathComponents(rel)
	if len(parts) == 0 {
		return false
	}
	logical := stripWindowsOldPrefix(parts)
	if len(logical) == 0 {
		return false
	}
	if strings.EqualFold(logical[0], "Users") && len(logical) <= 2 {
		return false
	}
	return true
}

func writeReport(dest, computer, employee string, roots []string, opt Options, sum *Summary, res *CopyResult, failPath string, started time.Time) error {
	if err := os.MkdirAll(longPath(dest), 0o755); err != nil {
		return err
	}
	p := filepath.Join(dest, reportFileName)
	f, err := os.Create(longPath(p))
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintf(f, "User Data Archive Report\r\n")
	fmt.Fprintf(f, "========================\r\n")
	fmt.Fprintf(f, "Computer: %s\r\n", computer)
	fmt.Fprintf(f, "Person: %s\r\n", employee)
	fmt.Fprintf(f, "Started: %s\r\n", started.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(f, "Finished: %s\r\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintf(f, "Mode: %s\r\n", opt.Mode.Title())
	fmt.Fprintf(f, "Transfer: %s\r\n", transferTitle(opt.Cut))
	fmt.Fprintf(f, "Include installed programs: %v\r\n", opt.IncludeProgramFiles)
	fmt.Fprintf(f, "Skip regeneratable folders: %v\r\n", opt.SkipRegeneratable)
	fmt.Fprintf(f, "Source: %s\r\n", joinComma(roots))
	fmt.Fprintf(f, "Destination: %s\r\n", dest)
	fmt.Fprintf(f, "\r\nPath mapping example:\r\n")
	fmt.Fprintf(f, "  C:\\Users\\alice\\Downloads\\a.pdf  ->  %s\\C\\Users\\alice\\Downloads\\a.pdf\r\n", dest)
	fmt.Fprintf(f, "  (Windows cannot use D:\\C:\\... as a folder name, so D:\\...\\C\\... stands for drive C:)\r\n")
	fmt.Fprintf(f, "\r\nScan total: %d files, %s\r\n", sum.Files, formatBytes(sum.Bytes))
	if res != nil {
		fmt.Fprintf(f, "Newly copied: %d files, %s\r\n", res.CopiedFiles, formatBytes(res.CopiedBytes))
		fmt.Fprintf(f, "Already present (resume skip): %d\r\n", res.SkippedSame)
		if opt.Cut {
			fmt.Fprintf(f, "Originals deleted: %d\r\n", res.CutFiles)
			fmt.Fprintf(f, "Delete original failed: %d\r\n", res.CutFailed)
		}
		fmt.Fprintf(f, "Failed: %d\r\n", res.Failed)
	}
	fmt.Fprintf(f, "\r\nBy top-level folder:\r\n")
	for _, folder := range sum.sortedFolders() {
		fmt.Fprintf(f, "  %-24s %8d files  %s\r\n", folder.Key, folder.Files, formatBytes(folder.Bytes))
	}
	if failPath != "" {
		fmt.Fprintf(f, "\r\nFailure list: %s\r\n", failPath)
	}
	return nil
}

func writeFailCSV(path string, rows [][]string) error {
	if len(rows) == 0 {
		return nil
	}
	if err := os.MkdirAll(longPath(filepath.Dir(path)), 0o755); err != nil {
		return err
	}
	f, err := os.Create(longPath(path))
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.Write([]byte{0xEF, 0xBB, 0xBF}); err != nil {
		return err
	}
	w := csv.NewWriter(f)
	_ = w.Write([]string{"source", "destination", "error"})
	for _, row := range rows {
		_ = w.Write(row)
	}
	w.Flush()
	return w.Error()
}

func joinComma(ss []string) string {
	out := ""
	for i, s := range ss {
		if i > 0 {
			out += ", "
		}
		out += s
	}
	return out
}

func formatBytes(n int64) string {
	if n < 0 {
		n = 0
	}
	units := []string{"B", "KB", "MB", "GB", "TB"}
	val := float64(n)
	i := 0
	for val >= 1024 && i < len(units)-1 {
		val /= 1024
		i++
	}
	if i == 0 {
		return fmt.Sprintf("%d B", n)
	}
	return fmt.Sprintf("%.2f %s", val, units[i])
}

func shortenPath(p string, max int) string {
	if max < 8 || utf8.RuneCountInString(p) <= max {
		return p
	}
	runes := []rune(p)
	keep := (max - 3) / 2
	return string(runes[:keep]) + "..." + string(runes[len(runes)-keep:])
}

func transferTitle(cut bool) string {
	if cut {
		return TransferCut.Title()
	}
	return TransferCopy.Title()
}
