@echo off
chcp 65001 >nul
echo.
echo ========================================
echo   Caddy Manager - Building Console Version
echo ========================================
echo.
echo Building console version (with visible window)...
go build -ldflags="-s -w" -o caddy-manager-console.exe
if errorlevel 1 goto :error
echo.
echo [SUCCESS] Console version build completed!

echo.
echo Building GUI version (without console window)...
go build -ldflags="-s -w -H=windowsgui" -o caddy-manager.exe
if errorlevel 1 goto :error
echo.
echo [SUCCESS] GUI version build completed!
echo.
echo Files created:
echo   - caddy-manager-console.exe (with console)
echo   - caddy-manager.exe (GUI mode, no console)
echo.
echo Note: Use caddy-manager-console.exe for debugging
echo       Use caddy-manager.exe for production
echo.
goto :end

:error
echo.
echo [ERROR] Build failed
echo.
exit /b 1

:end
pause
