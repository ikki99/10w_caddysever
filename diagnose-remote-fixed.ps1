# Remote Site Diagnostics Tool
param(
    [string]$Domain = "c808.333.606.f89f.top"
)

Write-Host "============================================" -ForegroundColor Cyan
Write-Host "     Remote Site Diagnostics" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
Write-Host ""

$baseUrl = "https://$Domain"

# 1. DNS Resolution
Write-Host "1. DNS Resolution Check" -ForegroundColor Yellow
Write-Host "-------------------" -ForegroundColor Gray
try {
    $ips = [System.Net.Dns]::GetHostAddresses($Domain)
    foreach ($ip in $ips) {
        Write-Host "  IP: $ip" -ForegroundColor Green
        
        # Check for Cloudflare
        $ipStr = $ip.ToString()
        if ($ipStr -match "^(104\.21\.|172\.67\.|104\.18\.)") {
            Write-Host "  Warning: Detected Cloudflare CDN" -ForegroundColor Yellow
        }
    }
} catch {
    Write-Host "  X DNS resolution failed: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# 2. SSL Certificate
Write-Host "2. SSL Certificate Check" -ForegroundColor Yellow
Write-Host "-------------------" -ForegroundColor Gray
try {
    $req = [System.Net.HttpWebRequest]::Create($baseUrl)
    $req.Timeout = 10000
    $response = $req.GetResponse()
    Write-Host "  OK HTTPS accessible" -ForegroundColor Green
    Write-Host "  Status: $($response.StatusCode)" -ForegroundColor Green
    $response.Close()
} catch {
    Write-Host "  X HTTPS access failed: $($_.Exception.Message)" -ForegroundColor Red
}
Write-Host ""

# 3. Homepage Content
Write-Host "3. Homepage Check" -ForegroundColor Yellow
Write-Host "-------------------" -ForegroundColor Gray
try {
    $html = (Invoke-WebRequest -Uri $baseUrl -UseBasicParsing).Content
    Write-Host "  OK Page size: $($html.Length) bytes" -ForegroundColor Green
    
    # Extract resources
    $cssLinks = [regex]::Matches($html, 'href=[''"]([^''"]*\.css[^''"]*)') | ForEach-Object { $_.Groups[1].Value }
    $jsLinks = [regex]::Matches($html, 'src=[''"]([^''"]*\.js[^''"]*)') | ForEach-Object { $_.Groups[1].Value }
    
    Write-Host "  Found $($cssLinks.Count) CSS files" -ForegroundColor Cyan
    Write-Host "  Found $($jsLinks.Count) JS files" -ForegroundColor Cyan
} catch {
    Write-Host "  X Cannot access homepage: $($_.Exception.Message)" -ForegroundColor Red
    exit 1
}
Write-Host ""

# 4. Static Resources Check
Write-Host "4. Static Resources Check" -ForegroundColor Yellow
Write-Host "-------------------" -ForegroundColor Gray

$allResources = @()
$cssLinks | ForEach-Object { 
    if ($_ -notmatch "^http") { 
        $allResources += @{Type="CSS"; Path=$_} 
    }
}
$jsLinks | ForEach-Object { 
    if ($_ -notmatch "^http") { 
        $allResources += @{Type="JS"; Path=$_} 
    }
}

$failedResources = @()

foreach ($resource in $allResources) {
    $path = $resource.Path
    $type = $resource.Type
    
    # Handle relative paths
    if ($path -notmatch "^/") {
        $path = "/$path"
    }
    
    $url = $baseUrl + $path
    
    try {
        $response = Invoke-WebRequest -Uri $url -UseBasicParsing -TimeoutSec 10
        $size = $response.RawContentLength
        $contentType = $response.Headers['Content-Type']
        
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
        $statusCode = "Unknown"
        if ($_.Exception.Response) {
            $statusCode = $_.Exception.Response.StatusCode.value__
        }
        
        Write-Host "  X [$type] $path - HTTP $statusCode" -ForegroundColor Red
        $failedResources += @{
            Path = $path
            Reason = "HTTP $statusCode"
            Type = $type
        }
    }
}
Write-Host ""

# 5. Diagnosis Results
Write-Host "5. Diagnosis Summary" -ForegroundColor Yellow
Write-Host "-------------------" -ForegroundColor Gray

if ($failedResources.Count -eq 0) {
    Write-Host "  OK No issues found, all resources load correctly" -ForegroundColor Green
} else {
    Write-Host "  X Found $($failedResources.Count) failed resources" -ForegroundColor Red
    Write-Host ""
    
    foreach ($failed in $failedResources) {
        Write-Host "  Issue: $($failed.Path)" -ForegroundColor Red
        Write-Host "  Reason: $($failed.Reason)" -ForegroundColor Yellow
        Write-Host ""
    }
    
    # Analysis
    Write-Host "Possible Causes:" -ForegroundColor Cyan
    Write-Host ""
    
    $hasEmptyFiles = $failedResources | Where-Object { $_.Reason -eq "File is empty" }
    if ($hasEmptyFiles) {
        Write-Host "  [Empty File Problem]" -ForegroundColor Yellow
        Write-Host "  Static files on server may not exist or are empty" -ForegroundColor White
        Write-Host ""
        Write-Host "  Solutions:" -ForegroundColor Green
        Write-Host "  1. SSH to server and check files:" -ForegroundColor White
        foreach ($file in $hasEmptyFiles) {
            Write-Host "     ls -lh /path/to/project$($file.Path)" -ForegroundColor Gray
        }
        Write-Host "  2. Confirm files exist and are not empty" -ForegroundColor White
        Write-Host "  3. Check file permissions (should be 644)" -ForegroundColor White
        Write-Host "  4. Re-upload files" -ForegroundColor White
        Write-Host ""
    }
    
    $has404 = $failedResources | Where-Object { $_.Reason -match "404" }
    if ($has404) {
        Write-Host "  [404 File Not Found]" -ForegroundColor Yellow
        Write-Host "  Static file path misconfigured or files not uploaded" -ForegroundColor White
        Write-Host ""
        Write-Host "  Solutions:" -ForegroundColor Green
        Write-Host "  1. Check Caddy configuration for static file path" -ForegroundColor White
        Write-Host "  2. Check resource paths in HTML are correct" -ForegroundColor White
        Write-Host "  3. Upload missing files to server" -ForegroundColor White
        Write-Host ""
    }
    
    Write-Host "  [Check Caddy Configuration]" -ForegroundColor Yellow
    Write-Host "  Caddyfile should contain:" -ForegroundColor White
    Write-Host @"
  
  $Domain {
      reverse_proxy localhost:PORT
      
      # Or for static files:
      file_server
      root * /path/to/static
  }
"@ -ForegroundColor Gray
    Write-Host ""
    
    Write-Host "  [Use Browser Developer Tools]" -ForegroundColor Yellow
    Write-Host "  1. Open https://$Domain" -ForegroundColor White
    Write-Host "  2. Press F12 to open Developer Tools" -ForegroundColor White
    Write-Host "  3. Network tab -> Refresh page" -ForegroundColor White
    Write-Host "  4. View failed requests (red)" -ForegroundColor White
    Write-Host "  5. Click for detailed error information" -ForegroundColor White
}

Write-Host ""
Write-Host "============================================" -ForegroundColor Cyan
Write-Host "Diagnosis Complete!" -ForegroundColor Cyan
Write-Host "============================================" -ForegroundColor Cyan
