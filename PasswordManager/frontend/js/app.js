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

function init() {
    loadEntries();
    loadLogs();
    setupEventListeners();
    updateStats();
}

function loadEntries() {
    entries = [
        { id: 1, title: 'Gmail', username: 'user1@gmail.com', password: 'encrypted1', remark: 'Personal account' },
        { id: 2, title: 'GitHub', username: 'dev1', password: 'encrypted2', remark: 'Work account' },
        { id: 3, title: 'example.com', username: 'user123', password: 'encrypted3', remark: '' },
        { id: 4, title: 'secure-site.com', username: 'admin', password: 'encrypted4', remark: 'Admin account' }
    ];
    renderEntries();
    renderRecentList();
}

function loadLogs() {
    const now = new Date();
    const formatTime = (date) => {
        return date.toLocaleString('zh-CN', {
            year: 'numeric',
            month: '2-digit',
            day: '2-digit',
            hour: '2-digit',
            minute: '2-digit'
        });
    };
    
    logs = [
        { id: 1, entry_id: 1, operation_type: 'add', operation_time: formatTime(new Date(now - 3600000)), user_id: 1 },
        { id: 2, entry_id: 2, operation_type: 'add', operation_time: formatTime(new Date(now - 7200000)), user_id: 1 },
        { id: 3, entry_id: 1, operation_type: 'update', operation_time: formatTime(new Date(now - 1800000)), user_id: 1 },
        { id: 4, entry_id: 3, operation_type: 'add', operation_time: formatTime(new Date(now - 900000)), user_id: 1 }
    ];
    renderLogs();
}

function renderEntries() {
    entriesTable.innerHTML = '';
    entries.forEach(entry => {
        const row = document.createElement('div');
        row.className = 'table-row';
        row.innerHTML = `
            <span>${entry.title}</span>
            <span>${entry.username}</span>
            <div class="table-actions">
                <button class="edit-btn" data-id="${entry.id}">编辑</button>
                <button class="delete-btn" data-id="${entry.id}">删除</button>
            </div>
        `;
        entriesTable.appendChild(row);
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
    const newEntry = {
        id: entries.length > 0 ? Math.max(...entries.map(e => e.id)) + 1 : 1,
        ...entryData
    };
    entries.push(newEntry);
    renderEntries();
    renderRecentList();
    updateStats();

    const now = new Date().toLocaleString('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit'
    });
    
    logs.push({
        id: logs.length > 0 ? Math.max(...logs.map(l => l.id)) + 1 : 1,
        entry_id: newEntry.id,
        operation_type: 'add',
        operation_time: now,
        user_id: 1
    });
    renderLogs();
}

function updateEntry(id, entryData) {
    const index = entries.findIndex(e => e.id === id);
    if (index !== -1) {
        entries[index] = { id, ...entryData };
        renderEntries();
        renderRecentList();

        const now = new Date().toLocaleString('zh-CN', {
            year: 'numeric',
            month: '2-digit',
            day: '2-digit',
            hour: '2-digit',
            minute: '2-digit'
        });
        
        logs.push({
            id: logs.length > 0 ? Math.max(...logs.map(l => l.id)) + 1 : 1,
            entry_id: id,
            operation_type: 'update',
            operation_time: now,
            user_id: 1
        });
        renderLogs();
    }
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
    const index = entries.findIndex(e => e.id === id);
    if (index !== -1) {
        const deletedEntry = entries[index];
        entries.splice(index, 1);
        renderEntries();
        renderRecentList();
        updateStats();

        const now = new Date().toLocaleString('zh-CN', {
            year: 'numeric',
            month: '2-digit',
            day: '2-digit',
            hour: '2-digit',
            minute: '2-digit'
        });
        
        logs.push({
            id: logs.length > 0 ? Math.max(...logs.map(l => l.id)) + 1 : 1,
            entry_id: id,
            operation_type: 'delete',
            operation_time: now,
            user_id: 1
        });
        renderLogs();
    }
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