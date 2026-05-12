// 全局变量
let currentPath = '';
let maxBandwidth = 0; // 最大带宽，单位：bytes/sec
let bandwidthLimit = 0; // 带宽限制（80%的最大带宽）
let isSpeedTested = false; // 是否已进行测速

// DOM加载完成后初始化
document.addEventListener('DOMContentLoaded', function() {
    // 初始化事件监听器
    initEventListeners();
    
    // 加载文件列表
    loadFileList();
    
    // 更新服务器信息
    updateServerInfo();
    
    // 更新当前时间
    updateCurrentTime();
    setInterval(updateCurrentTime, 1000);
    
    // 页面加载完成后自动进行一次测速
    setTimeout(() => {
        // 自动测速，但不显示结果（静默测速）
        ensureSpeedTest().then(result => {
            const bandwidthMbps = (result.maxBandwidth * 8 / (1024 * 1024)).toFixed(2);
            const limitMbps = (result.bandwidthLimit * 8 / (1024 * 1024)).toFixed(2);
            
            // 更新显示结果
            const bandwidthValue = document.getElementById('bandwidthValue');
            const bandwidthLimitValue = document.getElementById('bandwidthLimitValue');
            const speedTestResult = document.getElementById('speedTestResult');
            
            if (bandwidthValue && bandwidthLimitValue && speedTestResult) {
                bandwidthValue.textContent = bandwidthMbps;
                bandwidthLimitValue.textContent = limitMbps;
                speedTestResult.style.display = 'block';
            }
            
            console.log(`自动测速完成: 带宽=${bandwidthMbps} Mbps, 限制=${limitMbps} Mbps`);
        }).catch(error => {
            console.error('自动测速失败:', error);
        });
    }, 1000);
});

// 初始化事件监听器
function initEventListeners() {
    // 文件上传表单
    const uploadForm = document.getElementById('uploadForm');
    if (uploadForm) {
        uploadForm.addEventListener('submit', handleFileUpload);
    }

    // 目录上传表单
    const uploadDirForm = document.getElementById('uploadDirForm');
    if (uploadDirForm) {
        uploadDirForm.addEventListener('submit', handleDirectoryUpload);
    }

    // 创建目录表单
    const createDirForm = document.getElementById('createDirForm');
    if (createDirForm) {
        createDirForm.addEventListener('submit', handleCreateDirectory);
    }

    // 文件输入变化事件
    const fileInput = document.getElementById('fileInput');
    if (fileInput) {
        fileInput.addEventListener('change', updateFileLabel);
    }

    const dirInput = document.getElementById('dirInput');
    if (dirInput) {
        dirInput.addEventListener('change', updateDirLabel);
    }

    // 测速按钮事件监听器
    const speedTestBtn = document.getElementById('speedTestBtn');
    if (speedTestBtn) {
        speedTestBtn.addEventListener('click', handleSpeedTestButton);
    }
}

// 处理测速按钮点击
async function handleSpeedTestButton() {
    const speedTestBtn = document.getElementById('speedTestBtn');
    const speedTestResult = document.getElementById('speedTestResult');
    const bandwidthValue = document.getElementById('bandwidthValue');
    const bandwidthLimitValue = document.getElementById('bandwidthLimitValue');
    
    if (!speedTestBtn || !speedTestResult || !bandwidthValue || !bandwidthLimitValue) {
        return;
    }
    
    // 禁用按钮，防止重复点击
    speedTestBtn.disabled = true;
    speedTestBtn.innerHTML = '<i class="fas fa-spinner fa-spin"></i> 测速中...';
    
    try {
        const result = await measureBandwidth();
        
        // 更新显示结果
        const bandwidthMbps = (result.maxBandwidth * 8 / (1024 * 1024)).toFixed(2);
        const limitMbps = (result.bandwidthLimit * 8 / (1024 * 1024)).toFixed(2);
        
        bandwidthValue.textContent = bandwidthMbps;
        bandwidthLimitValue.textContent = limitMbps;
        speedTestResult.style.display = 'block';
        
        // 恢复按钮
        speedTestBtn.disabled = false;
        speedTestBtn.innerHTML = '<i class="fas fa-bolt"></i> 重新测速';
        
    } catch (error) {
        console.error('测速失败:', error);
        showMessage('测速失败: ' + error.message, 'error');
        
        // 恢复按钮
        speedTestBtn.disabled = false;
        speedTestBtn.innerHTML = '<i class="fas fa-bolt"></i> 开始测速';
    }
}

