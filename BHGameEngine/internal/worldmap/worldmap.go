package worldmap

import (
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	ChunkSize = 256 // 区块大小
	ViewRange = 3   // 视野范围（区块数量）
)

// ChunkPos 区块坐标
type ChunkPos struct {
	X, Y int // 区块坐标
}

// Chunk 地图区块
type Chunk struct {
	Pos      ChunkPos                 // 区块坐标
	Entities map[int64]interface{}    // 区块内实体
	Players  map[int64]*PlayerInChunk // 区块内玩家
	Loaded   bool                     // 是否已加载
	Active   bool                     // 是否活跃
	mu       sync.RWMutex             // 并发锁
}

// PlayerInChunk 区块内玩家信息
type PlayerInChunk struct {
	PlayerID int64 // 玩家ID
	Pos      Vec3  // 玩家位置
}

// Vec3 三维坐标
type Vec3 struct {
	X, Y, Z float64 // 三维坐标值
}

// WorldMap 世界地图
type WorldMap struct {
	chunks       map[ChunkPos]*Chunk // 区块列表
	chunkSize    float64             // 区块大小
	viewRange    int                 // 视野范围
	playerChunks map[int64]ChunkPos  // 玩家所在区块（玩家ID->区块坐标）
	mu           sync.RWMutex        // 并发锁
}

func NewWorldMap(chunkSize float64, viewRange int) *WorldMap {
	return &WorldMap{
		chunks:       make(map[ChunkPos]*Chunk),
		chunkSize:    chunkSize,
		viewRange:    viewRange,
		playerChunks: make(map[int64]ChunkPos),
	}
}

func (w *WorldMap) GetChunk(pos ChunkPos) *Chunk {
	w.mu.RLock()
	chunk, ok := w.chunks[pos]
	w.mu.RUnlock()

	if !ok {
		chunk = w.createChunk(pos)
	}

	return chunk
}

func (w *WorldMap) createChunk(pos ChunkPos) *Chunk {
	w.mu.Lock()
	defer w.mu.Unlock()

	if chunk, ok := w.chunks[pos]; ok {
		return chunk
	}

	chunk := &Chunk{
		Pos:      pos,
		Entities: make(map[int64]interface{}),
		Players:  make(map[int64]*PlayerInChunk),
		Loaded:   true,
		Active:   true,
	}

	w.chunks[pos] = chunk
	logrus.Info("Chunk created: ", pos.X, ",", pos.Y)
	return chunk
}

func (w *WorldMap) GetPlayerViewChunks(playerID int64, pos Vec3) []*Chunk {
	centerChunk := w.worldPosToChunk(pos)
	chunks := make([]*Chunk, 0)

	for dx := -w.viewRange; dx <= w.viewRange; dx++ {
		for dy := -w.viewRange; dy <= w.viewRange; dy++ {
			chunkPos := ChunkPos{centerChunk.X + dx, centerChunk.Y + dy}
			chunk := w.GetChunk(chunkPos)
			chunks = append(chunks, chunk)
		}
	}

	return chunks
}

func (w *WorldMap) UpdatePlayerPosition(playerID int64, newPos Vec3) ([]*Chunk, []*Chunk) {
	w.mu.Lock()
	defer w.mu.Unlock()

	oldChunk, oldOk := w.playerChunks[playerID]
	newChunk := w.worldPosToChunk(newPos)

	if oldOk && oldChunk == newChunk {
		return nil, nil
	}

	oldView := w.getViewChunks(oldChunk)
	newView := w.getViewChunks(newChunk)

	entering := w.diffChunks(newView, oldView)
	leaving := w.diffChunks(oldView, newView)

	for _, chunk := range entering {
		chunk.mu.Lock()
		chunk.Players[playerID] = &PlayerInChunk{PlayerID: playerID, Pos: newPos}
		chunk.mu.Unlock()
	}

	for _, chunk := range leaving {
		chunk.mu.Lock()
		delete(chunk.Players, playerID)
		chunk.mu.Unlock()
	}

	w.playerChunks[playerID] = newChunk

	return entering, leaving
}

func (w *WorldMap) getViewChunks(center ChunkPos) []*Chunk {
	chunks := make([]*Chunk, 0)
	for dx := -w.viewRange; dx <= w.viewRange; dx++ {
		for dy := -w.viewRange; dy <= w.viewRange; dy++ {
			if chunk, ok := w.chunks[ChunkPos{center.X + dx, center.Y + dy}]; ok {
				chunks = append(chunks, chunk)
			}
		}
	}
	return chunks
}

func (w *WorldMap) diffChunks(a, b []*Chunk) []*Chunk {
	set := make(map[ChunkPos]bool)
	for _, chunk := range b {
		set[chunk.Pos] = true
	}

	result := make([]*Chunk, 0)
	for _, chunk := range a {
		if !set[chunk.Pos] {
			result = append(result, chunk)
		}
	}
	return result
}

func (w *WorldMap) worldPosToChunk(pos Vec3) ChunkPos {
	return ChunkPos{
		X: int(math.Floor(pos.X / w.chunkSize)),
		Y: int(math.Floor(pos.Y / w.chunkSize)),
	}
}

