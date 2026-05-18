package item

import (
	"sync"
)

type ItemCategory int32

const (
	ItemCategoryConsumable ItemCategory = 1
	ItemCategoryEquipment  ItemCategory = 2
	ItemCategoryMaterial   ItemCategory = 3
	ItemCategoryQuest      ItemCategory = 4
	ItemCategoryCurrency   ItemCategory = 5
)

type ItemSubType int32

const (
	SubTypePotion      ItemSubType = 101
	SubTypeScroll      ItemSubType = 102
	SubTypeFood        ItemSubType = 103
	SubTypeWeapon      ItemSubType = 201
	SubTypeArmor       ItemSubType = 202
	SubTypeAccessory   ItemSubType = 203
	SubTypeBasicMat    ItemSubType = 301
	SubTypeAdvancedMat ItemSubType = 302
	SubTypeStory       ItemSubType = 401
	SubTypeReputation  ItemSubType = 402
	SubTypeGold        ItemSubType = 501
	SubTypeToken       ItemSubType = 502
)

type ItemRarity int32

const (
	RarityCommon    ItemRarity = 1
	RarityUncommon  ItemRarity = 2
	RarityRare      ItemRarity = 3
	RarityEpic      ItemRarity = 4
	RarityLegendary ItemRarity = 5
)

type EquipmentSlot int32

const (
	SlotWeapon   EquipmentSlot = 1
	SlotArmor    EquipmentSlot = 2
	SlotHelmet   EquipmentSlot = 3
	SlotBoots    EquipmentSlot = 4
	SlotGloves   EquipmentSlot = 5
	SlotRing     EquipmentSlot = 6
	SlotNecklace EquipmentSlot = 7
)

type EffectType int32

const (
	EffectHealHP          EffectType = 1
	EffectHealMP          EffectType = 2
	EffectAddStrength     EffectType = 3
	EffectAddAgility      EffectType = 4
	EffectAddIntelligence EffectType = 5
	EffectAddDefense      EffectType = 6
	EffectTempBuff        EffectType = 7
)

type ItemConfig struct {
	ID          int64
	Name        string
	Category    ItemCategory
	SubType     ItemSubType
	Rarity      ItemRarity
	Level       int32
	MaxStack    int32
	Description string
	Icon        string

	Attack       int32
	Defense      int32
	Strength     int32
	Agility      int32
	Intelligence int32

	EffectType  EffectType
	EffectValue int32
	Cooldown    int32
	Duration    int32

	RequiredLevel int32
	EquipmentSlot EquipmentSlot

	SellPrice int64
	BuyPrice  int64
}

type InventoryItem struct {
	ItemID     int64
	Slot       int32
	Count      int32
	Level      int32
	ExpireTime int64
	UID        string
}

type CooldownEntry struct {
	ItemID  int64
	EndTime int64
}

type Equipment struct {
	Slot   EquipmentSlot
	ItemID int64
	Level  int32
}

var (
	itemConfigs  = make(map[int64]*ItemConfig)
	itemConfigMu sync.RWMutex
)

func RegisterItemConfig(config *ItemConfig) {
	itemConfigMu.Lock()
	itemConfigs[config.ID] = config
	itemConfigMu.Unlock()
}

func GetItemConfig(itemID int64) (*ItemConfig, bool) {
	itemConfigMu.RLock()
	config, ok := itemConfigs[itemID]
	itemConfigMu.RUnlock()
	return config, ok
}

