package msg

import (
	"encoding/json"
	"sync"
)

type MessageRouter struct {
	msgToNodeType map[uint32]NodeType
	sync.RWMutex
}

var router = &MessageRouter{
	msgToNodeType: make(map[uint32]NodeType),
}

func RegisterMessageRoute(msgID uint32, nodeType NodeType) {
	router.Lock()
	defer router.Unlock()
	router.msgToNodeType[msgID] = nodeType
}

func GetMessageNodeType(msgID uint32) NodeType {
	router.RLock()
	defer router.RUnlock()
	if nodeType, ok := router.msgToNodeType[msgID]; ok {
		return nodeType
	}
	return NodeTypeGate
}

func init() {
	RegisterMessageRoute(MSG_LOGIN_REQ, NodeTypeLogin)
	RegisterMessageRoute(MSG_LOGIN_RES, NodeTypeLogin)
	RegisterMessageRoute(MSG_REGISTER_REQ, NodeTypeLogin)
	RegisterMessageRoute(MSG_REGISTER_RES, NodeTypeLogin)

	RegisterMessageRoute(MSG_LOGOUT_REQ, NodeTypeLogic)
	RegisterMessageRoute(MSG_LOGOUT_RES, NodeTypeLogic)
	RegisterMessageRoute(MSG_PLAYER_INFO_REQ, NodeTypeLogic)
	RegisterMessageRoute(MSG_PLAYER_INFO_RES, NodeTypeLogic)
	RegisterMessageRoute(MSG_PLAYER_MOVE_REQ, NodeTypeLogic)
	RegisterMessageRoute(MSG_PLAYER_MOVE_RES, NodeTypeLogic)
	RegisterMessageRoute(MSG_ATTACK_REQ, NodeTypeLogic)
	RegisterMessageRoute(MSG_ATTACK_RES, NodeTypeLogic)
	RegisterMessageRoute(MSG_SKILL_REQ, NodeTypeLogic)
	RegisterMessageRoute(MSG_SKILL_RES, NodeTypeLogic)
	RegisterMessageRoute(MSG_INVENTORY_REQ, NodeTypeLogic)
	RegisterMessageRoute(MSG_INVENTORY_RES, NodeTypeLogic)
	RegisterMessageRoute(MSG_ITEM_USE_REQ, NodeTypeLogic)
	RegisterMessageRoute(MSG_ITEM_USE_RES, NodeTypeLogic)
	RegisterMessageRoute(MSG_TASK_LIST_REQ, NodeTypeLogic)
	RegisterMessageRoute(MSG_TASK_LIST_RES, NodeTypeLogic)
	RegisterMessageRoute(MSG_TASK_ACCEPT_REQ, NodeTypeLogic)
	RegisterMessageRoute(MSG_TASK_ACCEPT_RES, NodeTypeLogic)
	RegisterMessageRoute(MSG_TASK_FINISH_REQ, NodeTypeLogic)
	RegisterMessageRoute(MSG_TASK_FINISH_RES, NodeTypeLogic)

	RegisterMessageRoute(MSG_MAP_LOAD_REQ, NodeTypeGridMap)
	RegisterMessageRoute(MSG_MAP_LOAD_RES, NodeTypeGridMap)
	RegisterMessageRoute(MSG_MAP_ENTITY_REQ, NodeTypeGridMap)
	RegisterMessageRoute(MSG_MAP_ENTITY_RES, NodeTypeGridMap)
	RegisterMessageRoute(MSG_MAP_PLAYER_ENTER, NodeTypeGridMap)
	RegisterMessageRoute(MSG_MAP_PLAYER_LEAVE, NodeTypeGridMap)
	RegisterMessageRoute(MSG_MAP_PLAYER_MOVE, NodeTypeGridMap)
	RegisterMessageRoute(MSG_MAP_PLAYER_SYNC, NodeTypeGridMap)
	RegisterMessageRoute(MSG_MAP_ENTITY_SYNC, NodeTypeGridMap)
	RegisterMessageRoute(MSG_MAP_CROSS_GRID_REQ, NodeTypeGridMap)
	RegisterMessageRoute(MSG_MAP_CROSS_GRID_RES, NodeTypeGridMap)
	RegisterMessageRoute(MSG_MAP_CHUNK_LOAD_REQ, NodeTypeGridMap)
	RegisterMessageRoute(MSG_MAP_CHUNK_LOAD_RES, NodeTypeGridMap)

	RegisterMessageRoute(MSG_PING, NodeTypeGate)
	RegisterMessageRoute(MSG_PONG, NodeTypeGate)
}

