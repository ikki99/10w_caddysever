# 更新日志 - v0.0.10

## 🎯 主要更新

### 系统托盘功能完善
- ✅ 添加自定义托盘图标（蓝色 C 字母）
- ✅ 托盘菜单增强，支持更多操作
- ✅ 实时状态显示（Caddy 运行状态）
- ✅ 控制台窗口显示/隐藏功能
- ✅ 关于对话框
- ✅ 退出确认对话框

### 应用程序优雅关闭
- ✅ Web 界面关闭应用功能
- ✅ 托盘菜单退出功能
- ✅ Ctrl+C 信号处理
- ✅ 优雅关闭流程（停止服务、关闭连接、清理资源）
- ✅ 自动停止所有运行中的项目

## 📋 详细功能

### 托盘菜单项
1. **状态显示区**
   - ● 状态: Caddy 运行中 ✓
   - ○ 状态: Caddy 未运行 ✗

2. **管理操作**
   - 🌐 打开管理面板
   - ▶ 启动 Caddy
   - ⏸ 停止 Caddy
   - 🔄 重启 Caddy

3. **窗口控制**
   - 📋 显示控制台
   - 🔽 隐藏控制台

4. **系统功能**
   - ℹ 关于
   - ❌ 退出程序

### 关闭应用的三种方式
1. **Web 界面**: 系统设置 → 应用程序控制 → 关闭应用程序
2. **托盘菜单**: 右键托盘 → 退出程序
3. **键盘快捷键**: 控制台窗口按 Ctrl+C

### 优雅关闭流程
```
用户触发退出
    ↓
显示确认对话框（托盘方式）
    ↓
停止 Caddy 服务
    ↓
停止所有项目进程
    ↓
关闭 HTTP 服务器（10秒超时）
    ↓
关闭数据库连接
    ↓
退出程序
```

## 🔧 技术实现

### 新增 API
- `POST /api/app/shutdown` - 关闭应用程序

### 新增函数
- `api.StopAllProjects()` - 批量停止项目
- `gracefulShutdown(server)` - 优雅关闭
- `tray.Quit()` - 退出托盘
- `tray.exitApplication()` - 处理退出逻辑
- `tray.showAbout()` - 显示关于信息
- `tray.showConsoleWindow()` - 显示控制台
- `tray.hideConsoleWindow()` - 隐藏控制台

### 信号处理
```go
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
```

### 图标嵌入
```go
//go:embed icon.ico
var iconData []byte
```

## 📁 文件变更

### 新增
- `internal/tray/icon.ico` - 托盘图标文件
- `TRAY_AND_SHUTDOWN_UPDATE.md` - 详细更新说明

### 修改
- `internal/tray/tray.go` - 托盘功能增强
- `main.go` - 信号处理和优雅关闭
- `internal/api/handlers.go` - 关闭应用 API
- `internal/api/projects.go` - 批量停止项目
- `internal/api/template.go` - 设置页面增强
- `web/static/app.js` - 关闭应用前端功能

## 🎨 用户体验改进

### 后台运行
- 启动后可隐藏控制台窗口
- 程序在系统托盘静默运行
- 需要时可随时显示控制台

### 状态监控
- 托盘图标提示显示当前状态
- 菜单中实时显示 Caddy 运行状态
- 每 5 秒自动更新状态

### 安全退出
- 退出前显示确认对话框
- 自动清理所有运行中的服务
- 防止数据丢失

## 🐛 修复的问题

1. ✅ 退出程序时项目进程没有停止
2. ✅ 托盘图标没有自定义图标
3. ✅ 没有控制台显示/隐藏功能
4. ✅ 强制关闭导致资源未释放

## 📝 使用示例

### 启动应用
```bash
# 带托盘启动
caddy-manager.exe

# 无托盘启动
caddy-manager.exe --no-tray

# 自定义端口
caddy-manager.exe --port 9090
```

### 后台运行
1. 启动程序
2. 右键托盘图标 → 隐藏控制台
3. 程序在后台运行

### 退出程序
1. 右键托盘图标
2. 选择"退出程序"
3. 确认退出
4. 等待服务停止

## ⚠️ 注意事项

- 控制台显示/隐藏功能仅支持 Windows
- 优雅关闭超时设置为 10 秒
- 托盘图标在系统通知区域显示
- 退出时会自动停止所有服务和项目

## 🚀 性能优化

- 托盘状态更新间隔优化为 5 秒
- 减少不必要的状态查询
- 优化关闭流程，避免资源泄漏

## 📊 兼容性

- **操作系统**: Windows 10/11
- **Go 版本**: 1.16+
- **依赖**: github.com/getlantern/systray

## 🔜 未来计划

- [ ] 托盘图标根据状态变色
- [ ] 托盘气泡通知
- [ ] 最小化到托盘
- [ ] 开机自启动
- [ ] 托盘快捷操作菜单

---

**版本**: v0.0.10  
**发布日期**: 2025年  
**作者**: Caddy Manager Team

**安装方式**: 直接运行 `caddy-manager.exe`  
**卸载方式**: 停止程序后删除程序文件和 `data` 目录
