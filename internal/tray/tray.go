package tray

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/getlantern/systray"
	
	"caddy-manager/internal/caddy"
)

var (
	port int
	statusItem *systray.MenuItem
	quitChannel chan bool
)

func Run(webPort int) {
	port = webPort
	quitChannel = make(chan bool)
	systray.Run(onReady, onExit)
}

func onReady() {
	// 设置托盘图标
	iconData, err := os.ReadFile("internal/tray/icon.ico")
	if err == nil {
		systray.SetIcon(iconData)
	}
	
	// 设置托盘标题和提示
	systray.SetTitle("Caddy")
	systray.SetTooltip("Caddy 管理器 - 运行在端口 " + fmt.Sprintf("%d", port))
	
	// 状态显示
	statusItem = systray.AddMenuItem("● 状态: 检查中...", "Caddy 服务运行状态")
	statusItem.Disable()
	
	systray.AddSeparator()
	
	// 管理面板
	mOpen := systray.AddMenuItem("🌐 打开管理面板", "在浏览器中打开 Web 管理界面")
	
	systray.AddSeparator()
	
	// Caddy 控制
	mCaddyStart := systray.AddMenuItem("▶ 启动 Caddy", "启动 Caddy 服务器")
	mCaddyStop := systray.AddMenuItem("⏸ 停止 Caddy", "停止 Caddy 服务器")
	mCaddyRestart := systray.AddMenuItem("🔄 重启 Caddy", "重启 Caddy 服务器")
	
	systray.AddSeparator()
	
	// 应用控制
	mShowWindow := systray.AddMenuItem("📋 显示控制台", "显示应用程序控制台窗口")
	mHideWindow := systray.AddMenuItem("🔽 隐藏控制台", "隐藏应用程序控制台窗口")
	
	systray.AddSeparator()
	
	// 关于和退出
	mAbout := systray.AddMenuItem("ℹ 关于", "关于 Caddy 管理器")
	mQuit := systray.AddMenuItem("❌ 退出程序", "停止所有服务并退出 Caddy 管理器")

	// 启动状态更新
	go updateStatus()

	// 事件处理
	go func() {
		for {
			select {
			case <-mOpen.ClickedCh:
				openBrowser()
			case <-mCaddyStart.ClickedCh:
				go func() {
					caddy.Start()
					time.Sleep(1 * time.Second)
					updateStatusOnce()
				}()
			case <-mCaddyStop.ClickedCh:
				go func() {
					caddy.Stop()
					time.Sleep(1 * time.Second)
					updateStatusOnce()
				}()
			case <-mCaddyRestart.ClickedCh:
				go func() {
					caddy.Restart()
					time.Sleep(2 * time.Second)
					updateStatusOnce()
				}()
			case <-mShowWindow.ClickedCh:
				showConsoleWindow()
			case <-mHideWindow.ClickedCh:
				hideConsoleWindow()
			case <-mAbout.ClickedCh:
				showAbout()
			case <-mQuit.ClickedCh:
				exitApplication()
				return
			case <-quitChannel:
				return
			}
		}
	}()
}

func onExit() {
	// 停止 Caddy 服务
	caddy.Stop()
}

func updateStatus() {
	// 立即更新一次
	updateStatusOnce()
	
	// 定期更新
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		updateStatusOnce()
	}
}

func updateStatusOnce() {
	if caddy.IsRunning() {
		statusItem.SetTitle("● 状态: Caddy 运行中 ✓")
		systray.SetTooltip("Caddy 管理器 - Caddy 运行中")
	} else {
		statusItem.SetTitle("○ 状态: Caddy 未运行 ✗")
		systray.SetTooltip("Caddy 管理器 - Caddy 未运行")
	}
}

func openBrowser() {
	url := fmt.Sprintf("http://localhost:%d", port)
	exec.Command("cmd", "/c", "start", url).Start()
}

func showConsoleWindow() {
	// Windows API 调用显示控制台
	cmd := exec.Command("powershell", "-Command", "Add-Type -Name Window -Namespace Console -MemberDefinition '[DllImport(\"Kernel32.dll\")]public static extern IntPtr GetConsoleWindow();[DllImport(\"user32.dll\")]public static extern bool ShowWindow(IntPtr hWnd, Int32 nCmdShow);';$consolePtr = [Console.Window]::GetConsoleWindow();[Console.Window]::ShowWindow($consolePtr, 5)")
	cmd.Run()
}

func hideConsoleWindow() {
	// Windows API 调用隐藏控制台
	cmd := exec.Command("powershell", "-Command", "Add-Type -Name Window -Namespace Console -MemberDefinition '[DllImport(\"Kernel32.dll\")]public static extern IntPtr GetConsoleWindow();[DllImport(\"user32.dll\")]public static extern bool ShowWindow(IntPtr hWnd, Int32 nCmdShow);';$consolePtr = [Console.Window]::GetConsoleWindow();[Console.Window]::ShowWindow($consolePtr, 0)")
	cmd.Run()
}

func showAbout() {
	aboutMsg := "Caddy 管理器 v0.0.11\n\n" +
		"一个现代化的 Caddy Web 服务器管理工具\n\n" +
		"功能特性:\n" +
		"• 项目管理和部署\n" +
		"• 自动 SSL 证书\n" +
		"• 文件管理\n" +
		"• 计划任务\n" +
		"• 系统监控\n\n" +
		fmt.Sprintf("管理面板: http://localhost:%d\n", port) +
		"\n© 2025 Caddy Manager"
	
	exec.Command("mshta", "javascript:alert('"+aboutMsg+"');close();").Start()
}

func exitApplication() {
	// 显示确认对话框
	cmd := exec.Command("powershell", "-Command", 
		"$result = [System.Windows.MessageBox]::Show('确定要退出 Caddy 管理器吗？这将停止所有正在运行的服务。','确认退出','YesNo','Question'); if($result -eq 'Yes'){exit 0}else{exit 1}")
	
	if err := cmd.Run(); err != nil {
		// 用户点击了 No，不退出
		return
	}
	
	// 停止 Caddy
	caddy.Stop()
	
	// 通知主程序退出
	close(quitChannel)
	
	// 退出托盘
	systray.Quit()
	
	// 退出应用程序
	os.Exit(0)
}

func Quit() {
	if quitChannel != nil {
		close(quitChannel)
	}
	systray.Quit()
}
