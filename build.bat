@echo off
chcp 65001 >nul
echo.
echo ========================================
echo   Caddy Manager - Building
echo ========================================
echo.
echo Building...
go build -ldflags="-s -w -H=windowsgui" -o caddy-manager.exe
if %errorlevel% == 0 (
    echo.
    echo [SUCCESS] Build completed successfully!
    echo.
    echo Run: caddy-manager.exe
    echo.
) else (
    echo.
    echo [ERROR] Build failed
    echo.
)
pause
