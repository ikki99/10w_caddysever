// Caddy Manager - Complete Version
let currentPath = '';
let currentEnvs = [];
let currentProjectId = null;
let currentStep = 1;

window.onload = function() {
    checkFirstRun();
};

async function checkFirstRun() {
    // 先检查是否有有效的 Session
    try {
        const sessionCheck = await fetch('/api/auth/check');
        if (sessionCheck.ok) {
            const data = await sessionCheck.json();
            if (data.authenticated) {
                // 已登录，直接显示仪表盘
                document.getElementById('dashboard').style.display = 'block';
                initializeFilePath();
                loadSystemInfo();
                loadProjects();
                checkCaddyStatus();
                setInterval(checkCaddyStatus, 10000);
                return;
            }
        }
    } catch (err) {
        // Session 无效或过期，继续检查是否首次运行
        console.log('Session check failed:', err);
    }
    
    // 检查是否首次运行
    const res = await fetch('/api/setup');
    const data = await res.json();
    if (data.firstRun) {
        document.getElementById('setup-page').style.display = 'block';
    } else {
        document.getElementById('login-page').style.display = 'block';
    }
}

// 设置表单
document.getElementById('setup-form')?.addEventListener('submit', async (e) => {
    e.preventDefault();
    const username = document.getElementById('setup-username').value;
    const password = document.getElementById('setup-password').value;
    const password2 = document.getElementById('setup-password2').value;
    
    if (password !== password2) {
        alert('两次密码不一致');
        return;
    }

    const res = await fetch('/api/setup', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password })
    });

    if (res.ok) {
        alert('设置完成，请登录');
        document.getElementById('setup-page').style.display = 'none';
        document.getElementById('login-page').style.display = 'block';
    }
});

// 登录表单
document.getElementById('login-form')?.addEventListener('submit', async (e) => {
    e.preventDefault();
    const username = document.getElementById('login-username').value;
    const password = document.getElementById('login-password').value;

    const res = await fetch('/api/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password })
    });

    if (res.ok) {
        document.getElementById('login-page').style.display = 'none';
        document.getElementById('dashboard').style.display = 'block';
        
        // 初始化 currentPath 为默认路径
        initializeFilePath();
        
        loadSystemInfo();
        loadProjects();
        checkCaddyStatus();
        setInterval(checkCaddyStatus, 10000);
    } else {
        alert('登录失败：用户名或密码错误');
    }
});

// 加载系统信息
async function loadSystemInfo() {
    try {
        const res = await fetch('/api/system/info');
        const data = await res.json();
        
        const grid = document.getElementById('sys-info-grid');
        grid.innerHTML = '<div class="info-item"><h4>操作系统</h4><p>' + data.os + '</p></div>' +
            '<div class="info-item"><h4>架构</h4><p>' + data.arch + '</p></div>' +
            '<div class="info-item"><h4>CPU核心</h4><p>' + data.cpu_cores + ' 核</p></div>';
        
        const envList = document.getElementById('env-info-list');
        currentEnvs = data.environments;
        envList.innerHTML = data.environments.map(env => 
            '<div class="file-item"><div><strong>' + env.name + '</strong> ' +
            '<span class="env-status ' + (env.installed ? 'env-installed' : 'env-not-installed') + '">' +
            (env.installed ? '已安装 ' + env.version : '未安装') + '</span></div></div>'
        ).join('');
    } catch (err) {
        console.error('加载系统信息失败:', err);
    }
}

// ===== 项目管理 =====

async function loadProjects() {
    const res = await fetch('/api/projects');
    const projects = await res.json();
    const list = document.getElementById('project-list');
    
    if (!projects || projects.length === 0) {
        list.innerHTML = '<p style="text-align:center;color:#909399;padding:40px;">暂无项目，点击"新建项目"开始部署</p>';
        return;
    }
    
    list.innerHTML = projects.map(p => {
        // SSL 状态检测 - 仅在启用 SSL 且有域名时显示检查按钮
        let sslStatus = '';
        if (p.ssl_enabled && p.domains) {
            sslStatus = '<button class="btn-link" style="color:#409EFF;font-size:12px;margin-left:10px;" onclick="checkProjectSSL(\'' + (p.domains.split('\\n')[0] || p.domains) + '\')">🔍 检查SSL</button>';
        } else if (p.ssl_enabled && !p.domains) {
            sslStatus = '<small style="color:#F56C6C;margin-left:10px;">⚠️ 未配置域名</small>';
        }
        
        return '<div class="project-card ' + p.status + '">' +
        '<div class="project-info">' +
        '<div class="project-details">' +
        '<h3>' + p.name + ' <span class="status-badge status-' + p.status + '">' + (p.status === 'running' ? '运行中' : '已停止') + '</span>' + sslStatus + '</h3>' +
        '<p style="color:#606266;margin:5px 0;"><strong>类型:</strong> ' + getProjectTypeName(p.project_type) + ' | <strong>端口:</strong> ' + p.port + ' | <strong>域名:</strong> ' + (p.domains || '无') + '</p>' +
        (p.description ? '<p style="color:#909399;font-size:13px;">' + p.description + '</p>' : '') +
        '</div>' +
        '<div class="project-actions">' +
        (p.status === 'running' ? 
            '<button class="btn btn-warning btn-sm" onclick="stopProject(' + p.id + ')">停止</button>' +
            '<button class="btn btn-primary btn-sm" onclick="restartProject(' + p.id + ')">重启</button>' :
            '<button class="btn btn-success btn-sm" onclick="startProject(' + p.id + ')">启动</button>') +
        '<button class="btn btn-primary btn-sm" onclick="editProject(' + p.id + ')">编辑</button>' +
        '<button class="btn btn-danger btn-sm" onclick="deleteProject(' + p.id + ')">删除</button>' +
        '</div></div></div>';
    }).join('');
}

function getProjectTypeName(type) {
    const types = {'go': 'Go', 'python': 'Python', 'nodejs': 'Node.js', 'java': 'Java', 'php': 'PHP', 'static': '静态站点'};
    return types[type] || type;
}

