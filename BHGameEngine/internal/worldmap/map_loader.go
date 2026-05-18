package worldmap

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/sirupsen/logrus"
)

type MapTile struct {
	TileID     int32   // 瓦片ID
	Terrain    int32   // 地形类型
	Height     float64 // 高度
	Walkable   bool    // 是否可行走
	WaterLevel float64 // 水位高度
}

type MapChunkData struct {
	ChunkPos  ChunkPos                      // 区块位置
	Tiles     [ChunkSize][ChunkSize]MapTile // 瓦片数据
	Entities  []MapEntityData               // 实体数据
	Timestamp int64                         // 时间戳
	Version   int32                         // 版本号
}

type MapEntityData struct {
	EntityID   int64
	EntityType int32
	Name       string
	Pos        Vec3
	Rotation   float64
	Properties map[string]interface{}
}

type MapLoader struct {
	mapDataDir   string                     // 地图数据目录
	loadedChunks map[ChunkPos]*MapChunkData // 已加载的区块
	mu           sync.RWMutex               // 读写锁
}

func NewMapLoader(mapDataDir string) *MapLoader {
	return &MapLoader{
		mapDataDir:   mapDataDir,
		loadedChunks: make(map[ChunkPos]*MapChunkData),
	}
}

func (l *MapLoader) LoadChunk(chunkPos ChunkPos) (*MapChunkData, error) {
	l.mu.RLock()
	if chunk, ok := l.loadedChunks[chunkPos]; ok {
		l.mu.RUnlock()
		return chunk, nil
	}
	l.mu.RUnlock()

	chunkData, err := l.loadChunkFromFile(chunkPos)
	if err != nil {
		return nil, err
	}

	l.mu.Lock()
	l.loadedChunks[chunkPos] = chunkData
	l.mu.Unlock()

	logrus.Info("Map chunk loaded: ", chunkPos.X, ",", chunkPos.Y)
	return chunkData, nil
}

func (l *MapLoader) loadChunkFromFile(chunkPos ChunkPos) (*MapChunkData, error) {
	filePath := l.getChunkFilePath(chunkPos)

	data, err := os.ReadFile(filePath)
	if err != nil {
		logrus.Warn("Failed to load chunk file, generating default: ", err)
		return l.generateDefaultChunk(chunkPos), nil
	}

	var chunkData MapChunkData
	err = json.Unmarshal(data, &chunkData)
	if err != nil {
		logrus.Error("Failed to parse chunk data: ", err)
		return nil, err
	}

	return &chunkData, nil
}

func (l *MapLoader) getChunkFilePath(chunkPos ChunkPos) string {
	return l.mapDataDir + "/chunk_" + fmt.Sprintf("%d_%d", chunkPos.X, chunkPos.Y) + ".json"
}

func (l *MapLoader) generateDefaultChunk(chunkPos ChunkPos) *MapChunkData {
	chunk := &MapChunkData{
		ChunkPos: chunkPos,
		Version:  1,
	}

	for x := 0; x < ChunkSize; x++ {
		for y := 0; y < ChunkSize; y++ {
			chunk.Tiles[x][y] = MapTile{
				TileID:   int32(chunkPos.X*ChunkSize + x),
				Terrain:  1,
				Height:   0,
				Walkable: true,
			}
		}
	}

	return chunk
}

func (l *MapLoader) UnloadChunk(chunkPos ChunkPos) {
	l.mu.Lock()
	delete(l.loadedChunks, chunkPos)
	l.mu.Unlock()
	logrus.Info("Map chunk unloaded: ", chunkPos.X, ",", chunkPos.Y)
}

func (l *MapLoader) IsChunkLoaded(chunkPos ChunkPos) bool {
	l.mu.RLock()
	defer l.mu.RUnlock()
	_, ok := l.loadedChunks[chunkPos]
	return ok
}

func (l *MapLoader) GetLoadedChunkCount() int {
	l.mu.RLock()
	defer l.mu.RUnlock()
	return len(l.loadedChunks)
}

func (l *MapLoader) SaveChunk(chunkData *MapChunkData) error {
	filePath := l.getChunkFilePath(chunkData.ChunkPos)

	data, err := json.Marshal(chunkData)
	if err != nil {
		return err
	}

	err = os.WriteFile(filePath, data, 0644)
	if err != nil {
		return err
	}

	logrus.Info("Map chunk saved: ", chunkData.ChunkPos.X, ",", chunkData.ChunkPos.Y)
	return nil
}

func (l *MapLoader) GetTileAtWorldPos(pos Vec3) (*MapTile, error) {
	chunkX := int(pos.X / float64(ChunkSize))
	chunkY := int(pos.Y / float64(ChunkSize))
	tileX := int(pos.X) % ChunkSize
	tileY := int(pos.Y) % ChunkSize

	chunk, err := l.LoadChunk(ChunkPos{X: chunkX, Y: chunkY})
	if err != nil {
		return nil, err
	}

	return &chunk.Tiles[tileX][tileY], nil
}

func (l *MapLoader) IsWalkable(pos Vec3) bool {
	tile, err := l.GetTileAtWorldPos(pos)
	if err != nil {
		return false
	}
	return tile.Walkable
}
