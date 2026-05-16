const API_BASE = '';

let currentUser = null;
let currentMedia = null;
let currentEpisodeIndex = 0;
let episodeList = [];
let imageList = [];
let currentImageIndex = 0;
let currentMediaList = [];
let currentMediaIndex = 0;
let videoList = [];
let audioList = [];
let novelList = [];

document.addEventListener('DOMContentLoaded', () => {
    loadMedia('video');
    setupEventListeners();
});

function setupEventListeners() {
    document.querySelectorAll('.nav-item').forEach(item => {
        item.addEventListener('click', (e) => {
            e.preventDefault();
            const target = item.getAttribute('href').substring(1);
            showSection(target);
            loadMedia(target);
        });
    });

    document.querySelectorAll('.category-list a').forEach(item => {
        item.addEventListener('click', (e) => {
            e.preventDefault();
            const type = item.dataset.type;
            loadMedia(type);
        });
    });

    document.getElementById('searchBtn').addEventListener('click', () => {
        const keyword = document.getElementById('searchInput').value;
        if (keyword) {
            searchMedia(keyword);
        }
    });

    document.getElementById('searchInput').addEventListener('keyup', (e) => {
        if (e.key === 'Enter') {
            const keyword = document.getElementById('searchInput').value;
            if (keyword) {
                searchMedia(keyword);
            }
        }
    });

    document.getElementById('loginBtn').addEventListener('click', () => {
        document.getElementById('loginModal').style.display = 'block';
    });

    document.getElementById('registerBtn').addEventListener('click', () => {
        document.getElementById('registerModal').style.display = 'block';
    });

    document.querySelectorAll('.close').forEach(close => {
        close.addEventListener('click', () => {
            const modal = close.closest('.modal');
            closeModal(modal);
        });
    });

    document.getElementById('loginForm').addEventListener('submit', handleLogin);
    document.getElementById('registerForm').addEventListener('submit', handleRegister);

    document.getElementById('playPauseBtn').addEventListener('click', togglePlayPause);
    document.getElementById('progressBar').addEventListener('input', seekTo);
    document.getElementById('progressBar').addEventListener('click', seekTo);
    document.getElementById('playbackSpeed').addEventListener('change', changeSpeed);
    document.getElementById('fullscreenBtn').addEventListener('click', toggleFullscreen);
    document.getElementById('screenshotBtn').addEventListener('click', takeScreenshot);
    document.getElementById('rotateBtn').addEventListener('click', rotateVideo);

    document.getElementById('prevImage').addEventListener('click', showPrevItem);
    document.getElementById('nextImage').addEventListener('click', showNextItem);
    document.getElementById('prevChapter').addEventListener('click', showPrevItem);
    document.getElementById('nextChapter').addEventListener('click', showNextItem);

    window.addEventListener('click', (e) => {
        if (e.target.classList.contains('modal')) {
            closeModal(e.target);
        }
    });
    
    document.getElementById('playerModal').addEventListener('click', (e) => {
        e.stopPropagation();
    });
    
    document.addEventListener('fullscreenchange', onFullscreenChange);
    document.addEventListener('webkitfullscreenchange', onFullscreenChange);
    document.addEventListener('mozfullscreenchange', onFullscreenChange);
    document.addEventListener('MSFullscreenChange', onFullscreenChange);
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

function showSection(sectionId) {
    document.querySelectorAll('.section').forEach(section => {
        section.classList.remove('active');
    });
    document.querySelectorAll('.nav-item').forEach(item => {
        item.classList.remove('active');
    });

    const section = document.getElementById(`${sectionId}Section`);
    if (section) {
        section.classList.add('active');
    }

    const navItem = document.querySelector(`[href="#${sectionId}"]`);
    if (navItem) {
        navItem.classList.add('active');
    }
}

async function loadMedia(type) {
    const response = await fetch(`${API_BASE}/api/media/list?type=${type}`);
    const mediaList = await response.json();
    displayMedia(mediaList, type);
}

function displayMedia(mediaList, type) {
    // 添加这几行
    if (type === 'video') videoList = mediaList;
    else if (type === 'audio') audioList = mediaList;
    else if (type === 'image') imageList = mediaList;
    else if (type === 'novel') novelList = mediaList;
    
    let gridId = `${type}Grid`;
    if (['short', 'movie', 'tv'].includes(type)) {
        gridId = 'videoGrid';
    }
    const grid = document.getElementById(gridId);
    if (!grid) {
        console.error(`Grid element ${gridId} not found`);
        return;
    }
    grid.innerHTML = '';

    mediaList.forEach(media => {
        const card = document.createElement('div');
        card.className = 'media-card';
        card.innerHTML = `
            <img src="${media.thumbnail || getDefaultThumbnail(media.type)}" alt="${media.title}" onerror="this.src='${getDefaultThumbnail(media.type)}';">
            <div class="play-overlay"></div>
            <div class="media-info">
                <div class="media-title">${media.title}</div>
                <div class="media-meta">
                    ${media.type === 'video' ? `${formatDuration(media.duration)}` : ''}
                    ${media.seriesID ? ` | 剧集 ${media.episode}` : ''}
                </div>
            </div>
        `;
        card.addEventListener('click', () => {
            currentMedia = media;
            openMedia(media);
        });
        grid.appendChild(card);
    });
}

function openMedia(media) {
    currentMedia = media;
    
    // 添加这部分
    switch(media.type) {
        case 'video':
            currentMediaList = videoList;
            currentMediaIndex = videoList.findIndex(m => m.id === media.id);
            break;
        case 'audio':
            currentMediaList = audioList;
            currentMediaIndex = audioList.findIndex(m => m.id === media.id);
            break;
        case 'image':
            currentMediaList = imageList;
            currentMediaIndex = imageList.findIndex(m => m.id === media.id);
            break;
        case 'novel':
            currentMediaList = novelList;
            currentMediaIndex = novelList.findIndex(m => m.id === media.id);
            break;
    }

    switch(media.type) {
        case 'video':
            openVideoPlayer(media);
            break;
        case 'audio':
            openAudioPlayer(media);
            break;
        case 'image':
            openImageViewer(media);
            break;
        case 'novel':
            openNovelReader(media);
            break;
        default:
            openVideoPlayer(media);
    }
}

async function openVideoPlayer(media) {
    const modal = document.getElementById('playerModal');
    modal.style.display = 'block';

    currentMedia = media;
    seekOffset = 0;

    const video = document.getElementById('videoPlayer');
    const btn = document.getElementById('playPauseBtn');
    const progress = document.getElementById('progressBar');
    const timeDisplay = document.getElementById('timeDisplay');
    
    if (hls) {
        hls.destroy();
        hls = null;
    }
    
    video.currentTime = 0;
    progress.value = 0;
    timeDisplay.textContent = `00:00 / ${formatTime(media.duration)}`;
    btn.textContent = '播放';
    
    const ext = media.file_path.split('.').pop().toLowerCase();
    
    
    
    if (Hls.isSupported()) {
        hls = new Hls({
            enableWorker: true,
            lowLatencyMode: false,
            autoStartLoad: true,
            maxBufferLength: 120,
            maxBufferSize: 200 * 1000 * 1000,
            maxBufferHole: 0.5,
            maxFragLookUpTolerance: 0.5,
            stretchShortVideoTrack: true,
            startLevel: -1,
            backBufferLength: 30,
            debug: true,
            maxLoadDelay: 0,
            abrEwmaFastLive: 3,
            abrEwmaSlowLive: 9,
            abrEwmaFastVoD: 3,
            abrEwmaSlowVoD: 9,
            maxStartBitrate: 10 * 1000 * 1000,
            enableStitching: true,
            manifestLoadingTimeOut: 10000,
            manifestLoadingMaxRetry: 2,
            levelLoadingTimeOut: 10000,
            levelLoadingMaxRetry: 2,
            fragLoadingTimeOut: 10000,
            fragLoadingMaxRetry: 2,
            enableManifestCache: true,
        });
        
        hls.on(Hls.Events.MANIFEST_PARSED, function() {
            // console.log('HLS manifest parsed, levels:', hls.levels ? hls.levels.length : 'N/A');
            // console.log('Available bandwidth:', hls.bandwidthEstimate);
            if (hls.levels && hls.levels.length > 0) {
                // console.log('Level details:', hls.levels[0]);
            }
            btn.textContent = '播放';
        });
        
        hls.on(Hls.Events.FRAG_LOADED, function(event, data) {
            // console.log('HLS fragment loaded:', data.frag.start, '-', data.frag.end);
        });
        
        hls.on(Hls.Events.FRAG_BUFFERED, function(event, data) {
            // console.log('HLS fragment buffered:', data.frag ? data.frag.start : 'unknown');
            if (video.buffered.length > 0) {
                let lastEnd = 0;
                for (let i = 0; i < video.buffered.length; i++) {
                    const start = video.buffered.start(i);
                    const end = video.buffered.end(i);
                    // console.log(`  Buffer ${i}: ${start.toFixed(2)} - ${end.toFixed(2)}s`);
                    if (end > lastEnd) lastEnd = end;
                }
                // console.log(`  Ready to play up to: ${lastEnd.toFixed(2)}s`);
            }
        });
        
        hls.on(Hls.Events.FRAG_CHANGED, function(event, data) {
            // console.log('HLS fragment changed:', data.frag ? data.frag.start : 'unknown');
        });
        
        hls.on(Hls.Events.BUFFER_CREATED, function() {
            // console.log('HLS buffer created');
        });
        
        hls.on(Hls.Events.BUFFER_APPENDED, function(event, data) {
            // console.log('HLS buffer appended:', data);
        });
        
        hls.on(Hls.Events.BUFFER_FULL, function() {
            // console.log('HLS buffer full');
        });
        
        hls.on(Hls.Events.ERROR, function(event, data) {
            console.error('HLS error:', data);
            if (data.fatal) {
                console.error('Fatal HLS error:', data.type, data.details);
                switch (data.type) {
                    case Hls.ErrorTypes.NETWORK_ERROR:
                        // console.log('Network error, trying to recover...');
                        setTimeout(() => hls.startLoad(), 1000);
                        break;
                    case Hls.ErrorTypes.MEDIA_ERROR:
                        // console.log('Media error, trying to recover...');
                        hls.recoverMediaError();
                        break;
                    default:
                        // console.log('Fatal error, cannot recover');
                        break;
                }
            } else {
                console.warn('Non-fatal HLS error:', data);
                if (data.details === 'bufferSeekOverHole') {
                    handleBufferHole(data);
                }
            }
        });
        
        hls.on(Hls.Events.LEVEL_SWITCH, function(event, data) {
            // console.log('HLS level switched to:', data.level);
        });
        
        // console.log('Loading HLS source:', `${API_BASE}/hls/playlist?id=${media.id}`);
        hls.loadSource(`${API_BASE}/hls/playlist?id=${media.id}`);
        hls.attachMedia(video);
    } else {
        // console.log('HLS not supported, using native video');
        video.src = `${API_BASE}/hls/playlist?id=${media.id}`;
        video.load();
    }

    if (media.isVertical) {
        video.style.maxHeight = '90vh';
        video.style.width = 'auto';
        video.style.maxWidth = '400px';
        video.style.margin = '0 auto';
    } else {
        video.style.maxHeight = '70vh';
        video.style.width = '100%';
        video.style.maxWidth = '100%';
    }

    loadEpisodes(media.seriesID);
    loadProgress(media.id);
}

async function loadEpisodes(seriesID) {
    if (!seriesID) {
        document.querySelector('.episode-list').style.display = 'none';
        return;
    }

    document.querySelector('.episode-list').style.display = 'block';
    const response = await fetch(`${API_BASE}/api/media/series?series_id=${seriesID}`);
    episodeList = await response.json();

    const container = document.getElementById('episodeContainer');
    container.innerHTML = '';

    episodeList.forEach((episode, index) => {
        const item = document.createElement('div');
        item.className = 'episode-item';
        item.textContent = `第 ${episode.episode} 集`;
        item.addEventListener('click', () => {
            currentEpisodeIndex = index;
            playEpisode(episode);
        });
        container.appendChild(item);
    });
}

function playEpisode(episode) {
    currentMedia = episode;
    const video = document.getElementById('videoPlayer');
    const btn = document.getElementById('playPauseBtn');
    
    if (hls) {
        hls.destroy();
        hls = null;
    }
    
    const ext = episode.file_path.split('.').pop().toLowerCase();
    const supportedExts = ['mp4', 'webm', 'ogg'];
    
    if (supportedExts.includes(ext)) {
        video.src = `${API_BASE}/stream?id=${episode.id}`;
        video.load();
    } else {
        if (Hls.isSupported()) {
            hls = new Hls({
                enableWorker: true,
                lowLatencyMode: true
            });
            hls.loadSource(`${API_BASE}/hls/playlist?id=${episode.id}`);
            hls.attachMedia(video);
        } else {
            video.src = `${API_BASE}/hls/playlist?id=${episode.id}`;
            video.load();
        }
    }
    
    video.currentTime = 0;
    
    video.addEventListener('loadeddata', function playAfterLoad() {
        video.removeEventListener('loadeddata', playAfterLoad);
        video.play().then(() => {
            btn.textContent = '暂停';
        }).catch(err => {
            // console.log('Auto-play blocked:', err);
            btn.textContent = '播放';
        });
    });
}

function togglePlayPause() {
    const video = document.getElementById('videoPlayer');
    const btn = document.getElementById('playPauseBtn');

    // console.log(`Toggle play/pause - paused: ${video.paused}, readyState: ${video.readyState}`);
    
    if (video.paused) {
        if (hls && !hls.levels.length) {
            // console.log('Starting HLS load...');
            btn.textContent = '加载中...';
            hls.startLoad();
            
            const waitForLoad = setInterval(() => {
                // console.log(`Waiting for HLS - readyState: ${video.readyState}`);
                if (video.readyState >= 2) {
                    clearInterval(waitForLoad);
                    // console.log('HLS ready, trying to play...');
                    video.play().then(() => {
                        btn.textContent = '暂停';
                        // console.log('Playback started');
                    }).catch(err => {
                        // console.log('Play error:', err);
                        btn.textContent = '播放';
                    });
                }
            }, 200);
            
            return;
        }
        
        if (video.readyState < 2) {
            // console.log('Video not ready yet, waiting for data...');
            btn.textContent = '加载中...';
            
            const waitForReady = setInterval(() => {
                // console.log(`Waiting... readyState: ${video.readyState}`);
                if (video.readyState >= 2) {
                    clearInterval(waitForReady);
                    // console.log('Video ready, trying to play...');
                    video.play().then(() => {
                        btn.textContent = '暂停';
                        // console.log('Playback started after waiting');
                    }).catch(err => {
                        // console.log('Play error after waiting:', err);
                        btn.textContent = '播放';
                    });
                }
            }, 200);
            
            return;
        }
        
        video.play().then(() => {
            btn.textContent = '暂停';
            // console.log('Playback started');
        }).catch(err => {
            // console.log('Play error:', err);
            btn.textContent = '播放';
        });
    } else {
        video.pause();
        btn.textContent = '播放';
        // console.log('Playback paused');
    }
}

function seekTo(e) {
    const video = document.getElementById('videoPlayer');
    const progress = document.getElementById('progressBar');
    
    if (!currentMedia) {
        // console.log('No current media');
        return;
    }
    
    const rect = progress.getBoundingClientRect();
    let percent;
    
    if (e && e.type === 'click') {
        percent = (e.clientX - rect.left) / rect.width;
        progress.value = percent * 100;
    } else {
        percent = progress.value / 100;
    }
    
    let targetTime = Math.floor(currentMedia.duration * percent);
    // console.log(`Seek to: ${targetTime}s (${percent * 100}%)`);
    
    if (video.readyState < 2) {
        // console.log('Video not ready, cannot seek');
        return;
    }
    
    if (hls) {
        targetTime = adjustSeekTimeToBuffer(targetTime, video);
        // console.log(`Adjusted seek time: ${targetTime}s`);
    }
    
    const wasPlaying = !video.paused;
    
    video.currentTime = targetTime;
    progress.value = (targetTime / currentMedia.duration) * 100;
    
    if (wasPlaying) {
        video.play().then(() => {
            // console.log('Resumed playback after seek');
        }).catch(err => {
            // console.log('Failed to resume after seek:', err);
        });
    }
}

function adjustSeekTimeToBuffer(targetTime, video) {
    const bufferThreshold = 2;
    
    for (let i = 0; i < video.buffered.length; i++) {
        const start = video.buffered.start(i);
        const end = video.buffered.end(i);
        
        if (targetTime >= start && targetTime <= end) {
            return targetTime;
        }
        
        if (targetTime >= end - bufferThreshold && targetTime <= end + bufferThreshold) {
            return end - 0.5;
        }
    }
    
    return targetTime;
}

function changeSpeed() {
    const video = document.getElementById('videoPlayer');
    const speed = document.getElementById('playbackSpeed').value;
    video.playbackRate = parseFloat(speed);
}

function toggleFullscreen() {
    const playerContainer = document.querySelector('.player-container');
    
    if (!document.fullscreenElement && !document.webkitFullscreenElement && !document.mozFullScreenElement && !document.msFullscreenElement) {
        try {
            if (playerContainer.requestFullscreen) {
                playerContainer.requestFullscreen();
            } else if (playerContainer.webkitRequestFullscreen) {
                playerContainer.webkitRequestFullscreen();
            } else if (playerContainer.mozRequestFullScreen) {
                playerContainer.mozRequestFullScreen();
            } else if (playerContainer.msRequestFullscreen) {
                playerContainer.msRequestFullscreen();
            }
        } catch (e) {
            console.log('Fullscreen error:', e);
        }
    } else {
        try {
            if (document.exitFullscreen) {
                document.exitFullscreen();
            } else if (document.webkitExitFullscreen) {
                document.webkitExitFullscreen();
            } else if (document.mozCancelFullScreen) {
                document.mozCancelFullScreen();
            } else if (document.msExitFullscreen) {
                document.msExitFullscreen();
            }
        } catch (e) {
            console.log('Exit fullscreen error:', e);
        }
    }
}

function onFullscreenChange() {
    const video = document.getElementById('videoPlayer');
    const playerControls = document.querySelector('.player-controls');
    const closeBtn = document.querySelector('.modal-content .close');
    
    if (document.fullscreenElement || document.webkitFullscreenElement || document.mozFullScreenElement || document.msFullscreenElement) {
        video.style.width = '100%';
        video.style.height = 'auto';
        video.style.maxHeight = 'calc(100vh - 80px)';
        video.style.objectFit = 'contain';
        video.style.borderRadius = '0';
        
        playerControls.style.position = 'fixed';
        playerControls.style.bottom = '0';
        playerControls.style.left = '0';
        playerControls.style.right = '0';
        playerControls.style.background = 'linear-gradient(to top, rgba(0,0,0,0.9), transparent)';
        playerControls.style.zIndex = '10000';
        playerControls.style.padding = '1rem';
        
        closeBtn.style.position = 'fixed';
        closeBtn.style.top = '20px';
        closeBtn.style.right = '20px';
        closeBtn.style.zIndex = '10000';
    } else {
        video.style.width = '';
        video.style.height = '';
        video.style.maxHeight = '';
        video.style.objectFit = '';
        video.style.borderRadius = '';
        
        playerControls.style.position = '';
        playerControls.style.bottom = '';
        playerControls.style.left = '';
        playerControls.style.right = '';
        playerControls.style.background = '';
        playerControls.style.zIndex = '';
        playerControls.style.padding = '';
        
        closeBtn.style.position = '';
        closeBtn.style.top = '';
        closeBtn.style.right = '';
        closeBtn.style.zIndex = '';
    }
}

function takeScreenshot() {
    const video = document.getElementById('videoPlayer');
    const canvas = document.createElement('canvas');
    canvas.width = video.videoWidth;
    canvas.height = video.videoHeight;
    canvas.getContext('2d').drawImage(video, 0, 0);

    const link = document.createElement('a');
    link.download = `screenshot-${Date.now()}.png`;
    link.href = canvas.toDataURL();
    link.click();
}

function rotateVideo() {
    const video = document.getElementById('videoPlayer');
    const currentRotation = parseInt(video.style.transform.replace('rotate(', '').replace('deg)', '')) || 0;
    video.style.transform = `rotate(${currentRotation + 90}deg)`;
}

async function loadProgress(mediaID) {
    if (!currentUser) return;

    const response = await fetch(`${API_BASE}/api/history/progress?user_id=${currentUser.id}&media_id=${mediaID}`);
    const data = await response.json();
    
    // console.log(`loadProgress - data.progress: ${data.progress}, currentMedia.duration: ${currentMedia ? currentMedia.duration : 'N/A'}`);
    
    if (data.progress > 0) {
        const video = document.getElementById('videoPlayer');
        const progressBar = document.getElementById('progressBar');
        let progress = data.progress;
        
        const maxProgress = currentMedia && currentMedia.duration > 0 ? currentMedia.duration - 1 : 100;
        
        if (progress >= maxProgress) {
            progress = Math.max(0, maxProgress - 5);
            // console.log(`Progress ${data.progress} exceeds max (${maxProgress}), adjusted to ${progress}`);
        }
        
        const setProgress = () => {
            const duration = currentMedia && currentMedia.duration > 0 ? currentMedia.duration : (video.duration > 0 ? video.duration : 100);
            video.currentTime = progress;
            progressBar.value = (progress / duration) * 100;
            // console.log(`Set currentTime: ${progress}, progressBar: ${progressBar.value}%, duration: ${duration}`);
        };
        
        if (video.readyState >= 2) {
            setProgress();
        } else {
            setTimeout(() => {
                if (video.readyState >= 2) {
                    setProgress();
                }
            }, 1000);
        }
    }
}

async function saveProgress() {
    if (!currentUser || !currentMedia) return;

    const video = document.getElementById('videoPlayer');
    let progress = seekOffset + video.currentTime;
    
    if (currentMedia.duration > 0 && progress > currentMedia.duration) {
        progress = currentMedia.duration;
    }
    
    await fetch(`${API_BASE}/api/history/save`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            user_id: currentUser.id,
            media_id: currentMedia.id,
            progress: Math.floor(progress)
        })
    });
}

