# 混合内容检测脚本
# 用于查找项目中所有可能导致 HTTPS 黄色叹号的 HTTP 资源

param(
    [string]$ProjectPath = ".",
    [switch]$ShowDetails,
    [switch]$AutoFix
)

$ErrorActionPreference = "SilentlyContinue"

Write-Host ""
Write-Host "===============================================" -ForegroundColor Cyan
Write-Host "        混合内容检测工具" -ForegroundColor Cyan
Write-Host "===============================================" -ForegroundColor Cyan
Write-Host ""

if (-not (Test-Path $ProjectPath)) {
    Write-Host "❌ 错误：路径不存在：$ProjectPath" -ForegroundColor Red
    exit 1
}

Write-Host "🔍 扫描目录：$ProjectPath" -ForegroundColor Yellow
Write-Host ""

# 定义要检查的文件类型
$fileExtensions = @("*.html", "*.htm", "*.php", "*.js", "*.css", "*.vue", "*.jsx", "*.tsx", "*.json")

# 查找所有匹配的文件
$files = Get-ChildItem -Path $ProjectPath -Recurse -Include $fileExtensions -File

if ($files.Count -eq 0) {
    Write-Host "⚠️  未找到任何需要检查的文件" -ForegroundColor Yellow
    exit 0
}

Write-Host "📁 找到 $($files.Count) 个文件需要检查" -ForegroundColor Green
Write-Host ""

# 存储结果
$results = @()
$totalIssues = 0

foreach ($file in $files) {
    $content = Get-Content $file.FullName -Raw -ErrorAction SilentlyContinue
    if (-not $content) { continue }
    
    # 查找 HTTP 链接（但排除 HTTPS）
    $pattern = 'http://[^"\s''<>)]+|src=["'']http://|href=["'']http://|url\([''"]?http://'
    $matches = [regex]::Matches($content, $pattern)
    
    if ($matches.Count -gt 0) {
        $lines = $content -split "`n"
        $fileIssues = @()
        
        foreach ($match in $matches) {
            # 跳过已经是 HTTPS 的
            if ($match.Value -match "https://") { continue }
            
            # 查找行号
            $lineNum = 1
            $charCount = 0
            foreach ($line in $lines) {
                $charCount += $line.Length + 1
                if ($charCount -ge $match.Index) {
                    break
                }
                $lineNum++
            }
            
            $fileIssues += [PSCustomObject]@{
                File = $file.FullName
                RelativePath = $file.FullName.Replace((Get-Location).Path, ".")
                Line = $lineNum
                MatchedText = $match.Value
                LineContent = $lines[$lineNum - 1].Trim()
            }
        }
        
        if ($fileIssues.Count -gt 0) {
            $results += $fileIssues
            $totalIssues += $fileIssues.Count
            
            Write-Host "⚠️  $($file.Name) - 发现 $($fileIssues.Count) 个问题" -ForegroundColor Yellow
            
            if ($ShowDetails) {
                foreach ($issue in $fileIssues) {
                    Write-Host "   第 $($issue.Line) 行: " -NoNewline -ForegroundColor Gray
                    Write-Host $issue.MatchedText -ForegroundColor Red
                }
                Write-Host ""
            }
        }
    }
}

Write-Host ""
Write-Host "===============================================" -ForegroundColor Cyan
Write-Host "检测结果汇总" -ForegroundColor Cyan
Write-Host "===============================================" -ForegroundColor Cyan
Write-Host ""

