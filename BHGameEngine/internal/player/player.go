package player

import (
	"math"
	"sync"
	"time"

	"github.com/openworld-server/internal/item"
	"github.com/openworld-server/internal/worldmap"
)

type Player struct {
	ID           int64             // 玩家唯一标识符
	Name         string            // 玩家名称
	AccountID    int64             // 所属账号ID
	Level        int32             // 玩家等级
	Exp          int64             // 当前经验值
	Pos          worldmap.Vec3     // 世界坐标位置
	Rotation     float64           // 朝向角度
	Health       int32             // 当前生命值
	MaxHealth    int32             // 最大生命值
	Mana         int32             // 当前魔法值
	MaxMana      int32             // 最大魔法值
	Stamina      int32             // 当前体力值
	MaxStamina   int32             // 最大体力值
	State        PlayerState       // 玩家状态
	ChunkPos     worldmap.ChunkPos // 所在地图区块位置
	TeamID       int64             // 队伍ID，0表示无队伍
	Buffs        []*Buff           // 增益/减益效果列表（旧版）
	BuffEffects  []*BuffEffect     // 技能效果列表（新版）
	Inventory    *item.Inventory   // 背包系统
	Strength     int32             // 力量属性（影响物理攻击）
	Agility      int32             // 敏捷属性（影响攻速、闪避、暴击）
	Intelligence int32             // 智力属性（影响魔法攻击、MP上限）
	Defense      int32             // 防御属性（减少受到的伤害）
	mu           sync.RWMutex      // 并发访问互斥锁
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

type BuffEffect struct {
	EffectID   int64
	EffectType int32
	Value      float32
	Duration   int
	StackCount int
	MaxStacks  int
	StartTime  int64
	IsDebuff   bool
}

type PlayerManager struct {
	players map[int64]*Player
	mu      sync.RWMutex
}

func NewPlayerManager() *PlayerManager {
	item.LoadDefaultConfigs()
	item.LoadDefaultRecipes()
	return &PlayerManager{
		players: make(map[int64]*Player),
	}
}

func (m *PlayerManager) CreatePlayer(id int64, name string, accountID int64) *Player {
	player := &Player{
		ID:           id,
		Name:         name,
		AccountID:    accountID,
		Level:        1,
		Exp:          0,
		Pos:          worldmap.Vec3{X: 0, Y: 0, Z: 0},
		Rotation:     0,
		Health:       100,
		MaxHealth:    100,
		Mana:         50,
		MaxMana:      50,
		Stamina:      100,
		MaxStamina:   100,
		State:        StateIdle,
		Buffs:        make([]*Buff, 0),
		BuffEffects:  make([]*BuffEffect, 0),
		Inventory:    item.NewInventory(id),
		Strength:     10,
		Agility:      10,
		Intelligence: 10,
		Defense:      5,
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

func (p *Player) GetID() int64 {
	return p.ID
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

func (p *Player) GetDefense() int32 {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.Defense
}

func (p *Player) IsDead() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.State == StateDead
}

func (p *Player) AddBuffEffect(buff *BuffEffect) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, existing := range p.BuffEffects {
		if existing.EffectID == buff.EffectID {
			if existing.StackCount < existing.MaxStacks {
				existing.StackCount++
			}
			existing.StartTime = buff.StartTime
			existing.Duration = buff.Duration
			return
		}
	}

	p.BuffEffects = append(p.BuffEffects, buff)
}

func (p *Player) RemoveBuffEffect(effectID int64) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for i, buff := range p.BuffEffects {
		if buff.EffectID == effectID {
			p.BuffEffects = append(p.BuffEffects[:i], p.BuffEffects[i+1:]...)
			break
		}
	}
}

func (p *Player) HasBuffEffect(effectID int64) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, buff := range p.BuffEffects {
		if buff.EffectID == effectID {
			return true
		}
	}
	return false
}

func (p *Player) GetBuffEffectStacks(effectID int64) int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, buff := range p.BuffEffects {
		if buff.EffectID == effectID {
			return buff.StackCount
		}
	}
	return 0
}