let isSeeking = false;
let seekOffset = 0;
let hls = null;

let timeupdateTimeout = null;

document.getElementById('videoPlayer').addEventListener('timeupdate', () => {
    if (timeupdateTimeout) {
        clearTimeout(timeupdateTimeout);
    }
    
    timeupdateTimeout = setTimeout(() => {
        const video = document.getElementById('videoPlayer');
        const progress = document.getElementById('progressBar');
        const timeDisplay = document.getElementById('timeDisplay');

        if (currentMedia && currentMedia.duration > 0) {
            const currentTime = !isNaN(video.currentTime) ? video.currentTime : 0;
            const duration = !isNaN(video.duration) && video.duration > 0 ? video.duration : currentMedia.duration;
            
            progress.value = (currentTime / duration) * 100;
            timeDisplay.textContent = `${formatTime(currentTime)} / ${formatTime(duration)}`;
        } else if (!isNaN(video.duration) && video.duration > 0) {
            progress.value = (video.currentTime / video.duration) * 100;
            timeDisplay.textContent = `${formatTime(video.currentTime)} / ${formatTime(video.duration)}`;
        } else {
            timeDisplay.textContent = `${formatTime(video.currentTime)} / Streaming...`;
        }
    }, 100);
});

document.getElementById('videoPlayer').addEventListener('loadedmetadata', () => {
    const video = document.getElementById('videoPlayer');
    const progress = document.getElementById('progressBar');
    const timeDisplay = document.getElementById('timeDisplay');
    
    // console.log(`loadedmetadata - video.duration: ${video.duration}, currentMedia.duration: ${currentMedia ? currentMedia.duration : 'N/A'}`);
    
    if (currentMedia && currentMedia.duration > 0 && (isNaN(video.duration) || video.duration < 10)) {
        Object.defineProperty(video, 'duration', {
            get: function() { return currentMedia.duration; },
            configurable: true
        });
        // console.log(`Fixed duration to: ${currentMedia.duration}s`);
    }
    
    if (currentMedia && currentMedia.duration > 0) {
        progress.max = 100;
        timeDisplay.textContent = `${formatTime(video.currentTime)} / ${formatTime(currentMedia.duration)}`;
    }
});

