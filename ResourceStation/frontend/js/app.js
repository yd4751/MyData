// Resource Station 前端JavaScript
// 主要功能：用户认证、文件上传、资源管理

const API_BASE_URL = `${window.location.protocol}//${window.location.host}/api/v1`;
let currentToken = localStorage.getItem('token');
let currentUser = JSON.parse(localStorage.getItem('user') || '{}');

// DOM元素
const authButtons = document.getElementById('auth-buttons');
const authForms = document.getElementById('auth-forms');
const resourceManager = document.getElementById('resource-manager');
const loginForm = document.getElementById('login-form');
const registerForm = document.getElementById('register-form');
const logoutBtn = document.getElementById('logout-btn');
const uploadArea = document.getElementById('upload-area');
const fileInput = document.getElementById('file-input');
const startUploadBtn = document.getElementById('start-upload');
const resourceTable = document.getElementById('resource-table');
const searchInput = document.getElementById('search-input');
const refreshBtn = document.getElementById('refresh-btn');
const uploadProgressCard = document.getElementById('upload-progress-card');
const uploadProgress = document.getElementById('upload-progress');

// 导航栏元素
const homeNav = document.getElementById('home-nav');
const categoryNav = document.getElementById('category-nav');
const resourcesNav = document.getElementById('resources-nav');
const settingsNav = document.getElementById('settings-nav');

// 初始化
document.addEventListener('DOMContentLoaded', () => {
    initEventListeners();
    checkAuthStatus();
    loadResources();
});

// 初始化事件监听器
function initEventListeners() {
    // 登录表单提交
    if (loginForm) {
        loginForm.addEventListener('submit', handleLogin);
    }

    // 注册表单提交
    if (registerForm) {
        registerForm.addEventListener('submit', handleRegister);
    }

    // 退出登录
    if (logoutBtn) {
        logoutBtn.addEventListener('click', handleLogout);
    }

    // 导航栏点击事件
    if (homeNav) homeNav.addEventListener('click', () => setActiveNav(homeNav));
    if (categoryNav) categoryNav.addEventListener('click', () => setActiveNav(categoryNav));
    if (resourcesNav) resourcesNav.addEventListener('click', () => setActiveNav(resourcesNav));
    if (settingsNav) settingsNav.addEventListener('click', () => setActiveNav(settingsNav));

    // 上传区域点击事件
    if (uploadArea) {
        uploadArea.addEventListener('click', () => fileInput.click());
        uploadArea.addEventListener('dragover', handleDragOver);
        uploadArea.addEventListener('drop', handleDrop);
    }

    // 文件选择事件
    if (fileInput) {
        fileInput.addEventListener('change', handleFileSelect);
    }

    // 开始上传按钮
    if (startUploadBtn) {
        startUploadBtn.addEventListener('click', handleStartUpload);
    }

    // 搜索输入
    if (searchInput) {
        searchInput.addEventListener('input', debounce(handleSearch, 300));
    }

    // 刷新按钮
    if (refreshBtn) {
        refreshBtn.addEventListener('click', loadResources);
    }
}

// 检查认证状态
function checkAuthStatus() {
    if (currentToken && currentUser.id) {
        showResourceManager();
        updateUserInfo();
    } else {
        showAuthForms();
    }
}

// 显示认证表单
function showAuthForms() {
    if (authForms) authForms.classList.remove('d-none');
    if (resourceManager) resourceManager.classList.add('d-none');
    updateAuthButtons();
}

// 显示资源管理器
function showResourceManager() {
    if (authForms) authForms.classList.add('d-none');
    if (resourceManager) resourceManager.classList.remove('d-none');
    updateAuthButtons();
}

