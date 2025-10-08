@echo off
color 0A
cls
echo.
echo     ========================================================
echo                   Caddy Manager - 修复版本
echo     ========================================================
echo.
echo     版本：v0.0.11-fix
echo     修复日期：2025-01-09
echo.
echo     ========================================================
echo                      修复内容概览
echo     ========================================================
echo.
echo     ✓ Session 超时问题（7天有效期，自动续期）
echo     ✓ 黑框闪烁问题（生成 Console 和 GUI 两个版本）
echo     ✓ 诊断按钮问题（完整错误处理）
echo     ✓ Caddy 状态显示问题（改进检测逻辑）
echo     ✓ SSL 混合内容检测工具（专门解决黄色叹号）
echo.
echo     ========================================================
echo                      快速开始
echo     ========================================================
echo.
echo     请选择：
echo.
echo       [1] 启动 Caddy Manager（选择调试/生产模式）
echo       [2] 检测 SSL 混合内容问题（解决黄色叹号）⭐
echo       [3] 查看完整修复文档
echo       [4] 退出
echo.
echo     ========================================================
echo.

set /p choice=请输入选择 (1-4): 

if "%choice%"=="1" (
    start 启动.bat
) else if "%choice%"=="2" (
    start 检测SSL问题.bat
) else if "%choice%"=="3" (
    start 最终修复总结.txt
    start 混合内容检测修复指南.md
) else if "%choice%"=="4" (
    exit
) else (
    echo.
    echo 无效选择
    pause
    cls
    goto :eof
)