func (p *Player) UpdateBuffEffects() {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now().UnixNano() / 1e6
	newEffects := make([]*BuffEffect, 0)
	for _, buff := range p.BuffEffects {
		if buff.StartTime+int64(buff.Duration) > now {
			newEffects = append(newEffects, buff)
		}
	}
	p.BuffEffects = newEffects
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
	p.mu.Lock()
	newBuffs := make([]*Buff, 0)
	for _, buff := range p.Buffs {
		newBuffs = append(newBuffs, buff)
	}
	p.Buffs = newBuffs
	p.mu.Unlock()
}

func (p *Player) CleanExpiredBuffs() {
	p.mu.Lock()
	now := time.Now().UnixNano() / 1e6
	newBuffs := make([]*Buff, 0)
	for _, buff := range p.Buffs {
		if buff.EndTime > now {
			newBuffs = append(newBuffs, buff)
		}
	}
	p.Buffs = newBuffs
	p.mu.Unlock()
}

func (p *Player) AddItem(itemID int64, count int32) error {
	return p.Inventory.AddItem(itemID, count)
}

func (p *Player) RemoveItem(slot int32, count int32) error {
	return p.Inventory.RemoveItem(slot, count)
}

func (p *Player) UseItem(slot int32) (bool, error) {
	p.mu.Lock()
	health := p.Health
	maxHealth := p.MaxHealth
	mana := p.Mana
	maxMana := p.MaxMana
	strength := p.Strength
	agility := p.Agility
	intelligence := p.Intelligence
	defense := p.Defense
	p.mu.Unlock()

	ctx := &item.EffectContext{
		PlayerID:     p.ID,
		Inventory:    p.Inventory,
		Health:       &health,
		MaxHealth:    &maxHealth,
		Mana:         &mana,
		MaxMana:      &maxMana,
		Strength:     &strength,
		Agility:      &agility,
		Intelligence: &intelligence,
		Defense:      &defense,
	}

	success, err := item.UseItem(p.Inventory, slot, ctx)
	if success {
		p.mu.Lock()
		p.Health = health
		p.Mana = mana
		p.Strength = strength
		p.Agility = agility
		p.Intelligence = intelligence
		p.Defense = defense
		p.mu.Unlock()
	}

	return success, err
}

func (p *Player) EquipItem(slot int32) error {
	return p.Inventory.EquipItem(slot)
}

func (p *Player) UnequipItem(slot item.EquipmentSlot) error {
	return p.Inventory.UnequipItem(slot)
}

func (p *Player) AddGold(amount int64) {
	p.Inventory.AddGold(amount)
}

func (p *Player) RemoveGold(amount int64) error {
	return p.Inventory.RemoveGold(amount)
}

func CalculateExpForLevel(level int32) int64 {
	return int64(100 * math.Pow(1.5, float64(level-1)))
}

func (p *Player) AddExp(exp int64) (bool, int32) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.Exp += exp
	oldLevel := p.Level
	leveledUp := false
	newLevel := oldLevel

	for {
		requiredExp := CalculateExpForLevel(p.Level)
		if p.Exp >= requiredExp {
			p.Exp -= requiredExp
			p.Level++
			newLevel = p.Level
			p.MaxHealth += 20
			p.Health = p.MaxHealth
			p.MaxMana += 10
			p.Mana = p.MaxMana
			p.Strength += 2
			p.Agility += 2
			p.Intelligence += 2
			p.Defense += 1
			leveledUp = true
		} else {
			break
		}
	}

	return leveledUp, newLevel - oldLevel
}

func (p *Player) GetExp() int64 {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.Exp
}

func (p *Player) GetExpProgress() float64 {
	p.mu.RLock()
	defer p.mu.RUnlock()
	required := CalculateExpForLevel(p.Level)
	return float64(p.Exp) / float64(required)
}

func (p *Player) SetPosition(pos worldmap.Vec3) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Pos = pos
}

func (p *Player) SetRotation(rotation float64) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Rotation = rotation
}

func (p *Player) SetLevel(level int32) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Level = level
}

func (p *Player) SetHealth(health int32) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.Health = health
}

func (p *Player) SetMaxHealth(maxHealth int32) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.MaxHealth = maxHealth
}

func (p *Player) SetChunkPos(chunkPos worldmap.ChunkPos) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.ChunkPos = chunkPos
}
