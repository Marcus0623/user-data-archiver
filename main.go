package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

func main() {
	setupConsole()

	dstFlag := flag.String("dst", "", "归档目标目录（最终目录，里面会生成 C、D 等盘符文件夹）")
	nameFlag := flag.String("name", "", "员工姓名；交互模式下会拼到目标父目录后面")
	srcFlag := flag.String("src", "", "源路径，逗号分隔。例如 C: 或 C:,D: 或 C:\\Users")
	modeFlag := flag.String("mode", "", "归档模式: personal | appdata | all。默认 appdata")
	includePF := flag.Bool("include-program-files", false, "包含 Program Files / ProgramData（体积很大，离职交接通常不需要）")
	includeRegen := flag.Bool("include-regeneratable", false, "包含 node_modules、__pycache__ 等可再生成目录")
	excludeFlag := flag.String("exclude", "", "额外排除的目录名（任意层级匹配），逗号分隔")
	dryRun := flag.Bool("dry-run", false, "只扫描预览，不复制")
	yes := flag.Bool("yes", false, "扫描后不询问，直接开始复制")
	flag.Parse()

	in := bufio.NewReader(os.Stdin)

	fmt.Println("========================================")
	fmt.Println("  离职资料归档工具")
	fmt.Println("========================================")
	fmt.Printf("计算机: %s\n", computerName())
	if isAdmin() {
		fmt.Println("权限: 已是管理员（可读取其他用户目录）")
	} else {
		fmt.Println("权限: 当前不是管理员。若要归档本机全部用户资料，请右键“以管理员身份运行”。")
	}
	fmt.Println()

	employee := strings.TrimSpace(*nameFlag)
	if employee == "" {
		if *yes {
			employee = defaultName()
		} else {
			employee = prompt(in, "员工姓名（用于文件夹命名）", defaultName())
		}
	}
	employee = sanitizeEmployeeName(employee)

	dest := strings.TrimSpace(*dstFlag)
	if dest == "" {
		if *yes {
			fatal("使用 -yes 时必须指定 -dst")
		}
		parent := prompt(in, "目标父目录（例如 D:\\离职归档）", `D:\离职归档`)
		dest = filepath.Join(strings.TrimSpace(parent), employee)
	}
	dest = normalizeAbs(dest)
	if vol := filepath.VolumeName(dest); vol != "" {
		if _, err := os.Stat(vol + `\`); err != nil {
			fmt.Printf("警告: 目标盘 %s 当前无法访问，稍后创建目录可能会失败。\n", vol)
		}
	}

	srcText := strings.TrimSpace(*srcFlag)
	if srcText == "" {
		if *yes {
			fatal("使用 -yes 时必须指定 -src")
		}
		srcText = prompt(in, "源路径（多个用逗号，盘符如 C: ）", "C:")
	}
	roots := splitRoots(srcText)
	if len(roots) == 0 {
		fatal("未指定源路径")
	}

	mode := ModeWithAppData
	if strings.TrimSpace(*modeFlag) != "" {
		mode = ParseMode(*modeFlag)
	} else if !*yes {
		fmt.Println()
		fmt.Println("归档模式:")
		fmt.Println("  1) " + ModePersonal.Title())
		fmt.Println("  2) " + ModeWithAppData.Title())
		fmt.Println("  3) " + ModeAllNonSystem.Title())
		mode = ParseMode(prompt(in, "请选择", "2"))
	}

	opt := Options{
		Mode:                mode,
		IncludeProgramFiles: *includePF,
		SkipRegeneratable:   !*includeRegen,
		DestAbs:             dest,
		EmployeeName:        employee,
		ExtraExclude:        splitRoots(*excludeFlag),
	}

	fmt.Println()
	fmt.Println("将按原目录结构归档，盘符冒号会改成文件夹名（Windows 不允许 D:\\C:\\...）:")
	fmt.Printf("  C:\\Users\\%s\\Downloads\\文件.pdf\n", employee)
	fmt.Printf("    -> %s\n", filepath.Join(dest, "C", "Users", employee, "Downloads", "文件.pdf"))
	fmt.Printf("模式: %s\n", opt.Mode.Title())
	fmt.Printf("源: %s\n", joinComma(roots))
	fmt.Printf("目标: %s\n", dest)
	fmt.Println("说明: 将跳过 Windows、回收站、页面文件等重装系统后仍在的系统目录；默认不拷贝 Program Files。")
	fmt.Println("建议: 若使用 OneDrive“仅联机”文件，请先设为“始终保留在此设备上”。")
	fmt.Println()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	fmt.Println("正在扫描（可能需要几分钟）...")
	started := time.Now()
	var scanErrs int
	sum, err := Scan(ctx, roots, opt, func(src string, e error) {
		scanErrs++
		if scanErrs <= 20 {
			fmt.Printf("  [扫描跳过] %s: %v\n", src, e)
		}
	})
	if err != nil {
		if ctx.Err() != nil {
			fatal("已中断")
		}
		fatal("扫描失败: " + err.Error())
	}

	fmt.Println()
	fmt.Println("预览（按顶层目录）:")
	for _, folder := range sum.sortedFolders() {
		if folder.Files == 0 && folder.Bytes == 0 {
			continue
		}
		fmt.Printf("  %-28s %8d 文件  %10s\n", folder.Key, folder.Files, formatBytes(folder.Bytes))
	}
	fmt.Printf("\n合计: %d 个文件, %s", sum.Files, formatBytes(sum.Bytes))
	if sum.Errors > 0 {
		fmt.Printf("（扫描中 %d 个路径无法读取）", sum.Errors)
	}
	fmt.Println()

	if vol := filepath.VolumeName(dest); vol != "" {
		if free, _, err := diskFreeBytes(vol + `\`); err == nil {
			fmt.Printf("目标盘剩余空间: %s\n", formatBytes(int64(free)))
			need := uint64(sum.Bytes) + uint64(sum.Bytes/20) + 64*1024*1024
			if free < need {
				fmt.Println("警告: 剩余空间可能不足。")
				if !*yes && !askYes(in, "空间可能不够，仍要继续", false) {
					fmt.Println("已取消。")
					return
				}
			}
		}
	}

	if *dryRun {
		if err := os.MkdirAll(longPath(dest), 0o755); err == nil {
			_ = writeReport(dest, computerName(), employee, roots, opt, sum, nil, "", started)
			fmt.Printf("已写入预览报告: %s\n", filepath.Join(dest, "_归档报告.txt"))
		}
		fmt.Println("dry-run 完成，未复制文件。")
		return
	}

	if !*yes {
		if !askYes(in, "开始复制到目标目录", true) {
			fmt.Println("已取消。")
			return
		}
	}

	if err := os.MkdirAll(longPath(dest), 0o755); err != nil {
		fatal("无法创建目标目录: " + err.Error())
	}

	fmt.Println("正在复制（可随时 Ctrl+C 中断，再次运行会跳过已相同的文件）...")
	lastPrint := time.Now()
	fails := make([][]string, 0, 64)
	res, err := Archive(ctx, roots, opt, dest, func(doneFiles, doneBytes int64, src string) {
		if time.Since(lastPrint) < 200*time.Millisecond {
			return
		}
		lastPrint = time.Now()
		pct := ""
		if sum.Bytes > 0 {
			pct = fmt.Sprintf(" %.1f%%", float64(doneBytes)*100/float64(sum.Bytes))
		}
		fmt.Printf("\r  %d/%d 文件%s  %s  %s          ",
			doneFiles, sum.Files, pct, formatBytes(doneBytes), shortenPath(src, 60))
	}, func(src, dst string, e error) {
		if len(fails) < 5000 {
			fails = append(fails, []string{src, dst, e.Error()})
		}
	})
	fmt.Println()
	if err != nil && ctx.Err() != nil {
		fmt.Println("已中断。可使用同一目标目录再次运行以续传。")
	} else if err != nil {
		fmt.Println("复制过程出错:", err)
	}

	failPath := ""
	if len(fails) > 0 {
		failPath = filepath.Join(dest, "_失败清单.csv")
		if werr := writeFailCSV(failPath, fails); werr != nil {
			fmt.Println("写入失败清单出错:", werr)
			failPath = ""
		}
	}
	if werr := writeReport(dest, computerName(), employee, roots, opt, sum, res, failPath, started); werr != nil {
		fmt.Println("写入报告出错:", werr)
	}

	if res != nil {
		fmt.Printf("完成: 新复制 %d 个文件 (%s)，续传跳过 %d，失败 %d\n",
			res.CopiedFiles, formatBytes(res.CopiedBytes), res.SkippedSame, res.Failed)
	}
	fmt.Printf("目标目录: %s\n", dest)
	fmt.Printf("报告: %s\n", filepath.Join(dest, "_归档报告.txt"))
	if failPath != "" {
		fmt.Printf("失败清单: %s\n", failPath)
	}
}

func defaultName() string {
	u := os.Getenv("USERNAME")
	if u == "" {
		u = "archive"
	}
	return u
}

func splitRoots(s string) []string {
	parts := strings.FieldsFunc(s, func(r rune) bool {
		return r == ',' || r == ';'
	})
	out := make([]string, 0, len(parts))
	seen := map[string]bool{}
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		key := strings.ToLower(p)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, p)
	}
	return out
}

func prompt(in *bufio.Reader, label, fallback string) string {
	if fallback != "" {
		fmt.Printf("%s [%s]: ", label, fallback)
	} else {
		fmt.Printf("%s: ", label)
	}
	text, err := in.ReadString('\n')
	if err != nil {
		return fallback
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return fallback
	}
	return text
}

func askYes(in *bufio.Reader, label string, defYes bool) bool {
	hint := "Y/n"
	if !defYes {
		hint = "y/N"
	}
	fmt.Printf("%s [%s]: ", label, hint)
	text, err := in.ReadString('\n')
	if err != nil {
		return defYes
	}
	text = strings.TrimSpace(strings.ToLower(text))
	if text == "" {
		return defYes
	}
	return text == "y" || text == "yes" || text == "是"
}

func fatal(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}