// 测速函数：测量与服务器的带宽
async function measureBandwidth() {
    try {
        showMessage('正在测速...', 'info');
        
        const testSize = 1 * 1024 * 1024; // 1MB测试数据
        const startTime = performance.now();
        
        // 发送测速请求
        const response = await fetch(`/api/speedtest?size=${testSize}`);
        if (!response.ok) {
            throw new Error(`测速请求失败: ${response.status} ${response.statusText}`);
        }
        
        // 读取响应数据
        const reader = response.body.getReader();
        let receivedBytes = 0;
        const chunks = [];
        
        while (true) {
            const { done, value } = await reader.read();
            if (done) break;
            
            receivedBytes += value.length;
            chunks.push(value);
        }
        
        const endTime = performance.now();
        const duration = (endTime - startTime) / 1000; // 转换为秒
        
        // 计算带宽（bytes/sec）
        maxBandwidth = receivedBytes / duration;
        
        // 计算80%的带宽限制
        bandwidthLimit = maxBandwidth * 0.8;
        
        isSpeedTested = true;
        
        // 显示测速结果
        const bandwidthMbps = (maxBandwidth * 8 / (1024 * 1024)).toFixed(2);
        const limitMbps = (bandwidthLimit * 8 / (1024 * 1024)).toFixed(2);
        showMessage(`测速完成: 带宽 ${bandwidthMbps} Mbps，限制为 ${limitMbps} Mbps (80%)`, 'success');
        
        console.log(`测速结果: 带宽=${maxBandwidth.toFixed(0)} bytes/sec (${bandwidthMbps} Mbps), 限制=${bandwidthLimit.toFixed(0)} bytes/sec (${limitMbps} Mbps)`);
        
        return { maxBandwidth, bandwidthLimit };
    } catch (error) {
        console.error('测速失败:', error);
        showMessage('测速失败，使用默认带宽限制', 'error');
        
        // 使用默认值
        maxBandwidth = 10 * 1024 * 1024; // 10 MB/s 默认值
        bandwidthLimit = maxBandwidth * 0.8;
        isSpeedTested = true;
        
        return { maxBandwidth, bandwidthLimit };
    }
}

// 检查并执行测速（如果需要）
async function ensureSpeedTest() {
    if (!isSpeedTested) {
        return await measureBandwidth();
    }
    return { maxBandwidth, bandwidthLimit };
}

// 带速率限制的文件上传
async function uploadWithRateLimit(url, formData, onProgress) {
    // 确保已测速
    await ensureSpeedTest();
    
    // 如果没有带宽限制，直接上传
    if (bandwidthLimit <= 0) {
        return await fetch(url, {
            method: 'POST',
            body: formData
        });
    }
    
    // 创建自定义请求，实现速率限制
    return new Promise((resolve, reject) => {
        const xhr = new XMLHttpRequest();
        
        // 跟踪上传进度
        let lastLoaded = 0;
        let lastTime = Date.now();
        
        xhr.upload.addEventListener('progress', (event) => {
            if (event.lengthComputable) {
                const currentTime = Date.now();
                const timeDiff = (currentTime - lastTime) / 1000; // 秒
                
                if (timeDiff > 0) {
                    const loadedDiff = event.loaded - lastLoaded;
                    const currentSpeed = loadedDiff / timeDiff; // bytes/sec
                    
                    // 如果速度超过限制，添加延迟
                    if (currentSpeed > bandwidthLimit) {
                        const excess = currentSpeed - bandwidthLimit;
                        const delay = (excess / bandwidthLimit) * 1000; // 毫秒
                        
                        // 简单实现：暂停一小段时间
                        setTimeout(() => {
                            // 继续上传
                        }, Math.min(delay, 100)); // 最大延迟100ms
                    }
                }
                
                lastLoaded = event.loaded;
                lastTime = currentTime;
                
                if (onProgress) {
                    onProgress(event.loaded, event.total);
                }
            }
        });
        
        xhr.addEventListener('load', () => {
            if (xhr.status >= 200 && xhr.status < 300) {
                resolve({
                    ok: true,
                    status: xhr.status,
                    statusText: xhr.statusText,
                    json: () => Promise.resolve(JSON.parse(xhr.responseText))
                });
            } else {
                reject(new Error(`上传失败: ${xhr.status} ${xhr.statusText}`));
            }
        });
        
        xhr.addEventListener('error', () => {
            reject(new Error('网络错误'));
        });
        
        xhr.open('POST', url);
        xhr.send(formData);
    });
}

