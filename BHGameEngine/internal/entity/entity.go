package entity

import (
	"math"
	"sync"

	"github.com/openworld-server/internal/worldmap"
)

type EntityType int32

const (
	EntityTypeMonster EntityType = 1
	EntityTypeNPC     EntityType = 2
	EntityTypeItem    EntityType = 3
	EntityTypeTrigger EntityType = 4
)

type EntityState int32

const (
	EntityStateIdle      EntityState = 0
	EntityStateMoving    EntityState = 1
	EntityStateAttacking EntityState = 2
	EntityStateDead      EntityState = 3
	EntityStateSleeping  EntityState = 4
)

type Entity struct {
	ID           int64
	Type         EntityType
	Name         string
	Pos          worldmap.Vec3
	Rotation     float64
	Health       int32
	MaxHealth    int32
	State        EntityState
	ChunkPos     worldmap.ChunkPos
	OwnerID      int64
	SpawnTime    int64
	DespawnTime  int64
	AIEnabled    bool
	InterestList map[int64]bool
	mu           sync.RWMutex
}

type Monster struct {
	Entity
	Level          int32
	ExpReward      int64
	DropTableID    int64
	SkillSet       []int32
	PatrolRadius   float64
	TargetID       int64
	LastAttackTime int64
}

type NPC struct {
	Entity
	DialogID    int64
	QuestID     int64
	ShopID      int64
	ServiceType int32
}

type ItemEntity struct {
	Entity
	ItemID       int64
	StackCount   int32
	ExpireTime   int64
	IsPickupable bool
}

type EntityManager struct {
	entities map[int64]*Entity
	monsters map[int64]*Monster
	npc      map[int64]*NPC
	items    map[int64]*ItemEntity
	mu       sync.RWMutex
}

func NewEntityManager() *EntityManager {
	return &EntityManager{
		entities: make(map[int64]*Entity),
		monsters: make(map[int64]*Monster),
		npc:      make(map[int64]*NPC),
		items:    make(map[int64]*ItemEntity),
	}
}

func (m *EntityManager) CreateMonster(id int64, name string, pos worldmap.Vec3, level int32) *Monster {
	monster := &Monster{
		Entity: Entity{
			ID:           id,
			Type:         EntityTypeMonster,
			Name:         name,
			Pos:          pos,
			Rotation:     0,
			Health:       int32(level * 20),
			MaxHealth:    int32(level * 20),
			State:        EntityStateIdle,
			AIEnabled:    true,
			InterestList: make(map[int64]bool),
		},
		Level:        level,
		ExpReward:    int64(level * 10),
		PatrolRadius: 10,
	}

	m.mu.Lock()
	m.entities[id] = &monster.Entity
	m.monsters[id] = monster
	m.mu.Unlock()

	return monster
}

func (m *EntityManager) CreateNPC(id int64, name string, pos worldmap.Vec3, dialogID int64) *NPC {
	npc := &NPC{
		Entity: Entity{
			ID:           id,
			Type:         EntityTypeNPC,
			Name:         name,
			Pos:          pos,
			Rotation:     0,
			Health:       1000,
			MaxHealth:    1000,
			State:        EntityStateIdle,
			AIEnabled:    false,
			InterestList: make(map[int64]bool),
		},
		DialogID: dialogID,
	}

	m.mu.Lock()
	m.entities[id] = &npc.Entity
	m.npc[id] = npc
	m.mu.Unlock()

	return npc
}

func (m *EntityManager) CreateItem(id int64, itemID int64, pos worldmap.Vec3, count int32) *ItemEntity {
	item := &ItemEntity{
		Entity: Entity{
			ID:           id,
			Type:         EntityTypeItem,
			Name:         "",
			Pos:          pos,
			Rotation:     0,
			Health:       1,
			MaxHealth:    1,
			State:        EntityStateIdle,
			AIEnabled:    false,
			InterestList: make(map[int64]bool),
		},
		ItemID:       itemID,
		StackCount:   count,
		IsPickupable: true,
	}

	m.mu.Lock()
	m.entities[id] = &item.Entity
	m.items[id] = item
	m.mu.Unlock()

	return item
}

