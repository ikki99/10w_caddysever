package api

import (
	"encoding/json"
	"net/http"
	
	"caddy-manager/internal/system"
)

// SystemMonitorHandler 系统监控 API
func SystemMonitorHandler(w http.ResponseWriter, r *http.Request) {
	stats := system.GetSystemStats()
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
