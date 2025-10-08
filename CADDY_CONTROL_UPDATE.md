# Caddy 控制功能更新

## 更新内容

### 新增功能

1. **Caddy 服务控制**
   - ✅ 启动 Caddy 服务
   - ✅ 停止 Caddy 服务  
   - ✅ 重启 Caddy 服务
   - ✅ 实时状态显示

2. **用户界面增强**
   - 在顶部状态栏添加 Caddy 控制按钮
   - 根据 Caddy 运行状态动态显示不同的控制按钮
   - 运行中：显示"停止"和"重启"按钮
   - 已停止：显示"启动"按钮

### API 接口

新增以下 API 端点：

- `POST /api/caddy/start` - 启动 Caddy
- `POST /api/caddy/stop` - 停止 Caddy
- `POST /api/caddy/restart` - 重启 Caddy（已优化）

### 代码变更

#### 后端变更

1. **internal/api/handlers.go**
   ```go
   // 新增 CaddyStartHandler - 启动 Caddy
   // 新增 CaddyStopHandler - 停止 Caddy
   // 优化 CaddyRestartHandler - 返回 JSON 响应
   ```

2. **main.go**
   ```go
   // 注册新的路由
   mux.HandleFunc("/api/caddy/start", auth.AuthMiddleware(api.CaddyStartHandler))
   mux.HandleFunc("/api/caddy/stop", auth.AuthMiddleware(api.CaddyStopHandler))
   ```

#### 前端变更

1. **web/static/app.js**
   ```javascript
   // 优化 checkCaddyStatus() - 动态显示控制按钮
   // 新增 startCaddy() - 启动 Caddy
   // 新增 stopCaddy() - 停止 Caddy
   // 新增 restartCaddy() - 重启 Caddy
   ```

2. **internal/api/template.go**
   ```html
   <!-- 添加 caddy-controls 容器用于显示控制按钮 -->
   <span id="caddy-controls"></span>
   ```

### 使用说明

1. **启动 Caddy**
   - 当 Caddy 未运行时，点击顶部的"启动"按钮
   - 系统会自动启动 Caddy 服务并加载配置

2. **停止 Caddy**
   - 当 Caddy 运行中时，点击"停止"按钮
   - 会弹出确认对话框，确认后停止服务
   - **注意**：停止后所有网站将无法访问

3. **重启 Caddy**
   - 当 Caddy 运行中时，点击"重启"按钮
   - 会弹出确认对话框，确认后重启服务
   - **用途**：配置文件更新后需要重启才能生效

### 安全特性

- 所有操作都需要登录认证
- 停止和重启操作需要二次确认
- 操作失败时会显示详细的错误信息
- 操作成功后自动刷新状态显示

### 状态显示

- **运行中**：绿色文字 "Caddy 运行中"
- **未运行**：红色文字 "Caddy 未运行"
- 状态每 10 秒自动刷新一次

### 错误处理

所有操作都有完善的错误处理：

- 网络错误提示
- 服务状态检查
- 操作失败时的详细错误信息
- 按钮禁用防止重复点击

### 测试建议

1. 测试启动功能：
   - 手动停止 Caddy
   - 通过界面启动
   - 检查服务是否正常运行

2. 测试停止功能：
   - 启动 Caddy
   - 通过界面停止
   - 确认服务已停止

3. 测试重启功能：
   - 修改 Caddyfile
   - 通过界面重启
   - 验证新配置是否生效

4. 测试状态显示：
   - 观察状态是否准确
   - 控制按钮是否正确切换

### 注意事项

⚠️ **重要提示**

1. 停止 Caddy 会导致所有托管的网站暂时无法访问
2. 重启 Caddy 会短暂中断服务（通常少于 1 秒）
3. 建议在低流量时段进行重启操作
4. 配置文件错误可能导致启动失败，请检查日志

### 下一步优化建议

1. **配置验证**
   - 在重启前验证 Caddyfile 语法
   - 提供配置测试功能

2. **优雅重启**
   - 实现零停机时间的重启
   - 使用 Caddy 的 reload 功能

3. **日志查看**
   - 集成实时日志查看
   - 启动失败时自动显示错误日志

4. **状态监控**
   - 显示 Caddy 运行时长
   - 显示内存和 CPU 使用情况
   - 显示当前处理的连接数

5. **批量操作**
   - 配置文件编辑器
   - 一键备份/恢复配置
   - 配置模板功能

## 版本信息

- 更新日期：2024
- 影响文件：
  - internal/api/handlers.go
  - internal/api/template.go
  - web/static/app.js
  - main.go

## 反馈

如有问题或建议，请通过 GitHub Issues 反馈。
