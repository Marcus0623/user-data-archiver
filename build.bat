@echo off
cd /d "%~dp0"

go run github.com/akavel/rsrc@v0.10.2 -manifest app.manifest -arch amd64 -o rsrc_windows_amd64.syso
if errorlevel 1 (
  echo Failed to embed app.manifest. The GUI needs Common Controls 6.
  exit /b 1
)

go build -ldflags="-s -w -H windowsgui" -o user-data-archiver.exe.new .
if errorlevel 1 (
  echo Build failed
  del /q user-data-archiver.exe.new 2>nul
  exit /b 1
)

del /q user-data-archiver.exe 2>nul
if exist user-data-archiver.exe (
  echo Cannot replace user-data-archiver.exe. Close it and run build.bat again.
  del /q user-data-archiver.exe.new 2>nul
  exit /b 1
)

for %%F in (*.exe) do (
  if /I not "%%~nxF"=="user-data-archiver.exe.new" del /q "%%F"
)

ren user-data-archiver.exe.new user-data-archiver.exe
if errorlevel 1 (
  echo Could not keep user-data-archiver.exe.new as the only executable.
  exit /b 1
)
echo Built %~dp0user-data-archiver.exe
