const entryForm = document.getElementById('entryForm');
const entriesTable = document.getElementById('entriesTable');
const logsContainer = document.getElementById('logsContainer');
const recentList = document.getElementById('recentList');
const totalEntries = document.getElementById('totalEntries');
const todayAdded = document.getElementById('todayAdded');
const navBtns = document.querySelectorAll('.nav-btn');

const titleInput = document.getElementById('title');
const usernameInput = document.getElementById('username');
const passwordInput = document.getElementById('password');
const remarkInput = document.getElementById('remark');

let entries = [];
let logs = [];
let currentEntryId = null;

const API_BASE = `${window.location.origin}/api`;

function init() {
    loadEntries();
    loadLogs();
    setupEventListeners();
    updateStats();
}

function loadEntries() {
    fetch(`${API_BASE}/password_entries`)
        .then(response => response.json())
        .then(data => {
            entries = data;
            renderEntries();
            renderRecentList();
            updateStats();
        })
        .catch(error => {
            console.error('Failed to load entries:', error);
            entries = [];
            renderEntries();
            renderRecentList();
            updateStats();
        });
}

function loadLogs() {
    fetch(`${API_BASE}/operation_logs`)
        .then(response => response.json())
        .then(data => {
            logs = data.map(log => ({
                ...log,
                operation_time: formatDateTime(log.operation_time)
            }));
            renderLogs();
            updateStats();
        })
        .catch(error => {
            console.error('Failed to load logs:', error);
            logs = [];
            renderLogs();
            updateStats();
        });
}

function formatDateTime(dateString) {
    const date = new Date(dateString);
    return date.toLocaleString('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit'
    });
}

function renderEntries() {
    entriesTable.innerHTML = '';
    entries.forEach(entry => {
        const row = document.createElement('div');
        row.className = 'table-row';
        row.innerHTML = `
            <span class="entry-title">${entry.title}</span>
            <span>${entry.username}</span>
            <div class="table-actions">
                <button class="edit-btn" data-id="${entry.id}">编辑</button>
                <button class="delete-btn" data-id="${entry.id}">删除</button>
            </div>
        `;
        entriesTable.appendChild(row);
    });

    document.querySelectorAll('.table-row').forEach(row => {
        row.addEventListener('click', (e) => {
            if (!e.target.closest('.table-actions')) {
                const id = parseInt(row.querySelector('.edit-btn').dataset.id);
                const entry = entries.find(e => e.id === id);
                if (entry) {
                    showEntryModal(entry);
                }
            }
        });
    });

    document.querySelectorAll('.edit-btn').forEach(btn => {
        btn.addEventListener('click', handleEdit);
    });
    document.querySelectorAll('.delete-btn').forEach(btn => {
        btn.addEventListener('click', handleDelete);
    });
}

function renderRecentList() {
    recentList.innerHTML = '';
    const recent = [...entries].reverse().slice(0, 5);
    recent.forEach(entry => {
        const item = document.createElement('div');
        item.className = 'recent-item';
        item.textContent = `${entry.title} - ${entry.username}`;
        item.addEventListener('click', () => showEntryModal(entry));
        recentList.appendChild(item);
    });
}

function renderLogs() {
    logsContainer.innerHTML = '';
    const sortedLogs = [...logs].reverse();
    sortedLogs.forEach(log => {
        const entry = entries.find(e => e.id === log.entry_id) || { title: 'Deleted Entry' };
        const row = document.createElement('div');
        row.className = 'log-item';
        const typeText = {
            'add': '添加',
            'update': '更新',
            'delete': '删除'
        };
        row.innerHTML = `
            <span class="log-time">${log.operation_time}</span>
            <span class="log-type ${log.operation_type}">${typeText[log.operation_type]}</span>
            <span>${entry.title}</span>
        `;
        logsContainer.appendChild(row);
    });
}

