package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/openworld-server/internal/worldmap"
	"github.com/openworld-server/pkg/config"
	"github.com/openworld-server/pkg/logger"
)

var configPath = flag.String("config", "./config/config.toml", "config file path")
var mapDataDir = flag.String("mapdir", "./data/maps", "map data directory")
var action = flag.String("action", "", "action: generate, info, edit, entity, save")
var chunkX = flag.Int("chunkx", 0, "chunk X coordinate")
var chunkY = flag.Int("chunky", 0, "chunk Y coordinate")
var tileX = flag.Int("tilex", 0, "tile X coordinate")
var tileY = flag.Int("tiley", 0, "tile Y coordinate")
var tileID = flag.Int("tileid", 1, "tile ID")
var terrain = flag.Int("terrain", 1, "terrain type")
var walkable = flag.Bool("walkable", true, "walkable flag")
var height = flag.Float64("height", 0, "height value")
var entityID = flag.Int64("entityid", 0, "entity ID")
var entityType = flag.Int("entitytype", 1, "entity type")
var entityName = flag.String("entityname", "", "entity name")
var posX = flag.Float64("posx", 0, "position X")
var posY = flag.Float64("posy", 0, "position Y")
var posZ = flag.Float64("posz", 0, "position Z")
var radius = flag.Int("radius", 1, "radius for generation")

func main() {
	flag.Parse()

	err := config.LoadConfig(*configPath)
	if err != nil {
		logger.Fatal("Failed to load config:", err)
	}

	loggerConfig := config.GetLoggerConfig()
	logger.Init(loggerConfig.Path, loggerConfig.Level, "maptool")

	if *action == "" {
		printUsage()
		return
	}

	mapLoader := worldmap.NewMapLoader(*mapDataDir)

	switch *action {
	case "generate":
		generateChunks(mapLoader, *chunkX, *chunkY, *radius)
	case "info":
		showChunkInfo(mapLoader, *chunkX, *chunkY)
	case "edit":
		editTile(mapLoader, *chunkX, *chunkY, *tileX, *tileY, *tileID, *terrain, *walkable, *height)
	case "entity":
		addEntity(mapLoader, *chunkX, *chunkY, *entityID, *entityType, *entityName, *posX, *posY, *posZ)
	case "save":
		saveChunk(mapLoader, *chunkX, *chunkY)
	case "list":
		listChunks(mapLoader)
	default:
		fmt.Println("Unknown action:", *action)
		printUsage()
	}
}

func printUsage() {
	fmt.Println("Map Tool - BHGameEngine Map Editor")
	fmt.Println("Usage:")
	fmt.Println("  maptool -action=<action> [options]")
	fmt.Println("")
	fmt.Println("Actions:")
	fmt.Println("  generate   - Generate map chunks")
	fmt.Println("  info       - Show chunk information")
	fmt.Println("  edit       - Edit tile properties")
	fmt.Println("  entity     - Add entity to chunk")
	fmt.Println("  save       - Save chunk to file")
	fmt.Println("  list       - List all loaded chunks")
	fmt.Println("")
	fmt.Println("Options:")
	fmt.Println("  -config=<path>    Config file path (default: ./config/config.toml)")
	fmt.Println("  -mapdir=<path>    Map data directory (default: ./data/maps)")
	fmt.Println("  -chunkx=<int>     Chunk X coordinate")
	fmt.Println("  -chunky=<int>     Chunk Y coordinate")
	fmt.Println("  -radius=<int>     Radius for generation (default: 1)")
	fmt.Println("  -tilex=<int>      Tile X coordinate (0-255)")
	fmt.Println("  -tiley=<int>      Tile Y coordinate (0-255)")
	fmt.Println("  -tileid=<int>     Tile ID")
	fmt.Println("  -terrain=<int>    Terrain type")
	fmt.Println("  -walkable=<bool>  Walkable flag")
	fmt.Println("  -height=<float>   Height value")
	fmt.Println("  -entityid=<int>   Entity ID")
	fmt.Println("  -entitytype=<int> Entity type")
	fmt.Println("  -entityname=<str> Entity name")
	fmt.Println("  -posx=<float>     Position X")
	fmt.Println("  -posy=<float>     Position Y")
	fmt.Println("  -posz=<float>     Position Z")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  maptool -action=generate -chunkx=0 -chunky=0 -radius=3")
	fmt.Println("  maptool -action=info -chunkx=0 -chunky=0")
	fmt.Println("  maptool -action=edit -chunkx=0 -chunky=0 -tilex=10 -tiley=10 -tileid=5 -terrain=2 -walkable=true")
	fmt.Println("  maptool -action=entity -chunkx=0 -chunky=0 -entityid=1001 -entitytype=1 -entityname=Monster -posx=128 -posy=128")
}

