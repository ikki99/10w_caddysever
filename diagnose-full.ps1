# Caddy Manager Comprehensive Diagnostics Tool
# This script checks the complete system status

Write-Host "============================================" -ForegroundColor Cyan
Write-Host "    Caddy Manager Diagnostics v1.0" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""

$issues = @()
$warnings = @()

# 1. Check Administrator Privileges
Write-Host "1. Checking Administrator Privileges..." -ForegroundColor Yellow
try {
    $isAdmin = ([Security.Principal.WindowsPrincipal][Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
    if ($isAdmin) {
        Write-Host "   OK Running as Administrator" -ForegroundColor Green
    } else {
        Write-Host "   WARNING Not running as Administrator" -ForegroundColor Yellow
        $warnings += "Not running as Administrator - Some features may not work (binding ports 80, 443)"
    }
} catch {
    Write-Host "   ERROR Cannot check administrator status" -ForegroundColor Red
}
Write-Host ""

# 2. Check Caddy Manager Process
Write-Host "2. Checking Caddy Manager Status..." -ForegroundColor Yellow
$managerRunning = Get-Process | Where-Object { $_.ProcessName -like "*caddy-manager*" }
if ($managerRunning) {
    Write-Host "   OK Caddy Manager is running (PID: $($managerRunning.Id))" -ForegroundColor Green
} else {
    Write-Host "   WARNING Caddy Manager is not running" -ForegroundColor Yellow
}
Write-Host ""

# 3. Check Caddy Process
Write-Host "3. Checking Caddy Service..." -ForegroundColor Yellow
$caddyRunning = Get-Process | Where-Object { $_.ProcessName -eq "caddy" }
if ($caddyRunning) {
    Write-Host "   OK Caddy is running (PID: $($caddyRunning.Id))" -ForegroundColor Green
} else {
    Write-Host "   WARNING Caddy is not running" -ForegroundColor Yellow
    $issues += "Caddy service is not running - Projects won't be accessible"
}
Write-Host ""

# 4. Check Port Status
Write-Host "4. Checking Port Status..." -ForegroundColor Yellow
$portCheck = @(80, 443, 8989, 2019)
foreach ($port in $portCheck) {
    $conn = Get-NetTCPConnection -LocalPort $port -ErrorAction SilentlyContinue | Where-Object { $_.State -eq "Listen" }
    if ($conn) {
        $proc = Get-Process -Id $conn.OwningProcess -ErrorAction SilentlyContinue
        Write-Host "   OK Port $port is listening ($($proc.ProcessName))" -ForegroundColor Green
    } else {
        if ($port -eq 80 -or $port -eq 443) {
            Write-Host "   WARNING Port $port is not listening" -ForegroundColor Yellow
            $warnings += "Port $port is not listening - Web services may not work"
        } elseif ($port -eq 8989) {
            Write-Host "   WARNING Port $port is not listening (Manager UI)" -ForegroundColor Yellow
            $issues += "Manager UI port 8989 is not listening"
        } else {
            Write-Host "   INFO Port $port is not listening" -ForegroundColor Gray
        }
    }
}
Write-Host ""

# 5. Check Firewall Rules
Write-Host "5. Checking Windows Firewall..." -ForegroundColor Yellow
try {
    $firewallStatus = Get-NetFirewallProfile | Select-Object Name, Enabled
    $enabled = $firewallStatus | Where-Object { $_.Enabled -eq $true }
    if ($enabled) {
        Write-Host "   INFO Firewall is enabled for: $($enabled.Name -join ', ')" -ForegroundColor Cyan
        Write-Host "   INFO Make sure ports 80, 443, 8989 are allowed" -ForegroundColor Cyan
    } else {
        Write-Host "   INFO Firewall is disabled" -ForegroundColor Gray
    }
} catch {
    Write-Host "   WARNING Cannot check firewall status" -ForegroundColor Yellow
}
Write-Host ""

# 6. Check Data Directory
Write-Host "6. Checking Data Directory..." -ForegroundColor Yellow
$dataDir = ".\data"
if (Test-Path $dataDir) {
    Write-Host "   OK Data directory exists: $dataDir" -ForegroundColor Green
    
    $caddyfile = Join-Path $dataDir "caddy\Caddyfile"
    if (Test-Path $caddyfile) {
        $lines = (Get-Content $caddyfile | Measure-Object -Line).Lines
        Write-Host "   OK Caddyfile exists ($lines lines)" -ForegroundColor Green
    } else {
        Write-Host "   WARNING Caddyfile not found" -ForegroundColor Yellow
        $issues += "Caddyfile not found at $caddyfile"
    }
    
    $db = Join-Path $dataDir "caddy-manager.db"
    if (Test-Path $db) {
        $size = (Get-Item $db).Length / 1KB
        Write-Host "   OK Database exists ($([math]::Round($size, 2)) KB)" -ForegroundColor Green
    } else {
        Write-Host "   ERROR Database not found" -ForegroundColor Red
        $issues += "Database not found - Manager won't work properly"
    }
} else {
    Write-Host "   ERROR Data directory not found" -ForegroundColor Red
    $issues += "Data directory missing: $dataDir"
}
Write-Host ""

# 7. Check DNS Resolution
Write-Host "7. Testing DNS Resolution..." -ForegroundColor Yellow
try {
    $testDomain = "google.com"
    $dns = Resolve-DnsName $testDomain -ErrorAction Stop
    Write-Host "   OK DNS resolution working" -ForegroundColor Green
} catch {
    Write-Host "   ERROR DNS resolution failed" -ForegroundColor Red
    $issues += "DNS resolution not working - Domain SSL won't work"
}
Write-Host ""

# 8. Check Internet Connectivity
Write-Host "8. Testing Internet Connectivity..." -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Uri "https://www.google.com" -UseBasicParsing -TimeoutSec 5 -ErrorAction Stop
    Write-Host "   OK Internet connection working" -ForegroundColor Green
} catch {
    Write-Host "   WARNING Cannot reach internet" -ForegroundColor Yellow
    $warnings += "Internet connectivity issues - SSL certificate requests may fail"
}
Write-Host ""

# 9. Check Caddy Binary
Write-Host "9. Checking Caddy Binary..." -ForegroundColor Yellow
$caddyPath = ".\data\caddy\caddy.exe"
if (Test-Path $caddyPath) {
    $version = & $caddyPath version 2>$null
    if ($version) {
        Write-Host "   OK Caddy version: $version" -ForegroundColor Green
    } else {
        Write-Host "   WARNING Cannot get Caddy version" -ForegroundColor Yellow
    }
} else {
    Write-Host "   ERROR Caddy binary not found at $caddyPath" -ForegroundColor Red
    $issues += "Caddy binary missing - System won't work"
}
Write-Host ""

# 10. Check Logs
Write-Host "10. Checking Recent Errors in Logs..." -ForegroundColor Yellow
$logDir = ".\data\logs"
if (Test-Path $logDir) {
    $caddyLog = Join-Path $logDir "caddy.log"
    if (Test-Path $caddyLog) {
        $recentErrors = Get-Content $caddyLog -Tail 50 | Select-String -Pattern "error|failed|fatal" -CaseSensitive:$false
        if ($recentErrors) {
            Write-Host "   WARNING Found $($recentErrors.Count) recent errors in Caddy log" -ForegroundColor Yellow
            Write-Host "   Recent errors:" -ForegroundColor Gray
            $recentErrors | Select-Object -First 3 | ForEach-Object {
                Write-Host "     - $($_.Line.Substring(0, [Math]::Min(100, $_.Line.Length)))" -ForegroundColor Gray
            }
        } else {
            Write-Host "   OK No recent errors in Caddy log" -ForegroundColor Green
        }
    } else {
        Write-Host "   INFO Caddy log not found (may not have run yet)" -ForegroundColor Gray
    }
} else {
    Write-Host "   INFO Log directory not found" -ForegroundColor Gray
}
Write-Host ""

# Summary
Write-Host "============================================" -ForegroundColor Cyan
Write-Host "           DIAGNOSIS SUMMARY" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""

if ($issues.Count -eq 0 -and $warnings.Count -eq 0) {
    Write-Host "OK System appears healthy!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Next steps:" -ForegroundColor Cyan
    Write-Host "  1. Access Manager UI at http://localhost:8989" -ForegroundColor White
    Write-Host "  2. Create a new project" -ForegroundColor White
    Write-Host "  3. Configure your domain" -ForegroundColor White
} else {
    if ($issues.Count -gt 0) {
        Write-Host "CRITICAL ISSUES FOUND:" -ForegroundColor Red
        $issues | ForEach-Object { Write-Host "  X $_" -ForegroundColor Red }
        Write-Host ""
    }
    
    if ($warnings.Count -gt 0) {
        Write-Host "WARNINGS:" -ForegroundColor Yellow
        $warnings | ForEach-Object { Write-Host "  ! $_" -ForegroundColor Yellow }
        Write-Host ""
    }
    
    Write-Host "RECOMMENDATIONS:" -ForegroundColor Cyan
    Write-Host ""
    
    if (-not $isAdmin) {
        Write-Host "  1. Run as Administrator:" -ForegroundColor White
        Write-Host "     - Right-click caddy-manager.exe" -ForegroundColor Gray
        Write-Host "     - Select 'Run as administrator'" -ForegroundColor Gray
        Write-Host ""
    }
    
    if (-not $caddyRunning) {
        Write-Host "  2. Start Caddy Service:" -ForegroundColor White
        Write-Host "     - Launch Caddy Manager" -ForegroundColor Gray
        Write-Host "     - Caddy should auto-start" -ForegroundColor Gray
        Write-Host ""
    }
    
    Write-Host "  3. Check documentation:" -ForegroundColor White
    Write-Host "     - README.md" -ForegroundColor Gray
    Write-Host "     - TROUBLESHOOTING.md" -ForegroundColor Gray
    Write-Host "     - SSL_TROUBLESHOOTING.md" -ForegroundColor Gray
}

Write-Host ""
Write-Host "============================================" -ForegroundColor Cyan
Write-Host "For detailed help, see TROUBLESHOOTING.md" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
