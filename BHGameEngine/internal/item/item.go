package item

import (
	"sync"
)

type ItemCategory int32 // 物品分类

const (
	ItemCategoryConsumable ItemCategory = 1 // 消耗品
	ItemCategoryEquipment  ItemCategory = 2 // 装备
	ItemCategoryMaterial   ItemCategory = 3 // 材料
	ItemCategoryQuest      ItemCategory = 4 // 任务物品
	ItemCategoryCurrency   ItemCategory = 5 // 货币
)

type ItemSubType int32 // 物品子类型

const (
	SubTypePotion      ItemSubType = 101 // 药水
	SubTypeScroll      ItemSubType = 102 // 卷轴
	SubTypeFood        ItemSubType = 103 // 食物
	SubTypeWeapon      ItemSubType = 201 // 武器
	SubTypeArmor       ItemSubType = 202 // 护甲
	SubTypeAccessory   ItemSubType = 203 // 饰品
	SubTypeBasicMat    ItemSubType = 301 // 基础材料
	SubTypeAdvancedMat ItemSubType = 302 // 高级材料
	SubTypeStory       ItemSubType = 401 // 剧情物品
	SubTypeReputation  ItemSubType = 402 // 声望物品
	SubTypeGold        ItemSubType = 501 // 金币
	SubTypeToken       ItemSubType = 502 // 代币
)

type ItemRarity int32 // 物品稀有度

const (
	RarityCommon    ItemRarity = 1 // 普通(白色)
	RarityUncommon  ItemRarity = 2 // 优秀(绿色)
	RarityRare      ItemRarity = 3 // 稀有(蓝色)
	RarityEpic      ItemRarity = 4 // 史诗(紫色)
	RarityLegendary ItemRarity = 5 // 传说(橙色)
)

type EquipmentSlot int32 // 装备槽位

const (
	SlotWeapon   EquipmentSlot = 1 // 武器槽
	SlotArmor    EquipmentSlot = 2 // 护甲槽
	SlotHelmet   EquipmentSlot = 3 // 头盔槽
	SlotBoots    EquipmentSlot = 4 // 靴子槽
	SlotGloves   EquipmentSlot = 5 // 手套槽
	SlotRing     EquipmentSlot = 6 // 戒指槽
	SlotNecklace EquipmentSlot = 7 // 项链槽
)

type EffectType int32 // 物品效果类型

const (
	EffectHealHP          EffectType = 1 // 恢复生命值
	EffectHealMP          EffectType = 2 // 恢复魔法值
	EffectAddStrength     EffectType = 3 // 增加力量
	EffectAddAgility      EffectType = 4 // 增加敏捷
	EffectAddIntelligence EffectType = 5 // 增加智力
	EffectAddDefense      EffectType = 6 // 增加防御
	EffectTempBuff        EffectType = 7 // 临时增益
)

type ItemConfig struct {
	ID          int64        // 物品唯一ID
	Name        string       // 物品名称
	Category    ItemCategory // 物品分类
	SubType     ItemSubType  // 物品子类型
	Rarity      ItemRarity   // 稀有度
	Level       int32        // 物品等级
	MaxStack    int32        // 最大堆叠数量
	Description string       // 物品描述
	Icon        string       // 图标路径

	Attack       int32 // 攻击力加成
	Defense      int32 // 防御力加成
	Strength     int32 // 力量加成
	Agility      int32 // 敏捷加成
	Intelligence int32 // 智力加成

	EffectType  EffectType // 效果类型
	EffectValue int32      // 效果数值
	Cooldown    int32      // 使用冷却时间(秒)
	Duration    int32      // 效果持续时间(秒)

	RequiredLevel int32         // 装备需求等级
	EquipmentSlot EquipmentSlot // 装备槽位

	SellPrice int64 // 出售价格
	BuyPrice  int64 // 购买价格
}

type InventoryItem struct {
	ItemID     int64  // 物品ID
	Slot       int32  // 背包槽位
	Count      int32  // 数量
	Level      int32  // 物品强化等级
	ExpireTime int64  // 过期时间戳
	UID        string // 物品唯一标识
}

type CooldownEntry struct {
	ItemID  int64 // 物品ID
	EndTime int64 // 冷却结束时间戳
}

type Equipment struct {
	Slot   EquipmentSlot // 装备槽位
	ItemID int64         // 物品ID
	Level  int32         // 强化等级
}

var (
	itemConfigs  = make(map[int64]*ItemConfig) // 物品配置表
	itemConfigMu sync.RWMutex                  // 配置表并发锁
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
