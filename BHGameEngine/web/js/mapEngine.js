class MapEngine {
    constructor(config) {
        this.config = Object.assign({
            chunkSize: 256,
            viewRange: 3,
            tileSize: 32,
            renderMode: '2d'
        }, config);

        this.chunkManager = new ChunkManager(this);
        this.viewport = new Viewport(this);
        this.renderer = new MapRenderer(this);
        this.isRunning = false;
        this.lastTime = 0;
        this.frameCount = 0;
        this.fps = 0;
        this.fpsUpdateTime = 0;
        
        this.playerWorldPos = { x: 512, y: 512, z: 0 };
        this.debugMode = false;
        
        this.onChunkLoaded = this.onChunkLoaded.bind(this);
        this.chunkManager.addEventListener('chunkLoaded', this.onChunkLoaded);
    }

    init(canvas) {
        const rect = canvas.getBoundingClientRect();
        const width = rect.width || canvas.offsetWidth || window.innerWidth;
        const height = rect.height || canvas.offsetHeight || window.innerHeight;
        
        canvas.width = width;
        canvas.height = height;
        
        this.renderer.init(canvas);
        this.viewport.resize(canvas.width, canvas.height);
        
        window.addEventListener('resize', () => {
            const r = canvas.getBoundingClientRect();
            canvas.width = r.width || canvas.offsetWidth;
            canvas.height = r.height || canvas.offsetHeight;
            this.viewport.resize(canvas.width, canvas.height);
        });

        this.start();
        this.loadInitialChunks();
    }

    start() {
        if (this.isRunning) return;
        this.isRunning = true;
        this.lastTime = performance.now();
        this.fpsUpdateTime = this.lastTime;
        this.gameLoop();
    }

    stop() {
        this.isRunning = false;
    }

    gameLoop() {
        if (!this.isRunning) return;

        const currentTime = performance.now();
        const deltaTime = (currentTime - this.lastTime) / 1000;
        this.lastTime = currentTime;

        this.frameCount++;
        if (currentTime - this.fpsUpdateTime >= 1000) {
            this.fps = this.frameCount;
            this.frameCount = 0;
            this.fpsUpdateTime = currentTime;
        }

        this.update(deltaTime);
        this.render();

        requestAnimationFrame(() => this.gameLoop());
    }

    update(deltaTime) {
        this.viewport.update();
        this.chunkManager.update(this.playerWorldPos);
    }

    render() {
        const ctx = this.renderer.ctx;
        const visibleChunks = this.chunkManager.getVisibleChunks();
        
        this.renderer.clear();
        
        if (visibleChunks.length === 0) {
            this.renderer.renderEmptyMap();
        } else {
            for (const chunk of visibleChunks) {
                this.renderer.renderChunk(chunk);
            }
        }
        
        this.renderer.renderGrid();
        
        for (const chunk of visibleChunks) {
            this.renderer.renderChunkEntities(chunk);
        }
        
        this.renderer.renderPlayer(this.playerWorldPos);
        
        if (this.debugMode) {
            this.renderer.renderDebug(this.fps, this.chunkManager.loadedChunks.size);
        }
        
        this.renderer.renderMiniMap(this.playerWorldPos);
    }

    forceRender() {
        this.render();
    }

    setPlayerPosition(worldPos) {
        this.playerWorldPos = { ...worldPos };
        this.viewport.setCenter(worldPos.x, worldPos.y);
    }

    getPlayerPosition() {
        return { ...this.playerWorldPos };
    }

    worldToScreen(worldPos) {
        return this.viewport.worldToScreen(worldPos);
    }

    screenToWorld(screenPos) {
        return this.viewport.screenToWorld(screenPos);
    }

    onChunkLoaded(event) {
        const chunk = event.detail.chunk;
        this.renderer.markDirty(chunk.pos);
    }

    loadInitialChunks() {
        this.chunkManager.preloadChunks(this.playerWorldPos);
    }

    setDebugMode(enabled) {
        this.debugMode = enabled;
    }

    setRenderMode(mode) {
        if (mode === '2d' || mode === 'pseudo3d') {
            this.config.renderMode = mode;
        }
    }
}