function showAddProject() {
    currentProjectId = null;
    currentStep = 1;
    clearProjectForm();
    document.getElementById('modal-title').textContent = '新建项目';
    document.getElementById('submit-project-btn').textContent = '创建项目';
    document.getElementById('add-project-modal').style.display = 'block';
    updateWizardSteps();
}

async function editProject(id) {
    currentProjectId = id;
    currentStep = 1;
    
    // 获取项目详情
    const res = await fetch('/api/projects');
    const projects = await res.json();
    const project = projects.find(p => p.id === id);
    
    if (!project) {
        alert('项目不存在');
        return;
    }
    
    // 填充表单
    document.getElementById('proj-type').value = project.project_type;
    document.getElementById('proj-name').value = project.name;
    document.getElementById('proj-root').value = project.root_dir;
    document.getElementById('proj-exec').value = project.exec_path || '';
    document.getElementById('proj-cmd').value = project.start_command || '';
    document.getElementById('proj-port').value = project.port;
    document.getElementById('proj-autostart').value = project.auto_start ? 'true' : 'false';
    document.getElementById('proj-domains').value = project.domains || '';
    document.getElementById('proj-ssl').value = project.ssl_enabled ? 'true' : 'false';
    document.getElementById('proj-email').value = project.ssl_email || '';
    document.getElementById('proj-proxy-path').value = project.reverse_proxy_path || '/';
    document.getElementById('proj-headers').value = project.extra_headers || '';
    document.getElementById('proj-desc').value = project.description || '';
    
    // 设置 IPv4 选项（默认为 true）
    const useIPv4El = document.getElementById('proj-use-ipv4');
    if (useIPv4El) {
        useIPv4El.value = (project.use_ipv4 !== false) ? 'true' : 'false';
    }
    
    onProjectTypeChange();
    document.getElementById('modal-title').textContent = '编辑项目 - ' + project.name;
    document.getElementById('submit-project-btn').textContent = '保存修改';
    document.getElementById('add-project-modal').style.display = 'block';
    updateWizardSteps();
}

function onProjectTypeChange() {
    const type = document.getElementById('proj-type').value;
    document.getElementById('exec-path-group').style.display = (type === 'go' || type === 'java') ? 'block' : 'none';
    // ������̬վ�㣬Ĭ�ϲ��� SSL ������ʾ��Ͷ���
    if (type === 'static') {
        const sslEl = document.getElementById('proj-ssl');
        const domainsEl = document.getElementById('proj-domains');
        if (sslEl) sslEl.value = 'false';
        if (domainsEl) domainsEl.value = '';
    }
}

function nextStep(step) {
    if (step === 2 && !document.getElementById('proj-type').value) {
        alert('请选择项目类型');
        return;
    }
    if (step === 3) {
        if (!document.getElementById('proj-name').value || !document.getElementById('proj-root').value || !document.getElementById('proj-port').value) {
            alert('请填写必填项');
            return;
        }
    }
    currentStep = step;
    updateWizardSteps();
}

function prevStep(step) {
    currentStep = step;
    updateWizardSteps();
}

function updateWizardSteps() {
    document.querySelectorAll('.wizard-step').forEach(el => {
        el.classList.toggle('active', parseInt(el.dataset.step) === currentStep);
    });
    document.querySelectorAll('.step-content').forEach(el => {
        el.classList.toggle('active', parseInt(el.dataset.step) === currentStep);
    });
}

async function submitProject() {
    const project = {
        name: document.getElementById('proj-name').value,
        project_type: document.getElementById('proj-type').value,
        root_dir: document.getElementById('proj-root').value,
        exec_path: document.getElementById('proj-exec').value,
        port: parseInt(document.getElementById('proj-port').value),
        start_command: document.getElementById('proj-cmd').value,
        auto_start: document.getElementById('proj-autostart').value === 'true',
        domains: document.getElementById('proj-domains').value,
        ssl_enabled: document.getElementById('proj-ssl').value === 'true',
        ssl_email: document.getElementById('proj-email').value,
        reverse_proxy_path: document.getElementById('proj-proxy-path').value || '/',
        extra_headers: document.getElementById('proj-headers').value,
        description: document.getElementById('proj-desc').value,
        use_ipv4: document.getElementById('proj-use-ipv4') ? document.getElementById('proj-use-ipv4').value === 'true' : true
    };
    
    // 验证域名格式
    if (project.domains) {
        const domains = project.domains.split('\n').map(d => d.trim()).filter(d => d);
        const invalidDomains = [];
        
        for (const domain of domains) {
            if (!isValidDomain(domain)) {
                invalidDomains.push(domain);
            }
        }
        
        if (invalidDomains.length > 0) {
            alert('以下域名格式不正确:\n\n' + invalidDomains.join('\n') + 
                  '\n\n请检查:\n1. 域名格式是否正确（如 example.com）\n2. 是否包含非法字符\n3. 是否有拼写错误');
            return;
        }
    }
    
    // 验证 SSL 配置
    if (project.ssl_enabled && !project.domains) {
        alert('启用 SSL 需要绑定域名');
        return;
    }
    
    if (project.ssl_enabled && !project.ssl_email) {
        if (!confirm('未设置证书邮箱，是否继续？\n\n建议填写邮箱以接收证书相关通知。')) {
            return;
        }
    }
    
    // 如果是编辑模式，添加 ID
    if (currentProjectId) {
        project.id = currentProjectId;
    }
    
    const url = currentProjectId ? '/api/projects/update' : '/api/projects/add';
    
    // 显示加载状态
    const submitBtn = document.getElementById('submit-project-btn');
    const originalText = submitBtn.textContent;
    submitBtn.disabled = true;
    submitBtn.textContent = '处理中...';
    
    try {
        const res = await fetch(url, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(project)
        });
        
        const data = await res.json();
        
        if (res.ok && data.success) {
            closeModal('add-project-modal');
            loadProjects();
            
            // 显示详细反馈
            let message = currentProjectId ? '✓ 项目更新成功！' : '✓ 项目创建成功！';
            
            if (data.ssl_warnings && data.ssl_warnings.length > 0) {
                message += '\n\n⚠ SSL 警告:\n' + data.ssl_warnings.join('\n');
            }
            
            if (data.start_message) {
                message += '\n\n' + data.start_message;
            }
            
            alert(message);
            clearProjectForm();
            currentProjectId = null;
            
            // 检查SSL配置状态
            if (project.ssl_enabled && project.domains) {
                setTimeout(() => checkSSLStatus(), 3000);
            }
        } else {
            // 显示详细错误
            showDetailedError(data);
        }
    } catch (err) {
        alert('网络错误: ' + err.message + '\n\n请检查网络连接后重试');
    } finally {
        submitBtn.disabled = false;
        submitBtn.textContent = originalText;
    }
}

