const apiBaseUrl = `${window.location.protocol}//${window.location.hostname}${window.location.port ? `:${window.location.port}` : ''}/api`;

let services = [];
const servicesContainer = document.getElementById('services-container');

// UI元素
const addServiceBtn = document.getElementById('add-service-btn');
const addServiceModal = document.getElementById('add-service-modal');
const addServiceForm = document.getElementById('add-service-form');
const importServicesBtn = document.getElementById('import-services-btn');
const importServiceModal = document.getElementById('import-service-modal');
const jsonFileInput = document.getElementById('json-file');
const jsonInput = document.getElementById('json-input');
const submitImportBtn = document.getElementById('submit-import');

// 初始化应用
document.addEventListener('DOMContentLoaded', () => {
    fetchServices();
    setupEventListeners();
});

// 获取服务列表
async function fetchServices() {
    try {
        const response = await fetch(`${apiBaseUrl}/services`);
        services = await response.json();
        renderServices();
    } catch (error) {
        console.error('获取服务列表失败:', error);
    }
}

// 渲染服务卡片
function renderServices() {
    servicesContainer.innerHTML = '';
    services.forEach(service => {
        const card = document.createElement('div');
        card.className = `service-card ${service.status}`;
        card.innerHTML = `
            <div class="card-header">
                <h3>${service.name}</h3>
                <span class="status-badge ${service.status}">${service.status === 'running' ? '运行中' : '已停止'}</span>
            </div>
            <div class="card-body">
                <p><strong>启动时间:</strong> ${service.startTime?.Valid ? new Date(service.startTime.Time).toLocaleString('zh-CN') : '-'}</p>
                <p><strong>运行时长:</strong> ${service.uptime?.Valid ? service.uptime.String : '-'}</p>
                <p><strong>命令:</strong> ${service.command}</p>
                ${service.url ? `<p><strong>URL:</strong> <a href="${service.url}" target="_blank">${service.url}</a></p>` : ''}
            </div>
            <div class="actions">
                <button class="button ${service.status === 'running' ? 'stop-btn' : 'start-btn'}" data-id="${service.id}">
                    ${service.status === 'running' ? '停止' : '启动'}
                </button>
                <button class="button delete-btn" data-id="${service.id}">删除</button>
            </div>
        `;
        servicesContainer.appendChild(card);
    });
}

// 设置事件监听器
function setupEventListeners() {
    // 关闭模态框
    document.querySelectorAll('.close').forEach(closeBtn => {
        closeBtn.addEventListener('click', () => {
            document.querySelectorAll('.modal').forEach(modal => {
                modal.style.display = 'none';
            });
        });
    });

    // 点击模态框外部关闭
    document.querySelectorAll('.modal').forEach(modal => {
        modal.addEventListener('click', (e) => {
            if (e.target === modal) {
                modal.style.display = 'none';
            }
        });
    });

    // 委托处理按钮点击
    servicesContainer.addEventListener('click', (e) => {
        if (e.target.classList.contains('start-btn')) {
            startService(e.target.dataset.id);
        } else if (e.target.classList.contains('stop-btn')) {
            stopService(e.target.dataset.id);
        } else if (e.target.classList.contains('delete-btn')) {
            if (confirm('确定要删除此服务吗？')) {
                deleteService(e.target.dataset.id);
            }
        }
    });

    // 添加服务按钮
    addServiceBtn.addEventListener('click', () => {
        addServiceModal.style.display = 'block';
    });

    // 导入服务按钮
    importServicesBtn.addEventListener('click', () => {
        importServiceModal.style.display = 'block';
    });

    // 表单提交
    addServiceForm.addEventListener('submit', async (e) => {
        e.preventDefault();
        const name = document.getElementById('service-name').value;
        const command = document.getElementById('service-command').value;
        const url = document.getElementById('service-url').value;

        try {
            await createService({ name, command, url });
            addServiceModal.style.display = 'none';
            addServiceForm.reset();
            fetchServices(); // 刷新服务列表
        } catch (error) {
            console.error('创建服务失败:', error);
        }
    });

    // 文件上传处理
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

    // 导入提交
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
            fetchServices(); // 刷新服务列表
        } catch (error) {
            alert(`导入失败: ${error.message}`);
            console.error('导入服务失败:', error);
        }
    });
}

// 读取文件为文本
function readFileAsText(file) {
    return new Promise((resolve, reject) => {
        const reader = new FileReader();
        reader.onload = (e) => resolve(e.target.result);
        reader.onerror = (e) => reject(new Error('文件读取失败'));
        reader.readAsText(file);
    });
}

// 导入服务
async function importServices(json) {
    try {
        const services = JSON.parse(json);
        if (!Array.isArray(services)) {
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
        throw error;
    }
}

// 创建服务
async function createService(serviceData) {
    try {
        await fetch(`${apiBaseUrl}/services`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(serviceData)
        });
    } catch (error) {
        console.error('创建服务失败:', error);
        throw error;
    }
}

// 启动服务
async function startService(serviceId) {
    try {
        await fetch(`${apiBaseUrl}/services/${serviceId}/start`, {
            method: 'POST'
        });
        fetchServices(); // 刷新服务列表
    } catch (error) {
        console.error('启动服务失败:', error);
    }
}

// 停止服务
async function stopService(serviceId) {
    try {
        await fetch(`${apiBaseUrl}/services/${serviceId}/stop`, {
            method: 'POST'
        });
        fetchServices(); // 刷新服务列表
    } catch (error) {
        console.error('停止服务失败:', error);
    }
}

// 删除服务
async function deleteService(serviceId) {
    try {
        await fetch(`${apiBaseUrl}/services/${serviceId}`, {
            method: 'DELETE'
        });
        fetchServices(); // 刷新服务列表
    } catch (error) {
        console.error('删除服务失败:', error);
    }
}
