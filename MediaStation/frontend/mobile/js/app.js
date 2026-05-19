const API_BASE = '';

let currentMedia = null;
let videoList = [];

document.addEventListener('DOMContentLoaded', () => {
    loadVideos();
    setupEventListeners();
});

function setupEventListeners() {
    document.querySelectorAll('.nav-item').forEach(item => {
        item.addEventListener('click', (e) => {
            e.preventDefault();
            const href = item.getAttribute('href');
            if (href && href.startsWith('#')) {
                const sectionId = href.substring(1);
                showSection(sectionId);
            }
        });
    });

    document.getElementById('searchBtn').addEventListener('click', () => {
        document.getElementById('searchModal').style.display = 'block';
    });

    document.getElementById('searchSubmitBtn').addEventListener('click', () => {
        const keyword = document.getElementById('searchInput').value;
        if (keyword) {
            searchVideos(keyword);
        }
    });

    document.getElementById('searchInput').addEventListener('keyup', (e) => {
        if (e.key === 'Enter') {
            const keyword = document.getElementById('searchInput').value;
            if (keyword) {
                searchVideos(keyword);
            }
        }
    });

    document.getElementById('playHeroBtn').addEventListener('click', () => {
        if (videoList.length > 0) {
            openVideoPlayer(videoList[0]);
        } else {
            openVideoPlayer({
                id: 8,
                title: '探索自然之美',
                duration: 3600,
                thumbnail: 'https://a0ai.marscode.cn/api/ide/v1/text_to_image?prompt=beautiful%20nature%20landscape&image_size=portrait_16_9'
            });
        }
    });

    document.getElementById('backBtn').addEventListener('click', () => {
        closeModal(document.getElementById('playerModal'));
    });

    document.querySelectorAll('.tab-btn').forEach(btn => {
        btn.addEventListener('click', () => {
            const tabId = btn.dataset.tab;
            showTab(tabId);
        });
    });

    document.querySelectorAll('.category-item').forEach(item => {
        item.addEventListener('click', () => {
            const mediaType = item.dataset.type;
            loadMediaByType(mediaType);
        });
    });

    document.getElementById('uploadBtn').addEventListener('click', () => {
        document.getElementById('uploadModal').style.display = 'block';
    });

    document.getElementById('uploadBackBtn').addEventListener('click', () => {
        closeModal(document.getElementById('uploadModal'));
    });

    document.getElementById('uploadForm').addEventListener('submit', async (e) => {
        e.preventDefault();
        await handleUpload();
    });

    window.addEventListener('click', (e) => {
        if (e.target.classList.contains('modal')) {
            closeModal(e.target);
        }
    });
}

function showSection(sectionId) {
    document.querySelectorAll('.section').forEach(section => {
        section.classList.remove('active');
    });
    document.querySelectorAll('.nav-item').forEach(item => {
        item.classList.remove('active');
    });

    const section = document.getElementById(sectionId);
    if (section) {
        section.classList.add('active');
    }

    const navItems = document.querySelectorAll(`[href="#${sectionId}"]`);
    navItems.forEach(item => item.classList.add('active'));
}

function showTab(tabId) {
    document.querySelectorAll('.tab-btn').forEach(btn => {
        btn.classList.remove('active');
    });
    document.querySelectorAll('.tab-content').forEach(content => {
        content.classList.remove('active');
    });

    const btn = document.querySelector(`[data-tab="${tabId}"]`);
    const content = document.getElementById(`${tabId}Content`);
    
    if (btn) btn.classList.add('active');
    if (content) content.classList.add('active');
}

function closeModal(modal) {
    const video = document.getElementById('videoPlayer');
    if (modal.id === 'playerModal') {
        if (!video.paused) {
            video.pause();
        }
        video.currentTime = 0;
        video.src = '';
    }
    modal.style.display = 'none';
}

