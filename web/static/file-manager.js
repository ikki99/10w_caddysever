// 增强的文件管理功能
// 文件选择状态
let selectedFiles = new Set();
let isMultiSelectMode = false;

// 重写 loadFiles 函数以支持多选
async function loadFilesEnhanced(path) {
    currentPath = path;
    selectedFiles.clear(); // 清除选择
    
    const res = await fetch('/api/files/browse?path=' + encodeURIComponent(path));
    const data = await res.json();
    
    document.getElementById('file-breadcrumb').textContent = '当前目录: ' + data.current_path;
    
    const browser = document.getElementById('file-browser');
    let html = '';
    
    // 工具栏
    html += '<div style="margin-bottom: 15px; padding: 10px; background: #f9f9f9; border-radius: 4px;">';
    html += '<button class="btn btn-primary btn-sm" onclick="toggleMultiSelect()"><span id="multi-select-text">多选模式</span></button> ';
    html += '<button class="btn btn-success btn-sm" id="compress-btn" style="display:none;" onclick="compressSelected()">压缩选中</button> ';
    html += '<button class="btn btn-warning btn-sm" id="delete-multiple-btn" style="display:none;" onclick="deleteSelected()">删除选中</button> ';
    html += '<span style="margin-left: 10px; color: #909399;"><span id="selected-count">0</span> 个文件已选择</span>';
    html += '</div>';
    
    // 返回上一级
    if (data.parent_path && data.parent_path !== data.current_path) {
        html += '<div class="file-item" onclick="loadFiles(\'' + data.parent_path.replace(/\\/g, '\\\\') + '\')">';
        html += '<div>📁 ..</div><div></div></div>';
    }
    
    // 文件列表
    html += data.files.map(f => {
        const escapedPath = f.path.replace(/\\/g, '\\\\').replace(/'/g, "\\'");
        const isZip = f.name.toLowerCase().endsWith('.zip');
        const isEditable = isEditableFile(f.name);
        
        return '<div class="file-item" data-path="' + f.path + '" id="file-' + btoa(f.path) + '">' +
            '<div style="display:flex; align-items:center;">' +
            (isMultiSelectMode ? '<input type="checkbox" class="file-checkbox" value="' + escapedPath + '" onchange="updateSelection()" style="margin-right:10px;">' : '') +
            '<span onclick="' + (f.is_dir ? 'loadFiles(\'' + escapedPath + '\')' : (isEditable ? 'openEditor(\'' + escapedPath + '\')' : '')) + '" style="cursor:pointer;">' +
            (f.is_dir ? '📁' : getFileIcon(f.name)) + ' ' + f.name +
            (!f.is_dir ? ' <small style="color:#909399;">(' + formatSize(f.size) + ')</small>' : '') +
            '</span></div>' +
            '<div>' +
            (isZip ? '<button class="btn btn-success btn-sm" onclick="decompressFile(\'' + escapedPath + '\')">解压</button> ' : '') +
            (isEditable && !f.is_dir ? '<button class="btn btn-primary btn-sm" onclick="openEditor(\'' + escapedPath + '\')">编辑</button> ' : '') +
            (!f.is_dir ? '<button class="btn btn-primary btn-sm" onclick="downloadFile(\'' + escapedPath + '\')">下载</button> ' : '') +
            '<button class="btn btn-warning btn-sm" onclick="renameFile(\'' + escapedPath + '\')">重命名</button> ' +
            '<button class="btn btn-danger btn-sm" onclick="deleteFile(\'' + escapedPath + '\')">删除</button>' +
            '</div></div>';
    }).join('');
    
    browser.innerHTML = html;
}

// 获取文件图标
function getFileIcon(filename) {
    const ext = filename.split('.').pop().toLowerCase();
    const icons = {
        'js': '📜',
        'json': '📋',
        'html': '🌐',
        'htm': '🌐',
        'css': '🎨',
        'php': '🐘',
        'py': '🐍',
        'go': '🔷',
        'java': '☕',
        'txt': '📝',
        'md': '📖',
        'zip': '📦',
        'rar': '📦',
        '7z': '📦',
        'jpg': '🖼️',
        'jpeg': '🖼️',
        'png': '🖼️',
        'gif': '🖼️',
        'svg': '🖼️',
        'pdf': '📕',
        'doc': '📘',
        'docx': '📘',
        'xls': '📗',
        'xlsx': '📗',
    };
    return icons[ext] || '📄';
}

// 判断文件是否可编辑
function isEditableFile(filename) {
    const ext = filename.split('.').pop().toLowerCase();
    const editableExts = ['js', 'json', 'html', 'htm', 'css', 'php', 'py', 'go', 'java', 'txt', 'md', 'xml', 'yml', 'yaml', 'ini', 'conf', 'log', 'sh', 'bat', 'sql', 'env', 'gitignore'];
    return editableExts.includes(ext) || filename.startsWith('.') || !filename.includes('.');
}

// 切换多选模式
function toggleMultiSelect() {
    isMultiSelectMode = !isMultiSelectMode;
    const text = document.getElementById('multi-select-text');
    text.textContent = isMultiSelectMode ? '退出多选' : '多选模式';
    loadFiles(currentPath);
}

// 更新选择状态
function updateSelection() {
    selectedFiles.clear();
    document.querySelectorAll('.file-checkbox:checked').forEach(cb => {
        selectedFiles.add(cb.value);
    });
    
    const count = selectedFiles.size;
    document.getElementById('selected-count').textContent = count;
    document.getElementById('compress-btn').style.display = count > 0 ? 'inline-block' : 'none';
    document.getElementById('delete-multiple-btn').style.display = count > 0 ? 'inline-block' : 'none';
}

// 压缩选中的文件
async function compressSelected() {
    if (selectedFiles.size === 0) {
        alert('请选择要压缩的文件');
        return;
    }
    
    const outputName = prompt('请输入压缩文件名（不含扩展名）:', 'archive');
    if (!outputName) return;
    
    try {
        const res = await fetch('/api/files/compress', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                files: Array.from(selectedFiles),
                output_name: outputName
            })
        });
        
        if (res.ok) {
            const data = await res.json();
            alert('✓ 压缩完成！\n文件: ' + data.file);
            toggleMultiSelect(); // 退出多选模式
            loadFiles(currentPath);
        } else {
            const error = await res.text();
            alert('✗ 压缩失败: ' + error);
        }
    } catch (err) {
        alert('✗ 压缩失败: ' + err.message);
    }
}

