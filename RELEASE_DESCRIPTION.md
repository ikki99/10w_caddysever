## 🎉 Caddy Manager v1.0.0

这是 Caddy Manager 的首个正式版本！一个强大的 Caddy Web 服务器可视化管理工具。

### ✨ 主要功能

- 🎯 **可视化管理** - 友好的 Web 界面，无需命令行操作
- 🚀 **项目部署** - 支持 Go、Node.js、Python、Java、PHP、静态站点等多种项目类型
- 🔒 **自动 SSL** - 自动申请和续期 Let's Encrypt 免费证书
- 📊 **实时监控** - 项目运行状态实时显示，在线查看日志
- 🔧 **诊断工具** - SSL 检测、混合内容检查、IPv4/IPv6 兼容性诊断
- 📁 **文件管理** - 在线文件浏览、上传、下载
- ⚙️ **灵活配置** - IPv4/IPv6 代理连接选择、反向代理、自定义头部
- 🎨 **系统托盘** - 最小化到系统托盘，便捷控制

### 📥 下载

下载对应版本：
- **caddy-manager.exe** - GUI 版本（推荐日常使用，无窗口运行）
- **caddy-manager-console.exe** - Console 版本（推荐调试，显示日志）

### 🚀 快速开始

1. 右键 `caddy-manager.exe` → 选择"以管理员身份运行"
2. 访问 http://localhost:8989
3. 创建管理员账户
4. 开始部署您的第一个项目！

### 🐛 重要修复

#### Session 管理
- **问题**：每次刷新页面都需要重新登录
- **修复**：延长 Session 有效期至 7 天，并自动续期

#### 黑框闪烁
- **问题**：使用 GUI 版本时不断闪出黑色命令行窗口
- **修复**：编译两个版本，所有系统命令添加 HideWindow 属性

#### IPv4/IPv6 兼容性 ⭐ 重要
- **问题**：Caddy 使用 IPv6 连接导致 502 Bad Gateway
- **现象**：日志显示 `dial tcp [::1]:端口: connectex: No connection...`
- **修复**：新增"代理连接方式"选项，支持强制使用 IPv4（127.0.0.1）

#### 诊断按钮
- **问题**：系统设置中的诊断按钮点击没有反应
- **修复**：添加完整的错误处理和 401 认证检测

#### Caddy 状态显示
- **问题**：Web 页面无法正确显示 Caddy 运行状态
- **修复**：改进 IsRunning() 和 GetVersion() 检测逻辑

### 🔧 新增工具

#### 混合内容检测
- **问题**：HTTPS 网站显示黄色叹号/不安全警告
- **原因**：混合内容（HTTPS 页面加载 HTTP 资源）
- **工具**：
  - `检测SSL问题.bat` - 一键检测工具
  - `检测混合内容.ps1` - PowerShell 自动检测脚本
  - 自动扫描项目文件并提供修复建议

### 📚 文档

- [README](https://github.com/ikki99/10w_caddysever/blob/main/README.md) - 完整使用指南
- [CHANGELOG](https://github.com/ikki99/10w_caddysever/blob/main/CHANGELOG.md) - 详细更新日志
- [IPv4/IPv6 兼容性](https://github.com/ikki99/10w_caddysever/blob/main/IPv4-IPv6兼容性问题.md) - 技术文档
- [混合内容检测](https://github.com/ikki99/10w_caddysever/blob/main/混合内容检测修复指南.md) - 使用指南

### ⚠️ 重要提示

- **管理员权限**：需要以管理员权限运行才能绑定 80 和 443 端口
- **502 错误**：如遇 502 Bad Gateway，请在项目配置中选择 IPv4 连接方式
- **SSL 黄色叹号**：HTTPS 显示黄色叹号，请使用混合内容检测工具
- **防火墙**：确保开放 80 和 443 端口

### 🛠️ 系统要求

- Windows 7 或更高版本
- 需要管理员权限
- 建议 2GB 以上内存

### 📝 技术栈

- **后端**：Go 1.21+
- **前端**：原生 JavaScript
- **数据库**：SQLite
- **Web 服务器**：Caddy 2.10.2

### 🙏 致谢

感谢所有测试用户的反馈和建议，特别是发现 IPv4/IPv6 兼容性问题的用户！

---

**制作者**: 10w  
**邮箱**: wngx99@gmail.com  
**GitHub**: [@10w-server](https://github.com/10w-server)

---

如有问题，请提交 [Issue](https://github.com/ikki99/10w_caddysever/issues)
