# Changelog

All notable changes to this project will be documented in this file.

## [1.0.0] - 2025-01-09

### 🎉 首个正式版本

这是 Caddy Manager 的第一个正式发布版本，集成了所有核心功能和重要修复。

### ✨ 新增功能

#### 项目管理
- **多项目支持** - 同时管理多个 Web 项目
- **项目类型** - 支持 Go、Node.js、Python、Java、PHP、静态站点
- **自动启动** - 支持开机自动启动项目
- **实时监控** - 项目运行状态实时显示
- **日志查看** - 在线查看项目运行日志

#### SSL 证书
- **自动申请** - 自动申请 Let's Encrypt 免费证书
- **自动续期** - 证书快过期时自动续期
- **SSL 检查** - 检测 SSL 配置问题
- **诊断工具** - 完整的 SSL 故障排查工具

#### 反向代理
- **IPv4/IPv6 选择** - 解决 502 Bad Gateway 问题
- **自定义 Header** - 支持添加自定义请求头
- **路径匹配** - 灵活的路径代理配置
- **智能生成** - 自动生成 Caddyfile 配置

#### 文件管理
- **在线浏览** - Web 界面浏览服务器文件
- **文件上传** - 拖拽上传或选择文件
- **文件下载** - 一键下载文件到本地
- **文件夹管理** - 创建、删除文件夹

#### 系统工具
- **系统诊断** - 检查端口占用、防火墙等
- **混合内容检测** - 自动检测 HTTPS 页面的 HTTP 资源
- **自动修复** - 部分问题支持一键修复
- **系统托盘** - 最小化到系统托盘运行

### 🐛 问题修复

#### Session 管理
- **问题**: 每次刷新页面都需要重新登录
- **修复**: 延长 Session 有效期至 7 天，并自动续期
- **影响文件**: `internal/auth/auth.go`, `internal/api/handlers.go`

#### 黑框闪烁
- **问题**: 使用 GUI 版本时不断闪出黑色命令行窗口
- **修复**: 
  - 编译两个版本：Console（调试）和 GUI（生产）
  - 所有系统命令添加 HideWindow 属性
- **影响文件**: `build.bat`, `internal/caddy/downloader.go`, `internal/diagnostics/diagnostics.go`

#### 诊断按钮无反应
- **问题**: 系统设置中的诊断按钮点击没有反应
- **修复**: 添加完整的错误处理和 401 认证检测
- **影响文件**: `web/static/app.js`

#### Caddy 状态显示
- **问题**: Web 页面无法正确显示 Caddy 运行状态
- **修复**: 改进 `IsRunning()` 和 `GetVersion()` 检测逻辑
- **影响文件**: `internal/caddy/downloader.go`, `web/static/app.js`

#### IPv4/IPv6 兼容性
- **问题**: Caddy 使用 IPv6 连接导致 502 Bad Gateway
- **现象**: 日志显示 `dial tcp [::1]:端口: connectex: No connection...`
- **原因**: Caddy 使用 `localhost` 被解析为 IPv6，但应用只监听 IPv4
- **修复**: 
  - 新增"代理连接方式"选项
  - 支持强制使用 IPv4（127.0.0.1）
  - 自动数据库迁移
- **影响文件**: `internal/database/database.go`, `internal/models/models.go`, `internal/api/projects.go`, `internal/api/template.go`, `web/static/app.js`

#### SSL 黄色叹号
- **问题**: HTTPS 网站显示黄色叹号/不安全警告
- **原因**: 混合内容（HTTPS 页面加载 HTTP 资源）
- **修复**: 
  - 创建混合内容检测工具
  - 自动扫描项目文件
  - 提供修复建议和自动修复选项
- **新增文件**: `检测混合内容.ps1`, `检测SSL问题.bat`, `混合内容检测修复指南.md`

### 🔨 改进

#### 性能优化
- 优化数据库查询效率
- 改进进程检测算法
- 减少不必要的文件读写

#### 用户体验
- 更友好的错误提示
- 详细的帮助文档
- 完善的新手引导

#### 安全性
- Session 自动续期机制
- Cookie SameSite 属性
- 更安全的密码存储

### 📚 文档

新增完整文档：
- ✅ 快速开始指南
- ✅ SSL 证书配置
- ✅ IPv4/IPv6 兼容性问题
- ✅ 混合内容检测修复
- ✅ 故障排查指南

### 🛠️ 技术栈

- **后端**: Go 1.21+
- **前端**: 原生 JavaScript
- **数据库**: SQLite
- **Web 服务器**: Caddy 2.10.2

### 📦 构建

```bash
# Console 版本
go build -ldflags="-s -w" -o caddy-manager-console.exe

# GUI 版本
go build -ldflags="-s -w -H=windowsgui" -o caddy-manager.exe
```

### 🙏 致谢

感谢所有测试用户的反馈和建议，特别是发现 IPv4/IPv6 兼容性问题的用户！

---

## 版本说明

版本号格式：`主版本号.次版本号.修订号`

- **主版本号**: 重大更新，可能包含不兼容的更改
- **次版本号**: 新功能添加，向后兼容
- **修订号**: Bug 修复和小改进

[1.0.0]: https://github.com/10w-server/caddy-manager/releases/tag/v1.0.0
