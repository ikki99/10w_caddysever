# Caddy Manager

<div align="center">

![Version](https://img.shields.io/badge/version-1.0.0-blue.svg)
![License](https://img.shields.io/badge/license-MIT-green.svg)
![Platform](https://img.shields.io/badge/platform-Windows-lightgrey.svg)

**一个强大的 Caddy Web 服务器可视化管理工具**

[English](README_EN.md) | 简体中文

</div>

---

## ✨ 特性

- 🎯 **可视化管理** - 友好的 Web 界面，无需命令行
- 🚀 **项目部署** - 支持 Go、Node.js、Python、Java、PHP 等多种项目
- 🔒 **自动 SSL** - 自动申请和续期 Let's Encrypt 证书
- 📊 **实时监控** - 项目状态、日志查看、性能监控
- 🔧 **诊断工具** - SSL 检测、混合内容检查、IPv4/IPv6 兼容性
- 📁 **文件管理** - 在线文件浏览、上传、下载
- ⚙️ **灵活配置** - 反向代理、负载均衡、自定义头部
- 🎨 **系统托盘** - 最小化到系统托盘，便捷控制

## 🚀 快速开始

### 系统要求

- Windows 7 或更高版本
- 需要管理员权限（用于绑定 80/443 端口）

### 下载安装

1. 从 [Releases](https://github.com/ikki99/10w_caddysever/releases)下载最新版本
2. 解压到任意目录
3. 右键 `caddy-manager.exe` → 选择"以管理员身份运行"

### 首次运行

1. 启动程序后，访问 http://localhost:8989
2. 创建管理员账户
3. 开始部署您的第一个项目！

### 两个版本

- **caddy-manager.exe** - GUI 版本，无窗口运行（推荐日常使用）
- **caddy-manager-console.exe** - Console 版本，显示日志（推荐调试）

## 📚 使用指南

### 部署 Web 项目

1. 点击"新建项目"
2. 选择项目类型（Go/Node.js/Python 等）
3. 填写项目信息（名称、路径、端口）
4. 配置域名和 SSL
5. 选择代理连接方式（推荐 IPv4）
6. 保存并启动

### SSL 证书配置

自动申请 Let's Encrypt 证书需要：

1. ✅ 有效的域名
2. ✅ 域名已解析到服务器
3. ✅ 开放 80 和 443 端口
4. ✅ 以管理员身份运行

### IPv4/IPv6 兼容性

如果遇到 **502 Bad Gateway** 错误：

1. 编辑项目配置
2. 在"代理连接方式"选择 **IPv4 (127.0.0.1)**
3. 保存配置即可解决

> 大多数 Go/Node.js 程序只监听 IPv4，使用 localhost 可能导致连接失败

### 混合内容检测

如果 HTTPS 网站显示黄色叹号：

1. 双击运行 `检测SSL问题.bat`
2. 选择"检测项目中的混合内容"
3. 根据报告修复 HTTP 资源

## 🔧 主要功能

### 项目管理

- ✅ 支持多种项目类型
- ✅ 自动启动和监控
- ✅ 实时日志查看
- ✅ 一键启动/停止/重启

### SSL 证书

- ✅ 自动申请 Let's Encrypt 证书
- ✅ 自动续期
- ✅ SSL 状态检查
- ✅ 证书诊断工具

### 反向代理

- ✅ HTTP/HTTPS 反向代理
- ✅ IPv4/IPv6 连接方式选择
- ✅ 自定义 Header
- ✅ 路径匹配

### 文件管理

- ✅ 在线文件浏览
- ✅ 文件上传下载
- ✅ 创建文件夹
- ✅ 删除文件

### 系统诊断

- ✅ 运行诊断检查
- ✅ SSL 配置检查
- ✅ 混合内容检测
- ✅ 自动修复工具

## 🛠️ 开发

### 构建

```bash
# 克隆仓库
git clone https://github.com/10w-server/caddy-manager.git
cd caddy-manager

# 编译
go build -ldflags="-s -w" -o caddy-manager-console.exe
go build -ldflags="-s -w -H=windowsgui" -o caddy-manager.exe
```

### 项目结构

```
caddy-manager/
├── internal/
│   ├── api/          # API 处理
│   ├── auth/         # 身份认证
│   ├── caddy/        # Caddy 管理
│   ├── config/       # 配置管理
│   ├── database/     # 数据库
│   ├── diagnostics/  # 诊断工具
│   ├── models/       # 数据模型
│   ├── system/       # 系统工具
│   └── tray/         # 系统托盘
├── web/
│   └── static/       # 前端资源
├── data/             # 数据目录
│   ├── caddy/        # Caddy 程序
│   ├── logs/         # 日志文件
│   └── www/          # 网站文件
└── main.go           # 主程序
```

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📝 更新日志

### v1.0.0 (2025-01-09)

**主要功能：**
- ✅ 完整的项目管理系统
- ✅ 自动 SSL 证书申请
- ✅ IPv4/IPv6 代理连接选择
- ✅ 混合内容检测工具
- ✅ 系统诊断和自动修复

**修复：**
- ✅ Session 超时问题（延长至 7 天）
- ✅ 黑框闪烁问题（双版本编译）
- ✅ 诊断按钮无反应
- ✅ Caddy 状态显示
- ✅ 502 Bad Gateway（IPv4/IPv6 兼容性）

详细更新日志请查看 [CHANGELOG.md](CHANGELOG.md)

## 📄 许可证

本项目采用 MIT 许可证 - 详见 [LICENSE](LICENSE) 文件

## 👤 作者

**10w**

- 📧 Email: wngx99@gmail.com
- 🐙 GitHub: [@10w-server](https://github.com/10w-server)

## ⭐ Star History

如果这个项目对您有帮助，请给个 Star ⭐

## 🙏 致谢

- [Caddy](https://caddyserver.com/) - 优秀的 Web 服务器
- [Let's Encrypt](https://letsencrypt.org/) - 免费 SSL 证书

---

<div align="center">

**[⬆ 回到顶部](#caddy-manager)**

Made with ❤️ by 10w

</div>
