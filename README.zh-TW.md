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

按兩下 `user-data-archiver.exe` 或 `start-archive.bat`。選擇複製／剪下後可設定姓名、來源、**Save to**（Browse）、磁碟類型／標籤／剩餘空間、**Read/Write Test**、模式、是否包含 Program Files 與 `node_modules`、僅預覽、開始／停止與日誌。系統管理員用來**讀取** C 槽其他使用者，**不能取代**目的碟寫入權限。`-cli` 或 `-yes` 不開視窗。

## 路徑與規則

不可使用 `D:\C:\Downloads`。磁碟機代號變成資料夾。請把目的地放在**另一顆碟或 USB**。預設略過 `Windows`、回收筒、`Program Files`、OEM、快取、`node_modules` 等；**appdata** 會收 `AppData\Roaming`（聊天／瀏覽器，不含快取）。會收 `Windows.old\Users`。複製中斷後對同一目的地可依大小與時間續傳；剪下仍會核對內容。報告 `_archive-report.txt`，失敗 `_failed-files.csv`。

## 編譯與執行檔

此資料夾只應留下一份 `user-data-archiver.exe`。`build.bat` 會刪除多餘的 `.exe` 再寫入最新版。`start-archive.bat` 在原始碼較新時會先重新編譯。

## 模式與命令列

**personal**／**appdata**（建議）／**all**。剪下：`-cut`。僅預覽：`-dry-run`。互動命令列：`-cli`。

```bat
user-data-archiver.exe -name alice -src C: -dst D:\offboarding-archive\alice -mode appdata -yes
user-data-archiver.exe -src C: -dst D:\offboarding-archive\alice -dry-run -yes
user-data-archiver.exe -src C: -dst D:\offboarding-archive\alice -cut -yes
user-data-archiver.exe -cli
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
| `-cli` | 命令列問答，不開視窗 |
| `-yes` | 不提問、不開視窗（必須提供 `-dst` 與 `-src`） |

僅供授權交接或本人備份。
