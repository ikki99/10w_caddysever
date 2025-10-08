package api

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"caddy-manager/internal/auth"
	"caddy-manager/internal/caddy"
	"caddy-manager/internal/config"
	"caddy-manager/internal/database"
	"caddy-manager/internal/models"
)

func IndexHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(IndexTemplate))
}

func SitesHandler(w http.ResponseWriter, r *http.Request) {
	db := database.GetDB()
	rows, err := db.Query("SELECT id, domain, type, target, ssl_enabled, environment, php_version FROM sites ORDER BY created_at DESC")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var sites []models.Site
	for rows.Next() {
		var site models.Site
		var env, phpVer *string
		if err := rows.Scan(&site.ID, &site.Domain, &site.Type, &site.Target, &site.SSLEnabled, &env, &phpVer); err != nil {
			continue
		}
		if env != nil {
			site.Environment = *env
		}
		if phpVer != nil {
			site.PHPVersion = *phpVer
		}
		sites = append(sites, site)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sites)
}

func AddSiteHandler(w http.ResponseWriter, r *http.Request) {
	var site models.Site
	if err := json.NewDecoder(r.Body).Decode(&site); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	db := database.GetDB()
	_, err := db.Exec("INSERT INTO sites (domain, type, target, ssl_enabled, environment, php_version) VALUES (?, ?, ?, ?, ?, ?)",
		site.Domain, site.Type, site.Target, site.SSLEnabled, site.Environment, site.PHPVersion)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	generateCaddyfile()
	caddy.Restart()

	w.WriteHeader(http.StatusOK)
}

func EditSiteHandler(w http.ResponseWriter, r *http.Request) {
	var site models.Site
	if err := json.NewDecoder(r.Body).Decode(&site); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	db := database.GetDB()
	_, err := db.Exec("UPDATE sites SET domain=?, type=?, target=?, ssl_enabled=?, environment=?, php_version=?, updated_at=CURRENT_TIMESTAMP WHERE id=?",
		site.Domain, site.Type, site.Target, site.SSLEnabled, site.Environment, site.PHPVersion, site.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	generateCaddyfile()
	caddy.Restart()

	w.WriteHeader(http.StatusOK)
}

func DeleteSiteHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	db := database.GetDB()
	_, err := db.Exec("DELETE FROM sites WHERE id=?", id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	generateCaddyfile()
	caddy.Restart()

	w.WriteHeader(http.StatusOK)
}

func CaddyStatusHandler(w http.ResponseWriter, r *http.Request) {
	running := caddy.IsRunning()
	version := caddy.GetVersion()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"running": running,
		"version": version,
	})
}

func CaddySSLStatusHandler(w http.ResponseWriter, r *http.Request) {
	errors := []string{}
	
	// 读取 Caddy 日志检查 SSL 错误
	logPath := config.CaddyLogFile
	if data, err := os.ReadFile(logPath); err == nil {
		logContent := string(data)
		
		// 检查常见的 SSL 错误
		if contains(logContent, "acme") && contains(logContent, "error") {
			errors = append(errors, "ACME证书申请失败")
		}
		if contains(logContent, "dns") && contains(logContent, "error") {
			errors = append(errors, "DNS验证失败，请检查域名解析")
		}
		if contains(logContent, "timeout") {
			errors = append(errors, "连接超时，请检查网络和防火墙设置")
		}
		if contains(logContent, "rate limit") {
			errors = append(errors, "证书申请频率限制，请稍后再试")
		}
		if contains(logContent, "unauthorized") {
			errors = append(errors, "域名验证失败，请确认域名已正确解析")
		}
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"errors": errors,
		"hasErrors": len(errors) > 0,
	})
}

func contains(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && 
		(len(s) >= len(substr)) && 
		(s == substr || len(s) > len(substr) && 
		func() bool {
			for i := 0; i <= len(s)-len(substr); i++ {
				if s[i:i+len(substr)] == substr {
					return true
				}
			}
			return false
		}())
}

func CaddyStartHandler(w http.ResponseWriter, r *http.Request) {
	if caddy.IsRunning() {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Caddy 已在运行中",
		})
		return
	}
	
	if err := caddy.Start(); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Caddy 启动成功",
	})
}

func CaddyStopHandler(w http.ResponseWriter, r *http.Request) {
	if !caddy.IsRunning() {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   "Caddy 未运行",
		})
		return
	}
	
	caddy.Stop()
	
	// 等待一小段时间确保停止
	time.Sleep(500 * time.Millisecond)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Caddy 已停止",
	})
}

func CaddyRestartHandler(w http.ResponseWriter, r *http.Request) {
	if err := caddy.Restart(); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Caddy 重启成功",
	})
}

func CaddyReloadHandler(w http.ResponseWriter, r *http.Request) {
	if err := caddy.Reload(); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Caddy 配置已重新加载（零停机）",
	})
}

func SetupHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		isFirst := database.IsFirstRun()
		json.NewEncoder(w).Encode(map[string]bool{"firstRun": isFirst})
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	db := database.GetDB()
	_, err = db.Exec("INSERT INTO users (username, password) VALUES (?, ?)", req.Username, hash)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	db := database.GetDB()
	var hash string
	err := db.QueryRow("SELECT password FROM users WHERE username = ?", req.Username).Scan(&hash)
	if err != nil || !auth.CheckPassword(req.Password, hash) {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	sessionID, err := auth.CreateSession(req.Username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    sessionID,
		MaxAge:   86400,
		HttpOnly: true,
		Path:     "/",
	})

	w.WriteHeader(http.StatusOK)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err == nil {
		auth.DeleteSession(cookie.Value)
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		MaxAge:   -1,
		HttpOnly: true,
		Path:     "/",
	})

	w.WriteHeader(http.StatusOK)
}

func EnvListHandler(w http.ResponseWriter, r *http.Request) {
	envs := []map[string]string{
		{"name": "Python", "status": "未检测"},
		{"name": "Node.js", "status": "未检测"},
		{"name": "Java", "status": "未检测"},
		{"name": "Go", "status": "未检测"},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(envs)
}

func EnvInstallHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func ShutdownHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "应用程序正在关闭..."})
	
	// 在单独的 goroutine 中执行关闭操作
	go func() {
		// 等待响应发送完成
		time.Sleep(1 * time.Second)
		
		// 发送中断信号触发优雅关闭
		p, err := os.FindProcess(os.Getpid())
		if err == nil {
			p.Signal(os.Interrupt)
		}
	}()
}

func generateCaddyfile() error {
	db := database.GetDB()
	rows, err := db.Query("SELECT domain, type, target, ssl_enabled, environment, php_version FROM sites")
	if err != nil {
		return err
	}
	defer rows.Close()

	var content string
	for rows.Next() {
		var domain, siteType, target string
		var environment, phpVersion *string
		var sslEnabled bool
		rows.Scan(&domain, &siteType, &target, &sslEnabled, &environment, &phpVersion)

		if siteType == "proxy" {
			content += domain + " {\n"
			content += "    reverse_proxy " + target + "\n"
			content += "}\n\n"
		} else if siteType == "static" {
			content += domain + " {\n"
			content += "    root * " + target + "\n"
			content += "    file_server\n"
			content += "}\n\n"
		} else if siteType == "php" {
			content += domain + " {\n"
			content += "    root * " + target + "\n"
			content += "    php_fastcgi localhost:9000\n"
			content += "    file_server\n"
			content += "}\n\n"
		}
	}

	return os.WriteFile(config.CaddyConfig, []byte(content), 0644)
}
