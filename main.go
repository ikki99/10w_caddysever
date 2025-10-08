package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"caddy-manager/internal/api"
	"caddy-manager/internal/auth"
	"caddy-manager/internal/caddy"
	"caddy-manager/internal/config"
	"caddy-manager/internal/database"
	"caddy-manager/internal/system"
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
	
	// 检查管理员权限
	checkAdminPrivileges()

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
	mux.HandleFunc("/api/caddy/start", auth.AuthMiddleware(api.CaddyStartHandler))
	mux.HandleFunc("/api/caddy/stop", auth.AuthMiddleware(api.CaddyStopHandler))
	mux.HandleFunc("/api/caddy/restart", auth.AuthMiddleware(api.CaddyRestartHandler))
	mux.HandleFunc("/api/caddy/reload", auth.AuthMiddleware(api.CaddyReloadHandler))
	mux.HandleFunc("/api/caddy/ssl-status", auth.AuthMiddleware(api.CaddySSLStatusHandler))
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
	mux.HandleFunc("/api/projects/status", auth.AuthMiddleware(api.GetProjectStatusHandler))
	
	// 任务管理
	mux.HandleFunc("/api/tasks", auth.AuthMiddleware(api.TasksHandler))
	mux.HandleFunc("/api/tasks/add", auth.AuthMiddleware(api.AddTaskHandler))
	mux.HandleFunc("/api/tasks/delete", auth.AuthMiddleware(api.DeleteTaskHandler))
	mux.HandleFunc("/api/tasks/execute", auth.AuthMiddleware(api.ExecuteTaskHandler))

	// 应用程序控制
	mux.HandleFunc("/api/app/shutdown", auth.AuthMiddleware(api.ShutdownHandler))
	
	// 诊断和修复
	mux.HandleFunc("/api/diagnostics/run", auth.AuthMiddleware(api.DiagnosticsHandler))
	mux.HandleFunc("/api/diagnostics/ssl", auth.AuthMiddleware(api.CheckSSLHandler))
	mux.HandleFunc("/api/diagnostics/autofix", auth.AuthMiddleware(api.AutoFixHandler))
	mux.HandleFunc("/api/system/status", auth.AuthMiddleware(api.SystemStatusHandler))

	// 静态文件
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	// 前端页面
	mux.HandleFunc("/", api.IndexHandler)

	// 使用安全路径中间件
	handler := api.SecurityPathMiddleware(mux)

	addr := fmt.Sprintf(":%d", *port)
	
	// 创建 HTTP 服务器
	server := &http.Server{
		Addr:    addr,
		Handler: handler,
	}
	
	fmt.Println("============================================================")
	fmt.Println("                  Caddy 管理器 v0.0.11")
	fmt.Println("============================================================")
	fmt.Printf("\n🌐 访问地址: http://localhost:%d\n", *port)
	if !*noTray {
		fmt.Println("📋 系统托盘: 已启用 (右键查看菜单)")
	}
	fmt.Println("📊 按 Ctrl+C 停止服务")
	fmt.Println()
	
	// 启动系统托盘（在单独的 goroutine 中）
	if !*noTray {
		go tray.Run(*port)
	}
	
	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	
	// 启动服务器
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("服务启动失败: %v", err)
		}
	}()
	
	// 等待信号
	<-sigChan
	
	fmt.Println("\n正在关闭服务...")
	
	// 优雅关闭
	gracefulShutdown(server)
}

func gracefulShutdown(server *http.Server) {
	// 创建超时上下文
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	
	// 停止 Caddy
	fmt.Println("停止 Caddy 服务...")
	caddy.Stop()
	
	// 停止所有项目
	fmt.Println("停止所有项目...")
	api.StopAllProjects()
	
	// 关闭 HTTP 服务器
	fmt.Println("关闭 HTTP 服务器...")
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("服务器关闭错误: %v", err)
	}
	
	// 关闭数据库
	fmt.Println("关闭数据库连接...")
	database.Close()
	
	fmt.Println("✓ 服务已安全关闭")
}

func checkAdminPrivileges() {
	if !system.IsAdmin() {
		fmt.Println("============================================================")
		fmt.Println("⚠️  警告: 未以管理员身份运行")
		fmt.Println("============================================================")
		fmt.Println("")
		fmt.Println("当前程序未以管理员权限运行，这可能导致以下问题：")
		fmt.Println("  • 无法绑定 80 和 443 端口")
		fmt.Println("  • 无法自动申请 SSL 证书")
		fmt.Println("  • 无法配置防火墙规则")
		fmt.Println("")
		fmt.Println("建议操作：")
		fmt.Println("  1. 关闭本程序")
		fmt.Println("  2. 右键程序图标 → 选择'以管理员身份运行'")
		fmt.Println("  3. 或在 Web 界面的诊断页面查看详细说明")
		fmt.Println("")
		fmt.Println("如果只使用非标准端口（如 8080）可以忽略此警告")
		fmt.Println("============================================================")
		fmt.Println("")
		
		// 等待用户确认
		fmt.Print("按 Enter 键继续，或 Ctrl+C 退出...")
		fmt.Scanln()
		fmt.Println("")
	} else {
		fmt.Println("✓ 已以管理员权限运行")
	}
}
