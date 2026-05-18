package main

import (
	"encoding/json"
	"net"
	"sync"

	"github.com/openworld-server/internal/cluster"
	"github.com/openworld-server/internal/log"
	"github.com/openworld-server/internal/msg"
	"github.com/openworld-server/internal/network"
	"github.com/openworld-server/internal/worldmap"
)

type GateHandler struct {
	cluster     *cluster.Cluster
	gridRouter  *worldmap.GridMapRouter
	playerGrids map[int64]int
	connPool    map[string]*connPoolEntry
	mu          sync.RWMutex
	poolMu      sync.RWMutex
}

type connPoolEntry struct {
	conns []net.Conn
	mu    sync.Mutex
}

func NewGateHandler(cluster *cluster.Cluster) *GateHandler {
	return &GateHandler{
		cluster:     cluster,
		gridRouter:  worldmap.NewGridMapRouter(cluster, 9),
		playerGrids: make(map[int64]int),
		connPool:    make(map[string]*connPoolEntry),
	}
}

func (h *GateHandler) getConnection(addr string) (net.Conn, error) {
	h.poolMu.RLock()
	entry, ok := h.connPool[addr]
	h.poolMu.RUnlock()

	if ok {
		entry.mu.Lock()
		if len(entry.conns) > 0 {
			conn := entry.conns[len(entry.conns)-1]
			entry.conns = entry.conns[:len(entry.conns)-1]
			entry.mu.Unlock()
			log.Info("Reusing connection from pool for ", addr)
			return conn, nil
		}
		entry.mu.Unlock()
	}

	log.Info("Creating new connection to ", addr)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	return conn, nil
}

func (h *GateHandler) releaseConnection(addr string, conn net.Conn) {
	h.poolMu.Lock()
	entry, ok := h.connPool[addr]
	if !ok {
		entry = &connPoolEntry{
			conns: make([]net.Conn, 0, 10),
		}
		h.connPool[addr] = entry
	}
	h.poolMu.Unlock()

	entry.mu.Lock()
	if len(entry.conns) < 10 {
		entry.conns = append(entry.conns, conn)
	} else {
		conn.Close()
	}
	entry.mu.Unlock()
}

func (h *GateHandler) Handle(msgObj *network.Message) {
	if msgObj.ID != msg.MSG_PING {
		log.Info("Gate received message from ", msgObj.Session.RemoteAddr(), " - MsgID:", msgObj.ID, "(", msg.GetMsgName(msgObj.ID), ")")
	}

	targetNodeType := msg.GetMessageNodeType(msgObj.ID)
	h.routeMessage(msgObj, targetNodeType)
}

func (h *GateHandler) routeMessage(msgObj *network.Message, targetNodeType msg.NodeType) {
	switch targetNodeType {
	case msg.NodeTypeGate:
		return
	case msg.NodeTypeGridMap:
		h.handleGridMapMessage(msgObj)
	default:
		h.forwardToService(msgObj, targetNodeType)
	}
}

func (h *GateHandler) handleGridMapMessage(msgObj *network.Message) {
	switch msgObj.ID {
	case msg.MSG_MAP_PLAYER_ENTER:
		h.handleMapPlayerEnter(msgObj)
	case msg.MSG_MAP_PLAYER_MOVE:
		h.handleMapPlayerMove(msgObj)
	case msg.MSG_MAP_PLAYER_LEAVE:
		h.handleMapPlayerLeave(msgObj)
	default:
		h.forwardToRandomGridMap(msgObj)
	}
}

func (h *GateHandler) handleMapPlayerEnter(msgObj *network.Message) {
	var req msg.MapPlayerEnterRequest
	if err := json.Unmarshal(msgObj.Data, &req); err != nil {
		log.Error("Failed to unmarshal MapPlayerEnterRequest: ", err)
		return
	}

	gridID := h.gridRouter.GetGridMapID(worldmap.Vec3{X: req.PosX, Y: req.PosY, Z: req.PosZ})

	h.mu.Lock()
	h.playerGrids[req.PlayerID] = gridID
	h.mu.Unlock()

	h.forwardToGridMap(gridID, msgObj)
}

