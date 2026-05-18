package skill

type SkillResult struct {
	SkillID   int64
	TargetID  int64
	Effects   []*EffectResult
	IsCombo   bool
	ComboName string
}

type EffectResult struct {
	EffectID   int64
	EffectType EffectType
	Value      float32
	Duration   int
}

type SkillInfo struct {
	SkillID     int64
	SkillName   string
	SkillType   SkillType
	SkillClass  SkillClass
	Level       int
	MaxLevel    int
	Cooldown    int
	CooldownEnd int64
	ManaCost    int
	Range       float32
	TargetType  TargetType
	Description string
}

type PlayerSkillInfo struct {
	SkillID     int64
	SkillName   string
	Level       int
	MaxLevel    int
	CooldownEnd int64
	IsUnlocked  bool
}