document.getElementById('videoPlayer').addEventListener('loadeddata', () => {
    if (isSeeking && seekOffset > 0 && currentMedia && currentMedia.duration > 0) {
        const progress = document.getElementById('progressBar');
        progress.value = (seekOffset / currentMedia.duration) * 100;
    }
});

document.getElementById('videoPlayer').addEventListener('pause', saveProgress);
document.getElementById('videoPlayer').addEventListener('ended', () => {
    saveProgress();
    if (episodeList.length > currentEpisodeIndex + 1) {
        currentEpisodeIndex++;
        playEpisode(episodeList[currentEpisodeIndex]);
    }
});

function formatDuration(seconds) {
    const mins = Math.floor(seconds / 60);
    const secs = seconds % 60;
    return `${mins}:${secs.toString().padStart(2, '0')}`;
}

function formatTime(seconds) {
    const mins = Math.floor(seconds / 60);
    const secs = Math.floor(seconds % 60);
    return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
}

async function searchMedia(keyword) {
    const response = await fetch(`${API_BASE}/api/media/search?keyword=${keyword}`);
    const mediaList = await response.json();
    
    const activeSection = document.querySelector('.section.active');
    if (!activeSection) {
        console.error('No active section found');
        return;
    }
    const sectionId = activeSection.id.replace('Section', '');
    const grid = document.getElementById(`${sectionId}Grid`);
    if (!grid) {
        console.error(`Grid element ${sectionId}Grid not found`);
        return;
    }
    grid.innerHTML = '';

    mediaList.forEach(media => {
        const card = document.createElement('div');
        card.className = 'media-card';
        card.innerHTML = `
            <img src="${media.thumbnail || getDefaultThumbnail(media.type)}" alt="${media.title}" onerror="this.src='${getDefaultThumbnail(media.type)}';">
            <div class="play-overlay"></div>
            <div class="media-info">
                <div class="media-title">${media.title}</div>
                <div class="media-meta">${media.type}</div>
            </div>
        `;
        card.addEventListener('click', () => {
            currentMedia = media;
            openMedia(media);
        });
        grid.appendChild(card);
    });
}

