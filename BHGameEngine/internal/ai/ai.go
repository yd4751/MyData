package ai

import (
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/openworld-server/internal/worldmap"
)

type AIType int32

const (
	AITypePassive    AIType = 1
	AITypeAggressive AIType = 2
	AITypePatrol     AIType = 3
	AITypeBoss       AIType = 4
)

type AIState int32

const (
	AIStateIdle   AIState = 1
	AIStatePatrol AIState = 2
	AIStateChase  AIState = 3
	AIStateAttack AIState = 4
	AIStateReturn AIState = 5
)

type AIBehavior struct {
	MonsterID  int64
	Type       AIType
	State      AIState
	HomePos    worldmap.Vec3
	TargetID   int64
	TargetPos  worldmap.Vec3
	MoveSpeed  float64
	LastUpdate int64
	mu         sync.RWMutex
}

type AIManager struct {
	behaviors map[int64]*AIBehavior
	ticker    *time.Ticker
	running   bool
	mu        sync.RWMutex
}

func NewAIManager() *AIManager {
	return &AIManager{
		behaviors: make(map[int64]*AIBehavior),
	}
}

func (m *AIManager) AddBehavior(monsterID int64, aiType AIType, homePos worldmap.Vec3) {
	behavior := &AIBehavior{
		MonsterID:  monsterID,
		Type:       aiType,
		State:      AIStateIdle,
		HomePos:    homePos,
		MoveSpeed:  1.5,
		LastUpdate: time.Now().Unix(),
	}

	m.mu.Lock()
	m.behaviors[monsterID] = behavior
	m.mu.Unlock()
}

func (m *AIManager) RemoveBehavior(monsterID int64) {
	m.mu.Lock()
	delete(m.behaviors, monsterID)
	m.mu.Unlock()
}

func (m *AIManager) GetBehavior(monsterID int64) (*AIBehavior, bool) {
	m.mu.RLock()
	behavior, ok := m.behaviors[monsterID]
	m.mu.RUnlock()
	return behavior, ok
}

func (m *AIManager) Start() {
	m.mu.Lock()
	if m.running {
		m.mu.Unlock()
		return
	}
	m.running = true
	m.ticker = time.NewTicker(100 * time.Millisecond)
	m.mu.Unlock()

	go func() {
		for range m.ticker.C {
			m.Update()
		}
	}()
}

func (m *AIManager) Stop() {
	m.mu.Lock()
	if m.ticker != nil {
		m.ticker.Stop()
	}
	m.running = false
	m.mu.Unlock()
}

func (m *AIManager) Update() {
	m.mu.RLock()
	behaviors := make([]*AIBehavior, 0, len(m.behaviors))
	for _, b := range m.behaviors {
		behaviors = append(behaviors, b)
	}
	m.mu.RUnlock()

	for _, behavior := range behaviors {
		go m.updateBehavior(behavior)
	}
}

func (m *AIManager) updateBehavior(behavior *AIBehavior) {
	behavior.mu.Lock()
	defer behavior.mu.Unlock()

	now := time.Now().Unix()
	if now-behavior.LastUpdate < 500 {
		return
	}
	behavior.LastUpdate = now

	switch behavior.Type {
	case AITypePassive:
		m.updatePassive(behavior)
	case AITypeAggressive:
		m.updateAggressive(behavior)
	case AITypePatrol:
		m.updatePatrol(behavior)
	case AITypeBoss:
		m.updateBoss(behavior)
	}
}

func (m *AIManager) updatePassive(behavior *AIBehavior) {
	behavior.State = AIStateIdle
}

func (m *AIManager) updateAggressive(behavior *AIBehavior) {
	if behavior.TargetID != 0 {
		m.updateChase(behavior)
	} else {
		behavior.State = AIStateIdle
	}
}

