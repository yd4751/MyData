package msg

import (
	"encoding/json"
	"sync"
)

type MessageRouter struct {
	msgToNodeType map[uint32]NodeType // 消息ID到节点类型的路由映射
	sync.RWMutex                      // 读写锁
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
	RegisterMessageRoute(MSG_ATTACK_REQ, NodeTypeBattle)
	RegisterMessageRoute(MSG_ATTACK_RES, NodeTypeBattle)
	RegisterMessageRoute(MSG_SKILL_REQ, NodeTypeBattle)
	RegisterMessageRoute(MSG_SKILL_RES, NodeTypeBattle)
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

func (n NodeType) ServiceName() string {
	switch n {
	case NodeTypeGate:
		return "gate"
	case NodeTypeLogin:
		return "login"
	case NodeTypeLogic:
		return "logic"
	case NodeTypeBattle:
		return "battle"
	case NodeTypeGridMap:
		return "gridmap"
	case NodeTypeCross:
		return "cross"
	case NodeTypeData:
		return "dataservice"
	case NodeTypeGM:
		return "gm"
	default:
		return ""
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
	ID       uint32   // 消息ID
	NodeType NodeType // 目标节点类型
	Data     []byte   // 消息数据
}

type LoginRequest struct {
	Account  string `json:"account"`   // 账号
	Password string `json:"password"`  // 密码
	DeviceID string `json:"device_id"` // 设备ID
}

type LoginResponse struct {
	Result     int    `json:"result"`      // 结果码(0成功)
	Message    string `json:"message"`     // 提示消息
	SessionID  string `json:"session_id"`  // 会话ID
	PlayerID   int64  `json:"player_id"`   // 玩家ID
	PlayerName string `json:"player_name"` // 玩家名称
	Level      int32  `json:"level"`       // 玩家等级
	Health     int32  `json:"health"`      // 当前生命值
}

type RegisterRequest struct {
	Account  string `json:"account"`  // 账号
	Password string `json:"password"` // 密码
}

type RegisterResponse struct {
	Result  int    `json:"result"`  // 结果码(0成功)
	Message string `json:"message"` // 提示消息
}

type LogoutRequest struct {
	SessionID string `json:"session_id"` // 会话ID
}

type LogoutResponse struct {
	Result  int    `json:"result"`  // 结果码(0成功)
	Message string `json:"message"` // 提示消息
}

type PlayerInfoRequest struct {
	PlayerID  int64  `json:"player_id"`  // 玩家ID
	SessionID string `json:"session_id"` // 会话ID
}

type PlayerInfoResponse struct {
	Result     int     `json:"result"`      // 结果码(0成功)
	Message    string  `json:"message"`     // 提示消息
	PlayerID   int64   `json:"player_id"`   // 玩家ID
	PlayerName string  `json:"player_name"` // 玩家名称
	Level      int32   `json:"level"`       // 等级
	Exp        int64   `json:"exp"`         // 经验值
	Health     int32   `json:"health"`      // 当前生命值
	MaxHealth  int32   `json:"max_health"`  // 最大生命值
	Mana       int32   `json:"mana"`        // 当前魔法值
	MaxMana    int32   `json:"max_mana"`    // 最大魔法值
	PositionX  float64 `json:"pos_x"`       // X坐标
	PositionY  float64 `json:"pos_y"`       // Y坐标
}

type PlayerMoveRequest struct {
	PlayerID  int64   `json:"player_id,omitempty"`  // 玩家ID
	SessionID string  `json:"session_id,omitempty"` // 会话ID
	TargetX   float64 `json:"target_x,omitempty"`   // 目标X坐标
	TargetY   float64 `json:"target_y,omitempty"`   // 目标Y坐标
}

type PlayerMoveResponse struct {
	Result  int     `json:"result"`  // 结果码(0成功)
	Message string  `json:"message"` // 提示消息
	PosX    float64 `json:"pos_x"`   // 当前X坐标
	PosY    float64 `json:"pos_y"`   // 当前Y坐标
}

type AttackRequest struct {
	PlayerID  int64  `json:"player_id"`  // 玩家ID
	SessionID string `json:"session_id"` // 会话ID
	TargetID  int64  `json:"target_id"`  // 目标ID
}

type AttackResponse struct {
	Result     int    `json:"result"`      // 结果码(0成功)
	Message    string `json:"message"`     // 提示消息
	AttackerID int64  `json:"attacker_id"` // 攻击者ID
	TargetID   int64  `json:"target_id"`   // 目标ID
	Damage     int32  `json:"damage"`      // 造成伤害
	TargetHP   int32  `json:"target_hp"`   // 目标剩余生命值
}

type SkillRequest struct {
	PlayerID  int64  `json:"player_id"`  // 玩家ID
	SessionID string `json:"session_id"` // 会话ID
	SkillID   uint32 `json:"skill_id"`   // 技能ID
	TargetID  int64  `json:"target_id"`  // 目标ID
}

type SkillResponse struct {
	Result   int    `json:"result"`    // 结果码(0成功)
	Message  string `json:"message"`   // 提示消息
	SkillID  uint32 `json:"skill_id"`  // 技能ID
	TargetID int64  `json:"target_id"` // 目标ID
	Damage   int32  `json:"damage"`    // 造成伤害
	TargetHP int32  `json:"target_hp"` // 目标剩余生命值
}

type InventoryRequest struct {
	PlayerID  int64  `json:"player_id"`  // 玩家ID
	SessionID string `json:"session_id"` // 会话ID
}

type EquipmentInfo struct {
	Slot   int32  `json:"slot"`    // 装备槽位
	ItemID uint32 `json:"item_id"` // 物品ID
	Level  int32  `json:"level"`   // 强化等级
}

type InventoryResponse struct {
	Result      int             `json:"result"`       // 结果码(0成功)
	Message     string          `json:"message"`      // 提示消息
	Items       []ItemInfo      `json:"items"`        // 背包物品列表
	Gold        int64           `json:"gold"`         // 金币数量
	Equipments  []EquipmentInfo `json:"equipments"`   // 装备列表
	Capacity    int32           `json:"capacity"`     // 背包容量
	ItemConfigs []ItemConfig    `json:"item_configs"` // 物品配置列表
}

type ItemInfo struct {
	ItemID      uint32 `json:"item_id"`     // 物品ID
	Name        string `json:"name"`        // 物品名称
	Icon        string `json:"icon"`        // 图标路径
	Count       int32  `json:"count"`       // 数量
	Slot        int32  `json:"slot"`        // 槽位
	Level       int32  `json:"level"`       // 强化等级
	UID         string `json:"uid"`         // 唯一标识符
	Description string `json:"description"` // 描述
}

type ItemUseRequest struct {
	PlayerID  int64  `json:"player_id"`  // 玩家ID
	SessionID string `json:"session_id"` // 会话ID
	ItemID    uint32 `json:"item_id"`    // 物品ID
	Position  int32  `json:"position"`   // 使用位置
}

type ItemUseResponse struct {
	Result  int    `json:"result"`  // 结果码(0成功)
	Message string `json:"message"` // 提示消息
	Slot    int32  `json:"slot"`    // 物品槽位
	Count   int32  `json:"count"`   // 剩余数量
}

type TaskListRequest struct {
	PlayerID  int64  `json:"player_id"`  // 玩家ID
	SessionID string `json:"session_id"` // 会话ID
}

type TaskListResponse struct {
	Result  int        `json:"result"`  // 结果码(0成功)
	Message string     `json:"message"` // 提示消息
	Tasks   []TaskInfo `json:"tasks"`   // 任务列表
}

type TaskInfo struct {
	TaskID      uint32 `json:"task_id"`      // 任务ID
	Title       string `json:"title"`        // 任务标题
	Status      int    `json:"status"`       // 状态(0-未接取,1-进行中,2-完成)
	Progress    int    `json:"progress"`     // 当前进度
	MaxProgress int    `json:"max_progress"` // 最大进度
}

type TaskAcceptRequest struct {
	PlayerID  int64  `json:"player_id"`  // 玩家ID
	SessionID string `json:"session_id"` // 会话ID
	TaskID    uint32 `json:"task_id"`    // 任务ID
}

type TaskAcceptResponse struct {
	Result  int    `json:"result"`  // 结果码(0成功)
	Message string `json:"message"` // 提示消息
}

type TaskFinishRequest struct {
	PlayerID  int64  `json:"player_id"`  // 玩家ID
	SessionID string `json:"session_id"` // 会话ID
	TaskID    uint32 `json:"task_id"`    // 任务ID
}

type TaskFinishResponse struct {
	Result  int          `json:"result"`  // 结果码(0成功)
	Message string       `json:"message"` // 提示消息
	Rewards []RewardInfo `json:"rewards"` // 奖励列表
}

type RewardInfo struct {
	Type  string `json:"type"`  // 奖励类型(gold/exp/item)
	Value int    `json:"value"` // 奖励值
}

type MapLoadRequest struct {
	PlayerID  int64   `json:"player_id"`  // 玩家ID
	SessionID string  `json:"session_id"` // 会话ID
	MapID     int32   `json:"map_id"`     // 地图ID
	TargetX   float64 `json:"target_x"`   // 目标X坐标
	TargetY   float64 `json:"target_y"`   // 目标Y坐标
}

type MapLoadResponse struct {
	Result    int           `json:"result"`     // 结果码(0成功)
	Message   string        `json:"message"`    // 提示消息
	MapID     int32         `json:"map_id"`     // 地图ID
	PlayerPos *PositionInfo `json:"player_pos"` // 玩家位置
	Chunks    []*ChunkInfo  `json:"chunks"`     // 区块列表
}

type PositionInfo struct {
	X      float64 `json:"x"`       // X坐标
	Y      float64 `json:"y"`       // Y坐标
	Z      float64 `json:"z"`       // Z坐标
	Rot    float64 `json:"rot"`     // 旋转角度
	GridID int     `json:"grid_id"` // 网格ID
}

type ChunkInfo struct {
	ChunkX   int32         `json:"chunk_x"`  // 区块X
	ChunkY   int32         `json:"chunk_y"`  // 区块Y
	Tiles    []int32       `json:"tiles"`    // 瓦片数据
	Entities []*EntityInfo `json:"entities"` // 实体列表
}

type EntityInfo struct {
	EntityID   int64   `json:"entity_id"`   // 实体ID
	EntityType int32   `json:"entity_type"` // 实体类型
	Name       string  `json:"name"`        // 名称
	PosX       float64 `json:"pos_x"`       // X坐标
	PosY       float64 `json:"pos_y"`       // Y坐标
	PosZ       float64 `json:"pos_z"`       // Z坐标
	Rotation   float64 `json:"rotation"`    // 旋转角度
	Health     int32   `json:"health"`      // 当前生命值
	MaxHealth  int32   `json:"max_health"`  // 最大生命值
	State      int32   `json:"state"`       // 状态
}

type MapPlayerEnterRequest struct {
	PlayerID  int64   `json:"player_id"`  // 玩家ID
	AccountID int64   `json:"account_id"` // 账号ID
	Name      string  `json:"name"`       // 玩家名称
	PosX      float64 `json:"pos_x"`      // X坐标
	PosY      float64 `json:"pos_y"`      // Y坐标
	PosZ      float64 `json:"pos_z"`      // Z坐标
	Rotation  float64 `json:"rotation"`   // 旋转角度
	Level     int32   `json:"level"`      // 等级
	Health    int32   `json:"health"`     // 当前生命值
	MaxHealth int32   `json:"max_health"` // 最大生命值
}

type MapPlayerEnterResponse struct {
	Result  int    `json:"result"`  // 结果码(0成功)
	Message string `json:"message"` // 提示消息
}

type MapPlayerLeaveRequest struct {
	PlayerID int64 `json:"player_id"` // 玩家ID
}

type MapPlayerLeaveResponse struct {
	Result  int    `json:"result"`  // 结果码(0成功)
	Message string `json:"message"` // 提示消息
}

type MapPlayerMoveRequest struct {
	PlayerID int64   `json:"player_id"` // 玩家ID
	PosX     float64 `json:"pos_x"`     // X坐标
	PosY     float64 `json:"pos_y"`     // Y坐标
	PosZ     float64 `json:"pos_z"`     // Z坐标
	Rotation float64 `json:"rotation"`  // 旋转角度
}

type MapPlayerMoveResponse struct {
	Result  int    `json:"result"`  // 结果码(0成功)
	Message string `json:"message"` // 提示消息
}

type MapPlayerSyncRequest struct {
	PlayerID int64         `json:"player_id"` // 玩家ID
	Players  []*PlayerSync `json:"players"`   // 玩家同步列表
}

type PlayerSync struct {
	PlayerID  int64   `json:"player_id"`  // 玩家ID
	Name      string  `json:"name"`       // 名称
	PosX      float64 `json:"pos_x"`      // X坐标
	PosY      float64 `json:"pos_y"`      // Y坐标
	PosZ      float64 `json:"pos_z"`      // Z坐标
	Rotation  float64 `json:"rotation"`   // 旋转角度
	State     int32   `json:"state"`      // 状态
	Health    int32   `json:"health"`     // 当前生命值
	MaxHealth int32   `json:"max_health"` // 最大生命值
	Level     int32   `json:"level"`      // 等级
}

type MapEntitySyncRequest struct {
	Entities []*EntityInfo `json:"entities"` // 实体同步列表
}

type MapCrossGridRequest struct {
	PlayerID   int64   `json:"player_id"`    // 玩家ID
	AccountID  int64   `json:"account_id"`   // 账号ID
	FromGridID int     `json:"from_grid_id"` // 原网格ID
	ToGridID   int     `json:"to_grid_id"`   // 目标网格ID
	PosX       float64 `json:"pos_x"`        // X坐标
	PosY       float64 `json:"pos_y"`        // Y坐标
	PosZ       float64 `json:"pos_z"`        // Z坐标
	Name       string  `json:"name"`         // 玩家名称
	Level      int32   `json:"level"`        // 等级
	Health     int32   `json:"health"`       // 当前生命值
	MaxHealth  int32   `json:"max_health"`   // 最大生命值
	Rotation   float64 `json:"rotation"`     // 旋转角度
}

type MapCrossGridResponse struct {
	Result     int    `json:"result"`      // 结果码(0成功)
	Message    string `json:"message"`     // 提示消息
	TargetGrid string `json:"target_grid"` // 目标网格地址
}

type MapChunkLoadRequest struct {
	PlayerID int64 `json:"player_id"` // 玩家ID
	ChunkX   int32 `json:"chunk_x"`   // 区块X
	ChunkY   int32 `json:"chunk_y"`   // 区块Y
}

type MapChunkLoadResponse struct {
	Result   int           `json:"result"`   // 结果码(0成功)
	Message  string        `json:"message"`  // 提示消息
	ChunkX   int32         `json:"chunk_x"`  // 区块X
	ChunkY   int32         `json:"chunk_y"`  // 区块Y
	Tiles    []int32       `json:"tiles"`    // 瓦片数据
	Entities []*EntityInfo `json:"entities"` // 实体列表
}

type DBResponse struct {
	Result  int         `json:"result"`  // 结果码(0成功)
	Message string      `json:"message"` // 提示消息
	Data    interface{} `json:"data"`    // 返回数据
}

type AccountData struct {
	ID       int64  `json:"id"`       // 账号ID
	Account  string `json:"account"`  // 账号名
	Password string `json:"password"` // 密码
	Salt     string `json:"salt"`     // 加密盐
}

type PlayerData struct {
	ID        int64   `json:"id"`         // 玩家ID
	Name      string  `json:"name"`       // 玩家名称
	AccountID int64   `json:"account_id"` // 账号ID
	Level     int32   `json:"level"`      // 等级
	Exp       int64   `json:"exp"`        // 经验值
	PosX      float64 `json:"pos_x"`      // X坐标
	PosY      float64 `json:"pos_y"`      // Y坐标
	Health    int32   `json:"health"`     // 当前生命值
	MaxHealth int32   `json:"max_health"` // 最大生命值
	Mana      int32   `json:"mana"`       // 当前魔法值
	MaxMana   int32   `json:"max_mana"`   // 最大魔法值
}

type InventoryItem struct {
	ID       int64 `json:"id"`        // 物品实例ID
	PlayerID int64 `json:"player_id"` // 玩家ID
	ItemID   int64 `json:"item_id"`   // 物品配置ID
	Slot     int32 `json:"slot"`      // 槽位
	Count    int32 `json:"count"`     // 数量
}

type ItemConfig struct {
	ID          int64  `json:"id"`           // 物品配置ID
	Name        string `json:"name"`         // 物品名称
	Type        int32  `json:"type"`         // 物品类型
	EffectType  int32  `json:"effect_type"`  // 效果类型
	EffectValue int32  `json:"effect_value"` // 效果值
	Icon        string `json:"icon"`         // 图标路径
	Description string `json:"description"`  // 描述
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
