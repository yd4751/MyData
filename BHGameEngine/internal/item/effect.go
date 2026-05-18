package item

import (
	"errors"
	"time"
)

type EffectContext struct {
	PlayerID     int64
	Inventory    *Inventory
	Health       *int32
	MaxHealth    *int32
	Mana         *int32
	MaxMana      *int32
	Strength     *int32
	Agility      *int32
	Intelligence *int32
	Defense      *int32
}

func UseItem(inv *Inventory, slot int32, ctx *EffectContext) (bool, error) {
	item, ok := inv.GetItem(slot)
	if !ok {
		return false, errors.New("item not found")
	}

	config, ok := GetItemConfig(item.ItemID)
	if !ok {
		return false, errors.New("item config not found")
	}

	if config.Category != ItemCategoryConsumable {
		return false, errors.New("item is not consumable")
	}

	if inv.IsOnCooldown(item.ItemID) {
		return false, errors.New("item is on cooldown")
	}

	if config.Cooldown > 0 {
		inv.SetCooldown(item.ItemID, config.Cooldown)
	}

	err := applyEffect(config, ctx)
	if err != nil {
		return false, err
	}

	err = inv.RemoveItem(slot, 1)
	if err != nil {
		return false, err
	}

	return true, nil
}

func applyEffect(config *ItemConfig, ctx *EffectContext) error {
	switch config.EffectType {
	case EffectHealHP:
		if ctx.Health == nil || ctx.MaxHealth == nil {
			return errors.New("health not available")
		}
		*ctx.Health += config.EffectValue
		if *ctx.Health > *ctx.MaxHealth {
			*ctx.Health = *ctx.MaxHealth
		}
	case EffectHealMP:
		if ctx.Mana == nil || ctx.MaxMana == nil {
			return errors.New("mana not available")
		}
		*ctx.Mana += config.EffectValue
		if *ctx.Mana > *ctx.MaxMana {
			*ctx.Mana = *ctx.MaxMana
		}
	case EffectAddStrength:
		if ctx.Strength == nil {
			return errors.New("strength not available")
		}
		*ctx.Strength += config.EffectValue
		if config.Duration > 0 {
			go removeStrengthAfterDelay(ctx.Strength, config.EffectValue, config.Duration)
		}
	case EffectAddAgility:
		if ctx.Agility == nil {
			return errors.New("agility not available")
		}
		*ctx.Agility += config.EffectValue
		if config.Duration > 0 {
			go removeAgilityAfterDelay(ctx.Agility, config.EffectValue, config.Duration)
		}
	case EffectAddIntelligence:
		if ctx.Intelligence == nil {
			return errors.New("intelligence not available")
		}
		*ctx.Intelligence += config.EffectValue
		if config.Duration > 0 {
			go removeIntelligenceAfterDelay(ctx.Intelligence, config.EffectValue, config.Duration)
		}
	case EffectAddDefense:
		if ctx.Defense == nil {
			return errors.New("defense not available")
		}
		*ctx.Defense += config.EffectValue
		if config.Duration > 0 {
			go removeDefenseAfterDelay(ctx.Defense, config.EffectValue, config.Duration)
		}
	default:
		return errors.New("unknown effect type")
	}

	return nil
}

func removeStrengthAfterDelay(strength *int32, value int32, duration int32) {
	time.Sleep(time.Duration(duration) * time.Second)
	*strength -= value
}

func removeAgilityAfterDelay(agility *int32, value int32, duration int32) {
	time.Sleep(time.Duration(duration) * time.Second)
	*agility -= value
}

func removeIntelligenceAfterDelay(intelligence *int32, value int32, duration int32) {
	time.Sleep(time.Duration(duration) * time.Second)
	*intelligence -= value
}

func removeDefenseAfterDelay(defense *int32, value int32, duration int32) {
	time.Sleep(time.Duration(duration) * time.Second)
	*defense -= value
}
