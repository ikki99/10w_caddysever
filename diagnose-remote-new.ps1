# Remote Website Diagnostics Script
# Usage: .\diagnose-remote-new.ps1 -Domain "yourdomain.com"

param(
    [Parameter(Mandatory=$true)]
    [string]$Domain
)

Write-Host "============================================" -ForegroundColor Cyan
Write-Host "  Remote Website Diagnostics" -ForegroundColor Cyan
Write-Host "  Domain: $Domain" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""

# 1. DNS Resolution Check
Write-Host "[1/6] Checking DNS Resolution..." -ForegroundColor Yellow
try {
    $dnsResult = Resolve-DnsName -Name $Domain -ErrorAction Stop
    Write-Host "  OK - Domain resolves to:" -ForegroundColor Green
    foreach ($record in $dnsResult) {
        if ($record.IP4Address) {
            Write-Host "    - $($record.IP4Address)" -ForegroundColor Green
        }
    }
} catch {
    Write-Host "  X ERROR - DNS resolution failed: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# 2. Port 80 (HTTP) Check
Write-Host "[2/6] Checking Port 80 (HTTP)..." -ForegroundColor Yellow
try {
    $tcpClient = New-Object System.Net.Sockets.TcpClient
    $tcpClient.ConnectAsync($Domain, 80).Wait(5000) | Out-Null
    if ($tcpClient.Connected) {
        Write-Host "  OK - Port 80 is reachable" -ForegroundColor Green
        $tcpClient.Close()
    } else {
        Write-Host "  X ERROR - Port 80 timeout" -ForegroundColor Red
    }
} catch {
    Write-Host "  X ERROR - Port 80 not reachable: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# 3. Port 443 (HTTPS) Check
Write-Host "[3/6] Checking Port 443 (HTTPS)..." -ForegroundColor Yellow
try {
    $tcpClient = New-Object System.Net.Sockets.TcpClient
    $tcpClient.ConnectAsync($Domain, 443).Wait(5000) | Out-Null
    if ($tcpClient.Connected) {
        Write-Host "  OK - Port 443 is reachable" -ForegroundColor Green
        $tcpClient.Close()
    } else {
        Write-Host "  X ERROR - Port 443 timeout" -ForegroundColor Red
    }
} catch {
    Write-Host "  X ERROR - Port 443 not reachable: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# 4. HTTP Response Check
Write-Host "[4/6] Checking HTTP Response..." -ForegroundColor Yellow
try {
    $httpResponse = Invoke-WebRequest -Uri "http://$Domain" -UseBasicParsing -TimeoutSec 10 -ErrorAction Stop
    Write-Host "  OK - HTTP Status: $($httpResponse.StatusCode)" -ForegroundColor Green
    Write-Host "  Content Length: $($httpResponse.RawContentLength) bytes" -ForegroundColor Gray
} catch {
    Write-Host "  X ERROR - HTTP request failed: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# 5. HTTPS/SSL Check
Write-Host "[5/6] Checking HTTPS/SSL..." -ForegroundColor Yellow
try {
    $httpsResponse = Invoke-WebRequest -Uri "https://$Domain" -UseBasicParsing -TimeoutSec 10 -ErrorAction Stop
    Write-Host "  OK - HTTPS Status: $($httpsResponse.StatusCode)" -ForegroundColor Green
    Write-Host "  SSL Certificate is valid" -ForegroundColor Green
} catch {
    Write-Host "  X ERROR - HTTPS request failed: $($_.Exception.Message)" -ForegroundColor Red
    if ($_.Exception.Message -like "*SSL*" -or $_.Exception.Message -like "*certificate*") {
        Write-Host "  ! This appears to be an SSL certificate issue" -ForegroundColor Yellow
    }
}
Write-Host ""

# 6. Static Resources Check
Write-Host "[6/6] Checking Static Resources..." -ForegroundColor Yellow

$staticResources = @(
    @{Path="/static/app.js"; Type="JavaScript"},
    @{Path="/static/style.css"; Type="CSS"},
    @{Path="/"; Type="HTML"}
)

$failedResources = @()

foreach ($resource in $staticResources) {
    $path = $resource.Path
    $type = $resource.Type
    $url = "https://$Domain$path"
    
    Write-Host "  Testing: $path" -ForegroundColor Gray
    
    try {
        $response = Invoke-WebRequest -Uri $url -UseBasicParsing -TimeoutSec 10
        $size = $response.RawContentLength
        
        if ($size -eq 0) {
            Write-Host "  X [$type] $path - File is empty (0 bytes)" -ForegroundColor Red
            $failedResources += @{
                Path = $path
                Reason = "File is empty"
                Type = $type
            }
        } else {
            Write-Host "  OK [$type] $path - $size bytes" -ForegroundColor Green
        }
    } catch {
        Write-Host "  X [$type] $path - Failed: $($_.Exception.Message)" -ForegroundColor Red
        $failedResources += @{
            Path = $path
            Reason = $_.Exception.Message
            Type = $type
        }
    }
}

Write-Host ""
Write-Host "============================================" -ForegroundColor Cyan
Write-Host "  Diagnostics Summary" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan

if ($failedResources.Count -eq 0) {
    Write-Host "  All checks passed!" -ForegroundColor Green
} else {
    Write-Host "  Found $($failedResources.Count) issues:" -ForegroundColor Red
    foreach ($failed in $failedResources) {
        Write-Host "    - $($failed.Path): $($failed.Reason)" -ForegroundColor Red
    }
}

Write-Host ""
Write-Host "Recommendations:" -ForegroundColor Yellow
Write-Host "  1. Check Caddyfile configuration (reverse_proxy should not have '/' path)" -ForegroundColor Gray
Write-Host "  2. Verify backend service is running on correct port" -ForegroundColor Gray
Write-Host "  3. Check firewall rules for ports 80 and 443" -ForegroundColor Gray
Write-Host "  4. Review Caddy logs: data\logs\caddy.log" -ForegroundColor Gray
Write-Host ""