async function loadVideos() {
    try {
        const response = await fetch(`${API_BASE}/api/media/list?type=video`);
        videoList = await response.json();
    } catch (error) {
        videoList = generateMockMedia('video');
    }
    displayMediaScroll(videoList, 'videoList', 'video');
    displayMediaGrid(videoList, 'categoryMediaGrid', 'video');
    displayHistory(videoList.slice(0, 3), 'historyList');
    displayMediaGrid(videoList.slice(3, 7), 'favoritesGrid', 'video');
    displayRelated(videoList.slice(4, 7), 'relatedList');
}

async function loadMediaByType(mediaType) {
    let mediaList;
    try {
        const response = await fetch(`${API_BASE}/api/media/list?type=${mediaType}`);
        mediaList = await response.json();
    } catch (error) {
        mediaList = generateMockMedia(mediaType);
    }
    displayMediaGrid(mediaList, 'categoryMediaGrid', mediaType);
}

function generateMockMedia(mediaType) {
    if (mediaType === 'image') {
        return [
            { id: 13, title: '风景图片1', type: 'image', thumbnail: 'https://a0ai.marscode.cn/api/ide/v1/text_to_image?prompt=beautiful%20scenery%20image&image_size=portrait_16_9' },
            { id: 14, title: '风景图片2', type: 'image', thumbnail: 'https://a0ai.marscode.cn/api/ide/v1/text_to_image?prompt=mountain%20landscape%20photo&image_size=portrait_16_9' },
            { id: 15, title: '风景图片3', type: 'image', thumbnail: 'https://a0ai.marscode.cn/api/ide/v1/text_to_image?prompt=ocean%20beach%20scene&image_size=portrait_16_9' }
        ];
    } else if (mediaType === 'audio') {
        return [
            { id: 17, title: '轻松背景音乐', type: 'audio', duration: 360, thumbnail: 'https://a0ai.marscode.cn/api/ide/v1/text_to_image?prompt=music%20notes%20audio%20wave%20art&image_size=square' },
            { id: 18, title: '自然声音合集', type: 'audio', duration: 480, thumbnail: 'https://a0ai.marscode.cn/api/ide/v1/text_to_image?prompt=forest%20nature%20sounds%20relaxing&image_size=square' },
            { id: 19, title: '古典音乐精选', type: 'audio', duration: 540, thumbnail: 'https://a0ai.marscode.cn/api/ide/v1/text_to_image?prompt=classical%20music%20piano%20elegant&image_size=square' }
        ];
    } else if (mediaType === 'novel') {
        return [
            { id: 12, title: '翻译库', type: 'novel', thumbnail: 'https://a0ai.marscode.cn/api/ide/v1/text_to_image?prompt=book%20document%20icon&image_size=square' },
            { id: 16, title: '全球主流媒体', type: 'novel', thumbnail: 'https://a0ai.marscode.cn/api/ide/v1/text_to_image?prompt=news%20article%20icon&image_size=square' }
        ];
    }
    return [
        { id: 1, title: '山川湖海的壮丽', duration: 1800, thumbnail: 'https://a0ai.marscode.cn/api/ide/v1/text_to_image?prompt=majestic%20mountains%20and%20lake&image_size=portrait_16_9' },
        { id: 2, title: '森林深处的秘密', duration: 1200, thumbnail: 'https://a0ai.marscode.cn/api/ide/v1/text_to_image?prompt=deep%20forest%20landscape&image_size=portrait_16_9' },
        { id: 3, title: '海洋世界探索', duration: 2400, thumbnail: 'https://a0ai.marscode.cn/api/ide/v1/text_to_image?prompt=underwater%20ocean%20world&image_size=portrait_16_9' },
        { id: 4, title: '日出日落美景', duration: 1500, thumbnail: 'https://a0ai.marscode.cn/api/ide/v1/text_to_image?prompt=beautiful%20sunrise%20mountains&image_size=portrait_16_9' },
        { id: 5, title: '野生动物集锦', duration: 1900, thumbnail: 'https://a0ai.marscode.cn/api/ide/v1/text_to_image?prompt=wild%20animals%20nature&image_size=portrait_16_9' },
        { id: 6, title: '城市夜景', duration: 1300, thumbnail: 'https://a0ai.marscode.cn/api/ide/v1/text_to_image?prompt=city%20night%20skyline&image_size=portrait_16_9' },
        { id: 7, title: '星空银河', duration: 2000, thumbnail: 'https://a0ai.marscode.cn/api/ide/v1/text_to_image?prompt=milky%20way%20starry%20night&image_size=portrait_16_9' },
        { id: 8, title: '瀑布奇观', duration: 1600, thumbnail: 'https://a0ai.marscode.cn/api/ide/v1/text_to_image?prompt=beautiful%20waterfall%20forest&image_size=portrait_16_9' }
    ];
}

