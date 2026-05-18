package skill

import "errors"

var (
	ErrSkillNotFound     = errors.New("skill not found")
	ErrSkillNotLearned   = errors.New("skill not learned")
	ErrSkillMaxLevel     = errors.New("skill already at max level")
	ErrSkillOnCooldown   = errors.New("skill is on cooldown")
	ErrInsufficientMana  = errors.New("insufficient mana")
	ErrCannotCastPassive = errors.New("cannot cast passive skill")
	ErrInvalidTarget     = errors.New("invalid target")
	ErrTargetOutOfRange  = errors.New("target out of range")
	ErrNoLineOfSight     = errors.New("no line of sight to target")
)

type SkillError struct {
	Code    int
	Message string
}

func (e *SkillError) Error() string {
	return e.Message
}

func NewSkillError(code int, message string) *SkillError {
	return &SkillError{
		Code:    code,
		Message: message,
	}
}