function showDetailedError(data) {
    let message = '❌ 操作失败\n\n';
    
    if (data.error) {
        message += '错误: ' + data.error + '\n\n';
    }
    
    if (data.details && data.details.length > 0) {
        message += '详细信息:\n';
        data.details.forEach(detail => {
            message += '  • ' + detail + '\n';
        });
        message += '\n';
    }
    
    if (data.suggestions && data.suggestions.length > 0) {
        message += '建议解决方案:\n';
        data.suggestions.forEach((suggestion, index) => {
            message += '  ' + (index + 1) + '. ' + suggestion + '\n';
        });
        message += '\n';
    }
    
    if (data.ssl_warnings && data.ssl_warnings.length > 0) {
        message += 'SSL 警告:\n';
        data.ssl_warnings.forEach(warning => {
            message += '  • ' + warning + '\n';
        });
    }
    
    alert(message);
}

// 验证域名格式
function isValidDomain(domain) {
    // 移除前后空格
    domain = domain.trim();
    
    // 检查长度
    if (domain.length === 0 || domain.length > 253) {
        return false;
    }
    
    // 检查是否包含空格
    if (domain.includes(' ') || domain.includes('\t')) {
        return false;
    }
    
    // localhost 是有效的
    if (domain === 'localhost') {
        return true;
    }
    
    // IP 地址模式
    const ipPattern = /^(\d{1,3}\.){3}\d{1,3}$/;
    if (ipPattern.test(domain)) {
        return true;
    }
    
    // 域名模式：允许字母、数字、连字符、点和下划线
    const domainPattern = /^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)*[a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?$/;
    
    return domainPattern.test(domain);
}

async function checkSSLStatus() {
    try {
        const res = await fetch('/api/caddy/ssl-status');
        const data = await res.json();
        
        if (data.errors && data.errors.length > 0) {
            let errorMsg = 'SSL证书申请可能存在问题:\n\n';
            data.errors.forEach(err => {
                errorMsg += '• ' + err + '\n';
            });
            errorMsg += '\n请检查:\n1. 域名DNS是否正确解析到本服务器\n2. 80和443端口是否开放\n3. 域名是否可以公网访问';
            alert(errorMsg);
        }
    } catch (err) {
        console.error('SSL状态检查失败:', err);
    }
}

function clearProjectForm() {
    document.getElementById('proj-type').value = '';
    document.getElementById('proj-name').value = '';
    document.getElementById('proj-root').value = '';
    document.getElementById('proj-exec').value = '';
    document.getElementById('proj-cmd').value = '';
    document.getElementById('proj-port').value = '';
    document.getElementById('proj-autostart').value = 'false';
    document.getElementById('proj-domains').value = '';
    document.getElementById('proj-ssl').value = 'true';
    document.getElementById('proj-email').value = '';
    document.getElementById('proj-proxy-path').value = '/';
    document.getElementById('proj-headers').value = '';
    document.getElementById('proj-desc').value = '';
    currentStep = 1;
    updateWizardSteps();
}

async function startProject(id) {
    const btn = event.target;
    btn.disabled = true;
    btn.textContent = '启动中...';
    
    try {
        const res = await fetch('/api/projects/start?id=' + id, { method: 'POST' });
        const data = await res.json();
        
        if (data.success) {
            alert('✓ ' + data.message + '\n\n端口: ' + data.port + '\n\n项目已在后台运行');
            setTimeout(loadProjects, 500);
        } else {
            showErrorDialog(data);
            setTimeout(loadProjects, 500);
        }
    } catch (err) {
        alert('启动失败: ' + err.message + '\n\n请检查网络连接');
        setTimeout(loadProjects, 500);
    }
}

function showErrorDialog(data) {
    let message = '❌ ' + (data.error || '启动失败') + '\n\n';
    
    if (data.code) {
        message += '错误代码: ' + data.code + '\n\n';
    }
    
    if (data.details && data.details.length > 0) {
        message += '详细信息:\n';
        data.details.forEach(detail => {
            message += '  • ' + detail + '\n';
        });
        message += '\n';
    }
    
    if (data.suggestions && data.suggestions.length > 0) {
        message += '💡 建议解决方案:\n';
        data.suggestions.forEach((suggestion, index) => {
            message += '  ' + (index + 1) + '. ' + suggestion + '\n';
        });
        message += '\n';
    }
    
    if (data.log_path) {
        message += '📄 日志文件: ' + data.log_path + '\n\n';
        message += '提示: 查看日志文件可能包含更多错误信息';
    }
    
    alert(message);
}

async function stopProject(id) {
    if (!confirm('确定停止该项目吗？\n\n停止后项目将无法访问，直到重新启动。')) return;
    
    const btn = event.target;
    btn.disabled = true;
    btn.textContent = '停止中...';
    
    try {
        const res = await fetch('/api/projects/stop?id=' + id, { method: 'POST' });
        if (res.ok) {
            alert('✓ 项目已停止');
        } else {
            alert('❌ 停止失败，请稍后重试');
        }
    } catch (err) {
        alert('操作失败: ' + err.message);
    } finally {
        setTimeout(loadProjects, 500);
    }
}

