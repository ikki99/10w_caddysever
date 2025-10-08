# IPv4/IPv6 兼容性问题解决方案

## 问题描述

您发现的问题：
```
dial tcp [::1]:6481: connectex: No connection could be made because the target machine actively refused it.
```

这是一个经典的 IPv4/IPv6 兼容性问题！

## 根本原因

1. **Caddy 默认行为**：
   - Caddy 使用 `localhost` 时，Windows 系统可能优先解析为 IPv6 地址 `[::1]`
   - Caddy 尝试连接 `[::1]:6481`（IPv6）

2. **应用程序监听**：
   - 大多数 Go、Node.js、Python 程序默认只监听 IPv4（`127.0.0.1`）
   - 当程序只监听 IPv4 时，IPv6 连接会被拒绝

3. **结果**：
   - Caddy → 尝试 IPv6 连接
   - 应用 → 只接受 IPv4 连接
   - 错误 → **502 Bad Gateway**

## 解决方案

### 已实现的功能

✅ **在项目配置中新增"代理连接方式"选项**

位置：新建/编辑项目 → 步骤 3（域名和SSL）→ 代理连接方式

选项说明：
- **IPv4 (127.0.0.1) - 推荐** ⭐
  - 强制 Caddy 使用 IPv4 地址连接
  - 避免 IPv6 兼容性问题
  - 适用于大多数应用程序
  - Caddyfile 生成：`reverse_proxy 127.0.0.1:端口`

- **localhost (可能IPv6)**
  - 让系统自动选择 IPv4 或 IPv6
  - 可能导致连接问题
  - 仅在明确支持 IPv6 时使用
  - Caddyfile 生成：`reverse_proxy localhost:端口`

### 使用指南

#### 新建项目时

1. 进入"新建项目"
2. 在步骤 3（域名和SSL配置）
3. 找到"代理连接方式"选项
4. 选择"IPv4 (127.0.0.1) - 推荐"（默认已选中）
5. 继续完成配置

#### 修复现有项目

如果您的项目出现 502 错误：

1. 编辑项目
2. 跳转到步骤 3
3. 将"代理连接方式"改为"IPv4 (127.0.0.1)"
4. 保存更改
5. Caddy 会自动重新加载配置

### 技术细节

#### Caddyfile 配置对比

**IPv4 模式（推荐）：**
```caddyfile
example.com {
    reverse_proxy 127.0.0.1:6481
}
```

**localhost 模式（可能有问题）：**
```caddyfile
example.com {
    reverse_proxy localhost:6481
}
```

区别：
- `127.0.0.1` - 明确指定 IPv4 地址
- `localhost` - 依赖系统 DNS 解析，可能返回 `::1`（IPv6）

#### 数据库字段

新增字段：`use_ipv4` (BOOLEAN, 默认值: 1)
- `true` (1) - 使用 IPv4（127.0.0.1）
- `false` (0) - 使用 localhost

#### 代码实现

**数据库迁移**：
```sql
ALTER TABLE projects ADD COLUMN use_ipv4 BOOLEAN DEFAULT 1;
```

**Caddyfile 生成逻辑**：
```go
var proxyTarget string
if useIPv4 {
    // 强制使用 IPv4 地址
    proxyTarget = fmt.Sprintf("127.0.0.1:%d", port)
} else {
    // 使用 localhost（可能 IPv6）
    proxyTarget = fmt.Sprintf("localhost:%d", port)
}
content += fmt.Sprintf("    reverse_proxy %s\n", proxyTarget)
```

## 常见问题

### Q1: 为什么默认选择 IPv4？

**A:** 因为：
1. 大多数应用程序默认只监听 IPv4
2. IPv4 兼容性更好
3. 避免不必要的连接问题

### Q2: 什么时候应该使用 localhost？

**A:** 仅在以下情况：
- 您的应用程序明确配置为监听 IPv6
- 例如 Go 程序使用 `[::]:端口` 或 `:端口`
- 或 Node.js 使用 `::` 作为 host

### Q3: 如何检查我的应用监听的是 IPv4 还是 IPv6？

**A:** 使用命令：
```powershell
netstat -ano | findstr ":6481"
```

