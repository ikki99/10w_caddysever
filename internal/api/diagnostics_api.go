package api

import (
	"encoding/json"
	"net/http"

	"caddy-manager/internal/diagnostics"
	"caddy-manager/internal/system"
)

// DiagnosticsHandler 运行系统诊断
func DiagnosticsHandler(w http.ResponseWriter, r *http.Request) {
	result := diagnostics.RunDiagnostics()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// CheckSSLHandler 检查 SSL 问题
func CheckSSLHandler(w http.ResponseWriter, r *http.Request) {
	domain := r.URL.Query().Get("domain")
	if domain == "" {
		http.Error(w, "domain parameter required", http.StatusBadRequest)
		return
	}
	
	issues := diagnostics.CheckSSLIssues(domain)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"domain": domain,
		"issues": issues,
	})
}

// AutoFixHandler 自动修复问题
func AutoFixHandler(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IssueCode string `json:"issue_code"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	
	err := diagnostics.AutoFix(req.IssueCode)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		})
		return
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "问题已修复",
	})
}

// SystemStatusHandler 获取系统状态
func SystemStatusHandler(w http.ResponseWriter, r *http.Request) {
	isAdmin := system.IsAdmin()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"is_admin": isAdmin,
	})
}
