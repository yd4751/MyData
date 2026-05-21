let historyChart = null;
let loadChart = null;

document.addEventListener('DOMContentLoaded', () => {
    initTabs();
    initHistoryControls();
    initProcessSearch();
    initModal();
    initLoadChart();
    fetchServerStatus();
    fetchHardwareInfo();
    fetchProcesses();
    loadHistoryChart(168);
    
    setInterval(fetchServerStatus, 2000);
});

function initTabs() {
    const navButtons = document.querySelectorAll('.nav-item');
    navButtons.forEach(btn => {
        btn.addEventListener('click', () => {
            const tabId = btn.dataset.tab;
            document.querySelectorAll('.nav-item').forEach(b => b.classList.remove('active'));
            document.querySelectorAll('.tab-content').forEach(t => t.classList.remove('active'));
            btn.classList.add('active');
            document.getElementById(tabId).classList.add('active');
            
            if (tabId === 'hardware') {
                fetchHardwareInfo();
            } else if (tabId === 'processes') {
                fetchProcesses();
            } else if (tabId === 'history') {
                loadHistoryChart(getActiveRange());
            }
        });
    });
}

function initHistoryControls() {
    const rangeButtons = document.querySelectorAll('.range-btn');
    rangeButtons.forEach(btn => {
        btn.addEventListener('click', () => {
            document.querySelectorAll('.range-btn').forEach(b => b.classList.remove('active'));
            btn.classList.add('active');
            loadHistoryChart(parseInt(btn.dataset.range));
        });
    });

    const checkboxes = document.querySelectorAll('input[name="metric"]');
    checkboxes.forEach(cb => {
        cb.addEventListener('change', () => {
            loadHistoryChart(getActiveRange());
        });
    });
}

function initProcessSearch() {
    const searchBtn = document.getElementById('process-search-btn');
    const refreshBtn = document.getElementById('refresh-processes');
    const searchInput = document.getElementById('process-search-input');

    searchBtn.addEventListener('click', () => {
        const query = searchInput.value.trim();
        if (query) {
            searchProcesses(query);
        } else {
            fetchProcesses();
        }
    });

    searchInput.addEventListener('keyup', (e) => {
        if (e.key === 'Enter') {
            searchBtn.click();
        }
    });

    refreshBtn.addEventListener('click', () => {
        searchInput.value = '';
        fetchProcesses();
    });
}

function initModal() {
    const modal = document.getElementById('process-detail-modal');
    const closeBtn = document.querySelector('.close');

    closeBtn.addEventListener('click', () => {
        modal.style.display = 'none';
    });

    window.addEventListener('click', (e) => {
        if (e.target === modal) {
            modal.style.display = 'none';
        }
    });
}

function initLoadChart() {
    const ctx = document.getElementById('loadChart').getContext('2d');
    const labels = Array.from({length: 10}, (_, i) => String(i + 1));
    const data = Array(10).fill(0);

    loadChart = new Chart(ctx, {
        type: 'line',
        data: {
            labels: labels,
            datasets: [{
                label: '负载',
                data: data,
                borderColor: '#22d3ee',
                backgroundColor: 'rgba(34, 211, 238, 0.1)',
                fill: true,
                tension: 0.4,
                pointRadius: 0,
                borderWidth: 3
            }]
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            plugins: {
                legend: {
                    display: false
                },
                tooltip: {
                    backgroundColor: 'rgba(0, 0, 0, 0.8)',
                    titleColor: '#fff',
                    bodyColor: '#fff'
                }
            },
            scales: {
                x: {
                    display: false
                },
                y: {
                    display: false,
                    min: 0,
                    max: 100
                }
            }
        }
    });
}

function updateLoadChart(newValue) {
    if (loadChart) {
        const data = loadChart.data.datasets[0].data;
        data.shift();
        data.push(newValue);
        loadChart.update();
    }
}

function getActiveRange() {
    const activeBtn = document.querySelector('.range-btn.active');
    return parseInt(activeBtn.dataset.range);
}

function getSelectedMetrics() {
    const metrics = [];
    document.querySelectorAll('input[name="metric"]:checked').forEach(cb => {
        metrics.push(cb.value);
    });
    return metrics;
}