func LoadDefaultConfigs() {
	RegisterItemConfig(&ItemConfig{
		ID:          1001,
		Name:        "小型生命药水",
		Category:    ItemCategoryConsumable,
		SubType:     SubTypePotion,
		Rarity:      RarityCommon,
		MaxStack:    20,
		Description: "恢复50点生命值",
		EffectType:  EffectHealHP,
		EffectValue: 50,
		SellPrice:   5,
		BuyPrice:    10,
	})

	RegisterItemConfig(&ItemConfig{
		ID:          1002,
		Name:        "中型生命药水",
		Category:    ItemCategoryConsumable,
		SubType:     SubTypePotion,
		Rarity:      RarityCommon,
		MaxStack:    20,
		Description: "恢复100点生命值",
		EffectType:  EffectHealHP,
		EffectValue: 100,
		SellPrice:   10,
		BuyPrice:    20,
	})

	RegisterItemConfig(&ItemConfig{
		ID:          1003,
		Name:        "大型生命药水",
		Category:    ItemCategoryConsumable,
		SubType:     SubTypePotion,
		Rarity:      RarityUncommon,
		MaxStack:    20,
		Description: "恢复200点生命值",
		EffectType:  EffectHealHP,
		EffectValue: 200,
		SellPrice:   20,
		BuyPrice:    40,
	})

	RegisterItemConfig(&ItemConfig{
		ID:          2001,
		Name:        "小型魔法药水",
		Category:    ItemCategoryConsumable,
		SubType:     SubTypePotion,
		Rarity:      RarityCommon,
		MaxStack:    20,
		Description: "恢复30点魔法值",
		EffectType:  EffectHealMP,
		EffectValue: 30,
		SellPrice:   6,
		BuyPrice:    12,
	})

	RegisterItemConfig(&ItemConfig{
		ID:          2002,
		Name:        "中型魔法药水",
		Category:    ItemCategoryConsumable,
		SubType:     SubTypePotion,
		Rarity:      RarityCommon,
		MaxStack:    20,
		Description: "恢复60点魔法值",
		EffectType:  EffectHealMP,
		EffectValue: 60,
		SellPrice:   12,
		BuyPrice:    24,
	})

	RegisterItemConfig(&ItemConfig{
		ID:          3001,
		Name:        "力量卷轴",
		Category:    ItemCategoryConsumable,
		SubType:     SubTypeScroll,
		Rarity:      RarityUncommon,
		MaxStack:    10,
		Description: "5分钟内力量+10",
		EffectType:  EffectAddStrength,
		EffectValue: 10,
		Cooldown:    300,
		Duration:    300,
		SellPrice:   25,
		BuyPrice:    50,
	})

	RegisterItemConfig(&ItemConfig{
		ID:          3002,
		Name:        "敏捷卷轴",
		Category:    ItemCategoryConsumable,
		SubType:     SubTypeScroll,
		Rarity:      RarityUncommon,
		MaxStack:    10,
		Description: "5分钟内敏捷+10",
		EffectType:  EffectAddAgility,
		EffectValue: 10,
		Cooldown:    300,
		Duration:    300,
		SellPrice:   25,
		BuyPrice:    50,
	})

	RegisterItemConfig(&ItemConfig{
		ID:            4001,
		Name:          "铁剑",
		Category:      ItemCategoryEquipment,
		SubType:       SubTypeWeapon,
		Rarity:        RarityCommon,
		Level:         1,
		MaxStack:      1,
		Description:   "普通的铁剑",
		Attack:        10,
		RequiredLevel: 1,
		EquipmentSlot: SlotWeapon,
		SellPrice:     20,
		BuyPrice:      50,
	})

	RegisterItemConfig(&ItemConfig{
		ID:            4002,
		Name:          "精钢剑",
		Category:      ItemCategoryEquipment,
		SubType:       SubTypeWeapon,
		Rarity:        RarityUncommon,
		Level:         10,
		MaxStack:      1,
		Description:   "精良的精钢剑",
		Attack:        25,
		Strength:      5,
		RequiredLevel: 10,
		EquipmentSlot: SlotWeapon,
		SellPrice:     100,
		BuyPrice:      250,
	})

	RegisterItemConfig(&ItemConfig{
		ID:            4003,
		Name:          "皮甲",
		Category:      ItemCategoryEquipment,
		SubType:       SubTypeArmor,
		Rarity:        RarityCommon,
		Level:         1,
		MaxStack:      1,
		Description:   "普通的皮甲",
		Defense:       8,
		RequiredLevel: 1,
		EquipmentSlot: SlotArmor,
		SellPrice:     15,
		BuyPrice:      40,
	})

	RegisterItemConfig(&ItemConfig{
		ID:            4004,
		Name:          "铁头盔",
		Category:      ItemCategoryEquipment,
		SubType:       SubTypeArmor,
		Rarity:        RarityCommon,
		Level:         5,
		MaxStack:      1,
		Description:   "坚固的铁头盔",
		Defense:       12,
		RequiredLevel: 5,
		EquipmentSlot: SlotHelmet,
		SellPrice:     30,
		BuyPrice:      80,
	})

	RegisterItemConfig(&ItemConfig{
		ID:          5001,
		Name:        "铁矿",
		Category:    ItemCategoryMaterial,
		SubType:     SubTypeBasicMat,
		Rarity:      RarityCommon,
		MaxStack:    100,
		Description: "常见的铁矿石",
		SellPrice:   2,
		BuyPrice:    5,
	})

	RegisterItemConfig(&ItemConfig{
		ID:          5002,
		Name:        "木材",
		Category:    ItemCategoryMaterial,
		SubType:     SubTypeBasicMat,
		Rarity:      RarityCommon,
		MaxStack:    100,
		Description: "普通的木材",
		SellPrice:   1,
		BuyPrice:    3,
	})

	RegisterItemConfig(&ItemConfig{
		ID:          5003,
		Name:        "红宝石",
		Category:    ItemCategoryMaterial,
		SubType:     SubTypeAdvancedMat,
		Rarity:      RarityRare,
		MaxStack:    20,
		Description: "珍贵的红宝石",
		SellPrice:   50,
		BuyPrice:    120,
	})

	RegisterItemConfig(&ItemConfig{
		ID:          6001,
		Name:        "神秘钥匙",
		Category:    ItemCategoryQuest,
		SubType:     SubTypeStory,
		Rarity:      RarityUncommon,
		MaxStack:    1,
		Description: "开启神秘之门的钥匙",
	})

	RegisterItemConfig(&ItemConfig{
		ID:          6002,
		Name:        "勇士勋章",
		Category:    ItemCategoryQuest,
		SubType:     SubTypeReputation,
		Rarity:      RarityRare,
		MaxStack:    10,
		Description: "提升声望的勋章",
	})

	RegisterItemConfig(&ItemConfig{
		ID:          7001,
		Name:        "金币",
		Category:    ItemCategoryCurrency,
		SubType:     SubTypeGold,
		Rarity:      RarityCommon,
		MaxStack:    999999,
		Description: "游戏内通用货币",
	})

	RegisterItemConfig(&ItemConfig{
		ID:          7002,
		Name:        "荣誉代币",
		Category:    ItemCategoryCurrency,
		SubType:     SubTypeToken,
		Rarity:      RarityUncommon,
		MaxStack:    9999,
		Description: "竞技场荣誉代币",
	})
}