function displayMediaScroll(mediaList, containerId, mediaType) {
    const container = document.getElementById(containerId);
    if (!container) return;
    container.innerHTML = '';

    mediaList.forEach(media => {
        const card = document.createElement('div');
        card.className = 'media-card';
        const meta = media.duration ? formatDuration(media.duration) : '';
        card.innerHTML = `
            <img src="${media.thumbnail || getDefaultThumbnail(mediaType)}" alt="${media.title}" class="media-thumbnail">
            <div class="media-info">
                <div class="media-title">${media.title}</div>
                <div class="media-meta">${meta}</div>
            </div>
        `;
        card.addEventListener('click', () => {
            openMedia(media);
        });
        container.appendChild(card);
    });
}

function displayMediaGrid(mediaList, gridId, mediaType) {
    const grid = document.getElementById(gridId);
    if (!grid) return;
    grid.innerHTML = '';
    const defaultThumbnail = getDefaultThumbnail(mediaType);

    mediaList.forEach(media => {
        const card = document.createElement('div');
        card.className = 'media-card';
        const meta = media.duration ? formatDuration(media.duration) : '';
        card.innerHTML = `
            <img src="${media.thumbnail || defaultThumbnail}" alt="${media.title}" class="media-thumbnail" onerror="this.src='${defaultThumbnail}'">
            <div class="media-info">
                <div class="media-title">${media.title}</div>
                <div class="media-meta">${meta}</div>
            </div>
        `;
        card.addEventListener('click', () => {
            openMedia(media);
        });
        grid.appendChild(card);
    });
}

function displayHistory(videos, listId) {
    const list = document.getElementById(listId);
    if (!list) return;
    list.innerHTML = '';
    const defaultThumbnail = getDefaultThumbnail('video');

    videos.forEach(video => {
        const item = document.createElement('div');
        item.className = 'history-item';
        item.innerHTML = `
            <img src="${video.thumbnail || defaultThumbnail}" alt="${video.title}" onerror="this.src='${defaultThumbnail}'">
            <div class="history-info">
                <div class="history-title">${video.title}</div>
                <div class="history-meta">${formatDuration(video.duration)}</div>
            </div>
        `;
        item.addEventListener('click', () => {
            openVideoPlayer(video);
        });
        list.appendChild(item);
    });
}

function displayRelated(videos, listId) {
    const list = document.getElementById(listId);
    if (!list) return;
    list.innerHTML = '';
    const defaultThumbnail = getDefaultThumbnail('video');

    videos.forEach(video => {
        const item = document.createElement('div');
        item.className = 'related-item';
        item.innerHTML = `
            <img src="${video.thumbnail || defaultThumbnail}" alt="${video.title}" onerror="this.src='${defaultThumbnail}'">
            <div class="related-title">${video.title}</div>
        `;
        item.addEventListener('click', () => {
            openVideoPlayer(video);
        });
        list.appendChild(item);
    });
}

function openVideoPlayer(video) {
    currentMedia = video;
    const modal = document.getElementById('playerModal');
    modal.style.display = 'block';

    const videoElement = document.getElementById('videoPlayer');
    videoElement.src = `${API_BASE}/hls/playlist?id=${video.id}`;
    videoElement.load();

    const continueItem = document.getElementById('continueItem');
    continueItem.innerHTML = `
        <img src="${video.thumbnail}" alt="${video.title}">
        <div class="continue-info">
            <div class="continue-title">${video.title}</div>
            <span class="continue-badge">继续观看</span>
        </div>
    `;
    continueItem.addEventListener('click', () => {
        videoElement.currentTime = 0;
        videoElement.play();
    });
}