async function fetchServerStatus() {
    try {
        const response = await fetch('/api/status');
        const data = await response.json();
        updateStatusUI(data);
    } catch (error) {
        console.error('Error fetching server status:', error);
    }
}

async function fetchHardwareInfo() {
    try {
        const response = await fetch('/api/hardware');
        const data = await response.json();
        updateHardwareUI(data);
    } catch (error) {
        console.error('Error fetching hardware info:', error);
    }
}

async function fetchProcesses() {
    try {
        const response = await fetch('/api/processes?limit=10');
        const data = await response.json();
        updateProcessesUI(data);
    } catch (error) {
        console.error('Error fetching processes:', error);
    }
}

async function searchProcesses(query) {
    try {
        const response = await fetch(`/api/processes?q=${encodeURIComponent(query)}`);
        const data = await response.json();
        updateProcessesUI(data);
    } catch (error) {
        console.error('Error searching processes:', error);
    }
}

async function fetchProcessDetail(pid) {
    try {
        const response = await fetch(`/api/process?pid=${pid}`);
        const data = await response.json();
        showProcessDetail(data);
    } catch (error) {
        console.error('Error fetching process detail:', error);
    }
}

async function loadHistoryChart(hours) {
    try {
        const response = await fetch(`/api/history?hours=${hours}`);
        const data = await response.json();
        renderHistoryChart(data, hours);
    } catch (error) {
        console.error('Error loading history:', error);
    }
}

function updateStatusUI(data) {
    const cpuUsage = data.cpu_usage || 0;
    document.getElementById('cpu-usage').textContent = cpuUsage.toFixed(1);
    document.getElementById('cpu-resource').style.width = `${cpuUsage}%`;
    document.getElementById('cpu-value').textContent = cpuUsage.toFixed(1) + '%';
    
    const memoryUsage = data.memory_usage || 0;
    document.getElementById('memory-usage').textContent = memoryUsage.toFixed(1);
    document.getElementById('memory-resource').style.width = `${memoryUsage}%`;
    document.getElementById('memory-value').textContent = memoryUsage.toFixed(1) + '%';
    
    const diskUsage = data.disk_usage || 0;
    document.getElementById('disk-usage').textContent = diskUsage.toFixed(1);
    document.getElementById('disk-resource').style.width = `${diskUsage}%`;
    document.getElementById('disk-value').textContent = diskUsage.toFixed(1) + '%';
    
    const networkRate = (data.network_in + data.network_out) / 1024 || 0;
    document.getElementById('network-rate').textContent = networkRate.toFixed(1);
    
    updateLoadChart(cpuUsage);
}

function updateHardwareUI(data) {
    let cpuHTML = '';
    data.cpu_info.forEach((cpu, index) => {
        cpuHTML += `
            <div class="hardware-item">
                <p><span>CPU ${index + 1}:</span>${cpu.model_name}</p>
                <p><span>厂商:</span>${cpu.vendor_id}</p>
                <p><span>主频:</span>${cpu.mhz} MHz</p>
                <p><span>缓存:</span>${cpu.cache_size} KB</p>
                <p><span>核心数:</span>${cpu.cores}</p>
            </div>
        `;
    });
    document.getElementById('cpu-info').innerHTML = cpuHTML;
    
    const memory = data.memory_info;
    document.getElementById('memory-info').innerHTML = `
        <p><span>总内存:</span>${formatBytes(memory.total)}</p>
        <p><span>可用内存:</span>${formatBytes(memory.available)}</p>
        <p><span>已用内存:</span>${formatBytes(memory.used)} (${memory.used_percent.toFixed(1)}%)</p>
        <p><span>空闲内存:</span>${formatBytes(memory.free)}</p>
    `;
    
    let diskHTML = '';
    data.disk_info.forEach(disk => {
        diskHTML += `
            <div class="hardware-item">
                <p><span>设备:</span>${disk.device}</p>
                <p><span>挂载点:</span>${disk.mountpoint}</p>
                <p><span>文件系统:</span>${disk.fstype}</p>
                <p><span>总容量:</span>${formatBytes(disk.total)}</p>
                <p><span>已用:</span>${formatBytes(disk.used)} (${disk.used_percent.toFixed(1)}%)</p>
            </div>
        `;
    });
    document.getElementById('disk-info').innerHTML = diskHTML;
    
    let networkHTML = '';
    data.network_info.forEach(net => {
        networkHTML += `
            <div class="hardware-item">
                <p><span>接口名称:</span>${net.name}</p>
                <p><span>MAC地址:</span>${net.hardware_addr}</p>
                <p><span>IP地址:</span>${net.addrs.join(', ')}</p>
            </div>
        `;
    });
    document.getElementById('network-info').innerHTML = networkHTML;
    
    const host = data.host_info;
    document.getElementById('host-info').innerHTML = `
        <p><span>主机名:</span>${host.hostname}</p>
        <p><span>操作系统:</span>${host.os}</p>
        <p><span>平台:</span>${host.platform} ${host.platform_version}</p>
        <p><span>内核版本:</span>${host.kernel_version}</p>
        <p><span>架构:</span>${host.kernel_arch}</p>
        <p><span>运行时间:</span>${formatUptime(host.uptime)}</p>
        <p><span>进程数:</span>${host.procs}</p>
    `;
}

