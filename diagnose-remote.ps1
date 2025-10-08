# Caddy Remote Project Diagnostics Script
# Usage: .\diagnose-remote.ps1 -Domain "your-domain.com"

param(
    [Parameter(Mandatory=$true)]
    [string]$Domain
)

Write-Host ""
Write-Host "============================================" -ForegroundColor Cyan
Write-Host "  Caddy Remote Project Diagnostics" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Domain: $Domain" -ForegroundColor Yellow
Write-Host ""

# Test DNS Resolution
Write-Host "[1] DNS Resolution Check" -ForegroundColor Green
Write-Host "----------------------------------------" -ForegroundColor Gray
try {
    $dnsResult = Resolve-DnsName -Name $Domain -ErrorAction Stop
    Write-Host "  [OK] DNS resolved successfully" -ForegroundColor Green
    foreach ($record in $dnsResult) {
        if ($record.Type -eq 'A') {
            Write-Host "    IP: $($record.IPAddress)" -ForegroundColor White
        }
    }
} catch {
    Write-Host "  [ERROR] DNS resolution failed" -ForegroundColor Red
    Write-Host "    Error: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# Test HTTP/HTTPS Connection
Write-Host "[2] HTTP/HTTPS Connection Test" -ForegroundColor Green
Write-Host "----------------------------------------" -ForegroundColor Gray

# Test HTTPS
Write-Host "  Testing HTTPS connection..." -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Uri "https://$Domain" -TimeoutSec 10 -UseBasicParsing -ErrorAction Stop
    Write-Host "  [OK] HTTPS connection successful" -ForegroundColor Green
    Write-Host "    Status: $($response.StatusCode) $($response.StatusDescription)" -ForegroundColor White
} catch {
    Write-Host "  [ERROR] HTTPS connection failed" -ForegroundColor Red
    Write-Host "    Error: $($_.Exception.Message)" -ForegroundColor Red
}

# Test HTTP
Write-Host "  Testing HTTP connection..." -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Uri "http://$Domain" -TimeoutSec 10 -UseBasicParsing -ErrorAction Stop
    Write-Host "  [OK] HTTP connection successful" -ForegroundColor Green
    Write-Host "    Status: $($response.StatusCode) $($response.StatusDescription)" -ForegroundColor White
} catch {
    Write-Host "  [ERROR] HTTP connection failed" -ForegroundColor Red
    Write-Host "    Error: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# Test SSL Certificate
Write-Host "[3] SSL Certificate Check" -ForegroundColor Green
Write-Host "----------------------------------------" -ForegroundColor Gray
try {
    $tcpClient = New-Object System.Net.Sockets.TcpClient($Domain, 443)
    $sslStream = New-Object System.Net.Security.SslStream($tcpClient.GetStream(), $false)
    $sslStream.AuthenticateAsClient($Domain)
    
    $cert = $sslStream.RemoteCertificate
    $cert2 = New-Object System.Security.Cryptography.X509Certificates.X509Certificate2($cert)
    
    Write-Host "  [OK] SSL Certificate found" -ForegroundColor Green
    Write-Host "    Subject: $($cert2.Subject)" -ForegroundColor White
    Write-Host "    Issuer: $($cert2.Issuer)" -ForegroundColor White
    Write-Host "    Valid From: $($cert2.NotBefore)" -ForegroundColor White
    Write-Host "    Valid To: $($cert2.NotAfter)" -ForegroundColor White
    
    if ($cert2.NotAfter -lt (Get-Date)) {
        Write-Host "    [WARNING] Certificate has expired!" -ForegroundColor Yellow
    } elseif ($cert2.NotAfter -lt (Get-Date).AddDays(30)) {
        Write-Host "    [WARNING] Certificate expires soon!" -ForegroundColor Yellow
    } else {
        Write-Host "    [OK] Certificate is valid" -ForegroundColor Green
    }
    
    $sslStream.Close()
    $tcpClient.Close()
} catch {
    Write-Host "  [ERROR] SSL Certificate check failed" -ForegroundColor Red
    Write-Host "    Error: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# Test Port Connectivity
Write-Host "[4] Port Connectivity Test" -ForegroundColor Green
Write-Host "----------------------------------------" -ForegroundColor Gray

$ports = @(80, 443)
foreach ($port in $ports) {
    Write-Host "  Testing port $port..." -ForegroundColor Yellow
    $tcpTest = Test-NetConnection -ComputerName $Domain -Port $port -WarningAction SilentlyContinue
    if ($tcpTest.TcpTestSucceeded) {
        Write-Host "    [OK] Port $port is accessible" -ForegroundColor Green
    } else {
        Write-Host "    [ERROR] Port $port is not accessible" -ForegroundColor Red
    }
}
Write-Host ""

Write-Host "============================================" -ForegroundColor Cyan
Write-Host "  Diagnostics Complete" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""