class ChunkManager {
    constructor(engine) {
        this.engine = engine;
        this.loadedChunks = new Map();
        this.pendingRequests = new Map();
        this.eventListeners = {};
    }

    addEventListener(event, callback) {
        if (!this.eventListeners[event]) {
            this.eventListeners[event] = [];
        }
        this.eventListeners[event].push(callback);
    }

    removeEventListener(event, callback) {
        if (!this.eventListeners[event]) return;
        this.eventListeners[event] = this.eventListeners[event].filter(cb => cb !== callback);
    }

    dispatchEvent(event, detail) {
        if (!this.eventListeners[event]) return;
        this.eventListeners[event].forEach(cb => cb({ detail }));
    }

    loadChunk(chunkPos) {
        const key = this.getChunkKey(chunkPos);
        
        if (this.loadedChunks.has(key)) {
            return Promise.resolve(this.loadedChunks.get(key));
        }
        
        if (this.pendingRequests.has(key)) {
            return this.pendingRequests.get(key);
        }

        const promise = new Promise((resolve) => {
            this.requestChunkData(chunkPos).then(chunkData => {
                const chunk = this.createChunk(chunkPos, chunkData);
                this.loadedChunks.set(key, chunk);
                this.pendingRequests.delete(key);
                this.dispatchEvent('chunkLoaded', { chunk });
                resolve(chunk);
            }).catch(() => {
                const chunk = this.createEmptyChunk(chunkPos);
                this.loadedChunks.set(key, chunk);
                this.pendingRequests.delete(key);
                resolve(chunk);
            });
        });

        this.pendingRequests.set(key, promise);
        return promise;
    }

    requestChunkData(chunkPos) {
        return new Promise((resolve, reject) => {
            fetch(`/api/map/chunk?x=${chunkPos.x}&y=${chunkPos.y}`)
                .then(res => {
                    if (!res.ok) throw new Error('Network response was not ok');
                    return res.json();
                })
                .then(data => {
                    const tiles = [];
                    const size = this.engine.config.chunkSize / this.engine.config.tileSize;
                    
                    if (data.tiles && data.tiles.length > 0) {
                        if (Array.isArray(data.tiles[0])) {
                            for (let y = 0; y < size; y++) {
                                for (let x = 0; x < size; x++) {
                                    if (data.tiles[y] && data.tiles[y][x]) {
                                        const tileData = data.tiles[y][x];
                                        tiles.push(typeof tileData === 'object' ? (tileData.Terrain || tileData.terrain || 0) : tileData);
                                    } else {
                                        tiles.push(0);
                                    }
                                }
                            }
                        } else {
                            for (let i = 0; i < size * size; i++) {
                                const tileData = data.tiles[i];
                                tiles.push(typeof tileData === 'object' ? (tileData.Terrain || tileData.terrain || 0) : (tileData || 0));
                            }
                        }
                    } else {
                        for (let y = 0; y < size; y++) {
                            for (let x = 0; x < size; x++) {
                                const worldX = chunkPos.x * this.engine.config.chunkSize + x * this.engine.config.tileSize;
                                const worldY = chunkPos.y * this.engine.config.chunkSize + y * this.engine.config.tileSize;
                                let tileId = this.generateTileId(worldX, worldY);
                                tiles.push(tileId);
                            }
                        }
                    }
                    
                    const entities = data.entities || [];
                    const processedEntities = entities.map(e => ({
                        id: e.EntityID || e.id || Math.random().toString(36).substr(2, 9),
                        type: e.EntityType || e.type || 1,
                        name: e.Name || e.name || 'Unknown',
                        x: e.Pos ? (e.Pos.X || e.Pos.x) : (e.pos_x || e.x || 0),
                        y: e.Pos ? (e.Pos.Y || e.Pos.y) : (e.pos_y || e.y || 0),
                        z: e.Pos ? (e.Pos.Z || e.Pos.z) : (e.pos_z || e.z || 0),
                        Health: e.Properties ? e.Properties.Health : (e.health || 100),
                        MaxHealth: e.Properties ? e.Properties.MaxHealth : (e.max_health || 100)
                    }));
                    
                    console.log(`[ChunkManager] Loaded chunk ${chunkPos.x},${chunkPos.y}: ${tiles.length} tiles, ${processedEntities.length} entities`);
                    resolve({ tiles, entities: processedEntities });
                }).catch(err => {
                    console.log(`[ChunkManager] Failed to load chunk ${chunkPos.x},${chunkPos.y}, using fallback:`, err.message);
                    reject(err);
                });
        });
    }