async function restartProject(id) {
    if (!confirm('确定重启该项目吗？\n\n项目将短暂中断服务。')) return;
    
    const btn = event.target;
    btn.disabled = true;
    btn.textContent = '重启中...';
    
    try {
        const res = await fetch('/api/projects/restart?id=' + id, { method: 'POST' });
        const data = await res.json();
        
        if (data && data.success) {
            alert('✓ 项目已重启');
        } else if (data && data.error) {
            showErrorDialog(data);
        } else {
            alert('❌ 重启可能失败，请查看项目状态');
        }
    } catch (err) {
        alert('操作失败: ' + err.message);
    } finally {
        setTimeout(loadProjects, 500);
    }
}

async function deleteProject(id) {
    if (!confirm('确定删除该项目吗？此操作不可恢复！')) return;
    await fetch('/api/projects/delete?id=' + id, { method: 'POST' });
    loadProjects();
}

// override: cancel restart feature to avoid hang
async function restartProject(id) {
    alert('已取消重启功能，请使用“停止/启动”切换');
}

// ===== 任务管理 =====

async function loadTasks() {
    const res = await fetch('/api/tasks');
    const tasks = await res.json();
    document.getElementById('task-list').innerHTML = (!tasks || tasks.length === 0) ? 
        '<p style="text-align:center;color:#909399;padding:40px;">暂无任务</p>' : 
        tasks.map(t => '<div class="file-item"><div><strong>' + t.name + '</strong><p style="color:#606266;margin:3px 0;font-size:13px;">命令: ' + t.command + '</p></div><div><button class="btn btn-danger btn-sm" onclick="deleteTask(' + t.id + ')">删除</button></div></div>').join('');
}

function showAddTask() {
    document.getElementById('add-task-modal').style.display = 'block';
}

async function submitTask() {
    const task = {
        name: document.getElementById('task-name').value,
        command: document.getElementById('task-cmd').value,
        schedule: document.getElementById('task-schedule').value,
        is_loop: document.getElementById('task-loop').value === 'true'
    };
    
    if (!task.name || !task.command || !task.schedule) {
        alert('请填写所有必填项');
        return;
    }
    
    const res = await fetch('/api/tasks/add', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(task)
    });
    
    if (res.ok) {
        closeModal('add-task-modal');
        loadTasks();
        document.getElementById('task-name').value = '';
        document.getElementById('task-cmd').value = '';
        document.getElementById('task-schedule').value = '';
    }
}

async function deleteTask(id) {
    if (!confirm('确定删除该任务吗？')) return;
    await fetch('/api/tasks/delete?id=' + id, { method: 'POST' });
    loadTasks();
}

// ===== 切换标签 =====

function switchTab(tab) {
    document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
    document.querySelectorAll('.tab-content').forEach(c => c.classList.remove('active'));
    event.target.classList.add('active');
    document.getElementById(tab + '-tab').classList.add('active');
    
    if (tab === 'dashboard') { 
        loadSystemInfo(); 
        startMonitorRefresh(); // 启动监控刷新
    }
    if (tab === 'projects') { 
        loadProjects(); 
        // 定期刷新项目状态
        if (window.projectStatusInterval) {
            clearInterval(window.projectStatusInterval);
        }
        window.projectStatusInterval = setInterval(loadProjects, 5000);
    } else {
        // 离开项目页面时停止刷新
        if (window.projectStatusInterval) {
            clearInterval(window.projectStatusInterval);
        }
    }
    if (tab === 'tasks') { loadTasks(); }
    if (tab === 'files') { loadFiles(''); }
    if (tab === 'logs') { refreshLogs(); }
    if (tab === 'env') { loadEnvs(); }
    if (tab === 'settings') { loadSettings(); }
}

// ===== 文件管理 =====

async function loadFiles(path) {
    currentPath = path;
    const res = await fetch('/api/files/browse?path=' + encodeURIComponent(path));
    const data = await res.json();
    
    document.getElementById('file-breadcrumb').textContent = '当前目录: ' + data.current_path;
    
    const browser = document.getElementById('file-browser');
    let html = '';
    
    if (data.parent_path && data.parent_path !== data.current_path) {
        html += '<div class="file-item" onclick="loadFiles(\'' + data.parent_path.replace(/\\/g, '\\\\') + '\')"><div>📁 ..</div></div>';
    }
    
    html += data.files.map(f => 
        '<div class="file-item">' +
        '<div onclick="' + (f.is_dir ? 'loadFiles(\'' + f.path.replace(/\\/g, '\\\\') + '\')' : '') + '">' +
        (f.is_dir ? '📁' : '📄') + ' ' + f.name + 
        (!f.is_dir ? ' <small>(' + formatSize(f.size) + ')</small>' : '') +
        '</div>' +
        '<div>' +
        (!f.is_dir ? '<button class="btn btn-primary btn-sm" onclick="downloadFile(\'' + f.path.replace(/\\/g, '\\\\') + '\')">下载</button>' : '') +
        '<button class="btn btn-danger btn-sm" onclick="deleteFile(\'' + f.path.replace(/\\/g, '\\\\') + '\')">删除</button>' +
        '</div></div>'
    ).join('');
    
    browser.innerHTML = html;
}

function formatSize(bytes) {
    if (bytes < 1024) return bytes + ' B';
    if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(2) + ' KB';
    return (bytes / (1024 * 1024)).toFixed(2) + ' MB';
}

function uploadFileToPath() {
    const files = document.getElementById('file-upload').files;
    if (!files || files.length === 0) {
        alert('请选择文件');
        return;
    }

    uploadFiles(files);
}

