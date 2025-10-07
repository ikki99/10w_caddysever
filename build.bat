@echo off
echo 正在编译 Caddy 管理器...
go build -ldflags="-s -w" -o caddy-manager.exe
if %errorlevel% == 0 (
    echo.
    echo ✅ 编译成功！
    echo 运行: caddy-manager.exe
) else (
    echo.
    echo ❌ 编译失败
)
pause