    generateTileId(x, y) {
        const noise = this.simplexNoise(x / 200, y / 200);
        const noise2 = this.simplexNoise(x / 150 + 100, y / 150 + 100);
        
        if (noise > 0.85) return 7;
        if (noise > 0.75) return 6;
        if (noise < 0.2) return 5;
        if (noise < 0.25) return 4;
        if (noise < 0.3) return 3;
        if (noise2 > 0.7) return 9;
        if (noise2 > 0.55) return 1;
        if (noise2 < 0.25) return 2;
        return 0;
    }

    simplexNoise(x, y) {
        const n = Math.sin(x * 0.05) * 0.5 + Math.cos(y * 0.07) * 0.3 + Math.sin((x + y) * 0.03) * 0.2;
        return (n + 1) / 2;
    }

    createChunk(chunkPos, data) {
        const size = this.engine.config.chunkSize / this.engine.config.tileSize;
        const tileMatrix = [];
        
        for (let y = 0; y < size; y++) {
            tileMatrix[y] = [];
            for (let x = 0; x < size; x++) {
                tileMatrix[y][x] = data.tiles[y * size + x];
            }
        }
        
        return {
            pos: chunkPos,
            tiles: tileMatrix,
            entities: data.entities || [],
            loaded: true,
            dirty: true
        };
    }

    createEmptyChunk(chunkPos) {
        const size = this.engine.config.chunkSize / this.engine.config.tileSize;
        const tileMatrix = [];
        
        for (let y = 0; y < size; y++) {
            tileMatrix[y] = [];
            for (let x = 0; x < size; x++) {
                const worldX = chunkPos.x * this.engine.config.chunkSize + x * this.engine.config.tileSize;
                const worldY = chunkPos.y * this.engine.config.chunkSize + y * this.engine.config.tileSize;
                let tileId = this.generateTileId(worldX, worldY);
                tileMatrix[y][x] = tileId;
            }
        }
        
        return {
            pos: chunkPos,
            tiles: tileMatrix,
            entities: [],
            loaded: true,
            dirty: true
        };
    }

    unloadChunk(chunkPos) {
        const key = this.getChunkKey(chunkPos);
        this.loadedChunks.delete(key);
        this.dispatchEvent('chunkUnloaded', { chunkPos });
    }

    getChunk(chunkPos) {
        const key = this.getChunkKey(chunkPos);
        return this.loadedChunks.get(key);
    }

    getVisibleChunks() {
        const centerChunk = this.worldToChunk(this.engine.playerWorldPos);
        const range = this.engine.config.viewRange;
        const visible = [];

        for (let dx = -range; dx <= range; dx++) {
            for (let dy = -range; dy <= range; dy++) {
                const chunkPos = { x: centerChunk.x + dx, y: centerChunk.y + dy };
                const chunk = this.loadedChunks.get(this.getChunkKey(chunkPos));
                if (chunk) {
                    visible.push(chunk);
                }
            }
        }

        return visible;
    }

    update(playerPos) {
        const centerChunk = this.worldToChunk(playerPos);
        const range = this.engine.config.viewRange;
        
        const neededChunks = new Set();
        for (let dx = -range; dx <= range; dx++) {
            for (let dy = -range; dy <= range; dy++) {
                neededChunks.add(this.getChunkKey({ x: centerChunk.x + dx, y: centerChunk.y + dy }));
            }
        }

        for (const key of neededChunks) {
            if (!this.loadedChunks.has(key) && !this.pendingRequests.has(key)) {
                const [x, y] = key.split(',').map(Number);
                this.loadChunk({ x, y });
            }
        }

        for (const [key, chunk] of this.loadedChunks) {
            if (!neededChunks.has(key)) {
                this.unloadChunk(chunk.pos);
            }
        }
    }

