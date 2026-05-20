let currentPath = '.';
let selectedItems = [];
let eventSource = null;
let isMobile = /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(navigator.userAgent);
let isBatchMode = false;

document.addEventListener('DOMContentLoaded', () => {
    loadFiles(currentPath);
    connectSSE();
    
    document.getElementById('uploadBtn').addEventListener('click', () => {
        document.getElementById('fileInput').click();
    });
    
    document.getElementById('fileInput').addEventListener('change', uploadFiles);
    
    document.getElementById('mkdirBtn').addEventListener('click', () => {
        document.getElementById('mkdirModal').style.display = 'block';
    });
    
    document.getElementById('pasteBtn').addEventListener('click', handlePasteClick);
    if (isMobile) {
        document.getElementById('pasteBtn').addEventListener('touchend', handlePasteClick);
    }
    
    function handlePasteClick(e) {
        if (e) {
            e.preventDefault();
            e.stopPropagation();
        }
        document.getElementById('pasteFileName').value = '';
        document.getElementById('pasteContent').value = '';
        document.getElementById('pasteModal').style.display = 'block';
    }
    
    document.getElementById('confirmPaste').addEventListener('click', confirmPaste);
    document.getElementById('cancelPaste').addEventListener('click', hidePasteModal);
    document.getElementById('pasteModalClose').addEventListener('click', hidePasteModal);
    
    document.getElementById('batchBtn').addEventListener('click', enterBatchMode);
    
    document.getElementById('deleteBtn').addEventListener('click', deleteSelectedFiles);
    
    document.getElementById('confirmMkdir').addEventListener('click', createDirectory);
    
    document.querySelector('.close').addEventListener('click', () => {
        document.getElementById('mkdirModal').style.display = 'none';
    });
    
    window.addEventListener('click', (e) => {
        const modal = document.getElementById('mkdirModal');
        if (e.target === modal) {
            modal.style.display = 'none';
        }
        
        const contextMenu = document.querySelector('.context-menu');
        if (contextMenu && e.target !== contextMenu && !contextMenu.contains(e.target)) {
            contextMenu.remove();
        }
        
        if (!e.target.closest('.file-item') && !e.target.closest('.context-menu')) {
            clearSelection();
        }
    });
    
    if (!isMobile) {
        document.addEventListener('keydown', (e) => {
            if (e.key === 'Delete' && selectedItems.length > 0) {
                e.preventDefault();
                deleteSelectedFiles();
            }
            if (e.key === 'Escape') {
                if (isBatchMode) {
                    exitBatchMode();
                } else {
                    clearSelection();
                }
            }
        });
    }
});

function loadFiles(path) {
    fetch(`/api/files?path=${encodeURIComponent(path)}`)
        .then(response => response.json())
        .then(files => {
            displayFiles(files);
            currentPath = path;
            updateBreadcrumb(path);
        })
        .catch(error => console.error('Error loading files:', error));
}

function displayFiles(files) {
    const fileList = document.getElementById('fileList');
    fileList.innerHTML = '';
    
    if (files.length === 0) {
        fileList.innerHTML = `
            <div class="empty-state">
                <div class="icon">📭</div>
                <p>空目录</p>
            </div>
        `;
        return;
    }
    
    files.forEach(file => {
        const fileItem = document.createElement('div');
        fileItem.className = 'file-item';
        fileItem.dataset.path = file.path;
        fileItem.dataset.isDir = file.isDir;
        
        fileItem.innerHTML = `
            <div class="file-icon">${file.isDir ? '📁' : getFileIcon(file.name)}</div>
            <div class="file-name">${file.name}</div>
            <div class="file-info">${file.isDir ? '文件夹' : formatSize(file.size)} · ${file.modTime}</div>
        `;
        
        if (isMobile) {
            fileItem.addEventListener('click', (e) => {
                e.preventDefault();
                if (isBatchMode) {
                    toggleSelection(file.path, fileItem);
                } else {
                    showMobileMenu(e, file, fileItem);
                }
            });
        } else {
            fileItem.addEventListener('click', (e) => {
                if (e.ctrlKey || e.metaKey) {
                    e.preventDefault();
                    toggleSelection(file.path, fileItem);
                } else if (file.isDir) {
                    loadFiles(file.path);
                } else {
                    downloadFile(file.path);
                }
            });
            
            fileItem.addEventListener('contextmenu', (e) => {
                e.preventDefault();
                if (!selectedItems.includes(file.path)) {
                    clearSelection();
                    toggleSelection(file.path, fileItem);
                }
                showContextMenu(e, file);
            });
        }
        
        
        
        fileList.appendChild(fileItem);
    });
}

