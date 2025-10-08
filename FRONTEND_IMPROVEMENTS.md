# 项目创建和启动 - 前端改进代码

将以下代码添加到 web/static/app.js

```javascript
// ===== 改进的项目提交函数 =====

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

    // 前端验证
    const validationErrors = validateProjectFrontend(project);
    if (validationErrors.length > 0) {
        showValidationErrors(validationErrors);
        return;
    }

    // 域名格式验证
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

    // SSL 验证
    if (project.ssl_enabled && !project.domains) {
        alert('❌ 启用 SSL 需要绑定域名\n\n请在"域名配置"步骤中添加域名');
        return;
    }

    if (project.ssl_enabled && !project.ssl_email) {
        if (!confirm('⚠ 未设置证书邮箱\n\n建议填写邮箱以接收证书相关通知。\n\n是否继续？')) {
            return;
        }
    }

    // 如果是编辑模式
    if (currentProjectId) {
        project.id = currentProjectId;
    }

    // 显示加载提示
    const submitBtn = document.getElementById('submit-project-btn');
    const originalText = submitBtn.textContent;
    submitBtn.disabled = true;
    submitBtn.textContent = '处理中...';

    try {
        const url = currentProjectId ? '/api/projects/update' : '/api/projects/add';
        const res = await fetch(url, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(project)
        });

        const data = await res.json();

        if (data.success) {
            showSuccessMessage(data);
            closeModal('add-project-modal');
            loadProjects();
            clearProjectForm();
            currentProjectId = null;
        } else {
            showErrorMessage(data);
        }
    } catch (err) {
        alert('❌ 操作失败\n\n' + err.message);
    } finally {
        submitBtn.disabled = false;
        submitBtn.textContent = originalText;
    }
}

function validateProjectFrontend(project) {
    const errors = [];

    if (!project.name || project.name.trim() === '') {
        errors.push('❌ 项目名称不能为空');
    }

    if (!project.project_type) {
        errors.push('❌ 请选择项目类型');
    }

    if (!project.root_dir || project.root_dir.trim() === '') {
        errors.push('❌ 项目根目录不能为空');
    }

    if (!project.port || project.port <= 0 || project.port > 65535) {
        errors.push('❌ 端口号无效（应在 1-65535 之间）');
    }

    if (!project.exec_path && !project.start_command) {
        errors.push('❌ 必须配置启动命令或可执行文件路径');
    }

    return errors;
}

function showValidationErrors(errors) {
    let message = '❌ 项目配置有误\n\n';
    errors.forEach((error, index) => {
        message += `${index + 1}. ${error}\n`;
    });
    message += '\n请检查并修正后重试';
    alert(message);
}

function showSuccessMessage(data) {
    let message = data.message + '\n';

    if (data.start_message) {
        message += '\n' + data.start_message + '\n';
    }

    if (data.ssl_warnings && data.ssl_warnings.length > 0) {
        message += '\n⚠ SSL 配置提示:\n';
        data.ssl_warnings.forEach((warning, index) => {
            message += `  ${index + 1}. ${warning}\n`;
        });
    }

    if (data.warning) {
        message += '\n⚠ 警告: ' + data.warning + '\n';
        if (data.suggestions) {
            message += '\n建议:\n';
            data.suggestions.forEach((s, i) => {
                message += `  ${i + 1}. ${s}\n`;
            });
        }
    }

    alert(message);
}

function showErrorMessage(data) {
    let message = data.message || data.error || '操作失败';
    message = '❌ ' + message + '\n\n';

    if (data.details) {
        if (Array.isArray(data.details)) {
            message += '详细信息:\n';
            data.details.forEach((detail, index) => {
                message += `  ${index + 1}. ${detail}\n`;
            });
        } else {
            message += '详细信息:\n  ' + data.details + '\n';
        }
        message += '\n';
    }

    if (data.suggestions && data.suggestions.length > 0) {
        message += '解决方案:\n';
        data.suggestions.forEach((suggestion, index) => {
            message += `  ${index + 1}. ${suggestion}\n`;
        });
        message += '\n';
    }

    if (data.log_path) {
        message += '日志文件: ' + data.log_path;
    }

    alert(message);
}
```

## 使用说明

1. 将上述代码中的函数替换 app.js 中对应的函数
2. 或者直接添加新函数并更新调用

## 功能改进

1. **前端验证** - 提交前检查必填项
2. **详细错误** - 显示具体的错误原因
3. **解决方案** - 每个错误都有解决建议
4. **SSL 警告** - 检测 Cloudflare 并提示
5. **加载状态** - 提交时禁用按钮显示"处理中"
6. **完整反馈** - 成功/警告/错误都有详细信息