// 更新认证按钮
function updateAuthButtons() {
    if (!authButtons) return;
    
    if (currentToken && currentUser.id) {
        authButtons.innerHTML = `
            <span class="navbar-text me-3">
                <i class="bi bi-person-circle"></i> ${currentUser.username}
            </span>
            <button class="btn btn-outline-light" id="logout-nav-btn">
                <i class="bi bi-box-arrow-right"></i> 退出
            </button>
        `;
        document.getElementById('logout-nav-btn')?.addEventListener('click', handleLogout);
    } else {
        authButtons.innerHTML = `
            <button class="btn btn-outline-light me-2" id="login-nav-btn">
                <i class="bi bi-box-arrow-in-right"></i> 登录
            </button>
            <button class="btn btn-primary" id="register-nav-btn">
                <i class="bi bi-person-plus"></i> 注册
            </button>
        `;
        document.getElementById('login-nav-btn')?.addEventListener('click', () => {
            showAuthForms();
            document.getElementById('login-tab').click();
        });
        document.getElementById('register-nav-btn')?.addEventListener('click', () => {
            showAuthForms();
            document.getElementById('register-tab').click();
        });
    }
}

// 更新用户信息
function updateUserInfo() {
    const userInfo = document.getElementById('user-info');
    if (!userInfo) return;
    
    userInfo.innerHTML = `
        <h5>${currentUser.username}</h5>
        <p class="text-muted mb-1">${currentUser.email || ''}</p>
        <p class="text-muted mb-0">
            <small>角色: ${currentUser.role === 'admin' ? '管理员' : '用户'}</small>
        </p>
    `;
}

// 处理登录
async function handleLogin(e) {
    e.preventDefault();
    
    const username = document.getElementById('login-username').value;
    const password = document.getElementById('login-password').value;
    
    try {
        const response = await fetch(`${API_BASE_URL}/auth/login`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username, password })
        });
        
        const data = await response.json();
        
        if (response.ok) {
            currentToken = data.token;
            currentUser = data.user;
            localStorage.setItem('token', currentToken);
            localStorage.setItem('user', JSON.stringify(currentUser));
            showResourceManager();
            updateUserInfo();
            loadResources();
            showToast('登录成功', 'success');
        } else {
            showToast(data.error || '登录失败', 'error');
        }
    } catch (error) {
        showToast('网络错误，请稍后重试', 'error');
    }
}

// 处理注册
async function handleRegister(e) {
    e.preventDefault();
    
    const username = document.getElementById('register-username').value;
    const email = document.getElementById('register-email').value;
    const password = document.getElementById('register-password').value;
    
    try {
        const response = await fetch(`${API_BASE_URL}/auth/register`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username, email, password })
        });
        
        const data = await response.json();
        
        if (response.ok) {
            showToast('注册成功，请登录', 'success');
            document.getElementById('login-tab').click();
            // 清空注册表单
            document.getElementById('register-form').reset();
        } else {
            showToast(data.error || '注册失败', 'error');
        }
    } catch (error) {
        showToast('网络错误，请稍后重试', 'error');
    }
}

// 处理退出登录
function handleLogout() {
    currentToken = null;
    currentUser = {};
    localStorage.removeItem('token');
    localStorage.removeItem('user');
    showAuthForms();
    showToast('已退出登录', 'info');
}

// 处理拖放
function handleDragOver(e) {
    e.preventDefault();
    e.stopPropagation();
    uploadArea.classList.add('border-primary');
}

function handleDrop(e) {
    e.preventDefault();
    e.stopPropagation();
    uploadArea.classList.remove('border-primary');
    
    const files = e.dataTransfer.files;
    if (files.length > 0) {
        handleFiles(files);
    }
}

// 处理文件选择
function handleFileSelect(e) {
    const files = e.target.files;
    if (files.length > 0) {
        handleFiles(files);
    }
}

// 处理文件
function handleFiles(files) {
    // 清除之前的预览
    clearFilePreviews();
    
    // 显示文件预览
    const fileList = Array.from(files);
    showToast(`已选择 ${files.length} 个文件`, 'info');
    
    // 为每个文件创建预览
    fileList.forEach((file, index) => {
        createFilePreview(file, index);
    });
    
    // 更新文件输入
    updateFileInput(files);
}