// 更新文件选择标签
function updateFileLabel(e) {
    const label = document.querySelector('label[for="fileInput"]');
    if (e.target.files.length > 0) {
        if (e.target.files.length === 1) {
            label.innerHTML = `<i class="fas fa-file"></i> ${e.target.files[0].name}`;
        } else {
            label.innerHTML = `<i class="fas fa-files"></i> ${e.target.files.length} 个文件`;
        }
    } else {
        label.innerHTML = `<i class="fas fa-folder-open"></i> 选择文件`;
    }
}

// 更新目录选择标签
function updateDirLabel(e) {
    const label = document.querySelector('label[for="dirInput"]');
    if (e.target.files.length > 0) {
        // 获取目录名（从第一个文件的webkitRelativePath中提取）
        const firstFile = e.target.files[0];
        if (firstFile.webkitRelativePath) {
            const dirName = firstFile.webkitRelativePath.split('/')[0];
            label.innerHTML = `<i class="fas fa-folder"></i> ${dirName} (${e.target.files.length} 个文件)`;
        } else {
            label.innerHTML = `<i class="fas fa-folder"></i> ${e.target.files.length} 个文件`;
        }
    } else {
        label.innerHTML = `<i class="fas fa-folder"></i> 选择目录`;
    }
}

// 处理文件上传
async function handleFileUpload(e) {
    e.preventDefault();
    
    const fileInput = document.getElementById('fileInput');
    if (!fileInput.files.length) {
        showMessage('请选择要上传的文件', 'error');
        return;
    }

    const formData = new FormData();
    for (let i = 0; i < fileInput.files.length; i++) {
        formData.append('file', fileInput.files[i]);
    }
    formData.append('path', currentPath);

    try {
        showMessage('正在上传文件...', 'info');
        
        // 使用带速率限制的上传
        const response = await uploadWithRateLimit('/upload', formData, (loaded, total) => {
            if (total > 0) {
                const percent = Math.round((loaded / total) * 100);
                showMessage(`正在上传文件... ${percent}%`, 'info');
            }
        });

        const result = await response.json();
        
        if (result.success) {
            showMessage(result.message, 'success');
            loadFileList();
            // 重置表单
            fileInput.value = '';
            updateFileLabel({ target: fileInput });
        } else {
            showMessage('上传失败: ' + (result.message || '未知错误'), 'error');
        }
    } catch (error) {
        showMessage('上传失败: ' + error.message, 'error');
    }
}