async function loadHistory() {
    if (!currentUser) {
        document.getElementById('historyList').innerHTML = '<li>请登录查看历史记录</li>';
        return;
    }

    const response = await fetch(`${API_BASE}/api/history?user_id=${currentUser.id}`);
    const history = await response.json();

    const list = document.getElementById('historyList');
    list.innerHTML = '';

    history.forEach(item => {
        const li = document.createElement('li');
        li.innerHTML = `<a href="#" data-media-id="${item.media_id}">媒体 ID: ${item.media_id} (进度: ${item.progress}s)</a>`;
        list.appendChild(li);
    });
}

async function handleLogin(e) {
    e.preventDefault();
    
    const username = document.getElementById('loginUsername').value;
    const password = document.getElementById('loginPassword').value;

    const response = await fetch(`${API_BASE}/api/user/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password })
    });

    if (response.ok) {
        currentUser = await response.json();
        document.getElementById('loginModal').style.display = 'none';
        document.querySelector('.user-panel').innerHTML = `
            <span>欢迎, ${currentUser.username}</span>
            <button id="logoutBtn">退出</button>
        `;
        document.getElementById('logoutBtn').addEventListener('click', logout);
        loadHistory();
    } else {
        alert('登录失败');
    }
}

async function handleRegister(e) {
    e.preventDefault();
    
    const username = document.getElementById('regUsername').value;
    const email = document.getElementById('regEmail').value;
    const password = document.getElementById('regPassword').value;

    const response = await fetch(`${API_BASE}/api/user/register`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, email, password })
    });

    if (response.ok) {
        alert('注册成功, 请登录');
        document.getElementById('registerModal').style.display = 'none';
    } else {
        alert('注册失败');
    }
}

function logout() {
    currentUser = null;
    document.querySelector('.user-panel').innerHTML = `
        <button id="loginBtn">登录</button>
        <button id="registerBtn">注册</button>
    `;
    document.getElementById('loginBtn').addEventListener('click', () => {
        document.getElementById('loginModal').style.display = 'block';
    });
    document.getElementById('registerBtn').addEventListener('click', () => {
        document.getElementById('registerModal').style.display = 'block';
    });
    loadHistory();
}

function showPrevImage() {
    if (currentImageIndex > 0) {
        currentImageIndex--;
        document.getElementById('viewImage').src = imageList[currentImageIndex].filePath;
    }
}

function showNextImage() {
    if (currentImageIndex < imageList.length - 1) {
        currentImageIndex++;
        document.getElementById('viewImage').src = imageList[currentImageIndex].filePath;
    }
}

function handleBufferHole(error) {
    const video = document.getElementById('videoPlayer');
    // console.log('Handling buffer hole:', error);
    
    if (hls && video.buffered.length > 0) {
        let maxEnd = 0;
        for (let i = 0; i < video.buffered.length; i++) {
            const end = video.buffered.end(i);
            if (end > maxEnd) maxEnd = end;
        }
        
        if (video.currentTime > maxEnd) {
            video.currentTime = Math.max(0, maxEnd - 1);
            // console.log('Adjusted currentTime to avoid buffer hole:', video.currentTime);
        }
    }
}

async function openAudioPlayer(media) {
    const modal = document.getElementById('audioPlayerModal');
    if (!modal) {
        createAudioPlayerModal();
    }
    
    document.getElementById('audioPlayerModal').style.display = 'block';
    document.getElementById('audioTitle').textContent = media.title;
    document.getElementById('audioSource').src = `${API_BASE}/api/media/stream/${media.id}`;
    
    const audioPlayer = document.getElementById('audioPlayer');
    audioPlayer.load();
    
    const audioProgress = document.getElementById('audioProgress');
    const audioTimeDisplay = document.getElementById('audioTimeDisplay');
    audioProgress.value = 0;
    audioTimeDisplay.textContent = `00:00 / ${formatTime(media.duration)}`;
    
    document.getElementById('audioPlayPauseBtn').textContent = '播放';
}

function createAudioPlayerModal() {
    const modal = document.createElement('div');
    modal.id = 'audioPlayerModal';
    modal.className = 'modal';
    modal.innerHTML = `
        <div class="modal-content audio-player-content">
            <span class="close">&times;</span>
            <div class="audio-player-container">
                <h3 id="audioTitle"></h3>
                <img id="audioCover" src="" alt="封面" style="max-width:200px; margin:1rem auto; display:block; border-radius:8px;">
                <audio id="audioPlayer" controls style="width:100%; margin:1rem 0;">
                    <source id="audioSource" src="" type="audio/mpeg">
                </audio>
                <div class="audio-controls">
                    <button id="audioPlayPauseBtn">播放</button>
                    <input type="range" id="audioProgress" min="0" max="100" value="0">
                    <span id="audioTimeDisplay">00:00 / 00:00</span>
                    <select id="audioSpeed">
                        <option value="0.5">0.5x</option>
                        <option value="0.75">0.75x</option>
                        <option value="1" selected>1x</option>
                        <option value="1.25">1.25x</option>
                        <option value="1.5">1.5x</option>
                        <option value="2">2x</option>
                    </select>
                </div>
            </div>
        </div>
    `;
    document.body.appendChild(modal);
    
    const closeBtn = modal.querySelector('.close');
    closeBtn.addEventListener('click', () => {
        modal.style.display = 'none';
        document.getElementById('audioPlayer').pause();
    });
    
    const playPauseBtn = document.getElementById('audioPlayPauseBtn');
    const audioPlayer = document.getElementById('audioPlayer');
    const audioProgress = document.getElementById('audioProgress');
    const audioTimeDisplay = document.getElementById('audioTimeDisplay');
    
    playPauseBtn.addEventListener('click', () => {
        if (audioPlayer.paused) {
            audioPlayer.play();
            playPauseBtn.textContent = '暂停';
        } else {
            audioPlayer.pause();
            playPauseBtn.textContent = '播放';
        }
    });
    
    audioPlayer.addEventListener('timeupdate', () => {
        const progress = (audioPlayer.currentTime / audioPlayer.duration) * 100;
        audioProgress.value = progress;
        audioTimeDisplay.textContent = `${formatTime(audioPlayer.currentTime)} / ${formatTime(audioPlayer.duration)}`;
    });
    
    audioProgress.addEventListener('input', () => {
        const time = (audioProgress.value / 100) * audioPlayer.duration;
        audioPlayer.currentTime = time;
    });
    
    document.getElementById('audioSpeed').addEventListener('change', (e) => {
        audioPlayer.playbackRate = parseFloat(e.target.value);
    });
    
    modal.addEventListener('click', (e) => {
        if (e.target === modal) {
            modal.style.display = 'none';
            audioPlayer.pause();
        }
    });
}

function openImageViewer(media) {
    const modal = document.getElementById('imageViewer');
    modal.style.display = 'block';
    
    const img = document.getElementById('viewImage');
    img.src = `${API_BASE}/api/media/stream/${media.id}`;
    img.alt = media.title;
}

async function openNovelReader(media) {
    const modal = document.getElementById('novelReader');
    modal.style.display = 'block';
    
    document.getElementById('novelTitle').textContent = media.title;
    
    const response = await fetch(`${API_BASE}/api/media/stream/${media.id}`);
    const content = await response.text();
    document.getElementById('novelContent').textContent = content;
}

function getDefaultThumbnail(type) {
    const colors = {
        video: '#e53935',
        audio: '#1e88e5',
        image: '#43a047',
        novel: '#fb8c00'
    };
    const icons = {
        video: '<polygon points="80,60 140,100 80,140" fill="white"/>',
        audio: '<circle cx="125" cy="100" r="30" fill="none" stroke="white" stroke-width="4"/><circle cx="125" cy="100" r="15" fill="white"/>',
        image: '<rect x="80" y="60" width="90" height="60" fill="none" stroke="white" stroke-width="4"/><polygon points="170,60 140,90 140,60" fill="white"/>',
        novel: '<rect x="75" y="70" width="35" height="80" fill="white" opacity="0.9"/><rect x="105" y="85" width="35" height="65" fill="white" opacity="0.7"/>'
    };
    const labels = { video: '视频', audio: '音频', image: '图片', novel: '小说' };
    
    const color = colors[type] || colors.video;
    const icon = icons[type] || icons.video;
    const label = labels[type] || '视频';
    
    const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="250" height="340" viewBox="0 0 250 340">
        <defs>
            <linearGradient id="grad" x1="0%" y1="0%" x2="0%" y2="100%">
                <stop offset="0%" style="stop-color:#fff;stop-opacity:0.2"/>
                <stop offset="100%" style="stop-color:#000;stop-opacity:0.3"/>
            </linearGradient>
        </defs>
        <rect fill="${color}" width="250" height="340"/>
        <rect fill="url(#grad)" width="250" height="340"/>
        <g transform="translate(0, 70)">${icon}</g>
        <text fill="white" font-family="Arial" font-size="20" font-weight="bold" x="125" y="280" text-anchor="middle">${label}</text>
    </svg>`;
    
    return 'data:image/svg+xml,' + encodeURIComponent(svg);
}

function showHistory(history) {
    const list = document.getElementById('historyList');
    list.innerHTML = '';

    history.forEach(item => {
        const li = document.createElement('li');
        li.innerHTML = `<a href="#" data-media-id="${item.media_id}">媒体 ID: ${item.media_id} (进度: ${item.progress}s)</a>`;
        list.appendChild(li);
    });
}

async function handleLogin(e) {
    e.preventDefault();
    
    const username = document.getElementById('loginUsername').value;
    const password = document.getElementById('loginPassword').value;

    const response = await fetch(`${API_BASE}/api/user/login`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password })
    });

    if (response.ok) {
        currentUser = await response.json();
        document.getElementById('loginModal').style.display = 'none';
        document.querySelector('.user-panel').innerHTML = `
            <span>欢迎, ${currentUser.username}</span>
            <button id="logoutBtn">退出</button>
        `;
        document.getElementById('logoutBtn').addEventListener('click', logout);
        loadHistory();
    } else {
        alert('登录失败');
    }
}

