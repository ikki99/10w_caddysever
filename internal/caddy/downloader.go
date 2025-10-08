package caddy

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"caddy-manager/internal/config"
)

const (
	// 使用官方最新版本下载地址
	caddyDownloadURL = "https://github.com/caddyserver/caddy/releases/download/v2.10.2/caddy_2.10.2_windows_amd64.zip"
)

var caddyCmd *exec.Cmd

// CheckAndDownload 检查并下载 Caddy
func CheckAndDownload() error {
	// 检查 Caddy 是否已存在
	if _, err := os.Stat(config.CaddyBin); err == nil {
		fmt.Println("✅ Caddy 已安装")
		return nil
	}

	fmt.Println("🔍 未检测到 Caddy，开始自动下载...")
	
	// 创建临时目录
	tempFile := filepath.Join(os.TempDir(), "caddy.zip")
	defer os.Remove(tempFile)

	// 下载 Caddy
	if err := downloadFile(tempFile, caddyDownloadURL); err != nil {
		return fmt.Errorf("下载失败: %v", err)
	}

	// 解压
	if err := unzip(tempFile, config.CaddyDir); err != nil {
		return fmt.Errorf("解压失败: %v", err)
	}

	fmt.Println("✅ Caddy 下载安装完成")
	return nil
}

// AutoStart 自动启动 Caddy
func AutoStart() {
	time.Sleep(2 * time.Second)
	
	// 确保日志目录存在
	logDir := filepath.Dir(config.CaddyLogFile)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Printf("创建日志目录失败: %v", err)
		return
	}
	
	// 检查 Caddyfile 是否存在
	if _, err := os.Stat(config.CaddyConfig); err != nil {
		// 创建默认的 Caddyfile
		defaultConfig := `# Caddy 配置文件
# 通过管理界面添加站点后会自动生成配置

# 默认监听配置
:80 {
	respond "Caddy 正在运行" 200
}
`
		if err := os.WriteFile(config.CaddyConfig, []byte(defaultConfig), 0644); err != nil {
			log.Printf("创建配置文件失败: %v", err)
			return
		}
	}
	
	log.Println("🚀 正在启动 Caddy...")
	if err := Start(); err != nil {
		log.Printf("❌ Caddy 启动失败: %v", err)
	}
}

// Start 启动 Caddy
func Start() error {
	// 先停止已有进程
	Stop()
	
	// 检查 Caddy 可执行文件是否存在
	if _, err := os.Stat(config.CaddyBin); err != nil {
		return fmt.Errorf("Caddy 未安装: %v", err)
	}
	
	// 创建日志文件
	logFile, err := os.OpenFile(config.CaddyLogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	
	caddyCmd = exec.Command(config.CaddyBin, "run", "--config", config.CaddyConfig, "--adapter", "caddyfile")
	caddyCmd.Dir = config.CaddyDir
	caddyCmd.Stdout = logFile
	caddyCmd.Stderr = logFile
	
	if err := caddyCmd.Start(); err != nil {
		logFile.Close()
		return err
	}
	
	log.Println("✅ Caddy 已启动")
	return nil
}

// Stop 停止 Caddy
func Stop() {
	exec.Command("taskkill", "/F", "/IM", "caddy.exe").Run()
	if caddyCmd != nil && caddyCmd.Process != nil {
		caddyCmd.Process.Kill()
	}
}

// Restart 重启 Caddy
func Restart() error {
	log.Println("🔄 重启 Caddy...")
	Stop()
	time.Sleep(1 * time.Second)
	return Start()
}

// Reload 重新加载配置（优雅重启）
func Reload() error {
	log.Println("🔄 重新加载 Caddy 配置...")
	
	// 检查 Caddy 是否在运行
	if !IsRunning() {
		return fmt.Errorf("Caddy 未运行")
	}
	
	// 使用 caddy reload 命令
	cmd := exec.Command(config.CaddyBin, "reload", "--config", config.CaddyConfig, "--adapter", "caddyfile")
	cmd.Dir = config.CaddyDir
	
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("重新加载失败: %v\n输出: %s", err, string(output))
	}
	
	log.Println("✅ Caddy 配置已重新加载")
	return nil
}

// IsRunning 检查 Caddy 是否运行
func IsRunning() bool {
	cmd := exec.Command("tasklist", "/FI", "IMAGENAME eq caddy.exe")
	output, _ := cmd.Output()
	return len(output) > 100
}

// GetLogs 获取 Caddy 日志
func GetLogs(lines int) (string, error) {
	data, err := os.ReadFile(config.CaddyLogFile)
	if err != nil {
		return "", err
	}
	
	content := string(data)
	allLines := strings.Split(content, "\n")
	
	if len(allLines) > lines {
		allLines = allLines[len(allLines)-lines:]
	}
	
	return strings.Join(allLines, "\n"), nil
}

func downloadFile(filepath string, url string) error {
	fmt.Print("📥 正在下载 Caddy")
	
	// 创建文件
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// 下载
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("下载失败，状态码: %d", resp.StatusCode)
	}

	// 获取文件大小
	total := resp.ContentLength
	downloaded := int64(0)

	// 创建缓冲区
	buf := make([]byte, 32*1024)
	for {
		n, err := resp.Body.Read(buf)
		if n > 0 {
			out.Write(buf[:n])
			downloaded += int64(n)
			
			// 显示进度
			if total > 0 {
				percent := float64(downloaded) / float64(total) * 100
				fmt.Printf("\r📥 正在下载 Caddy... %.1f%%", percent)
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
	}
	
	fmt.Println("\n✅ 下载完成")
	return nil
}

func unzip(src, dest string) error {
	fmt.Println("📦 正在解压...")
	
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		// 只解压 caddy.exe
		if !strings.Contains(f.Name, "caddy.exe") {
			continue
		}

		fpath := filepath.Join(dest, filepath.Base(f.Name))

		if f.FileInfo().IsDir() {
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return err
		}

		rc, err := f.Open()
		if err != nil {
			outFile.Close()
			return err
		}

		_, err = io.Copy(outFile, rc)
		outFile.Close()
		rc.Close()

		if err != nil {
			return err
		}
	}

	return nil
}

// GetVersion 获取 Caddy 版本
func GetVersion() string {
	cmd := exec.Command(config.CaddyBin, "version")
	output, err := cmd.Output()
	if err != nil {
		return "未知"
	}
	
	// 解析版本号，例如 "v2.10.2"
	version := strings.TrimSpace(string(output))
	parts := strings.Fields(version)
	if len(parts) > 0 {
		return parts[0]
	}
	
	return "未知"
}