function updateProcessesUI(processes) {
    let html = '';
    processes.forEach(p => {
        html += `
            <tr>
                <td>${p.pid}</td>
                <td>${p.name || '-'}</td>
                <td>${p.cpu_percent.toFixed(2)}%</td>
                <td>${p.memory_percent.toFixed(2)}%</td>
                <td>${formatBytes(p.memory_rss)}</td>
                <td>${p.status || '-'}</td>
                <td>${p.username || '-'}</td>
                <td><button class="btn-detail" onclick="fetchProcessDetail(${p.pid})">详情</button></td>
            </tr>
        `;
    });
    document.getElementById('process-list').innerHTML = html;
}

function showProcessDetail(process) {
    const modal = document.getElementById('process-detail-modal');
    const content = document.getElementById('process-detail-content');
    
    content.innerHTML = `
        <div class="process-detail-item"><span>PID:</span><span>${process.pid}</span></div>
        <div class="process-detail-item"><span>进程名:</span><span>${process.name || '-'}</span></div>
        <div class="process-detail-item"><span>执行路径:</span><span>${process.exe || '-'}</span></div>
        <div class="process-detail-item"><span>命令行:</span><span>${process.cmdline || '-'}</span></div>
        <div class="process-detail-item"><span>CPU使用率:</span><span>${process.cpu_percent.toFixed(2)}%</span></div>
        <div class="process-detail-item"><span>内存使用率:</span><span>${process.memory_percent.toFixed(2)}%</span></div>
        <div class="process-detail-item"><span>内存占用:</span><span>${formatBytes(process.memory_rss)}</span></div>
        <div class="process-detail-item"><span>状态:</span><span>${process.status || '-'}</span></div>
        <div class="process-detail-item"><span>用户:</span><span>${process.username || '-'}</span></div>
        <div class="process-detail-item"><span>创建时间:</span><span>${formatDateTime(process.create_time)}</span></div>
    `;
    
    modal.style.display = 'block';
}