type NodeType uint32

const (
	NodeTypeGate    NodeType = 0
	NodeTypeLogin   NodeType = 1
	NodeTypeLogic   NodeType = 2
	NodeTypeBattle  NodeType = 3
	NodeTypeGridMap NodeType = 4
	NodeTypeCross   NodeType = 5
	NodeTypeData    NodeType = 6
	NodeTypeGM      NodeType = 7
)

func (n NodeType) String() string {
	switch n {
	case NodeTypeGate:
		return "Gate"
	case NodeTypeLogin:
		return "Login"
	case NodeTypeLogic:
		return "Logic"
	case NodeTypeBattle:
		return "Battle"
	case NodeTypeGridMap:
		return "GridMap"
	case NodeTypeCross:
		return "Cross"
	case NodeTypeData:
		return "Data"
	case NodeTypeGM:
		return "GM"
	default:
		return "Unknown"
	}
}

const (
	MSG_LOGIN_REQ    uint32 = 1001
	MSG_LOGIN_RES    uint32 = 1002
	MSG_REGISTER_REQ uint32 = 1003
	MSG_REGISTER_RES uint32 = 1004
	MSG_LOGOUT_REQ   uint32 = 1005
	MSG_LOGOUT_RES   uint32 = 1006
)

const (
	MSG_PLAYER_INFO_REQ uint32 = 3001
	MSG_PLAYER_INFO_RES uint32 = 3002
	MSG_PLAYER_MOVE_REQ uint32 = 3003
	MSG_PLAYER_MOVE_RES uint32 = 3004
)

const (
	MSG_ATTACK_REQ     uint32 = 4001
	MSG_ATTACK_RES     uint32 = 4002
	MSG_SKILL_REQ      uint32 = 4003
	MSG_SKILL_RES      uint32 = 4004
	MSG_BATTLE_END_REQ uint32 = 4005
	MSG_BATTLE_END_RES uint32 = 4006
)

const (
	MSG_MAP_LOAD_REQ       uint32 = 5001
	MSG_MAP_LOAD_RES       uint32 = 5002
	MSG_MAP_ENTITY_REQ     uint32 = 5003
	MSG_MAP_ENTITY_RES     uint32 = 5004
	MSG_MAP_PLAYER_ENTER   uint32 = 5005
	MSG_MAP_PLAYER_LEAVE   uint32 = 5006
	MSG_MAP_PLAYER_MOVE    uint32 = 5007
	MSG_MAP_PLAYER_SYNC    uint32 = 5008
	MSG_MAP_ENTITY_SYNC    uint32 = 5009
	MSG_MAP_CROSS_GRID_REQ uint32 = 5010
	MSG_MAP_CROSS_GRID_RES uint32 = 5011
	MSG_MAP_CHUNK_LOAD_REQ uint32 = 5012
	MSG_MAP_CHUNK_LOAD_RES uint32 = 5013
)

const (
	MSG_INVENTORY_REQ uint32 = 6001
	MSG_INVENTORY_RES uint32 = 6002
	MSG_ITEM_USE_REQ  uint32 = 6003
	MSG_ITEM_USE_RES  uint32 = 6004
)

const (
	MSG_TASK_LIST_REQ   uint32 = 7001
	MSG_TASK_LIST_RES   uint32 = 7002
	MSG_TASK_ACCEPT_REQ uint32 = 7003
	MSG_TASK_ACCEPT_RES uint32 = 7004
	MSG_TASK_FINISH_REQ uint32 = 7005
	MSG_TASK_FINISH_RES uint32 = 7006
)

const (
	MSG_PING uint32 = 9999
	MSG_PONG uint32 = 9998
)

