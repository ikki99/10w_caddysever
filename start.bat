@echo off
chcp 65001 >nul
title Caddy 管理器
echo.
echo ============================================================
echo                   Caddy 管理器 启动器
echo ============================================================
echo.
echo 正在启动服务...
echo.

caddy-manager.exe

if %errorlevel% neq 0 (
    echo.
    echo ❌ 程序异常退出
    pause
) else (
    echo.
    echo ✅ 程序已退出
)
