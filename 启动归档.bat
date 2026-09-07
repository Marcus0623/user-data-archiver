@echo off
setlocal
cd /d "%~dp0"
if not exist "%~dp0user-data-archiver.exe" (
  echo 未找到 user-data-archiver.exe，正在编译...
  go build -ldflags="-s -w" -o user-data-archiver.exe .
  if errorlevel 1 (
    echo 编译失败。请先安装 Go，或把已经编好的 exe 放到本目录。
    pause
    exit /b 1
  )
)
"%~dp0user-data-archiver.exe" %*
echo.
pause