查看输出：
- `127.0.0.1:6481` - 仅 IPv4
- `[::]:6481` 或 `0.0.0.0:6481` - 同时支持 IPv4 和 IPv6
- `[::1]:6481` - 仅 IPv6

### Q4: 我的程序如何同时支持 IPv4 和 IPv6？

**A:** 取决于编程语言：

**Go:**
```go
// 监听所有接口（IPv4 + IPv6）
http.ListenAndServe(":6481", handler)

// 或明确指定
http.ListenAndServe("0.0.0.0:6481", handler)  // 仅 IPv4
http.ListenAndServe("[::]:6481", handler)     // IPv4 + IPv6
```

**Node.js:**
```javascript
// 监听所有接口
app.listen(6481);

// 或明确指定
app.listen(6481, '0.0.0.0');  // 仅 IPv4
app.listen(6481, '::');       // IPv4 + IPv6
```

**Python (Flask):**
```python
# 监听所有接口
app.run(host='0.0.0.0', port=6481)  # 仅 IPv4

# IPv4 + IPv6 需要额外配置
app.run(host='::', port=6481)
```

### Q5: 旧项目会自动更新吗？

**A:** 会！
- 数据库迁移会自动添加 `use_ipv4` 字段
- 默认值为 `true`（使用 IPv4）
- 所有旧项目自动获得 IPv4 兼容性

## 故障排查

### 症状 1：502 Bad Gateway

**可能原因：**
- Caddy 使用 IPv6，但应用只监听 IPv4

**解决方案：**
1. 编辑项目
2. 设置"代理连接方式"为"IPv4"
3. 保存并重新加载 Caddy

### 症状 2：日志显示 IPv6 连接错误

**日志内容：**
```
dial tcp [::1]:端口: connectex: No connection could be made...
```

**解决方案：**
同上，强制使用 IPv4

### 症状 3：本地可访问，Caddy 代理不行

**检查步骤：**
1. 确认应用在运行：`netstat -ano | findstr ":端口"`
2. 检查监听地址（IPv4 vs IPv6）
3. 修改 Caddy 配置使用正确的协议

## 最佳实践

### 推荐配置

1. **新项目**：
   - 默认使用 IPv4（已自动配置）
   - 除非有特殊需求，不要更改

2. **现有项目**：
   - 如遇到 502 错误，立即切换到 IPv4
   - 重新生成 Caddyfile

3. **应用程序开发**：
   - 推荐监听 `0.0.0.0`（IPv4）或 `[::]`（双栈）
   - 明确指定监听地址，避免歧义

### 配置检查清单

部署新项目前，确认：
- [ ] 应用程序正在运行
- [ ] 使用 `netstat` 确认监听地址
- [ ] Caddy 配置匹配应用监听的协议
- [ ] 测试反向代理是否工作

## 示例场景

### 场景 1：Go Web 应用

**应用代码：**
```go
package main

import "net/http"

func main() {
    // 默认监听 0.0.0.0:8080（IPv4）
    http.ListenAndServe(":8080", nil)
}
```

**Caddy 配置：**
- 选择"IPv4" ✅
- 生成：`reverse_proxy 127.0.0.1:8080`

### 场景 2：Node.js Express 应用

**应用代码：**
```javascript
const express = require('express');
const app = express();

// 默认监听 IPv4
app.listen(3000);
```

**Caddy 配置：**
- 选择"IPv4" ✅
- 生成：`reverse_proxy 127.0.0.1:3000`

### 场景 3：支持双栈的应用

**应用代码：**
```go
// 明确监听 IPv4 和 IPv6
http.ListenAndServe("[::]:9000", nil)
```

**Caddy 配置：**
- 可以选择"localhost" 或 "IPv4"
- 两者都可以工作

## 总结

- ✅ **默认使用 IPv4** - 最安全、最兼容
- ⚠️ **遇到 502** - 检查 IPv4/IPv6 配置
- 📝 **新功能** - 可选择代理连接方式
- 🔧 **自动迁移** - 旧项目自动支持

这个功能已经集成到系统中，开箱即用！

---

**相关文件：**
- `internal/database/database.go` - 数据库架构
- `internal/models/models.go` - 数据模型
- `internal/api/projects.go` - API 和 Caddyfile 生成
- `internal/api/template.go` - UI 界面
- `web/static/app.js` - 前端逻辑
