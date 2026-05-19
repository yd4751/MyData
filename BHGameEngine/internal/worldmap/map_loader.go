package worldmap

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"sync"
	"time"

	"github.com/openworld-server/pkg/snowflake"
	"github.com/sirupsen/logrus"
)

// Perlin noise constants
const (
	noiseScale     = 0.01
	octaves        = 4
	persistence    = 0.5
	lacunarity     = 2.0
	seedMultiplier = 1000000
)

// noise2D generates 2D Perlin-like noise
func noise2D(x, y, seed float64) float64 {
	var total float64
	var amplitude float64 = 1
	var frequency float64 = 1
	var maxValue float64 = 0

	for i := 0; i < octaves; i++ {
		total += interpolatedNoise(x*frequency+seed, y*frequency+seed) * amplitude
		maxValue += amplitude
		amplitude *= persistence
		frequency *= lacunarity
	}

	return total / maxValue
}

func interpolatedNoise(x, y float64) float64 {
	intX := int(math.Floor(x))
	intY := int(math.Floor(y))
	fracX := x - float64(intX)
	fracY := y - float64(intY)

	v1 := smoothNoise(intX, intY)
	v2 := smoothNoise(intX+1, intY)
	v3 := smoothNoise(intX, intY+1)
	v4 := smoothNoise(intX+1, intY+1)

	i1 := interpolate(v1, v2, fracX)
	i2 := interpolate(v3, v4, fracX)

	return interpolate(i1, i2, fracY)
}

func smoothNoise(x, y int) float64 {
	corners := (noise(x-1, y-1) + noise(x+1, y-1) + noise(x-1, y+1) + noise(x+1, y+1)) / 16
	sides := (noise(x-1, y) + noise(x+1, y) + noise(x, y-1) + noise(x, y+1)) / 8
	center := noise(x, y) / 4
	return corners + sides + center
}

func noise(x, y int) float64 {
	n := int64(x)*9301 + int64(y)*49297 + 49297
	n = (n << 13) ^ n
	return 1.0 - float64((n*(n*n*15731+789221)+1376312589)&0x7fffffff)/1073741824.0
}

func interpolate(a, b, t float64) float64 {
	ft := t * math.Pi
	f := (1.0 - math.Cos(ft)) * 0.5
	return a*(1.0-f) + b*f
}

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

	if chunkData.Entities == nil || len(chunkData.Entities) == 0 {
		l.generateEntities(chunkData)
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
		return l.GenerateDefaultChunk(chunkPos), nil
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

func (l *MapLoader) GenerateDefaultChunk(chunkPos ChunkPos) *MapChunkData {
	chunk := &MapChunkData{
		ChunkPos:  chunkPos,
		Version:   1,
		Timestamp: time.Now().Unix(),
	}

	seed := float64(chunkPos.X*seedMultiplier + chunkPos.Y)

	for x := 0; x < ChunkSize; x++ {
		for y := 0; y < ChunkSize; y++ {
			worldX := float64(chunkPos.X*ChunkSize + x)
			worldY := float64(chunkPos.Y*ChunkSize + y)

			noiseVal := noise2D(worldX*noiseScale, worldY*noiseScale, seed)

			var terrainType int32
			var height float64
			var walkable bool

			if noiseVal < -0.2 {
				terrainType = 4
				height = 0
				walkable = false
			} else if noiseVal < -0.05 {
				terrainType = 3
				height = 0
				walkable = true
			} else if noiseVal < 0.1 {
				terrainType = 0
				height = 0
				walkable = true
			} else if noiseVal < 0.3 {
				terrainType = 1
				height = noiseVal * 10
				walkable = true
			} else if noiseVal < 0.5 {
				terrainType = 1
				height = noiseVal * 15
				walkable = true
			} else if noiseVal < 0.7 {
				terrainType = 2
				height = noiseVal * 20
				walkable = true
			} else {
				terrainType = 5
				height = noiseVal * 25
				walkable = false
			}

			chunk.Tiles[x][y] = MapTile{
				TileID:     int32(chunkPos.X*ChunkSize + x),
				Terrain:    terrainType,
				Height:     height,
				Walkable:   walkable,
				WaterLevel: 0,
			}
		}
	}

	l.generateEntities(chunk)

	return chunk
}

func (l *MapLoader) generateEntities(chunk *MapChunkData) {
	rand.Seed(time.Now().UnixNano() + int64(chunk.ChunkPos.X)*1000 + int64(chunk.ChunkPos.Y))

	entities := make([]MapEntityData, 0)

	if rand.Float64() > 0.7 {
		entities = append(entities, MapEntityData{
			EntityID:   snowflake.GenerateID(),
			EntityType: 2,
			Name:       "守卫",
			Pos: Vec3{
				X: float64(chunk.ChunkPos.X*ChunkSize) + 128 + rand.Float64()*64,
				Y: float64(chunk.ChunkPos.Y*ChunkSize) + 128 + rand.Float64()*64,
				Z: 0,
			},
			Rotation: 0,
		})
	}

	if rand.Float64() > 0.3 {
		entities = append(entities, MapEntityData{
			EntityID:   snowflake.GenerateID(),
			EntityType: 1,
			Name:       "史莱姆",
			Pos: Vec3{
				X: float64(chunk.ChunkPos.X*ChunkSize) + 64 + rand.Float64()*128,
				Y: float64(chunk.ChunkPos.Y*ChunkSize) + 64 + rand.Float64()*128,
				Z: 0,
			},
			Rotation: 0,
		})
	}

	if rand.Float64() > 0.5 {
		entities = append(entities, MapEntityData{
			EntityID:   snowflake.GenerateID(),
			EntityType: 1,
			Name:       "哥布林",
			Pos: Vec3{
				X: float64(chunk.ChunkPos.X*ChunkSize) + 100 + rand.Float64()*56,
				Y: float64(chunk.ChunkPos.Y*ChunkSize) + 100 + rand.Float64()*56,
				Z: 0,
			},
			Rotation: 0,
		})
	}

	if rand.Float64() > 0.6 {
		monsterNames := []string{"蝙蝠", "骷髅", "狼"}
		name := monsterNames[rand.Intn(len(monsterNames))]
		entities = append(entities, MapEntityData{
			EntityID:   snowflake.GenerateID(),
			EntityType: 1,
			Name:       name,
			Pos: Vec3{
				X: float64(chunk.ChunkPos.X*ChunkSize) + 80 + rand.Float64()*96,
				Y: float64(chunk.ChunkPos.Y*ChunkSize) + 80 + rand.Float64()*96,
				Z: 0,
			},
			Rotation: 0,
		})
	}

	chunk.Entities = entities
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
