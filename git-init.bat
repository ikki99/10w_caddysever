@echo off
chcp 65001 >nul
cls
echo.
echo ====================================================================
echo            Caddy Manager v1.0.0 - Git 发布准备
echo ====================================================================
echo.
echo 制作者: 10w
echo 邮箱: wngx99@gmail.com
echo GitHub: https://github.com/10w-server/caddy-manager
echo.
echo ====================================================================
echo.

REM 检查是否已经初始化 Git
if not exist ".git" (
    echo 📦 初始化 Git 仓库...
    git init
    echo ✓ Git 仓库已初始化
    echo.
)

REM 添加文件
echo 📝 添加文件到 Git...
echo.

REM 核心文件
git add .gitignore
git add README.md
git add CHANGELOG.md
git add LICENSE
git add VERSION
git add go.mod go.sum
git add main.go
git add Caddyfile

REM 源代码
git add internal/
git add web/

REM 工具和文档
git add build.bat
git add 开始.bat
git add 启动.bat
git add 检测SSL问题.bat
git add 检测混合内容.ps1
git add IPv4-IPv6兼容性问题.md
git add 混合内容检测修复指南.md
git add IPv6兼容性更新说明.txt
git add IPv6快速参考.txt
git add SSL问题快速解决.txt
git add 修复完成-README.md

echo ✓ 文件已添加
echo.

REM 提交
echo 💾 提交更改...
git commit -m "Release v1.0.0" -m "" -m "🎉 首个正式版本发布" -m "" -m "主要功能:" -m "- 完整的项目管理系统" -m "- 自动 SSL 证书申请" -m "- IPv4/IPv6 代理连接选择" -m "- 混合内容检测工具" -m "- 系统诊断和自动修复" -m "" -m "修复:" -m "- Session 超时问题（延长至 7 天）" -m "- 黑框闪烁问题（双版本编译）" -m "- 诊断按钮无反应" -m "- Caddy 状态显示" -m "- 502 Bad Gateway（IPv4/IPv6 兼容性）" -m "" -m "制作者: 10w" -m "邮箱: wngx99@gmail.com"

echo.
echo ✓ 更改已提交
echo.

REM 显示状态
echo 📊 Git 状态:
git status
echo.

echo ====================================================================
echo.
echo 下一步:
echo.
echo 1. 添加远程仓库:
echo    git remote add origin https://github.com/10w-server/caddy-manager.git
echo.
echo 2. 推送到 GitHub:
echo    git push -u origin main
echo.
echo 3. 在 GitHub 上创建 Release:
echo    - Tag: v1.0.0
echo    - Title: Caddy Manager v1.0.0
echo    - 上传编译文件: caddy-manager.exe 和 caddy-manager-console.exe
echo.
echo ====================================================================
echo.
pause
