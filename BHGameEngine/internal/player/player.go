package player

import (
	"sync"
	"time"

	"github.com/openworld-server/internal/worldmap"
)

type ItemType int32

const (
	ItemTypeConsumable ItemType = 1
	ItemTypeEquipment  ItemType = 2
	ItemTypeMaterial   ItemType = 3
	ItemTypeQuest      ItemType = 4
)

type ItemConfig struct {
	ID          int64
	Name        string
	Type        ItemType
	EffectType  int32
	EffectValue int32
	Cooldown    int32
	MaxStack    int32
}

var ItemConfigs = map[int64]*ItemConfig{
	1001: {ID: 1001, Name: "小型生命药水", Type: ItemTypeConsumable, EffectType: 1, EffectValue: 50, Cooldown: 0, MaxStack: 20},
	1002: {ID: 1002, Name: "中型生命药水", Type: ItemTypeConsumable, EffectType: 1, EffectValue: 100, Cooldown: 0, MaxStack: 20},
	1003: {ID: 1003, Name: "大型生命药水", Type: ItemTypeConsumable, EffectType: 1, EffectValue: 200, Cooldown: 0, MaxStack: 20},
	2001: {ID: 2001, Name: "小型魔法药水", Type: ItemTypeConsumable, EffectType: 2, EffectValue: 30, Cooldown: 0, MaxStack: 20},
	2002: {ID: 2002, Name: "中型魔法药水", Type: ItemTypeConsumable, EffectType: 2, EffectValue: 60, Cooldown: 0, MaxStack: 20},
	3001: {ID: 3001, Name: "力量药剂", Type: ItemTypeConsumable, EffectType: 3, EffectValue: 10, Cooldown: 300, MaxStack: 10},
	3002: {ID: 3002, Name: "敏捷药剂", Type: ItemTypeConsumable, EffectType: 4, EffectValue: 10, Cooldown: 300, MaxStack: 10},
}

type InventoryItem struct {
	ItemID     int64
	Slot       int32
	Count      int32
	Level      int32
	ExpireTime int64
}

type CooldownEntry struct {
	ItemID  int64
	EndTime int64
}

type Player struct {
	ID         int64
	Name       string
	AccountID  int64
	Level      int32
	Exp        int64
	Pos        worldmap.Vec3
	Rotation   float64
	Health     int32
	MaxHealth  int32
	Mana       int32
	MaxMana    int32
	Stamina    int32
	MaxStamina int32
	State      PlayerState
	ChunkPos   worldmap.ChunkPos
	TeamID     int64
	Buffs      []*Buff
	Equipments []*Equipment
	Inventory  map[int32]*InventoryItem
	Cooldowns  map[int64]*CooldownEntry
	Strength   int32
	Agility    int32
	mu         sync.RWMutex
}

type PlayerState int32

const (
	StateIdle        PlayerState = 0
	StateMoving      PlayerState = 1
	StateFlying      PlayerState = 2
	StateMounted     PlayerState = 3
	StateCombat      PlayerState = 4
	StateDead        PlayerState = 5
	StateTeleporting PlayerState = 6
)

type Buff struct {
	ID      int32
	Name    string
	Stacks  int32
	EndTime int64
}

type Equipment struct {
	Slot   int32
	ItemID int64
	Level  int32
}

type PlayerManager struct {
	players map[int64]*Player
	mu      sync.RWMutex
}

func NewPlayerManager() *PlayerManager {
	return &PlayerManager{
		players: make(map[int64]*Player),
	}
}

func (m *PlayerManager) CreatePlayer(id int64, name string, accountID int64) *Player {
	player := &Player{
		ID:         id,
		Name:       name,
		AccountID:  accountID,
		Level:      1,
		Exp:        0,
		Pos:        worldmap.Vec3{X: 0, Y: 0, Z: 0},
		Rotation:   0,
		Health:     100,
		MaxHealth:  100,
		Mana:       50,
		MaxMana:    50,
		Stamina:    100,
		MaxStamina: 100,
		State:      StateIdle,
		Buffs:      make([]*Buff, 0),
		Equipments: make([]*Equipment, 0),
		Inventory:  make(map[int32]*InventoryItem),
		Cooldowns:  make(map[int64]*CooldownEntry),
		Strength:   10,
		Agility:    10,
	}

	m.mu.Lock()
	m.players[id] = player
	m.mu.Unlock()

	return player
}

