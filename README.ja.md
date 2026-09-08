# User Data Archiver

<p>
  <a href="README.md"><kbd>English</kbd></a>
  &nbsp;<a href="README.zh-CN.md"><kbd>简体中文</kbd></a>
  &nbsp;<a href="README.zh-TW.md"><kbd>繁體中文</kbd></a>
  &nbsp;<kbd><b>日本語</b></kbd>
  &nbsp;<a href="README.fr.md"><kbd>Français</kbd></a>
  &nbsp;<a href="README.ru.md"><kbd>Русский</kbd></a>
  &nbsp;<a href="README.vi.md"><kbd>Tiếng Việt</kbd></a>
</p>

Windows 用ツール。**OS 以外のユーザーファイル**を別ディスクまたは USB に保存し、**元のフォルダー構造を保ちます**。退職者の個人データの収集などに使います。

独立した Go プログラムで、他プロジェクトは不要です。更新のたびに実行ファイルは `user-data-archiver.exe` **1 つだけ**残してください。

## 機能

`C:` などを走査し、**再インストールしても戻らないデータ**（デスクトップ、文書、ダウンロード、プロジェクト、チャット、推奨モードではブラウザ設定）を元のフォルダー構造のまま保存します。OS、ごみ箱、ページファイル、インストール済みプログラム、OEM、キャッシュは既定で除外します。

## コピーまたはカット

最初のウィンドウ：