function renderHistoryChart(data, hours) {
    const ctx = document.getElementById('historyChart').getContext('2d');
    const metrics = getSelectedMetrics();
    
    if (historyChart) {
        historyChart.destroy();
    }

    const labels = data.map(d => formatTimestamp(d.created_at));
    const datasets = [];

    if (metrics.includes('cpu')) {
        datasets.push({
            label: 'CPU (%)',
            data: data.map(d => d.cpu_usage),
            borderColor: '#fbbf24',
            backgroundColor: 'rgba(251, 191, 36, 0.1)',
            fill: true,
            tension: 0.4,
            yAxisID: 'y'
        });
    }

    if (metrics.includes('memory')) {
        datasets.push({
            label: '内存 (%)',
            data: data.map(d => d.memory_usage),
            borderColor: '#22d3ee',
            backgroundColor: 'rgba(34, 211, 238, 0.1)',
            fill: true,
            tension: 0.4,
            yAxisID: 'y'
        });
    }

    if (metrics.includes('disk')) {
        datasets.push({
            label: '磁盘 (%)',
            data: data.map(d => d.disk_usage),
            borderColor: '#34d399',
            backgroundColor: 'rgba(52, 211, 153, 0.1)',
            fill: true,
            tension: 0.4,
            yAxisID: 'y'
        });
    }

    if (metrics.includes('network')) {
        datasets.push({
            label: '入站 (KB/s)',
            data: data.map(d => (d.network_in / 1024).toFixed(1)),
            borderColor: '#8b5cf6',
            backgroundColor: 'rgba(139, 92, 246, 0.1)',
            fill: true,
            tension: 0.4,
            yAxisID: 'y1'
        });
        datasets.push({
            label: '出站 (KB/s)',
            data: data.map(d => (d.network_out / 1024).toFixed(1)),
            borderColor: '#ec4899',
            backgroundColor: 'rgba(236, 72, 153, 0.1)',
            fill: true,
            tension: 0.4,
            yAxisID: 'y1'
        });
    }

    historyChart = new Chart(ctx, {
        type: 'line',
        data: {
            labels: labels,
            datasets: datasets
        },
        options: {
            responsive: true,
            maintainAspectRatio: false,
            interaction: {
                mode: 'index',
                intersect: false
            },
            plugins: {
                legend: {
                    position: 'top',
                    labels: {
                        color: '#fff'
                    }
                },
                tooltip: {
                    backgroundColor: 'rgba(0, 0, 0, 0.8)',
                    titleColor: '#fff',
                    bodyColor: '#fff'
                }
            },
            scales: {
                x: {
                    ticks: {
                        color: '#9ca3af',
                        maxTicksLimit: 12
                    },
                    grid: {
                        color: 'rgba(255, 255, 255, 0.1)'
                    }
                },
                y: {
                    type: 'linear',
                    display: true,
                    position: 'left',
                    min: 0,
                    max: 100,
                    ticks: {
                        color: '#9ca3af',
                        callback: (value) => value + '%'
                    },
                    grid: {
                        color: 'rgba(255, 255, 255, 0.1)'
                    },
                    title: {
                        display: true,
                        text: '使用率 (%)',
                        color: '#9ca3af'
                    }
                },
                y1: {
                    type: 'linear',
                    display: metrics.includes('network'),
                    position: 'right',
                    ticks: {
                        color: '#9ca3af',
                        callback: (value) => value + ' KB/s'
                    },
                    grid: {
                        drawOnChartArea: false
                    },
                    title: {
                        display: true,
                        text: '网络 (KB/s)',
                        color: '#9ca3af'
                    }
                }
            }
        }
    });
}

function formatBytes(bytes) {
    if (bytes === 0) return '0 B';
    const k = 1024;
    const sizes = ['B', 'KB', 'MB', 'GB', 'TB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
}

function formatUptime(seconds) {
    if (!seconds) return '-';
    const days = Math.floor(seconds / 86400);
    const hours = Math.floor((seconds % 86400) / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    const secs = Math.floor(seconds % 60);
    
    if (days > 0) {
        return `${days}天 ${hours}时 ${minutes}分`;
    } else if (hours > 0) {
        return `${hours}时 ${minutes}分 ${secs}秒`;
    } else if (minutes > 0) {
        return `${minutes}分 ${secs}秒`;
    } else {
        return `${secs}秒`;
    }
}

function formatTime(timestamp) {
    if (!timestamp) return '-';
    const date = new Date(timestamp);
    return date.toLocaleString('zh-CN', {
        year: 'numeric',
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit'
    });
}

function formatTimestamp(timestamp) {
    if (!timestamp) return '';
    const date = new Date(timestamp);
    return date.toLocaleString('zh-CN', {
        month: '2-digit',
        day: '2-digit',
        hour: '2-digit',
        minute: '2-digit'
    });
}

function formatDateTime(timestamp) {
    if (!timestamp) return '-';
    const date = new Date(timestamp);
    return date.toLocaleString('zh-CN');
}