// 清除文件预览
function clearFilePreviews() {
    const previewContainer = document.getElementById('file-preview-container');
    if (previewContainer) {
        previewContainer.remove();
    }
}

// 创建文件预览
function createFilePreview(file, index) {
    // 创建或获取预览容器
    let previewContainer = document.getElementById('file-preview-container');
    if (!previewContainer) {
        previewContainer = document.createElement('div');
        previewContainer.id = 'file-preview-container';
        previewContainer.className = 'file-preview-container mt-3';
        uploadArea.parentNode.insertBefore(previewContainer, uploadArea.nextSibling);
    }
    
    // 创建预览卡片
    const previewCard = document.createElement('div');
    previewCard.className = 'file-preview-card';
    previewCard.dataset.index = index;
    
    // 根据文件类型创建预览内容
    if (file.type.startsWith('image/')) {
        // 图片文件：显示缩略图
        const reader = new FileReader();
        reader.onload = function(e) {
            const img = document.createElement('img');
            img.src = e.target.result;
            img.className = 'file-preview-thumbnail';
            img.alt = file.name;
            
            const thumbnailContainer = previewCard.querySelector('.file-preview-thumbnail-container');
            if (thumbnailContainer) {
                thumbnailContainer.appendChild(img);
            }
        };
        reader.readAsDataURL(file);
        
        previewCard.innerHTML = `
            <div class="file-preview-content">
                <div class="file-preview-thumbnail-container d-flex justify-content-center align-items-center">
                    <div class="spinner-border spinner-border-sm text-primary" role="status">
                        <span class="visually-hidden">加载中...</span>
                    </div>
                </div>
                <div class="file-preview-info">
                    <div class="file-preview-name">${escapeHtml(file.name)}</div>
                    <div class="file-preview-meta">
                        <span class="file-preview-type">${getFileTypeText(file.type)}</span>
                        <span class="file-preview-size">${formatFileSize(file.size)}</span>
                    </div>
                </div>
                <button type="button" class="btn btn-sm btn-outline-danger file-preview-remove" data-index="${index}">
                    <i class="bi bi-x"></i>
                </button>
            </div>
        `;
    } else {
        // 非图片文件：显示文件图标
        const fileIcon = getFileIcon(file.type);
        
        previewCard.innerHTML = `
            <div class="file-preview-content">
                <div class="file-preview-icon">
                    <i class="bi ${fileIcon}"></i>
                </div>
                <div class="file-preview-info">
                    <div class="file-preview-name">${escapeHtml(file.name)}</div>
                    <div class="file-preview-meta">
                        <span class="file-preview-type">${getFileTypeText(file.type)}</span>
                        <span class="file-preview-size">${formatFileSize(file.size)}</span>
                    </div>
                </div>
                <button type="button" class="btn btn-sm btn-outline-danger file-preview-remove" data-index="${index}">
                    <i class="bi bi-x"></i>
                </button>
            </div>
        `;
    }
    
    previewContainer.appendChild(previewCard);
    
    // 添加删除按钮事件
    const removeBtn = previewCard.querySelector('.file-preview-remove');
    removeBtn.addEventListener('click', function() {
        removeFilePreview(index);
    });
}

// 根据文件类型获取图标
function getFileIcon(fileType) {
    if (fileType.startsWith('image/')) return 'bi-file-image';
    if (fileType.startsWith('video/')) return 'bi-file-play';
    if (fileType.startsWith('audio/')) return 'bi-file-music';
    if (fileType === 'application/pdf') return 'bi-file-pdf';
    if (fileType.includes('#document') || fileType.includes('msword') || fileType.includes('wordprocessingml')) return 'bi-file-word';
    if (fileType.includes('spreadsheetml') || fileType.includes('excel')) return 'bi-file-excel';
    if (fileType.includes('presentationml') || fileType.includes('powerpoint')) return 'bi-file-ppt';
    if (fileType.includes('zip') || fileType.includes('compressed')) return 'bi-file-zip';
    if (fileType.startsWith('text/')) return 'bi-file-text';
    return 'bi-file-earmark';
}

