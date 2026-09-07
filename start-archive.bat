@echo off
setlocal
cd /d "%~dp0"

where go >nul 2>&1
if errorlevel 1 goto :run

powershell -NoProfile -Command "if (-not (Test-Path -LiteralPath 'user-data-archiver.exe')) { exit 1 }; $t = (Get-Item -LiteralPath 'user-data-archiver.exe').LastWriteTime; if (Get-ChildItem -File -Filter '*.go' | Where-Object { $_.LastWriteTime -gt $t }) { exit 1 }; if (Get-ChildItem -File -Filter '*.exe' | Where-Object { $_.Name -ne 'user-data-archiver.exe' }) { exit 1 }; exit 0"
if errorlevel 1 (
  echo Rebuilding user-data-archiver.exe...
  call "%~dp0build.bat"
  if errorlevel 1 (
    if not exist "%~dp0user-data-archiver.exe" (
      echo Build failed. Install Go, or copy a built exe into this folder.
      pause
      exit /b 1
    )
    echo Build failed; starting the existing executable.
  )
)
goto :run

:run
if not exist "%~dp0user-data-archiver.exe" (
  echo user-data-archiver.exe not found, building...
  call "%~dp0build.bat"
  if errorlevel 1 (
    echo Build failed. Install Go, or copy a built exe into this folder.
    pause
    exit /b 1
  )
)
"%~dp0user-data-archiver.exe" %*
echo.
pause
