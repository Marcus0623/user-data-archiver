# User Data Archiver

<p>
  <a href="README.md"><kbd>English</kbd></a>
  &nbsp;<a href="README.zh-CN.md"><kbd>简体中文</kbd></a>
  &nbsp;<kbd><b>繁體中文</b></kbd>
  &nbsp;<a href="README.ja.md"><kbd>日本語</kbd></a>
  &nbsp;<a href="README.fr.md"><kbd>Français</kbd></a>
  &nbsp;<a href="README.ru.md"><kbd>Русский</kbd></a>
  &nbsp;<a href="README.vi.md"><kbd>Tiếng Việt</kbd></a>
</p>

Windows 工具：將**非系統使用者檔案**歸檔到另一顆硬碟或 USB，並**維持原資料夾結構**。典型用途：收集離職同事的個人資料。

獨立 Go 程式，不依賴其他專案。每次更新後，此資料夾應只留下 **一份** 執行檔：`user-data-archiver.exe`。

## 功能

掃描 `C:` 等來源，保存**重裝 Windows 後不會自動回來的資料**（桌面、文件、下載、專案、聊天紀錄；建議模式下還有瀏覽器／軟體設定），並維持原資料夾結構。預設略過系統目錄、資源回收筒、分頁檔、已安裝程式、OEM 驅動目錄與快取。

## 複製或剪下

啟動後第一個視窗：

| 選項 | 作用 |
| --- | --- |
| **Copy files** | 存到目的地，原始檔保留 |
| **Cut files (move)** | 先寫入並刷盤，**目的檔可讀且與來源逐位元組相同後才刪除原始檔**。僅大小／時間相同不夠。不刪除僅雲端佔位檔。無法復原。不會刪除 `C:\`、`C:\Users` 或使用者主目錄 |
| **Cancel** | 離開 |

剪下在真正變更檔案前會再確認（預設否）。預覽／掃描不會複製或刪除。

## 圖形介面

按兩下 `user-data-archiver.exe` 或 `start-archive.bat`。選擇複製／剪下後可設定姓名、來源、**儲存到**（瀏覽）、**排除**、磁碟類型／標籤／剩餘空間、**讀寫測試**、模式、是否包含 Program Files 與 `node_modules`、僅預覽、開始／停止與日誌。來源可填多個路徑（逗號或分號，例如 `C:,D:`）；**瀏覽**會追加，不會清空已有路徑。勾選 **僅預覽** 時請用 **預覽**；**開始**一定是真正傳檔。系統管理員用來**讀取** C 槽其他使用者，**不能取代**目的碟寫入權限。`-cli` 或 `-yes` 不開視窗。雙擊程式時若加 `-cut`，會直接進入剪下視窗。介面語言可用 `-lang`：`en`、`zh-CN`、`zh-TW`、`ja`、`fr`、`ru`、`vi`。

## 範例

同事 Alice 離職。系統與個人資料在 `C:`，專案在 `D:\Projects`。U 盤為 `E:`。目標是**複製**（保留原始檔）。

1. 按兩下 `user-data-archiver.exe`，選 **複製檔案**。
2. 姓名填 `alice`。來源填 `C:,D:\Projects`（或先瀏覽 `C:\` 再瀏覽 `D:\Projects`，第二次會追加）。
3. 儲存到 `E:\offboarding-archive\alice`，先做讀寫測試，再 **預覽**，確認無誤後 **開始複製**。

| 原路徑 | 歸檔後 |
| --- | --- |
| `C:\Users\alice\Desktop\handoff.docx` | `E:\offboarding-archive\alice\C\Users\alice\Desktop\handoff.docx` |
| `C:\Downloads\contract.pdf` | `E:\offboarding-archive\alice\C\Downloads\contract.pdf` |
| `D:\Projects\api\readme.md` | `E:\offboarding-archive\alice\D\Projects\api\readme.md` |

同一資料夾會有 `_archive-report.txt`。命令列：

```bat
user-data-archiver.exe -lang zh-TW -name alice -src C:,D:\Projects -dst E:\offboarding-archive\alice -mode appdata -yes
```

請把目的地放在**另一顆碟或 USB**，不要放在正在掃描的來源裡面。

## 聊天紀錄

**appdata**（建議）會收本機已落地的聊天資料，例如 `Documents\WeChat Files`、`AppData\Roaming`（微信／QQ、釘釘、飛書、Telegram、Teams 等）。只在雲端或手機裡的訊息、**personal** 模式下的 AppData、程式開著被鎖的檔、OneDrive 僅雲端檔可能沒有或不完整。這是檔案備份，不是可讀的聊天匯出。

## 路徑與規則

不可使用 `D:\C:\Downloads`。磁碟機代號變成資料夾。請把目的地放在**另一顆碟或 USB**。預設略過 `Windows`、回收筒、`Program Files`、OEM、快取、`node_modules` 等；**appdata** 會收 `AppData\Roaming`（聊天／瀏覽器，不含快取）。會收 `Windows.old\Users`。複製中斷後對同一目的地可依大小與時間續傳；剪下仍會核對內容。報告 `_archive-report.txt`，失敗 `_failed-files.csv`。

## 編譯與執行檔

此資料夾只應留下一份 `user-data-archiver.exe`。`build.bat` 會刪除多餘的 `.exe` 再寫入最新版。`start-archive.bat` 在原始碼較新時會先重新編譯。

## 模式與命令列

**personal**／**appdata**（建議）／**all**。剪下：`-cut`。僅預覽：`-dry-run`。互動命令列：`-cli`。

```bat
user-data-archiver.exe -name alice -src C:,D:\Projects -dst E:\offboarding-archive\alice -mode appdata -yes
user-data-archiver.exe -src C: -dst E:\offboarding-archive\alice -dry-run -yes
user-data-archiver.exe -src C: -dst E:\offboarding-archive\alice -cut -yes
user-data-archiver.exe -cli
user-data-archiver.exe -lang zh-TW
```

| 參數 | 含義 |
| --- | --- |
| `-name` | 姓名（報告與預設資料夾名） |
| `-src` | 來源路徑，逗號分隔 |
| `-dst` | 最終歸檔目錄 |
| `-mode` | `personal` \| `appdata` \| `all`（預設 `appdata`） |
| `-include-program-files` | 同時包含 Program Files / ProgramData |
| `-include-regeneratable` | 包含 `node_modules`、`__pycache__` 等 |
| `-exclude` | 依目錄名額外排除 |
| `-dry-run` | 只掃描，不複製、不剪下 |
| `-cut` | 移動：目的檔刷盤且內容相符後才刪除原始檔 |
| `-lang` | 介面語言：`en`、`zh-CN`、`zh-TW`、`ja`、`fr`、`ru`、`vi` |
| `-cli` | 命令列問答，不開視窗 |
| `-yes` | 不提問、不開視窗（必須提供 `-dst` 與 `-src`） |

僅供授權交接或本人備份。