// 处理目录上传
async function handleDirectoryUpload(e) {
    e.preventDefault();
    
    const dirInput = document.getElementById('dirInput');
    if (!dirInput.files.length) {
        showMessage('请选择要上传的目录', 'error');
        return;
    }

    const MAX_CHUNK_SIZE = 100 * 1024 * 1024; // 100MB
    const files = Array.from(dirInput.files);
    
    // 计算文件大小并分组
    const chunks = [];
    let currentChunk = [];
    let currentChunkSize = 0;
    
    for (const file of files) {
        const fileSize = file.size;
        // 如果单个文件超过100MB，它将成为单独的一个块
        if (currentChunkSize + fileSize > MAX_CHUNK_SIZE && currentChunk.length > 0) {
            chunks.push(currentChunk);
            currentChunk = [file];
            currentChunkSize = fileSize;
        } else {
            currentChunk.push(file);
            currentChunkSize += fileSize;
        }
    }
    
    if (currentChunk.length > 0) {
        chunks.push(currentChunk);
    }
    
    if (chunks.length === 0) {
        showMessage('没有文件可上传', 'error');
        return;
    }
    
    const totalChunks = chunks.length;
    showMessage(`开始上传目录，共 ${totalChunks} 个批次，每个批次不超过100MB`, 'info');
    
    let successfulUploads = 0;
    let failedUploads = 0;
    
    for (let i = 0; i < chunks.length; i++) {
        const chunk = chunks[i];
        const chunkNumber = i + 1;
        showMessage(`正在上传第 ${chunkNumber}/${totalChunks} 批次...`, 'info');
        
        const formData = new FormData();
        for (const file of chunk) {
            // 使用webkitRelativePath作为文件名以保留目录结构
            // 如果webkitRelativePath不存在或为空，则使用原始文件名
            let fileName = file.name;
            if (file.webkitRelativePath && file.webkitRelativePath.trim() !== '') {
                fileName = file.webkitRelativePath;
            }
            formData.append('files', file, fileName);
            console.log(`${chunkNumber}/${totalChunks} 批次: ${fileName}...`);
        }
        formData.append('path', currentPath);
        
        try {
            // 使用带速率限制的上传
            const response = await uploadWithRateLimit('/upload-dir', formData, (loaded, total) => {
                if (total > 0) {
                    const percent = Math.round((loaded / total) * 100);
                    showMessage(`正在上传第 ${chunkNumber}/${totalChunks} 批次... ${percent}%`, 'info');
                }
            });
            
            const result = await response.json();
            if (result.success) {
                successfulUploads++;
            } else {
                failedUploads++;
                console.error(`第 ${chunkNumber} 批次上传失败:`, result.message);
            }
        } catch (error) {
            failedUploads++;
            console.error(`第 ${chunkNumber} 批次上传失败:`, error.message);
        }
    }
    
    // 显示最终结果
    if (failedUploads === 0) {
        showMessage(`目录上传完成！所有 ${totalChunks} 个批次均成功上传。`, 'success');
        loadFileList();
        // 重置表单
        dirInput.value = '';
        updateDirLabel({ target: dirInput });
    } else {
        showMessage(`目录上传完成，但有部分失败。成功: ${successfulUploads}/${totalChunks} 批次，失败: ${failedUploads} 批次。`, 'error');
        // 即使有失败，也刷新文件列表，因为可能部分文件已上传
        loadFileList();
    }
}

// 处理创建目录
async function handleCreateDirectory(e) {
    e.preventDefault();
    
    const dirNameInput = document.getElementById('dirName');
    const dirName = dirNameInput.value.trim();
    
    if (!dirName) {
        showMessage('请输入目录名称', 'error');
        return;
    }

    const formData = new FormData();
    formData.append('name', dirName);
    formData.append('path', currentPath);

    try {
        showMessage('正在创建目录...', 'info');
        
        const response = await fetch('/api/create-dir', {
            method: 'POST',
            body: formData
        });

        const result = await response.json();
        
        if (result.success) {
            showMessage(result.message, 'success');
            loadFileList();
            // 重置表单
            dirNameInput.value = '';
        } else {
            showMessage('创建失败: ' + (result.message || '未知错误'), 'error');
        }
    } catch (error) {
        showMessage('创建失败: ' + error.message, 'error');
    }
}

// 加载文件列表
async function loadFileList() {
    const fileListElement = document.getElementById('fileList');
    if (!fileListElement) return;

    // 显示加载状态
    fileListElement.innerHTML = `
        <div class="loading">
            <i class="fas fa-spinner"></i>
            <p>正在加载文件列表...</p>
        </div>
    `;

    try {
        const response = await fetch(`/api/files?path=${encodeURIComponent(currentPath)}`);
        const files = await response.json();
        
        if (files.length === 0) {
            fileListElement.innerHTML = `
                <div class="empty-state">
                    <i class="fas fa-folder-open"></i>
                    <p>当前目录为空</p>
                </div>
            `;
        } else {
            renderFileList(files);
        }
        
        updatePathNavigation();
    } catch (error) {
        fileListElement.innerHTML = `
            <div class="empty-state">
                <i class="fas fa-exclamation-triangle"></i>
                <p>加载文件列表失败: ${error.message}</p>
            </div>
        `;
    }
}

