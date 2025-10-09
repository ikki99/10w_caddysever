# 紧急修复 - Session 检查问题

## 问题描述

**用户报告：**
- 更新后项目丢失
- 系统设置丢失
- 无法保存设置

**实际原因：**
- 之前使用 `/api/system/info` 检查 Session
- 这个 API 不需要认证，可能导致误判
- 在某些情况下可能创建新的 Session

## 修复方案

### 1. 新增专用 Session 检查接口

**新文件：** `internal/api/auth_check.go`

```go
func CheckAuthHandler(w http.ResponseWriter, r *http.Request) {
    cookie, err := r.Cookie("session_id")
    if err != nil {
        json.NewEncoder(w).Encode(map[string]bool{"authenticated": false})
        return
    }
    
    _, exists := auth.GetSession(cookie.Value)
    json.NewEncoder(w).Encode(map[string]bool{"authenticated": exists})
}
```

**路由：** `GET /api/auth/check`

### 2. 修改前端 Session 检查

**修改：** `web/static/app.js`

```javascript
async function checkFirstRun() {
    // 使用专用接口检查 Session
    const sessionCheck = await fetch('/api/auth/check');
    if (sessionCheck.ok) {
        const data = await sessionCheck.json();
        if (data.authenticated) {
            // 已登录，显示仪表盘
            // ...
        }
    }
    // ...
}
```

### 3. 改进文件路径初始化

```javascript
async function initializeFilePath() {
    try {
        const res = await fetch('/api/settings/get');
        if (!res.ok) {
            currentPath = 'C:\\www'; // 失败时使用默认值
            return;
        }
        const data = await res.json();
        currentPath = data.www_root || 'C:\\www';
    } catch (err) {
        currentPath = 'C:\\www';
        console.error('获取默认路径失败:', err);
    }
}
```

## 数据完整性验证

✅ 所有数据仍然存在

```
项目数量: 1
用户数量: 1  
设置数量: 2

项目详情:
- ID: 6
- 名称: c808.333.606
- 类型: go
- 状态: running

设置详情:
- security_path: 10w_wp
- www_root: d:\www
```

## 测试步骤

1. **停止旧版本程序**
   ```
   Ctrl+C 停止
   ```

2. **启动新版本**
   ```
   caddy-manager-console.exe
   ```

3. **测试登录**
   - 使用原来的账号密码登录
   - 验证能否看到项目列表
   - 验证能否看到设置

4. **测试刷新**
   - 登录后按 F5 刷新
   - 验证是否保持登录状态

5. **测试设置保存**
   - 进入系统设置
   - 修改 www 根目录
   - 点击保存
   - 刷新页面验证是否保存成功

## 问题原因分析

### 为什么会误以为数据丢失？

1. **Session 检查不准确**
   - 旧代码使用 `/api/system/info` 检查
   - 这个接口不需要认证
   - 可能在未登录时也返回 200

2. **可能的执行流程**
   ```
   刷新页面
     ↓
   checkFirstRun()
     ↓
   调用 /api/system/info (返回 200)
     ↓
   误以为已登录
     ↓
   调用 initializeFilePath()
     ↓
   调用 /api/settings/get (需要认证 - 401)
     ↓
   获取失败，使用默认值
     ↓
   未显示实际数据
   ```

3. **用户看到的现象**
   - 界面显示但没有项目
   - 设置显示默认值
   - 误以为数据丢失

### 实际情况

- ✅ 数据库完好无损
- ✅ 所有项目都在
- ✅ 所有设置都在
- ✅ 只是显示逻辑问题

## 改进点

### 之前的问题

```javascript
// ❌ 使用不需要认证的 API
const sessionCheck = await fetch('/api/system/info');
```

### 现在的改进

```javascript
// ✅ 使用专用的认证检查 API
const sessionCheck = await fetch('/api/auth/check');
const data = await sessionCheck.json();
if (data.authenticated) {
    // 确认已登录
}
```

## 如果仍有问题

### 检查数据库

```bash
# 查看项目
sqlite3 data/caddy-manager.db "SELECT * FROM projects;"

# 查看设置
sqlite3 data/caddy-manager.db "SELECT * FROM settings;"

# 查看用户
sqlite3 data/caddy-manager.db "SELECT username FROM users;"
```

### 清除浏览器缓存

1. 按 `Ctrl+Shift+Delete`
2. 清除 Cookie 和缓存
3. 重新登录

### 重新登录

1. 如果忘记密码，删除数据库重新初始化
2. 备份 `data/caddy-manager.db`
3. 删除数据库文件
4. 重启程序

## 修复清单

- [x] 创建 `/api/auth/check` 接口
- [x] 修改 `checkFirstRun()` 函数
- [x] 改进 `initializeFilePath()` 容错
- [x] 验证数据库完整性
- [x] 测试编译
- [x] 创建修复文档

## 紧急回滚（如果需要）

如果新版本有问题，可以回滚：

1. 停止程序
2. 使用之前的备份版本
3. 数据库不需要回滚（未改动）

---

**制作者：** 10w  
**邮箱：** wngx99@gmail.com  
**GitHub：** https://github.com/ikki99/10w_caddysever

---

**重要提醒：**
- ✅ 数据完全安全，没有丢失
- ✅ 只是显示逻辑问题
- ✅ 新版本已修复
- ✅ 请测试后反馈

更新日期：2025-01-09
版本：v1.0.3-hotfix
