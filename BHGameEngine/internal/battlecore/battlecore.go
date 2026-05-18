package battlecore

import (
	"math"
	"sync"
	"time"

	"github.com/openworld-server/internal/entity"
	"github.com/openworld-server/internal/player"
	"github.com/openworld-server/internal/worldmap"
)

type Battle struct {
	ID        int64
	Players   map[int64]*player.Player
	Monsters  map[int64]*entity.Monster
	StartTime int64
	EndTime   int64
	Active    bool
	mu        sync.RWMutex
}

type Skill struct {
	ID       int32
	Name     string
	Damage   int32
	Range    float64
	CastTime float64
	Cooldown float64
	LastCast int64
}

type CombatLog struct {
	Timestamp int64
	Attacker  int64
	Target    int64
	SkillID   int32
	Damage    int32
}

type BattleManager struct {
	battles map[int64]*Battle
	mu      sync.RWMutex
}

func NewBattleManager() *BattleManager {
	return &BattleManager{
		battles: make(map[int64]*Battle),
	}
}

func (m *BattleManager) CreateBattle(battleID int64) *Battle {
	battle := &Battle{
		ID:        battleID,
		Players:   make(map[int64]*player.Player),
		Monsters:  make(map[int64]*entity.Monster),
		StartTime: time.Now().Unix(),
		Active:    true,
	}

	m.mu.Lock()
	m.battles[battleID] = battle
	m.mu.Unlock()

	return battle
}

func (m *BattleManager) GetBattle(battleID int64) (*Battle, bool) {
	m.mu.RLock()
	battle, ok := m.battles[battleID]
	m.mu.RUnlock()
	return battle, ok
}

func (m *BattleManager) EndBattle(battleID int64) {
	m.mu.Lock()
	battle, ok := m.battles[battleID]
	if ok {
		battle.mu.Lock()
		battle.Active = false
		battle.EndTime = time.Now().Unix()
		battle.mu.Unlock()
		delete(m.battles, battleID)
	}
	m.mu.Unlock()
}

func (b *Battle) AddPlayer(p *player.Player) {
	b.mu.Lock()
	b.Players[p.ID] = p
	b.mu.Unlock()
}

func (b *Battle) RemovePlayer(playerID int64) {
	b.mu.Lock()
	delete(b.Players, playerID)
	if len(b.Players) == 0 && len(b.Monsters) == 0 {
		b.Active = false
	}
	b.mu.Unlock()
}

func (b *Battle) AddMonster(m *entity.Monster) {
	b.mu.Lock()
	b.Monsters[m.ID] = m
	b.mu.Unlock()
}

func (b *Battle) RemoveMonster(monsterID int64) {
	b.mu.Lock()
	delete(b.Monsters, monsterID)
	if len(b.Players) == 0 && len(b.Monsters) == 0 {
		b.Active = false
	}
	b.mu.Unlock()
}

func (b *Battle) IsActive() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.Active
}

func CalculateDamage(attacker *player.Player, target *entity.Monster, skill *Skill) int32 {
	baseDamage := skill.Damage
	levelBonus := int32(attacker.GetLevel()) * 2
	weaponBonus := int32(10)
	total := baseDamage + levelBonus + weaponBonus

	armor := int32(target.Level * 5)
	damage := total - armor/2
	if damage < 1 {
		damage = 1
	}

	return damage
}

func CalculateHeal(player *player.Player, amount int32) int32 {
	player.Heal(amount)
	return amount
}

func IsInRange(pos1, pos2 worldmap.Vec3, r float64) bool {
	dx := pos1.X - pos2.X
	dy := pos1.Y - pos2.Y
	dz := pos1.Z - pos2.Z
	dist := math.Sqrt(dx*dx + dy*dy + dz*dz)
	return dist <= r
}

func CheckCooldown(skill *Skill) bool {
	now := time.Now().UnixNano() / 1e6
	return now-skill.LastCast >= int64(skill.Cooldown*1000)
}

func ApplyBuff(target *player.Player, buff *player.Buff) {
	target.ApplyBuff(buff)
}

func RemoveBuff(target *player.Player, buffID int32) {
	target.RemoveBuff(buffID)
}

func UpdateBuffs(target *player.Player) {
	target.UpdateBuffs()
}

func ProcessSkill(attacker *player.Player, target *entity.Monster, skill *Skill) *CombatLog {
	if !IsInRange(attacker.GetPosition(), target.GetPosition(), skill.Range) {
		return nil
	}

	if !CheckCooldown(skill) {
		return nil
	}

	damage := CalculateDamage(attacker, target, skill)
	target.TakeDamage(damage)
	skill.LastCast = time.Now().UnixNano() / 1e6

	log := &CombatLog{
		Timestamp: time.Now().Unix(),
		Attacker:  attacker.ID,
		Target:    target.ID,
		SkillID:   skill.ID,
		Damage:    damage,
	}

	return log
}

func ProcessMonsterAttack(monster *entity.Monster, player *player.Player) *CombatLog {
	now := time.Now().Unix()
	if now-monster.LastAttackTime < 1000 {
		return nil
	}

	if !IsInRange(monster.GetPosition(), player.GetPosition(), 3) {
		return nil
	}

	damage := int32(monster.Level * 5)
	player.TakeDamage(damage)
	monster.LastAttackTime = now

	log := &CombatLog{
		Timestamp: now,
		Attacker:  monster.ID,
		Target:    player.ID,
		SkillID:   0,
		Damage:    damage,
	}

	return log
}
