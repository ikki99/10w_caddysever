# SSL 证书问题快速修复脚本
# 需要管理员权限运行

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  SSL 证书问题修复工具" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# 检查是否以管理员身份运行
$isAdmin = ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
if (-not $isAdmin) {
    Write-Host "⚠ 警告: 未以管理员身份运行" -ForegroundColor Red
    Write-Host "某些操作需要管理员权限" -ForegroundColor Yellow
    Write-Host ""
}

# 1. 检查端口占用
Write-Host "1. 检查端口占用..." -ForegroundColor Yellow
$port80 = Get-NetTCPConnection -LocalPort 80 -ErrorAction SilentlyContinue
$port443 = Get-NetTCPConnection -LocalPort 443 -ErrorAction SilentlyContinue

if ($port80) {
    Write-Host "  ⚠ 端口 80 被占用" -ForegroundColor Red
    $process80 = Get-Process -Id $port80[0].OwningProcess -ErrorAction SilentlyContinue
    if ($process80) {
        Write-Host "    占用进程: $($process80.ProcessName) (PID: $($process80.Id))" -ForegroundColor Yellow
        
        if ($isAdmin) {
            $stop = Read-Host "是否停止该进程? (y/n)"
            if ($stop -eq 'y') {
                Stop-Process -Id $process80.Id -Force
                Write-Host "    ✓ 进程已停止" -ForegroundColor Green
            }
        }
    }
} else {
    Write-Host "  ✓ 端口 80 可用" -ForegroundColor Green
}

if ($port443) {
    Write-Host "  ⚠ 端口 443 被占用" -ForegroundColor Red
    $process443 = Get-Process -Id $port443[0].OwningProcess -ErrorAction SilentlyContinue
    if ($process443) {
        Write-Host "    占用进程: $($process443.ProcessName) (PID: $($process443.Id))" -ForegroundColor Yellow
        
        if ($isAdmin) {
            $stop = Read-Host "是否停止该进程? (y/n)"
            if ($stop -eq 'y') {
                Stop-Process -Id $process443.Id -Force
                Write-Host "    ✓ 进程已停止" -ForegroundColor Green
            }
        }
    }
} else {
    Write-Host "  ✓ 端口 443 可用" -ForegroundColor Green
}

Write-Host ""

# 2. 检查防火墙规则
Write-Host "2. 配置防火墙规则..." -ForegroundColor Yellow
if ($isAdmin) {
    try {
        # 检查现有规则
        $httpRule = Get-NetFirewallRule -DisplayName "Caddy HTTP" -ErrorAction SilentlyContinue
        $httpsRule = Get-NetFirewallRule -DisplayName "Caddy HTTPS" -ErrorAction SilentlyContinue
        
        if (-not $httpRule) {
            New-NetFirewallRule -DisplayName "Caddy HTTP" -Direction Inbound -Protocol TCP -LocalPort 80 -Action Allow | Out-Null
            Write-Host "  ✓ 已添加 HTTP (80) 防火墙规则" -ForegroundColor Green
        } else {
            Write-Host "  ✓ HTTP (80) 防火墙规则已存在" -ForegroundColor Green
        }
        
        if (-not $httpsRule) {
            New-NetFirewallRule -DisplayName "Caddy HTTPS" -Direction Inbound -Protocol TCP -LocalPort 443 -Action Allow | Out-Null
            Write-Host "  ✓ 已添加 HTTPS (443) 防火墙规则" -ForegroundColor Green
        } else {
            Write-Host "  ✓ HTTPS (443) 防火墙规则已存在" -ForegroundColor Green
        }
    } catch {
        Write-Host "  ✗ 添加防火墙规则失败: $_" -ForegroundColor Red
    }
} else {
    Write-Host "  ⚠ 需要管理员权限才能配置防火墙" -ForegroundColor Yellow
}

Write-Host ""

# 3. 检查域名解析
Write-Host "3. 检查域名解析..." -ForegroundColor Yellow
$domain = Read-Host "请输入你的域名 (例如: c808.333.606.f89f.top)"
if ($domain) {
    try {
        $dnsResult = Resolve-DnsName $domain -ErrorAction Stop
        Write-Host "  ✓ 域名可以解析" -ForegroundColor Green
        Write-Host "    解析到: $($dnsResult.IPAddress -join ', ')" -ForegroundColor Cyan
        
        # 检查是否是 Cloudflare
        if ($dnsResult.IPAddress -like "104.21.*" -or $dnsResult.IPAddress -like "172.67.*") {
            Write-Host "" 
            Write-Host "  ⚠ 检测到 Cloudflare CDN" -ForegroundColor Yellow
            Write-Host "    建议:" -ForegroundColor Cyan
            Write-Host "    1. 使用 Flexible SSL (Cloudflare 控制台)" -ForegroundColor Gray
            Write-Host "    2. 或临时关闭 Cloudflare 代理（橙色云变灰色）" -ForegroundColor Gray
            Write-Host "    3. 或使用 Cloudflare Origin CA 证书" -ForegroundColor Gray
        }
    } catch {
        Write-Host "  ✗ 域名解析失败: $_" -ForegroundColor Red
    }
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  修复建议" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "如果使用 Cloudflare:" -ForegroundColor Yellow
Write-Host "  方案 1: 使用 Flexible SSL (最简单)" -ForegroundColor Cyan
Write-Host "    - Cloudflare 控制台 → SSL/TLS → Flexible"
Write-Host "    - 服务器不需要配置 SSL"
Write-Host ""
Write-Host "  方案 2: 临时关闭代理申请证书" -ForegroundColor Cyan
Write-Host "    - DNS 记录点击橙色云变成灰色"
Write-Host "    - 等待 5-10 分钟"
Write-Host "    - 重新申请 SSL 证书"
Write-Host "    - 成功后重新开启橙色云"
Write-Host ""
Write-Host "如果不使用 Cloudflare:" -ForegroundColor Yellow
Write-Host "  1. 以管理员身份运行 Caddy Manager"
Write-Host "  2. 确保 80 和 443 端口未被占用"
Write-Host "  3. 确保防火墙规则已添加"
Write-Host "  4. 确保域名解析到本服务器公网 IP"
Write-Host ""
Write-Host "详细文档: SSL_TROUBLESHOOTING.md" -ForegroundColor Cyan
Write-Host ""
