package api

import (
	"encoding/json"
	"net/http"
	"os"

	"caddy-manager/internal/auth"
	"caddy-manager/internal/caddy"
	"caddy-manager/internal/database"
)

// GetSettingsHandler 获取设置
func GetSettingsHandler(w http.ResponseWriter, r *http.Request) {
	db := database.GetDB()
	
	var securityPath, wwwRoot string
	db.QueryRow("SELECT value FROM settings WHERE key = 'security_path'").Scan(&securityPath)
	db.QueryRow("SELECT value FROM settings WHERE key = 'www_root'").Scan(&wwwRoot)
	
	settings := map[string]string{
		"security_path": securityPath,
		"www_root":      wwwRoot,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settings)
}

// UpdateSettingsHandler 更新设置
func UpdateSettingsHandler(w http.ResponseWriter, r *http.Request) {
	var req map[string]string
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	db := database.GetDB()
	
	// 更新安全路径
	if securityPath, ok := req["security_path"]; ok {
		_, err := db.Exec("UPDATE settings SET value = ?, updated_at = CURRENT_TIMESTAMP WHERE key = 'security_path'", securityPath)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	
	// 更新 www 根目录
	if wwwRoot, ok := req["www_root"]; ok {
		_, err := db.Exec("UPDATE settings SET value = ?, updated_at = CURRENT_TIMESTAMP WHERE key = 'www_root'", wwwRoot)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// 创建目录
		os.MkdirAll(wwwRoot, 0755)
	}
	
	w.WriteHeader(http.StatusOK)
}

// ChangePasswordHandler 修改密码
func ChangePasswordHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username    string `json:"username"`
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	db := database.GetDB()
	
	// 验证旧密码
	var hash string
	err := db.QueryRow("SELECT password FROM users WHERE username = ?", req.Username).Scan(&hash)
	if err != nil || !auth.CheckPassword(req.OldPassword, hash) {
		http.Error(w, "旧密码错误", http.StatusUnauthorized)
		return
	}
	
	// 更新密码
	newHash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	_, err = db.Exec("UPDATE users SET password = ? WHERE username = ?", newHash, req.Username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	
	w.WriteHeader(http.StatusOK)
}

// CaddyLogsHandler 获取 Caddy 日志
func CaddyLogsHandler(w http.ResponseWriter, r *http.Request) {
	logs, err := caddy.GetLogs(100)
	if err != nil {
		http.Error(w, "无法读取日志", http.StatusInternalServerError)
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"logs": logs})
}