function getFileIcon(filename) {
    const ext = filename.split('.').pop().toLowerCase();
    const icons = {
        'pdf': '📕',
        'doc': '📘', 'docx': '📘',
        'xls': '📗', 'xlsx': '📗',
        'ppt': '📙', 'pptx': '📙',
        'txt': '📝',
        'zip': '📦', 'rar': '📦', '7z': '📦',
        'jpg': '🖼️', 'jpeg': '🖼️', 'png': '🖼️', 'gif': '🖼️',
        'mp3': '🎵', 'wav': '🎵',
        'mp4': '🎬', 'avi': '🎬', 'mov': '🎬',
        'html': '🌐', 'css': '🎨', 'js': '⚙️', 'go': '🐹'
    };
    return icons[ext] || '📄';
}

function formatSize(bytes) {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

function updateBreadcrumb(path) {
    const breadcrumb = document.getElementById('breadcrumb');
    breadcrumb.innerHTML = '';
    
    const crumbs = path === '.' ? [] : path.split('/');
    let currentPath = '.';
    
    const homeCrumb = document.createElement('span');
    homeCrumb.className = 'crumb';
    homeCrumb.dataset.path = '.';
    homeCrumb.textContent = '🏠 根目录';
    homeCrumb.addEventListener('click', () => loadFiles('.'));
    breadcrumb.appendChild(homeCrumb);
    
    crumbs.forEach(crumb => {
        currentPath = currentPath === '.' ? crumb : `${currentPath}/${crumb}`;
        const span = document.createElement('span');
        span.className = 'crumb';
        span.dataset.path = currentPath;
        span.textContent = crumb;
        span.addEventListener('click', () => loadFiles(currentPath));
        breadcrumb.appendChild(span);
    });
}

function uploadFiles(e) {
    const files = e.target.files;
    if (files.length === 0) return;
    
    Array.from(files).forEach(file => {
        const formData = new FormData();
        formData.append('file', file);
        formData.append('path', currentPath);
        
        fetch('/api/upload', {
            method: 'POST',
            body: formData
        }).then(() => {
            loadFiles(currentPath);
        }).catch(error => console.error('Upload error:', error));
    });
    
    e.target.value = '';
}

function createDirectory() {
    const dirName = document.getElementById('dirName').value.trim();
    if (!dirName) {
        alert('请输入文件夹名称');
        return;
    }
    
    const path = currentPath === '.' ? dirName : `${currentPath}/${dirName}`;
    
    fetch('/api/mkdir', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/x-www-form-urlencoded'
        },
        body: `path=${encodeURIComponent(path)}`
    }).then(() => {
        loadFiles(currentPath);
        document.getElementById('mkdirModal').style.display = 'none';
        document.getElementById('dirName').value = '';
    }).catch(error => console.error('Create dir error:', error));
}

function downloadFile(path) {
    window.location.href = `/api/download?path=${encodeURIComponent(path)}`;
}

function deleteFile(path) {
    if (!confirm('确定要删除吗？')) return;
    
    fetch('/api/delete', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/x-www-form-urlencoded'
        },
        body: `path=${encodeURIComponent(path)}`
    }).then(() => {
        loadFiles(currentPath);
    }).catch(error => console.error('Delete error:', error));
}