const (
	MSG_DB_ACCOUNT_GET      uint32 = 10001
	MSG_DB_ACCOUNT_CREATE   uint32 = 10002
	MSG_DB_ACCOUNT_EXISTS   uint32 = 10003
	MSG_DB_PLAYER_GET       uint32 = 10004
	MSG_DB_PLAYER_CREATE    uint32 = 10005
	MSG_DB_PLAYER_UPDATE    uint32 = 10006
	MSG_DB_INVENTORY_GET    uint32 = 10007
	MSG_DB_INVENTORY_UPDATE uint32 = 10008
	MSG_DB_ITEM_GET         uint32 = 10009
)

type Message struct {
	ID       uint32
	NodeType NodeType
	Data     []byte
}

type LoginRequest struct {
	Account  string `json:"account"`
	Password string `json:"password"`
	DeviceID string `json:"device_id"`
}

type LoginResponse struct {
	Result     int    `json:"result"`
	Message    string `json:"message"`
	SessionID  string `json:"session_id"`
	PlayerID   int64  `json:"player_id"`
	PlayerName string `json:"player_name"`
	Level      int32  `json:"level"`
	Health     int32  `json:"health"`
}

type RegisterRequest struct {
	Account  string `json:"account"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	Result  int    `json:"result"`
	Message string `json:"message"`
}

type LogoutRequest struct {
	SessionID string `json:"session_id"`
}

type LogoutResponse struct {
	Result  int    `json:"result"`
	Message string `json:"message"`
}

type PlayerInfoRequest struct {
	PlayerID  int64  `json:"player_id"`
	SessionID string `json:"session_id"`
}

type PlayerInfoResponse struct {
	Result     int     `json:"result"`
	Message    string  `json:"message"`
	PlayerID   int64   `json:"player_id"`
	PlayerName string  `json:"player_name"`
	Level      int32   `json:"level"`
	Exp        int64   `json:"exp"`
	Health     int32   `json:"health"`
	MaxHealth  int32   `json:"max_health"`
	Mana       int32   `json:"mana"`
	MaxMana    int32   `json:"max_mana"`
	PositionX  float64 `json:"pos_x"`
	PositionY  float64 `json:"pos_y"`
}

type PlayerMoveRequest struct {
	PlayerID  int64   `json:"player_id,omitempty"`
	SessionID string  `json:"session_id,omitempty"`
	TargetX   float64 `json:"target_x,omitempty"`
	TargetY   float64 `json:"target_y,omitempty"`
}

type PlayerMoveResponse struct {
	Result  int     `json:"result"`
	Message string  `json:"message"`
	PosX    float64 `json:"pos_x"`
	PosY    float64 `json:"pos_y"`
}

type AttackRequest struct {
	PlayerID  int64  `json:"player_id"`
	SessionID string `json:"session_id"`
	TargetID  int64  `json:"target_id"`
}

type AttackResponse struct {
	Result     int    `json:"result"`
	Message    string `json:"message"`
	AttackerID int64  `json:"attacker_id"`
	TargetID   int64  `json:"target_id"`
	Damage     int32  `json:"damage"`
	TargetHP   int32  `json:"target_hp"`
}

type SkillRequest struct {
	PlayerID  int64  `json:"player_id"`
	SessionID string `json:"session_id"`
	SkillID   uint32 `json:"skill_id"`
	TargetID  int64  `json:"target_id"`
}

type SkillResponse struct {
	Result   int    `json:"result"`
	Message  string `json:"message"`
	SkillID  uint32 `json:"skill_id"`
	TargetID int64  `json:"target_id"`
	Damage   int32  `json:"damage"`
	TargetHP int32  `json:"target_hp"`
}

type InventoryRequest struct {
	PlayerID  int64  `json:"player_id"`
	SessionID string `json:"session_id"`
}

type EquipmentInfo struct {
	Slot   int32  `json:"slot"`
	ItemID uint32 `json:"item_id"`
	Level  int32  `json:"level"`
}

type InventoryResponse struct {
	Result      int             `json:"result"`
	Message     string          `json:"message"`
	Items       []ItemInfo      `json:"items"`
	Gold        int64           `json:"gold"`
	Equipments  []EquipmentInfo `json:"equipments"`
	Capacity    int32           `json:"capacity"`
	ItemConfigs []ItemConfig    `json:"item_configs"`
}

type ItemInfo struct {
	ItemID      uint32 `json:"item_id"`
	Name        string `json:"name"`
	Icon        string `json:"icon"`
	Count       int32  `json:"count"`
	Slot        int32  `json:"slot"`
	Level       int32  `json:"level"`
	UID         string `json:"uid"`
	Description string `json:"description"`
}

type ItemUseRequest struct {
	PlayerID  int64  `json:"player_id"`
	SessionID string `json:"session_id"`
	ItemID    uint32 `json:"item_id"`
	Position  int32  `json:"position"`
}

type ItemUseResponse struct {
	Result  int    `json:"result"`
	Message string `json:"message"`
}

type TaskListRequest struct {
	PlayerID  int64  `json:"player_id"`
	SessionID string `json:"session_id"`
}

type TaskListResponse struct {
	Result  int        `json:"result"`
	Message string     `json:"message"`
	Tasks   []TaskInfo `json:"tasks"`
}

type TaskInfo struct {
	TaskID      uint32 `json:"task_id"`
	Title       string `json:"title"`
	Status      int    `json:"status"`
	Progress    int    `json:"progress"`
	MaxProgress int    `json:"max_progress"`
}

type TaskAcceptRequest struct {
	PlayerID  int64  `json:"player_id"`
	SessionID string `json:"session_id"`
	TaskID    uint32 `json:"task_id"`
}

type TaskAcceptResponse struct {
	Result  int    `json:"result"`
	Message string `json:"message"`
}

type TaskFinishRequest struct {
	PlayerID  int64  `json:"player_id"`
	SessionID string `json:"session_id"`
	TaskID    uint32 `json:"task_id"`
}

type TaskFinishResponse struct {
	Result  int          `json:"result"`
	Message string       `json:"message"`
	Rewards []RewardInfo `json:"rewards"`
}

type RewardInfo struct {
	Type  string `json:"type"`
	Value int    `json:"value"`
}

type MapLoadRequest struct {
	PlayerID  int64   `json:"player_id"`
	SessionID string  `json:"session_id"`
	MapID     int32   `json:"map_id"`
	TargetX   float64 `json:"target_x"`
	TargetY   float64 `json:"target_y"`
}

type MapLoadResponse struct {
	Result    int           `json:"result"`
	Message   string        `json:"message"`
	MapID     int32         `json:"map_id"`
	PlayerPos *PositionInfo `json:"player_pos"`
	Chunks    []*ChunkInfo  `json:"chunks"`
}

type PositionInfo struct {
	X      float64 `json:"x"`
	Y      float64 `json:"y"`
	Z      float64 `json:"z"`
	Rot    float64 `json:"rot"`
	GridID int     `json:"grid_id"`
}

type ChunkInfo struct {
	ChunkX   int32         `json:"chunk_x"`
	ChunkY   int32         `json:"chunk_y"`
	Tiles    []int32       `json:"tiles"`
	Entities []*EntityInfo `json:"entities"`
}

type EntityInfo struct {
	EntityID   int64   `json:"entity_id"`
	EntityType int32   `json:"entity_type"`
	Name       string  `json:"name"`
	PosX       float64 `json:"pos_x"`
	PosY       float64 `json:"pos_y"`
	PosZ       float64 `json:"pos_z"`
	Rotation   float64 `json:"rotation"`
	Health     int32   `json:"health"`
	MaxHealth  int32   `json:"max_health"`
	State      int32   `json:"state"`
}

type MapPlayerEnterRequest struct {
	PlayerID  int64   `json:"player_id"`
	Name      string  `json:"name"`
	PosX      float64 `json:"pos_x"`
	PosY      float64 `json:"pos_y"`
	PosZ      float64 `json:"pos_z"`
	Rotation  float64 `json:"rotation"`
	Level     int32   `json:"level"`
	Health    int32   `json:"health"`
	MaxHealth int32   `json:"max_health"`
}

type MapPlayerEnterResponse struct {
	Result  int    `json:"result"`
	Message string `json:"message"`
}

type MapPlayerLeaveRequest struct {
	PlayerID int64 `json:"player_id"`
}

type MapPlayerLeaveResponse struct {
	Result  int    `json:"result"`
	Message string `json:"message"`
}

type MapPlayerMoveRequest struct {
	PlayerID int64   `json:"player_id"`
	PosX     float64 `json:"pos_x"`
	PosY     float64 `json:"pos_y"`
	PosZ     float64 `json:"pos_z"`
	Rotation float64 `json:"rotation"`
}

type MapPlayerMoveResponse struct {
	Result  int    `json:"result"`
	Message string `json:"message"`
}

type MapPlayerSyncRequest struct {
	PlayerID int64         `json:"player_id"`
	Players  []*PlayerSync `json:"players"`
}

type PlayerSync struct {
	PlayerID  int64   `json:"player_id"`
	Name      string  `json:"name"`
	PosX      float64 `json:"pos_x"`
	PosY      float64 `json:"pos_y"`
	PosZ      float64 `json:"pos_z"`
	Rotation  float64 `json:"rotation"`
	State     int32   `json:"state"`
	Health    int32   `json:"health"`
	MaxHealth int32   `json:"max_health"`
	Level     int32   `json:"level"`
}

type MapEntitySyncRequest struct {
	Entities []*EntityInfo `json:"entities"`
}

type MapCrossGridRequest struct {
	PlayerID   int64   `json:"player_id"`
	FromGridID int     `json:"from_grid_id"`
	ToGridID   int     `json:"to_grid_id"`
	PosX       float64 `json:"pos_x"`
	PosY       float64 `json:"pos_y"`
	PosZ       float64 `json:"pos_z"`
}

type MapCrossGridResponse struct {
	Result     int    `json:"result"`
	Message    string `json:"message"`
	TargetGrid string `json:"target_grid"`
}

type MapChunkLoadRequest struct {
	PlayerID int64 `json:"player_id"`
	ChunkX   int32 `json:"chunk_x"`
	ChunkY   int32 `json:"chunk_y"`
}

type MapChunkLoadResponse struct {
	Result   int           `json:"result"`
	Message  string        `json:"message"`
	ChunkX   int32         `json:"chunk_x"`
	ChunkY   int32         `json:"chunk_y"`
	Tiles    []int32       `json:"tiles"`
	Entities []*EntityInfo `json:"entities"`
}

type DBResponse struct {
	Result  int         `json:"result"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type AccountData struct {
	ID       int64  `json:"id"`
	Account  string `json:"account"`
	Password string `json:"password"`
	Salt     string `json:"salt"`
}

type PlayerData struct {
	ID        int64   `json:"id"`
	Name      string  `json:"name"`
	AccountID int64   `json:"account_id"`
	Level     int32   `json:"level"`
	Exp       int64   `json:"exp"`
	PosX      float64 `json:"pos_x"`
	PosY      float64 `json:"pos_y"`
	Health    int32   `json:"health"`
	MaxHealth int32   `json:"max_health"`
	Mana      int32   `json:"mana"`
	MaxMana   int32   `json:"max_mana"`
}

type InventoryItem struct {
	ID       int64 `json:"id"`
	PlayerID int64 `json:"player_id"`
	ItemID   int64 `json:"item_id"`
	Slot     int32 `json:"slot"`
	Count    int32 `json:"count"`
}

type ItemConfig struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Type        int32  `json:"type"`
	EffectType  int32  `json:"effect_type"`
	EffectValue int32  `json:"effect_value"`
	Icon        string `json:"icon"`
	Description string `json:"description"`
}

