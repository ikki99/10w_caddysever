// Caddy Manager - Complete Version
let currentPath = '';
let currentEnvs = [];
let currentProjectId = null;
let currentStep = 1;

window.onload = function() {
    checkFirstRun();
};

async function checkFirstRun() {
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
    
    list.innerHTML = projects.map(p => 
        '<div class="project-card ' + p.status + '">' +
        '<div class="project-info">' +
        '<div class="project-details">' +
        '<h3>' + p.name + ' <span class="status-badge status-' + p.status + '">' + (p.status === 'running' ? '运行中' : '已停止') + '</span></h3>' +
        '<p style="color:#606266;margin:5px 0;"><strong>类型:</strong> ' + getProjectTypeName(p.project_type) + ' | <strong>端口:</strong> ' + p.port + ' | <strong>域名:</strong> ' + (p.domains || '无') + '</p>' +
        (p.description ? '<p style="color:#909399;font-size:13px;">' + p.description + '</p>' : '') +
        '</div>' +
        '<div class="project-actions">' +
        (p.status === 'running' ? 
            '<button class="btn btn-warning btn-sm" onclick="stopProject(' + p.id + ')">停止</button>' +
            '<button class="btn btn-primary btn-sm" onclick="restartProject(' + p.id + ')">重启</button>' :
            '<button class="btn btn-success btn-sm" onclick="startProject(' + p.id + ')">启动</button>') +
        '<button class="btn btn-danger btn-sm" onclick="deleteProject(' + p.id + ')">删除</button>' +
        '</div></div></div>'
    ).join('');
}

function getProjectTypeName(type) {
    const types = {'go': 'Go', 'python': 'Python', 'nodejs': 'Node.js', 'java': 'Java', 'php': 'PHP', 'static': '静态站点'};
    return types[type] || type;
}

function showAddProject() {
    currentStep = 1;
    document.getElementById('add-project-modal').style.display = 'block';
    updateWizardSteps();
}

function onProjectTypeChange() {
    const type = document.getElementById('proj-type').value;
    document.getElementById('exec-path-group').style.display = (type === 'go' || type === 'java') ? 'block' : 'none';
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
        description: document.getElementById('proj-desc').value
    };
    
    const res = await fetch('/api/projects/add', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(project)
    });
    
    if (res.ok) {
        closeModal('add-project-modal');
        loadProjects();
        alert('项目创建成功！');
        clearProjectForm();
    } else {
        alert('创建失败: ' + await res.text());
    }
}

function clearProjectForm() {
    document.getElementById('proj-type').value = '';
    document.getElementById('proj-name').value = '';
    document.getElementById('proj-root').value = '';
    document.getElementById('proj-exec').value = '';
    document.getElementById('proj-cmd').value = '';
    document.getElementById('proj-port').value = '';
    document.getElementById('proj-domains').value = '';
    document.getElementById('proj-desc').value = '';
}

async function startProject(id) {
    await fetch('/api/projects/start?id=' + id, { method: 'POST' });
    setTimeout(loadProjects, 500);
}

async function stopProject(id) {
    if (!confirm('确定停止该项目吗？')) return;
    await fetch('/api/projects/stop?id=' + id, { method: 'POST' });
    setTimeout(loadProjects, 500);
}

async function restartProject(id) {
    await fetch('/api/projects/restart?id=' + id, { method: 'POST' });
    setTimeout(loadProjects, 500);
}

async function deleteProject(id) {
    if (!confirm('确定删除该项目吗？此操作不可恢复！')) return;
    await fetch('/api/projects/delete?id=' + id, { method: 'POST' });
    loadProjects();
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
    
    if (tab === 'dashboard') { loadSystemInfo(); }
    if (tab === 'projects') { loadProjects(); }
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
    const file = document.getElementById('file-upload').files[0];
    if (!file) {
        alert('请选择文件');
        return;
    }
    
    const formData = new FormData();
    formData.append('file', file);
    formData.append('path', currentPath);
    
    fetch('/api/files/upload', {
        method: 'POST',
        body: formData
    }).then(() => {
        loadFiles(currentPath);
        document.getElementById('file-upload').value = '';
    });
}

function createNewFolder() {
    const name = prompt('请输入文件夹名称:');
    if (!name) return;
    
    fetch('/api/files/create-folder', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ path: currentPath, name })
    }).then(() => loadFiles(currentPath));
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

// ===== 其他 =====

async function checkCaddyStatus() {
    const res = await fetch('/api/caddy/status');
    const data = await res.json();
    const statusEl = document.getElementById('caddy-status');
    if (data.running) {
        statusEl.textContent = 'Caddy 运行中';
        statusEl.style.color = '#67C23A';
    } else {
        statusEl.textContent = 'Caddy 未运行';
        statusEl.style.color = '#F56C6C';
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