    preloadChunks(worldPos) {
        const centerChunk = this.worldToChunk(worldPos);
        const range = this.engine.config.viewRange;
        
        const chunkPromises = [];
        for (let dx = -range; dx <= range; dx++) {
            for (let dy = -range; dy <= range; dy++) {
                chunkPromises.push(this.loadChunk({ x: centerChunk.x + dx, y: centerChunk.y + dy }));
            }
        }
        
        Promise.all(chunkPromises).then(() => {
            console.log('[ChunkManager] All initial chunks loaded');
        }).catch(() => {
            console.log('[ChunkManager] Some chunks failed to load');
        });
    }

    worldToChunk(worldPos) {
        return {
            x: Math.floor(worldPos.x / this.engine.config.chunkSize),
            y: Math.floor(worldPos.y / this.engine.config.chunkSize)
        };
    }

    chunkToWorld(chunkPos) {
        return {
            x: chunkPos.x * this.engine.config.chunkSize,
            y: chunkPos.y * this.engine.config.chunkSize
        };
    }

    getChunkKey(chunkPos) {
        return `${chunkPos.x},${chunkPos.y}`;
    }
}

class Viewport {
    constructor(engine) {
        this.engine = engine;
        this.centerX = 0;
        this.centerY = 0;
        this.width = 0;
        this.height = 0;
        this.scale = 1;
        this.targetCenterX = 0;
        this.targetCenterY = 0;
        this.smoothFactor = 0.15;
    }

    resize(width, height) {
        this.width = width;
        this.height = height;
    }

    setCenter(x, y) {
        this.targetCenterX = x;
        this.targetCenterY = y;
    }

    update() {
        this.centerX += (this.targetCenterX - this.centerX) * this.smoothFactor;
        this.centerY += (this.targetCenterY - this.centerY) * this.smoothFactor;
    }

    worldToScreen(worldPos) {
        return {
            x: (worldPos.x - this.centerX) * this.scale + this.width / 2,
            y: (worldPos.y - this.centerY) * this.scale + this.height / 2
        };
    }

    screenToWorld(screenPos) {
        return {
            x: (screenPos.x - this.width / 2) / this.scale + this.centerX,
            y: (screenPos.y - this.height / 2) / this.scale + this.centerY
        };
    }

    getVisibleRect() {
        const halfWidth = (this.width / 2) / this.scale;
        const halfHeight = (this.height / 2) / this.scale;
        
        return {
            left: this.centerX - halfWidth,
            right: this.centerX + halfWidth,
            top: this.centerY - halfHeight,
            bottom: this.centerY + halfHeight
        };
    }

    getScale() {
        return this.scale;
    }
}

class MapRenderer {
    constructor(engine) {
        this.engine = engine;
        this.ctx = null;
        this.canvas = null;
        this.miniMapEnabled = true;
        this.miniMapSize = 150;
        this.miniMapMargin = 10;
        this.tileTypes = [
            { name: 'grass', color: '#3d7a37', icon: '🌿' },
            { name: 'forest', color: '#1e4d23', icon: '🌲' },
            { name: 'dirt', color: '#8b7355', icon: '🪨' },
            { name: 'sand', color: '#f4d03f', icon: '🏜️' },
            { name: 'water', color: '#3498db', icon: '🌊' },
            { name: 'deep_water', color: '#2c3e50', icon: '🌊' },
            { name: 'mountain', color: '#95a5a6', icon: '⛰️' },
            { name: 'snow', color: '#ecf0f1', icon: '❄️' },
            { name: 'lava', color: '#e74c3c', icon: '🔥' },
            { name: 'swamp', color: '#27ae60', icon: '🟩' }
        ];
        this.gridSpacing = 64;
    }

    init(canvas) {
        this.canvas = canvas;
        this.ctx = canvas.getContext('2d');
    }