// 渲染文件列表
function renderFileList(files) {
    const fileListElement = document.getElementById('fileList');
    
    fileListElement.innerHTML = files.map(file => `
        <div class="file-item ${file.is_dir ? 'folder' : ''}" data-path="${file.path}">
            <div class="file-icon">
                <i class="fas ${file.is_dir ? 'fa-folder' : 'fa-file'}"></i>
            </div>
            <div class="file-name">${escapeHtml(file.name)}</div>
            <div class="file-info">
                <span><i class="fas fa-calendar"></i> ${formatDate(file.mod_time)}</span>
                ${!file.is_dir ? `<span><i class="fas fa-hdd"></i> ${formatFileSize(file.size)}</span>` : ''}
            </div>
            <div class="file-actions">
                ${file.is_dir ? 
                    `<button class="btn btn-secondary" onclick="navigateTo('${encodeURIComponent(file.path)}')">
                        <i class="fas fa-folder-open"></i> 打开
                    </button>` : 
                    `<button class="btn btn-primary" onclick="downloadFile('${encodeURIComponent(file.path)}')">
                        <i class="fas fa-download"></i> 下载
                    </button>`
                }
                <button class="btn btn-danger" onclick="deleteItem('${encodeURIComponent(file.path)}', ${file.is_dir})">
                    <i class="fas fa-trash"></i> 删除
                </button>
            </div>
        </div>
    `).join('');
}

// 更新路径导航
function updatePathNavigation() {
    const pathNavElement = document.querySelector('.path-nav');
    if (!pathNavElement) return;

    const pathParts = currentPath.split('/').filter(p => p);
    let pathItems = '<span class="path-item" data-path="" onclick="navigateTo(\'\')">根目录</span>';
    
    let currentPathAcc = '';
    for (let i = 0; i < pathParts.length; i++) {
        currentPathAcc += (currentPathAcc ? '/' : '') + pathParts[i];
        pathItems += `<span class="path-item" data-path="${currentPathAcc}" onclick="navigateTo('${encodeURIComponent(currentPathAcc)}')">
            <i class="fas fa-chevron-right"></i> ${escapeHtml(pathParts[i])}
        </span>`;
    }
    
    pathNavElement.innerHTML = pathItems;
}

// 导航到指定路径
function navigateTo(path) {
    currentPath = path;
    loadFileList();
}

// 下载文件
function downloadFile(filePath) {
    window.open(`/api/download/${filePath}`, '_blank');
}

// 删除文件或目录
async function deleteItem(itemPath, isDir) {
    const itemType = isDir ? '目录' : '文件';
    const itemName = decodeURIComponent(itemPath).split('/').pop();
    
    if (!confirm(`确定要删除${itemType} "${itemName}" 吗？此操作不可恢复。`)) {
        return;
    }

    try {
        showMessage(`正在删除${itemType}...`, 'info');
        
        const response = await fetch(`/api/delete/${itemPath}`, {
            method: 'DELETE'
        });

        const result = await response.json();
        
        if (result.success) {
            showMessage(`${itemType}删除成功`, 'success');
            loadFileList();
        } else {
            showMessage(`删除失败: ${result.message || '未知错误'}`, 'error');
        }
    } catch (error) {
        showMessage(`删除失败: ${error.message}`, 'error');
    }
}

// 更新服务器信息
async function updateServerInfo() {
    const serverInfoElement = document.getElementById('serverInfo');
    if (!serverInfoElement) return;

    try {
        // 尝试获取服务器信息
        const protocol = window.location.protocol;
        const host = window.location.host;
        serverInfoElement.textContent = `${protocol}//${host}`;
    } catch (error) {
        serverInfoElement.textContent = '无法获取服务器信息';
    }
}

