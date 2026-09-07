package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
	"unicode/utf8"
)

type CopyResult struct {
	CopiedFiles int64
	CopiedBytes int64
	SkippedSame int64
	Failed      int64
	Dirs        int64
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

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	if err := out.Close(); err != nil {
		return err
	}
	_ = os.Chtimes(longPath(tmp), srcInfo.ModTime(), srcInfo.ModTime())
	_ = os.Remove(longPath(dst))
	if err := os.Rename(longPath(tmp), longPath(dst)); err != nil {
		if copyErr := replaceByCopy(tmp, dst); copyErr != nil {
			return err
		}
	}
	ok = true
	return nil
}

func replaceByCopy(tmp, dst string) error {
	in, err := os.Open(longPath(tmp))
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(longPath(dst), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(out, in)
	closeErr := out.Close()
	_ = os.Remove(longPath(tmp))
	if copyErr != nil {
		return copyErr
	}
	return closeErr
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
		if sameFile(info, dst) {
			res.SkippedSame++
			if onProgress != nil {
				onProgress(res.CopiedFiles+res.SkippedSame, res.CopiedBytes+info.Size(), src)
			}
			return nil
		}
		if err := copyFileWithRetry(src, dst, info); err != nil {
			res.Failed++
			if onErr != nil {
				onErr(src, dst, err)
			}
			return nil
		}
		res.CopiedFiles++
		res.CopiedBytes += info.Size()
		if onProgress != nil {
			onProgress(res.CopiedFiles+res.SkippedSame, res.CopiedBytes, src)
		}
		return nil
	})
	return res, err
}

func writeReport(dest, computer, employee string, roots []string, opt Options, sum *Summary, res *CopyResult, failPath string, started time.Time) error {
	if err := os.MkdirAll(longPath(dest), 0o755); err != nil {
		return err
	}
	p := filepath.Join(dest, "_归档报告.txt")
	f, err := os.Create(longPath(p))
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintf(f, "离职资料归档报告\r\n")
	fmt.Fprintf(f, "================\r\n")
	fmt.Fprintf(f, "计算机名: %s\r\n", computer)
	fmt.Fprintf(f, "员工标识: %s\r\n", employee)
	fmt.Fprintf(f, "开始时间: %s\r\n", started.Format("2006-01-02 15:04:05"))
	fmt.Fprintf(f, "结束时间: %s\r\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintf(f, "归档模式: %s\r\n", opt.Mode.Title())
	fmt.Fprintf(f, "包含已安装程序: %v\r\n", opt.IncludeProgramFiles)
	fmt.Fprintf(f, "跳过可再生成目录: %v\r\n", opt.SkipRegeneratable)
	fmt.Fprintf(f, "源路径: %s\r\n", joinComma(roots))
	fmt.Fprintf(f, "目标目录: %s\r\n", dest)
	fmt.Fprintf(f, "\r\n路径对照示例:\r\n")
	fmt.Fprintf(f, "  C:\\Users\\张三\\Downloads\\a.pdf  ->  %s\\C\\Users\\张三\\Downloads\\a.pdf\r\n", dest)
	fmt.Fprintf(f, "  （Windows 路径不能使用 D:\\C:\\... 这种带冒号的目录名，因此用 D:\\...\\C\\... 表示 C 盘）\r\n")
	fmt.Fprintf(f, "\r\n扫描合计: %d 个文件, %s\r\n", sum.Files, formatBytes(sum.Bytes))
	if res != nil {
		fmt.Fprintf(f, "新复制: %d 个文件, %s\r\n", res.CopiedFiles, formatBytes(res.CopiedBytes))
		fmt.Fprintf(f, "已存在且相同（续传跳过）: %d\r\n", res.SkippedSame)
		fmt.Fprintf(f, "失败: %d\r\n", res.Failed)
	}
	fmt.Fprintf(f, "\r\n按顶层目录:\r\n")
	for _, folder := range sum.sortedFolders() {
		fmt.Fprintf(f, "  %-24s %8d 文件  %s\r\n", folder.Key, folder.Files, formatBytes(folder.Bytes))
	}
	if failPath != "" {
		fmt.Fprintf(f, "\r\n失败清单: %s\r\n", failPath)
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
	_ = w.Write([]string{"源路径", "目标路径", "错误"})
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