function formatDuration(seconds) {
    const hours = Math.floor(seconds / 3600);
    const mins = Math.floor((seconds % 3600) / 60);
    const secs = seconds % 60;
    
    if (hours > 0) {
        return `${hours}:${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
    }
    return `${mins}:${secs.toString().padStart(2, '0')}`;
}

async function searchVideos(keyword) {
    let results;
    try {
        const response = await fetch(`${API_BASE}/api/media/search?keyword=${keyword}`);
        results = await response.json();
    } catch (error) {
        results = videoList.filter(v => v.title.includes(keyword));
    }

    const container = document.getElementById('searchResults');
    container.innerHTML = '';
    const defaultThumbnail = getDefaultThumbnail('video');

    results.forEach(video => {
        const item = document.createElement('div');
        item.className = 'search-item';
        item.innerHTML = `
            <img src="${video.thumbnail || defaultThumbnail}" alt="${video.title}" onerror="this.src='${defaultThumbnail}'">
            <div class="search-info">
                <div class="search-title">${video.title}</div>
                <div class="search-meta">${formatDuration(video.duration)}</div>
            </div>
        `;
        item.addEventListener('click', () => {
            closeModal(document.getElementById('searchModal'));
            openVideoPlayer(video);
        });
        container.appendChild(item);
    });
}

function openMedia(media) {
    currentMedia = media;
    if (media.type === 'video' || !media.type) {
        openVideoPlayer(media);
    } else if (media.type === 'audio') {
        openAudioPlayer(media);
    } else if (media.type === 'image') {
        openImageViewer(media);
    } else if (media.type === 'novel') {
        openTextViewer(media);
    }
}

function openAudioPlayer(audio) {
    const modal = document.createElement('div');
    modal.id = 'audioModal';
    modal.className = 'modal';
    modal.innerHTML = `
        <div class="modal-content audio-modal">
            <div class="modal-header">
                <button class="back-btn" id="audioBackBtn">←</button>
                <h3 class="modal-title">${audio.title}</h3>
            </div>
            <div class="audio-player">
                <img src="${audio.thumbnail || 'https://a0ai.marscode.cn/api/ide/v1/text_to_image?prompt=music%20player%20interface&image_size=square'}" alt="${audio.title}" class="audio-cover">
                <audio id="audioPlayer" controls preload="metadata">
                    <source src="${API_BASE}/api/media/stream/${audio.id}" type="audio/mpeg">
                
                <div class="audio-controls">
                    <span id="audioTime">00:00 / ${formatDuration(audio.duration || 0)}</span>
                </div>
            </div>
        </div>
    `;
    document.body.appendChild(modal);
    modal.style.display = 'block';
    
    const audioElement = modal.querySelector('#audioPlayer');
    const timeDisplay = modal.querySelector('#audioTime');
    
    audioElement.addEventListener('timeupdate', () => {
        const current = Math.floor(audioElement.currentTime);
        const duration = Math.floor(audioElement.duration) || (audio.duration || 0);
        timeDisplay.textContent = `${formatDuration(current)} / ${formatDuration(duration)}`;
    });
    
    modal.querySelector('.back-btn').addEventListener('click', () => {
        audioElement.pause();
        modal.remove();
    });
    modal.addEventListener('click', (e) => {
        if (e.target === modal) {
            audioElement.pause();
            modal.remove();
        }
    });
}

function openImageViewer(image) {
    const modal = document.createElement('div');
    modal.id = 'imageModal';
    modal.className = 'modal';
    modal.innerHTML = `
        <div class="modal-content">
            <div class="modal-header">
                <button class="back-btn" id="imageBackBtn">←</button>
                <h3 class="modal-title">${image.title}</h3>
            </div>
            <img src="${API_BASE}/api/media/stream/${image.id}" alt="${image.title}" style="max-width:100%; max-height:70vh; margin:1rem auto; display:block;">
        </div>
    `;
    document.body.appendChild(modal);
    modal.style.display = 'block';
    
    modal.querySelector('.back-btn').addEventListener('click', () => {
        modal.remove();
    });
    modal.addEventListener('click', (e) => {
        if (e.target === modal) modal.remove();
    });
}

function openTextViewer(text) {
    const modal = document.createElement('div');
    modal.id = 'textModal';
    modal.className = 'modal';
    modal.innerHTML = `
        <div class="modal-content">
            <div class="modal-header">
                <button class="back-btn" id="textBackBtn">←</button>
                <h3 class="modal-title">${text.title}</h3>
            </div>
            <div id="textContent" style="padding:1rem; max-height:60vh; overflow-y:auto; white-space: pre-wrap; word-break: break-word;"></div>
        </div>
    `;
    document.body.appendChild(modal);
    modal.style.display = 'block';
    
    fetch(`${API_BASE}/api/media/stream/${text.id}`)
        .then(res => res.text())
        .then(content => {
            document.getElementById('textContent').textContent = content;
        });
    
    modal.querySelector('.back-btn').addEventListener('click', () => {
        modal.remove();
    });
    modal.addEventListener('click', (e) => {
        if (e.target === modal) modal.remove();
    });
}

function getDefaultThumbnail(type) {
    const colors = {
        video: '#e53935',
        image: '#43a047',
        audio: '#1e88e5',
        novel: '#fb8c00'
    };
    const icons = {
        video: '<polygon points="60,45 105,75 60,105" fill="white"/>',
        image: '<rect x="60" y="45" width="68" height="45" fill="none" stroke="white" stroke-width="3"/><polygon points="128,45 105,67 105,45" fill="white"/>',
        audio: '<ellipse cx="94" cy="70" rx="35" ry="25" fill="none" stroke="white" stroke-width="3"/><path d="M75 70 Q75 55, 94 55 Q113 55, 113 70" fill="white" opacity="0.8"/>',
        novel: '<rect x="56" y="52" width="26" height="60" fill="white" opacity="0.9"/><rect x="79" y="64" width="26" height="48" fill="white" opacity="0.7"/>'
    };
    
    const color = colors[type] || colors.video;
    const icon = icons[type] || icons.video;
    
    const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="188" height="135" viewBox="0 0 188 135">
        <rect fill="${color}" width="188" height="135"/>
        <g transform="translate(0, 22)">${icon}</g>
    </svg>`;
    
    return 'data:image/svg+xml,' + encodeURIComponent(svg);
}

async function handleUpload() {
    const form = document.getElementById('uploadForm');
    const fileInput = document.getElementById('uploadFile');
    const progressContainer = document.querySelector('.upload-progress');
    const progressBar = document.getElementById('uploadProgress');
    const status = document.getElementById('uploadStatus');

    if (!fileInput.files || fileInput.files.length === 0) {
        alert('请选择要上传的文件');
        return;
    }

    const formData = new FormData(form);
    
    progressContainer.classList.add('active');
    status.textContent = '上传中...';
    progressBar.style.setProperty('--progress', '0%');

    try {
        const response = await fetch(`${API_BASE}/api/media/add`, {
            method: 'POST',
            body: formData
        });

        const result = await response.json();
        
        if (response.ok) {
            progressBar.style.setProperty('--progress', '100%');
            status.textContent = `上传成功！ID: ${result.id}`;
            
            setTimeout(() => {
                closeModal(document.getElementById('uploadModal'));
                form.reset();
                progressContainer.classList.remove('active');
                status.textContent = '';
                loadVideos();
            }, 1500);
        } else {
            status.textContent = `上传失败: ${result.message || '未知错误'}`;
        }
    } catch (error) {
        status.textContent = `上传失败: ${error.message}`;
    }
}