func (m *EntityManager) GetEntity(id int64) (*Entity, bool) {
	m.mu.RLock()
	entity, ok := m.entities[id]
	m.mu.RUnlock()
	return entity, ok
}

func (m *EntityManager) GetMonster(id int64) (*Monster, bool) {
	m.mu.RLock()
	monster, ok := m.monsters[id]
	m.mu.RUnlock()
	return monster, ok
}

func (m *EntityManager) RemoveEntity(id int64) {
	m.mu.Lock()
	entity, ok := m.entities[id]
	if ok {
		switch entity.Type {
		case EntityTypeMonster:
			delete(m.monsters, id)
		case EntityTypeNPC:
			delete(m.npc, id)
		case EntityTypeItem:
			delete(m.items, id)
		}
		delete(m.entities, id)
	}
	m.mu.Unlock()
}

func (m *EntityManager) UpdatePosition(id int64, pos worldmap.Vec3) {
	m.mu.RLock()
	entity, ok := m.entities[id]
	m.mu.RUnlock()

	if ok {
		entity.mu.Lock()
		entity.Pos = pos
		entity.mu.Unlock()
	}
}

func (m *EntityManager) UpdateState(id int64, state EntityState) {
	m.mu.RLock()
	entity, ok := m.entities[id]
	m.mu.RUnlock()

	if ok {
		entity.mu.Lock()
		entity.State = state
		entity.mu.Unlock()
	}
}

func (m *EntityManager) AddInterest(id int64, playerID int64) {
	m.mu.RLock()
	entity, ok := m.entities[id]
	m.mu.RUnlock()

	if ok {
		entity.mu.Lock()
		entity.InterestList[playerID] = true
		entity.mu.Unlock()
	}
}

func (m *EntityManager) RemoveInterest(id int64, playerID int64) {
	m.mu.RLock()
	entity, ok := m.entities[id]
	m.mu.RUnlock()

	if ok {
		entity.mu.Lock()
		delete(entity.InterestList, playerID)
		entity.mu.Unlock()
	}
}

func (m *EntityManager) GetInterestCount(id int64) int {
	m.mu.RLock()
	entity, ok := m.entities[id]
	m.mu.RUnlock()

	if !ok {
		return 0
	}

	entity.mu.RLock()
	count := len(entity.InterestList)
	entity.mu.RUnlock()
	return count
}

func (m *EntityManager) GetAllMonsters() []*Monster {
	m.mu.RLock()
	defer m.mu.RUnlock()

	monsters := make([]*Monster, 0, len(m.monsters))
	for _, m := range m.monsters {
		monsters = append(monsters, m)
	}
	return monsters
}

func (m *EntityManager) GetEntitiesInRadius(center worldmap.Vec3, radius float64) []*Entity {
	m.mu.RLock()
	defer m.mu.RUnlock()

	entities := make([]*Entity, 0)
	for _, entity := range m.entities {
		entity.mu.RLock()
		dx := entity.Pos.X - center.X
		dy := entity.Pos.Y - center.Y
		dz := entity.Pos.Z - center.Z
		dist := math.Sqrt(dx*dx + dy*dy + dz*dz)
		entity.mu.RUnlock()

		if dist <= radius {
			entities = append(entities, entity)
		}
	}

	return entities
}

func (m *EntityManager) GetEntityCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.entities)
}

func (e *Entity) GetPosition() worldmap.Vec3 {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.Pos
}

func (e *Entity) GetState() EntityState {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.State
}

func (e *Entity) TakeDamage(damage int32) {
	e.mu.Lock()
	e.Health -= damage
	if e.Health <= 0 {
		e.Health = 0
		e.State = EntityStateDead
	}
	e.mu.Unlock()
}

func (e *Entity) IsSleeping() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.State == EntityStateSleeping
}

func (e *Entity) SetSleeping(sleeping bool) {
	e.mu.Lock()
	if sleeping {
		e.State = EntityStateSleeping
	} else if e.State == EntityStateSleeping {
		e.State = EntityStateIdle
	}
	e.mu.Unlock()
}

func (e *Entity) HasInterest(playerID int64) bool {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.InterestList[playerID]
}
