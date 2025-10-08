# Caddy 管理器诊断工具
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Caddy 管理器诊断工具" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# 检查 Caddyfile
Write-Host "📄 检查 Caddyfile..." -ForegroundColor Yellow
$caddyfile = "data\caddy\Caddyfile"
if (Test-Path $caddyfile) {
    Write-Host "✓ Caddyfile 存在" -ForegroundColor Green
    $content = Get-Content $caddyfile -Raw
    Write-Host "文件大小: $($content.Length) 字节" -ForegroundColor Cyan
    
    if ($content.Length -eq 0) {
        Write-Host "⚠ 警告: Caddyfile 为空" -ForegroundColor Red
    } else {
        Write-Host "`nCaddyfile 内容预览:" -ForegroundColor Cyan
        Write-Host "----------------------------------------"
        Get-Content $caddyfile | Select-Object -First 20
        Write-Host "----------------------------------------"
    }
} else {
    Write-Host "✗ Caddyfile 不存在" -ForegroundColor Red
}

Write-Host ""

# 检查 Caddy 进程
Write-Host "🔍 检查 Caddy 进程..." -ForegroundColor Yellow
$caddyProcess = Get-Process caddy -ErrorAction SilentlyContinue
if ($caddyProcess) {
    Write-Host "✓ Caddy 正在运行 (PID: $($caddyProcess.Id))" -ForegroundColor Green
} else {
    Write-Host "✗ Caddy 未运行" -ForegroundColor Red
}

Write-Host ""

# 检查端口占用
Write-Host "🌐 检查端口占用..." -ForegroundColor Yellow
$ports = @(80, 443, 8989, 2019)
foreach ($port in $ports) {
    $listening = Get-NetTCPConnection -LocalPort $port -State Listen -ErrorAction SilentlyContinue
    if ($listening) {
        Write-Host "✓ 端口 $port : 正在监听" -ForegroundColor Green
    } else {
        Write-Host "○ 端口 $port : 未使用" -ForegroundColor Gray
    }
}

Write-Host ""

# 检查 Caddy 日志
Write-Host "�� 检查 Caddy 日志..." -ForegroundColor Yellow
$logFile = "data\caddy\caddy.log"
if (Test-Path $logFile) {
    Write-Host "✓ 日志文件存在" -ForegroundColor Green
    $errors = Get-Content $logFile -Tail 50 | Select-String -Pattern "error|Error|ERROR"
    if ($errors) {
        Write-Host "`n⚠ 发现错误:" -ForegroundColor Red
        $errors | Select-Object -First 5 | ForEach-Object {
            Write-Host "  - $_" -ForegroundColor Yellow
        }
    } else {
        Write-Host "✓ 最近 50 行日志中没有错误" -ForegroundColor Green
    }
} else {
    Write-Host "✗ 日志文件不存在" -ForegroundColor Red
}

Write-Host ""

# 检查数据库
Write-Host "💾 检查数据库..." -ForegroundColor Yellow
$dbFile = "data\caddy-manager.db"
if (Test-Path $dbFile) {
    $dbSize = (Get-Item $dbFile).Length
    Write-Host "✓ 数据库存在 (大小: $([math]::Round($dbSize/1KB, 2)) KB)" -ForegroundColor Green
} else {
    Write-Host "✗ 数据库不存在" -ForegroundColor Red
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  诊断完成" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "💡 提示:" -ForegroundColor Yellow
Write-Host "  - 如果发现配置错误，请检查项目的域名格式"
Write-Host "  - 域名应该类似: example.com 或 www.example.com"
Write-Host "  - 不要使用包含特殊字符的域名"
Write-Host "  - 查看详细文档: CADDYFILE_FIX.md"
Write-Host ""
