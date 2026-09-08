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

Nhấp đúp `user-data-archiver.exe` hoặc `start-archive.bat`. Sau đó: tên, nguồn, **Save to** (Browse), loại ổ / nhãn / dung lượng trống, **Read/Write Test**, mode, tùy chọn Program Files / `node_modules`, preview, Start / Stop, nhật ký. Quyền administrator để **đọc** hồ sơ khác trên C:, không thay quyền ghi ổ đích. `-cli` hoặc `-yes` không mở cửa sổ.

## Đường dẫn, lọc, tiếp tục

Không dùng `D:\C:\Downloads`. Ký tự ổ thành thư mục. Đích nên là **ổ khác hoặc USB**. Bỏ `Windows`, `Program Files`, cache, `node_modules`. **appdata** gồm `AppData\Roaming` (chat/trình duyệt, không cache). Có `Windows.old\Users`. Chạy lại copy **cùng thư mục đích** sẽ bỏ qua file trùng size và thời gian. Cut chỉ xóa sau khi đối chiếu nội dung. Báo cáo `_archive-report.txt`, lỗi `_failed-files.csv`.

## Biên dịch và file exe

Chỉ giữ **một** `user-data-archiver.exe` trong thư mục. `build.bat` xóa các `.exe` cũ rồi ghi bản mới. `start-archive.bat` biên dịch lại nếu mã nguồn mới hơn.

## Chế độ và dòng lệnh

**personal** / **appdata** (khuyến nghị) / **all**. Di chuyển: `-cut`. Chỉ xem: `-dry-run`.

```bat
user-data-archiver.exe -name alice -src C: -dst D:\offboarding-archive\alice -mode appdata -yes
user-data-archiver.exe -src C: -dst D:\offboarding-archive\alice -dry-run -yes
user-data-archiver.exe -src C: -dst D:\offboarding-archive\alice -cut -yes
user-data-archiver.exe -cli
```

| Cờ | Ý nghĩa |
| --- | --- |
| `-name` | Tên (báo cáo và thư mục mặc định) |
| `-src` | Đường dẫn nguồn, cách nhau bằng dấu phẩy |
| `-dst` | Thư mục đích |
| `-mode` | `personal` \| `appdata` \| `all` (mặc định `appdata`) |
| `-include-program-files` | Gồm cả Program Files / ProgramData |
| `-include-regeneratable` | Gồm `node_modules`, `__pycache__`, v.v. |
| `-exclude` | Thêm tên thư mục cần bỏ |
| `-dry-run` | Chỉ quét |
| `-cut` | Di chuyển: xóa gốc sau khi đích đã được xác minh |
| `-cli` | Hỏi đáp dòng lệnh, không mở cửa sổ |
| `-yes` | Không hỏi, không GUI (cần `-dst` và `-src`) |

Chỉ dùng khi được phép hoặc sao lưu của bạn.