    clear() {
        this.ctx.fillStyle = '#0d0d14';
        this.ctx.fillRect(0, 0, this.canvas.width, this.canvas.height);
    }

    renderChunk(chunk) {
        const chunkWorldPos = this.engine.chunkManager.chunkToWorld(chunk.pos);
        const screenPos = this.engine.viewport.worldToScreen(chunkWorldPos);
        const tileSize = this.engine.config.tileSize * this.engine.viewport.getScale();
        const chunkSize = this.engine.config.chunkSize;
        const tilesPerChunk = chunkSize / this.engine.config.tileSize;

        for (let ty = 0; ty < tilesPerChunk; ty++) {
            for (let tx = 0; tx < tilesPerChunk; tx++) {
                const tileId = chunk.tiles[ty][tx];
                const x = screenPos.x + tx * tileSize;
                const y = screenPos.y + ty * tileSize;

                this.renderTile(x, y, tileSize, tileId);
            }
        }
    }

    renderTile(x, y, size, tileId) {
        this.ctx.fillStyle = '#1a1a2e';
        this.ctx.fillRect(x, y, size, size);
        
        const tile = this.tileTypes[tileId];
        if (!tile) return;
        
        this.ctx.fillStyle = tile.color;
        this.ctx.font = `${size * 0.7}px serif`;
        this.ctx.textAlign = 'center';
        this.ctx.textBaseline = 'middle';
        this.ctx.fillText(tile.icon, x + size / 2, y + size / 2);
        
        this.ctx.strokeStyle = 'rgba(255, 255, 255, 0.05)';
        this.ctx.lineWidth = 0.5;
        this.ctx.strokeRect(x, y, size, size);
    }

    renderEmptyMap() {
        const scale = this.engine.viewport.getScale();
        const tileSize = this.engine.config.tileSize * scale;
        const visible = this.engine.viewport.getVisibleRect();
        
        const startTileX = Math.floor(visible.left / this.engine.config.tileSize);
        const endTileX = Math.ceil(visible.right / this.engine.config.tileSize);
        const startTileY = Math.floor(visible.top / this.engine.config.tileSize);
        const endTileY = Math.ceil(visible.bottom / this.engine.config.tileSize);

        for (let ty = startTileY; ty <= endTileY; ty++) {
            for (let tx = startTileX; tx <= endTileX; tx++) {
                const worldX = tx * this.engine.config.tileSize;
                const worldY = ty * this.engine.config.tileSize;
                const screenPos = this.engine.viewport.worldToScreen({ x: worldX, y: worldY });
                
                if (screenPos.x + tileSize < 0 || screenPos.x > this.canvas.width ||
                    screenPos.y + tileSize < 0 || screenPos.y > this.canvas.height) {
                    continue;
                }
                
                const noise = this.simplexNoise(worldX / 200, worldY / 200);
                const noise2 = this.simplexNoise(worldX / 150 + 100, worldY / 150 + 100);
                
                let tileId;
                if (noise > 0.85) tileId = 7;
                else if (noise > 0.75) tileId = 6;
                else if (noise < 0.2) tileId = 5;
                else if (noise < 0.25) tileId = 4;
                else if (noise < 0.3) tileId = 3;
                else if (noise2 > 0.7) tileId = 9;
                else if (noise2 > 0.55) tileId = 1;
                else if (noise2 < 0.25) tileId = 2;
                else tileId = 0;
                
                this.renderTile(screenPos.x, screenPos.y, tileSize, tileId);
            }
        }
    }

    simplexNoise(x, y) {
        const n = Math.sin(x * 0.05) * 0.5 + Math.cos(y * 0.07) * 0.3 + Math.sin((x + y) * 0.03) * 0.2;
        return (n + 1) / 2;
    }

