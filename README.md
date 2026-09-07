# 离职资料归档工具

把电脑上**除 Windows 系统初始文件以外**的资料，按原来的文件夹结构拷到另一块盘/U 盘，方便同事离职后交接个人文档。

这是独立的 Windows 本地工具仓库，不依赖其他项目。

## 路径怎么放

Windows **不允许**目录名带冒号，所以不能建成 `D:\C:\Downloads`。本工具用盘符文件夹代替：

| 原路径 | 归档后 |
| --- | --- |
| `C:\Downloads\资料.zip` | `D:\离职归档\张三\C\Downloads\资料.zip` |
| `C:\Users\张三\Desktop\a.docx` | `D:\离职归档\张三\C\Users\张三\Desktop\a.docx` |

建议目标选 **另一块硬盘或 U 盘**，不要写回正在扫描的 C 盘同一位置。

## 会跳过什么（“重装系统后仍在的东西”）

默认跳过：

- `Windows`、回收站、`System Volume Information`、页面文件/休眠文件
- `Program Files`、`ProgramData`（程序可重装，默认不当作个人资料）
- 品牌机自带的 `Intel` / `Dell` 等驱动目录
- 系统默认用户（`Default` 等）
- 软件缓存（`AppData` 下的 Cache、Temp 等）
- `node_modules`、`__pycache__` 等可再生成目录

默认**会拷贝**：桌面/文档/下载/图片、项目目录、微信文件、浏览器与聊天软件配置（`AppData\Roaming` 等，不含缓存）。

升级残留的 `Windows.old\Users` 会归档，其中的 `Windows.old\Windows` 不会。

## 三种模式

1. **个人资料**：用户文档，不含 AppData  
2. **个人资料 + 软件配置（推荐）**：含微信/浏览器等，排除缓存  
3. **非系统全量**：C 盘上除 Windows/系统卷/已装程序外的自定义目录都收

## 怎么运行

### 方式 A：双击

1. 若尚未编译，先运行 `build.bat`（需已安装 Go），得到 `user-data-archiver.exe`
2. 双击 `启动归档.bat`（或直接运行 exe）
3. 按提示输入员工姓名、目标盘、源路径（一般是 `C:`）
4. 先看预览体积，确认后再复制

需要读取其他用户目录时，请右键 **以管理员身份运行**。

### 方式 B：命令行

```bat
user-data-archiver.exe -name 张三 -src C: -dst D:\离职归档\张三 -mode appdata -yes
```

只预览不复制：

```bat
user-data-archiver.exe -src C: -dst D:\离职归档\张三 -dry-run -yes
```

常用参数：

- `-mode personal|appdata|all`
- `-include-program-files` 连安装目录一起拷（很占空间）
- `-include-regeneratable` 包含 `node_modules` 等
- `-exclude games,Steam` 按目录名额外排除
- `-yes` 扫描后不询问

中断后用**同一目标目录**再跑一次，已拷过且大小/时间相同的文件会跳过。

完成后会在目标目录生成 `_归档报告.txt`，失败项在 `_失败清单.csv`。

## 使用前请注意

- 仅用于公司授权的离职交接、本人电脑备份，不要对未授权设备做全盘拷贝
- OneDrive「仅联机」文件可能只拷到占位符：请先改为「始终保留在此设备上」
- 正在打开的 Outlook/微信数据库可能拷失败，可关掉对应软件后重跑续传
- 杀毒软件若拦截本程序，请添加信任（本工具只做本地拷贝，无后台驻留）
