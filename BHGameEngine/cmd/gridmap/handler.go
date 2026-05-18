package main

import (
	"encoding/json"
	"net"
	"sync"
	"time"

	"github.com/openworld-server/internal/ai"
	"github.com/openworld-server/internal/cluster"
	"github.com/openworld-server/internal/connector"
	"github.com/openworld-server/internal/entity"
	"github.com/openworld-server/internal/msg"
	"github.com/openworld-server/internal/network"
	"github.com/openworld-server/internal/player"
	"github.com/openworld-server/internal/worldmap"
	"github.com/openworld-server/pkg/logger"
)

type GridMapHandler struct {
	cluster     *cluster.Cluster
	worldMap    *worldmap.WorldMap
	entityMgr   *entity.EntityManager
	playerMgr   *player.PlayerManager
	aiManager   *ai.AIManager
	gridRouter  *worldmap.GridMapRouter
	mapLoader   *worldmap.MapLoader
	gridID      int
	playerConns map[int64]net.Conn
	connector   *connector.Connector
	mu          sync.RWMutex
}

func NewGridMapHandler(cluster *cluster.Cluster, gridID int) *GridMapHandler {
	return &GridMapHandler{
		cluster:     cluster,
		worldMap:    worldmap.NewWorldMap(256, 3),
		entityMgr:   entity.NewEntityManager(),
		playerMgr:   player.NewPlayerManager(),
		aiManager:   ai.NewAIManager(),
		gridRouter:  worldmap.NewGridMapRouter(cluster, 9),
		mapLoader:   worldmap.NewMapLoader("./data/maps"),
		gridID:      gridID,
		playerConns: make(map[int64]net.Conn),
		connector:   connector.NewConnector(cluster),
	}
}

func (h *GridMapHandler) Handle(msgObj *network.Message) {
	switch msgObj.ID {
	case msg.MSG_MAP_PLAYER_ENTER:
		h.handlePlayerEnter(msgObj)
	case msg.MSG_MAP_PLAYER_LEAVE:
		h.handlePlayerLeave(msgObj)
	case msg.MSG_MAP_PLAYER_MOVE:
		h.handlePlayerMove(msgObj)
	case msg.MSG_MAP_CROSS_GRID_REQ:
		h.handleCrossGrid(msgObj)
	case msg.MSG_MAP_CHUNK_LOAD_REQ:
		h.handleChunkLoad(msgObj)
	case msg.MSG_MAP_ENTITY_SYNC:
		h.handleEntitySync(msgObj)
	default:
		logger.Warn("Unknown message ID: ", msgObj.ID)
	}
}

func (h *GridMapHandler) handlePlayerEnter(msgObj *network.Message) {
	var req msg.MapPlayerEnterRequest
	err := json.Unmarshal(msgObj.Data, &req)
	if err != nil {
		logger.Error("Failed to unmarshal MapPlayerEnterRequest: ", err)
		return
	}

	player := h.playerMgr.CreatePlayer(req.PlayerID, req.Name, 0)
	player.Pos = worldmap.Vec3{X: req.PosX, Y: req.PosY, Z: req.PosZ}
	player.Rotation = req.Rotation
	player.Level = req.Level
	player.Health = req.Health
	player.MaxHealth = req.MaxHealth

	h.worldMap.UpdatePlayerPosition(req.PlayerID, player.Pos)

	h.mu.Lock()
	h.playerConns[req.PlayerID] = msgObj.Session
	h.mu.Unlock()

	logger.Info("Player entered gridmap ", h.gridID, ": ", req.PlayerID, " at (", req.PosX, ",", req.PosY, ")")

	resp := msg.MapPlayerEnterResponse{
		Result:  0,
		Message: "success",
	}
	h.sendResponse(msgObj.Session, msg.MSG_MAP_PLAYER_ENTER, &resp)
}