    renderChunkEntities(chunk) {
        if (!chunk.entities || chunk.entities.length === 0) return;

        for (const entity of chunk.entities) {
            let entX = entity.x || entity.pos_x || entity.PosX;
            let entY = entity.y || entity.pos_y || entity.PosY;
            
            if (entity.Pos) {
                entX = entity.Pos.X || entity.Pos.x || entX;
                entY = entity.Pos.Y || entity.Pos.y || entY;
            }
            
            const screenPos = this.engine.viewport.worldToScreen({ x: entX, y: entY });
            
            let entityType = entity.type || entity.EntityType;
            if (typeof entityType === 'number') {
                switch(entityType) {
                    case 1: entityType = 'monster'; break;
                    case 2: entityType = 'npc'; break;
                    case 3: entityType = 'treasure'; break;
                    case 4: entityType = 'portal'; break;
                }
            }
            
            let color = '#00d4ff';
            let size = 16;
            let icon = '👤';
            
            if (entityType === 'npc') {
                color = '#f59e0b';
                icon = '🧙';
            } else if (entityType === 'monster') {
                color = '#ef4444';
                icon = '👹';
            } else if (entityType === 'treasure') {
                color = '#fbbf24';
                icon = '📦';
            } else if (entityType === 'portal') {
                color = '#a855f7';
                icon = '🌀';
            }

            this.ctx.fillStyle = color;
            this.ctx.font = `${size}px serif`;
            this.ctx.textAlign = 'center';
            this.ctx.textBaseline = 'middle';
            this.ctx.fillText(icon, screenPos.x, screenPos.y);

            this.ctx.shadowColor = color;
            this.ctx.shadowBlur = size * 2;
            this.ctx.fillText(icon, screenPos.x, screenPos.y);
            this.ctx.shadowBlur = 0;
        }
    }

    renderGrid() {
        const scale = this.engine.viewport.getScale();
        const gridSize = this.gridSpacing * scale;
        const visible = this.engine.viewport.getVisibleRect();
        
        const startX = Math.floor(visible.left / this.gridSpacing) * this.gridSpacing;
        const endX = Math.ceil(visible.right / this.gridSpacing) * this.gridSpacing;
        const startY = Math.floor(visible.top / this.gridSpacing) * this.gridSpacing;
        const endY = Math.ceil(visible.bottom / this.gridSpacing) * this.gridSpacing;

        this.ctx.strokeStyle = 'rgba(0, 212, 255, 0.15)';
        this.ctx.lineWidth = 1;

        for (let x = startX; x <= endX; x += this.gridSpacing) {
            const screenX = (x - this.engine.viewport.centerX) * scale + this.canvas.width / 2;
            this.ctx.beginPath();
            this.ctx.moveTo(screenX, 0);
            this.ctx.lineTo(screenX, this.canvas.height);
            this.ctx.stroke();
        }

        for (let y = startY; y <= endY; y += this.gridSpacing) {
            const screenY = (y - this.engine.viewport.centerY) * scale + this.canvas.height / 2;
            this.ctx.beginPath();
            this.ctx.moveTo(0, screenY);
            this.ctx.lineTo(this.canvas.width, screenY);
            this.ctx.stroke();
        }

        this.renderChunkBorders(startX, endX, startY, endY, scale);
    }

    renderChunkBorders(startX, endX, startY, endY, scale) {
        const chunkSize = this.engine.config.chunkSize;
        
        this.ctx.strokeStyle = 'rgba(0, 217, 255, 0.4)';
        this.ctx.lineWidth = 2;

        for (let x = startX; x <= endX; x += chunkSize) {
            const screenX = (x - this.engine.viewport.centerX) * scale + this.canvas.width / 2;
            this.ctx.beginPath();
            this.ctx.moveTo(screenX, 0);
            this.ctx.lineTo(screenX, this.canvas.height);
            this.ctx.stroke();
        }

        for (let y = startY; y <= endY; y += chunkSize) {
            const screenY = (y - this.engine.viewport.centerY) * scale + this.canvas.height / 2;
            this.ctx.beginPath();
            this.ctx.moveTo(0, screenY);
            this.ctx.lineTo(this.canvas.width, screenY);
            this.ctx.stroke();
        }
    }