function showContextMenu(e, file) {
    const existingMenu = document.querySelector('.context-menu');
    if (existingMenu) existingMenu.remove();
    
    const menu = document.createElement('div');
    menu.className = 'context-menu';
    menu.style.left = `${e.pageX}px`;
    menu.style.top = `${e.pageY}px`;
    
    const downloadBtn = document.createElement('button');
    downloadBtn.innerHTML = '⬇️ 下载';
    downloadBtn.addEventListener('click', () => {
        downloadFile(file.path);
        menu.remove();
    });
    
    const deleteBtn = document.createElement('button');
    deleteBtn.innerHTML = '🗑️ 删除';
    deleteBtn.addEventListener('click', () => {
        deleteFile(file.path);
        menu.remove();
    });
    
    menu.appendChild(downloadBtn);
    menu.appendChild(deleteBtn);
    
    document.body.appendChild(menu);
}

function showMobileMenu(e, file, fileItem) {
    const existingMenu = document.querySelector('.mobile-menu');
    if (existingMenu) existingMenu.remove();
    
    const menu = document.createElement('div');
    menu.className = 'mobile-menu';
    
    const menuContent = document.createElement('div');
    menuContent.className = 'mobile-menu-content';
    
    const iconDiv = document.createElement('div');
    iconDiv.className = 'mobile-menu-icon';
    iconDiv.innerHTML = file.isDir ? '📁' : getFileIcon(file.name);
    menuContent.appendChild(iconDiv);
    
    const nameDiv = document.createElement('div');
    nameDiv.className = 'mobile-menu-name';
    nameDiv.textContent = file.name;
    menuContent.appendChild(nameDiv);
    
    const infoDiv = document.createElement('div');
    infoDiv.className = 'mobile-menu-info';
    infoDiv.textContent = file.isDir ? '文件夹' : formatSize(file.size);
    menuContent.appendChild(infoDiv);
    
    const actionsDiv = document.createElement('div');
    actionsDiv.className = 'mobile-menu-actions';
    
    if (file.isDir) {
        const enterBtn = document.createElement('button');
        enterBtn.className = 'mobile-menu-btn primary';
        enterBtn.innerHTML = '📂 进入';
        enterBtn.addEventListener('click', () => {
            loadFiles(file.path);
            menu.remove();
        });
        actionsDiv.appendChild(enterBtn);
    } else {
        const downloadBtn = document.createElement('button');
        downloadBtn.className = 'mobile-menu-btn primary';
        downloadBtn.innerHTML = '⬇️ 下载';
        downloadBtn.addEventListener('click', () => {
            downloadFile(file.path);
            menu.remove();
        });
        actionsDiv.appendChild(downloadBtn);
    }
    
    const deleteBtn = document.createElement('button');
    deleteBtn.className = 'mobile-menu-btn danger';
    deleteBtn.innerHTML = '🗑️ 删除';
    deleteBtn.addEventListener('click', () => {
        deleteFile(file.path);
        menu.remove();
    });
    actionsDiv.appendChild(deleteBtn);
    
    menuContent.appendChild(actionsDiv);
    menu.appendChild(menuContent);
    
    menu.addEventListener('click', (e) => {
        if (e.target === menu) {
            menu.remove();
        }
    });
    
    document.body.appendChild(menu);
}

function enterBatchMode() {
    isBatchMode = true;
    clearSelection();
    
    const batchBar = document.createElement('div');
    batchBar.id = 'batchBar';
    batchBar.className = 'batch-bar';
    
    const cancelBtn = document.createElement('button');
    cancelBtn.className = 'batch-btn cancel';
    cancelBtn.innerHTML = '❌ 取消';
    cancelBtn.addEventListener('click', exitBatchMode);
    
    const selectAllBtn = document.createElement('button');
    selectAllBtn.className = 'batch-btn';
    selectAllBtn.innerHTML = '☑️ 全选';
    selectAllBtn.addEventListener('click', selectAllFiles);
    
    const deleteBtn = document.createElement('button');
    deleteBtn.className = 'batch-btn danger';
    deleteBtn.innerHTML = '🗑️ 删除';
    deleteBtn.addEventListener('click', deleteSelectedFiles);
    
    const countSpan = document.createElement('span');
    countSpan.className = 'batch-count';
    countSpan.textContent = '已选择 0 个';
    
    batchBar.appendChild(cancelBtn);
    batchBar.appendChild(selectAllBtn);
    batchBar.appendChild(countSpan);
    batchBar.appendChild(deleteBtn);
    
    document.body.appendChild(batchBar);
    
    document.querySelector('header').style.display = 'none';
}

