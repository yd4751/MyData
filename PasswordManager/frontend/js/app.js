// DOM Elements
const entryForm = document.getElementById('entryForm');
const entriesTable = document.getElementById('entriesTable').querySelector('tbody');
const logsTable = document.getElementById('logsTable').querySelector('tbody');

// Form inputs
const titleInput = document.getElementById('title');
const usernameInput = document.getElementById('username');
const passwordInput = document.getElementById('password');
const remarkInput = document.getElementById('remark');

// Sample data - will be replaced with API calls
let entries = [];
let logs = [];
let currentEntryId = null;

// Initialize the application
function init() {
    loadEntries();
    loadLogs();
    setupEventListeners();
}

// Load password entries
function loadEntries() {
    // TODO: Replace with API call to /api/password_entries
    entries = [
        {id: 1, title: 'Gmail', username: 'user1@gmail.com', password: 'encrypted1', remark: 'Personal account'},
        {id: 2, title: 'GitHub', username: 'dev1', password: 'encrypted2', remark: 'Work account'}
    ];
    renderEntries();
}

// Load operation logs
function loadLogs() {
    // TODO: Replace with API call to /api/operation_logs
    logs = [
        {id: 1, entry_id: 1, operation_type: 'add', operation_time: '2026-05-12 10:00', user_id: 1},
        {id: 2, entry_id: 2, operation_type: 'add', operation_time: '2026-05-12 10:30', user_id: 1}
    ];
    renderLogs();
}

// Render password entries to the table
function renderEntries() {
    entriesTable.innerHTML = '';
    entries.forEach(entry => {
        const row = document.createElement('tr');
        row.innerHTML = `
            <td>${entry.title}</td>
            <td>${entry.username}</td>
            <td>
                <button class="edit-btn" data-id="${entry.id}">Edit</button>
                <button class="delete-btn" data-id="${entry.id}">Delete</button>
            </td>
        `;
        entriesTable.appendChild(row);
    });

    // Add event listeners to buttons
    document.querySelectorAll('.edit-btn').forEach(btn => {
        btn.addEventListener('click', handleEdit);
    });
    document.querySelectorAll('.delete-btn').forEach(btn => {
        btn.addEventListener('click', handleDelete);
    });
}

// Render operation logs to the table
function renderLogs() {
    logsTable.innerHTML = '';
    logs.forEach(log => {
        const entry = entries.find(e => e.id === log.entry_id) || {title: 'Deleted Entry'};
        const row = document.createElement('tr');
        row.innerHTML = `
            <td>${log.operation_time}</td>
            <td>${log.operation_type}</td>
            <td>${entry.title}</td>
        `;
        logsTable.appendChild(row);
    });
}

// Handle form submission
function handleSubmit(e) {
    e.preventDefault();
    
    const entryData = {
        title: titleInput.value,
        username: usernameInput.value,
        password: passwordInput.value,
        remark: remarkInput.value
    };

    if (currentEntryId) {
        // Update existing entry
        updateEntry(currentEntryId, entryData);
    } else {
        // Create new entry
        createEntry(entryData);
    }

    resetForm();
}

// Create new password entry
function createEntry(entryData) {
    // TODO: Replace with API POST to /api/password_entries
    const newEntry = {
        id: entries.length > 0 ? Math.max(...entries.map(e => e.id)) + 1 : 1,
        ...entryData
    };
    entries.push(newEntry);
    renderEntries();

    // Simulate log creation
    logs.push({
        id: logs.length > 0 ? Math.max(...logs.map(l => l.id)) + 1 : 1,
        entry_id: newEntry.id,
        operation_type: 'add',
        operation_time: new Date().toISOString(),
        user_id: 1 // TODO: Replace with actual user ID
    });
    renderLogs();
}

// Update existing password entry
function updateEntry(id, entryData) {
    // TODO: Replace with API PUT to /api/password_entries/{id}
    const index = entries.findIndex(e => e.id === id);
    if (index !== -1) {
        entries[index] = {id, ...entryData};
        renderEntries();

        // Simulate log creation
        logs.push({
            id: logs.length > 0 ? Math.max(...logs.map(l => l.id)) + 1 : 1,
            entry_id: id,
            operation_type: 'update',
            operation_time: new Date().toISOString(),
            user_id: 1 // TODO: Replace with actual user ID
        });
        renderLogs();
    }
}

// Handle edit button click
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

// Handle delete button click
function handleDelete(e) {
    const id = parseInt(e.target.dataset.id);
    if (confirm('Are you sure you want to delete this entry?')) {
        deleteEntry(id);
    }
}

// Delete password entry
function deleteEntry(id) {
    // TODO: Replace with API DELETE to /api/password_entries/{id}
    const index = entries.findIndex(e => e.id === id);
    if (index !== -1) {
        entries.splice(index, 1);
        renderEntries();

        // Simulate log creation
        logs.push({
            id: logs.length > 0 ? Math.max(...logs.map(l => l.id)) + 1 : 1,
            entry_id: id,
            operation_type: 'delete',
            operation_time: new Date().toISOString(),
            user_id: 1 // TODO: Replace with actual user ID
        });
        renderLogs();
    }
}

// Reset form
function resetForm() {
    currentEntryId = null;
    entryForm.reset();
}

// Set up event listeners
function setupEventListeners() {
    entryForm.addEventListener('submit', handleSubmit);
}

// Initialize the app when DOM is loaded
document.addEventListener('DOMContentLoaded', init);
