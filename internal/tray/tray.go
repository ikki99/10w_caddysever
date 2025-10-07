package tray

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/getlantern/systray"
	
	"caddy-manager/internal/caddy"
)

var (
	port int
	statusItem *systray.MenuItem
)

func Run(webPort int) {
	port = webPort
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetTitle("Caddy")
	systray.SetTooltip("Caddy 管理器")
	
	statusItem = systray.AddMenuItem("状态: 检查中...", "Caddy 运行状态")
	statusItem.Disable()
	
	systray.AddSeparator()
	
	mOpen := systray.AddMenuItem("打开管理面板", "在浏览器中打开")
	mCaddyStart := systray.AddMenuItem("启动 Caddy", "启动 Caddy 服务")
	mCaddyStop := systray.AddMenuItem("停止 Caddy", "停止 Caddy 服务")
	mCaddyRestart := systray.AddMenuItem("重启 Caddy", "重启 Caddy 服务")
	
	systray.AddSeparator()
	
	mQuit := systray.AddMenuItem("退出程序", "退出 Caddy 管理器")

	go updateStatus()

	go func() {
		for {
			select {
			case <-mOpen.ClickedCh:
				openBrowser()
			case <-mCaddyStart.ClickedCh:
				caddy.Start()
			case <-mCaddyStop.ClickedCh:
				caddy.Stop()
			case <-mCaddyRestart.ClickedCh:
				caddy.Restart()
			case <-mQuit.ClickedCh:
				caddy.Stop()
				systray.Quit()
				return
			}
		}
	}()
}

func onExit() {
	// 清理
}

func updateStatus() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	
	for range ticker.C {
		if caddy.IsRunning() {
			statusItem.SetTitle("状态: Caddy 运行中")
		} else {
			statusItem.SetTitle("状态: Caddy 未运行")
		}
	}
}

func openBrowser() {
	url := fmt.Sprintf("http://localhost:%d", port)
	exec.Command("cmd", "/c", "start", url).Start()
}
