@echo off
setlocal
cd /d "%~dp0"

where go >nul 2>nul
if errorlevel 1 (
    echo ERROR: Go is not installed or is not available in PATH.
    echo Install Go 1.20.x only if you want to build the executable yourself.
    exit /b 1
)

for /f "tokens=3" %%V in ('go version') do set GOVERSION=%%V
echo Using %GOVERSION%

echo %GOVERSION% | findstr /b /c:"go1.20" >nul
if errorlevel 1 (
    echo ERROR: This compatibility build must use Go 1.20.x.
    echo Go 1.21+ no longer officially supports Windows 7 / Windows Server 2008 R2.
    exit /b 1
)

if not exist dist mkdir dist
set GOOS=windows
set GOARCH=amd64
set CGO_ENABLED=0

go build -trimpath -ldflags="-s -w -H=windowsgui" -o dist\OracleReportGroupFinder.exe .
if errorlevel 1 exit /b 1

echo.
echo Build complete: dist\OracleReportGroupFinder.exe
endlocal
