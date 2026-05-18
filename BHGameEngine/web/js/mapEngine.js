class MapEngine {
    constructor(config) {
        this.config = Object.assign({
            chunkSize: 256,
            viewRange: 3,
            tileSize: 32,
            mapWidth: 100,
            mapHeight: 100,
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
        
        this.playerWorldPos = { x: 0, y: 0, z: 0 };
        this.entities = [];
        this.loadingChunks = new Set();
        
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
        
        for (const chunk of visibleChunks) {
            this.renderer.renderChunk(chunk);
        }
        
        this.renderer.renderGrid();
        this.renderer.renderEntities(this.entities);
        
        this.renderer.renderPlayer(this.playerWorldPos);
        
        if (this.debugMode) {
            this.renderer.renderDebug(this.fps, this.chunkManager.loadedChunks.size);
        }
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
        return new Promise((resolve) => {
            setTimeout(() => {
                const tiles = [];
                const size = this.engine.config.chunkSize / this.engine.config.tileSize;
                for (let y = 0; y < size; y++) {
                    for (let x = 0; x < size; x++) {
                        const worldX = chunkPos.x * this.engine.config.chunkSize + x * this.engine.config.tileSize;
                        const worldY = chunkPos.y * this.engine.config.chunkSize + y * this.engine.config.tileSize;
                        let tileId = this.generateTileId(worldX, worldY);
                        tiles.push(tileId);
                    }
                }
                
                const entities = this.generateEntities(chunkPos);
                
                resolve({ tiles, entities });
            }, Math.random() * 100 + 50);
        });
    }

    generateTileId(x, y) {
        const noise = this.simplexNoise(x / 200, y / 200);
        
        if (noise > 0.7) return 2;
        if (noise > 0.5) return 1;
        if (noise > 0.3) return 3;
        if (noise > 0.1) return 0;
        return 4;
    }

    simplexNoise(x, y) {
        const n = Math.sin(x * 0.05) * 0.5 + Math.cos(y * 0.07) * 0.3 + Math.sin((x + y) * 0.03) * 0.2;
        return (n + 1) / 2;
    }

    generateEntities(chunkPos) {
        const entities = [];
        const rand = Math.random();
        
        if (rand > 0.7) {
            entities.push({
                id: Math.random().toString(36).substr(2, 9),
                type: 'npc',
                name: 'NPC',
                x: chunkPos.x * this.engine.config.chunkSize + 128,
                y: chunkPos.y * this.engine.config.chunkSize + 128,
                z: 0
            });
        }
        
        if (rand > 0.85) {
            entities.push({
                id: Math.random().toString(36).substr(2, 9),
                type: 'monster',
                name: 'Monster',
                x: chunkPos.x * this.engine.config.chunkSize + 64 + Math.random() * 128,
                y: chunkPos.y * this.engine.config.chunkSize + 64 + Math.random() * 128,
                z: 0
            });
        }
        
        return entities;
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
            entities: data.entities,
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
                tileMatrix[y][x] = 0;
            }
        }
        
        return {
            pos: chunkPos,
            tiles: tileMatrix,
            entities: [],
            loaded: true,
            dirty: false
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
        
        for (let dx = -range; dx <= range; dx++) {
            for (let dy = -range; dy <= range; dy++) {
                this.loadChunk({ x: centerChunk.x + dx, y: centerChunk.y + dy });
            }
        }
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
        this.smoothFactor = 0.1;
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

    setScale(scale) {
        this.scale = Math.max(0.5, Math.min(3, scale));
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
        this.dirtyChunks = new Set();
        this.tileColors = [
            '#2d5a27',
            '#3d7a37',
            '#4a90d9',
            '#8b7355',
            '#c4a35a'
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

        if (chunk.dirty) {
            for (let ty = 0; ty < tilesPerChunk; ty++) {
                for (let tx = 0; tx < tilesPerChunk; tx++) {
                    const tileId = chunk.tiles[ty][tx];
                    const x = screenPos.x + tx * tileSize;
                    const y = screenPos.y + ty * tileSize;

                    this.ctx.fillStyle = this.tileColors[tileId] || '#333';
                    this.ctx.fillRect(x, y, tileSize, tileSize);

                    this.ctx.strokeStyle = 'rgba(255, 255, 255, 0.05)';
                    this.ctx.strokeRect(x, y, tileSize, tileSize);
                }
            }
            chunk.dirty = false;
        }
    }

    renderEntities(entities) {
        for (const entity of entities) {
            const screenPos = this.engine.viewport.worldToScreen({ x: entity.x, y: entity.y });
            
            let color = '#00d4ff';
            let size = 16;
            
            if (entity.type === 'npc') {
                color = '#f59e0b';
                size = 14;
            } else if (entity.type === 'monster') {
                color = '#ef4444';
                size = 12;
            }

            this.ctx.fillStyle = color;
            this.ctx.beginPath();
            this.ctx.arc(screenPos.x, screenPos.y, size, 0, Math.PI * 2);
            this.ctx.fill();

            this.ctx.shadowColor = color;
            this.ctx.shadowBlur = size * 2;
            this.ctx.beginPath();
            this.ctx.arc(screenPos.x, screenPos.y, size, 0, Math.PI * 2);
            this.ctx.fill();
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

        this.renderGridLabels(startX, endX, startY, endY, scale);
    }

    renderGridLabels(startX, endX, startY, endY, scale) {
        this.ctx.font = `${10 * scale}px monospace`;
        this.ctx.fillStyle = 'rgba(0, 212, 255, 0.6)';
        this.ctx.textAlign = 'center';
        this.ctx.textBaseline = 'top';

        const edgeOffset = 15 * scale;
        const bottomY = this.canvas.height - edgeOffset;

        for (let x = startX; x <= endX; x += this.gridSpacing) {
            const screenX = (x - this.engine.viewport.centerX) * scale + this.canvas.width / 2;
            
            if (screenX > 30 * scale && screenX < this.canvas.width - 30 * scale) {
                this.ctx.fillText(`${Math.floor(x)}`, screenX, bottomY);
            }
        }

        this.ctx.textAlign = 'left';
        this.ctx.textBaseline = 'middle';

        const rightX = this.canvas.width - edgeOffset;

        for (let y = startY; y <= endY; y += this.gridSpacing) {
            const screenY = (y - this.engine.viewport.centerY) * scale + this.canvas.height / 2;
            
            if (screenY > 20 * scale && screenY < this.canvas.height - 20 * scale) {
                this.ctx.fillText(`${Math.floor(y)}`, rightX, screenY);
            }
        }
    }

    renderPlayer(worldPos) {
        if (!this.canvas || !this.ctx) return;
        
        const screenPos = { x: this.canvas.width / 2, y: this.canvas.height / 2 };
        
        this.ctx.save();
        
        this.ctx.shadowColor = '#22c55e';
        this.ctx.shadowBlur = 25;
        
        this.ctx.fillStyle = '#22c55e';
        this.ctx.beginPath();
        this.ctx.arc(screenPos.x, screenPos.y, 18, 0, Math.PI * 2);
        this.ctx.fill();
        
        this.ctx.fillStyle = '#86efac';
        this.ctx.beginPath();
        this.ctx.arc(screenPos.x, screenPos.y, 10, 0, Math.PI * 2);
        this.ctx.fill();
        
        this.ctx.fillStyle = '#ffffff';
        this.ctx.beginPath();
        this.ctx.arc(screenPos.x, screenPos.y, 5, 0, Math.PI * 2);
        this.ctx.fill();
        
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

    markDirty(chunkPos) {
        const chunk = this.engine.chunkManager.getChunk(chunkPos);
        if (chunk) {
            chunk.dirty = true;
        }
    }
}

export { MapEngine, ChunkManager, Viewport, MapRenderer };