| 選択 | 動作 |
| --- | --- |
| **Copy files** | 保存先へコピーし、元ファイルは残す |
| **Cut files (move)** | 書き込み・フラッシュ後、**保存先が読め、内容が一致したファイルだけ**元を削除。サイズ／時刻が同じだけでは不十分。オンラインのみのプレースホルダーは削除しない。取り消し不可。`C:\`、`C:\Users`、プロファイル直下は削除しない |
| **Cancel** | 終了 |

カットは実行前に再確認します（既定はいいえ）。プレビューはコピーも削除もしません。

## GUI

`user-data-archiver.exe` または `start-archive.bat` をダブルクリック。氏名、ソース、**保存先**（参照。ソースドライブ上のフォルダーは可、`D:\` 直下は不可）、**除外**、ドライブ種別／ラベル／空き容量、**読み書きテスト**、モード、Program Files / `node_modules`、プレビュー、開始／停止、ログ。ソースは複数指定できます（カンマまたはセミコロン、例 `C:,D:`）。**参照**は既存の入力を消さず追加します。**プレビューのみ**をオンにすると **開始** は無効になり、**プレビュー** を使います。**開始** は必ず実コピー／カットです。管理者権限は C: の他ユーザーを**読む**ためで、保存先の書き込み権限の代わりにはなりません。`-cli` / `-yes` でウィンドウなし。`-cut` 付きで起動するとカット画面になります。言語は `-lang`：`en`, `zh-CN`, `zh-TW`, `ja`, `fr`, `ru`, `vi`。

## 例

退職者 Alice。OS と個人データは `C:`、プロジェクトは `D:\Projects`。USB は `E:`。**コピー**（元ファイルは残す）します。

1. `user-data-archiver.exe` を開き **コピー** を選ぶ。
2. 氏名 `alice`。ソース `C:,D:\Projects`（または `C:\` を参照したあと `D:\Projects` を参照して追加）。
3. 保存先 `E:\offboarding-archive\alice`。読み書きテスト → **プレビュー** → **開始（コピー）**。

| 元のパス | アーカイブ内 |
| --- | --- |
| `C:\Users\alice\Desktop\handoff.docx` | `E:\offboarding-archive\alice\C\Users\alice\Desktop\handoff.docx` |
| `C:\Downloads\contract.pdf` | `E:\offboarding-archive\alice\C\Downloads\contract.pdf` |
| `D:\Projects\api\readme.md` | `E:\offboarding-archive\alice\D\Projects\api\readme.md` |

同じフォルダーに `_archive-report.txt`。コマンド例：

```bat
user-data-archiver.exe -lang ja -name alice -src C:,D:\Projects -dst E:\offboarding-archive\alice -mode appdata -yes
```

USB が無い例：

```bat
user-data-archiver.exe -lang ja -name alice -src C:,D: -dst D:\offboarding-archive\alice -mode appdata -yes
```

保存先は**別ディスクまたは USB**が望ましいです。USB が無い場合はソース `C:,D:`、保存先 `D:\offboarding-archive\alice`（`D:\` 直下は不可）。ログに「保存先がソースの中」と出るのは正常です。D: にはコピー分の空きが必要です。プログラムは保存先フォルダー自身をスキップし、入れ子コピーしません。

## チャット履歴

**appdata**（推奨）では、この PC に既にあるチャットデータを保存します。例: `Documents\WeChat Files`、`AppData\Roaming`（WeChat/QQ、DingTalk、Lark、Telegram、Teams など）。クラウドやスマホのみのメッセージ、**personal** の AppData、起動中でロックされたファイル、OneDrive のオンラインのみファイルは欠けることがあります。ファイルのバックアップであり、読めるチャット書き出しではありません。

## パス・除外・再開

`D:\C:\Downloads` は不可。ドライブレターがフォルダーになります。保存先は**別ディスクまたは USB**が望ましいです。`C:,D:` を `D:\offboarding-archive\氏名` へ保存することもできます（保存先フォルダーはスキップ、`D:\` ルートは不可）。既定で `Windows`、`Program Files`、キャッシュ、`node_modules` などを除外。**appdata** は `AppData\Roaming` を含む（チャット／ブラウザ、キャッシュ除く）。`Windows.old\Users` は対象。コピーの再開は同一保存先でサイズとタイムスタンプが同じファイルをスキップ。カットは内容照合後にだけ削除。報告 `_archive-report.txt`、失敗 `_failed-files.csv`。

## ビルドと実行ファイル

このフォルダーには `user-data-archiver.exe` を **1 つだけ**置きます。`build.bat` は余った `.exe` を削除して最新版だけ残します。`start-archive.bat` はソースが新しいと再ビルドします。

## モードと CLI

**personal** / **appdata**（推奨）/ **all**。移動は `-cut`、プレビューは `-dry-run`。

```bat
user-data-archiver.exe -name alice -src C:,D:\Projects -dst E:\offboarding-archive\alice -mode appdata -yes
user-data-archiver.exe -name alice -src C:,D: -dst D:\offboarding-archive\alice -mode appdata -yes
user-data-archiver.exe -src C: -dst E:\offboarding-archive\alice -dry-run -yes
user-data-archiver.exe -src C: -dst E:\offboarding-archive\alice -cut -yes
user-data-archiver.exe -cli
user-data-archiver.exe -lang ja
```

| フラグ | 意味 |
| --- | --- |
| `-name` | 氏名（レポートと既定フォルダー名） |
| `-src` | ソース（カンマ区切り） |
| `-dst` | 保存先フォルダー（ソースドライブ上は可、`D:\` ルートは不可） |
| `-mode` | `personal` \| `appdata` \| `all`（既定 `appdata`） |
| `-include-program-files` | Program Files / ProgramData も含める |
| `-include-regeneratable` | `node_modules` なども含める |
| `-exclude` | 追加で除外するフォルダー名 |
| `-dry-run` | スキャンのみ |
| `-cut` | 移動：保存先を検証してから元を削除 |
| `-lang` | UI 言語: `en`, `zh-CN`, `zh-TW`, `ja`, `fr`, `ru`, `vi` |
| `-cli` | ウィンドウなしの対話 CLI |
| `-yes` | 確認なし（`-dst` と `-src` 必須） |

許可された引き継ぎまたは自分のバックアップ専用。
