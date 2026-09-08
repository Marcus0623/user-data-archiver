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
| **复制文件** | 拷到目标盘，源文件保留 |
| **剪切文件（移动）** | 先保存并刷盘，**目标可读且与源逐字节一致后才删除源文件**。仅大小/时间相同不够；内容损坏会先重拷再核对。不会删除仅云端占位的文件。不可撤销。不会删除 `C:\`、`C:\Users` 或用户主目录 |
| **取消** | 退出 |

剪切在真正改文件前还会再确认一次（默认否）。预览/扫描不会拷贝或删除。

## 图形界面

双击 `user-data-archiver.exe` 或 `start-archive.bat`。程序按 Windows 图形界面编译（`-H windowsgui`），资源管理器里双击时不应再闪过命令行窗口。选完复制/剪切后：

- **姓名**：报告和默认目标文件夹名
- **源路径**：可同时填多块磁盘或多个文件夹，逗号或分号分隔，例如 `C:`、`C:,D:`、`C:\Users;E:\data`。点 **浏览** 会**追加**到列表，不会覆盖已有路径；重复项会跳过。要去掉某一项，在框里删除即可
- **保存到**：最终归档目录，可浏览（默认 `D:\offboarding-archive\<姓名>`）。可以放在源盘上的文件夹里；不能是 `D:\` 这种盘符根目录
- 目标盘类型、卷标、剩余空间
- **读写测试**：创建、写入、读回并删除 `_write-test.tmp`；光驱、未解锁、写保护、无权限会失败
- 模式：个人文件 / 个人+应用数据 / 所有非系统文件
- 可选：包含 Program Files；包含 `node_modules` 等
- **排除**：按目录名排除（与 `-exclude` 相同）
- **仅预览**：勾选后 **开始复制/剪切** 不可用，请点 **预览**
- **预览**：只扫描，不拷、不剪
- **开始复制** 或 **开始剪切**：一定是真正传输，不会静默变成预览
- **停止**
- 日志：体积预览、进度、错误

管理员权限用于**读取** C 盘其他用户，**不能代替**目标盘写权限。`-cli` 或 `-yes` 不弹窗。只加 `-cut`、不加 `-cli` 时，会直接进入剪切窗口。界面语言跟随 Windows 显示语言，也可用 `-lang` 或窗口里的语言列表：`en`、`zh-CN`、`zh-TW`、`ja`、`fr`、`ru`、`vi`。

## 样例

同事小王离职。电脑系统和个人资料在 `C:`，项目代码在 `D:\Projects`。你有一块 U 盘 `E:`（或另一块硬盘）。目标是**复制**（源盘文件保留）个人文件、已落在磁盘上的聊天数据，以及项目目录。

1. 双击 `user-data-archiver.exe`。
2. 选 **复制文件**（除非确定要清空源盘，否则不要选剪切）。
3. **姓名：** `xiaowang`
4. **源路径：** `C:,D:\Projects`  
   可直接输入，或先 **浏览** 选 `C:\`，再 **浏览** 选 `D:\Projects`（第二次会追加，不会把 `C:` 清掉）。
5. **保存到：** `E:\offboarding-archive\xiaowang`
6. 点 **读写测试**。失败则换盘或修权限。
7. 模式保持推荐项：**个人文件 + 应用数据**。
8. 先点 **预览**，看日志里按目录的体积和剩余空间。这一步不会拷贝。
9. 再点 **开始复制**。结束后打开目标文件夹核对。

归档后的路径（盘符会变成文件夹，因为 Windows 不能创建 `E:\C:\...`）：

| 小王电脑上 | 归档里 |
| --- | --- |
| `C:\Users\xiaowang\Desktop\交接.docx` | `E:\offboarding-archive\xiaowang\C\Users\xiaowang\Desktop\交接.docx` |
| `C:\Users\xiaowang\Documents\WeChat Files\...` | `E:\offboarding-archive\xiaowang\C\Users\xiaowang\Documents\WeChat Files\...` |
| `C:\Downloads\合同.pdf` | `E:\offboarding-archive\xiaowang\C\Downloads\合同.pdf` |
| `D:\Projects\api\readme.md` | `E:\offboarding-archive\xiaowang\D\Projects\api\readme.md` |

同一目录还会有 `_archive-report.txt`；失败文件见 `_failed-files.csv`。

命令行做同一件事：

```bat
user-data-archiver.exe -lang zh-CN -name xiaowang -src C:,D:\Projects -dst E:\offboarding-archive\xiaowang -mode appdata -yes
```

只预览、不拷贝：

```bat
user-data-archiver.exe -lang zh-CN -name xiaowang -src C:,D:\Projects -dst E:\offboarding-archive\xiaowang -dry-run -yes
```

### 没有 U 盘：源包含目标盘（C: 和 D: 都拷到 D:）

没有 E: 盘时，仍可以把 **C: 和 D:** 拷到 **D: 上的专用文件夹**。不要填 `D:\` 根目录。

1. **源路径：** `C:,D:`（或 `C:,D:\Projects`）
2. **保存到：** `D:\offboarding-archive\xiaowang` — 不要填 `D:\`
3. **读写测试**，再 **预览**。日志会警告「目标位于源路径内部」，这是正常的。
4. 确认 D 盘剩余空间够再放**一整份**要归档的文件（C 盘那份 + D 盘那份），然后 **开始复制**。

| 小王电脑上 | 归档里 |
| --- | --- |
| `C:\Downloads\合同.pdf` | `D:\offboarding-archive\xiaowang\C\Downloads\合同.pdf` |
| `D:\Projects\api\readme.md` | `D:\offboarding-archive\xiaowang\D\Projects\api\readme.md` |

扫描 D 盘时会**跳过归档目录本身**，不会把已经拷进去的文件再拷一层。再跑一次也不会越拷越深。

```bat
user-data-archiver.exe -lang zh-CN -name xiaowang -src C:,D: -dst D:\offboarding-archive\xiaowang -mode appdata -yes
```

**保存到** 不能是盘符根目录（`C:\`、`D:\`），读写测试会拒绝。有 U 盘或第三块盘时优先用；拷到同一块 D 盘可以，但 D 盘要多占一份空间。

## 路径

不能使用 `D:\C:\Downloads`。盘符变成文件夹，例如 `C:\Downloads\a.zip` → `D:\offboarding-archive\alice\C\Downloads\a.zip`。有条件时请把目标放在**另一块盘或 U 盘**。也可以扫 `C:,D:` 并存到 `D:\offboarding-archive\姓名`：程序会跳过该归档文件夹，避免拷进自己；预览会提示目标在源路径内部；D 盘需要能再放下要拷的全部文件。不能把 `D:\` 整盘当作保存位置。相同目标的联接/符号链接只存一份。支持超长路径。

## 跳过与拷贝

**默认跳过：** `Windows`、回收站、系统卷、Recovery/Boot/EFI、pagefile/hiberfil、`Program Files`/`ProgramData`、OEM 目录、`Default` 等系统用户、AppData 缓存、`node_modules` 等可再生成目录、`NTUSER.DAT`。会收 `Windows.old\Users`，不收 `Windows.old\Windows`。

**默认拷贝：** 用户文档和 `C:\Downloads` 等自定义目录；**appdata** 模式还包含 `AppData\Roaming`（聊天/浏览器，不含缓存）。

## 模式

1. **personal**：仅文档，不含 AppData
2. **appdata**（推荐）：文档 + 聊天/浏览器，排除缓存
3. **all**：非系统自定义目录；除非勾选，否则仍跳过已安装程序

`-exclude games,Steam` 按目录名额外排除。窗口里的 **Exclude** 填同样的名单。

## 聊天记录

默认的 **appdata** 模式会收**已经落在这台电脑磁盘上**的聊天数据，例如：

- `Documents\WeChat Files`、`Documents\Tencent Files`、`Documents\WXWork`
- `AppData\Roaming`（微信/QQ、钉钉、飞书、Telegram、Slack、Teams、Discord 等）
- 其它盘上自建的聊天目录（把该盘加进 **Source**）

不会收、或不完整：

- 只在云端或手机里的消息
- **personal** 模式跳过全部 AppData（文档里的微信文件目录仍会收）
- 微信/QQ/Outlook/Teams 正在打开导致文件被锁——先退出再跑
- OneDrive「仅联机」占位文件

这是**文件备份**，不是可读的聊天导出。换机后一般要把同一款客户端指到这些目录才能打开。

## 扫描、续传、报告

按顶层目录显示文件数和体积，空间不足会警告。中断后对**同一目标**再跑，复制模式会跳过大小和时间戳相同的文件。**剪切**仍会再读目标，内容一致才删源。报告 `_archive-report.txt`，失败 `_failed-files.csv`。

## 编译与可执行文件

本目录应只保留 **一份** 程序：`user-data-archiver.exe`。项目每次更新后都要换掉旧 exe，不要留下多份。

| 操作 | 结果 |
| --- | --- |
| `build.bat` | 需要已安装 Go。先把 `app.manifest`（通用控件 6）编进资源，再按图形界面程序编译（`-H windowsgui`），**删掉多余的 `.exe`**，只留下 `user-data-archiver.exe`。若程序正在运行导致无法替换，请先退出再编。 |
| `start-archive.bat` | 若已安装 Go，且源码比 exe 新（或目录里还有其它 `.exe`），会先重新编译再打开界面。 |

编好后可以把 `user-data-archiver.exe`（以及可选的 `start-archive.bat`）拷到没有安装 Go 的电脑上使用。

## 命令行

```bat
user-data-archiver.exe -lang zh-CN -name xiaowang -src C:,D:\Projects -dst E:\offboarding-archive\xiaowang -mode appdata -yes
user-data-archiver.exe -lang zh-CN -name xiaowang -src C:,D: -dst D:\offboarding-archive\xiaowang -mode appdata -yes
user-data-archiver.exe -src C: -dst E:\offboarding-archive\xiaowang -dry-run -yes
user-data-archiver.exe -src C: -dst E:\offboarding-archive\xiaowang -cut -yes
user-data-archiver.exe -cli
user-data-archiver.exe -lang zh-CN
```

| 参数 | 含义 |
| --- | --- |
| `-name` | 姓名（报告和默认文件夹名） |
| `-src` | 源路径，逗号或分号分隔（`C:`、`C:,D:` 或 `C:\Users;E:\data`） |
| `-dst` | 最终归档目录（可放在源盘上的文件夹；不能是 `D:\` 这种盘符根目录） |
| `-mode` | `personal` \| `appdata` \| `all`（默认 `appdata`） |
| `-include-program-files` | 同时包含 Program Files / ProgramData |
| `-include-regeneratable` | 包含 `node_modules`、`__pycache__` 等 |
| `-exclude` | 按目录名额外排除（任意深度） |
| `-dry-run` | 只扫描，不拷贝、不剪切 |
| `-cut` | 移动：目标刷盘且内容一致后才删除源文件 |
| `-lang` | 界面语言：`en`、`zh-CN`、`zh-TW`、`ja`、`fr`、`ru`、`vi`（默认跟随 Windows 显示语言） |
| `-cli` | 命令行问答，不打开窗口 |
| `-yes` | 不提问、不弹窗（必须提供 `-dst` 和 `-src`） |

## 使用前

仅用于授权交接或本人备份。OneDrive「仅联机」请改为始终保留在此设备。Outlook/聊天软件打开中的文件可能失败。若杀毒软件拦截 exe，可添加信任规则（仅本地拷贝/移动）。
