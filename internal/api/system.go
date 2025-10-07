package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"caddy-manager/internal/database"
)

type EnvInfo struct {
	Name      string `json:"name"`
	Installed bool   `json:"installed"`
	Version   string `json:"version"`
	Path      string `json:"path"`
}

type SystemInfo struct {
	OS          string    `json:"os"`
	Arch        string    `json:"arch"`
	CPUCores    int       `json:"cpu_cores"`
	Environments []EnvInfo `json:"environments"`
}

// GetSystemInfoHandler 获取系统信息
func GetSystemInfoHandler(w http.ResponseWriter, r *http.Request) {
	info := SystemInfo{
		OS:       runtime.GOOS,
		Arch:     runtime.GOARCH,
		CPUCores: runtime.NumCPU(),
		Environments: []EnvInfo{
			detectPHP(),
			detectPython(),
			detectNodeJS(),
			detectJava(),
			detectGo(),
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(info)
}

func detectPHP() EnvInfo {
	info := EnvInfo{Name: "PHP", Installed: false}
	
	cmd := exec.Command("php", "-v")
	output, err := cmd.Output()
	if err == nil {
		info.Installed = true
		lines := strings.Split(string(output), "\n")
		if len(lines) > 0 {
			parts := strings.Fields(lines[0])
			if len(parts) > 1 {
				info.Version = parts[1]
			}
		}
		path, _ := exec.LookPath("php")
		info.Path = path
	}
	
	return info
}

func detectPython() EnvInfo {
	info := EnvInfo{Name: "Python", Installed: false}
	
	cmd := exec.Command("python", "--version")
	output, err := cmd.CombinedOutput()
	if err == nil {
		info.Installed = true
		parts := strings.Fields(string(output))
		if len(parts) > 1 {
			info.Version = parts[1]
		}
		path, _ := exec.LookPath("python")
		info.Path = path
	}
	
	return info
}

func detectNodeJS() EnvInfo {
	info := EnvInfo{Name: "Node.js", Installed: false}
	
	cmd := exec.Command("node", "--version")
	output, err := cmd.Output()
	if err == nil {
		info.Installed = true
		info.Version = strings.TrimSpace(string(output))
		path, _ := exec.LookPath("node")
		info.Path = path
	}
	
	return info
}

func detectJava() EnvInfo {
	info := EnvInfo{Name: "Java", Installed: false}
	
	cmd := exec.Command("java", "-version")
	output, err := cmd.CombinedOutput()
	if err == nil {
		info.Installed = true
		lines := strings.Split(string(output), "\n")
		if len(lines) > 0 {
			parts := strings.Fields(lines[0])
			if len(parts) > 2 {
				info.Version = strings.Trim(parts[2], "\"")
			}
		}
		path, _ := exec.LookPath("java")
		info.Path = path
	}
	
	return info
}

func detectGo() EnvInfo {
	info := EnvInfo{Name: "Go", Installed: false}
	
	cmd := exec.Command("go", "version")
	output, err := cmd.Output()
	if err == nil {
		info.Installed = true
		parts := strings.Fields(string(output))
		if len(parts) > 2 {
			info.Version = strings.TrimPrefix(parts[2], "go")
		}
		path, _ := exec.LookPath("go")
		info.Path = path
	}
	
	return info
}

// InstallEnvGuideHandler 环境安装引导
func InstallEnvGuideHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Env string `json:"env"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	guides := map[string]map[string]string{
		"php": {
			"title": "PHP 安装指南",
			"steps": `1. 访问 https://windows.php.net/download/
2. 下载 PHP 8.x (Thread Safe) 压缩包
3. 解压到 C:\php
4. 添加 C:\php 到系统 PATH 环境变量
5. 重启命令行窗口，输入 php -v 验证`,
			"download": "https://windows.php.net/download/",
		},
		"python": {
			"title": "Python 安装指南",
			"steps": `1. 访问 https://www.python.org/downloads/
2. 下载最新版 Python 安装程序
3. 运行安装程序，勾选 "Add Python to PATH"
4. 点击 Install Now
5. 安装完成后，打开命令行输入 python --version 验证`,
			"download": "https://www.python.org/downloads/",
		},
		"nodejs": {
			"title": "Node.js 安装指南",
			"steps": `1. 访问 https://nodejs.org/
2. 下载 LTS 版本安装程序
3. 运行安装程序，按默认选项安装
4. 安装完成后，打开命令行输入 node --version 验证
5. 同时会安装 npm 包管理器`,
			"download": "https://nodejs.org/",
		},
		"java": {
			"title": "Java 安装指南",
			"steps": `1. 访问 https://adoptium.net/
2. 下载 Java 17 LTS 版本
3. 运行安装程序
4. 设置 JAVA_HOME 环境变量
5. 添加 %JAVA_HOME%\bin 到 PATH
6. 重启命令行，输入 java -version 验证`,
			"download": "https://adoptium.net/",
		},
		"go": {
			"title": "Go 安装指南",
			"steps": `1. 访问 https://go.dev/dl/
2. 下载最新版 Go 安装程序
3. 运行安装程序，按默认选项安装
4. 安装完成后，打开命令行输入 go version 验证
5. 设置 GOPATH 环境变量（可选）`,
			"download": "https://go.dev/dl/",
		},
	}

	guide, exists := guides[req.Env]
	if !exists {
		http.Error(w, "未知环境", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(guide)
}

// BrowseFilesHandler 浏览文件
func BrowseFilesHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	
	// 获取 www 根目录设置
	db := database.GetDB()
	var wwwRoot string
	db.QueryRow("SELECT value FROM settings WHERE key = 'www_root'").Scan(&wwwRoot)
	
	if path == "" {
		path = wwwRoot
	}
	
	// 确保路径在 www 根目录下（安全检查）
	absPath, err := filepath.Abs(path)
	if err != nil {
		http.Error(w, "无效路径", http.StatusBadRequest)
		return
	}
	
	absWwwRoot, _ := filepath.Abs(wwwRoot)
	if !strings.HasPrefix(absPath, absWwwRoot) {
		absPath = absWwwRoot
	}
	
	// 读取目录
	entries, err := os.ReadDir(absPath)
	if err != nil {
		// 如果目录不存在，创建它
		os.MkdirAll(absPath, 0755)
		entries, _ = os.ReadDir(absPath)
	}
	
	type FileInfo struct {
		Name  string `json:"name"`
		IsDir bool   `json:"is_dir"`
		Size  int64  `json:"size"`
		Path  string `json:"path"`
	}
	
	var files []FileInfo
	for _, entry := range entries {
		info, _ := entry.Info()
		files = append(files, FileInfo{
			Name:  entry.Name(),
			IsDir: entry.IsDir(),
			Size:  info.Size(),
			Path:  filepath.Join(absPath, entry.Name()),
		})
	}
	
	result := map[string]interface{}{
		"current_path": absPath,
		"parent_path":  filepath.Dir(absPath),
		"files":        files,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// CreateFolderHandler 创建文件夹
func CreateFolderHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Path string `json:"path"`
		Name string `json:"name"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	folderPath := filepath.Join(req.Path, req.Name)
	if err := os.MkdirAll(folderPath, 0755); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
}

// DeleteFileHandler 删除文件或文件夹
func DeleteFileHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	
	if err := os.RemoveAll(path); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
}

// UploadFileHandler 上传文件
func UploadFileHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()
	
	uploadPath := r.FormValue("path")
	if uploadPath == "" {
		db := database.GetDB()
		db.QueryRow("SELECT value FROM settings WHERE key = 'www_root'").Scan(&uploadPath)
	}
	
	destPath := filepath.Join(uploadPath, header.Filename)
	
	dst, err := os.Create(destPath)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer dst.Close()
	
	if _, err := dst.ReadFrom(file); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
}

// DownloadFileHandler 下载文件
func DownloadFileHandler(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		http.Error(w, "文件不存在", http.StatusNotFound)
		return
	}
	
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filepath.Base(path)))
	http.ServeFile(w, r, path)
}
