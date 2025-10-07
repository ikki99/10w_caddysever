package api

const IndexTemplate = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Caddy 管理器</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body { font-family: 'Microsoft YaHei', Arial, sans-serif; background: #f5f7fa; }
        .container { max-width: 1200px; margin: 0 auto; padding: 20px; }
        .header { background: #fff; padding: 20px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); margin-bottom: 20px; }
        .header h1 { color: #2c3e50; font-size: 24px; }
        .card { background: #fff; padding: 20px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); margin-bottom: 20px; }
        .project-card { background: #fff; padding: 20px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); margin-bottom: 15px; border-left: 4px solid #409EFF; }
        .project-card.running { border-left-color: #67C23A; }
        .project-card.stopped { border-left-color: #909399; }
        .btn { padding: 10px 20px; border: none; border-radius: 4px; cursor: pointer; font-size: 14px; margin-right: 10px; }
        .btn-primary { background: #409EFF; color: #fff; }
        .btn-primary:hover { background: #66b1ff; }
        .btn-success { background: #67C23A; color: #fff; }
        .btn-danger { background: #F56C6C; color: #fff; }
        .btn-warning { background: #E6A23C; color: #fff; }
        .btn-sm { padding: 6px 12px; font-size: 12px; }
        input[type="text"], input[type="password"], input[type="number"], select, textarea { width: 100%; padding: 10px; border: 1px solid #dcdfe6; border-radius: 4px; margin-bottom: 10px; font-size: 14px; }
        textarea { resize: vertical; min-height: 80px; font-family: 'Consolas', monospace; }
        .form-group { margin-bottom: 15px; }
        .form-row { display: grid; grid-template-columns: 1fr 1fr; gap: 15px; }
        label { display: block; margin-bottom: 5px; color: #606266; font-weight: 500; }
        #login-page, #setup-page, #dashboard, #system-info { display: none; }
        .tabs { display: flex; border-bottom: 2px solid #e4e7ed; margin-bottom: 20px; }
        .tab { padding: 10px 20px; cursor: pointer; border-bottom: 2px solid transparent; margin-bottom: -2px; }
        .tab.active { color: #409EFF; border-bottom-color: #409EFF; }
        .tab-content { display: none; }
        .tab-content.active { display: block; }
        .status-badge { display: inline-block; padding: 4px 12px; border-radius: 12px; font-size: 12px; font-weight: 600; }
        .status-running { background: #67C23A; color: white; }
        .status-stopped { background: #909399; color: white; }
        .status-error { background: #F56C6C; color: white; }
        .info-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(250px, 1fr)); gap: 15px; margin-bottom: 20px; }
        .info-item { padding: 15px; background: #f9f9f9; border-radius: 4px; border-left: 4px solid #409EFF; }
        .info-item h4 { color: #606266; margin-bottom: 5px; }
        .info-item p { font-size: 18px; color: #2c3e50; font-weight: 600; }
        .env-status { display: inline-block; padding: 3px 10px; border-radius: 12px; font-size: 12px; }
        .env-installed { background: #67C23A; color: white; }
        .env-not-installed { background: #909399; color: white; }
        .file-item { padding: 10px; border-bottom: 1px solid #eee; display: flex; justify-content: space-between; align-items: center; cursor: pointer; }
        .file-item:hover { background: #f5f7fa; }
        .file-icon { margin-right: 10px; font-size: 20px; }
        small { color: #909399; font-size: 12px; }
        hr { margin: 20px 0; border: none; border-top: 1px solid #ebeef5; }
        .modal { display: none; position: fixed; top: 0; left: 0; width: 100%; height: 100%; background: rgba(0,0,0,0.5); z-index: 1000; overflow-y: auto; }
        .modal-content { position: relative; background: white; width: 90%; max-width: 800px; margin: 50px auto; padding: 30px; border-radius: 8px; }
        .modal-close { position: absolute; right: 15px; top: 15px; font-size: 28px; cursor: pointer; color: #909399; }
        .breadcrumb { padding: 10px; background: #f5f7fa; border-radius: 4px; margin-bottom: 15px; }
        .project-info { display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 15px; }
        .project-details { flex: 1; }
        .project-actions { display: flex; gap: 5px; }
        .wizard-steps { display: flex; justify-content: space-between; margin-bottom: 30px; }
        .wizard-step { flex: 1; text-align: center; padding: 10px; position: relative; }
        .wizard-step.active { color: #409EFF; font-weight: 600; }
        .wizard-step.active::after { content: ''; position: absolute; bottom: 0; left: 0; right: 0; height: 2px; background: #409EFF; }
        .step-content { display: none; }
        .step-content.active { display: block; }
    </style>
</head>
<body>
    <!-- 设置页面 -->
    <div id="setup-page">
        <div class="container" style="max-width: 500px; margin-top: 100px;">
            <div class="card">
                <h2 style="text-align: center; margin-bottom: 20px;">欢迎使用 Caddy 管理器</h2>
                <p style="text-align: center; color: #909399; margin-bottom: 20px;">请创建管理员账户</p>
                <form id="setup-form">
                    <div class="form-group">
                        <label>用户名</label>
                        <input type="text" id="setup-username" required>
                    </div>
                    <div class="form-group">
                        <label>密码</label>
                        <input type="password" id="setup-password" required>
                    </div>
                    <div class="form-group">
                        <label>确认密码</label>
                        <input type="password" id="setup-password2" required>
                    </div>
                    <button type="submit" class="btn btn-primary" style="width: 100%;">完成设置</button>
                </form>
            </div>
        </div>
    </div>

    <!-- 登录页面 -->
    <div id="login-page">
        <div class="container" style="max-width: 400px; margin-top: 100px;">
            <div class="card">
                <h2 style="text-align: center; margin-bottom: 20px;">Caddy 管理器</h2>
                <form id="login-form">
                    <div class="form-group">
                        <label>用户名</label>
                        <input type="text" id="login-username" required>
                    </div>
                    <div class="form-group">
                        <label>密码</label>
                        <input type="password" id="login-password" required>
                    </div>
                    <button type="submit" class="btn btn-primary" style="width: 100%;">登录</button>
                </form>
            </div>
        </div>
    </div>

    <div id="dashboard">
        <div class="container">
            <div class="header">
                <div style="display: flex; justify-content: space-between; align-items: center;">
                    <h1>🖥️ Caddy 管理器</h1>
                    <div>
                        <span id="caddy-status">状态检查中...</span>
                        <button class="btn btn-primary" onclick="logout()">退出</button>
                    </div>
                </div>
            </div>

            <div class="tabs">
                <div class="tab active" onclick="switchTab('dashboard')">仪表盘</div>
                <div class="tab" onclick="switchTab('projects')">项目管理</div>
                <div class="tab" onclick="switchTab('tasks')">计划任务</div>
                <div class="tab" onclick="switchTab('files')">文件管理</div>
                <div class="tab" onclick="switchTab('logs')">运行日志</div>
                <div class="tab" onclick="switchTab('env')">环境部署</div>
                <div class="tab" onclick="switchTab('settings')">系统设置</div>
            </div>

            <!-- 仪表盘 -->
            <div id="dashboard-tab" class="tab-content active">
                <div class="card">
                    <h3 style="margin-bottom: 20px;">系统信息</h3>
                    <div class="info-grid" id="sys-info-grid"></div>
                </div>
                <div class="card">
                    <h3 style="margin-bottom: 20px;">运行环境</h3>
                    <div id="env-info-list"></div>
                </div>
            </div>

            <!-- 项目管理 -->
            <div id="projects-tab" class="tab-content">
                <div class="card">
                    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px;">
                        <h3>项目列表</h3>
                        <button class="btn btn-primary" onclick="showAddProject()">+ 新建项目</button>
                    </div>
                    <div id="project-list"></div>
                </div>
            </div>

            <!-- 计划任务 -->
            <div id="tasks-tab" class="tab-content">
                <div class="card">
                    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px;">
                        <h3>计划任务</h3>
                        <button class="btn btn-primary" onclick="showAddTask()">+ 新建任务</button>
                    </div>
                    <div id="task-list"></div>
                </div>
            </div>

            <!-- 站点管理 -->
            <div id="sites-tab" class="tab-content">
                <div class="card">
                    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px;">
                        <h3>站点列表</h3>
                        <button class="btn btn-primary" onclick="showAddSite()">+ 添加站点</button>
                    </div>
                    <ul class="site-list" id="site-list"></ul>
                </div>
            </div>

            <div id="files-tab" class="tab-content">
                <div class="card">
                    <h3 style="margin-bottom: 20px;">文件管理器</h3>
                    <div class="breadcrumb" id="file-breadcrumb"></div>
                    <div style="margin-bottom: 15px;">
                        <input type="file" id="file-upload">
                        <button class="btn btn-primary" onclick="uploadFileToPath()">上传文件</button>
                        <button class="btn btn-success" onclick="createNewFolder()">新建文件夹</button>
                    </div>
                    <div id="file-browser"></div>
                </div>
            </div>

            <div id="logs-tab" class="tab-content">
                <div class="card">
                    <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px;">
                        <h3>Caddy 运行日志</h3>
                        <button class="btn btn-primary" onclick="refreshLogs()">刷新日志</button>
                    </div>
                    <div id="log-content" style="background: #2c3e50; color: #ecf0f1; padding: 15px; border-radius: 4px; font-family: 'Consolas', monospace; font-size: 12px; max-height: 500px; overflow-y: auto; white-space: pre-wrap;"></div>
                </div>
            </div>

            <div id="env-tab" class="tab-content">
                <div class="card">
                    <h3 style="margin-bottom: 20px;">运行环境部署</h3>
                    <div id="env-list"></div>
                </div>
            </div>

            <div id="settings-tab" class="tab-content">
                <div class="card">
                    <h3 style="margin-bottom: 20px;">系统设置</h3>
                    <div class="form-group">
                        <label>安全访问路径（类似宝塔面板）</label>
                        <input type="text" id="security-path" placeholder="留空表示不启用，例如：admin123">
                        <small>设置后访问地址变为: http://localhost:8989/yourpath</small>
                    </div>
                    <div class="form-group">
                        <label>网站根目录</label>
                        <input type="text" id="www-root" placeholder="C:\\www">
                        <small>所有网站文件的默认存放目录</small>
                    </div>
                    <button class="btn btn-primary" onclick="saveSettings()">保存设置</button>
                    <hr>
                    <h3 style="margin-bottom: 20px;">修改管理员密码</h3>
                    <div class="form-group">
                        <label>用户名</label>
                        <input type="text" id="change-username">
                    </div>
                    <div class="form-group">
                        <label>旧密码</label>
                        <input type="password" id="old-password">
                    </div>
                    <div class="form-group">
                        <label>新密码</label>
                        <input type="password" id="new-password">
                    </div>
                    <div class="form-group">
                        <label>确认新密码</label>
                        <input type="password" id="new-password2">
                    </div>
                    <button class="btn btn-success" onclick="changePassword()">修改密码</button>
                </div>
            </div>
        </div>
    </div>

    <!-- 添加项目模态框 -->
    <div id="add-project-modal" class="modal">
        <div class="modal-content">
            <span class="modal-close" onclick="closeModal('add-project-modal')">&times;</span>
            <h2 style="margin-bottom: 20px;">新建项目</h2>
            
            <div class="wizard-steps">
                <div class="wizard-step active" data-step="1">1. 选择类型</div>
                <div class="wizard-step" data-step="2">2. 基本配置</div>
                <div class="wizard-step" data-step="3">3. 域名配置</div>
                <div class="wizard-step" data-step="4">4. 高级选项</div>
            </div>

            <div class="step-content active" data-step="1">
                <div class="form-group">
                    <label>项目类型</label>
                    <select id="proj-type" onchange="onProjectTypeChange()">
                        <option value="">请选择</option>
                        <option value="go">Go 项目</option>
                        <option value="python">Python 项目</option>
                        <option value="nodejs">Node.js 项目</option>
                        <option value="java">Java 项目</option>
                        <option value="php">PHP 站点</option>
                        <option value="static">静态网站</option>
                    </select>
                </div>
                <button class="btn btn-primary" onclick="nextStep(2)">下一步</button>
            </div>

            <div class="step-content" data-step="2">
                <div class="form-group">
                    <label>项目名称 *</label>
                    <input type="text" id="proj-name" placeholder="我的API项目">
                </div>
                <div class="form-group">
                    <label>项目根目录 *</label>
                    <input type="text" id="proj-root" placeholder="C:\\projects\\myapi">
                    <small>项目文件所在目录</small>
                </div>
                <div class="form-group" id="exec-path-group" style="display:none;">
                    <label>可执行文件路径</label>
                    <input type="text" id="proj-exec" placeholder="C:\\projects\\myapi\\myapi.exe">
                    <small>编译后的可执行文件（Go/Java）</small>
                </div>
                <div class="form-group">
                    <label>启动命令</label>
                    <input type="text" id="proj-cmd" placeholder="./myapp.exe">
                    <small id="cmd-hint">根据项目类型输入启动命令</small>
                </div>
                <div class="form-row">
                    <div class="form-group">
                        <label>监听端口 *</label>
                        <input type="number" id="proj-port" placeholder="8080">
                    </div>
                    <div class="form-group">
                        <label>开机自启</label>
                        <select id="proj-autostart">
                            <option value="false">否</option>
                            <option value="true">是</option>
                        </select>
                    </div>
                </div>
                <div class="form-group">
                    <label>项目说明</label>
                    <textarea id="proj-desc" rows="3" placeholder="可选的项目描述信息"></textarea>
                </div>
                <button class="btn" onclick="prevStep(1)">上一步</button>
                <button class="btn btn-primary" onclick="nextStep(3)">下一步</button>
            </div>

            <div class="step-content" data-step="3">
                <div class="form-group">
                    <label>绑定域名</label>
                    <textarea id="proj-domains" rows="4" placeholder="example.com&#10;www.example.com&#10;每行一个域名"></textarea>
                    <small>留空则不绑定域名，只能通过端口访问</small>
                </div>
                <div class="form-row">
                    <div class="form-group">
                        <label>启用 SSL</label>
                        <select id="proj-ssl">
                            <option value="true">是（自动申请证书）</option>
                            <option value="false">否</option>
                        </select>
                    </div>
                    <div class="form-group">
                        <label>证书邮箱</label>
                        <input type="text" id="proj-email" placeholder="admin@example.com">
                    </div>
                </div>
                <button class="btn" onclick="prevStep(2)">上一步</button>
                <button class="btn btn-primary" onclick="nextStep(4)">下一步</button>
            </div>

            <div class="step-content" data-step="4">
                <div class="form-group">
                    <label>反向代理路径</label>
                    <input type="text" id="proj-proxy-path" value="/" placeholder="/">
                    <small>默认 / 代理所有请求</small>
                </div>
                <div class="form-group">
                    <label>额外 Header</label>
                    <textarea id="proj-headers" rows="3" placeholder="X-Real-IP {remote_host}&#10;X-Forwarded-For {remote_host}"></textarea>
                    <small>每行一个 Header</small>
                </div>
                <div style="text-align: right; margin-top: 20px;">
                    <button class="btn" onclick="prevStep(3)">上一步</button>
                    <button class="btn btn-success" onclick="submitProject()">创建项目</button>
                    <button class="btn" onclick="closeModal('add-project-modal')">取消</button>
                </div>
            </div>
        </div>
    </div>

    <!-- 项目日志模态框 -->
    <div id="project-logs-modal" class="modal">
        <div class="modal-content">
            <span class="modal-close" onclick="closeModal('project-logs-modal')">&times;</span>
            <h2 id="log-title" style="margin-bottom: 20px;">项目日志</h2>
            <div style="background: #2c3e50; color: #ecf0f1; padding: 15px; border-radius: 4px; font-family: 'Consolas', monospace; font-size: 12px; max-height: 500px; overflow-y: auto; white-space: pre-wrap;" id="project-log-content"></div>
            <div style="text-align: right; margin-top: 15px;">
                <button class="btn btn-primary" onclick="refreshProjectLogs()">刷新</button>
                <button class="btn" onclick="closeModal('project-logs-modal')">关闭</button>
            </div>
        </div>
    </div>

    <!-- 添加任务模态框 -->
    <div id="add-task-modal" class="modal">
        <div class="modal-content">
            <span class="modal-close" onclick="closeModal('add-task-modal')">&times;</span>
            <h2 style="margin-bottom: 20px;">新建任务</h2>
            <div class="form-group">
                <label>任务名称 *</label>
                <input type="text" id="task-name" placeholder="每日备份">
            </div>
            <div class="form-group">
                <label>执行命令 *</label>
                <input type="text" id="task-cmd" placeholder="backup.bat">
            </div>
            <div class="form-group">
                <label>执行时间</label>
                <input type="text" id="task-schedule" placeholder="0 2 * * * 或 每天 02:00">
                <small>支持 cron 表达式</small>
            </div>
            <div class="form-group">
                <label>循环执行</label>
                <select id="task-loop">
                    <option value="true">是</option>
                    <option value="false">否（只执行一次）</option>
                </select>
            </div>
            <div style="text-align: right;">
                <button class="btn btn-primary" onclick="submitTask()">创建</button>
                <button class="btn" onclick="closeModal('add-task-modal')">取消</button>
            </div>
        </div>
    </div>

    <div id="add-site-modal" class="modal">
        <div class="modal-content">
            <span class="modal-close" onclick="closeModal('add-site-modal')">&times;</span>
            <h2 style="margin-bottom: 20px;">添加站点</h2>
            <div class="form-group">
                <label>域名</label>
                <input type="text" id="site-domain" placeholder="example.com">
            </div>
            <div class="form-group">
                <label>站点类型</label>
                <select id="site-type" onchange="onSiteTypeChange()">
                    <option value="static">静态站点</option>
                    <option value="proxy">反向代理</option>
                    <option value="php">PHP站点</option>
                </select>
            </div>
            <div class="form-group" id="target-group">
                <label id="target-label">网站目录</label>
                <input type="text" id="site-target" placeholder="C:\\www\\mysite">
            </div>
            <div class="form-group" id="env-group" style="display:none;">
                <label>PHP版本</label>
                <select id="php-version">
                    <option value="7.4">PHP 7.4</option>
                    <option value="8.0">PHP 8.0</option>
                    <option value="8.1">PHP 8.1</option>
                    <option value="8.2">PHP 8.2</option>
                </select>
            </div>
            <div style="text-align: right; margin-top: 20px;">
                <button class="btn btn-primary" onclick="submitAddSite()">添加</button>
                <button class="btn" onclick="closeModal('add-site-modal')">取消</button>
            </div>
        </div>
    </div>

    <div id="env-guide-modal" class="modal">
        <div class="modal-content">
            <span class="modal-close" onclick="closeModal('env-guide-modal')">&times;</span>
            <h2 id="guide-title" style="margin-bottom: 20px;"></h2>
            <pre id="guide-steps" style="background: #f5f7fa; padding: 15px; border-radius: 4px; white-space: pre-wrap;"></pre>
            <div style="text-align: right; margin-top: 20px;">
                <button class="btn btn-primary" id="guide-download-btn">打开下载页</button>
                <button class="btn" onclick="closeModal('env-guide-modal')">关闭</button>
            </div>
        </div>
    </div>

    <script src="/static/app.js"></script>
</body>
</html>`