func (w *WorldMap) chunkToWorldPos(chunk ChunkPos) Vec3 {
	return Vec3{
		X: float64(chunk.X)*w.chunkSize + w.chunkSize/2,
		Y: float64(chunk.Y)*w.chunkSize + w.chunkSize/2,
		Z: 0,
	}
}

func (w *WorldMap) WorldPosToChunk(pos Vec3) ChunkPos {
	return w.worldPosToChunk(pos)
}

func (w *WorldMap) GetNearbyPlayers(pos Vec3, radius float64) []*PlayerInChunk {
	centerChunk := w.worldPosToChunk(pos)
	players := make([]*PlayerInChunk, 0)

	for dx := -w.viewRange; dx <= w.viewRange; dx++ {
		for dy := -w.viewRange; dy <= w.viewRange; dy++ {
			chunkPos := ChunkPos{centerChunk.X + dx, centerChunk.Y + dy}
			w.mu.RLock()
			chunk, ok := w.chunks[chunkPos]
			w.mu.RUnlock()

			if ok {
				chunk.mu.RLock()
				for _, player := range chunk.Players {
					dx := player.Pos.X - pos.X
					dy := player.Pos.Y - pos.Y
					dz := player.Pos.Z - pos.Z
					dist := math.Sqrt(dx*dx + dy*dy + dz*dz)
					if dist <= radius {
						players = append(players, player)
					}
				}
				chunk.mu.RUnlock()
			}
		}
	}

	return players
}

func (w *WorldMap) AddEntity(chunkPos ChunkPos, entityID int64, entity interface{}) {
	chunk := w.GetChunk(chunkPos)
	chunk.mu.Lock()
	chunk.Entities[entityID] = entity
	chunk.mu.Unlock()
}

func (w *WorldMap) RemoveEntity(chunkPos ChunkPos, entityID int64) {
	chunk := w.GetChunk(chunkPos)
	chunk.mu.Lock()
	delete(chunk.Entities, entityID)
	chunk.mu.Unlock()
}

func (w *WorldMap) GetEntitiesInView(playerID int64) []interface{} {
	w.mu.RLock()
	chunkPos, ok := w.playerChunks[playerID]
	w.mu.RUnlock()

	if !ok {
		return nil
	}

	entities := make([]interface{}, 0)
	for dx := -w.viewRange; dx <= w.viewRange; dx++ {
		for dy := -w.viewRange; dy <= w.viewRange; dy++ {
			if chunk, ok := w.chunks[ChunkPos{chunkPos.X + dx, chunkPos.Y + dy}]; ok {
				chunk.mu.RLock()
				for _, entity := range chunk.Entities {
					entities = append(entities, entity)
				}
				chunk.mu.RUnlock()
			}
		}
	}

	return entities
}

func (w *WorldMap) UnloadInactiveChunks(timeout time.Duration) {
	w.mu.Lock()
	defer w.mu.Unlock()

	for pos, chunk := range w.chunks {
		if len(chunk.Players) == 0 {
			chunk.mu.Lock()
			chunk.Active = false
			chunk.Loaded = false
			chunk.mu.Unlock()
			delete(w.chunks, pos)
			logrus.Info("Chunk unloaded: ", pos.X, ",", pos.Y)
		}
	}
}

func (w *WorldMap) GetChunkLoadStatus(pos ChunkPos) (bool, bool) {
	w.mu.RLock()
	chunk, ok := w.chunks[pos]
	w.mu.RUnlock()

	if !ok {
		return false, false
	}

	return chunk.Loaded, chunk.Active
}

func (w *WorldMap) PreloadChunks(targetPos Vec3, radius int) error {
	targetChunk := w.worldPosToChunk(targetPos)

	for dx := -radius; dx <= radius; dx++ {
		for dy := -radius; dy <= radius; dy++ {
			chunkPos := ChunkPos{targetChunk.X + dx, targetChunk.Y + dy}
			w.GetChunk(chunkPos)
		}
	}

	logrus.Info("Preloaded chunks around: ", targetPos.X, ",", targetPos.Y)
	return nil
}

func (w *WorldMap) GetChunkCount() int {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return len(w.chunks)
}

func (w *WorldMap) GetPlayerCountInChunk(chunkPos ChunkPos) int {
	chunk := w.GetChunk(chunkPos)
	chunk.mu.RLock()
	defer chunk.mu.RUnlock()
	return len(chunk.Players)
}

func (c *Chunk) GetNearbyPlayers(pos Vec3, radius float64) []*PlayerInChunk {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]*PlayerInChunk, 0)
	for _, player := range c.Players {
		dx := player.Pos.X - pos.X
		dy := player.Pos.Y - pos.Y
		dz := player.Pos.Z - pos.Z
		dist := math.Sqrt(dx*dx + dy*dy + dz*dz)
		if dist <= radius {
			result = append(result, player)
		}
	}
	return result
}

func (c *Chunk) IsActive() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Active
}

func (c *Chunk) GetEntityCount() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.Entities)
}

func (c *Chunk) GetPlayerIDs() []int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	ids := make([]int64, 0, len(c.Players))
	for id := range c.Players {
		ids = append(ids, id)
	}
	return ids
}

func (pos ChunkPos) String() string {
	return fmt.Sprintf("(%d,%d)", pos.X, pos.Y)
}