func (h *GridMapHandler) handlePlayerLeave(msgObj *network.Message) {
	var req msg.MapPlayerLeaveRequest
	err := json.Unmarshal(msgObj.Data, &req)
	if err != nil {
		logger.Error("Failed to unmarshal MapPlayerLeaveRequest: ", err)
		return
	}

	h.worldMap.UpdatePlayerPosition(req.PlayerID, worldmap.Vec3{X: -1, Y: -1, Z: 0})
	h.playerMgr.RemovePlayer(req.PlayerID)

	h.mu.Lock()
	delete(h.playerConns, req.PlayerID)
	h.mu.Unlock()

	logger.Info("Player left gridmap ", h.gridID, ": ", req.PlayerID)

	resp := msg.MapPlayerLeaveResponse{
		Result:  0,
		Message: "success",
	}
	h.sendResponse(msgObj.Session, msg.MSG_MAP_PLAYER_LEAVE, &resp)
}

func (h *GridMapHandler) handlePlayerMove(msgObj *network.Message) {
	var req msg.MapPlayerMoveRequest
	err := json.Unmarshal(msgObj.Data, &req)
	if err != nil {
		logger.Error("Failed to unmarshal MapPlayerMoveRequest: ", err)
		return
	}

	newPos := worldmap.Vec3{X: req.PosX, Y: req.PosY, Z: req.PosZ}

	player, ok := h.playerMgr.GetPlayer(req.PlayerID)
	if !ok {
		return
	}

	if h.gridRouter.IsCrossingBoundary(player.Pos, newPos) {
		h.handleCrossGridInternal(req.PlayerID, newPos)
		return
	}

	h.playerMgr.UpdatePosition(req.PlayerID, newPos)
	h.worldMap.UpdatePlayerPosition(req.PlayerID, newPos)

	h.syncNearbyPlayers(req.PlayerID, newPos)

	resp := msg.MapPlayerMoveResponse{
		Result:  0,
		Message: "success",
	}
	h.sendResponse(msgObj.Session, msg.MSG_MAP_PLAYER_MOVE, &resp)
}

func (h *GridMapHandler) handleCrossGridInternal(playerID int64, newPos worldmap.Vec3) {
	player, ok := h.playerMgr.GetPlayer(playerID)
	if !ok {
		return
	}

	oldGridID := h.gridID
	newGridID := h.gridRouter.GetGridMapID(newPos)

	if newGridID == oldGridID {
		return
	}

	logger.Info("Player ", playerID, " crossing from grid ", oldGridID, " to grid ", newGridID)

	oldPos := player.GetPosition()

	h.worldMap.UpdatePlayerPosition(playerID, worldmap.Vec3{X: -1, Y: -1, Z: 0})

	req := msg.MapCrossGridRequest{
		PlayerID:   playerID,
		FromGridID: oldGridID,
		ToGridID:   newGridID,
		PosX:       newPos.X,
		PosY:       newPos.Y,
		PosZ:       newPos.Z,
		Name:       player.Name,
		Level:      player.GetLevel(),
		Health:     player.Health,
		MaxHealth:  player.GetMaxHealth(),
		Rotation:   player.Rotation,
	}

	service, err := h.gridRouter.GetGridMapByID(newGridID)
	if err != nil || service == nil {
		logger.Error("Failed to find target gridmap service: ", err)
		h.handleCrossGridFailure(playerID, oldPos)
		return
	}

	nodeID := "gridmap:" + service.Addr
	conn, err := h.connector.GetConnectionByNodeID(nodeID)
	if err != nil {
		logger.Error("Failed to get connection to target gridmap: ", err)
		h.handleCrossGridFailure(playerID, oldPos)
		return
	}

	data, _ := json.Marshal(req)
	err = conn.Send(msg.MSG_MAP_CROSS_GRID_REQ, msg.NodeTypeGridMap, data)
	if err != nil {
		logger.Error("Failed to send cross grid request: ", err)
		h.handleCrossGridFailure(playerID, oldPos)
		return
	}

	_, _, respData, err := conn.Receive(5 * time.Second)
	if err != nil {
		logger.Error("Failed to read cross grid response: ", err)
		h.handleCrossGridFailure(playerID, oldPos)
		return
	}

	var resp msg.MapCrossGridResponse
	err = json.Unmarshal(respData, &resp)
	if err != nil {
		logger.Error("Failed to unmarshal cross grid response: ", err)
		h.handleCrossGridFailure(playerID, oldPos)
		return
	}

	if resp.Result == 0 {
		h.playerMgr.RemovePlayer(playerID)
		h.mu.Lock()
		delete(h.playerConns, playerID)
		h.mu.Unlock()
		logger.Info("Player ", playerID, " successfully transferred to grid ", newGridID)
	} else {
		logger.Error("Cross grid failed: ", resp.Message)
		h.handleCrossGridFailure(playerID, oldPos)
	}
}