async function uploadFiles(files) {
    const progressDiv = document.getElementById('upload-progress');
    const statusSpan = document.getElementById('upload-status');
    const percentSpan = document.getElementById('upload-percent');
    const barDiv = document.getElementById('upload-bar');
    
    progressDiv.style.display = 'block';
    
    let totalFiles = files.length;
    let uploadedFiles = 0;
    let failedFiles = 0;
    
    for (let i = 0; i < files.length; i++) {
        const file = files[i];
        
        statusSpan.textContent = `上传中 (${uploadedFiles + 1}/${totalFiles}): ${file.name}`;
        
        const formData = new FormData();
        formData.append('file', file);
        formData.append('path', currentPath);
        
        try {
            const response = await fetch('/api/files/upload', {
                method: 'POST',
                body: formData
            });
            
            if (response.ok) {
                uploadedFiles++;
            } else {
                failedFiles++;
                console.error(`上传失败: ${file.name}`);
            }
        } catch (err) {
            failedFiles++;
            console.error(`上传错误: ${file.name}`, err);
        }
        
        const percent = Math.round(((uploadedFiles + failedFiles) / totalFiles) * 100);
        percentSpan.textContent = percent + '%';
        barDiv.style.width = percent + '%';
    }
    
    if (failedFiles > 0) {
        statusSpan.textContent = `上传完成: ${uploadedFiles} 成功, ${failedFiles} 失败`;
        barDiv.style.background = '#E6A23C';
    } else {
        statusSpan.textContent = `✓ 全部上传成功 (${uploadedFiles} 个文件)`;
        barDiv.style.background = '#67C23A';
    }
    
    loadFiles(currentPath);
    document.getElementById('file-upload').value = '';
    document.getElementById('selected-files').textContent = '';
    
    setTimeout(() => {
        progressDiv.style.display = 'none';
        barDiv.style.width = '0%';
        barDiv.style.background = '#409EFF';
    }, 3000);
}

function createNewFolder() {
    const name = prompt('请输入文件夹名称:');
    if (!name) return;
    
    // 验证文件夹名称
    if (!/^[a-zA-Z0-9_\-\u4e00-\u9fa5\s]+$/.test(name)) {
        alert('文件夹名称包含非法字符！\n只允许字母、数字、中文、下划线和连字符');
        return;
    }
    
    fetch('/api/files/create-folder', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ path: currentPath, name })
    })
    .then(async res => {
        if (res.ok) {
            const data = await res.json();
            alert('✓ ' + data.message);
            loadFiles(currentPath);
        } else {
            const error = await res.text();
            alert('✗ 创建失败: ' + error);
        }
    })
    .catch(err => {
        alert('✗ 创建失败: ' + err.message);
    });
}

function deleteFile(path) {
    if (!confirm('确定删除吗？')) return;
    fetch('/api/files/delete?path=' + encodeURIComponent(path), { method: 'POST' })
        .then(() => loadFiles(currentPath));
}

function downloadFile(path) {
    window.open('/api/files/download?path=' + encodeURIComponent(path), '_blank');
}

// ===== 日志 =====

async function refreshLogs() {
    const res = await fetch('/api/caddy/logs');
    const data = await res.json();
    document.getElementById('log-content').textContent = data.logs || '暂无日志';
    document.getElementById('log-content').scrollTop = document.getElementById('log-content').scrollHeight;
}

// ===== 环境 =====

async function loadEnvs() {
    const list = document.getElementById('env-list');
    list.innerHTML = currentEnvs.map(env => 
        '<div class="file-item"><div><strong>' + env.name + '</strong> ' +
        '<span class="env-status ' + (env.installed ? 'env-installed' : 'env-not-installed') + '">' +
        (env.installed ? '已安装 ' + env.version : '未安装') + '</span></div>' +
        '<div>' + (!env.installed ? '<button class="btn btn-warning btn-sm" onclick="showEnvGuide(\'' + env.name.toLowerCase() + '\')">安装指南</button>' : '') +
        '</div></div>'
    ).join('');
}

async function showEnvGuide(env) {
    const res = await fetch('/api/env/guide', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ env })
    });
    
    const guide = await res.json();
    if (confirm(guide.title + '\n\n' + guide.steps + '\n\n点击"确定"打开下载页面')) {
        window.open(guide.download, '_blank');
    }
}

// ===== 设置 =====

async function loadSettings() {
    const res = await fetch('/api/settings/get');
    const data = await res.json();
    document.getElementById('security-path').value = data.security_path || '';
    document.getElementById('www-root').value = data.www_root || 'C:\\www';
}

async function saveSettings() {
    const securityPath = document.getElementById('security-path').value;
    const wwwRoot = document.getElementById('www-root').value;
    
    const res = await fetch('/api/settings/update', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ security_path: securityPath, www_root: wwwRoot })
    });
    
    if (res.ok) {
        alert('设置已保存！');
    }
}

async function changePassword() {
    const username = document.getElementById('change-username').value;
    const oldPassword = document.getElementById('old-password').value;
    const newPassword = document.getElementById('new-password').value;
    const newPassword2 = document.getElementById('new-password2').value;
    
    if (!username || !oldPassword || !newPassword) {
        alert('请填写完整信息');
        return;
    }
    
    if (newPassword !== newPassword2) {
        alert('两次新密码不一致');
        return;
    }
    
    const res = await fetch('/api/user/password', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, old_password: oldPassword, new_password: newPassword })
    });
    
    if (res.ok) {
        alert('密码修改成功，请重新登录');
        logout();
    } else {
        alert('修改失败: ' + await res.text());
    }
}

async function shutdownApplication() {
    if (!confirm('确定要关闭 Caddy 管理器吗？\n\n这将停止所有正在运行的服务和项目。')) {
        return;
    }
    
    try {
        const res = await fetch('/api/app/shutdown', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' }
        });
        
        if (res.ok) {
            alert('应用程序正在关闭...\n\n所有服务将被安全停止。');
            // 不需要做其他操作，服务器会自动关闭
        }
    } catch (err) {
        // 请求可能因为服务器关闭而失败，这是正常的
        console.log('应用程序已关闭');
    }
}

// ===== 其他 =====

