# User Data Archiver

<p>
  <kbd><b>English</b></kbd>
  &nbsp;<a href="README.zh-CN.md"><kbd>简体中文</kbd></a>
  &nbsp;<a href="README.zh-TW.md"><kbd>繁體中文</kbd></a>
  &nbsp;<a href="README.ja.md"><kbd>日本語</kbd></a>
  &nbsp;<a href="README.fr.md"><kbd>Français</kbd></a>
  &nbsp;<a href="README.ru.md"><kbd>Русский</kbd></a>
  &nbsp;<a href="README.vi.md"><kbd>Tiếng Việt</kbd></a>
</p>

Windows tool that archives **non-OS user files** to another disk or USB drive and **keeps the original folder tree**. Typical use: collect a departing colleague’s personal files.

Standalone Go program. No other project is required. After every update, the folder should contain **only one** executable: `user-data-archiver.exe`.

## What it does

Walks a source such as `C:` (or several paths) and saves files that would **not** come back after a clean Windows reinstall: Desktop, Documents, Downloads, Pictures, project folders, chat data, and — in the recommended mode — browser/app settings. OS folders, Recycle Bin, pagefile, installed programs (by default), OEM driver folders, and app caches are skipped.

## Copy or cut

The first window asks how to transfer:

| Choice | Effect |
| --- | --- |
| **Copy files** | Save to the destination; originals stay on the source disk |
| **Cut files (move)** | Save first, flush to disk, then delete each original **only after the destination is readable and byte-for-byte equal**. Matching size/timestamp alone is not enough; a corrupt same-size file is recopied first. Online-only cloud placeholders are not deleted. Cannot be undone. Does not delete `C:\`, `C:\Users`, or a user profile root |
| **Cancel** | Exit |

Cut asks for a second confirmation (default No) before changing files. Preview/scan never copies or deletes.

## Graphical window

Double-click `user-data-archiver.exe` or `start-archive.bat`. After Copy/Cut:

- **Person name** — used in the report and in the default destination folder
- **Source** — e.g. `C:` or `C:\Users`; **Browse...**
- **Save to** — final archive folder; **Browse...** (default `D:\offboarding-archive\<name>`)
- Drive line — type (local disk / USB / network / CD-ROM), volume label, free space
- **Read/Write Test** — create, write, read, and delete `_write-test.tmp`. Refuses CD-ROM, missing/locked disks, write-protect, and access denied
- **Mode** — personal / appdata / all
- Optional: include Program Files / ProgramData; include `node_modules` and similar
- **Preview only** — scan without copy or cut
- **Preview** / **Start Copy** or **Start Cut** / **Stop**
- Log of scan sizes, progress, and errors

Administrator rights help **read** other profiles on C:. They do **not** replace write permission on the destination.

Use `-cli` or `-yes` to skip the windows.

## Path layout

Windows cannot create `D:\C:\Downloads`. The drive letter becomes a folder:

| Original | Archived |
| --- | --- |
| `C:\Downloads\notes.zip` | `D:\offboarding-archive\alice\C\Downloads\notes.zip` |
| `C:\Users\alice\Desktop\a.docx` | `D:\offboarding-archive\alice\C\Users\alice\Desktop\a.docx` |

Put the destination on **another disk or USB**, not the volume you are scanning. Junctions/symlinks that point at the same place are stored once. Long paths are supported.

## What is skipped vs copied

**Skipped by default:** `Windows`, Recycle Bin, `System Volume Information`, Recovery/Boot/EFI, pagefile/hiberfil/swapfile, `Program Files` / `ProgramData`, OEM folders (`Intel`, `Dell`, …), built-in profiles (`Default`, …), AppData caches (`Temp`, `Cache`, …), regeneratable folders (`node_modules`, `__pycache__`, …), `NTUSER.DAT`. `Windows.old\Users` is included; `Windows.old\Windows` is not.

**Copied by default:** user documents and custom folders such as `C:\Downloads`; in **appdata** mode also `AppData\Roaming` (chat/browser), excluding caches.

## Modes and extra options

1. **personal** — user documents, no AppData
2. **appdata** (recommended) — documents + chat/browser data, caches excluded
3. **all** — non-system custom folders; Windows and installed programs still skipped unless you include Program Files

`-exclude games,Steam` skips those directory names at any depth.

## Scan, resume, reports

Before copy/cut: file count and size by top-level folder, free-space warning. Interrupted **copy** runs on the **same destination** skip files that already match size and timestamp. **Cut** still re-reads the destination and deletes the source only when the contents match. Output: `_archive-report.txt`; failures: `_failed-files.csv`.

## Build and the executable

Keep **one** program in this folder: `user-data-archiver.exe`. After every project update, replace the old binary; do not keep extra `.exe` files.

| Action | Result |
| --- | --- |
| `build.bat` | Needs Go. Compiles a new binary, **deletes leftover `.exe` files**, then keeps only `user-data-archiver.exe`. Close the running program first if Windows cannot replace it. |
| `start-archive.bat` | If Go is installed and the `.go` sources are newer than the exe (or extra `.exe` files are present), rebuilds first, then starts the GUI. |

You can copy the finished `user-data-archiver.exe` (and optionally `start-archive.bat`) to another PC that does not have Go.

## Command line

```bat
user-data-archiver.exe -name alice -src C: -dst D:\offboarding-archive\alice -mode appdata -yes
user-data-archiver.exe -src C: -dst D:\offboarding-archive\alice -dry-run -yes
user-data-archiver.exe -src C: -dst D:\offboarding-archive\alice -cut -yes
user-data-archiver.exe -cli
```

| Flag | Meaning |
| --- | --- |
| `-name` | Person name (report and default folder name) |
| `-src` | Source paths, comma-separated (`C:`, `C:,D:`, or `C:\Users`) |
| `-dst` | Final destination folder |
| `-mode` | `personal` \| `appdata` \| `all` (default `appdata`) |
| `-include-program-files` | Also include Program Files / ProgramData |
| `-include-regeneratable` | Include `node_modules`, `__pycache__`, etc. |
| `-exclude` | Extra directory names to skip at any depth |
| `-dry-run` | Scan only; do not copy or cut |
| `-cut` | Move: delete each original only after the destination is flushed and contents match |
| `-cli` | Command-line prompts instead of the windows |
| `-yes` | No prompts and no GUI (`-dst` and `-src` required) |

## Before you start

Authorized offboarding or your own backups only. Set OneDrive files to Always keep on this device. Close Outlook/chat if those files fail. If antivirus blocks the exe, add a trust rule (local copy/move only).