func (h *GateHandler) handleMapPlayerMove(msgObj *network.Message) {
	var req msg.MapPlayerMoveRequest
	if err := json.Unmarshal(msgObj.Data, &req); err != nil {
		log.Error("Failed to unmarshal MapPlayerMoveRequest: ", err)
		return
	}

	h.mu.RLock()
	currentGridID, ok := h.playerGrids[req.PlayerID]
	h.mu.RUnlock()

	newGridID := h.gridRouter.GetGridMapID(worldmap.Vec3{X: req.PosX, Y: req.PosY, Z: req.PosZ})

	if !ok {
		currentGridID = newGridID
		h.mu.Lock()
		h.playerGrids[req.PlayerID] = currentGridID
		h.mu.Unlock()
	}

	if newGridID != currentGridID {
		log.Info("Player ", req.PlayerID, " crossing grid boundary: ", currentGridID, " -> ", newGridID)

		h.mu.Lock()
		h.playerGrids[req.PlayerID] = newGridID
		h.mu.Unlock()
	}

	h.forwardToGridMap(newGridID, msgObj)
}

func (h *GateHandler) handleMapPlayerLeave(msgObj *network.Message) {
	var req msg.MapPlayerLeaveRequest
	if err := json.Unmarshal(msgObj.Data, &req); err != nil {
		log.Error("Failed to unmarshal MapPlayerLeaveRequest: ", err)
		return
	}

	h.mu.RLock()
	gridID, ok := h.playerGrids[req.PlayerID]
	h.mu.RUnlock()

	if ok {
		h.forwardToGridMap(gridID, msgObj)

		h.mu.Lock()
		delete(h.playerGrids, req.PlayerID)
		h.mu.Unlock()
	}
}

func (h *GateHandler) forwardToRandomGridMap(msgObj *network.Message) {
	services, err := h.cluster.DiscoverServices("gridmap-0")
	if err != nil || len(services) == 0 {
		log.Error("Failed to find gridmap service: ", err)
		return
	}
	h.sendToService(services[0].Addr, msgObj)
}

func (h *GateHandler) forwardToService(msgObj *network.Message, nodeType msg.NodeType) {
	serviceName := h.getServiceNameByNodeType(nodeType)
	if serviceName == "" {
		log.Error("Unknown node type:", nodeType, "(", nodeType.String(), ")")
		return
	}

	log.Info("Forwarding message to ", serviceName, " - MsgID:", msgObj.ID, "(", msg.GetMsgName(msgObj.ID), ")")
	service, err := h.cluster.GetRandomService(serviceName)
	if err != nil {
		log.Error("Failed to get service ", serviceName, ": ", err)
		return
	}

	h.sendToService(service.Addr, msgObj)
}

func (h *GateHandler) forwardToGridMap(gridID int, msgObj *network.Message) {
	serviceName := "gridmap-" + string(rune('0'+gridID))
	services, err := h.cluster.DiscoverServices(serviceName)
	if err != nil || len(services) == 0 {
		log.Error("Failed to find gridmap service: ", serviceName, err)
		return
	}

	h.sendToService(services[0].Addr, msgObj)
}

func (h *GateHandler) sendToService(addr string, msgObj *network.Message) {
	log.Info("Connecting to service at ", addr)
	conn, err := h.getConnection(addr)
	if err != nil {
		log.Error("Failed to connect to service at ", addr, ": ", err)
		return
	}
	defer h.releaseConnection(addr, conn)

	err = network.SendRawMessage(conn, msgObj.ID, msgObj.NodeType, msgObj.Data)
	if err != nil {
		log.Error("Failed to send message to service: ", err)
		return
	}

	buf := make([]byte, 4096)
	n, err := conn.Read(buf)
	if err != nil {
		log.Error("Failed to read response from service: ", err)
		return
	}

	respMsgID, respNodeType, respData, err := network.ParseMessage(buf[:n])
	if err != nil {
		log.Error("Failed to parse response: ", err)
		return
	}

	log.Info("Received response from service - MsgID:", respMsgID, "(", msg.GetMsgName(respMsgID), "), NodeType:", respNodeType, "(", respNodeType.String(), ")")
	err = network.SendRawMessage(msgObj.Session, respMsgID, respNodeType, respData)
	if err != nil {
		log.Error("Failed to send response to client: ", err)
	}
}

func (h *GateHandler) getServiceNameByNodeType(nodeType msg.NodeType) string {
	switch nodeType {
	case msg.NodeTypeLogin:
		return "login"
	case msg.NodeTypeLogic:
		return "logic"
	case msg.NodeTypeBattle:
		return "battle"
	case msg.NodeTypeGridMap:
		return "gridmap"
	case msg.NodeTypeCross:
		return "cross"
	case msg.NodeTypeData:
		return "dataservice"
	case msg.NodeTypeGM:
		return "gm"
	default:
		return ""
	}
}