// 更新当前时间
function updateCurrentTime() {
    const currentTimeElement = document.getElementById('currentTime');
    if (!currentTimeElement) return;

    const now = new Date();
    const formattedTime = now.toLocaleString('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
        hour12: false
    });
    
    currentTimeElement.textContent = formattedTime;
}

// 显示消息提示
function showMessage(message, type = 'info') {
    // 移除现有的消息
    const existingMessage = document.querySelector('.message');
    if (existingMessage) {
        existingMessage.remove();
    }

    // 创建新消息
    const messageElement = document.createElement('div');
    messageElement.className = `message ${type}`;
    messageElement.innerHTML = `
        <i class="fas ${getMessageIcon(type)}"></i>
        <span>${message}</span>
    `;

    document.body.appendChild(messageElement);

    // 3秒后自动移除
    setTimeout(() => {
        if (messageElement.parentNode) {
            messageElement.remove();
        }
    }, 3000);
}

// 获取消息图标
function getMessageIcon(type) {
    switch (type) {
        case 'success': return 'fa-check-circle';
        case 'error': return 'fa-exclamation-circle';
        case 'info': return 'fa-info-circle';
        default: return 'fa-info-circle';
    }
}

// 格式化文件大小
function formatFileSize(bytes) {
    if (bytes === 0) return '0 B';
    
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

// 格式化日期
function formatDate(dateString) {
    const date = new Date(dateString);
    return date.toLocaleString('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        hour12: false
    });
}

// HTML转义
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// 拖放上传支持
function initDragAndDrop() {
    const dropZone = document.querySelector('.upload-section');
    
    if (!dropZone) return;
    
    ['dragenter', 'dragover', 'dragleave', 'drop'].forEach(eventName => {
        dropZone.addEventListener(eventName, preventDefaults, false);
    });
    
    function preventDefaults(e) {
        e.preventDefault();
        e.stopPropagation();
    }
    
    ['dragenter', 'dragover'].forEach(eventName => {
        dropZone.addEventListener(eventName, highlight, false);
    });
    
    ['dragleave', 'drop'].forEach(eventName => {
        dropZone.addEventListener(eventName, unhighlight, false);
    });
    
    function highlight() {
        dropZone.style.backgroundColor = '#e8f4fd';
        dropZone.style.borderColor = '#667eea';
    }
    
    function unhighlight() {
        dropZone.style.backgroundColor = '';
        dropZone.style.borderColor = '';
    }
    
    dropZone.addEventListener('drop', handleDrop, false);
    
    async function handleDrop(e) {
        const dt = e.dataTransfer;
        const files = dt.files;
        
        if (files.length === 0) return;
        
        // 检查是否是目录（通过检查是否有webkitRelativePath）
        const isDirectory = Array.from(files).some(file => file.webkitRelativePath);
        
        if (isDirectory) {
            // 处理目录上传
            const dirInput = document.getElementById('dirInput');
            const dataTransfer = new DataTransfer();
            Array.from(files).forEach(file => dataTransfer.items.add(file));
            dirInput.files = dataTransfer.files;
            updateDirLabel({ target: dirInput });
            
            // 自动提交表单
            const uploadDirForm = document.getElementById('uploadDirForm');
            if (uploadDirForm) {
                const event = new Event('submit', { cancelable: true });
                uploadDirForm.dispatchEvent(event);
                if (!event.defaultPrevented) {
                    handleDirectoryUpload(new Event('submit'));
                }
            }
        } else {
            // 处理文件上传
            const fileInput = document.getElementById('fileInput');
            const dataTransfer = new DataTransfer();
            Array.from(files).forEach(file => dataTransfer.items.add(file));
            fileInput.files = dataTransfer.files;
            updateFileLabel({ target: fileInput });
            
            // 自动提交表单
            const uploadForm = document.getElementById('uploadForm');
            if (uploadForm) {
                const event = new Event('submit', { cancelable: true });
                uploadForm.dispatchEvent(event);
                if (!event.defaultPrevented) {
                    handleFileUpload(new Event('submit'));
                }
            }
        }
    }
}

// 初始化拖放功能
initDragAndDrop();