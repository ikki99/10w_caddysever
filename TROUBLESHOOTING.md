# Caddy 管理器 - 故障排查指南

## 🔍 常见问题和解决方案

### 问题 1: Caddy 启动失败 - EOF 错误

**错误信息**:
```
Error: adapting config using caddyfile: EOF
```

**原因**: Caddyfile 为空或格式不正确

**解决方案**:
1. 检查是否有项目配置
2. 运行诊断脚本: `powershell -ExecutionPolicy Bypass -File diagnose.ps1`
3. 如果没有项目，系统会自动生成默认配置
4. 重启 Caddy 服务

**预防措施**:
- v0.0.11 已修复此问题
- 没有项目时会自动生成默认配置

---

### 问题 2: 域名格式错误

**错误信息**:
```
Error: adapting config using caddyfile: ambiguous site definition: c808.333.606.f89f.top
```

**原因**: 输入的域名格式不正确

**常见错误域名**:
- `c808.333.606.f89f.top` ❌ (数字点分格式错误)
- `my site.com` ❌ (包含空格)
- `-example.com` ❌ (以连字符开头)
- `example..com` ❌ (连续点号)

**正确格式**:
- `example.com` ✅
- `www.example.com` ✅
- `api.subdomain.example.com` ✅
- `localhost` ✅

**解决方案**:
1. 编辑项目，修改域名为正确格式
2. 删除无效域名的项目
3. 重新创建项目时注意域名格式
4. v0.0.11 已添加前端验证，会自动检测错误

---

### 问题 3: SSL 证书申请失败

**错误信息**:
```
{"level":"warn","msg":"stapling OCSP","error":"no OCSP stapling..."}
```

**原因**: 
- 域名未正确解析
- 端口未开放
- 证书申请频率限制

**解决方案**:

1. **检查域名解析**:
```bash
nslookup your-domain.com
ping your-domain.com
```

2. **检查端口开放**:
```bash
netstat -ano | findstr ":80"
netstat -ano | findstr ":443"
```

3. **检查防火墙**:
```bash
# 允许 80 和 443 端口
netsh advfirewall firewall add rule name="Caddy HTTP" dir=in action=allow protocol=TCP localport=80
netsh advfirewall firewall add rule name="Caddy HTTPS" dir=in action=allow protocol=TCP localport=443
```

4. **等待 DNS 生效**: 
   - DNS 解析可能需要 1-48 小时
   - 使用 `nslookup` 确认解析正确

5. **检查证书限制**:
   - Let's Encrypt 限制: 每周每域名 5 次
   - 如果超限，等待一周后重试

---

### 问题 4: 项目无法启动

**症状**: 点击启动按钮后项目状态仍为"已停止"

**排查步骤**:

1. **检查启动命令**:
   - 确保启动命令正确
   - 确保可执行文件存在
   - 确保有执行权限

2. **查看项目日志**:
   - 位置: `data/logs/project_<id>.log`
   - 查看错误信息

3. **检查端口占用**:
```bash
netstat -ano | findstr ":<端口号>"
```

4. **手动测试启动**:
   - 打开命令行
   - 进入项目目录
   - 手动执行启动命令

---

### 问题 5: 托盘图标不显示

**原因**: 
- 系统托盘设置隐藏了图标
- 使用了 `--no-tray` 参数

**解决方案**:

1. **检查系统托盘设置**:
   - Windows 10/11: 设置 → 个性化 → 任务栏
   - 选择在任务栏上显示的图标
   - 启用 Caddy 管理器

2. **检查启动参数**:
   - 不要使用 `--no-tray` 参数
   - 使用 `caddy-manager.exe` 直接启动

3. **重启程序**: 退出后重新启动

---

### 问题 6: 无法访问管理面板

**症状**: 浏览器打不开 http://localhost:8989

**排查步骤**:

1. **检查程序是否运行**:
```bash
tasklist | findstr "caddy-manager"
```

2. **检查端口占用**:
```bash
netstat -ano | findstr ":8989"
```

