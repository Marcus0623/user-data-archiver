# User Data Archiver

<p>
  <a href="README.md"><kbd>English</kbd></a>
  &nbsp;<a href="README.zh-CN.md"><kbd>简体中文</kbd></a>
  &nbsp;<a href="README.zh-TW.md"><kbd>繁體中文</kbd></a>
  &nbsp;<a href="README.ja.md"><kbd>日本語</kbd></a>
  &nbsp;<a href="README.fr.md"><kbd>Français</kbd></a>
  &nbsp;<a href="README.ru.md"><kbd>Русский</kbd></a>
  &nbsp;<kbd><b>Tiếng Việt</b></kbd>
</p>

Công cụ Windows: lưu **file người dùng không thuộc hệ điều hành** sang ổ khác hoặc USB và **giữ nguyên cây thư mục**. Thường dùng để thu thập dữ liệu cá nhân khi nhân sự nghỉ việc.

Chương trình Go độc lập. Sau mỗi lần cập nhật, thư mục chỉ nên còn **một** file: `user-data-archiver.exe`.

## Chức năng

Quét `C:` (hoặc nhiều đường dẫn) và lưu những gì **sẽ không còn sau khi cài lại Windows**: Desktop, tài liệu, Downloads, dự án, chat, và ở chế độ khuyến nghị cả cấu hình trình duyệt. Mặc định bỏ OS, Thùng rác, pagefile, chương trình đã cài, OEM và cache.

## Copy hoặc Cut

Cửa sổ đầu tiên:

