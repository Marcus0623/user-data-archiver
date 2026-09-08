# User Data Archiver

<p>
  <a href="README.md"><kbd>English</kbd></a>
  &nbsp;<kbd><b>简体中文</b></kbd>
  &nbsp;<a href="README.zh-TW.md"><kbd>繁體中文</kbd></a>
  &nbsp;<a href="README.ja.md"><kbd>日本語</kbd></a>
  &nbsp;<a href="README.fr.md"><kbd>Français</kbd></a>
  &nbsp;<a href="README.ru.md"><kbd>Русский</kbd></a>
  &nbsp;<a href="README.vi.md"><kbd>Tiếng Việt</kbd></a>
</p>

Windows 工具：把**非系统用户文件**归档到另一块硬盘或 U 盘，并**保持原目录结构**。典型用途：收集离职同事的个人资料。

独立 Go 程序，不依赖其他项目。每次更新后，本目录应只保留 **一份** 可执行文件：`user-data-archiver.exe`。

## 功能

扫描 `C:` 等源路径，保存**重装 Windows 后不会自动回来的资料**（桌面、文档、下载、项目、聊天记录；推荐模式下还有浏览器/软件配置），并保持原目录结构。默认跳过系统目录、回收站、页面文件、已安装程序、OEM 驱动目录和软件缓存。

## 复制或剪切

启动后第一个窗口：

| 选项 | 作用 |
| --- | --- |
| **Copy files** | 拷到目标盘，源文件保留 |
| **Cut files (move)** | 先保存并刷盘，**目标可读且与源逐字节一致后才删除源文件**。仅大小/时间相同不够；内容损坏会先重拷再核对。不会删除仅云端占位的文件。不可撤销。不会删除 `C:\`、`C:\Users` 或用户主目录 |
| **Cancel** | 退出 |

剪切在真正改文件前还会再确认一次（默认否）。预览/扫描不会拷贝或删除。

## 图形界面

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

## 路径

不能使用 `D:\C:\Downloads`。盘符变成文件夹，例如 `C:\Downloads\a.zip` → `D:\offboarding-archive\alice\C\Downloads\a.zip`。请把目标放在**另一块盘或 U 盘**。相同目标的联接/符号链接只存一份。支持超长路径。

## 跳过与拷贝

**默认跳过：** `Windows`、回收站、系统卷、Recovery/Boot/EFI、pagefile/hiberfil、`Program Files`/`ProgramData`、OEM 目录、`Default` 等系统用户、AppData 缓存、`node_modules` 等可再生成目录、`NTUSER.DAT`。会收 `Windows.old\Users`，不收 `Windows.old\Windows`。

**默认拷贝：** 用户文档和 `C:\Downloads` 等自定义目录；**appdata** 模式还包含 `AppData\Roaming`（聊天/浏览器，不含缓存）。

## 模式

1. **personal**：仅文档，不含 AppData
2. **appdata**（推荐）：文档 + 聊天/浏览器，排除缓存
3. **all**：非系统自定义目录；除非勾选，否则仍跳过已安装程序

`-exclude games,Steam` 按目录名额外排除。

## 扫描、续传、报告

按顶层目录显示文件数和体积，空间不足会警告。中断后对**同一目标**再跑，复制模式会跳过大小和时间戳相同的文件。**剪切**仍会再读目标，内容一致才删源。报告 `_archive-report.txt`，失败 `_failed-files.csv`。

## 编译与可执行文件

本目录应只保留 **一份** 程序：`user-data-archiver.exe`。项目每次更新后都要换掉旧 exe，不要留下多份。

| 操作 | 结果 |
| --- | --- |
| `build.bat` | 需要已安装 Go。先编出新文件，**删掉多余的 `.exe`**，只留下 `user-data-archiver.exe`。若程序正在运行导致无法替换，请先退出再编。 |
| `start-archive.bat` | 若已安装 Go，且源码比 exe 新（或目录里还有其它 `.exe`），会先重新编译再打开界面。 |

编好后可以把 `user-data-archiver.exe`（以及可选的 `start-archive.bat`）拷到没有安装 Go 的电脑上使用。

## 命令行

```bat
user-data-archiver.exe -name alice -src C: -dst D:\offboarding-archive\alice -mode appdata -yes
user-data-archiver.exe -src C: -dst D:\offboarding-archive\alice -dry-run -yes
user-data-archiver.exe -src C: -dst D:\offboarding-archive\alice -cut -yes
user-data-archiver.exe -cli
```

| 参数 | 含义 |
| --- | --- |
| `-name` | 姓名（报告和默认文件夹名） |
| `-src` | 源路径，逗号分隔（`C:`、`C:,D:` 或 `C:\Users`） |
| `-dst` | 最终归档目录 |
| `-mode` | `personal` \| `appdata` \| `all`（默认 `appdata`） |
| `-include-program-files` | 同时包含 Program Files / ProgramData |
| `-include-regeneratable` | 包含 `node_modules`、`__pycache__` 等 |
| `-exclude` | 按目录名额外排除（任意深度） |
| `-dry-run` | 只扫描，不拷贝、不剪切 |
| `-cut` | 移动：目标刷盘且内容一致后才删除源文件 |
| `-cli` | 命令行问答，不打开窗口 |
| `-yes` | 不提问、不弹窗（必须提供 `-dst` 和 `-src`） |

## 使用前

仅用于授权交接或本人备份。OneDrive「仅联机」请改为始终保留在此设备。Outlook/聊天软件打开中的文件可能失败。若杀毒软件拦截 exe，可添加信任规则（仅本地拷贝/移动）。