// 删除选中的文件
async function deleteSelected() {
    if (selectedFiles.size === 0) {
        alert('请选择要删除的文件');
        return;
    }
    
    if (!confirm(`确定要删除选中的 ${selectedFiles.size} 个文件/文件夹吗？\n\n此操作不可恢复！`)) {
        return;
    }
    
    let deleted = 0;
    let failed = 0;
    
    for (const file of selectedFiles) {
        try {
            const res = await fetch('/api/files/delete?path=' + encodeURIComponent(file), {
                method: 'POST'
            });
            if (res.ok) {
                deleted++;
            } else {
                failed++;
            }
        } catch (err) {
            failed++;
        }
    }
    
    alert(`删除完成\n成功: ${deleted}\n失败: ${failed}`);
    toggleMultiSelect();
    loadFiles(currentPath);
}

// 解压文件
async function decompressFile(filepath) {
    if (!confirm('确定要解压此文件吗？\n\n将解压到同名文件夹')) {
        return;
    }
    
    try {
        const res = await fetch('/api/files/decompress', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                file: filepath
            })
        });
        
        if (res.ok) {
            const data = await res.json();
            alert('✓ 解压完成！\n目录: ' + data.target);
            loadFiles(currentPath);
        } else {
            const error = await res.text();
            alert('✗ 解压失败: ' + error);
        }
    } catch (err) {
        alert('✗ 解压失败: ' + err.message);
    }
}

// 重命名文件
async function renameFile(filepath) {
    const oldName = filepath.split(/[\\\/]/).pop();
    const newName = prompt('请输入新名称:', oldName);
    
    if (!newName || newName === oldName) return;
    
    try {
        const res = await fetch('/api/files/rename', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                old_path: filepath,
                new_name: newName
            })
        });
        
        if (res.ok) {
            alert('✓ 重命名成功');
            loadFiles(currentPath);
        } else {
            const error = await res.text();
            alert('✗ 重命名失败: ' + error);
        }
    } catch (err) {
        alert('✗ 重命名失败: ' + err.message);
    }
}

// 打开代码编辑器
async function openEditor(filepath) {
    try {
        const res = await fetch('/api/files/read?path=' + encodeURIComponent(filepath));
        if (!res.ok) {
            alert('无法打开文件');
            return;
        }
        
        const data = await res.json();
        
        // 显示编辑器模态框
        document.getElementById('editor-modal').style.display = 'block';
        document.getElementById('editor-filename').textContent = data.name;
        document.getElementById('editor-filepath').value = filepath;
        document.getElementById('code-editor').value = data.content;
        
        // 设置语法高亮（如果有）
        detectLanguage(data.name);
        
    } catch (err) {
        alert('打开文件失败: ' + err.message);
    }
}

// 保存文件
async function saveFile() {
    const filepath = document.getElementById('editor-filepath').value;
    const content = document.getElementById('code-editor').value;
    
    try {
        const res = await fetch('/api/files/save', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                path: filepath,
                content: content
            })
        });
        
        if (res.ok) {
            alert('✓ 文件已保存');
        } else {
            const error = await res.text();
            alert('✗ 保存失败: ' + error);
        }
    } catch (err) {
        alert('✗ 保存失败: ' + err.message);
    }
}

// 关闭编辑器
function closeEditor() {
    if (document.getElementById('code-editor').value !== '') {
        if (!confirm('确定要关闭编辑器吗？未保存的更改将丢失。')) {
            return;
        }
    }
    document.getElementById('editor-modal').style.display = 'none';
}

// 检测语言（用于语法高亮提示）
function detectLanguage(filename) {
    const ext = filename.split('.').pop().toLowerCase();
    const languages = {
        'js': 'JavaScript',
        'json': 'JSON',
        'html': 'HTML',
        'css': 'CSS',
        'php': 'PHP',
        'py': 'Python',
        'go': 'Go',
        'java': 'Java',
        'md': 'Markdown',
        'sql': 'SQL',
        'xml': 'XML',
        'yml': 'YAML',
        'yaml': 'YAML'
    };
    
    const lang = languages[ext] || 'Text';
    document.getElementById('editor-language').textContent = lang;
}

// 编辑器快捷键支持
function setupEditorShortcuts() {
    const editor = document.getElementById('code-editor');
    if (!editor) return;
    
    editor.addEventListener('keydown', function(e) {
        // Ctrl+S 保存
        if (e.ctrlKey && e.key === 's') {
            e.preventDefault();
            saveFile();
        }
        
        // Tab 键插入空格
        if (e.key === 'Tab') {
            e.preventDefault();
            const start = this.selectionStart;
            const end = this.selectionEnd;
            this.value = this.value.substring(0, start) + '    ' + this.value.substring(end);
            this.selectionStart = this.selectionEnd = start + 4;
        }
    });
}

// 页面加载时初始化
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', setupEditorShortcuts);
} else {
    setupEditorShortcuts();
}

// 替换原有的 loadFiles 函数
window.loadFiles = loadFilesEnhanced;