func (m *AIManager) updatePatrol(behavior *AIBehavior) {
	switch behavior.State {
	case AIStateIdle:
		m.startPatrol(behavior)
	case AIStatePatrol:
		m.continuePatrol(behavior)
	case AIStateChase:
		m.updateChase(behavior)
	case AIStateReturn:
		m.updateReturn(behavior)
	}
}

func (m *AIManager) updateBoss(behavior *AIBehavior) {
	if behavior.TargetID != 0 {
		m.updateChase(behavior)
	} else {
		behavior.State = AIStateIdle
	}
}

func (m *AIManager) startPatrol(behavior *AIBehavior) {
	rand.Seed(time.Now().UnixNano())
	angle := rand.Float64() * math.Pi * 2
	dist := rand.Float64() * 10

	behavior.TargetPos = worldmap.Vec3{
		X: behavior.HomePos.X + math.Cos(angle)*dist,
		Y: behavior.HomePos.Y + math.Sin(angle)*dist,
		Z: 0,
	}
	behavior.State = AIStatePatrol
}

func (m *AIManager) continuePatrol(behavior *AIBehavior) {
	dx := behavior.TargetPos.X - behavior.HomePos.X
	dy := behavior.TargetPos.Y - behavior.HomePos.Y
	dist := math.Sqrt(dx*dx + dy*dy)

	if dist < 1 {
		behavior.State = AIStateIdle
		return
	}

	if behavior.TargetID != 0 {
		behavior.State = AIStateChase
	}
}

func (m *AIManager) updateChase(behavior *AIBehavior) {
	if behavior.TargetID == 0 {
		behavior.State = AIStateReturn
		return
	}

	dx := behavior.TargetPos.X - behavior.HomePos.X
	dy := behavior.TargetPos.Y - behavior.HomePos.Y
	dist := math.Sqrt(dx*dx + dy*dy)

	if dist > 30 {
		behavior.TargetID = 0
		behavior.State = AIStateReturn
		return
	}

	if dist < 3 {
		behavior.State = AIStateAttack
		return
	}

	behavior.State = AIStateChase
}

func (m *AIManager) updateReturn(behavior *AIBehavior) {
	dx := behavior.HomePos.X - behavior.HomePos.X
	dy := behavior.HomePos.Y - behavior.HomePos.Y
	dist := math.Sqrt(dx*dx + dy*dy)

	if dist < 1 {
		behavior.State = AIStateIdle
		return
	}

	behavior.State = AIStateReturn
}

func (m *AIManager) SetTarget(monsterID int64, targetID int64, targetPos worldmap.Vec3) {
	m.mu.RLock()
	behavior, ok := m.behaviors[monsterID]
	m.mu.RUnlock()

	if ok {
		behavior.mu.Lock()
		behavior.TargetID = targetID
		behavior.TargetPos = targetPos
		if behavior.Type == AITypeAggressive || behavior.Type == AITypeBoss {
			behavior.State = AIStateChase
		}
		behavior.mu.Unlock()
	}
}

func (m *AIManager) ClearTarget(monsterID int64) {
	m.mu.RLock()
	behavior, ok := m.behaviors[monsterID]
	m.mu.RUnlock()

	if ok {
		behavior.mu.Lock()
		behavior.TargetID = 0
		if behavior.Type == AITypePatrol {
			behavior.State = AIStateReturn
		} else {
			behavior.State = AIStateIdle
		}
		behavior.mu.Unlock()
	}
}

func (m *AIManager) UpdatePosition(monsterID int64, pos worldmap.Vec3) {
	m.mu.RLock()
	behavior, ok := m.behaviors[monsterID]
	m.mu.RUnlock()

	if ok {
		behavior.mu.Lock()
		behavior.HomePos = pos
		behavior.mu.Unlock()
	}
}

func (m *AIManager) GetActiveCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	count := 0
	for _, b := range m.behaviors {
		b.mu.RLock()
		if b.State != AIStateIdle {
			count++
		}
		b.mu.RUnlock()
	}
	return count
}