func (h *GridMapHandler) handleCrossGridFailure(playerID int64, oldPos worldmap.Vec3) {
	logger.Info("Rolling back player ", playerID, " to position (", oldPos.X, ",", oldPos.Y, ")")

	h.worldMap.UpdatePlayerPosition(playerID, oldPos)

	player, ok := h.playerMgr.GetPlayer(playerID)
	if ok {
		player.Pos = oldPos
	}

	h.syncNearbyPlayers(playerID, oldPos)
}

func (h *GridMapHandler) handleCrossGrid(msgObj *network.Message) {
	var req msg.MapCrossGridRequest
	err := json.Unmarshal(msgObj.Data, &req)
	if err != nil {
		logger.Error("Failed to unmarshal MapCrossGridRequest: ", err)
		return
	}

	player := h.playerMgr.CreatePlayer(req.PlayerID, req.Name, 0)
	player.Pos = worldmap.Vec3{X: req.PosX, Y: req.PosY, Z: req.PosZ}
	player.Rotation = req.Rotation
	player.Level = req.Level
	player.Health = req.Health
	player.MaxHealth = req.MaxHealth

	h.worldMap.UpdatePlayerPosition(req.PlayerID, player.Pos)

	h.mu.Lock()
	h.playerConns[req.PlayerID] = msgObj.Session
	h.mu.Unlock()

	logger.Info("Player ", req.PlayerID, " entered from grid ", req.FromGridID)

	resp := msg.MapCrossGridResponse{
		Result:     0,
		Message:    "success",
		TargetGrid: "gridmap",
	}
	h.sendResponse(msgObj.Session, msg.MSG_MAP_CROSS_GRID_RES, &resp)
}

func (h *GridMapHandler) handleChunkLoad(msgObj *network.Message) {
	var req msg.MapChunkLoadRequest
	err := json.Unmarshal(msgObj.Data, &req)
	if err != nil {
		logger.Error("Failed to unmarshal MapChunkLoadRequest: ", err)
		return
	}

	chunkPos := worldmap.ChunkPos{X: int(req.ChunkX), Y: int(req.ChunkY)}
	chunkData, err := h.mapLoader.LoadChunk(chunkPos)
	if err != nil {
		logger.Error("Failed to load chunk: ", err)
		resp := msg.MapChunkLoadResponse{
			Result:  1,
			Message: "failed",
		}
		h.sendResponse(msgObj.Session, msg.MSG_MAP_CHUNK_LOAD_RES, &resp)
		return
	}

	entities := make([]*msg.EntityInfo, 0)
	for _, entityData := range chunkData.Entities {
		entities = append(entities, &msg.EntityInfo{
			EntityID:   entityData.EntityID,
			EntityType: entityData.EntityType,
			Name:       entityData.Name,
			PosX:       entityData.Pos.X,
			PosY:       entityData.Pos.Y,
			PosZ:       entityData.Pos.Z,
			Rotation:   entityData.Rotation,
		})
	}

	tiles := make([]int32, 0)
	for x := 0; x < worldmap.ChunkSize; x++ {
		for y := 0; y < worldmap.ChunkSize; y++ {
			tiles = append(tiles, chunkData.Tiles[x][y].TileID)
		}
	}

	resp := msg.MapChunkLoadResponse{
		Result:   0,
		Message:  "success",
		ChunkX:   req.ChunkX,
		ChunkY:   req.ChunkY,
		Tiles:    tiles,
		Entities: entities,
	}
	h.sendResponse(msgObj.Session, msg.MSG_MAP_CHUNK_LOAD_RES, &resp)
}