// checkProjectSSL - 快速检查项目的 SSL 状态
async function checkProjectSSL(domain) {
    if (!domain) {
        alert('未配置域名');
        return;
    }
    
    // 使用模态框显示结果
    const modal = document.getElementById('diagnostics-modal') || createDiagnosticsModal();
    const resultDiv = document.getElementById('diagnostics-result');
    resultDiv.innerHTML = '<p style="color:#909399;">正在检查 SSL 配置...</p>';
    modal.style.display = 'block';
    
    try {
        const res = await fetch('/api/diagnostics/ssl?domain=' + encodeURIComponent(domain));
        const data = await res.json();
        
        let html = '<div style="background:#f5f7fa;padding:15px;border-radius:4px;">';
        html += '<h4>SSL 诊断结果 - ' + data.domain + '</h4>';
        
        if (data.issues && data.issues.length > 0) {
            for (const issue of data.issues) {
                const color = issue.severity === 'error' ? '#F56C6C' :
                             issue.severity === 'warning' ? '#E6A23C' : '#67C23A';
                
                html += '<div style="background:white;padding:15px;margin:10px 0;border-left:4px solid ' + color + ';border-radius:4px;">';
                html += '<h5 style="margin:0 0 10px 0;color:' + color + ';">';
                
                if (issue.severity === 'info') {
                    html += '✅ ' + issue.title;
                } else if (issue.severity === 'warning') {
                    html += '⚠️ ' + issue.title;
                } else {
                    html += '❌ ' + issue.title;
                }
                
                html += '</h5>';
                html += '<p style="color:#606266;white-space:pre-wrap;margin:5px 0;">' + issue.description + '</p>';
                
                if (issue.solutions && issue.solutions.length > 0) {
                    html += '<div style="margin-top:10px;"><strong style="color:#909399;">解决方案:</strong><ul style="margin:5px 0;padding-left:20px;">';
                    issue.solutions.forEach(sol => {
                        html += '<li style="color:#606266;margin:3px 0;">' + sol + '</li>';
                    });
                    html += '</ul></div>';
                }
                
                html += '</div>';
            }
        } else {
            html += '<p style="color:#67C23A;margin:10px 0;">✓ SSL 配置正常</p>';
        }
        
        html += '</div>';
        resultDiv.innerHTML = html;
    } catch (err) {
        resultDiv.innerHTML = '<p style="color:#F56C6C;">检查失败: ' + err.message + '</p>';
    }
}

function createDiagnosticsModal() {
    // 如果诊断模态框不存在，创建它
    const modal = document.createElement('div');
    modal.id = 'diagnostics-modal';
    modal.className = 'modal';
    modal.innerHTML = `
        <div class="modal-content" style="max-width:600px;">
            <div class="modal-header">
                <h2>SSL 诊断</h2>
                <span class="close" onclick="document.getElementById('diagnostics-modal').style.display='none'">&times;</span>
            </div>
            <div id="diagnostics-result" class="modal-body"></div>
        </div>
    `;
    document.body.appendChild(modal);
    return modal;
}

async function checkCaddyStatus() {
    try {
        const res = await fetch('/api/caddy/status');
        if (!res.ok) {
            console.error('Caddy status check failed:', res.status);
            return;
        }
        const data = await res.json();
        const statusEl = document.getElementById('caddy-status');
        const controlsEl = document.getElementById('caddy-controls');
        
        if (!statusEl) return; // Element not yet loaded
        
        if (data.running) {
            const versionText = data.version ? ` (${data.version})` : '';
            statusEl.textContent = 'Caddy 运行中' + versionText;
            statusEl.style.color = '#67C23A';
            if (controlsEl) {
                controlsEl.innerHTML = '<button class="btn btn-warning btn-sm" onclick="stopCaddy()">停止</button>' +
                                       '<button class="btn btn-primary btn-sm" onclick="reloadCaddy()">重载配置</button>' +
                                       '<button class="btn btn-primary btn-sm" onclick="restartCaddy()">重启</button>';
            }
        } else {
            statusEl.textContent = 'Caddy 未运行';
            statusEl.style.color = '#F56C6C';
            if (controlsEl) {
                controlsEl.innerHTML = '<button class="btn btn-success btn-sm" onclick="startCaddy()">启动</button>';
            }
        }
    } catch (err) {
        console.error('Caddy status check error:', err);
        // Silently fail to avoid disturbing user experience
    }
}

async function startCaddy() {
    const btn = event.target;
    btn.disabled = true;
    btn.textContent = '启动中...';
    
    try {
        const res = await fetch('/api/caddy/start', { method: 'POST' });
        const data = await res.json();
        
        if (data.success) {
            alert('✓ Caddy 启动成功');
        } else {
            alert('❌ 启动失败: ' + (data.error || '未知错误'));
        }
    } catch (err) {
        alert('❌ 启动失败: ' + err.message);
    } finally {
        btn.disabled = false;
        btn.textContent = '启动';
        setTimeout(checkCaddyStatus, 500);
    }
}

async function stopCaddy() {
    if (!confirm('确定要停止 Caddy 吗？\n\n这将导致所有网站暂时无法访问。')) {
        return;
    }
    
    const btn = event.target;
    btn.disabled = true;
    btn.textContent = '停止中...';
    
    try {
        const res = await fetch('/api/caddy/stop', { method: 'POST' });
        const data = await res.json();
        
        if (data.success) {
            alert('✓ Caddy 已停止');
        } else {
            alert('❌ 停止失败: ' + (data.error || '未知错误'));
        }
    } catch (err) {
        alert('❌ 停止失败: ' + err.message);
    } finally {
        btn.disabled = false;
        btn.textContent = '停止';
        setTimeout(checkCaddyStatus, 500);
    }
}