function exitBatchMode() {
    isBatchMode = false;
    clearSelection();
    
    const batchBar = document.getElementById('batchBar');
    if (batchBar) batchBar.remove();
    
    document.querySelector('header').style.display = 'flex';
}

function selectAllFiles() {
    const fileItems = document.querySelectorAll('.file-item');
    selectedItems = [];
    
    fileItems.forEach(item => {
        const path = item.dataset.path;
        selectedItems.push(path);
        item.classList.add('selected');
    });
    
    updateBatchCount();
}

function updateBatchCount() {
    const countSpan = document.querySelector('.batch-count');
    if (countSpan) {
        countSpan.textContent = `已选择 ${selectedItems.length} 个`;
    }
}

function updateDeleteButton() {
    const btn = document.getElementById('deleteBtn');
    if (btn && !isMobile) {
        if (selectedItems.length > 0) {
            btn.style.display = 'block';
            btn.textContent = `🗑️ 删除选中 (${selectedItems.length})`;
        } else {
            btn.style.display = 'none';
        }
    }
    if (isMobile && isBatchMode) {
        updateBatchCount();
    }
}

function toggleSelection(path, element) {
    const index = selectedItems.indexOf(path);
    if (index > -1) {
        selectedItems.splice(index, 1);
        element.classList.remove('selected');
    } else {
        selectedItems.push(path);
        element.classList.add('selected');
    }
    updateDeleteButton();
}

function clearSelection() {
    selectedItems.forEach(path => {
        const element = document.querySelector(`[data-path="${path}"]`);
        if (element) element.classList.remove('selected');
    });
    selectedItems = [];
    updateDeleteButton();
}

function updateDeleteButton() {
    const btn = document.getElementById('deleteBtn');
    if (selectedItems.length > 0) {
        btn.style.display = 'block';
        btn.textContent = `🗑️ 删除选中 (${selectedItems.length})`;
    } else {
        btn.style.display = 'none';
    }
}

function deleteSelectedFiles() {
    if (selectedItems.length === 0) return;
    
    const count = selectedItems.length;
    if (!confirm(`确定要删除 ${count} 个文件/文件夹吗？`)) return;
    
    const deletePromises = selectedItems.map(path => {
        return fetch('/api/delete', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/x-www-form-urlencoded'
            },
            body: `path=${encodeURIComponent(path)}`
        });
    });
    
    Promise.all(deletePromises).then(() => {
        loadFiles(currentPath);
        clearSelection();
    }).catch(error => console.error('Delete error:', error));
}

function hidePasteModal() {
    document.getElementById('pasteModal').style.display = 'none';
}

function confirmPaste() {
    const content = document.getElementById('pasteContent').value;
    if (!content || content.trim() === '') {
        alert('请输入内容');
        return;
    }
    
    let fileName = document.getElementById('pasteFileName').value.trim();
    if (!fileName) {
        const timestamp = new Date().toISOString().replace(/[:.]/g, '-');
        fileName = `粘贴_${timestamp}.txt`;
    } else if (!fileName.endsWith('.txt')) {
        fileName += '.txt';
    }
    
    fetch('/api/create-file', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/x-www-form-urlencoded'
        },
        body: `path=${encodeURIComponent(currentPath)}&filename=${encodeURIComponent(fileName)}&content=${encodeURIComponent(content)}`
    }).then(response => {
        if (response.ok) {
            loadFiles(currentPath);
            hidePasteModal();
        } else {
            alert('创建文件失败');
        }
    }).catch(error => {
        console.error('Create file error:', error);
        alert('创建文件失败');
    });
}

function connectSSE() {
    if (eventSource) {
        eventSource.close();
    }
    
    eventSource = new EventSource('/api/events');
    
    eventSource.addEventListener('change', (event) => {
        loadFiles(currentPath);
    });
    
    eventSource.addEventListener('ping', () => {
    });
    
    eventSource.onerror = (error) => {
        console.error('SSE connection error:', error);
        setTimeout(connectSSE, 5000);
    };
}