    renderPlayer(worldPos) {
        const screenPos = { x: this.canvas.width / 2, y: this.canvas.height / 2 };
        
        this.ctx.save();
        
        this.ctx.shadowColor = '#22c55e';
        this.ctx.shadowBlur = 25;
        
        this.ctx.fillStyle = '#22c55e';
        this.ctx.font = '24px serif';
        this.ctx.textAlign = 'center';
        this.ctx.textBaseline = 'middle';
        this.ctx.fillText('🧑', screenPos.x, screenPos.y);
        
        this.ctx.restore();
    }

    renderDebug(fps, chunkCount) {
        this.ctx.save();
        this.ctx.font = '12px monospace';
        this.ctx.fillStyle = '#888';
        
        let y = 20;
        this.ctx.fillText(`FPS: ${fps}`, 10, y); y += 15;
        this.ctx.fillText(`Chunks: ${chunkCount}`, 10, y); y += 15;
        this.ctx.fillText(`Player: (${Math.floor(this.engine.playerWorldPos.x)}, ${Math.floor(this.engine.playerWorldPos.y)})`, 10, y); y += 15;
        this.ctx.fillText(`Viewport: (${Math.floor(this.engine.viewport.centerX)}, ${Math.floor(this.engine.viewport.centerY)})`, 10, y);
        
        this.ctx.restore();
    }

    renderMiniMap(playerPos) {
        if (!this.miniMapEnabled || !this.canvas || !this.ctx) return;
        
        const size = this.miniMapSize;
        const margin = this.miniMapMargin;
        const x = this.canvas.width - size - margin;
        const y = margin;
        
        this.ctx.save();
        
        this.ctx.fillStyle = 'rgba(0, 0, 0, 0.7)';
        this.ctx.fillRect(x, y, size, size);
        
        this.ctx.strokeStyle = '#fff';
        this.ctx.lineWidth = 1;
        this.ctx.strokeRect(x, y, size, size);
        
        const chunkSize = this.engine.config.chunkSize;
        const miniMapScale = size / (chunkSize * 3);
        
        for (let dx = -1; dx <= 1; dx++) {
            for (let dy = -1; dy <= 1; dy++) {
                const chunk = this.engine.chunkManager.getChunk({ x: Math.floor(playerPos.x / chunkSize) + dx, y: Math.floor(playerPos.y / chunkSize) + dy });
                const chunkWorldX = Math.floor(playerPos.x / chunkSize) * chunkSize + dx * chunkSize;
                const chunkWorldY = Math.floor(playerPos.y / chunkSize) * chunkSize + dy * chunkSize;
                
                const miniX = x + (chunkWorldX - playerPos.x + chunkSize * 1.5) * miniMapScale;
                const miniY = y + (chunkWorldY - playerPos.y + chunkSize * 1.5) * miniMapScale;
                const miniChunkSize = chunkSize * miniMapScale;
                
                if (chunk && chunk.loaded) {
                    this.ctx.fillStyle = 'rgba(45, 90, 39, 0.6)';
                } else {
                    this.ctx.fillStyle = 'rgba(30, 30, 30, 0.6)';
                }
                this.ctx.fillRect(miniX, miniY, miniChunkSize, miniChunkSize);
                
                this.ctx.strokeStyle = 'rgba(255, 255, 255, 0.2)';
                this.ctx.lineWidth = 0.5;
                this.ctx.strokeRect(miniX, miniY, miniChunkSize, miniChunkSize);
            }
        }
        
        const playerMiniX = x + size / 2;
        const playerMiniY = y + size / 2;
        
        this.ctx.fillStyle = '#22c55e';
        this.ctx.beginPath();
        this.ctx.arc(playerMiniX, playerMiniY, 4, 0, Math.PI * 2);
        this.ctx.fill();
        
        this.ctx.fillStyle = '#fff';
        this.ctx.font = '10px sans-serif';
        this.ctx.textAlign = 'center';
        this.ctx.fillText('小地图', x + size / 2, y + size + 15);
        
        this.ctx.restore();
    }

    markDirty(chunkPos) {
        const chunk = this.engine.chunkManager.getChunk(chunkPos);
        if (chunk) {
            chunk.dirty = true;
        }
    }
}

export { MapEngine, ChunkManager, Viewport, MapRenderer };