async function handleRegister(e) {
    e.preventDefault();
    
    const username = document.getElementById('regUsername').value;
    const email = document.getElementById('regEmail').value;
    const password = document.getElementById('regPassword').value;

    const response = await fetch(`${API_BASE}/api/user/register`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, email, password })
    });

    if (response.ok) {
        alert('注册成功, 请登录');
        document.getElementById('registerModal').style.display = 'none';
    } else {
        alert('注册失败');
    }
}

function logout() {
    currentUser = null;
    document.querySelector('.user-panel').innerHTML = `
        <button id="loginBtn">登录</button>
        <button id="registerBtn">注册</button>
    `;
    document.getElementById('loginBtn').addEventListener('click', () => {
        document.getElementById('loginModal').style.display = 'block';
    });
    document.getElementById('registerBtn').addEventListener('click', () => {
        document.getElementById('registerModal').style.display = 'block';
    });
    loadHistory();
}

function showPrevImage() {
    if (currentImageIndex > 0) {
        currentImageIndex--;
        document.getElementById('viewImage').src = imageList[currentImageIndex].filePath;
    }
}

function showNextImage() {
    if (currentImageIndex < imageList.length - 1) {
        currentImageIndex++;
        document.getElementById('viewImage').src = imageList[currentImageIndex].filePath;
    }
}