async function restartCaddy() {
    if (!confirm('确定要重启 Caddy 吗？\n\n网站将短暂中断服务（约1-2秒）。\n\n💡 提示：如果只是修改了配置，建议使用"重载配置"功能，可实现零停机更新。')) {
        return;
    }
    
    const btn = event.target;
    btn.disabled = true;
    btn.textContent = '重启中...';
    
    try {
        const res = await fetch('/api/caddy/restart', { method: 'POST' });
        const data = await res.json();
        
        if (data.success) {
            alert('✓ Caddy 重启成功');
        } else {
            alert('❌ 重启失败: ' + (data.error || '未知错误'));
        }
    } catch (err) {
        alert('❌ 重启失败: ' + err.message);
    } finally {
        btn.disabled = false;
        btn.textContent = '重启';
        setTimeout(checkCaddyStatus, 500);
    }
}

async function reloadCaddy() {
    if (!confirm('确定要重新加载 Caddy 配置吗？\n\n✅ 此操作不会中断服务（零停机更新）\n✅ 适合在修改配置后使用')) {
        return;
    }
    
    const btn = event.target;
    btn.disabled = true;
    btn.textContent = '重载中...';
    
    try {
        const res = await fetch('/api/caddy/reload', { method: 'POST' });
        const data = await res.json();
        
        if (data.success) {
            alert('✓ ' + data.message);
        } else {
            alert('❌ 重载失败: ' + (data.error || '未知错误') + '\n\n如果配置文件有语法错误，请检查后重试。');
        }
    } catch (err) {
        alert('❌ 重载失败: ' + err.message);
    } finally {
        btn.disabled = false;
        btn.textContent = '重载配置';
        setTimeout(checkCaddyStatus, 500);
    }
}

function logout() {
    fetch('/api/logout', { method: 'POST' })
        .then(() => location.reload());
}

function closeModal(id) {
    document.getElementById(id).style.display = 'none';
}

window.onclick = function(event) {
    if (event.target.classList.contains('modal')) {
        event.target.style.display = 'none';
    }
}

// ===== 诊断功能 =====

async function runDiagnostics() {
    const resultDiv = document.getElementById('diagnostics-result');
    resultDiv.innerHTML = '<p style="color:#909399;">正在检查...</p>';
    
    try {
        const res = await fetch('/api/diagnostics/run');
        if (!res.ok) {
            if (res.status === 401) {
                alert('登录已过期，请重新登录');
                logout();
                return;
            }
            throw new Error(`HTTP ${res.status}: ${res.statusText}`);
        }
        const data = await res.json();
        
        let html = '<div style="background:#f5f7fa;padding:15px;border-radius:4px;margin-top:10px;">';
        html += '<h4>诊断结果 (' + new Date(data.timestamp).toLocaleString() + ')</h4>';
        
        if (data.issues && data.issues.length > 0) {
            for (const issue of data.issues) {
                const color = issue.severity === 'error' ? '#F56C6C' : 
                             issue.severity === 'warning' ? '#E6A23C' : '#409EFF';
                
                html += '<div style="background:white;padding:15px;margin:10px 0;border-left:4px solid ' + color + ';border-radius:4px;">';
                html += '<h5 style="margin:0 0 10px 0;color:' + color + ';">' + issue.title + '</h5>';
                html += '<p style="margin:5px 0;color:#606266;">' + issue.description + '</p>';
                
                if (issue.solutions && issue.solutions.length > 0) {
                    html += '<p style="margin:10px 0 5px 0;font-weight:600;">解决方案:</p>';
                    html += '<ul style="margin:0;padding-left:20px;">';
                    for (const solution of issue.solutions) {
                        html += '<li style="margin:5px 0;">' + solution + '</li>';
                    }
                    html += '</ul>';
                }
                
                if (issue.auto_fix) {
                    html += '<button class="btn btn-warning btn-sm" style="margin-top:10px;" onclick="autoFix(\'' + issue.code + '\')">自动修复</button>';
                }
                
                html += '</div>';
            }
        } else {
            html += '<p style="color:#67C23A;margin:10px 0;">✓ 未发现问题，系统运行正常</p>';
        }
        
        html += '</div>';
        resultDiv.innerHTML = html;
    } catch (err) {
        resultDiv.innerHTML = '<p style="color:#F56C6C;">诊断失败: ' + err.message + '</p>';
    }
}

async function checkSSLIssues() {
    const domain = prompt('请输入要检查的域名:');
    if (!domain) return;
    
    const resultDiv = document.getElementById('diagnostics-result');
    resultDiv.innerHTML = '<p style="color:#909399;">正在检查 SSL 配置...</p>';
    
    try {
        const res = await fetch('/api/diagnostics/ssl?domain=' + encodeURIComponent(domain));
        if (!res.ok) {
            if (res.status === 401) {
                alert('登录已过期，请重新登录');
                logout();
                return;
            }
            throw new Error(`HTTP ${res.status}: ${res.statusText}`);
        }
        const data = await res.json();
        
        let html = '<div style="background:#f5f7fa;padding:15px;border-radius:4px;margin-top:10px;">';
        html += '<h4>SSL 诊断结果 - ' + data.domain + '</h4>';
        
        if (data.issues && data.issues.length > 0) {
            for (const issue of data.issues) {
                const color = issue.severity === 'error' ? '#F56C6C' : 
                             issue.severity === 'warning' ? '#E6A23C' : '#409EFF';
                
                html += '<div style="background:white;padding:15px;margin:10px 0;border-left:4px solid ' + color + ';border-radius:4px;">';
                html += '<h5 style="margin:0 0 10px 0;color:' + color + ';">' + issue.title + '</h5>';
                html += '<p style="margin:5px 0;color:#606266;white-space:pre-line;">' + issue.description + '</p>';
                
                if (issue.solutions && issue.solutions.length > 0) {
                    html += '<p style="margin:10px 0 5px 0;font-weight:600;">建议:</p>';
                    html += '<ul style="margin:0;padding-left:20px;">';
                    for (const solution of issue.solutions) {
                        html += '<li style="margin:5px 0;">' + solution + '</li>';
                    }
                    html += '</ul>';
                }
                
                html += '</div>';
            }
        } else {
            html += '<p style="color:#67C23A;margin:10px 0;">✓ SSL 配置正常</p>';
        }
        
        html += '</div>';
        resultDiv.innerHTML = html;
    } catch (err) {
        resultDiv.innerHTML = '<p style="color:#F56C6C;">检查失败: ' + err.message + '</p>';
    }
}

