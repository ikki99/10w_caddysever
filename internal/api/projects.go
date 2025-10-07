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
	rows, err := db.Query("SELECT id, name, project_type, root_dir, exec_path, port, start_command, auto_start, status, domains, ssl_enabled, description FROM projects ORDER BY created_at DESC")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var projects []models.Project
	for rows.Next() {
		var p models.Project
		var execPath, startCmd, domains, desc *string
		if err := rows.Scan(&p.ID, &p.Name, &p.ProjectType, &p.RootDir, &execPath, &p.Port, &startCmd, &p.AutoStart, &p.Status, &domains, &p.SSLEnabled, &desc); err != nil {
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

	db := database.GetDB()
	result, err := db.Exec(`INSERT INTO projects 
		(name, project_type, root_dir, exec_path, port, start_command, auto_start, status, domains, ssl_enabled, ssl_email, reverse_proxy_path, extra_headers, description) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.Name, p.ProjectType, p.RootDir, p.ExecPath, p.Port, p.StartCommand, p.AutoStart, "stopped", p.Domains, p.SSLEnabled, p.SSLEmail, p.ReverseProxyPath, p.ExtraHeaders, p.Description)
	
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
		name=?, project_type=?, root_dir=?, exec_path=?, port=?, start_command=?, auto_start=?, domains=?, ssl_enabled=?, ssl_email=?, reverse_proxy_path=?, extra_headers=?, description=?, updated_at=CURRENT_TIMESTAMP 
		WHERE id=?`,
		p.Name, p.ProjectType, p.RootDir, p.ExecPath, p.Port, p.StartCommand, p.AutoStart, p.Domains, p.SSLEnabled, p.SSLEmail, p.ReverseProxyPath, p.ExtraHeaders, p.Description, p.ID)
	
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
	err := db.QueryRow("SELECT project_type, root_dir, exec_path, port, start_command FROM projects WHERE id=?", id).
		Scan(&p.ProjectType, &p.RootDir, &p.ExecPath, &p.Port, &p.StartCommand)
	
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := startProject(id, &p); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 更新状态
	db.Exec("UPDATE projects SET status='running' WHERE id=?", id)

	w.WriteHeader(http.StatusOK)
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
	
	stopProject(id)
	
	db := database.GetDB()
	var p models.Project
	err := db.QueryRow("SELECT project_type, root_dir, exec_path, port, start_command FROM projects WHERE id=?", id).
		Scan(&p.ProjectType, &p.RootDir, &p.ExecPath, &p.Port, &p.StartCommand)
	
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := startProject(id, &p); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	db.Exec("UPDATE projects SET status='running' WHERE id=?", id)

	w.WriteHeader(http.StatusOK)
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
	
	// 通过端口检测
	if port > 0 {
		cmd := exec.Command("netstat", "-ano", "|", "findstr", fmt.Sprintf(":%d", port))
		if output, err := cmd.Output(); err == nil && len(output) > 0 {
			return "running"
		}
	}
	
	return "stopped"
}

func generateCaddyfileForProjects() error {
	db := database.GetDB()
	rows, err := db.Query("SELECT domains, port, ssl_enabled, reverse_proxy_path, extra_headers FROM projects WHERE domains != ''")
	if err != nil {
		return err
	}
	defer rows.Close()

	var content string
	for rows.Next() {
		var domains string
		var port int
		var sslEnabled bool
		var reverseProxyPath, extraHeaders *string
		
		rows.Scan(&domains, &port, &sslEnabled, &reverseProxyPath, &extraHeaders)

		domainList := strings.Split(domains, "\n")
		for _, domain := range domainList {
			domain = strings.TrimSpace(domain)
			if domain == "" {
				continue
			}

			content += domain + " {\n"
			
			// 反向代理路径
			proxyPath := "/"
			if reverseProxyPath != nil && *reverseProxyPath != "" {
				proxyPath = *reverseProxyPath
			}
			
			content += fmt.Sprintf("    reverse_proxy %s localhost:%d\n", proxyPath, port)
			
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

	return os.WriteFile(config.CaddyConfig, []byte(content), 0644)
}
