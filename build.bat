@echo off
cd /d "%~dp0"
go build -ldflags="-s -w" -o user-data-archiver.exe .
if errorlevel 1 (
  echo 编译失败
  exit /b 1
)
echo 已生成 %~dp0user-data-archiver.exe
