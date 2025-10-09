package api

import (
	"net/http"
	"encoding/json"
	
	"caddy-manager/internal/auth"
)

// CheckAuthHandler 检查是否已登录
func CheckAuthHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_id")
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"authenticated": false})
		return
	}
	
	_, exists := auth.GetSession(cookie.Value)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"authenticated": exists})
}
