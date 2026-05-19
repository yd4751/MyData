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
            const sectionId = item.getAttribute('href').substring(1);
            showSection(sectionId);
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

    document.querySelectorAll('.close').forEach(close => {
        close.addEventListener('click', () => {
            const modal = close.closest('.modal');
            closeModal(modal);
        });
    });

    document.querySelectorAll('.category-card').forEach(card => {
        card.addEventListener('click', () => {
            const mediaType = card.dataset.type;
            loadMediaByType(mediaType);
        });
    });

    document.getElementById('uploadBtn').addEventListener('click', () => {
        document.getElementById('uploadModal').style.display = 'block';
    });

    document.getElementById('uploadForm').addEventListener('submit', async (e) => {
        e.preventDefault();
        await handleUpload();
    });

    document.getElementById('playHeroBtn').addEventListener('click', () => {
        if (videoList.length > 0) {
            openVideoPlayer(videoList[0]);
        } else {
            openVideoPlayer({
                id: 8,
                title: '探索自然之美',
                duration: 3600,
                thumbnail: 'https://a0ai.marscode.cn/api/ide/v1/text_to_image?prompt=beautiful%20nature%20landscape&image_size=landscape_16_9'
            });
        }
    });

    document.getElementById('playPauseBtn').addEventListener('click', togglePlayPause);
    document.getElementById('progressBar').addEventListener('input', seekTo);
    document.getElementById('playbackSpeed').addEventListener('change', changeSpeed);
    document.getElementById('fullscreenBtn').addEventListener('click', toggleFullscreen);

    document.querySelectorAll('.tab-btn').forEach(btn => {
        btn.addEventListener('click', () => {
            const tabId = btn.dataset.tab;
            showTab(tabId);
        });
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

    const navItem = document.querySelector(`[href="#${sectionId}"]`);
    if (navItem) {
        navItem.classList.add('active');
    }
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
        videoList = generateMockVideos();
    }
    displayMedia(videoList, 'videoGrid', 'video');
    displayMedia(videoList.slice(0, 4), 'categoryMediaGrid', 'video');
    displayMedia(videoList.slice(0, 3), 'historyGrid', 'video');
    displayMedia(videoList.slice(3, 6), 'favoritesGrid', 'video');
}

async function loadMediaByType(mediaType) {
    let mediaList;
    try {
        const response = await fetch(`${API_BASE}/api/media/list?type=${mediaType}`);
        mediaList = await response.json();
    } catch (error) {
        mediaList = generateMockMedia(mediaType);
    }
    displayMedia(mediaList, 'categoryMediaGrid', mediaType);
}