func generateChunks(loader *worldmap.MapLoader, centerX, centerY, radius int) {
	fmt.Printf("Generating chunks around (%d,%d) with radius %d...\n", centerX, centerY, radius)

	for dx := -radius; dx <= radius; dx++ {
		for dy := -radius; dy <= radius; dy++ {
			chunkPos := worldmap.ChunkPos{X: centerX + dx, Y: centerY + dy}
			chunk := loader.GenerateDefaultChunk(chunkPos)
			err := loader.SaveChunk(chunk)
			if err != nil {
				fmt.Printf("Failed to save chunk (%d,%d): %v\n", chunkPos.X, chunkPos.Y, err)
			} else {
				fmt.Printf("Generated chunk (%d,%d)\n", chunkPos.X, chunkPos.Y)
			}
		}
	}

	fmt.Println("Generation complete!")
}

func showChunkInfo(loader *worldmap.MapLoader, x, y int) {
	chunkPos := worldmap.ChunkPos{X: x, Y: y}
	chunk, err := loader.LoadChunk(chunkPos)
	if err != nil {
		fmt.Printf("Failed to load chunk (%d,%d): %v\n", x, y, err)
		return
	}

	fmt.Printf("Chunk (%d,%d) Information:\n", chunkPos.X, chunkPos.Y)
	fmt.Printf("  Version: %d\n", chunk.Version)
	fmt.Printf("  Timestamp: %d\n", chunk.Timestamp)
	fmt.Printf("  Entity Count: %d\n", len(chunk.Entities))

	tileStats := make(map[int32]int)
	for _, row := range chunk.Tiles {
		for _, tile := range row {
			tileStats[tile.Terrain]++
		}
	}

	fmt.Printf("  Terrain Distribution:\n")
	for terrain, count := range tileStats {
		fmt.Printf("    Type %d: %d tiles\n", terrain, count)
	}
}

func editTile(loader *worldmap.MapLoader, chunkX, chunkY, tileX, tileY int, tileID, terrain int, walkable bool, height float64) {
	if tileX < 0 || tileX >= worldmap.ChunkSize || tileY < 0 || tileY >= worldmap.ChunkSize {
		fmt.Println("Tile coordinates must be between 0 and", worldmap.ChunkSize-1)
		return
	}

	chunkPos := worldmap.ChunkPos{X: chunkX, Y: chunkY}
	chunk, err := loader.LoadChunk(chunkPos)
	if err != nil {
		fmt.Printf("Failed to load chunk (%d,%d): %v\n", chunkX, chunkY, err)
		return
	}

	chunk.Tiles[tileX][tileY] = worldmap.MapTile{
		TileID:     int32(tileID),
		Terrain:    int32(terrain),
		Height:     height,
		Walkable:   walkable,
		WaterLevel: 0,
	}

	err = loader.SaveChunk(chunk)
	if err != nil {
		fmt.Printf("Failed to save chunk: %v\n", err)
	} else {
		fmt.Printf("Edited tile (%d,%d) in chunk (%d,%d)\n", tileX, tileY, chunkX, chunkY)
	}
}

func addEntity(loader *worldmap.MapLoader, chunkX, chunkY int, entityID int64, entityType int, name string, posX, posY, posZ float64) {
	chunkPos := worldmap.ChunkPos{X: chunkX, Y: chunkY}
	chunk, err := loader.LoadChunk(chunkPos)
	if err != nil {
		fmt.Printf("Failed to load chunk (%d,%d): %v\n", chunkX, chunkY, err)
		return
	}

	entity := worldmap.MapEntityData{
		EntityID:   entityID,
		EntityType: int32(entityType),
		Name:       name,
		Pos:        worldmap.Vec3{X: posX, Y: posY, Z: posZ},
		Rotation:   0,
		Properties: make(map[string]interface{}),
	}

	chunk.Entities = append(chunk.Entities, entity)

	err = loader.SaveChunk(chunk)
	if err != nil {
		fmt.Printf("Failed to save chunk: %v\n", err)
	} else {
		fmt.Printf("Added entity %d (%s) to chunk (%d,%d) at (%.1f,%.1f,%.1f)\n", entityID, name, chunkX, chunkY, posX, posY, posZ)
	}
}

func saveChunk(loader *worldmap.MapLoader, x, y int) {
	chunkPos := worldmap.ChunkPos{X: x, Y: y}
	chunk, err := loader.LoadChunk(chunkPos)
	if err != nil {
		fmt.Printf("Failed to load chunk (%d,%d): %v\n", x, y, err)
		return
	}

	err = loader.SaveChunk(chunk)
	if err != nil {
		fmt.Printf("Failed to save chunk (%d,%d): %v\n", x, y, err)
	} else {
		fmt.Printf("Saved chunk (%d,%d)\n", x, y)
	}
}

func listChunks(loader *worldmap.MapLoader) {
	files, err := os.ReadDir(*mapDataDir)
	if err != nil {
		fmt.Printf("Failed to read map directory: %v\n", err)
		return
	}

	fmt.Println("Loaded chunks:")
	for _, file := range files {
		if !file.IsDir() {
			fmt.Println("  ", file.Name())
		}
	}

	fmt.Printf("Total: %d chunks\n", len(files))
}
