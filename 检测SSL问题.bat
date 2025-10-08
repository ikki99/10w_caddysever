@echo off
chcp 65001 >nul
cls
echo.
echo ====================================================================
echo              SSL 黄色叹号问题 - 混合内容检测工具
echo ====================================================================
echo.
echo 您的网站：https://c808.333.606.f89f.top/
echo SSL 证书：Let's Encrypt (正常有效)
echo 问题：浏览器显示黄色叹号/不安全警告
echo.
echo 可能原因：页面中包含 HTTP 资源（混合内容）
echo.
echo ====================================================================
echo.
echo 请选择操作：
echo.
echo   [1] 🔍 检测项目中的混合内容（推荐先做这个）
echo   [2] 📖 查看详细解决方案文档
echo   [3] 🌐 在浏览器中手动检测（最准确）
echo   [4] 🔧 自动修复混合内容（谨慎使用）
echo   [5] ❌ 退出
echo.
echo ====================================================================
echo.

set /p choice=请输入选择 (1-5): 

if "%choice%"=="1" (
    echo.
    echo 请输入您的项目根目录路径：
    echo 例如：C:\www\myproject
    echo 或直接按回车使用当前目录
    echo.
    set /p projectPath=项目路径: 
    
    if "%projectPath%"=="" (
        set projectPath=.
    )
    
    echo.
    echo 正在扫描项目...
    echo.
    powershell -ExecutionPolicy Bypass -File "检测混合内容.ps1" -ProjectPath "%projectPath%" -ShowDetails
    pause
    
) else if "%choice%"=="2" (
    echo.
    echo 正在打开文档...
    start 混合内容检测修复指南.md
    echo.
    echo 文档已打开，请查看详细的解决方案
    pause
    
) else if "%choice%"=="3" (
    echo.
    echo ====================================================================
    echo                    浏览器手动检测步骤
    echo ====================================================================
    echo.
    echo 1. 打开您的网站：https://c808.333.606.f89f.top/
    echo.
    echo 2. 按 F12 键打开开发者工具
    echo.
    echo 3. 切换到 "Console" (控制台) 标签
    echo.
    echo 4. 刷新页面 (F5)
    echo.
    echo 5. 查找黄色或红色的警告信息：
    echo    - "Mixed Content" (混合内容)
    echo    - "blocked loading mixed active content" 
    echo    - HTTP 资源的 URL 会被列出
    echo.
    echo 6. 记下所有 HTTP 开头的资源地址
    echo.
    echo 7. 在项目中找到这些资源并改为 HTTPS 或相对路径
    echo.
    echo ====================================================================
    echo.
    echo 按任意键打开网站...
    pause >nul
    start https://c808.333.606.f89f.top/
    echo.
    echo 网站已在浏览器中打开，请按照上述步骤检查
    echo.
    pause
    
) else if "%choice%"=="4" (
    echo.
    echo ⚠️  警告：自动修复会修改您的项目文件！
    echo.
    echo 请输入您的项目根目录路径：
    set /p projectPath=项目路径: 
    
    if "%projectPath%"=="" (
        echo 未输入路径，已取消
        pause
        goto :eof
    )
    
    echo.
    echo 即将自动修复项目中的混合内容
    echo 所有 http:// 将被替换为 https://
    echo.
    set /p confirm=确定继续吗？(y/N): 
    
    if /i "%confirm%"=="y" (
        echo.
        echo 正在修复...
        powershell -ExecutionPolicy Bypass -File "检测混合内容.ps1" -ProjectPath "%projectPath%" -AutoFix
        echo.
        echo 修复完成！
        echo.
        echo 重要提示：
        echo 1. 请测试网站确保一切正常
        echo 2. 如有问题，可用 .bak 备份文件恢复
        echo 3. 清除浏览器缓存并硬刷新 (Ctrl+F5)
        echo.
    ) else (
        echo 已取消自动修复
    )
    pause
    
) else if "%choice%"=="5" (
    echo.
    echo 再见！
    exit /b 0
    
) else (
    echo.
    echo 无效选择，请重新运行
    pause
)

echo.
goto :eof