function handleBufferHole(error) {
    const video = document.getElementById('videoPlayer');
    // console.log('Handling buffer hole:', error);
    
    if (hls && video.buffered.length > 0) {
        let maxEnd = 0;
        for (let i = 0; i < video.buffered.length; i++) {
            const end = video.buffered.end(i);
            if (end > maxEnd) maxEnd = end;
        }
        
        if (video.currentTime > maxEnd) {
            video.currentTime = Math.max(0, maxEnd - 1);
            // console.log('Adjusted currentTime to avoid buffer hole:', video.currentTime);
        }
    }
}

async function openAudioPlayer(media) {
    const modal = document.getElementById('audioPlayerModal');
    if (!modal) {
        createAudioPlayerModal();
    }
    
    document.getElementById('audioPlayerModal').style.display = 'block';
    document.getElementById('audioTitle').textContent = media.title;
    document.getElementById('audioSource').src = `${API_BASE}/api/media/stream/${media.id}`;
    
    const audioPlayer = document.getElementById('audioPlayer');
    audioPlayer.load();
    
    const audioProgress = document.getElementById('audioProgress');
    const audioTimeDisplay = document.getElementById('audioTimeDisplay');
    audioProgress.value = 0;
    audioTimeDisplay.textContent = `00:00 / ${formatTime(media.duration)}`;
    
    document.getElementById('audioPlayPauseBtn').textContent = '播放';
}

