# 文件清理脚本
# 保留必要文件，删除开发过程中的临时文档

@echo off
chcp 65001 >nul
cls
echo.
echo ====================================================================
echo                   清理无用文件
echo ====================================================================
echo.
echo 此脚本将删除开发过程中的临时文档和无用文件
echo.
echo 保留文件：
echo   - README.md, CHANGELOG.md, LICENSE
echo   - 主程序文件（.exe）
echo   - 工具脚本（检测SSL问题.bat, 检测混合内容.ps1等）
echo   - 核心文档（IPv4-IPv6兼容性问题.md等）
echo   - 源代码文件
echo.
echo ====================================================================
echo.

set /p confirm=确定要清理吗？(y/N): 

if /i not "%confirm%"=="y" (
    echo 已取消
    exit /b 0
)

echo.
echo 正在清理...
echo.

REM 删除旧版本文档
del /q README_*.md 2>nul
del /q CHANGELOG_*.md 2>nul

REM 删除开发文档
del /q 功能完善总结.md 2>nul
del /q 快速参考.md 2>nul
del /q 完善说明.md 2>nul
del /q 问题修复说明.md 2>nul
del /q 修复完成说明.txt 2>nul
del /q 最终修复总结.txt 2>nul
del /q CADDY_CONTROL_UPDATE.md 2>nul
del /q CADDYFILE_FIX.md 2>nul
del /q COMPLETE_TUTORIAL.md 2>nul
del /q COMPLETE_UPDATE_SUMMARY.md 2>nul
del /q DIAGNOSTICS_GUIDE.md 2>nul
del /q EDIT_PROJECT_GUIDE.md 2>nul
del /q FEATURES.md 2>nul
del /q FILE_UPLOAD_GUIDE.md 2>nul
del /q FIXES_SUMMARY.md 2>nul
del /q FIXES_v0.0.11.md 2>nul
del /q FRONTEND_IMPROVEMENTS.md 2>nul
del /q IMPROVEMENTS.md 2>nul
del /q QUICK_FIX_GUIDE.md 2>nul
del /q QUICKSTART.md 2>nul
del /q SSL_TROUBLESHOOTING.md 2>nul
del /q STATIC_FILES_TROUBLESHOOTING.md 2>nul
del /q SUMMARY_TRAY_SHUTDOWN.md 2>nul
del /q SUMMARY.md 2>nul
del /q TEST_GUIDE.md 2>nul
del /q TEST_REPORT.md 2>nul
del /q TEST_RESULTS.md 2>nul
del /q TRAY_AND_SHUTDOWN_UPDATE.md 2>nul
del /q TRAY_GUIDE.md 2>nul
del /q TROUBLESHOOTING.md 2>nul
del /q USAGE.md 2>nul
del /q USER_GUIDE.md 2>nul

REM 删除旧的诊断脚本
del /q diagnose-full.ps1 2>nul
del /q diagnose-remote-fixed.ps1 2>nul
del /q diagnose-remote-new.ps1 2>nul
del /q diagnose-remote.ps1 2>nul
del /q diagnose.ps1 2>nul
del /q fix-ssl.ps1 2>nul

REM 删除旧的启动脚本
del /q start.bat 2>nul

REM 删除备份文件
del /q *.bak 2>nul
del /q *.old 2>nul
del /q *.tmp 2>nul

REM 删除 API 备份文件
del /q internal\api\*.bak 2>nul
del /q internal\api\*.old 2>nul
del /q internal\api\*.old2 2>nul
del /q internal\api\*.tmp 2>nul

echo.
echo ====================================================================
echo.
echo ✓ 清理完成！
echo.
echo 保留的文件：
echo   √ README.md - 项目说明
echo   √ CHANGELOG.md - 更新日志
echo   √ LICENSE - 许可证
echo   √ 主程序文件
echo   √ 工具脚本（开始.bat, 启动.bat, 检测SSL问题.bat等）
echo   √ 核心文档（IPv4-IPv6兼容性问题.md等）
echo   √ 源代码文件
echo.
echo ====================================================================
echo.
pause
