package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"caddy-manager/internal/api"
	"caddy-manager/internal/auth"
	"caddy-manager/internal/caddy"
	"caddy-manager/internal/config"
	"caddy-manager/internal/database"
	"caddy-manager/internal/tray"
)

func main() {
	port := flag.Int("port", 8989, "Web UI 端口")
	noTray := flag.Bool("no-tray", false, "禁用系统托盘")
	flag.Parse()

	// 初始化配置
	if err := config.Init(); err != nil {
		log.Fatalf("配置初始化失败: %v", err)
	}

	// 初始化数据库
	if err := database.Init(); err != nil {
		log.Fatalf("数据库初始化失败: %v", err)
	}
	defer database.Close()

	// 检查并下载 Caddy
	if err := caddy.CheckAndDownload(); err != nil {
		log.Printf("⚠️  Caddy 自动下载失败: %v", err)
		log.Println("💡 请手动下载 Caddy 到:", config.CaddyBin)
	} else {
		// 自动启动 Caddy
		go caddy.AutoStart()
	}

	// 检查是否首次运行
	if database.IsFirstRun() {
		fmt.Println("============================================================")
		fmt.Println("🎉 欢迎使用 Caddy 管理器")
		fmt.Println("============================================================")
		fmt.Printf("\n请访问: http://localhost:%d\n", *port)
		fmt.Println("首次运行将引导您完成初始化设置")
		fmt.Println()
	}

	// 设置路由
	mux := http.NewServeMux()

	// API 路由
	mux.HandleFunc("/api/setup", api.SetupHandler)
	mux.HandleFunc("/api/login", api.LoginHandler)
	mux.HandleFunc("/api/logout", api.LogoutHandler)
	mux.HandleFunc("/api/system/info", api.GetSystemInfoHandler)
	
	// 需要认证的路由
	mux.HandleFunc("/api/sites", auth.AuthMiddleware(api.SitesHandler))
	mux.HandleFunc("/api/sites/add", auth.AuthMiddleware(api.AddSiteHandler))
	mux.HandleFunc("/api/sites/edit", auth.AuthMiddleware(api.EditSiteHandler))
	mux.HandleFunc("/api/sites/delete", auth.AuthMiddleware(api.DeleteSiteHandler))
	mux.HandleFunc("/api/caddy/status", auth.AuthMiddleware(api.CaddyStatusHandler))
	mux.HandleFunc("/api/caddy/restart", auth.AuthMiddleware(api.CaddyRestartHandler))
	mux.HandleFunc("/api/caddy/logs", auth.AuthMiddleware(api.CaddyLogsHandler))
	mux.HandleFunc("/api/files/browse", auth.AuthMiddleware(api.BrowseFilesHandler))
	mux.HandleFunc("/api/files/upload", auth.AuthMiddleware(api.UploadFileHandler))
	mux.HandleFunc("/api/files/download", auth.AuthMiddleware(api.DownloadFileHandler))
	mux.HandleFunc("/api/files/delete", auth.AuthMiddleware(api.DeleteFileHandler))
	mux.HandleFunc("/api/files/create-folder", auth.AuthMiddleware(api.CreateFolderHandler))
	mux.HandleFunc("/api/env/list", auth.AuthMiddleware(api.EnvListHandler))
	mux.HandleFunc("/api/env/install", auth.AuthMiddleware(api.EnvInstallHandler))
	mux.HandleFunc("/api/env/guide", auth.AuthMiddleware(api.InstallEnvGuideHandler))
	mux.HandleFunc("/api/settings/get", auth.AuthMiddleware(api.GetSettingsHandler))
	mux.HandleFunc("/api/settings/update", auth.AuthMiddleware(api.UpdateSettingsHandler))
	mux.HandleFunc("/api/user/password", auth.AuthMiddleware(api.ChangePasswordHandler))

	// 项目管理
	mux.HandleFunc("/api/projects", auth.AuthMiddleware(api.ProjectsHandler))
	mux.HandleFunc("/api/projects/add", auth.AuthMiddleware(api.AddProjectHandler))
	mux.HandleFunc("/api/projects/update", auth.AuthMiddleware(api.UpdateProjectHandler))
	mux.HandleFunc("/api/projects/delete", auth.AuthMiddleware(api.DeleteProjectHandler))
	mux.HandleFunc("/api/projects/start", auth.AuthMiddleware(api.StartProjectHandler))
	mux.HandleFunc("/api/projects/stop", auth.AuthMiddleware(api.StopProjectHandler))
	mux.HandleFunc("/api/projects/restart", auth.AuthMiddleware(api.RestartProjectHandler))
	mux.HandleFunc("/api/projects/logs", auth.AuthMiddleware(api.GetProjectLogsHandler))
	
	// 任务管理
	mux.HandleFunc("/api/tasks", auth.AuthMiddleware(api.TasksHandler))
	mux.HandleFunc("/api/tasks/add", auth.AuthMiddleware(api.AddTaskHandler))
	mux.HandleFunc("/api/tasks/delete", auth.AuthMiddleware(api.DeleteTaskHandler))
	mux.HandleFunc("/api/tasks/execute", auth.AuthMiddleware(api.ExecuteTaskHandler))

	// 静态文件
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	// 前端页面
	mux.HandleFunc("/", api.IndexHandler)

	// 使用安全路径中间件
	handler := api.SecurityPathMiddleware(mux)

	addr := fmt.Sprintf(":%d", *port)
	
	fmt.Println("============================================================")
	fmt.Println("                  Caddy 管理器")
	fmt.Println("============================================================")
	fmt.Printf("\n🌐 访问地址: http://localhost:%d\n", *port)
	fmt.Println("📊 按 Ctrl+C 停止服务")
	fmt.Println()
	
	// 启动系统托盘（在单独的 goroutine 中）
	if !*noTray {
		go tray.Run(*port)
	}
	
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("服务启动失败: %v", err)
	}
}