// 获取文件类型文本
function getFileTypeText(fileType) {
    if (fileType.startsWith('image/')) return '图片';
    if (fileType.startsWith('video/')) return '视频';
    if (fileType.startsWith('audio/')) return '音频';
    if (fileType === 'application/pdf') return 'PDF文档';
    if (fileType.includes('document') || fileType.includes('word')) return 'Word文档';
    if (fileType.includes('spreadsheet') || fileType.includes('excel')) return 'Excel文档';
    if (fileType.includes('presentation') || fileType.includes('powerpoint')) return 'PPT文档';
    if (fileType.includes('zip') || fileType.includes('compressed')) return '压缩文件';
    if (fileType.startsWith('text/')) return '文本文件';
    return '文件';
}

// 格式化文件大小
function formatFileSize(bytes) {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

// HTML转义
function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

// 移除文件预览
function removeFilePreview(index) {
    const previewCard = document.querySelector(`.file-preview-card[data-index="${index}"]`);
    if (previewCard) {
        previewCard.remove();
    }
    
    // 更新文件输入
    updateFileInputAfterRemoval();
}

// 更新文件输入（删除文件后）
function updateFileInputAfterRemoval() {
    // 创建一个新的DataTransfer对象
    const dataTransfer = new DataTransfer();
    
    // 获取所有预览卡片
    const previewCards = document.querySelectorAll('.file-preview-card');
    
    // 如果没有预览卡片了，清空文件输入
    if (previewCards.length === 0) {
        fileInput.value = '';
        clearFilePreviews();
        return;
    }
    
    // 重新构建文件列表
    const files = Array.from(fileInput.files);
    const remainingFiles = [];
    
    // 根据剩余预览卡片筛选文件
    previewCards.forEach(card => {
        const index = parseInt(card.dataset.index);
        if (index >= 0 && index < files.length) {
            remainingFiles.push(files[index]);
        }
    });
    
    // 更新文件输入
    remainingFiles.forEach(file => {
        dataTransfer.items.add(file);
    });
    
    fileInput.files = dataTransfer.files;
}

// 更新文件输入（选择新文件后）
function updateFileInput(files) {
    // 创建一个新的DataTransfer对象
    const dataTransfer = new DataTransfer();
    
    // 添加所有文件
    Array.from(files).forEach(file => {
        dataTransfer.items.add(file);
    });
    
    // 更新文件输入
    fileInput.files = dataTransfer.files;
}

// 处理开始上传
async function handleStartUpload() {
    // 检查用户是否已登录
    if (!currentToken) {
        showToast('请先登录', 'warning');
        showAuthForms();
        return;
    }
    
    const files = fileInput.files;
    if (files.length === 0) {
        showToast('请先选择文件', 'warning');
        return;
    }
    
    const isPublic = document.getElementById('is-public').checked;
    const description = document.getElementById('file-description').value;
    const tags = document.getElementById('file-tags').value;
    const enableChunking = document.getElementById('enable-chunking').checked;
    
    // 显示上传进度
    uploadProgressCard.style.display = 'block';
    uploadProgress.innerHTML = '<div class="text-center"><div class="spinner-border text-primary" role="status"></div><p class="mt-2">准备上传...</p></div>';
    
    // 上传每个文件
    for (let i = 0; i < files.length; i++) {
        const file = files[i];
        try {
            await uploadFile(file, isPublic, description, tags, enableChunking, i + 1, files.length);
        } catch (error) {
            console.error('上传文件失败:', error);
            showToast(`文件 "${file.name}" 上传失败: ${error.message}`, 'error');
        }
    }
    
    // 所有文件上传完成后刷新资源列表
    setTimeout(() => {
        uploadProgressCard.style.display = 'none';
        loadResources();
    }, 1000);
}

// 上传单个文件
async function uploadFile(file, isPublic, description, tags, enableChunking, currentIndex, totalFiles) {
    // 计算文件哈希值 (SHA-256)
    file.hash = ''; // 默认值
    if (true) {
        try {
            const arrayBuffer = await file.arrayBuffer();
            const wordArray = CryptoJS.lib.WordArray.create(arrayBuffer);
            file.hash = CryptoJS.SHA256(wordArray).toString(CryptoJS.enc.Hex);
            console.log(`文件 "${file.name}" 的哈希值: ${file.hash}`);
        } catch (error) {
            console.error('计算文件哈希失败:', error);
        }
    } else {
        console.warn('浏览器不支持Web Crypto API，跳过哈希计算');
    }
    // 更新进度显示
    uploadProgress.innerHTML = `
        <div class="text-center">
            <div class="spinner-border text-primary" role="status"></div>
            <p class="mt-2">正在上传文件 ${currentIndex}/${totalFiles}: ${file.name}</p>
            <div class="progress mt-3">
                <div class="progress-bar" role="progressbar" style="width: 0%" id="upload-progress-bar">0%</div>
            </div>
        </div>
    `;
    
    const progressBar = document.getElementById('upload-progress-bar');
    
    try {
        // 1. 创建资源记录
        const createResponse = await fetch(`${API_BASE_URL}/resources`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${currentToken}`
            },
            body: JSON.stringify({
                filename: file.name,
                file_type: file.type || 'application/octet-stream',
                file_size: file.size,
                description: description,
                tags: tags,
                hash:file.hash,
                is_public: isPublic,
                chunk_count: enableChunking ? Math.ceil(file.size / (10 * 1024 * 1024)) : 1, // 10MB每片
                chunk_size: enableChunking ? 10 * 1024 * 1024 : file.size // 10MB
            })
        });
        
        if (!createResponse.ok) {
            const errorData = await createResponse.json();
            throw new Error(errorData.error || '创建资源失败');
        }
        
        const createData = await createResponse.json();
        const resourceId = createData.upload_id;
        const chunkSize = createData.chunk_size;
        const chunkCount = enableChunking ? Math.ceil(file.size / chunkSize) : 1;
        
        // 2. 上传文件（分片或整体）
        if (enableChunking && chunkCount > 1) {
            // 分片上传
            await uploadFileInChunks(file, resourceId, chunkSize, chunkCount, progressBar);
        } else {
            // 单文件上传
            await uploadFileAsWhole(file, resourceId, progressBar);
        }
        
        // 3. 完成上传
        const completeResponse = await fetch(`${API_BASE_URL}/resources/${resourceId}/complete`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${currentToken}`
            }
        });
        
        if (!completeResponse.ok) {
            const errorData = await completeResponse.json();
            throw new Error(errorData.error || '完成上传失败');
        }
        
        // 更新进度到100%
        if (progressBar) {
            progressBar.style.width = '100%';
            progressBar.textContent = '100%';
        }
        
        showToast(`文件 "${file.name}" 上传成功`, 'success');
        
    } catch (error) {
        console.error('上传文件失败:', error);
        throw error;
    }
}

