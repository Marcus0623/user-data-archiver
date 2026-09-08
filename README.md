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

Double-click `user-data-archiver.exe` or `start-archive.bat`. The exe is built as a Windows GUI program (`-H windowsgui`), so Explorer should not flash a console. After Copy/Cut:

- **Person name** — used in the report and in the default destination folder
- **Source** — one or more drives or folders, separated by commas or semicolons (`C:`, `C:,D:`, or `C:\Users;E:\data`). **Browse** adds a folder and does not clear paths already in the box; duplicates are skipped. Delete an entry in the box to remove it
- **Save to** — final archive folder; **Browse...** (default `D:\offboarding-archive\<name>`)
- Drive line — type (local disk / USB / network / CD-ROM), volume label, free space
- **Read/Write Test** — create, write, read, and delete `_write-test.tmp`. Refuses CD-ROM, missing/locked disks, write-protect, and access denied
- **Mode** — personal / appdata / all
- Optional: include Program Files / ProgramData; include `node_modules` and similar
- **Exclude** — extra folder names to skip at any depth (same as `-exclude`)
- **Preview only** — when checked, **Start Copy/Cut** is disabled; use **Preview**
- **Preview** — scan only, never copy or cut
- **Start Copy** or **Start Cut** — always transfers files (never a silent preview)
- **Stop**
- Log of scan sizes, progress, and errors

Administrator rights help **read** other profiles on C:. They do **not** replace write permission on the destination.

Use `-cli` or `-yes` to skip the windows. Passing `-cut` without `-cli` opens the cut window directly (no Copy/Cut chooser). The window follows the Windows display language, or use `-lang` / the Language list: `en`, `zh-CN`, `zh-TW`, `ja`, `fr`, `ru`, `vi`.

## Example

Alice is leaving. Her PC has Windows and her profile on `C:`, plus project files on `D:\Projects`. You have a USB drive `E:` (or another internal disk). You want a **copy** (keep the originals on her PC) of personal files, chat data already on disk, and the projects folder.

1. Double-click `user-data-archiver.exe`.
2. Choose **Copy files** (not Cut, unless you intend to empty the source).
3. **Person name:** `alice`
4. **Source:** `C:,D:\Projects`  
   Type it, or **Browse** `C:\` then **Browse** `D:\Projects` (the second click appends).
5. **Save to:** `E:\offboarding-archive\alice`
6. Click **Read/Write Test**. Fix the destination if it fails.
7. Leave mode on **Personal files + app data** (appdata).
8. Click **Preview**. Check the log (size by top-level folder, free space). Nothing is copied yet.
9. Click **Start Copy**. When it finishes, open the destination.

What you get (drive letters become folders because Windows cannot use `E:\C:\...`):

| On Alice’s PC | In the archive |
| --- | --- |
| `C:\Users\alice\Desktop\handoff.docx` | `E:\offboarding-archive\alice\C\Users\alice\Desktop\handoff.docx` |
| `C:\Users\alice\Documents\WeChat Files\...` | `E:\offboarding-archive\alice\C\Users\alice\Documents\WeChat Files\...` |
| `C:\Downloads\contract.pdf` | `E:\offboarding-archive\alice\C\Downloads\contract.pdf` |
| `D:\Projects\api\readme.md` | `E:\offboarding-archive\alice\D\Projects\api\readme.md` |

Also in that folder: `_archive-report.txt`. If some files fail: `_failed-files.csv`.

Same job from the command line:

```bat
user-data-archiver.exe -name alice -src C:,D:\Projects -dst E:\offboarding-archive\alice -mode appdata -yes
```

Preview only (no copy):

```bat
user-data-archiver.exe -name alice -src C:,D:\Projects -dst E:\offboarding-archive\alice -dry-run -yes
```

Do **not** set **Save to** to a folder inside a source path (for example `C:\offboarding-archive` while scanning `C:`). Put the archive on another volume.

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

`-exclude games,Steam` skips those directory names at any depth. The window has an **Exclude** field for the same list.

## Chat and messaging data

In **appdata** mode (default), local chat data is archived when it is already on this PC, for example:

- `Documents\WeChat Files`, `Documents\Tencent Files`, `Documents\WXWork`
- `AppData\Roaming` (WeChat/QQ, DingTalk, Feishu/Lark, Telegram, Slack, Teams, Discord, …)
- Folders you created on another drive, if that drive is in **Source**

Not archived, or often incomplete:

- Messages that exist only in the cloud or on a phone
- **personal** mode skips all of AppData (chat folders under Documents are still kept)
- Files locked because WeChat/QQ/Outlook/Teams is open — close the app and run again
- OneDrive online-only placeholders

This is a **file backup**, not a readable chat export. Restore usually means pointing the same app at the copied folders.

## Scan, resume, reports

Before copy/cut: file count and size by top-level folder, free-space warning. Interrupted **copy** runs on the **same destination** skip files that already match size and timestamp. **Cut** still re-reads the destination and deletes the source only when the contents match. Output: `_archive-report.txt`; failures: `_failed-files.csv`.

## Build and the executable

Keep **one** program in this folder: `user-data-archiver.exe`. After every project update, replace the old binary; do not keep extra `.exe` files.

| Action | Result |
| --- | --- |
| `build.bat` | Needs Go. Embeds `app.manifest` (Common Controls 6), compiles a GUI binary (`-H windowsgui`), **deletes leftover `.exe` files**, then keeps only `user-data-archiver.exe`. Close the running program first if Windows cannot replace it. |
| `start-archive.bat` | If Go is installed and the `.go` sources are newer than the exe (or extra `.exe` files are present), rebuilds first, then starts the GUI. |

You can copy the finished `user-data-archiver.exe` (and optionally `start-archive.bat`) to another PC that does not have Go.

## Command line

```bat
user-data-archiver.exe -name alice -src C:,D:\Projects -dst E:\offboarding-archive\alice -mode appdata -yes
user-data-archiver.exe -src C: -dst E:\offboarding-archive\alice -dry-run -yes
user-data-archiver.exe -src C: -dst E:\offboarding-archive\alice -cut -yes
user-data-archiver.exe -cli
user-data-archiver.exe -lang zh-CN
```

| Flag | Meaning |
| --- | --- |
| `-name` | Person name (report and default folder name) |
| `-src` | Source paths, comma or semicolon (`C:`, `C:,D:`, or `C:\Users;E:\data`) |
| `-dst` | Final destination folder |
| `-mode` | `personal` \| `appdata` \| `all` (default `appdata`) |
| `-include-program-files` | Also include Program Files / ProgramData |
| `-include-regeneratable` | Include `node_modules`, `__pycache__`, etc. |
| `-exclude` | Extra directory names to skip at any depth |
| `-dry-run` | Scan only; do not copy or cut |
| `-cut` | Move: delete each original only after the destination is flushed and contents match |
| `-lang` | UI language: `en`, `zh-CN`, `zh-TW`, `ja`, `fr`, `ru`, `vi` (default: Windows display language) |
| `-cli` | Command-line prompts instead of the windows |
| `-yes` | No prompts and no GUI (`-dst` and `-src` required) |

## Before you start

Authorized offboarding or your own backups only. Set OneDrive files to Always keep on this device. Close Outlook/chat if those files fail. If antivirus blocks the exe, add a trust rule (local copy/move only).