function createAudioPlayerModal() {
    const modal = document.createElement('div');
    modal.id = 'audioPlayerModal';
    modal.className = 'modal';
    modal.innerHTML = `
        <div class="modal-content audio-player-content">
            <span class="close">&times;</span>
            <div class="audio-player-container">
                <h3 id="audioTitle"></h3>
                <img id="audioCover" src="" alt="封面" style="max-width:200px; margin:1rem auto; display:block; border-radius:8px;">
                <audio id="audioPlayer" controls style="width:100%; margin:1rem 0;">
                    <source id="audioSource" src="" type="audio/mpeg">
                </audio>
                <div class="audio-controls">
                    <button id="audioPlayPauseBtn">播放</button>
                    <input type="range" id="audioProgress" min="0" max="100" value="0">
                    <span id="audioTimeDisplay">00:00 / 00:00</span>
                    <select id="audioSpeed">
                        <option value="0.5">0.5x</option>
                        <option value="0.75">0.75x</option>
                        <option value="1" selected>1x</option>
                        <option value="1.25">1.25x</option>
                        <option value="1.5">1.5x</option>
                        <option value="2">2x</option>
                    </select>
                </div>
            </div>
        </div>
    `;
    document.body.appendChild(modal);
    
    const closeBtn = modal.querySelector('.close');
    closeBtn.addEventListener('click', () => {
        modal.style.display = 'none';
        document.getElementById('audioPlayer').pause();
    });
    
    const playPauseBtn = document.getElementById('audioPlayPauseBtn');
    const audioPlayer = document.getElementById('audioPlayer');
    const audioProgress = document.getElementById('audioProgress');
    const audioTimeDisplay = document.getElementById('audioTimeDisplay');
    
    playPauseBtn.addEventListener('click', () => {
        if (audioPlayer.paused) {
            audioPlayer.play();
            playPauseBtn.textContent = '暂停';
        } else {
            audioPlayer.pause();
            playPauseBtn.textContent = '播放';
        }
    });
    
    audioPlayer.addEventListener('timeupdate', () => {
        const progress = (audioPlayer.currentTime / audioPlayer.duration) * 100;
        audioProgress.value = progress;
        audioTimeDisplay.textContent = `${formatTime(audioPlayer.currentTime)} / ${formatTime(audioPlayer.duration)}`;
    });
    
    audioProgress.addEventListener('input', () => {
        const time = (audioProgress.value / 100) * audioPlayer.duration;
        audioPlayer.currentTime = time;
    });
    
    document.getElementById('audioSpeed').addEventListener('change', (e) => {
        audioPlayer.playbackRate = parseFloat(e.target.value);
    });
    
    modal.addEventListener('click', (e) => {
        if (e.target === modal) {
            modal.style.display = 'none';
            audioPlayer.pause();
        }
    });
}

function openImageViewer(media) {
    const modal = document.getElementById('imageViewer');
    modal.style.display = 'block';
    
    const img = document.getElementById('viewImage');
    img.src = `${API_BASE}/api/media/stream/${media.id}`;
    img.alt = media.title;
}

async function openNovelReader(media) {
    const modal = document.getElementById('novelReader');
    modal.style.display = 'block';
    
    document.getElementById('novelTitle').textContent = media.title;
    
    const response = await fetch(`${API_BASE}/api/media/stream/${media.id}`);
    const content = await response.text();
    document.getElementById('novelContent').textContent = content;
}

function setCurrentMediaList(list) {
    currentMediaList = list;
}

function showPrevItem() {
    if (currentMediaList.length === 0) return;
    currentMediaIndex = (currentMediaIndex - 1 + currentMediaList.length) % currentMediaList.length;
    openMedia(currentMediaList[currentMediaIndex]);
}

function showNextItem() {
    if (currentMediaList.length === 0) return;
    currentMediaIndex = (currentMediaIndex + 1) % currentMediaList.length;
    openMedia(currentMediaList[currentMediaIndex]);
}