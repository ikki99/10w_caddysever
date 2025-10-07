package api

import (
	"encoding/json"
	"net/http"
	"os"

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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{"running": running})
}

func CaddyRestartHandler(w http.ResponseWriter, r *http.Request) {
	caddy.Restart()
	w.WriteHeader(http.StatusOK)
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
