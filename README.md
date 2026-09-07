# User Data Archiver

Windows tool that archives **non-OS user files** to another disk or USB drive and **keeps the original folder tree**. Typical use: collect a departing colleague’s personal files.

**Languages:** [English](#english) · [简体中文](#zh-cn) · [繁體中文](#zh-tw) · [日本語](#ja) · [Français](#fr) · [Русский](#ru) · [Tiếng Việt](#vi)

Standalone Go program. No other project is required. After every update, the folder should contain **only one** executable: `user-data-archiver.exe`.

---

<a id="english"></a>

## English

### What it does

Walks a source such as `C:` (or several paths) and saves files that would **not** come back after a clean Windows reinstall: Desktop, Documents, Downloads, Pictures, project folders, chat data, and — in the recommended mode — browser/app settings. OS folders, Recycle Bin, pagefile, installed programs (by default), OEM driver folders, and app caches are skipped.

### Copy or cut

The first window asks how to transfer:

| Choice | Effect |
| --- | --- |
| **Copy files** | Save to the destination; originals stay on the source disk |
| **Cut files (move)** | Save first, flush to disk, then delete each original **only after the destination is readable and byte-for-byte equal**. Matching size/timestamp alone is not enough; a corrupt same-size file is recopied first. Online-only cloud placeholders are not deleted. Cannot be undone. Does not delete `C:\`, `C:\Users`, or a user profile root |
| **Cancel** | Exit |

Cut asks for a second confirmation (default No) before changing files. Preview/scan never copies or deletes.

### Graphical window

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

### Path layout

Windows cannot create `D:\C:\Downloads`. The drive letter becomes a folder:

| Original | Archived |
| --- | --- |
| `C:\Downloads\notes.zip` | `D:\offboarding-archive\alice\C\Downloads\notes.zip` |
| `C:\Users\alice\Desktop\a.docx` | `D:\offboarding-archive\alice\C\Users\alice\Desktop\a.docx` |

Put the destination on **another disk or USB**, not the volume you are scanning. Junctions/symlinks that point at the same place are stored once. Long paths are supported.

### What is skipped vs copied

**Skipped by default:** `Windows`, Recycle Bin, `System Volume Information`, Recovery/Boot/EFI, pagefile/hiberfil/swapfile, `Program Files` / `ProgramData`, OEM folders (`Intel`, `Dell`, …), built-in profiles (`Default`, …), AppData caches (`Temp`, `Cache`, …), regeneratable folders (`node_modules`, `__pycache__`, …), `NTUSER.DAT`. `Windows.old\Users` is included; `Windows.old\Windows` is not.

**Copied by default:** user documents and custom folders such as `C:\Downloads`; in **appdata** mode also `AppData\Roaming` (chat/browser), excluding caches.

### Modes and extra options

1. **personal** — user documents, no AppData
2. **appdata** (recommended) — documents + chat/browser data, caches excluded
3. **all** — non-system custom folders; Windows and installed programs still skipped unless you include Program Files

`-exclude games,Steam` skips those directory names at any depth.

### Scan, resume, reports

Before copy/cut: file count and size by top-level folder, free-space warning. Interrupted **copy** runs on the **same destination** skip files that already match size and timestamp. **Cut** still re-reads the destination and deletes the source only when the contents match. Output: `_archive-report.txt`; failures: `_failed-files.csv`.

### Build and the executable

Keep **one** program in this folder: `user-data-archiver.exe`. After every project update, replace the old binary; do not keep extra `.exe` files.

| Action | Result |
| --- | --- |
| `build.bat` | Needs Go. Compiles a new binary, **deletes leftover `.exe` files**, then keeps only `user-data-archiver.exe`. Close the running program first if Windows cannot replace it. |
| `start-archive.bat` | If Go is installed and the `.go` sources are newer than the exe (or extra `.exe` files are present), rebuilds first, then starts the GUI. |

You can copy the finished `user-data-archiver.exe` (and optionally `start-archive.bat`) to another PC that does not have Go.

### Command line

```bat
user-data-archiver.exe -name alice -src C: -dst D:\offboarding-archive\alice -mode appdata -yes
user-data-archiver.exe -src C: -dst D:\offboarding-archive\alice -dry-run -yes
user-data-archiver.exe -src C: -dst D:\offboarding-archive\alice -cut -yes
user-data-archiver.exe -cli
```

Flags are listed at the bottom of this file.

### Before you start

Authorized offboarding or your own backups only. Set OneDrive files to Always keep on this device. Close Outlook/chat if those files fail. If antivirus blocks the exe, add a trust rule (local copy/move only).

---

<a id="zh-cn"></a>

## 简体中文

### 功能

扫描 `C:` 等源路径，保存**重装 Windows 后不会自动回来的资料**（桌面、文档、下载、项目、聊天记录；推荐模式下还有浏览器/软件配置），并保持原目录结构。默认跳过系统目录、回收站、页面文件、已安装程序、OEM 驱动目录和软件缓存。

### 复制或剪切

启动后第一个窗口：

| 选项 | 作用 |
| --- | --- |
| **Copy files** | 拷到目标盘，源文件保留 |
| **Cut files (move)** | 先保存并刷盘，**目标可读且与源逐字节一致后才删除源文件**。仅大小/时间相同不够；内容损坏会先重拷再核对。不会删除仅云端占位的文件。不可撤销。不会删除 `C:\`、`C:\Users` 或用户主目录 |
| **Cancel** | 退出 |

剪切在真正改文件前还会再确认一次（默认否）。预览/扫描不会拷贝或删除。

### 图形界面

双击 `user-data-archiver.exe` 或 `start-archive.bat`。选完复制/剪切后：

- **Person name**：报告和默认目标文件夹名
- **Source**：如 `C:`，可 Browse
- **Save to**：最终归档目录，可 Browse（默认 `D:\offboarding-archive\<姓名>`）
- 目标盘类型、卷标、剩余空间
- **Read/Write Test**：创建、写入、读回并删除 `_write-test.tmp`；光驱、未解锁、写保护、无权限会失败
- 模式 personal / appdata / all
- 可选：包含 Program Files；包含 `node_modules` 等
- **Preview only** / **Preview** / **Start Copy 或 Start Cut** / **Stop**
- 日志：体积预览、进度、错误

管理员权限用于**读取** C 盘其他用户，**不能代替**目标盘写权限。`-cli` 或 `-yes` 不弹窗。

### 路径

不能使用 `D:\C:\Downloads`。盘符变成文件夹，例如 `C:\Downloads\a.zip` → `D:\offboarding-archive\alice\C\Downloads\a.zip`。请把目标放在**另一块盘或 U 盘**。相同目标的联接/符号链接只存一份。支持超长路径。

### 跳过与拷贝

**默认跳过：** `Windows`、回收站、系统卷、Recovery/Boot/EFI、pagefile/hiberfil、`Program Files`/`ProgramData`、OEM 目录、`Default` 等系统用户、AppData 缓存、`node_modules` 等可再生成目录、`NTUSER.DAT`。会收 `Windows.old\Users`，不收 `Windows.old\Windows`。

**默认拷贝：** 用户文档和 `C:\Downloads` 等自定义目录；**appdata** 模式还包含 `AppData\Roaming`（不含缓存）。

### 模式

1. **personal**：仅文档，不含 AppData
2. **appdata**（推荐）：文档 + 聊天/浏览器，排除缓存
3. **all**：非系统自定义目录；除非勾选，否则仍跳过已安装程序

`-exclude games,Steam` 按目录名额外排除。

### 扫描、续传、报告

按顶层目录显示文件数和体积，空间不足会警告。中断后对**同一目标**再跑，复制模式会跳过大小和时间戳相同的文件。**剪切**仍会再读目标，内容一致才删源。报告 `_archive-report.txt`，失败 `_failed-files.csv`。

### 编译与可执行文件

本目录应只保留 **一份** 程序：`user-data-archiver.exe`。项目每次更新后都要换掉旧 exe，不要留下多份。

| 操作 | 结果 |
| --- | --- |
| `build.bat` | 需要已安装 Go。先编出新文件，**删掉多余的 `.exe`**，只留下 `user-data-archiver.exe`。若程序正在运行导致无法替换，请先退出再编。 |
| `start-archive.bat` | 若已安装 Go，且源码比 exe 新（或目录里还有其它 `.exe`），会先重新编译再打开界面。 |

编好后可以把 `user-data-archiver.exe`（以及可选的 `start-archive.bat`）拷到没有安装 Go 的电脑上使用。

### 命令行

```bat
user-data-archiver.exe -name alice -src C: -dst D:\offboarding-archive\alice -mode appdata -yes
user-data-archiver.exe -src C: -dst D:\offboarding-archive\alice -dry-run -yes
user-data-archiver.exe -src C: -dst D:\offboarding-archive\alice -cut -yes
user-data-archiver.exe -cli
```

完整参数见文末表格。仅用于授权交接或本人备份。OneDrive「仅联机」请改为始终保留在此设备。Outlook/聊天软件打开中的文件可能失败。

---

<a id="zh-tw"></a>

## 繁體中文

### 功能

掃描 `C:` 等來源，保存**重裝 Windows 後不會自動回來的資料**（桌面、文件、下載、專案、聊天紀錄；建議模式下還有瀏覽器／軟體設定），並維持原資料夾結構。預設略過系統目錄、資源回收筒、分頁檔、已安裝程式、OEM 驅動目錄與快取。

### 複製或剪下

啟動後第一個視窗：

| 選項 | 作用 |
| --- | --- |
| **Copy files** | 存到目的地，原始檔保留 |
| **Cut files (move)** | 先寫入並刷盤，**目的檔可讀且與來源逐位元組相同後才刪除原始檔**。僅大小／時間相同不夠。不刪除僅雲端佔位檔。無法復原。不會刪除 `C:\`、`C:\Users` 或使用者主目錄 |
| **Cancel** | 離開 |

剪下在真正變更檔案前會再確認（預設否）。預覽／掃描不會複製或刪除。

### 圖形介面

按兩下 `user-data-archiver.exe` 或 `start-archive.bat`。選擇複製／剪下後可設定姓名、來源、**Save to**（Browse）、磁碟類型／標籤／剩餘空間、**Read/Write Test**、模式、是否包含 Program Files 與 `node_modules`、僅預覽、開始／停止與日誌。系統管理員用來**讀取** C 槽其他使用者，**不能取代**目的碟寫入權限。`-cli` 或 `-yes` 不開視窗。

### 路徑與規則

不可使用 `D:\C:\Downloads`。磁碟機代號變成資料夾。請把目的地放在**另一顆碟或 USB**。預設略過 `Windows`、回收筒、`Program Files`、OEM、快取、`node_modules` 等；**appdata** 會收 `AppData\Roaming`（不含快取）。會收 `Windows.old\Users`。複製中斷後對同一目的地可依大小與時間續傳；剪下仍會核對內容。報告 `_archive-report.txt`，失敗 `_failed-files.csv`。

### 編譯與執行檔

此資料夾只應留下一份 `user-data-archiver.exe`。`build.bat` 會刪除多餘的 `.exe` 再寫入最新版。`start-archive.bat` 在原始碼較新時會先重新編譯。

### 模式與命令列

**personal**／**appdata**（建議）／**all**。剪下：`-cut`。僅預覽：`-dry-run`。互動命令列：`-cli`。參數表在文末。僅供授權交接或本人備份。

---

<a id="ja"></a>

## 日本語

### 機能

`C:` などを走査し、**再インストールしても戻らないデータ**（デスクトップ、文書、ダウンロード、プロジェクト、チャット、推奨モードではブラウザ設定）を元のフォルダー構造のまま保存します。OS、ごみ箱、ページファイル、インストール済みプログラム、OEM、キャッシュは既定で除外します。

### コピーまたはカット

最初のウィンドウ：

| 選択 | 動作 |
| --- | --- |
| **Copy files** | 保存先へコピーし、元ファイルは残す |
| **Cut files (move)** | 書き込み・フラッシュ後、**保存先が読め、内容が一致したファイルだけ**元を削除。サイズ／時刻が同じだけでは不十分。オンラインのみのプレースホルダーは削除しない。取り消し不可。`C:\`、`C:\Users`、プロファイル直下は削除しない |
| **Cancel** | 終了 |

カットは実行前に再確認します（既定はいいえ）。プレビューはコピーも削除もしません。

### GUI

`user-data-archiver.exe` または `start-archive.bat` をダブルクリック。氏名、ソース、**Save to**（Browse）、ドライブ種別／ラベル／空き容量、**Read/Write Test**、モード、Program Files / `node_modules` の含める指定、プレビュー、開始／停止、ログ。管理者権限は C: の他ユーザーを**読む**ためで、保存先の書き込み権限の代わりにはなりません。`-cli` / `-yes` でウィンドウなし。

### パス・除外・再開

`D:\C:\Downloads` は不可。ドライブレターがフォルダーになります。保存先は**別ディスクまたは USB**。既定で `Windows`、`Program Files`、キャッシュ、`node_modules` などを除外。**appdata** は `AppData\Roaming` を含む（キャッシュ除く）。`Windows.old\Users` は対象。コピーの再開は同一保存先でサイズとタイムスタンプが同じファイルをスキップ。カットは内容照合後にだけ削除。報告 `_archive-report.txt`、失敗 `_failed-files.csv`。

### ビルドと実行ファイル

このフォルダーには `user-data-archiver.exe` を **1 つだけ**置きます。`build.bat` は余った `.exe` を削除して最新版だけ残します。`start-archive.bat` はソースが新しいと再ビルドします。

### モードと CLI

**personal** / **appdata**（推奨）/ **all**。移動は `-cut`、プレビューは `-dry-run`。フラグ表は末尾。許可された引き継ぎまたは自分のバックアップ専用。

---

<a id="fr"></a>

## Français

### Fonction

Parcourt `C:` (ou d’autres chemins) et enregistre ce qui **ne reviendrait pas** après une réinstallation propre : bureau, documents, téléchargements, projets, messagerie, et en mode recommandé les réglages navigateur. Ignore par défaut l’OS, la Corbeille, le fichier d’échange, les programmes installés, les dossiers OEM et les caches.

### Copier ou couper

Première fenêtre :

| Choix | Effet |
| --- | --- |
| **Copy files** | Enregistre la destination ; les originaux restent |
| **Cut files (move)** | Enregistre, vide les tampons, puis **supprime l’original seulement si la destination est lisible et identique octet par octet**. Taille/date identiques ne suffisent pas. Les fichiers cloud « en ligne seulement » ne sont pas supprimés. Irréversible. Ne supprime pas `C:\`, `C:\Users` ni la racine d’un profil |
| **Cancel** | Quitter |

Le mode couper redemande confirmation (Non par défaut). L’aperçu ne copie ni ne supprime.

### Fenêtre

Double-clic sur `user-data-archiver.exe` ou `start-archive.bat`. Ensuite : nom, source, **Save to** (Browse), type/libellé/espace libre du disque, **Read/Write Test**, mode, options Program Files / `node_modules`, aperçu, démarrer / arrêter, journal. Les droits administrateur servent à **lire** les autres profils sur C:, pas à écrire sur la destination. `-cli` ou `-yes` : pas de fenêtre.

### Chemins, filtres, reprise

Pas de `D:\C:\Downloads` : la lettre de lecteur devient un dossier. Destination sur **un autre disque ou USB**. Ignore `Windows`, `Program Files`, caches, `node_modules`, etc. Mode **appdata** : `AppData\Roaming` sans caches. `Windows.old\Users` est inclus. Reprise de copie sur la **même destination** si taille et date identiques. Couper ne supprime qu’après comparaison du contenu. Rapport `_archive-report.txt` ; échecs `_failed-files.csv`.

### Compilation et exécutable

Un seul `user-data-archiver.exe` dans ce dossier. `build.bat` supprime les autres `.exe` et n’en garde qu’un. `start-archive.bat` recompile si les sources sont plus récentes.

### Modes et ligne de commande

**personal** / **appdata** (recommandé) / **all**. Déplacer : `-cut`. Aperçu : `-dry-run`. Tableau des options en bas. Usage autorisé uniquement.

---

<a id="ru"></a>

## Русский

### Назначение

Обходит `C:` и сохраняет то, **чего не будет после чистой переустановки**: рабочий стол, документы, загрузки, проекты, мессенджеры и в рекомендуемом режиме настройки браузера. По умолчанию пропускаются ОС, Корзина, файл подкачки, установленные программы, OEM и кэши.

### Копирование или вырезание

Первое окно:

| Выбор | Действие |
| --- | --- |
| **Copy files** | Сохранить копию; оригиналы остаются |
| **Cut files (move)** | Сохранить, сбросить на диск, затем **удалить оригинал только если назначение читается и побайтно совпадает**. Совпадения размера и времени недостаточно. Облачные заглушки не удаляются. Необратимо. Не удаляет `C:\`, `C:\Users` и корень профиля |
| **Cancel** | Выход |

Перед вырезанием — повторное подтверждение (по умолчанию Нет). Просмотр ничего не копирует и не удаляет.

### Окно

Двойной щелчок по `user-data-archiver.exe` или `start-archive.bat`. Далее: имя, источник, **Save to**, тип/метка/свободное место диска, **Read/Write Test**, режим, Program Files / `node_modules`, просмотр, старт / стоп, журнал. Права администратора нужны, чтобы **читать** чужие профили на C:, а не чтобы писать на диск назначения. `-cli` / `-yes` — без окон.

### Пути, исключения, докачка

Нельзя `D:\C:\Downloads`: буква диска становится папкой. Пишите на **другой диск или USB**. Пропускаются `Windows`, `Program Files`, кэши, `node_modules`. В **appdata** копируется `AppData\Roaming` без кэша. `Windows.old\Users` включается. Повтор **копирования** на ту же папку пропускает файлы с тем же размером и временем. Вырезание удаляет только после сверки содержимого. Отчёт `_archive-report.txt`, ошибки `_failed-files.csv`.

### Сборка и exe

В папке должен быть **один** `user-data-archiver.exe`. `build.bat` удаляет лишние `.exe` и оставляет свежую сборку. `start-archive.bat` пересобирает, если исходники новее.

### Режимы и CLI

**personal** / **appdata** (рекомендуется) / **all**. Перенос: `-cut`. Просмотр: `-dry-run`. Таблица флагов в конце. Только для разрешённой передачи дел или своей копии.

---

<a id="vi"></a>

## Tiếng Việt

### Chức năng

Quét `C:` (hoặc nhiều đường dẫn) và lưu những gì **sẽ không còn sau khi cài lại Windows**: Desktop, tài liệu, Downloads, dự án, chat, và ở chế độ khuyến nghị cả cấu hình trình duyệt. Mặc định bỏ OS, Thùng rác, pagefile, chương trình đã cài, OEM và cache.

### Copy hoặc Cut

Cửa sổ đầu tiên:

| Chọn | Tác dụng |
| --- | --- |
| **Copy files** | Lưu sang đích, giữ file gốc |
| **Cut files (move)** | Ghi xong, flush đĩa, **chỉ xóa file gốc khi đích đọc được và trùng từng byte**. Trùng size/thời gian chưa đủ. Không xóa file chỉ trên cloud. Không hoàn tác. Không xóa `C:\`, `C:\Users` hay thư mục gốc hồ sơ |
| **Cancel** | Thoát |

Cut hỏi lại trước khi đổi file (mặc định No). Preview không copy và không xóa.

### Giao diện

Nhấp đúp `user-data-archiver.exe` hoặc `start-archive.bat`. Sau đó: tên, nguồn, **Save to** (Browse), loại ổ / nhãn / dung lượng trống, **Read/Write Test**, mode, tùy chọn Program Files / `node_modules`, preview, Start / Stop, nhật ký. Quyền administrator để **đọc** hồ sơ khác trên C:, không thay quyền ghi ổ đích. `-cli` hoặc `-yes` không mở cửa sổ.

### Đường dẫn, lọc, tiếp tục

Không dùng `D:\C:\Downloads`. Ký tự ổ thành thư mục. Đích nên là **ổ khác hoặc USB**. Bỏ `Windows`, `Program Files`, cache, `node_modules`. **appdata** gồm `AppData\Roaming` (không cache). Có `Windows.old\Users`. Chạy lại copy **cùng thư mục đích** sẽ bỏ qua file trùng size và thời gian. Cut chỉ xóa sau khi đối chiếu nội dung. Báo cáo `_archive-report.txt`, lỗi `_failed-files.csv`.

### Biên dịch và file exe

Chỉ giữ **một** `user-data-archiver.exe` trong thư mục. `build.bat` xóa các `.exe` cũ rồi ghi bản mới. `start-archive.bat` biên dịch lại nếu mã nguồn mới hơn.

### Chế độ và dòng lệnh

**personal** / **appdata** (khuyến nghị) / **all**. Di chuyển: `-cut`. Chỉ xem: `-dry-run`. Bảng cờ ở cuối. Chỉ dùng khi được phép hoặc sao lưu của bạn.

---

## Command-line flags

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

**Build:** `build.bat` (needs Go) deletes leftover `.exe` files and writes a new `user-data-archiver.exe`. **Launch:** `start-archive.bat` or the exe (`start-archive.bat` rebuilds first if sources are newer). Default destination parent: `D:\offboarding-archive`.