3. **检查防火墙**:
   - 允许程序通过防火墙
   - 或临时关闭防火墙测试

4. **尝试其他端口**:
```bash
caddy-manager.exe --port 9090
```

5. **检查浏览器**:
   - 清除缓存
   - 尝试无痕模式
   - 尝试其他浏览器

---

### 问题 7: 数据库锁定错误

**错误信息**: `database is locked`

**原因**: 多个进程同时访问数据库

**解决方案**:
1. 关闭所有 Caddy 管理器实例
2. 只启动一个实例
3. 如果问题持续，重启计算机

---

### 问题 8: Caddyfile 格式错误

**错误信息**: 
```
Caddyfile input is not formatted; run 'caddy fmt --overwrite'
```

**解决方案**:
1. 这只是警告，不影响运行
2. 自动格式化:
```bash
cd data\caddy
caddy fmt --overwrite Caddyfile
```

---

## 🛠️ 诊断工具

### 快速诊断
运行诊断脚本获取系统状态：
```bash
powershell -ExecutionPolicy Bypass -File diagnose.ps1
```

### 手动检查

#### 1. 检查 Caddyfile
```bash
type data\caddy\Caddyfile
```

#### 2. 检查 Caddy 日志
```bash
type data\caddy\caddy.log | findstr /i "error"
```

#### 3. 检查项目日志
```bash
dir data\logs\
type data\logs\project_1.log
```

#### 4. 检查进程
```bash
tasklist | findstr "caddy"
```

#### 5. 检查端口
```bash
netstat -ano | findstr ":80"
netstat -ano | findstr ":443"
netstat -ano | findstr ":8989"
```

---

## 📝 日志分析

### Caddy 日志位置
`data/caddy/caddy.log`

### 项目日志位置
`data/logs/project_<id>.log`

### 查看最新日志
```powershell
Get-Content data\caddy\caddy.log -Tail 50
```

### 查找错误
```powershell
Get-Content data\caddy\caddy.log | Select-String -Pattern "error|Error|ERROR"
```

---

## 🔧 配置文件位置

| 文件 | 位置 | 说明 |
|------|------|------|
| Caddyfile | `data/caddy/Caddyfile` | Caddy 配置文件 |
| 数据库 | `data/caddy-manager.db` | SQLite 数据库 |
| Caddy 日志 | `data/caddy/caddy.log` | Caddy 运行日志 |
| 项目日志 | `data/logs/project_*.log` | 各项目日志 |
| Caddy 程序 | `data/caddy/caddy.exe` | Caddy 可执行文件 |

---

## 💡 最佳实践

### 添加项目
1. ✅ 使用标准域名格式
2. ✅ 先测试端口访问再配置域名
3. ✅ 确保域名已解析再启用 SSL
4. ✅ 填写有效的证书邮箱

### SSL 配置
1. ✅ 确保域名已解析
2. ✅ 开放 80 和 443 端口
3. ✅ 不要频繁申请证书
4. ✅ 使用测试域名先验证

### 日常维护
1. ✅ 定期查看日志
2. ✅ 及时更新域名解析
3. ✅ 备份数据库文件
4. ✅ 关注证书到期时间

---

## 🆘 获取帮助

### 查看文档
- `README.md` - 项目介绍
- `TRAY_GUIDE.md` - 托盘功能指南
- `CADDYFILE_FIX.md` - 配置修复说明
- `TROUBLESHOOTING.md` - 本文档

### 运行诊断
```bash
powershell -ExecutionPolicy Bypass -File diagnose.ps1
```

### 查看日志
```bash
type data\caddy\caddy.log
```

### 重置配置
1. 停止 Caddy 管理器
2. 删除 `data` 目录
3. 重新启动程序
4. 重新配置

---

## 📞 联系支持

如果以上方法都无法解决问题，请：
1. 收集错误日志
2. 记录操作步骤
3. 提供系统信息
4. 联系技术支持

---

**版本**: v0.0.11  
**最后更新**: 2025年