function showEntryModal(entry) {
    const modal = document.createElement('div');
    modal.className = 'modal-overlay';
    modal.innerHTML = `
        <div class="modal-content">
            <div class="modal-header">
                <h3>${entry.title}</h3>
                <button class="modal-close">&times;</button>
            </div>
            <div class="modal-body">
                <div class="modal-item">
                    <label>网站/标题</label>
                    <span>${entry.title}</span>
                </div>
                <div class="modal-item">
                    <label>用户名</label>
                    <span>${entry.username}</span>
                </div>
                <div class="modal-item">
                    <label>密码</label>
                    <span class="password-display">${entry.password}</span>
                </div>
                <div class="modal-item">
                    <label>备注</label>
                    <span>${entry.remark || '无'}</span>
                </div>
            </div>
            <div class="modal-footer">
                <button class="modal-close-btn">关闭</button>
            </div>
        </div>
    `;
    
    document.body.appendChild(modal);
    
    const closeModal = () => {
        document.body.removeChild(modal);
    };
    
    modal.querySelector('.modal-close').addEventListener('click', closeModal);
    modal.querySelector('.modal-close-btn').addEventListener('click', closeModal);
    modal.addEventListener('click', (e) => {
        if (e.target === modal) {
            closeModal();
        }
    });
}

function handleSubmit(e) {
    e.preventDefault();
    
    const entryData = {
        title: titleInput.value,
        username: usernameInput.value,
        password: passwordInput.value,
        remark: remarkInput.value
    };

    if (currentEntryId) {
        updateEntry(currentEntryId, entryData);
    } else {
        createEntry(entryData);
    }

    resetForm();
}

function createEntry(entryData) {
    fetch(`${API_BASE}/password_entries`, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(entryData)
    })
    .then(response => response.json())
    .then(newEntry => {
        entries.push(newEntry);
        renderEntries();
        renderRecentList();
        loadLogs();
    })
    .catch(error => {
        console.error('Failed to create entry:', error);
    });
}

function updateEntry(id, entryData) {
    fetch(`${API_BASE}/password_entries/${id}`, {
        method: 'PUT',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify({ ...entryData, id })
    })
    .then(response => response.json())
    .then(updatedEntry => {
        const index = entries.findIndex(e => e.id === id);
        if (index !== -1) {
            entries[index] = updatedEntry;
            renderEntries();
            renderRecentList();
            loadLogs();
        }
    })
    .catch(error => {
        console.error('Failed to update entry:', error);
    });
}

function handleEdit(e) {
    const id = parseInt(e.target.dataset.id);
    const entry = entries.find(e => e.id === id);
    if (entry) {
        currentEntryId = id;
        titleInput.value = entry.title;
        usernameInput.value = entry.username;
        passwordInput.value = entry.password;
        remarkInput.value = entry.remark;
    }
}

function handleDelete(e) {
    const id = parseInt(e.target.dataset.id);
    if (confirm('确定要删除这条记录吗？')) {
        deleteEntry(id);
    }
}

function deleteEntry(id) {
    fetch(`${API_BASE}/password_entries/${id}`, {
        method: 'DELETE'
    })
    .then(response => {
        if (response.ok) {
            const index = entries.findIndex(e => e.id === id);
            if (index !== -1) {
                entries.splice(index, 1);
                renderEntries();
                renderRecentList();
                loadLogs();
            }
        }
    })
    .catch(error => {
        console.error('Failed to delete entry:', error);
    });
}

function resetForm() {
    currentEntryId = null;
    entryForm.reset();
}

function updateStats() {
    totalEntries.textContent = entries.length;
    todayAdded.textContent = logs.filter(log => log.operation_type === 'add').length;
}

function switchTab(tabName) {
    document.querySelectorAll('.tab-content').forEach(tab => {
        tab.style.display = 'none';
    });
    document.getElementById(`${tabName}-tab`).style.display = 'block';
    
    navBtns.forEach(btn => btn.classList.remove('active'));
    document.querySelector(`[data-tab="${tabName}"]`).classList.add('active');
}

function setupEventListeners() {
    entryForm.addEventListener('submit', handleSubmit);
    
    navBtns.forEach(btn => {
        btn.addEventListener('click', () => {
            switchTab(btn.dataset.tab);
        });
    });
}

document.addEventListener('DOMContentLoaded', init);