function generateMockMedia(mediaType) {
    if (mediaType === 'image') {
        return [
            { id: 13, title: '风景图片1', type: 'image', thumbnail: 'https://a0ai.marscode.cn/api/ide/v1/text_to_image?prompt=beautiful%20scenery%20image&image_size=landscape_16_9' },
            { id: 14, title: '风景图片2', type: 'image', thumbnail: 'https://a0ai.marscode.cn/api/ide/v1/text_to_image?prompt=mountain%20landscape%20photo&image_size=landscape_16_9' },
            { id: 15, title: '风景图片3', type: 'image', thumbnail: 'https://a0ai.marscode.cn/api/ide/v1/text_to_image?prompt=ocean%20beach%20scene&image_size=landscape_16_9' }
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
    return generateMockVideos();
}

function generateMockVideos() {
    return [
        { id: 1, title: '山川湖海的壮丽', duration: 1800, thumbnail: 'https://a0ai.marscode.cn/api/ide/v1/text_to_image?prompt=majestic%20mountains%20and%20lake&image_size=landscape_16_9' },
        { id: 2, title: '森林深处的秘密', duration: 1200, thumbnail: 'https://a0ai.marscode.cn/api/ide/v1/text_to_image?prompt=deep%20forest%20landscape&image_size=landscape_16_9' },
        { id: 3, title: '海洋世界探索', duration: 2400, thumbnail: 'https://a0ai.marscode.cn/api/ide/v1/text_to_image?prompt=underwater%20ocean%20world&image_size=landscape_16_9' },
        { id: 4, title: '日出日落美景', duration: 1500, thumbnail: 'https://a0ai.marscode.cn/api/ide/v1/text_to_image?prompt=beautiful%20sunrise%20over%20mountains&image_size=landscape_16_9' },
        { id: 5, title: '野生动物集锦', duration: 1900, thumbnail: 'https://a0ai.marscode.cn/api/ide/v1/text_to_image?prompt=wild%20animals%20in%20nature&image_size=landscape_16_9' },
        { id: 6, title: '城市夜景', duration: 1300, thumbnail: 'https://a0ai.marscode.cn/api/ide/v1/text_to_image?prompt=city%20night%20skyline&image_size=landscape_16_9' },
        { id: 7, title: '星空银河', duration: 2000, thumbnail: 'https://a0ai.marscode.cn/api/ide/v1/text_to_image?prompt=milky%20way%20starry%20night&image_size=landscape_16_9' },
        { id: 8, title: '瀑布奇观', duration: 1600, thumbnail: 'https://a0ai.marscode.cn/api/ide/v1/text_to_image?prompt=beautiful%20waterfall%20in%20forest&image_size=landscape_16_9' }
    ];
}

function displayMedia(mediaList, gridId, mediaType) {
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

function openVideoPlayer(video) {
    currentMedia = video;
    const modal = document.getElementById('playerModal');
    modal.style.display = 'block';

    const videoElement = document.getElementById('videoPlayer');
    const progress = document.getElementById('progressBar');
    const timeDisplay = document.getElementById('timeDisplay');

    videoElement.src = `${API_BASE}/hls/playlist?id=${video.id}`;
    videoElement.load();

    progress.value = 0;
    timeDisplay.textContent = `00:00 / ${formatDuration(video.duration)}`;
}

function togglePlayPause() {
    const video = document.getElementById('videoPlayer');
    const btn = document.getElementById('playPauseBtn');

    if (video.paused) {
        video.play();
        btn.textContent = '⏸';
    } else {
        video.pause();
        btn.textContent = '▶';
    }
}

function seekTo() {
    const video = document.getElementById('videoPlayer');
    const progress = document.getElementById('progressBar');
    const time = (progress.value / 100) * video.duration;
    video.currentTime = time;
}

function changeSpeed() {
    const video = document.getElementById('videoPlayer');
    const speed = document.getElementById('playbackSpeed').value;
    video.playbackRate = parseFloat(speed);
}

function toggleFullscreen() {
    const playerContainer = document.querySelector('.player-container');

    if (!document.fullscreenElement) {
        playerContainer.requestFullscreen();
    } else {
        document.exitFullscreen();
    }
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

    const defaultVideoThumbnail = getDefaultThumbnail('video');
    results.forEach(video => {
        const item = document.createElement('div');
        item.className = 'search-item';
        item.innerHTML = `
            <img src="${video.thumbnail || defaultVideoThumbnail}" alt="${video.title}" onerror="this.src='${defaultVideoThumbnail}'">
            <div class="search-item-info">
                <div class="search-item-title">${video.title}</div>
                <div class="search-item-meta">${formatDuration(video.duration)}</div>
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
            <span class="close">&times;</span>
            <div class="audio-player">
                <img src="${audio.thumbnail || 'https://a0ai.marscode.cn/api/ide/v1/text_to_image?prompt=music%20player%20interface&image_size=square'}" alt="${audio.title}" class="audio-cover">
                <h3 class="audio-title">${audio.title}</h3>
                <audio id="audioPlayer" controls preload="metadata">
                    <source src="${API_BASE}/api/media/stream/${audio.id}" type="audio/mpeg">
                </audio>
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
    
    modal.querySelector('.close').addEventListener('click', () => {
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
            <span class="close">&times;</span>
            <div style="padding: 2rem; text-align: center;">
                <h3 style="margin-bottom: 1rem; color: var(--text-primary);">${image.title}</h3>
                <img src="${API_BASE}/api/media/stream/${image.id}" alt="${image.title}" style="max-width:100%; max-height:70vh;">
            </div>
        </div>
    `;
    document.body.appendChild(modal);
    modal.style.display = 'block';
    
    modal.querySelector('.close').addEventListener('click', () => {
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
            <span class="close">&times;</span>
            <div style="padding: 2rem;">
                <h3 style="margin-bottom: 1rem; color: var(--text-primary);">${text.title}</h3>
                <div id="textContent" style="max-height:60vh; overflow-y:auto; white-space: pre-wrap; word-break: break-word;"></div>
            </div>
        </div>
    `;
    document.body.appendChild(modal);
    modal.style.display = 'block';
    
    fetch(`${API_BASE}/api/media/stream/${text.id}`)
        .then(res => res.text())
        .then(content => {
            document.getElementById('textContent').textContent = content;
        });
    
    modal.querySelector('.close').addEventListener('click', () => {
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
        video: '<polygon points="80,60 140,100 80,140" fill="white"/>',
        image: '<rect x="80" y="60" width="90" height="60" fill="none" stroke="white" stroke-width="4"/><polygon points="170,60 140,90 140,60" fill="white"/>',
        audio: '<ellipse cx="125" cy="100" rx="45" ry="35" fill="none" stroke="white" stroke-width="4"/><path d="M100 100 Q100 75, 125 75 Q150 75, 150 100" fill="white" opacity="0.8"/>',
        novel: '<rect x="75" y="70" width="35" height="80" fill="white" opacity="0.9"/><rect x="105" y="85" width="35" height="65" fill="white" opacity="0.7"/>'
    };
    
    const color = colors[type] || colors.video;
    const icon = icons[type] || icons.video;
    
    const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="250" height="180" viewBox="0 0 250 180">
        <rect fill="${color}" width="250" height="180"/>
        <g transform="translate(0, 30)">${icon}</g>
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

document.getElementById('videoPlayer').addEventListener('timeupdate', () => {
    const video = document.getElementById('videoPlayer');
    const progress = document.getElementById('progressBar');
    const timeDisplay = document.getElementById('timeDisplay');

    const currentTime = video.currentTime;
    const duration = video.duration;

    progress.value = (currentTime / duration) * 100;
    timeDisplay.textContent = `${formatDuration(Math.floor(currentTime))} / ${formatDuration(Math.floor(duration))}`;
});