func GetMsgName(msgID uint32) string {
	switch msgID {
	case MSG_LOGIN_REQ:
		return "MSG_LOGIN_REQ"
	case MSG_LOGIN_RES:
		return "MSG_LOGIN_RES"
	case MSG_REGISTER_REQ:
		return "MSG_REGISTER_REQ"
	case MSG_REGISTER_RES:
		return "MSG_REGISTER_RES"
	case MSG_LOGOUT_REQ:
		return "MSG_LOGOUT_REQ"
	case MSG_LOGOUT_RES:
		return "MSG_LOGOUT_RES"
	case MSG_PLAYER_INFO_REQ:
		return "MSG_PLAYER_INFO_REQ"
	case MSG_PLAYER_INFO_RES:
		return "MSG_PLAYER_INFO_RES"
	case MSG_PLAYER_MOVE_REQ:
		return "MSG_PLAYER_MOVE_REQ"
	case MSG_PLAYER_MOVE_RES:
		return "MSG_PLAYER_MOVE_RES"
	case MSG_ATTACK_REQ:
		return "MSG_ATTACK_REQ"
	case MSG_ATTACK_RES:
		return "MSG_ATTACK_RES"
	case MSG_SKILL_REQ:
		return "MSG_SKILL_REQ"
	case MSG_SKILL_RES:
		return "MSG_SKILL_RES"
	case MSG_INVENTORY_REQ:
		return "MSG_INVENTORY_REQ"
	case MSG_INVENTORY_RES:
		return "MSG_INVENTORY_RES"
	case MSG_ITEM_USE_REQ:
		return "MSG_ITEM_USE_REQ"
	case MSG_ITEM_USE_RES:
		return "MSG_ITEM_USE_RES"
	case MSG_TASK_LIST_REQ:
		return "MSG_TASK_LIST_REQ"
	case MSG_TASK_LIST_RES:
		return "MSG_TASK_LIST_RES"
	case MSG_TASK_ACCEPT_REQ:
		return "MSG_TASK_ACCEPT_REQ"
	case MSG_TASK_ACCEPT_RES:
		return "MSG_TASK_ACCEPT_RES"
	case MSG_TASK_FINISH_REQ:
		return "MSG_TASK_FINISH_REQ"
	case MSG_TASK_FINISH_RES:
		return "MSG_TASK_FINISH_RES"
	case MSG_PING:
		return "MSG_PING"
	case MSG_PONG:
		return "MSG_PONG"
	case MSG_MAP_LOAD_REQ:
		return "MSG_MAP_LOAD_REQ"
	case MSG_MAP_LOAD_RES:
		return "MSG_MAP_LOAD_RES"
	case MSG_MAP_ENTITY_REQ:
		return "MSG_MAP_ENTITY_REQ"
	case MSG_MAP_ENTITY_RES:
		return "MSG_MAP_ENTITY_RES"
	case MSG_MAP_PLAYER_ENTER:
		return "MSG_MAP_PLAYER_ENTER"
	case MSG_MAP_PLAYER_LEAVE:
		return "MSG_MAP_PLAYER_LEAVE"
	case MSG_MAP_PLAYER_MOVE:
		return "MSG_MAP_PLAYER_MOVE"
	case MSG_MAP_PLAYER_SYNC:
		return "MSG_MAP_PLAYER_SYNC"
	case MSG_MAP_ENTITY_SYNC:
		return "MSG_MAP_ENTITY_SYNC"
	case MSG_MAP_CROSS_GRID_REQ:
		return "MSG_MAP_CROSS_GRID_REQ"
	case MSG_MAP_CROSS_GRID_RES:
		return "MSG_MAP_CROSS_GRID_RES"
	case MSG_MAP_CHUNK_LOAD_REQ:
		return "MSG_MAP_CHUNK_LOAD_REQ"
	case MSG_MAP_CHUNK_LOAD_RES:
		return "MSG_MAP_CHUNK_LOAD_RES"
	case MSG_DB_ACCOUNT_GET:
		return "MSG_DB_ACCOUNT_GET"
	case MSG_DB_ACCOUNT_CREATE:
		return "MSG_DB_ACCOUNT_CREATE"
	case MSG_DB_ACCOUNT_EXISTS:
		return "MSG_DB_ACCOUNT_EXISTS"
	case MSG_DB_PLAYER_GET:
		return "MSG_DB_PLAYER_GET"
	case MSG_DB_PLAYER_CREATE:
		return "MSG_DB_PLAYER_CREATE"
	case MSG_DB_PLAYER_UPDATE:
		return "MSG_DB_PLAYER_UPDATE"
	case MSG_DB_INVENTORY_GET:
		return "MSG_DB_INVENTORY_GET"
	case MSG_DB_INVENTORY_UPDATE:
		return "MSG_DB_INVENTORY_UPDATE"
	case MSG_DB_ITEM_GET:
		return "MSG_DB_ITEM_GET"
	default:
		return "Unknown"
	}
}

func UnmarshalLoginRequest(data []byte) (*LoginRequest, error) {
	var req LoginRequest
	err := json.Unmarshal(data, &req)
	return &req, err
}

func UnmarshalRegisterRequest(data []byte) (*RegisterRequest, error) {
	var req RegisterRequest
	err := json.Unmarshal(data, &req)
	return &req, err
}

func MarshalResponse(data interface{}) ([]byte, error) {
	return json.Marshal(data)
}