func (m *PlayerManager) GetPlayer(id int64) (*Player, bool) {
	m.mu.RLock()
	player, ok := m.players[id]
	m.mu.RUnlock()
	return player, ok
}

func (m *PlayerManager) RemovePlayer(id int64) {
	m.mu.Lock()
	delete(m.players, id)
	m.mu.Unlock()
}

func (m *PlayerManager) UpdatePosition(id int64, pos worldmap.Vec3) {
	m.mu.RLock()
	player, ok := m.players[id]
	m.mu.RUnlock()

	if ok {
		player.mu.Lock()
		player.Pos = pos
		player.mu.Unlock()
	}
}

func (m *PlayerManager) UpdateState(id int64, state PlayerState) {
	m.mu.RLock()
	player, ok := m.players[id]
	m.mu.RUnlock()

	if ok {
		player.mu.Lock()
		player.State = state
		player.mu.Unlock()
	}
}

func (m *PlayerManager) AddBuff(id int64, buff *Buff) {
	m.mu.RLock()
	player, ok := m.players[id]
	m.mu.RUnlock()

	if ok {
		player.mu.Lock()
		player.Buffs = append(player.Buffs, buff)
		player.mu.Unlock()
	}
}

func (m *PlayerManager) RemoveBuff(id int64, buffID int32) {
	m.mu.RLock()
	player, ok := m.players[id]
	m.mu.RUnlock()

	if ok {
		player.mu.Lock()
		for i, b := range player.Buffs {
			if b.ID == buffID {
				player.Buffs = append(player.Buffs[:i], player.Buffs[i+1:]...)
				break
			}
		}
		player.mu.Unlock()
	}
}

func (m *PlayerManager) GetOnlineCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.players)
}

func (m *PlayerManager) GetAllPlayers() []*Player {
	m.mu.RLock()
	defer m.mu.RUnlock()

	players := make([]*Player, 0, len(m.players))
	for _, p := range m.players {
		players = append(players, p)
	}
	return players
}

func (p *Player) GetPosition() worldmap.Vec3 {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.Pos
}

func (p *Player) GetState() PlayerState {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.State
}

func (p *Player) GetLevel() int32 {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.Level
}

func (p *Player) GetMaxHealth() int32 {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.MaxHealth
}

func (p *Player) TakeDamage(damage int32) {
	p.mu.Lock()
	p.Health -= damage
	if p.Health < 0 {
		p.Health = 0
		p.State = StateDead
	}
	p.mu.Unlock()
}

func (p *Player) Heal(amount int32) {
	p.mu.Lock()
	p.Health += amount
	if p.Health > p.MaxHealth {
		p.Health = p.MaxHealth
	}
	p.mu.Unlock()
}

func (p *Player) ConsumeMana(amount int32) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.Mana >= amount {
		p.Mana -= amount
		return true
	}
	return false
}

func (p *Player) ConsumeStamina(amount int32) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.Stamina >= amount {
		p.Stamina -= amount
		return true
	}
	return false
}

func (p *Player) Regenerate() {
	p.mu.Lock()
	p.Stamina += 1
	if p.Stamina > p.MaxStamina {
		p.Stamina = p.MaxStamina
	}
	p.Mana += 1
	if p.Mana > p.MaxMana {
		p.Mana = p.MaxMana
	}
	p.mu.Unlock()
}

func (p *Player) ApplyBuff(buff *Buff) {
	p.mu.Lock()
	found := false
	for _, b := range p.Buffs {
		if b.ID == buff.ID {
			b.Stacks++
			b.EndTime = time.Now().Unix() + 300
			found = true
			break
		}
	}
	if !found {
		p.Buffs = append(p.Buffs, buff)
	}
	p.mu.Unlock()
}

