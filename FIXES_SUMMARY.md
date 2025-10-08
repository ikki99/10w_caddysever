# 问题修复说明

## 修复的问题

### 1. 项目列表可以修改了
**问题**: 项目列表不能修改，只能删除
**解决方案**:
- 在每个项目卡片上添加了"编辑"按钮
- 新增 `editProject(id)` 函数，可以加载现有项目的所有配置
- 修改 `submitProject()` 函数，支持创建和更新两种模式
- 更新项目时会调用 `/api/projects/update` 接口
- 模态框标题会根据模式显示"新建项目"或"编辑项目"

**修改的文件**:
- `web/static/app.js` - 添加编辑功能
- `internal/api/template.go` - 添加动态标题支持

### 2. 项目状态实时检测
**问题**: 项目状态没有实时更新
**解决方案**:
- 改进了 `getProjectStatus()` 函数，使用更可靠的 Windows netstat 命令检测端口占用
- 添加了 `/api/projects/status` 接口，可以单独查询项目状态
- 在项目列表页面添加了自动刷新机制，每5秒更新一次状态
- 状态检测会同时检查进程记录和端口监听状态

**修改的文件**:
- `internal/api/projects.go` - 改进状态检测逻辑，添加状态查询接口
- `web/static/app.js` - 添加定时刷新
- `main.go` - 注册新的状态查询路由

### 3. SSL错误检测功能
**问题**: 申请SSL证书失败时没有错误提示
**解决方案**:
- 新增 `CaddySSLStatusHandler` 处理器，检查 Caddy 日志中的 SSL 相关错误
- 添加 `/api/caddy/ssl-status` 接口
- 在提交项目后如果启用了 SSL，会自动检查 SSL 状态
- 检测常见的 SSL 错误类型：
  - ACME 证书申请失败
  - DNS 验证失败
  - 连接超时
  - 证书申请频率限制
  - 域名验证失败
- 在项目列表中，启用了 SSL 的项目会显示警告图标
- 提供详细的错误提示和解决建议

**修改的文件**:
- `internal/api/handlers.go` - 添加 SSL 状态检查功能
- `web/static/app.js` - 添加 SSL 状态检查和错误提示
- `main.go` - 注册 SSL 状态查询路由

## 技术细节

### 状态检测改进
原来的代码使用管道命令 `netstat -ano | findstr`，在 Go 的 exec.Command 中不起作用。
新代码改为：
```go
cmd := exec.Command("netstat", "-ano")
output, err := cmd.Output()
// 然后在代码中搜索端口号
```

### SSL 错误检测
通过读取 Caddy 日志文件，搜索常见的错误关键词：
- "acme" + "error" - ACME 协议错误
- "dns" + "error" - DNS 验证问题
- "timeout" - 网络超时
- "rate limit" - 证书申请频率限制
- "unauthorized" - 授权失败

### 自动刷新机制
使用 JavaScript setInterval 在项目页面每5秒刷新一次：
```javascript
window.projectStatusInterval = setInterval(loadProjects, 5000);
```
离开项目页面时会自动清除定时器，避免资源浪费。

## 使用说明

### 编辑项目
1. 在项目列表中点击"编辑"按钮
2. 修改需要的配置
3. 点击"创建项目"按钮保存（会自动识别为更新操作）

### 查看 SSL 状态
1. 创建或编辑项目时启用 SSL
2. 保存后系统会自动检查 SSL 状态
3. 如果有错误会弹窗提示详细信息
4. 项目列表中启用 SSL 的项目会显示 ⚠️ SSL 标记

### 项目状态
- 项目状态每5秒自动更新
- 绿色徽章表示"运行中"
- 灰色徽章表示"已停止"
- 状态通过进程和端口双重检测确保准确性

## 注意事项

1. **SSL 证书申请前提条件**:
   - 域名必须正确解析到服务器
   - 服务器的 80 和 443 端口必须开放
   - 域名必须可以从公网访问
   - 不能频繁申请同一个域名的证书

2. **状态检测**:
   - 需要管理员权限运行 netstat 命令
   - Windows 系统默认支持，Linux 可能需要安装 net-tools

3. **日志文件**:
   - Caddy 日志位于 `data/caddy/caddy.log`
   - 项目日志位于 `data/logs/project_*.log`

## 构建

已重新编译项目：
```bash
go build -o caddy-manager.exe -ldflags="-H windowsgui"
```

所有更改已编译到新的可执行文件中。