func (h *GridMapHandler) handleEntitySync(msgObj *network.Message) {
	var req msg.MapEntitySyncRequest
	err := json.Unmarshal(msgObj.Data, &req)
	if err != nil {
		logger.Error("Failed to unmarshal MapEntitySyncRequest: ", err)
		return
	}

	for _, entityInfo := range req.Entities {
		chunkPos := h.worldMap.WorldPosToChunk(worldmap.Vec3{X: entityInfo.PosX, Y: entityInfo.PosY, Z: entityInfo.PosZ})
		h.entityMgr.CreateMonster(entityInfo.EntityID, entityInfo.Name,
			worldmap.Vec3{X: entityInfo.PosX, Y: entityInfo.PosY, Z: entityInfo.PosZ},
			1)
		h.worldMap.AddEntity(chunkPos, entityInfo.EntityID, nil)
	}
}

func (h *GridMapHandler) syncNearbyPlayers(playerID int64, pos worldmap.Vec3) {
	nearbyPlayers := h.worldMap.GetNearbyPlayers(pos, 200)

	if len(nearbyPlayers) == 0 {
		return
	}

	playerSyncs := make([]*msg.PlayerSync, 0)
	for _, nearby := range nearbyPlayers {
		if nearby.PlayerID == playerID {
			continue
		}
		p, ok := h.playerMgr.GetPlayer(nearby.PlayerID)
		if ok {
			playerSyncs = append(playerSyncs, &msg.PlayerSync{
				PlayerID:  p.ID,
				Name:      p.Name,
				PosX:      p.Pos.X,
				PosY:      p.Pos.Y,
				PosZ:      p.Pos.Z,
				Rotation:  p.Rotation,
				State:     int32(p.State),
				Health:    p.Health,
				MaxHealth: p.MaxHealth,
				Level:     p.Level,
			})
		}
	}

	if len(playerSyncs) > 0 {
		req := msg.MapPlayerSyncRequest{
			PlayerID: playerID,
			Players:  playerSyncs,
		}
		data, _ := json.Marshal(req)

		h.mu.RLock()
		conn, ok := h.playerConns[playerID]
		h.mu.RUnlock()

		if ok {
			network.SendRawMessage(conn, msg.MSG_MAP_PLAYER_SYNC, msg.NodeTypeGridMap, data)
		}
	}
}

func (h *GridMapHandler) sendResponse(conn net.Conn, msgID uint32, data interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		logger.Error("Failed to marshal response: ", err)
		return
	}
	network.SendRawMessage(conn, msgID, msg.NodeTypeGridMap, jsonData)
}

func (h *GridMapHandler) Start() {
	h.aiManager.Start()

	go func() {
		ticker := time.NewTicker(100 * time.Millisecond)
		for range ticker.C {
			h.updateAI()
		}
	}()

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		for range ticker.C {
			h.worldMap.UnloadInactiveChunks(30 * time.Second)
		}
	}()
}

func (h *GridMapHandler) Stop() {
	h.aiManager.Stop()
}

func (h *GridMapHandler) updateAI() {
	monsters := h.entityMgr.GetAllMonsters()
	for _, monster := range monsters {
		if monster.AIEnabled {
			pos := monster.GetPosition()
			nearbyPlayers := h.worldMap.GetNearbyPlayers(pos, 50)

			if len(nearbyPlayers) > 0 {
				h.aiManager.SetTarget(monster.ID, nearbyPlayers[0].PlayerID, nearbyPlayers[0].Pos)
			} else {
				h.aiManager.ClearTarget(monster.ID)
			}
		}
	}
}