if ($totalIssues -eq 0) {
    Write-Host "✅ 太好了！未发现混合内容问题" -ForegroundColor Green
    Write-Host ""
    Write-Host "您的网站应该显示绿色锁头 🔒" -ForegroundColor Green
} else {
    Write-Host "发现 $totalIssues 个可能的混合内容问题" -ForegroundColor Red
    Write-Host ""
    
    # 按文件分组显示
    $groupedResults = $results | Group-Object -Property RelativePath
    
    Write-Host "问题文件列表：" -ForegroundColor Yellow
    Write-Host ""
    
    foreach ($group in $groupedResults) {
        Write-Host "📄 $($group.Name)" -ForegroundColor Cyan
        Write-Host "   问题数：$($group.Count)" -ForegroundColor Yellow
        
        if ($ShowDetails) {
            foreach ($item in $group.Group) {
                Write-Host "   ├─ 第 $($item.Line) 行" -ForegroundColor Gray
                Write-Host "   │  $($item.MatchedText)" -ForegroundColor Red
                Write-Host "   │  上下文: $($item.LineContent.Substring(0, [Math]::Min(80, $item.LineContent.Length)))..." -ForegroundColor DarkGray
            }
        }
        Write-Host ""
    }
    
    Write-Host ""
    Write-Host "===============================================" -ForegroundColor Cyan
    Write-Host "修复建议" -ForegroundColor Cyan
    Write-Host "===============================================" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "1️⃣  手动修复（推荐）" -ForegroundColor Green
    Write-Host "   打开上述文件，将 http:// 改为 https:// 或使用相对路径" -ForegroundColor Gray
    Write-Host ""
    Write-Host "2️⃣  自动修复（谨慎使用）" -ForegroundColor Yellow
    Write-Host "   运行: .\检测混合内容.ps1 -AutoFix" -ForegroundColor Gray
    Write-Host "   注意：会自动将所有 http:// 替换为 https://" -ForegroundColor Red
    Write-Host ""
    Write-Host "3️⃣  添加 CSP 头部" -ForegroundColor Cyan
    Write-Host "   在 Caddyfile 中添加:" -ForegroundColor Gray
    Write-Host "   header Content-Security-Policy `"upgrade-insecure-requests`"" -ForegroundColor DarkCyan
    Write-Host ""
    
    if ($AutoFix) {
        Write-Host ""
        Write-Host "⚠️  准备自动修复..." -ForegroundColor Yellow
        $confirm = Read-Host "这将修改 $($groupedResults.Count) 个文件，是否继续？(y/N)"
        
        if ($confirm -eq 'y' -or $confirm -eq 'Y') {
            Write-Host ""
            Write-Host "🔧 开始修复..." -ForegroundColor Green
            
            $fixedFiles = 0
            foreach ($group in $groupedResults) {
                $filePath = Join-Path (Get-Location) $group.Name.TrimStart('.')
                
                if (Test-Path $filePath) {
                    # 备份文件
                    $backupPath = "$filePath.bak"
                    Copy-Item $filePath $backupPath -Force
                    
                    # 读取并替换
                    $content = Get-Content $filePath -Raw
                    $newContent = $content -replace 'http://', 'https://'
                    Set-Content $filePath $newContent -NoNewline
                    
                    Write-Host "   ✓ 已修复: $($group.Name)" -ForegroundColor Green
                    Write-Host "   备份: $backupPath" -ForegroundColor Gray
                    $fixedFiles++
                }
            }
            
            Write-Host ""
            Write-Host "✅ 修复完成！共修复 $fixedFiles 个文件" -ForegroundColor Green
            Write-Host ""
            Write-Host "⚠️  重要提示：" -ForegroundColor Yellow
            Write-Host "   1. 请测试网站确保一切正常" -ForegroundColor Gray
            Write-Host "   2. 如有问题，可用 .bak 文件恢复" -ForegroundColor Gray
            Write-Host "   3. 清除浏览器缓存并硬刷新（Ctrl+F5）" -ForegroundColor Gray
        } else {
            Write-Host "已取消自动修复" -ForegroundColor Yellow
        }
    }
}

Write-Host ""
Write-Host "💡 提示：" -ForegroundColor Cyan
Write-Host "   使用 -ShowDetails 参数查看详细信息" -ForegroundColor Gray
Write-Host "   使用 -AutoFix 参数自动修复（需确认）" -ForegroundColor Gray
Write-Host ""
Write-Host "示例：" -ForegroundColor Cyan
Write-Host "   .\检测混合内容.ps1 -ShowDetails" -ForegroundColor DarkCyan
Write-Host "   .\检测混合内容.ps1 -ProjectPath C:\www\mysite -ShowDetails" -ForegroundColor DarkCyan
Write-Host ""

# 导出结果到文件
if ($totalIssues -gt 0) {
    $reportPath = Join-Path (Get-Location) "mixed-content-report.txt"
    $results | Format-Table -AutoSize | Out-File $reportPath
    Write-Host "📊 详细报告已保存到: $reportPath" -ForegroundColor Green
    Write-Host ""
}
