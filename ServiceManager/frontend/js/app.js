const apiBaseUrl = `${window.location.protocol}//${window.location.hostname}${window.location.port ? `:${window.location.port}` : ''}/api`;

let services = [];
const servicesContainer = document.getElementById('services-container');

const addServiceBtn = document.getElementById('add-service-btn');
const addServiceModal = document.getElementById('add-service-modal');
const addServiceForm = document.getElementById('add-service-form');
const importServicesBtn = document.getElementById('import-services-btn');
const importServiceModal = document.getElementById('import-service-modal');
const jsonFileInput = document.getElementById('json-file');
const jsonInput = document.getElementById('json-input');
const submitImportBtn = document.getElementById('submit-import');

document.addEventListener('DOMContentLoaded', () => {
    fetchServices();
    setupEventListeners();
});

async function fetchServices() {
    try {
        const response = await fetch(`${apiBaseUrl}/services`);
        services = await response.json();
        renderServices();
    } catch (error) {
        console.error('获取服务列表失败:', error);
        renderEmptyState();
    }
}

function renderServices() {
    if (!services || services.length === 0) {
        renderEmptyState();
        return;
    }

    servicesContainer.innerHTML = '';
    services.forEach(service => {
        const card = document.createElement('div');
        card.className = `service-card ${service.status}`;
        card.innerHTML = `
            <div class="card-header">
                <div class="card-title">
                    <h3>${service.name}</h3>
                    <p>ID: ${service.id}</p>
                </div>
                <span class="status-badge ${service.status}">${service.status === 'running' ? '运行中' : '已停止'}</span>
            </div>
            <div class="card-body">
                <div class="info-row">
                    <div class="info-icon">🕐</div>
                    <div class="info-content">
                        <div class="info-label">启动时间</div>
                        <div class="info-value">${service.startTime?.Valid ? new Date(service.startTime.Time).toLocaleString('zh-CN') : '-'}</div>
                    </div>
                </div>
                <div class="info-row">
                    <div class="info-icon">⏱️</div>
                    <div class="info-content">
                        <div class="info-label">运行时长</div>
                        <div class="info-value">${service.uptime?.Valid ? service.uptime.String : '-'}</div>
                    </div>
                </div>
                <div class="info-row">
                    <div class="info-icon">📝</div>
                    <div class="info-content">
                        <div class="info-label">命令</div>
                        <div class="info-value">${service.command}</div>
                    </div>
                </div>
                ${service.url ? `
                <div class="info-row">
                    <div class="info-icon">🌐</div>
                    <div class="info-content">
                        <div class="info-label">URL</div>
                        <div class="info-value"><a href="${service.url}" target="_blank">${service.url}</a></div>
                    </div>
                </div>
                ` : ''}
            </div>
            <div class="card-actions">
                <button class="card-button ${service.status === 'running' ? 'stop-btn' : 'start-btn'}" data-id="${service.id}">
                    ${service.status === 'running' ? '⏹ 停止' : '▶ 启动'}
                </button>
                <button class="card-button edit-btn" data-id="${service.id}">✏️ 编辑</button>
                <button class="card-button delete-btn" data-id="${service.id}">🗑️ 删除</button>
            </div>
        `;

        card.addEventListener('click', (e) => {
            if (e.target.tagName === 'BUTTON' || e.target.closest('button')) return;
            showServiceDetails(service);
        });

        servicesContainer.appendChild(card);
    });
}

function renderEmptyState() {
    servicesContainer.innerHTML = `
        <div class="empty-state">
            <div class="empty-icon">📭</div>
            <h3>暂无服务</h3>
            <p>点击上方"添加服务"按钮创建您的第一个服务</p>
        </div>
    `;
}

