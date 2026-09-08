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
)

func main() {
	dstFlag := flag.String("dst", "", "archive destination folder (drive letters become C, D, ... subfolders)")
	nameFlag := flag.String("name", "", "person name; in interactive mode this is appended to the parent folder")
	srcFlag := flag.String("src", "", "source paths, comma-separated. Example: C: or C:,D: or C:\\Users")
	modeFlag := flag.String("mode", "", "archive mode: personal | appdata | all. Default appdata")
	includePF := flag.Bool("include-program-files", false, "include Program Files / ProgramData (large; usually not needed)")
	includeRegen := flag.Bool("include-regeneratable", false, "include node_modules, __pycache__, and similar folders")
	excludeFlag := flag.String("exclude", "", "extra directory names to skip at any depth, comma-separated")
	dryRun := flag.Bool("dry-run", false, "scan and preview only, do not copy")
	yes := flag.Bool("yes", false, "do not ask for confirmation after the scan")
	cliFlag := flag.Bool("cli", false, "use the command-line interface instead of the window")
	cutFlag := flag.Bool("cut", false, "move files: delete originals only after the destination is flushed and contents match")
	langFlag := flag.String("lang", "", "UI language: en, zh-CN, zh-TW, ja, fr, ru, vi (default: Windows display language)")
	flag.Parse()
	if strings.TrimSpace(*langFlag) != "" {
		setLang(ParseLang(*langFlag))
	}

	if *cliFlag || *yes {
		attachParentConsole()
		setupConsole()
		runCLI(*dstFlag, *nameFlag, *srcFlag, *modeFlag, *includePF, *includeRegen, *excludeFlag, *dryRun, *yes, *cutFlag)
		return
	}

	runGUI(guiPrefill{
		Name:                strings.TrimSpace(*nameFlag),
		Dest:                strings.TrimSpace(*dstFlag),
		Src:                 strings.TrimSpace(*srcFlag),
		Mode:                strings.TrimSpace(*modeFlag),
		IncludeProgramFiles: *includePF,
		IncludeRegen:        *includeRegen,
		Exclude:             strings.TrimSpace(*excludeFlag),
		DryRun:              *dryRun,
		Cut:                 *cutFlag,
		ActionChosen:        *cutFlag,
	})
	os.Exit(0)
}

func runCLI(dstFlag, nameFlag, srcFlag, modeFlag string, includePF, includeRegen bool, excludeFlag string, dryRun, yes, cut bool) {
	in := bufio.NewReader(os.Stdin)

	fmt.Println(T("CLIHeader"))
	fmt.Println("  " + T("AppName"))
	fmt.Println(T("CLIHeader"))
	fmt.Printf("%s\n", T("CLIComputer", computerName()))
	if isAdmin() {
		fmt.Println(T("CLIRightsAdmin"))
	} else {
		fmt.Println(T("CLIRightsUser"))
	}
	fmt.Println()

	employee := strings.TrimSpace(nameFlag)
	if employee == "" {
		if yes {
			employee = defaultName()
		} else {
			employee = prompt(in, T("CLIPersonPrompt"), defaultName())
		}
	}
	employee = sanitizeEmployeeName(employee)

	dest := strings.TrimSpace(dstFlag)
	if dest == "" {
		if yes {
			fatal(T("CLIYesNeedDst"))
		}
		parent := prompt(in, T("CLIParentPrompt"), defaultDestParent)
		dest = filepath.Join(strings.TrimSpace(parent), employee)
	}
	dest = normalizeAbs(dest)
	fmt.Printf("%s\n", T("CLIDestDrive", destDriveInfo(dest)))
	fmt.Printf("%s\n", T("CLIDestFolder", dest))
	fmt.Print(T("CLIChecking"))
	if err := probeDestWritable(dest); err != nil {
		fmt.Println(T("CLIFailed"))
		fatal(T("CLIYesNeedPerm", err.Error()))
	}
	fmt.Println(T("CLIOK"))

	srcText := strings.TrimSpace(srcFlag)
	if srcText == "" {
		if yes {
			fatal(T("CLIYesNeedSrc"))
		}
		srcText = prompt(in, T("CLISourcePrompt"), "C:")
	}
	roots := splitRoots(srcText)
	if len(roots) == 0 {
		fatal(T("CLINoSource"))
	}

	mode := ModeWithAppData
	if strings.TrimSpace(modeFlag) != "" {
		mode = ParseMode(modeFlag)
	} else if !yes {
		fmt.Println()
		fmt.Println(T("CLIModeTitle"))
		fmt.Println("  1) " + ModePersonal.Title())
		fmt.Println("  2) " + ModeWithAppData.Title())
		fmt.Println("  3) " + ModeAllNonSystem.Title())
		mode = ParseMode(prompt(in, T("CLIChoose"), "2"))
	}

	opt := Options{
		Mode:                mode,
		IncludeProgramFiles: includePF,
		SkipRegeneratable:   !includeRegen,
		DestAbs:             dest,
		EmployeeName:        employee,
		ExtraExclude:        splitRoots(excludeFlag),
		Cut:                 cut,
	}

	fmt.Println()
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	err := runArchiveJob(ctx, roots, dest, employee, opt, dryRun, yes, jobUI{
		log: func(format string, args ...any) {
			fmt.Printf(format+"\n", args...)
		},
		progress: func(doneFiles, totalFiles, doneBytes, totalBytes int64, src string) {
			pct := ""
			if totalBytes > 0 {
				pct = fmt.Sprintf(" %.1f%%", float64(doneBytes)*100/float64(totalBytes))
			}
			fmt.Printf("\r%s          ", T("LogProgress", doneFiles, totalFiles, pct, formatBytes(doneBytes), shortenPath(src, 60)))
		},
		ask: func(question string, defYes bool) bool {
			return askYes(in, question, defYes)
		},
	})
	fmt.Println()
	if err != nil {
		fatal(err.Error())
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
		key := strings.ToLower(canonicalRootToken(p))
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, p)
	}
	return out
}

func canonicalRootToken(p string) string {
	p = strings.TrimSpace(p)
	p = strings.ReplaceAll(p, `/`, `\`)
	p = strings.TrimRight(p, `\`)
	if len(p) >= 2 && p[1] == ':' {
		return strings.ToUpper(p[:1]) + p[1:]
	}
	return p
}

func rootsEqual(a, b string) bool {
	return strings.EqualFold(canonicalRootToken(a), canonicalRootToken(b))
}

func initialBrowsePath(current string) string {
	parts := splitRoots(current)
	if len(parts) == 0 {
		return strings.TrimSpace(current)
	}
	p := strings.TrimSpace(parts[len(parts)-1])
	if len(p) == 2 && p[1] == ':' {
		return strings.ToUpper(p[:1]) + `:\`
	}
	return p
}

func mergePickedRoot(current, picked string) string {
	picked = strings.TrimSpace(picked)
	if picked == "" {
		return strings.TrimSpace(current)
	}
	existing := splitRoots(current)
	for _, e := range existing {
		if rootsEqual(e, picked) {
			return joinComma(existing)
		}
	}
	return joinComma(append(existing, picked))
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
	return isAffirmative(text)
}

func fatal(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}