| Chọn | Tác dụng |
| --- | --- |
| **Copy files** | Lưu sang đích, giữ file gốc |
| **Cut files (move)** | Ghi xong, flush đĩa, **chỉ xóa file gốc khi đích đọc được và trùng từng byte**. Trùng size/thời gian chưa đủ. Không xóa file chỉ trên cloud. Không hoàn tác. Không xóa `C:\`, `C:\Users` hay thư mục gốc hồ sơ |
| **Cancel** | Thoát |

Cut hỏi lại trước khi đổi file (mặc định No). Preview không copy và không xóa.

## Giao diện

Nhấp đúp `user-data-archiver.exe` hoặc `start-archive.bat`. Sau đó: tên, nguồn, **Lưu vào** (Duyệt; thư mục trên ổ nguồn được phép, không dùng gốc `D:\`), **Loại trừ**, loại ổ / nhãn / dung lượng trống, **Kiểm tra đọc/ghi**, mode, tùy chọn Program Files / `node_modules`, preview, Start / Stop, nhật ký. Nguồn có thể nhiều đường dẫn (phẩy hoặc chấm phẩy, ví dụ `C:,D:`). **Duyệt** sẽ thêm, không xóa ô sẵn có. Khi **Chỉ xem trước** được chọn, dùng **Xem trước**. **Bắt đầu** luôn copy hoặc cut thật. Quyền administrator để **đọc** hồ sơ khác trên C:, không thay quyền ghi ổ đích. `-cli` hoặc `-yes` không mở cửa sổ. `-cut` mở thẳng cửa sổ Cut. Ngôn ngữ: `-lang` (`en`, `zh-CN`, `zh-TW`, `ja`, `fr`, `ru`, `vi`).

## Ví dụ

Alice nghỉ việc. Windows và hồ sơ trên `C:`, dự án trên `D:\Projects`. USB `E:`. Mục tiêu: **sao chép** (giữ file gốc).

1. Mở `user-data-archiver.exe`, chọn **Sao chép tệp**.
2. Tên `alice`. Nguồn `C:,D:\Projects` (hoặc Duyệt `C:\` rồi Duyệt `D:\Projects` — lần sau sẽ thêm vào).
3. Đích `E:\offboarding-archive\alice`. Kiểm tra đọc/ghi → **Xem trước** → **Bắt đầu sao chép**.

| Trên máy | Trong bản lưu |
| --- | --- |
| `C:\Users\alice\Desktop\handoff.docx` | `E:\offboarding-archive\alice\C\Users\alice\Desktop\handoff.docx` |
| `C:\Downloads\contract.pdf` | `E:\offboarding-archive\alice\C\Downloads\contract.pdf` |
| `D:\Projects\api\readme.md` | `E:\offboarding-archive\alice\D\Projects\api\readme.md` |

Cùng thư mục có `_archive-report.txt`. Dòng lệnh:

```bat
user-data-archiver.exe -lang vi -name alice -src C:,D:\Projects -dst E:\offboarding-archive\alice -mode appdata -yes
```

Không USB:

```bat
user-data-archiver.exe -lang vi -name alice -src C:,D: -dst D:\offboarding-archive\alice -mode appdata -yes
```

Đích nên là **ổ khác hoặc USB** nếu có. Không có USB: nguồn `C:,D:`, đích `D:\offboarding-archive\alice` (không dùng `D:\`). Cảnh báo “đích nằm trong nguồn” là bình thường; thư mục lưu trữ bị bỏ qua nên không tự sao vào chính nó. Ổ D: cần đủ chỗ cho **một bản sao thêm**.

## Chat

Chế độ **appdata** lưu dữ liệu chat **đã có trên máy**, ví dụ `Documents\WeChat Files` và `AppData\Roaming` (WeChat/QQ, DingTalk, Lark, Telegram, Teams…). Không (hoặc thiếu): tin chỉ trên cloud/điện thoại, AppData ở mode **personal**, file đang bị app khóa, OneDrive chỉ online. Đây là sao lưu file, không phải bản chat đọc được.

## Đường dẫn, lọc, tiếp tục

Không dùng `D:\C:\Downloads`. Ký tự ổ thành thư mục. Ưu tiên đích **ổ khác hoặc USB**. Có thể quét `C:,D:` và lưu vào `D:\offboarding-archive\<tên>` (bỏ qua thư mục đích; không dùng gốc `D:\`). Bỏ `Windows`, `Program Files`, cache, `node_modules`. **appdata** gồm `AppData\Roaming` (chat/trình duyệt, không cache). Có `Windows.old\Users`. Chạy lại copy **cùng thư mục đích** sẽ bỏ qua file trùng size và thời gian. Cut chỉ xóa sau khi đối chiếu nội dung. Báo cáo `_archive-report.txt`, lỗi `_failed-files.csv`.

## Biên dịch và file exe

Chỉ giữ **một** `user-data-archiver.exe` trong thư mục. `build.bat` xóa các `.exe` cũ rồi ghi bản mới. `start-archive.bat` biên dịch lại nếu mã nguồn mới hơn.

## Chế độ và dòng lệnh

**personal** / **appdata** (khuyến nghị) / **all**. Di chuyển: `-cut`. Chỉ xem: `-dry-run`.

```bat
user-data-archiver.exe -name alice -src C:,D:\Projects -dst E:\offboarding-archive\alice -mode appdata -yes
user-data-archiver.exe -name alice -src C:,D: -dst D:\offboarding-archive\alice -mode appdata -yes
user-data-archiver.exe -src C: -dst E:\offboarding-archive\alice -dry-run -yes
user-data-archiver.exe -src C: -dst E:\offboarding-archive\alice -cut -yes
user-data-archiver.exe -cli
user-data-archiver.exe -lang vi
```

| Cờ | Ý nghĩa |
| --- | --- |
| `-name` | Tên (báo cáo và thư mục mặc định) |
| `-src` | Đường dẫn nguồn, cách nhau bằng dấu phẩy |
| `-dst` | Thư mục đích (trên ổ nguồn được phép; không dùng gốc `D:\`) |
| `-mode` | `personal` \| `appdata` \| `all` (mặc định `appdata`) |
| `-include-program-files` | Gồm cả Program Files / ProgramData |
| `-include-regeneratable` | Gồm `node_modules`, `__pycache__`, v.v. |
| `-exclude` | Thêm tên thư mục cần bỏ |
| `-dry-run` | Chỉ quét |
| `-cut` | Di chuyển: xóa gốc sau khi đích đã được xác minh |
| `-lang` | Ngôn ngữ: `en`, `zh-CN`, `zh-TW`, `ja`, `fr`, `ru`, `vi` |
| `-cli` | Hỏi đáp dòng lệnh, không mở cửa sổ |
| `-yes` | Không hỏi, không GUI (cần `-dst` và `-src`) |

Chỉ dùng khi được phép hoặc sao lưu của bạn.