async function autoFix(issueCode) {
    if (!confirm('确定要自动修复此问题吗？\n\n某些操作可能需要管理员权限。')) {
        return;
    }
    
    try {
        const res = await fetch('/api/diagnostics/autofix', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ issue_code: issueCode })
        });
        
        const data = await res.json();
        
        if (data.success) {
            alert('✓ 修复成功！\n\n' + (data.message || ''));
            runDiagnostics();
        } else {
            alert('✗ 修复失败\n\n' + (data.error || ''));
        }
    } catch (err) {
        alert('✗ 修复失败\n\n' + err.message);
    }
}

// ===== 文件上传增强 =====

// 页面加载完成后初始化文件上传功能
window.addEventListener('load', function() {
    const fileInput = document.getElementById('file-upload');
    if (fileInput) {
        fileInput.addEventListener('change', function() {
            const files = this.files;
            const selectedFilesSpan = document.getElementById('selected-files');
            if (files.length > 0) {
                const names = Array.from(files).map(f => f.name).join(', ');
                const text = files.length === 1 ? 
                    `已选择: ${names}` : 
                    `已选择 ${files.length} 个文件`;
                if (selectedFilesSpan) {
                    selectedFilesSpan.textContent = text;
                }
            }
        });
    }
    
    const dropZone = document.getElementById('upload-drop-zone');
    if (dropZone) {
        dropZone.addEventListener('click', function() {
            document.getElementById('file-upload').click();
        });
        
        dropZone.addEventListener('dragover', function(e) {
            e.preventDefault();
            e.stopPropagation();
            this.style.borderColor = '#409EFF';
            this.style.background = '#ecf5ff';
        });
        
        dropZone.addEventListener('dragleave', function(e) {
            e.preventDefault();
            e.stopPropagation();
            this.style.borderColor = '#dcdfe6';
            this.style.background = '#fafafa';
        });
        
        dropZone.addEventListener('drop', function(e) {
            e.preventDefault();
            e.stopPropagation();
            this.style.borderColor = '#dcdfe6';
            this.style.background = '#fafafa';
            
            const files = e.dataTransfer.files;
            if (files.length > 0) {
                uploadFiles(files);
            }
        });
    }
});

// ===== 系统监控 =====

let monitorInterval = null;

async function refreshMonitor() {
    try {
        const res = await fetch('/api/system/monitor');
        const data = await res.json();
        
        // 更新 CPU
        document.getElementById('cpu-percent').textContent = data.cpu.used_percent.toFixed(1) + '%';
        document.getElementById('cpu-cores').textContent = data.cpu.cores;
        document.getElementById('cpu-bar').style.width = data.cpu.used_percent + '%';
        
        // 更新内存
        document.getElementById('memory-percent').textContent = data.memory.used_percent.toFixed(1) + '%';
        document.getElementById('memory-used').textContent = (data.memory.used_mb / 1024).toFixed(1) + ' GB';
        document.getElementById('memory-total').textContent = (data.memory.total_mb / 1024).toFixed(1) + ' GB';
        document.getElementById('memory-bar').style.width = data.memory.used_percent + '%';
        
        // 根据使用率改变颜色
        const memoryBar = document.getElementById('memory-bar');
        if (data.memory.used_percent > 90) {
            memoryBar.style.background = '#F56C6C';
        } else if (data.memory.used_percent > 70) {
            memoryBar.style.background = '#E6A23C';
        } else {
            memoryBar.style.background = '#67C23A';
        }
        
        // 更新磁盘
        const diskInfo = document.getElementById('disk-info');
        diskInfo.innerHTML = data.disks.map(disk => {
            let color = '#67C23A';
            if (disk.used_percent > 90) color = '#F56C6C';
            else if (disk.used_percent > 70) color = '#E6A23C';
            
            return `
                <div style="background: #f5f7fa; padding: 15px; border-radius: 4px;">
                    <h4 style="color: #606266; margin-bottom: 8px;">${disk.drive}</h4>
                    <div style="font-size: 18px; font-weight: 600; color: ${color};">
                        ${disk.used_percent.toFixed(1)}%
                    </div>
                    <div style="color: #909399; font-size: 13px; margin-top: 5px;">
                        已用: ${disk.used_gb} GB / ${disk.total_gb} GB
                    </div>
                    <div class="progress-bar" style="margin-top: 8px;">
                        <div class="progress-fill" style="width: ${disk.used_percent}%; background: ${color};"></div>
                    </div>
                    <div style="color: #909399; font-size: 12px; margin-top: 5px;">
                        可用: ${disk.free_gb} GB
                    </div>
                </div>
            `;
        }).join('');
        
    } catch (err) {
        console.error('获取监控数据失败:', err);
    }
}

// 启动定时刷新
function startMonitorRefresh() {
    refreshMonitor();
    if (monitorInterval) clearInterval(monitorInterval);
    monitorInterval = setInterval(refreshMonitor, 3000); // 每3秒刷新
}

// 停止定时刷新
function stopMonitorRefresh() {
    if (monitorInterval) {
        clearInterval(monitorInterval);
        monitorInterval = null;
    }
}

// 初始化文件路径
async function initializeFilePath() {
    try {
        const res = await fetch('/api/settings/get');
        if (!res.ok) {
            console.warn('获取设置失败，使用默认路径');
            currentPath = 'C:\\www';
            return;
        }
        const data = await res.json();
        currentPath = data.www_root || 'C:\\www';
        console.log('初始化文件路径:', currentPath);
    } catch (err) {
        currentPath = 'C:\\www'; // 默认路径
        console.error('获取默认路径失败，使用 C:\\www:', err);
    }
}