function showServiceDetails(service) {
    const modal = document.getElementById('service-modal');
    const modalContent = document.getElementById('modal-content');
    const modalTitle = document.getElementById('modal-title');

    modalTitle.textContent = service.name;
    modalContent.innerHTML = `
        <div class="service-details">
            <p><strong>服务ID:</strong> <span>${service.id}</span></p>
            <p><strong>状态:</strong> <span class="status-badge ${service.status}">${service.status === 'running' ? '运行中' : '已停止'}</span></p>
            <p><strong>启动时间:</strong> <span>${service.startTime?.Valid ? new Date(service.startTime.Time).toLocaleString('zh-CN') : '-'}</span></p>
            <p><strong>运行时长:</strong> <span>${service.uptime?.Valid ? service.uptime.String : '-'}</span></p>
            <p><strong>启动命令:</strong> <span>${service.command}</span></p>
            ${service.url ? `<p><strong>访问地址:</strong> <a href="${service.url}" target="_blank">${service.url}</a></p>` : ''}
            <p><strong>创建时间:</strong> <span>${new Date(service.createdAt).toLocaleString('zh-CN')}</span></p>
            ${service.updatedAt ? `<p><strong>更新时间:</strong> <span>${new Date(service.updatedAt).toLocaleString('zh-CN')}</span></p>` : ''}
        </div>
    `;

    modal.style.display = 'flex';

    document.querySelector('#service-modal .close').addEventListener('click', () => {
        modal.style.display = 'none';
    });
}

function setupEventListeners() {
    document.querySelectorAll('.close').forEach(closeBtn => {
        closeBtn.addEventListener('click', () => {
            document.querySelectorAll('.modal').forEach(modal => {
                modal.style.display = 'none';
            });
        });
    });

    document.querySelectorAll('.modal').forEach(modal => {
        modal.addEventListener('click', (e) => {
            if (e.target === modal) {
                modal.style.display = 'none';
            }
        });
    });

    servicesContainer.addEventListener('click', (e) => {
        const button = e.target.closest('button');
        if (!button) return;

        const serviceId = button.dataset.id;

        if (button.classList.contains('start-btn')) {
            startService(serviceId);
        } else if (button.classList.contains('stop-btn')) {
            stopService(serviceId);
        } else if (button.classList.contains('delete-btn')) {
            if (confirm('确定要删除此服务吗？')) {
                deleteService(serviceId);
            }
        } else if (button.classList.contains('edit-btn')) {
            editService(serviceId);
        }
    });

    addServiceBtn.addEventListener('click', () => {
        addServiceModal.style.display = 'flex';
    });

    importServicesBtn.addEventListener('click', () => {
        importServiceModal.style.display = 'flex';
    });

    addServiceForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        const name = document.getElementById('service-name').value;
        const command = document.getElementById('service-command').value;
        const url = document.getElementById('service-url').value;

        try {
            await createService({ name, command, url });
            addServiceModal.style.display = 'none';
            addServiceForm.reset();
            fetchServices();
        } catch (error) {
            console.error('创建服务失败:', error);
            alert('创建服务失败: ' + error.message);
        }
    });

    jsonFileInput.addEventListener('change', async (e) => {
        const file = e.target.files[0];
        if (file) {
            try {
                const content = await readFileAsText(file);
                jsonInput.value = content;
            } catch (error) {
                console.error('读取文件失败:', error);
                alert('读取文件失败，请检查文件格式');
            }
        }
    });

    submitImportBtn.addEventListener('click', async () => {
        try {
            const json = jsonInput.value.trim();
            if (!json) {
                alert('请输入或上传JSON数据');
                return;
            }

            const result = await importServices(json);
            alert(`成功导入 ${result.count} 个服务`);
            importServiceModal.style.display = 'none';
            jsonFileInput.value = '';
            jsonInput.value = '';
            fetchServices();
        } catch (error) {
            alert(`导入失败: ${error.message}`);
            console.error('导入服务失败:', error);
        }
    });
}

function readFileAsText(file) {
    return new Promise((resolve, reject) => {
        const reader = new FileReader();
        reader.onload = (e) => resolve(e.target.result);
        reader.onerror = (e) => reject(new Error('文件读取失败'));
        reader.readAsText(file);
    });
}

async function importServices(json) {
    try {
        const servicesData = JSON.parse(json);
        if (!Array.isArray(servicesData)) {
            throw new Error('JSON数据必须是服务数组');
        }

        const response = await fetch(`${apiBaseUrl}/services/import`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: json
        });

        if (!response.ok) {
            const errorData = await response.json();
            throw new Error(errorData.message || '导入失败');
        }

        return await response.json();
    } catch (error) {
        console.error('导入服务失败:', error);
        throw new Error(`导入失败: ${error.message}. 请检查JSON格式是否正确`);
    }
}

async function createService(serviceData) {
    try {
        const response = await fetch(`${apiBaseUrl}/services`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(serviceData)
        });

        if (!response.ok) {
            const errorData = await response.json();
            throw new Error(errorData.message || '创建失败');
        }
    } catch (error) {
        console.error('创建服务失败:', error);
        throw error;
    }
}

