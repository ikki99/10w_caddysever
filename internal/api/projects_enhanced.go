package api

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	
	"caddy-manager/internal/caddy"
	"caddy-manager/internal/database"
	"caddy-manager/internal/models"
)

// AddProjectHandlerV2 增强版添加项目处理器
func AddProjectHandlerV2(w http.ResponseWriter, r *http.Request) {
	var p models.Project
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		sendJSONResponse(w, false, "请求数据格式错误", map[string]interface{}{
			"details": err.Error(),
		})
		return
	}

	// 验证配置
	validationErrors := validateProjectConfig(&p)
	if len(validationErrors) > 0 {
		sendJSONResponse(w, false, "项目配置验证失败", map[string]interface{}{
			"details": validationErrors,
		})
		return
	}

	db := database.GetDB()
	result, err := db.Exec(`INSERT INTO projects
		(name, project_type, root_dir, exec_path, port, start_command, auto_start, status, domains, ssl_enabled, ssl_email, reverse_proxy_path, extra_headers, description)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.Name, p.ProjectType, p.RootDir, p.ExecPath, p.Port, p.StartCommand, p.AutoStart, "stopped", p.Domains, p.SSLEnabled, p.SSLEmail, p.ReverseProxyPath, p.ExtraHeaders, p.Description)

	if err != nil {
		sendJSONResponse(w, false, "数据库保存失败", map[string]interface{}{
			"details": err.Error(),
		})
		return
	}

	projectID, _ := result.LastInsertId()
	p.ID = int(projectID)

	// 生成 Caddyfile
	if err := generateCaddyfileForProjects(); err != nil {
		sendJSONResponse(w, true, "项目已创建，但 Caddy 配置可能有问题", map[string]interface{}{
			"warning": "Caddyfile 生成失败",
			"details": err.Error(),
		})
		return
	}

	// 重启 Caddy
	if err := caddy.Restart(); err != nil {
		sendJSONResponse(w, true, "项目已创建，但 Caddy 未能重启", map[string]interface{}{
			"warning": "Caddy 重启失败",
			"details": err.Error(),
			"suggestions": []string{
				"请手动重启 Caddy 服务",
				"或查看 Caddy 日志了解详情",
			},
		})
		return
	}

	// SSL 检查
	sslWarnings := []string{}
	if p.SSLEnabled && p.Domains != "" {
		if !checkAdminPrivileges() {
			sslWarnings = append(sslWarnings, "⚠ 未以管理员身份运行，无法绑定 443 端口")
		}
		
		domains := strings.Split(p.Domains, "\n")
		for _, domain := range domains {
			domain = strings.TrimSpace(domain)
			if domain != "" {
				warnings := checkDomainForSSL(domain)
				sslWarnings = append(sslWarnings, warnings...)
			}
		}
	}

	// 自动启动
	startMessage := ""
	if p.AutoStart {
		if err := startProject(int(projectID), &p); err != nil {
			startMessage = "⚠ 自动启动失败: " + err.Error()
		} else {
			startMessage = "✓ 项目已自动启动"
		}
	}

	response := map[string]interface{}{
		"success":    true,
		"message":    fmt.Sprintf("✓ 项目 '%s' 创建成功", p.Name),
		"project_id": projectID,
	}
	
	if startMessage != "" {
		response["start_message"] = startMessage
	}
	
	if len(sslWarnings) > 0 {
		response["ssl_warnings"] = sslWarnings
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func checkDomainForSSL(domain string) []string {
	warnings := []string{}
	
	ips, err := net.LookupIP(domain)
	if err != nil {
		warnings = append(warnings, fmt.Sprintf("❌ 域名 %s 无法解析", domain))
		return warnings
	}
	
	for _, ip := range ips {
		ipStr := ip.String()
		if strings.HasPrefix(ipStr, "104.21.") || strings.HasPrefix(ipStr, "172.67.") || strings.HasPrefix(ipStr, "104.18.") {
			warnings = append(warnings, fmt.Sprintf("⚠ 域名 %s 使用 Cloudflare CDN，建议:\n  1. 使用 Flexible SSL 模式\n  2. 或临时关闭 Cloudflare 代理申请证书", domain))
			break
		}
	}
	
	return warnings
}

func sendJSONResponse(w http.ResponseWriter, success bool, message string, extra map[string]interface{}) {
	response := map[string]interface{}{
		"success": success,
		"message": message,
	}
	
	for k, v := range extra {
		response[k] = v
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
