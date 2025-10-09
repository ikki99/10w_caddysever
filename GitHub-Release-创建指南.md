# GitHub Release 创建指南 - v1.0.4

## 📦 准备好的文件

✅ **发布包：** `caddy-manager-v1.0.4-windows.zip` (9.41 MB)
✅ **发布说明：** `RELEASE_v1.0.4.md`
✅ **Git Tag：** v1.0.4 已推送到 GitHub

---

## 🚀 手动创建 Release 步骤

### 1. 访问 GitHub Releases 页面

打开浏览器，访问：
```
https://github.com/ikki99/10w_caddysever/releases
```

### 2. 点击 "Draft a new release"

在页面右上角找到并点击绿色按钮：**Draft a new release**

### 3. 选择 Tag

- **Choose a tag:** 选择 `v1.0.4`（应该已经存在）
- 如果没有，可以输入 `v1.0.4` 创建新的

### 4. 填写 Release 标题

在 **Release title** 输入框中填写：
```
Caddy Manager v1.0.4 - 紧急修复版本
```

### 5. 填写 Release 描述

复制以下内容到 **Describe this release** 文本框：

---

## 🐛 紧急修复版本

**发布日期：** 2025-01-09

这是一个紧急修复版本，解决了 v1.0.0-v1.0.3 中发现的重要问题。**强烈建议所有用户更新！**

---

## 🔥 重要修复

### 1. Session 持久化问题 ✅
**问题：** 刷新页面会跳转到登录页，即使 Session 仍然有效。

**修复：**
- 新增专用 `/api/auth/check` Session 检查接口
- 修改页面加载逻辑，先检查登录状态
- Session 7天有效期 + 自动续期正常工作

**影响：** 现在刷新页面可以保持登录状态，不需要重新登录。

---

### 2. 文件管理器新建文件夹位置错误 ✅
**问题：** 点击"新建文件夹"会在程序运行目录创建，而不是当前浏览目录。

**修复：**
- 添加 `initializeFilePath()` 函数
- 登录后立即初始化文件路径
- 改进错误处理容错机制

**影响：** 新建文件夹现在会在正确的位置创建。

---

### 3. 安全性改进 ✅
**改进：**
- Cookie 使用更严格的 `SameSite: Strict` 模式
- `HttpOnly` 防止 XSS 攻击
- 更完善的错误处理

---

## 📦 下载

### Windows 版本

| 文件 | 大小 | 说明 |
|------|------|------|
| **caddy-manager.exe** | ~11 MB | GUI 版本，无窗口运行（推荐） |
| **caddy-manager-console.exe** | ~11 MB | Console 版本，显示日志（调试用） |

### 安装步骤

1. 下载下方的 `caddy-manager-v1.0.4-windows.zip`
2. 解压到任意目录
3. **右键 `caddy-manager.exe`** → 选择"**以管理员身份运行**"
4. 访问 http://localhost:8989
5. 开始使用！

---

## ⚠️ 升级说明

### 从 v1.0.0-v1.0.3 升级

**数据安全：**
- ✅ 数据库完全兼容，无需迁移
- ✅ 所有项目和设置保持不变
- ✅ 直接替换 exe 文件即可

**升级步骤：**
1. 停止旧版本程序
2. 备份 `data` 目录（可选，但建议）
3. 用新版本 exe 替换旧版本
4. 启动新版本
5. 登录验证一切正常

---

## 🔧 技术细节

### 新增文件
- `internal/api/auth_check.go` - Session 检查处理器
- `internal/api/files.go` - 文件管理增强
- `internal/api/monitor.go` - 系统监控
- `internal/system/monitor.go` - 系统资源监控
- `web/static/file-manager.js` - 文件管理器前端

### 修改的文件
- `web/static/app.js` - 修复 Session 检查逻辑
- `internal/api/handlers.go` - Cookie 安全配置
- `internal/api/template.go` - 版本号更新

### API 变更
- **新增：** `GET /api/auth/check` - Session 检查接口
- **改进：** 所有受保护的 API 认证机制

---

## 📝 完整更新日志

查看 [CHANGELOG.md](https://github.com/ikki99/10w_caddysever/blob/main/CHANGELOG.md) 了解详细的版本历史。

---

## 🐛 已知问题

**无** - 本版本修复了所有已知的重要问题。

---

## 💬 反馈与支持

- **GitHub Issues：** https://github.com/ikki99/10w_caddysever/issues
- **邮箱：** wngx99@gmail.com
- **作者：** 10w

---

## 📜 许可证

MIT License - 详见 [LICENSE](https://github.com/ikki99/10w_caddysever/blob/main/LICENSE)

---

**感谢使用 Caddy Manager！** 🎉

如果这个项目对你有帮助，欢迎给个 ⭐ Star！

---

### 6. 上传发布文件

在 **Attach binaries** 区域：
- 拖拽或点击上传 `caddy-manager-v1.0.4-windows.zip`
- 等待上传完成

### 7. 发布 Release

- 确认所有信息正确
- 点击绿色按钮：**Publish release**

---

## ✅ 完成！

Release 创建成功后：
- 访问：https://github.com/ikki99/10w_caddysever/releases/tag/v1.0.4
- 验证所有信息显示正确
- 测试下载链接是否正常

---

**制作者：** 10w  
**邮箱：** wngx99@gmail.com  
**项目：** https://github.com/ikki99/10w_caddysever
