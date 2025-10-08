# Caddyfile 配置错误修复

## 🐛 发现的问题

### 1. EOF 错误
**错误信息**: `Error: adapting config using caddyfile: EOF`

**原因**: 
- 当没有项目时，`generateCaddyfileForProjects()` 生成空文件
- Caddy 无法解析空的配置文件

**解决方案**:
- 没有项目时生成默认配置
- 默认配置在 80 端口返回 "Caddy 正在运行"

### 2. 域名格式错误
**错误信息**: `Error: adapting config using caddyfile: ambiguous site definition: c808.333.606.f89f.top`

**原因**:
- 用户输入的域名格式不正确
- 包含非法字符或格式错误
- 例如: `c808.333.606.f89f.top` (数字点分格式错误)

**解决方案**:
- 添加域名验证函数 `isValidDomain()`
- 前端验证域名格式
- 后端生成 Caddyfile 时过滤无效域名

---

## ✅ 修复内容

### 1. 后端修复 (`internal/api/projects.go`)

#### 新增域名验证函数
```go
func isValidDomain(domain string) bool {
    // 验证域名格式
    // 检查长度、字符、结构等
    // 支持: example.com, subdomain.example.com, localhost, IP
    // 不支持: 包含空格、特殊字符、格式错误的域名
}
```

#### 改进 Caddyfile 生成
```go
func generateCaddyfileForProjects() error {
    // 1. 添加注释头
    // 2. 验证每个域名
    // 3. 过滤无效域名
    // 4. 没有项目时生成默认配置
}
```

**默认配置**:
```caddyfile
# Caddy 配置文件
# 由 Caddy 管理器自动生成

# 暂无项目配置
# 通过管理界面添加项目后会自动生成配置

:80 {
    respond "Caddy 正在运行" 200
}
```

### 2. 前端验证 (`web/static/app.js`)

#### 新增验证函数
```javascript
function isValidDomain(domain) {
    // 验证域名格式
    // 支持标准域名、localhost、IP地址
}
```

#### 提交前验证
```javascript
async function submitProject() {
    // 验证域名格式
    if (project.domains) {
        const domains = project.domains.split('\n')...
        const invalidDomains = [];
        
        // 检查每个域名
        for (const domain of domains) {
            if (!isValidDomain(domain)) {
                invalidDomains.push(domain);
            }
        }
        
        // 发现无效域名时提示用户
        if (invalidDomains.length > 0) {
            alert('以下域名格式不正确:\n\n' + ...);
            return;
        }
    }
    
    // SSL 配置验证
    if (project.ssl_enabled && !project.domains) {
        alert('启用 SSL 需要绑定域名');
        return;
    }
}
```

### 3. UI 改进 (`internal/api/template.go`)

添加了更详细的提示信息：
```
注意: 请使用有效的域名格式，如 example.com 或 subdomain.example.com
不支持包含特殊字符或格式错误的域名

SSL 需要: 1.有效域名 2.域名已解析 3.80/443端口开放
```

---

## 🎯 域名验证规则

### 有效域名格式
✅ `example.com`  
✅ `www.example.com`  
✅ `subdomain.example.com`  
✅ `api.v2.example.com`  
✅ `localhost`  
✅ `192.168.1.1`  
✅ `my-site.com` (连字符)  
✅ `my_site.com` (下划线)

### 无效域名格式
❌ `c808.333.606.f89f.top` (格式错误)  
❌ `my site.com` (包含空格)  
❌ `example..com` (连续点)  
❌ `-example.com` (以连字符开头)  
❌ `example-.com` (以连字符结尾)  
❌ `exa mple.com` (包含空格)  
❌ `很长的域名超过253个字符...` (超长)

---

## 🔧 验证逻辑

### 前端验证
1. 检查域名长度 (1-253 字符)
2. 检查是否包含空格或制表符
3. 允许 localhost
4. 允许 IP 地址格式
5. 使用正则表达式验证标准域名格式

### 后端验证
1. 移除端口号（如果有）
2. 检查长度限制
3. 检查非法字符
4. 验证每个标签（点分隔的部分）
5. 标签不能以连字符开头或结尾
6. 每个标签长度 1-63 字符

---

## 📝 使用建议

### 添加项目时
1. **域名格式**: 使用标准格式 `example.com`
2. **多个域名**: 每行一个，不要有空行
3. **子域名**: 支持，如 `api.example.com`
4. **测试域名**: 可以使用 `localhost`

### SSL 配置
1. **前提条件**:
   - 域名已正确解析到服务器
   - 服务器 80 和 443 端口已开放
   - 域名可以从公网访问

2. **证书邮箱**:
   - 建议填写有效邮箱
   - 用于接收证书到期通知
   - 用于紧急联系

3. **常见问题**:
   - DNS 未解析：等待 DNS 生效（最多 48 小时）
   - 端口未开放：检查防火墙和路由器设置
   - 频率限制：Let's Encrypt 有速率限制

---

## 🐛 错误处理

### Caddyfile 生成失败
**检查**:
1. 查看 `data/caddy/Caddyfile`
2. 检查域名格式是否正确
3. 手动运行 `caddy fmt` 验证格式

### Caddy 启动失败
**解决**:
1. 查看 Caddy 日志: `data/caddy/caddy.log`
2. 检查错误信息
3. 修正配置后重启 Caddy

### 域名无法访问
**排查**:
1. 检查域名 DNS 解析
2. 检查防火墙规则
3. 检查 Caddy 是否运行
4. 检查端口占用情况

---

## 📊 测试建议

### 测试用例

#### 测试 1: 有效域名
```
输入: example.com
预期: 通过验证，生成配置
```

#### 测试 2: 多个域名
```
输入:
example.com
www.example.com
api.example.com

预期: 全部通过，生成 3 个配置块
```

#### 测试 3: 无效域名
```
输入: c808.333.606.f89f.top
预期: 验证失败，提示格式错误
```

#### 测试 4: 空域名
```
输入: (空)
预期: 生成默认配置，监听 80 端口
```

#### 测试 5: localhost
```
输入: localhost
预期: 通过验证，生成配置
```

---

## 🔜 后续改进

- [ ] 支持通配符域名 (*.example.com)
- [ ] 域名 DNS 验证（检查是否解析正确）
- [ ] SSL 证书状态实时监控
- [ ] 自动修复常见域名错误
- [ ] 批量域名导入验证

---

**修复版本**: v0.0.11  
**修复日期**: 2025年  
**影响范围**: Caddyfile 生成、域名验证、SSL 配置
