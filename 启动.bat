@echo off
chcp 65001 >nul
cls
echo.
echo ====================================================================
echo                  Caddy Manager 快速启动
echo ====================================================================
echo.
echo 请选择启动模式:
echo.
echo   [1] 调试模式 (Console) - 显示日志输出，方便排查问题
echo   [2] 生产模式 (GUI)     - 无窗口运行，适合日常使用
echo.
echo ====================================================================
echo.

set /p choice=请输入选择 (1 或 2): 

if "%choice%"=="1" (
    echo.
    echo 正在以调试模式启动...
    echo 您将看到详细的日志输出
    echo.
    caddy-manager-console.exe
) else if "%choice%"=="2" (
    echo.
    echo 正在以生产模式启动...
    echo 程序将在后台运行，无窗口
    echo.
    echo 访问地址: http://localhost:8989
    echo.
    start caddy-manager.exe
    echo.
    echo 程序已启动！
    echo.
    timeout /t 3
) else (
    echo.
    echo 无效选择，请重新运行脚本
    pause
)
