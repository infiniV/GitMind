@echo off
REM Build script for GitMind (Windows)

echo Building GitMind...

REM Create bin directory if it doesn't exist
if not exist bin mkdir bin

REM Build the binary
go build -o bin\gm.exe ./cmd/gm

if %ERRORLEVEL% EQU 0 (
    echo.
    echo ✓ Build successful!
    echo.
    echo Binary: bin\gm.exe
    echo Size:
    dir bin\gm.exe | find "gm.exe"
    echo.
    echo Run 'bin\gm.exe --version' to test
    echo Run 'bin\gm.exe config' to configure
) else (
    echo.
    echo ✗ Build failed!
    exit /b %ERRORLEVEL%
)