// 分片上传文件
async function uploadFileInChunks(file, resourceId, chunkSize, chunkCount, progressBar) {
    for (let chunkIndex = 0; chunkIndex < chunkCount; chunkIndex++) {
        const start = chunkIndex * chunkSize;
        const end = Math.min(start + chunkSize, file.size);
        const chunk = file.slice(start, end);
        
        const formData = new FormData();
        formData.append('file', chunk, file.name);
        
        // 更新进度显示
        const chunkProgress = Math.floor((chunkIndex / chunkCount) * 100);
        if (progressBar) {
            progressBar.style.width = `${chunkProgress}%`;
            progressBar.textContent = `${chunkProgress}%`;
        }
        
        const response = await fetch(`${API_BASE_URL}/resources/${resourceId}/chunks/${chunkIndex}`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${currentToken}`
            },
            body: formData
        });
        
        if (!response.ok) {
            const errorData = await response.json();
            throw new Error(`分片 ${chunkIndex + 1}/${chunkCount} 上传失败: ${errorData.error || '未知错误'}`);
        }
        
        // 短暂延迟，避免请求过快
        await new Promise(resolve => setTimeout(resolve, 100));
    }
    
    // 所有分片上传完成
    if (progressBar) {
        progressBar.style.width = '100%';
        progressBar.textContent = '100%';
    }
}

// 单文件上传（不分片）
async function uploadFileAsWhole(file, resourceId, progressBar) {
    const formData = new FormData();
    formData.append('file', file);
    
    // 使用XMLHttpRequest来获取上传进度
    return new Promise((resolve, reject) => {
        const xhr = new XMLHttpRequest();
        
        xhr.upload.addEventListener('progress', (event) => {
            if (event.lengthComputable && progressBar) {
                const percentComplete = Math.round((event.loaded / event.total) * 100);
                progressBar.style.width = `${percentComplete}%`;
                progressBar.textContent = `${percentComplete}%`;
            }
        });
        
        xhr.addEventListener('load', () => {
            if (xhr.status >= 200 && xhr.status < 300) {
                resolve(xhr.response);
            } else {
                reject(new Error(`上传失败: ${xhr.statusText}`));
            }
        });
        
        xhr.addEventListener('error', () => {
            reject(new Error('网络错误'));
        });
        
        xhr.open('POST', `${API_BASE_URL}/resources/${resourceId}/chunks/0`);
        xhr.setRequestHeader('Authorization', `Bearer ${currentToken}`);
        xhr.send(formData);
    });
}

// 加载资源列表
async function loadResources(page = 1, pageSize = 10) {
    if (!currentToken) return;
    
    try {
        const response = await fetch(`${API_BASE_URL}/resources?page=${page}&page_size=${pageSize}`, {
            headers: {
                'Authorization': `Bearer ${currentToken}`
            }
        });
        
        if (response.ok) {
            const data = await response.json();
            renderResourceTable(data.resources || []);
            renderPagination(data.total || 0, page, pageSize);
        } else {
            showToast('加载资源失败', 'error');
        }
    } catch (error) {
        showToast('网络错误，请稍后重试', 'error');
    }
}

// 渲染资源表格
function renderResourceTable(resources) {
    if (!resourceTable) return;
    
    if (resources.length === 0) {
        resourceTable.innerHTML = `
            <tr>
                <td colspan="6" class="text-center text-muted py-4">
                    <i class="bi bi-folder-x"></i> 暂无资源
                </td>
            </tr>
        `;
        return;
    }
    
    resourceTable.innerHTML = resources.map(resource => `
        <tr>
            <td>
                <i class="bi ${getFileIcon(resource.file_type)} me-2"></i>
                ${resource.filename}
                ${resource.is_public ? '<span class="badge bg-success ms-2">公开</span>' : ''}
            </td>
            <td>${getFileTypeName(resource.file_type)}</td>
            <td>${formatFileSize(resource.file_size)}</td>
            <td>
                <span class="badge ${getStatusBadgeClass(resource.status)}">
                    ${getStatusText(resource.status)}
                </span>
            </td>
            <td>${new Date(resource.created_at).toLocaleString()}</td>
            <td>
                <button class="btn btn-sm btn-outline-primary me-1" onclick="viewResource('${resource.id}')">
                    <i class="bi bi-eye"></i>
                </button>
                <button class="btn btn-sm btn-outline-danger" onclick="deleteResource('${resource.id}')">
                    <i class="bi bi-trash"></i>
                </button>
            </td>
        </tr>
    `).join('');
}

// 渲染分页
function renderPagination(total, currentPage, pageSize) {
    const pagination = document.getElementById('pagination');
    if (!pagination) return;
    
    const totalPages = Math.ceil(total / pageSize);
    if (totalPages <= 1) {
        pagination.innerHTML = '';
        return;
    }
    
    let html = '';
    
    // 上一页
    html += `
        <li class="page-item ${currentPage === 1 ? 'disabled' : ''}">
            <a class="page-link" href="#" onclick="loadResources(${currentPage - 1})">上一页</a>
        </li>
    `;
    
    // 页码
    for (let i = 1; i <= totalPages; i++) {
        if (i === 1 || i === totalPages || (i >= currentPage - 2 && i <= currentPage + 2)) {
            html += `
                <li class="page-item ${i === currentPage ? 'active' : ''}">
                    <a class="page-link" href="#" onclick="loadResources(${i})">${i}</a>
                </li>
            `;
        } else if (i === currentPage - 3 || i === currentPage + 3) {
            html += '<li class="page-item disabled"><span class="page-link">...</span></li>';
        }
    }
    
    // 下一页
    html += `
        <li class="page-item ${currentPage === totalPages ? 'disabled' : ''}">
            <a class="page-link" href="#" onclick="loadResources(${currentPage + 1})">下一页</a>
        </li>
    `;
    
    pagination.innerHTML = html;
}

// 处理搜索
function handleSearch() {
    const keyword = searchInput.value.trim();
    // 这里可以实现搜索功能
    console.log('搜索关键词:', keyword);
}

// 查看资源详情
async function viewResource(resourceId) {
    try {
        const response = await fetch(`${API_BASE_URL}/resources/${resourceId}`, {
            headers: {
                'Authorization': `Bearer ${currentToken}`
            }
        });
        
        if (response.ok) {
            const resource = await response.json();
            showResourceModal(resource);
        } else {
            showToast('获取资源详情失败', 'error');
        }
    } catch (error) {
        showToast('网络错误，请稍后重试', 'error');
    }
}

// 显示资源详情模态框
function showResourceModal(resource) {
    const modal = new bootstrap.Modal(document.getElementById('resource-modal'));
    const modalBody = document.getElementById('resource-details');
    const downloadBtn = document.getElementById('download-btn');
    
    modalBody.innerHTML = `
        <div class="row">
            <div class="col-md-4">
                <div class="card">
                    <div class="card-body text-center">
                        <i class="bi ${getFileIcon(resource.file_type)} display-1 text-primary"></i>
                        <h5 class="mt-3">${resource.filename}</h5>
                        <p class="text-muted">${resource.original_name}</p>
                    </div>
                </div>
            </div>
            <div class="col-md-8">
                <table class="table">
                    <tr>
                        <th width="30%">文件类型</th>
                        <td>${getFileTypeName(resource.file_type)}</td>
                    </tr>
                    <tr>
                        <th>文件大小</th>
                        <td>${formatFileSize(resource.file_size)}</td>
                    </tr>
                    <tr>
                        <th>状态</th>
                        <td><span class="badge ${getStatusBadgeClass(resource.status)}">${getStatusText(resource.status)}</span></td>
                    </tr>
                    <tr>
                        <th>上传时间</th>
                        <td>${new Date(resource.created_at).toLocaleString()}</td>
                    </tr>
                    <tr>
                        <th>最后更新</th>
                        <td>${new Date(resource.updated_at).toLocaleString()}</td>
                    </tr>
                    <tr>
                        <th>描述</th>
                        <td>${resource.description || '无'}</td>
                    </tr>
                    <tr>
                        <th>标签</th>
                        <td>${resource.tags || '无'}</td>
                    </tr>
                    <tr>
                        <th>可见性</th>
                        <td>${resource.is_public ? '公开' : '私有'}</td>
                    </tr>
                    <tr>
                        <th>文件哈希</th>
                        <td><code>${resource.hash || '未计算'}</code></td>
                    </tr>
                </table>
            </div>
        </div>
    `;
    
    // 设置下载按钮
    if (resource.status === 'completed' && resource.download_url) {
        downloadBtn.onclick = () => {
            window.open(resource.download_url, '_blank');
        };
        downloadBtn.disabled = false;
    } else {
        downloadBtn.disabled = true;
        downloadBtn.title = '文件未就绪，无法下载';
    }
    
    modal.show();
}

// 删除资源
async function deleteResource(resourceId) {
    if (!confirm('确定要删除这个资源吗？此操作不可撤销。')) {
        return;
    }
    
    try {
        const response = await fetch(`${API_BASE_URL}/resources/${resourceId}`, {
            method: 'DELETE',
            headers: {
                'Authorization': `Bearer ${currentToken}`
            }
        });
        
        if (response.ok) {
            showToast('资源删除成功', 'success');
            loadResources(); // 刷新列表
        } else {
            const data = await response.json();
            showToast(data.error || '删除失败', 'error');
        }
    } catch (error) {
        showToast('网络错误，请稍后重试', 'error');
    }
}

// 工具函数
function getFileIcon(fileType) {
    if (fileType.startsWith('image/')) return 'bi-file-image';
    if (fileType.startsWith('video/')) return 'bi-file-play';
    if (fileType.startsWith('audio/')) return 'bi-file-music';
    if (fileType.startsWith('text/')) return 'bi-file-text';
    if (fileType === 'application/pdf') return 'bi-file-pdf';
    if (fileType === 'application/zip') return 'bi-file-zip';
    return 'bi-file-earmark';
}

function getFileTypeName(fileType) {
    if (fileType.startsWith('image/')) return '图片';
    if (fileType.startsWith('video/')) return '视频';
    if (fileType.startsWith('audio/')) return '音频';
    if (fileType.startsWith('text/')) return '文本';
    if (fileType === 'application/pdf') return 'PDF文档';
    if (fileType === 'application/zip') return '压缩文件';
    return fileType;
}

function formatFileSize(bytes) {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

function getStatusBadgeClass(status) {
    switch (status) {
        case 'completed': return 'bg-success';
        case 'uploading': return 'bg-warning';
        case 'error': return 'bg-danger';
        default: return 'bg-secondary';
    }
}

function getStatusText(status) {
    switch (status) {
        case 'completed': return '已完成';
        case 'uploading': return '上传中';
        case 'error': return '错误';
        default: return status;
    }
}

function showToast(message, type = 'info') {
    // 简单的toast实现
    const toast = document.createElement('div');
    toast.className = `toast align-items-center text-bg-${type === 'error' ? 'danger' : type} border-0`;
    toast.setAttribute('role', 'alert');
    toast.setAttribute('aria-live', 'assertive');
    toast.setAttribute('aria-atomic', 'true');
    
    toast.innerHTML = `
        <div class="d-flex">
            <div class="toast-body">
                ${message}
            </div>
            <button type="button" class="btn-close btn-close-white me-2 m-auto" data-bs-dismiss="toast"></button>
        </div>
    `;
    
    const container = document.getElementById('toast-container') || (() => {
        const div = document.createElement('div');
        div.id = 'toast-container';
        div.className = 'toast-container position-fixed bottom-0 end-0 p-3';
        document.body.appendChild(div);
        return div;
    })();
    
    container.appendChild(toast);
    
    const bsToast = new bootstrap.Toast(toast);
    bsToast.show();
    
    toast.addEventListener('hidden.bs.toast', () => {
        toast.remove();
    });
}

// 防抖函数
function debounce(func, wait) {
    let timeout;
    return function executedFunction(...args) {
        const later = () => {
            clearTimeout(timeout);
            func(...args);
        };
        clearTimeout(timeout);
        timeout = setTimeout(later, wait);
    };
}

// 设置导航栏激活状态
function setActiveNav(navElement) {
    document.querySelectorAll('.navbar-nav .nav-link').forEach(nav => {
        nav.classList.remove('active');
    });
    navElement.classList.add('active');
}

// 全局函数
window.viewResource = viewResource;
window.deleteResource = deleteResource;
window.loadResources = loadResources;