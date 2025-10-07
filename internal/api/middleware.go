package api

import (
	"caddy-manager/internal/database"
	"net/http"
	"strings"
)

// SecurityPathMiddleware 安全路径中间件
func SecurityPathMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 获取安全路径设置
		db := database.GetDB()
		var securityPath string
		db.QueryRow("SELECT value FROM settings WHERE key = 'security_path'").Scan(&securityPath)

		// 如果设置了安全路径
		if securityPath != "" && securityPath != "/" {
			path := r.URL.Path
			
			// API 和静态资源不受安全路径限制
			if strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/static/") {
				next.ServeHTTP(w, r)
				return
			}
			
			// 检查是否以安全路径开头
			if strings.HasPrefix(path, "/"+securityPath+"/") || path == "/"+securityPath {
				// 移除安全路径前缀
				newPath := strings.TrimPrefix(path, "/"+securityPath)
				if newPath == "" {
					newPath = "/"
				}
				r.URL.Path = newPath
				next.ServeHTTP(w, r)
			} else {
				// 不是安全路径，返回403错误页面
				w.WriteHeader(http.StatusForbidden)
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.Write([]byte(`<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>访问被拒绝</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Microsoft YaHei", sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            display: flex;
            align-items: center;
            justify-content: center;
            min-height: 100vh;
            color: #fff;
        }
        .container {
            text-align: center;
            max-width: 500px;
            padding: 40px;
            background: rgba(255, 255, 255, 0.1);
            backdrop-filter: blur(10px);
            border-radius: 20px;
            box-shadow: 0 8px 32px rgba(0, 0, 0, 0.3);
        }
        .error-code {
            font-size: 100px;
            font-weight: bold;
            margin-bottom: 20px;
            text-shadow: 2px 2px 4px rgba(0, 0, 0, 0.3);
        }
        h1 {
            font-size: 32px;
            margin-bottom: 15px;
        }
        p {
            font-size: 16px;
            margin-bottom: 30px;
            opacity: 0.9;
        }
        .icon {
            font-size: 80px;
            margin-bottom: 20px;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="icon">🔒</div>
        <div class="error-code">403</div>
        <h1>访问被拒绝</h1>
        <p>您访问的路径受安全保护，请使用正确的访问地址。</p>
    </div>
</body>
</html>`))
				return
			}
		} else {
			// 没有设置安全路径，直接放行
			next.ServeHTTP(w, r)
		}
	})
}
