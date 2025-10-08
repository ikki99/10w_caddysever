package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"caddy-manager/internal/caddy"
	"caddy-manager/internal/config"
	"caddy-manager/internal/database"
	"caddy-manager/internal/models"
)

var (
	projectProcesses = make(map[int]*exec.Cmd)
	processMutex     sync.RWMutex
)

// ProjectsHandler 获取项目列表
func ProjectsHandler(w http.ResponseWriter, r *http.Request) {
	db := database.GetDB()
	rows, err := db.Query("SELECT id, name, project_type, root_dir, exec_path, port, start_command, auto_start, status, domains, ssl_enabled, description, COALESCE(use_ipv4, 1) FROM projects ORDER BY created_at DESC")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var projects []models.Project
	for rows.Next() {
		var p models.Project
		var execPath, startCmd, domains, desc *string
		if err := rows.Scan(&p.ID, &p.Name, &p.ProjectType, &p.RootDir, &execPath, &p.Port, &startCmd, &p.AutoStart, &p.Status, &domains, &p.SSLEnabled, &desc, &p.UseIPv4); err != nil {
			continue
		}
		if execPath != nil {
			p.ExecPath = *execPath
		}
		if startCmd != nil {
			p.StartCommand = *startCmd
		}
		if domains != nil {
			p.Domains = *domains
		}
		if desc != nil {
			p.Description = *desc
		}
		
		// 更新实时状态
		p.Status = getProjectStatus(p.ID, p.Port)
		
		projects = append(projects, p)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(projects)
}

// AddProjectHandler 添加项目
func AddProjectHandler(w http.ResponseWriter, r *http.Request) {
	var p models.Project
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 默认使用 IPv4
	if !p.UseIPv4 {
		p.UseIPv4 = true
	}

	db := database.GetDB()
	result, err := db.Exec(`INSERT INTO projects 
		(name, project_type, root_dir, exec_path, port, start_command, auto_start, status, domains, ssl_enabled, ssl_email, reverse_proxy_path, extra_headers, description, use_ipv4) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.Name, p.ProjectType, p.RootDir, p.ExecPath, p.Port, p.StartCommand, p.AutoStart, "stopped", p.Domains, p.SSLEnabled, p.SSLEmail, p.ReverseProxyPath, p.ExtraHeaders, p.Description, p.UseIPv4)
	
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	projectID, _ := result.LastInsertId()
	
	// 生成 Caddyfile
	generateCaddyfileForProjects()
	caddy.Restart()
	
	// 如果设置了自动启动，启动项目
	if p.AutoStart {
		startProject(int(projectID), &p)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]int64{"id": projectID})
}

// UpdateProjectHandler 更新项目
func UpdateProjectHandler(w http.ResponseWriter, r *http.Request) {
	var p models.Project
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	db := database.GetDB()
	_, err := db.Exec(`UPDATE projects SET 
		name=?, project_type=?, root_dir=?, exec_path=?, port=?, start_command=?, auto_start=?, domains=?, ssl_enabled=?, ssl_email=?, reverse_proxy_path=?, extra_headers=?, description=?, use_ipv4=?, updated_at=CURRENT_TIMESTAMP 
		WHERE id=?`,
		p.Name, p.ProjectType, p.RootDir, p.ExecPath, p.Port, p.StartCommand, p.AutoStart, p.Domains, p.SSLEnabled, p.SSLEmail, p.ReverseProxyPath, p.ExtraHeaders, p.Description, p.UseIPv4, p.ID)
	
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	generateCaddyfileForProjects()
	caddy.Restart()

	w.WriteHeader(http.StatusOK)
}

// DeleteProjectHandler 删除项目
func DeleteProjectHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.Atoi(idStr)
	
	// 先停止项目
	stopProject(id)
	
	db := database.GetDB()
	_, err := db.Exec("DELETE FROM projects WHERE id=?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	generateCaddyfileForProjects()
	caddy.Restart()

	w.WriteHeader(http.StatusOK)
}

// StartProjectHandler 启动项目
func StartProjectHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.Atoi(idStr)

	db := database.GetDB()
	var p models.Project
	err := db.QueryRow("SELECT name, project_type, root_dir, exec_path, port, start_command FROM projects WHERE id=?", id).
		Scan(&p.Name, &p.ProjectType, &p.RootDir, &p.ExecPath, &p.Port, &p.StartCommand)

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "项目不存在",
			"code":    "PROJECT_NOT_FOUND",
		})
		return
	}
	
	p.ID = id

	// 检查管理员权限（用于绑定端口）
	if !checkAdminPrivileges() {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "权限不足",
			"code":    "ADMIN_REQUIRED",
			"details": []string{"需要管理员权限来启动项目"},
			"suggestions": []string{
				"右键点击程序图标选择'以管理员身份运行'",
				"或在 PowerShell 中以管理员身份运行: .\\caddy-manager.exe",
			},
		})
		return
	}

	// 验证项目配置
	validationErrors := validateProjectConfig(&p)
	if len(validationErrors) > 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "项目配置错误",
			"code":    "CONFIG_ERROR",
			"details": validationErrors,
		})
		return
	}
	
	// 检查端口占用
	if isPortInUse(p.Port) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "端口已被占用",
			"code":    "PORT_IN_USE",
			"details": []string{fmt.Sprintf("端口 %d 已被其他程序占用", p.Port)},
			"suggestions": []string{
				fmt.Sprintf("运行诊断工具查看端口占用: netstat -ano | findstr :%d", p.Port),
				"停止占用该端口的程序",
				"或修改项目使用其他端口",
			},
		})
		return
	}

	// 尝试启动项目
	if err := startProject(id, &p); err != nil {
		w.Header().Set("Content-Type", "application/json")
		
		// 分析错误类型
		errorCode, errorMsg, suggestions := analyzeStartError(err, &p)
		
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":     false,
			"error":       errorMsg,
			"code":        errorCode,
			"suggestions": suggestions,
			"log_path":    filepath.Join(config.DataDir, "logs", fmt.Sprintf("project_%d.log", id)),
		})
		return
	}

	// 更新状态
	db.Exec("UPDATE projects SET status='running' WHERE id=?", id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("项目 '%s' 启动成功", p.Name),
		"port":    p.Port,
	})
}

// validateProjectConfig 验证项目配置
func validateProjectConfig(p *models.Project) []string {
	errors := []string{}
	
	if p.RootDir == "" {
		errors = append(errors, "❌ 项目根目录未配置")
	} else if _, err := os.Stat(p.RootDir); os.IsNotExist(err) {
		errors = append(errors, fmt.Sprintf("❌ 项目根目录不存在: %s", p.RootDir))
	}
	
	if p.Port <= 0 || p.Port > 65535 {
		errors = append(errors, fmt.Sprintf("❌ 端口号无效: %d (应在 1-65535 之间)", p.Port))
	}
	
	hasStartConfig := false
	if p.ExecPath != "" {
		hasStartConfig = true
		if _, err := os.Stat(p.ExecPath); os.IsNotExist(err) {
			errors = append(errors, fmt.Sprintf("❌ 可执行文件不存在: %s", p.ExecPath))
		}
	}
	if p.StartCommand != "" {
		hasStartConfig = true
	}
	
	if !hasStartConfig {
		errors = append(errors, "❌ 未配置启动命令或可执行文件路径")
	}
	
	return errors
}

// analyzeStartError 分析启动错误
func analyzeStartError(err error, p *models.Project) (code string, message string, suggestions []string) {
	errMsg := err.Error()
	
	if strings.Contains(errMsg, "no such file") || strings.Contains(errMsg, "cannot find") {
		return "FILE_NOT_FOUND",
			"启动失败: 找不到可执行文件或脚本",
			[]string{
				"检查可执行文件路径: " + p.ExecPath,
				"检查启动命令: " + p.StartCommand,
				"确认文件存在于: " + p.RootDir,
			}
	}
	
	if strings.Contains(errMsg, "permission denied") {
		return "PERMISSION_DENIED",
			"启动失败: 权限不足",
			[]string{
				"以管理员身份运行 Caddy Manager",
				"检查文件权限: " + p.ExecPath,
			}
	}
	
	if strings.Contains(errMsg, "address already in use") || strings.Contains(errMsg, "bind") {
		return "PORT_IN_USE",
			fmt.Sprintf("启动失败: 端口 %d 已被占用", p.Port),
			[]string{
				"运行系统诊断检查端口占用",
				fmt.Sprintf("停止占用端口 %d 的程序", p.Port),
				"或修改项目使用其他端口",
			}
	}
	
	if p.ProjectType == "python" && strings.Contains(errMsg, "executable file not found") {
		return "PYTHON_NOT_FOUND",
			"启动失败: 未安装 Python",
			[]string{
				"访问 https://www.python.org 下载 Python",
				"安装后确保添加到环境变量",
				"运行 'python --version' 验证",
			}
	}
	
	if p.ProjectType == "nodejs" && strings.Contains(errMsg, "executable file not found") {
		return "NODEJS_NOT_FOUND",
			"启动失败: 未安装 Node.js",
			[]string{
				"访问 https://nodejs.org 下载 Node.js",
				"安装后确保添加到环境变量",
				"运行 'node --version' 验证",
			}
	}
	
	if p.ProjectType == "java" && strings.Contains(errMsg, "executable file not found") {
		return "JAVA_NOT_FOUND",
			"启动失败: 未安装 Java",
			[]string{
				"下载并安装 JRE 或 JDK",
				"安装后确保添加到环境变量",
				"运行 'java -version' 验证",
			}
	}
	
	return "START_FAILED",
		"启动失败: " + errMsg,
		[]string{
			"查看日志: data/logs/project_" + strconv.Itoa(p.ID) + ".log",
			"检查项目配置是否正确",
			"尝试手动启动获取更多信息",
		}
}

// StopProjectHandler 停止项目
func StopProjectHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.Atoi(idStr)
	
	if err := stopProject(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	db := database.GetDB()
	db.Exec("UPDATE projects SET status='stopped' WHERE id=?", id)

	w.WriteHeader(http.StatusOK)
}

// RestartProjectHandler 重启项目
func RestartProjectHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.Atoi(idStr)
	
	// 先停止
	stopProject(id)
	
	db := database.GetDB()
	var p models.Project
	err := db.QueryRow("SELECT name, project_type, root_dir, exec_path, port, start_command FROM projects WHERE id=?", id).
		Scan(&p.Name, &p.ProjectType, &p.RootDir, &p.ExecPath, &p.Port, &p.StartCommand)
	
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "项目不存在",
		})
		return
	}
	
	p.ID = id

	// 重新启动
	if err := startProject(id, &p); err != nil {
		w.Header().Set("Content-Type", "application/json")
		errorCode, errorMsg, suggestions := analyzeStartError(err, &p)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":     false,
			"error":       errorMsg,
			"code":        errorCode,
			"suggestions": suggestions,
		})
		return
	}

	db.Exec("UPDATE projects SET status='running' WHERE id=?", id)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("项目 '%s' 重启成功", p.Name),
	})
}

// GetProjectLogsHandler 获取项目日志
func GetProjectLogsHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	logPath := filepath.Join(config.DataDir, "logs", fmt.Sprintf("project_%s.log", idStr))
	
	content := "暂无日志"
	if data, err := os.ReadFile(logPath); err == nil {
		lines := strings.Split(string(data), "\n")
		if len(lines) > 100 {
			lines = lines[len(lines)-100:]
		}
		content = strings.Join(lines, "\n")
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"logs": content})
}

// GetProjectStatusHandler 获取单个项目状态
func GetProjectStatusHandler(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.Atoi(idStr)
	
	db := database.GetDB()
	var port int
	err := db.QueryRow("SELECT port FROM projects WHERE id=?", id).Scan(&port)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	status := getProjectStatus(id, port)
	
	// 更新数据库中的状态
	db.Exec("UPDATE projects SET status=? WHERE id=?", status, id)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": status})
}

// StopAllProjects 停止所有运行中的项目
func StopAllProjects() {
	processMutex.Lock()
	defer processMutex.Unlock()
	
	for id, cmd := range projectProcesses {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		
		// 更新数据库状态
		db := database.GetDB()
		db.Exec("UPDATE projects SET status='stopped' WHERE id=?", id)
	}
	
	// 清空进程映射
	projectProcesses = make(map[int]*exec.Cmd)
}

// 内部函数
func startProject(id int, p *models.Project) error {
	processMutex.Lock()
	defer processMutex.Unlock()

	// 如果已经在运行，先停止
	if cmd, exists := projectProcesses[id]; exists {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		delete(projectProcesses, id)
	}

	// 创建日志目录
	logDir := filepath.Join(config.DataDir, "logs")
	os.MkdirAll(logDir, 0755)
	
	logPath := filepath.Join(logDir, fmt.Sprintf("project_%d.log", id))
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	var cmd *exec.Cmd
	
	switch p.ProjectType {
	case "go":
		if p.ExecPath != "" {
			cmd = exec.Command(p.ExecPath)
		} else if p.StartCommand != "" {
			parts := strings.Fields(p.StartCommand)
			cmd = exec.Command(parts[0], parts[1:]...)
		}
	case "python":
		if p.StartCommand != "" {
			parts := strings.Fields(p.StartCommand)
			cmd = exec.Command("python", parts...)
		}
	case "nodejs":
		if p.StartCommand != "" {
			parts := strings.Fields(p.StartCommand)
			cmd = exec.Command("node", parts...)
		}
	case "java":
		if p.StartCommand != "" {
			parts := strings.Fields(p.StartCommand)
			cmd = exec.Command("java", parts...)
		}
	default:
		if p.StartCommand != "" {
			parts := strings.Fields(p.StartCommand)
			cmd = exec.Command(parts[0], parts[1:]...)
		}
	}

	if cmd == nil {
		return fmt.Errorf("无法启动项目：未配置启动命令")
	}

	cmd.Dir = p.RootDir
	cmd.Stdout = logFile
	cmd.Stderr = logFile

	if err := cmd.Start(); err != nil {
		logFile.Close()
		return err
	}

	projectProcesses[id] = cmd
	
	// 后台监控进程
	go func() {
		cmd.Wait()
		logFile.Close()
		processMutex.Lock()
		delete(projectProcesses, id)
		processMutex.Unlock()
		
		// 更新状态
		db := database.GetDB()
		db.Exec("UPDATE projects SET status='stopped' WHERE id=?", id)
	}()

	return nil
}

func stopProject(id int) error {
	processMutex.Lock()
	defer processMutex.Unlock()

	if cmd, exists := projectProcesses[id]; exists {
		if cmd.Process != nil {
			cmd.Process.Kill()
		}
		delete(projectProcesses, id)
	}

	return nil
}

func getProjectStatus(id int, port int) string {
	processMutex.RLock()
	_, exists := projectProcesses[id]
	processMutex.RUnlock()

	if exists {
		return "running"
	}
	
	// 通过端口检测（Windows）
	if port > 0 {
		cmd := exec.Command("netstat", "-ano")
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
		output, err := cmd.Output()
		if err == nil {
			lines := strings.Split(string(output), "\n")
			portStr := fmt.Sprintf(":%d", port)
			for _, line := range lines {
				if strings.Contains(line, portStr) && strings.Contains(line, "LISTENING") {
					return "running"
				}
			}
		}
	}
	
	return "stopped"
}

func checkAdminPrivileges() bool {
	cmd := exec.Command("net", "session")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	err := cmd.Run()
	return err == nil
}

func isPortInUse(port int) bool {
	cmd := exec.Command("netstat", "-ano")
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	
	lines := strings.Split(string(output), "\n")
	portStr := fmt.Sprintf(":%d", port)
	for _, line := range lines {
		if strings.Contains(line, portStr) && strings.Contains(line, "LISTENING") {
			return true
		}
	}
	
	return false
}

func generateCaddyfileForProjects() error {
	db := database.GetDB()
	rows, err := db.Query("SELECT domains, port, ssl_enabled, reverse_proxy_path, extra_headers, COALESCE(use_ipv4, 1) FROM projects WHERE domains != ''")
	if err != nil {
		return err
	}
	defer rows.Close()

	var content string

	// 添加注释头
	content += "# Caddy 配置文件\n"
	content += "# 由 Caddy 管理器自动生成\n\n"

	hasProjects := false

	for rows.Next() {
		var domains string
		var port int
		var sslEnabled bool
		var useIPv4 bool
		var reverseProxyPath, extraHeaders *string

		rows.Scan(&domains, &port, &sslEnabled, &reverseProxyPath, &extraHeaders, &useIPv4)

		domainList := strings.Split(domains, "\n")
		for _, domain := range domainList {
			domain = strings.TrimSpace(domain)
			if domain == "" {
				continue
			}

			// 验证域名格式
			if !isValidDomain(domain) {
				continue
			}

			hasProjects = true
			content += domain + " {\n"

			// 反向代理 - 根据 use_ipv4 设置决定使用 IPv4 或 localhost
			var proxyTarget string
			if useIPv4 {
				// 强制使用 IPv4 地址，避免 IPv6 连接问题
				proxyTarget = fmt.Sprintf("127.0.0.1:%d", port)
			} else {
				// 使用 localhost（可能使用 IPv6）
				proxyTarget = fmt.Sprintf("localhost:%d", port)
			}
			
			content += fmt.Sprintf("    reverse_proxy %s\n", proxyTarget)

			// 额外 Header
			if extraHeaders != nil && *extraHeaders != "" {
				headers := strings.Split(*extraHeaders, "\n")
				for _, header := range headers {
					if strings.TrimSpace(header) != "" {
						content += "    header_up " + header + "\n"
					}
				}
			}

			content += "}\n\n"
		}
	}

	// 如果没有项目，添加默认配置
	if !hasProjects {
		content += "# 暂无项目配置\n"
		content += "# 通过管理界面添加项目后会自动生成配置\n\n"
		content += ":80 {\n"
		content += "    respond \"Caddy 正在运行\" 200\n"
		content += "}\n"
	}

	return os.WriteFile(config.CaddyConfig, []byte(content), 0644)
}

// 验证域名格式
func isValidDomain(domain string) bool {
	// 移除端口号（如果有）
	if idx := strings.Index(domain, ":"); idx > 0 {
		domain = domain[:idx]
	}
	
	// 检查长度
	if len(domain) == 0 || len(domain) > 253 {
		return false
	}
	
	// 检查是否包含空格或其他非法字符
	if strings.Contains(domain, " ") || strings.Contains(domain, "\t") {
		return false
	}
	
	// 简单的域名格式检查
	parts := strings.Split(domain, ".")
	
	// 至少要有一个点（如 example.com）
	if len(parts) < 2 {
		// 除非是 localhost 或 IP 地址
		if domain != "localhost" && !strings.Contains(domain, ":") {
			// 检查是否是合法的单字符标签（用于测试）
			if len(domain) == 0 {
				return false
			}
		}
	}
	
	// 检查每个部分
	for _, part := range parts {
		if len(part) == 0 || len(part) > 63 {
			return false
		}
		
		// 每个部分只能包含字母、数字、连字符
		for i, c := range part {
			if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || 
				 (c >= '0' && c <= '9') || c == '-' || c == '_') {
				return false
			}
			// 不能以连字符开头或结尾
			if c == '-' && (i == 0 || i == len(part)-1) {
				return false
			}
		}
	}
	
	return true
}