func (p *Player) RemoveBuff(buffID int32) {
	p.mu.Lock()
	for i, b := range p.Buffs {
		if b.ID == buffID {
			p.Buffs = append(p.Buffs[:i], p.Buffs[i+1:]...)
			break
		}
	}
	p.mu.Unlock()
}

func (p *Player) UpdateBuffs() {
	now := time.Now().Unix()
	p.mu.Lock()
	newBuffs := make([]*Buff, 0)
	for _, buff := range p.Buffs {
		if buff.EndTime > now {
			newBuffs = append(newBuffs, buff)
		}
	}
	p.Buffs = newBuffs
	p.mu.Unlock()
}

func (p *Player) AddItem(itemID int64, count int32) bool {
	itemConfig, ok := ItemConfigs[itemID]
	if !ok {
		return false
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	for _, item := range p.Inventory {
		if item.ItemID == itemID && item.Count < itemConfig.MaxStack {
			item.Count += count
			return true
		}
	}

	for slot := int32(0); slot < 100; slot++ {
		if _, ok := p.Inventory[slot]; !ok {
			p.Inventory[slot] = &InventoryItem{
				ItemID: itemID,
				Slot:   slot,
				Count:  count,
			}
			return true
		}
	}

	return false
}

func (p *Player) RemoveItem(slot int32, count int32) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	item, ok := p.Inventory[slot]
	if !ok || item.Count < count {
		return false
	}

	item.Count -= count
	if item.Count <= 0 {
		delete(p.Inventory, slot)
	}

	return true
}

func (p *Player) GetItem(slot int32) (*InventoryItem, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	item, ok := p.Inventory[slot]
	return item, ok
}

func (p *Player) GetInventory() map[int32]*InventoryItem {
	p.mu.RLock()
	defer p.mu.RUnlock()

	result := make(map[int32]*InventoryItem)
	for slot, item := range p.Inventory {
		result[slot] = item
	}
	return result
}

func (p *Player) IsOnCooldown(itemID int64) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	cooldown, ok := p.Cooldowns[itemID]
	if !ok {
		return false
	}

	return cooldown.EndTime > time.Now().Unix()
}

func (p *Player) SetCooldown(itemID int64, duration int32) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.Cooldowns[itemID] = &CooldownEntry{
		ItemID:  itemID,
		EndTime: time.Now().Unix() + int64(duration),
	}
}

func (p *Player) UseItem(slot int32) (bool, string) {
	item, ok := p.GetItem(slot)
	if !ok {
		return false, "道具不存在"
	}

	itemConfig, ok := ItemConfigs[item.ItemID]
	if !ok {
		return false, "道具配置不存在"
	}

	if itemConfig.Type != ItemTypeConsumable {
		return false, "该道具不可使用"
	}

	if p.IsOnCooldown(item.ItemID) {
		return false, "道具冷却中"
	}

	p.mu.Lock()
	if p.State == StateDead {
		p.mu.Unlock()
		return false, "死亡状态无法使用道具"
	}
	p.mu.Unlock()

	switch itemConfig.EffectType {
	case 1:
		p.Heal(itemConfig.EffectValue)
	case 2:
		p.mu.Lock()
		p.Mana += itemConfig.EffectValue
		if p.Mana > p.MaxMana {
			p.Mana = p.MaxMana
		}
		p.mu.Unlock()
	case 3:
		buff := &Buff{
			ID:      101,
			Name:    "力量提升",
			Stacks:  1,
			EndTime: time.Now().Unix() + 300,
		}
		p.ApplyBuff(buff)
		p.mu.Lock()
		p.Strength += itemConfig.EffectValue
		p.mu.Unlock()
	case 4:
		buff := &Buff{
			ID:      102,
			Name:    "敏捷提升",
			Stacks:  1,
			EndTime: time.Now().Unix() + 300,
		}
		p.ApplyBuff(buff)
		p.mu.Lock()
		p.Agility += itemConfig.EffectValue
		p.mu.Unlock()
	default:
		return false, "未知的道具效果类型"
	}

	if itemConfig.Cooldown > 0 {
		p.SetCooldown(item.ItemID, itemConfig.Cooldown)
	}

	p.RemoveItem(slot, 1)

	return true, "使用成功"
}