async function startService(serviceId) {
    try {
        const service = services.find(s => s.id == serviceId);
        if (!service) {
            throw new Error('服务不存在');
        }

        const response = await fetch(`${apiBaseUrl}/services/start`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(service)
        });

        if (!response.ok) {
            const errorData = await response.json();
            throw new Error(errorData.message || '启动失败');
        }

        fetchServices();
    } catch (error) {
        console.error('启动服务失败:', error);
        alert('启动服务失败: ' + error.message);
    }
}

async function stopService(serviceId) {
    try {
        const service = services.find(s => s.id == serviceId);
        if (!service) {
            throw new Error('服务不存在');
        }

        const response = await fetch(`${apiBaseUrl}/services/stop`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(service)
        });

        if (!response.ok) {
            const errorData = await response.json();
            throw new Error(errorData.message || '停止失败');
        }

        fetchServices();
    } catch (error) {
        console.error('停止服务失败:', error);
        alert('停止服务失败: ' + error.message);
    }
}

async function editService(serviceId) {
    try {
        const response = await fetch(`${apiBaseUrl}/services`);
        const servicesData = await response.json();
        const service = servicesData.find(s => s.id == serviceId);

        if (!service) {
            throw new Error('服务不存在');
        }

        const editModal = document.createElement('div');
        editModal.className = 'modal';
        editModal.id = 'edit-service-modal';
        editModal.innerHTML = `
            <div class="modal-content">
                <div class="modal-header">
                    <h2>编辑服务</h2>
                    <span class="close">&times;</span>
                </div>
                <form id="edit-service-form">
                    <input type="hidden" id="edit-service-id" value="${service.id}">
                    <div class="form-group">
                        <label for="edit-service-name">服务名称</label>
                        <input type="text" id="edit-service-name" value="${service.name}" required>
                    </div>
                    <div class="form-group">
                        <label for="edit-service-status">状态</label>
                        <select id="edit-service-status">
                            <option value="running" ${service.status === 'running' ? 'selected' : ''}>运行中</option>
                            <option value="stopped" ${service.status === 'stopped' ? 'selected' : ''}>已停止</option>
                        </select>
                    </div>
                    <div class="form-group">
                        <label for="edit-service-url">URL</label>
                        <input type="text" id="edit-service-url" value="${service.url || ''}">
                    </div>
                    <div class="form-group">
                        <label for="edit-service-command">命令</label>
                        <input type="text" id="edit-service-command" value="${service.command || ''}" required>
                    </div>
                    <button type="submit" class="submit-btn">保存更改</button>
                </form>
            </div>
        `;

        document.body.appendChild(editModal);
        editModal.style.display = 'flex';

        editModal.querySelector('.close').addEventListener('click', () => {
            editModal.style.display = 'none';
            editModal.remove();
        });

        document.getElementById('edit-service-form').addEventListener('submit', async (e) => {
            e.preventDefault();

            const serviceData = {
                id: parseInt(document.getElementById('edit-service-id').value),
                name: document.getElementById('edit-service-name').value,
                status: document.getElementById('edit-service-status').value,
                url: document.getElementById('edit-service-url').value || null,
                command: document.getElementById('edit-service-command').value
            };

            try {
                await updateService(serviceData);
                editModal.style.display = 'none';
                editModal.remove();
                fetchServices();
            } catch (error) {
                console.error('编辑服务失败:', error);
                alert('编辑服务失败: ' + error.message);
            }
        });
    } catch (error) {
        console.error('获取服务详情失败:', error);
        alert('无法编辑服务: ' + error.message);
    }
}

async function updateService(serviceData) {
    try {
        const response = await fetch(`${apiBaseUrl}/services/edit`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(serviceData)
        });

        if (!response.ok) {
            const errorData = await response.json();
            throw new Error(errorData.message || '更新失败');
        }
    } catch (error) {
        console.error('更新服务失败:', error);
        throw error;
    }
}

async function deleteService(serviceId) {
    try {
        const response = await fetch(`${apiBaseUrl}/services/${serviceId}`, {
            method: 'DELETE'
        });

        if (!response.ok) {
            const errorData = await response.json();
            throw new Error(errorData.message || '删除失败');
        }

        fetchServices();
    } catch (error) {
        console.error('删除服务失败:', error);
        alert('删除服务失败: ' + error.message);
    }
}