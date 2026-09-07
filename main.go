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
	setupConsole()

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
	flag.Parse()

	if !*cliFlag && !*yes {
		if consoleProcessCount() <= 1 {
			hideConsoleWindow()
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
		})
		return
	}

	runCLI(*dstFlag, *nameFlag, *srcFlag, *modeFlag, *includePF, *includeRegen, *excludeFlag, *dryRun, *yes, *cutFlag)
}

func runCLI(dstFlag, nameFlag, srcFlag, modeFlag string, includePF, includeRegen bool, excludeFlag string, dryRun, yes, cut bool) {
	in := bufio.NewReader(os.Stdin)

	fmt.Println("========================================")
	fmt.Println("  User Data Archiver")
	fmt.Println("========================================")
	fmt.Printf("Computer: %s\n", computerName())
	if isAdmin() {
		fmt.Println("Rights: running as administrator (can read other user profiles)")
	} else {
		fmt.Println("Rights: not administrator. Right-click and Run as administrator to archive all local user profiles.")
	}
	fmt.Println()

	employee := strings.TrimSpace(nameFlag)
	if employee == "" {
		if yes {
			employee = defaultName()
		} else {
			employee = prompt(in, "Person name (used in the folder name)", defaultName())
		}
	}
	employee = sanitizeEmployeeName(employee)

	dest := strings.TrimSpace(dstFlag)
	if dest == "" {
		if yes {
			fatal("-yes requires -dst")
		}
		parent := prompt(in, "Destination parent folder (example D:\\offboarding-archive)", defaultDestParent)
		dest = filepath.Join(strings.TrimSpace(parent), employee)
	}
	dest = normalizeAbs(dest)
	fmt.Printf("Destination drive: %s\n", destDriveInfo(dest))
	fmt.Printf("Destination folder: %s\n", dest)
	fmt.Print("Checking that the destination is writable... ")
	if err := probeDestWritable(dest); err != nil {
		fmt.Println("FAILED")
		fatal(err.Error() + "\nIf you lack permission: pick a folder where you can create files; on USB drives turn off the write-protect switch and unlock BitLocker; or ask an admin for Modify rights. Running as administrator is for reading other profiles on C:, and does not replace write access to the destination.")
	}
	fmt.Println("OK")

	srcText := strings.TrimSpace(srcFlag)
	if srcText == "" {
		if yes {
			fatal("-yes requires -src")
		}
		srcText = prompt(in, "Source path (comma-separated, drive letter like C:)", "C:")
	}
	roots := splitRoots(srcText)
	if len(roots) == 0 {
		fatal("no source path specified")
	}

	mode := ModeWithAppData
	if strings.TrimSpace(modeFlag) != "" {
		mode = ParseMode(modeFlag)
	} else if !yes {
		fmt.Println()
		fmt.Println("Archive mode:")
		fmt.Println("  1) " + ModePersonal.Title())
		fmt.Println("  2) " + ModeWithAppData.Title())
		fmt.Println("  3) " + ModeAllNonSystem.Title())
		mode = ParseMode(prompt(in, "Choose", "2"))
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
			fmt.Printf("\r  %d/%d files%s  %s  %s          ",
				doneFiles, totalFiles, pct, formatBytes(doneBytes), shortenPath(src, 60))
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
	return text == "y" || text == "yes"
}

func fatal(msg string) {
	fmt.Fprintln(os.Stderr, msg)
	os.Exit(1)
}
