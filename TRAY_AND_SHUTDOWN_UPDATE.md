# 托盘图标和应用关闭功能更新

## 新增功能

### 🎯 系统托盘增强

#### 托盘图标
- ✅ 添加了自定义托盘图标 (`internal/tray/icon.ico`)
- ✅ 托盘提示信息显示运行端口和 Caddy 状态
- ✅ 图标使用 embed 方式嵌入，无需外部文件

#### 托盘菜单功能
1. **状态显示**
   - 实时显示 Caddy 运行状态
   - 每 5 秒自动更新
   - 运行中: `● 状态: Caddy 运行中 ✓`
   - 未运行: `○ 状态: Caddy 未运行 ✗`

2. **管理面板**
   - 🌐 打开管理面板 - 自动在浏览器中打开

3. **Caddy 控制**
   - ▶ 启动 Caddy - 启动 Caddy 服务器
   - ⏸ 停止 Caddy - 停止 Caddy 服务器
   - 🔄 重启 Caddy - 重启 Caddy 服务器
   - 操作后自动刷新状态

4. **应用控制**
   - 📋 显示控制台 - 显示程序控制台窗口
   - 🔽 隐藏控制台 - 隐藏程序控制台窗口（后台运行）

5. **关于和退出**
   - ℹ 关于 - 显示应用程序版本和功能信息
   - ❌ 退出程序 - 优雅退出（带确认对话框）

### 🛑 应用程序优雅关闭

#### Web 界面关闭
- 在"系统设置"页面新增"应用程序控制"区域
- 添加"关闭应用程序"按钮
- 点击前需要确认
- 关闭时会提示用户

#### 托盘退出
- 右键托盘图标选择"退出程序"
- 显示确认对话框
- 确认后执行优雅关闭流程

#### 键盘中断 (Ctrl+C)
- 支持 Ctrl+C 优雅关闭
- 自动触发关闭流程

#### 关闭流程
1. 停止 Caddy 服务
2. 停止所有运行中的项目
3. 关闭 HTTP 服务器（10秒超时）
4. 关闭数据库连接
5. 退出程序

### 🔧 技术改进

#### 信号处理
```go
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
```

#### 优雅关闭
- 使用 context.WithTimeout 控制关闭超时
- 先停止接收新请求
- 等待现有请求完成
- 清理所有资源

#### 新增 API
- `POST /api/app/shutdown` - 关闭应用程序

#### 新增函数
- `api.StopAllProjects()` - 停止所有运行中的项目
- `gracefulShutdown()` - 执行优雅关闭
- `tray.Quit()` - 退出托盘

## 使用说明

### 启动应用
```bash
# 正常启动（带托盘）
caddy-manager.exe

# 不显示托盘
caddy-manager.exe --no-tray

# 自定义端口
caddy-manager.exe --port 9090
```

### 后台运行
1. 启动应用程序
2. 右键托盘图标
3. 选择"隐藏控制台"
4. 程序在后台运行，只显示托盘图标

### 退出应用

**方法 1: 托盘菜单**
1. 右键托盘图标
2. 选择"❌ 退出程序"
3. 确认退出

**方法 2: Web 界面**
1. 打开管理面板
2. 进入"系统设置"
3. 点击"关闭应用程序"
4. 确认关闭

**方法 3: 键盘**
- 控制台窗口中按 `Ctrl+C`

## 文件修改

### 新增文件
- `internal/tray/icon.ico` - 托盘图标文件

### 修改文件
1. `internal/tray/tray.go`
   - 增强托盘菜单
   - 添加图标支持
   - 添加控制台显示/隐藏
   - 添加关于对话框
   - 添加退出确认

2. `main.go`
   - 添加信号处理
   - 实现优雅关闭
   - 改进服务器启动方式

3. `internal/api/handlers.go`
   - 新增 ShutdownHandler

4. `internal/api/projects.go`
   - 新增 StopAllProjects 函数

5. `internal/api/template.go`
   - 设置页面添加关闭按钮

6. `web/static/app.js`
   - 新增 shutdownApplication 函数

## 注意事项

### Windows 特性
- 控制台显示/隐藏功能仅支持 Windows
- 使用 PowerShell 调用 Windows API
- 托盘图标在 Windows 通知区域显示

### 托盘图标
- 图标使用 embed 嵌入到可执行文件
- 不需要外部 .ico 文件
- 图标为蓝色 "C" 字母（代表 Caddy）

### 优雅关闭
- 超时时间为 10 秒
- 如果有长时间运行的请求可能被强制终止
- 所有项目进程会被停止
- 数据库连接正确关闭

## 版本信息

- **版本**: v0.0.10
- **更新日期**: 2025年
- **兼容性**: Windows 10/11

## 下一步计划

- [ ] 托盘图标根据状态变化（运行/停止不同颜色）
- [ ] 托盘气泡通知
- [ ] 托盘右键菜单添加快速操作
- [ ] 支持最小化到托盘
- [ ] 开机自启动选项

---

**构建命令**:
```bash
go build -o caddy-manager.exe
```

**注意**: 如果需要无控制台窗口启动，使用：
```bash
go build -o caddy-manager.exe -ldflags="-H windowsgui"
```
