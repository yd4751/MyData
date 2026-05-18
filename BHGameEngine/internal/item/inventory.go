package item

import (
	"errors"
	"sync"
	"time"
)

const (
	DefaultInventorySize = 100 // 默认背包容量
	MaxEquipSlots        = 7   // 最大装备槽位数量
)

// Inventory 背包系统
type Inventory struct {
	OwnerID    int64                        // 所有者ID
	Items      map[int32]*InventoryItem     // 物品列表（槽位->物品）
	Equipments map[EquipmentSlot]*Equipment // 装备列表（槽位->装备）
	Cooldowns  map[int64]*CooldownEntry     // 物品冷却列表（物品ID->冷却记录）
	Gold       int64                        // 金币数量
	Capacity   int32                        // 背包容量
	mu         sync.RWMutex                 // 并发访问锁
}

func NewInventory(ownerID int64) *Inventory {
	return &Inventory{
		OwnerID:    ownerID,
		Items:      make(map[int32]*InventoryItem),
		Equipments: make(map[EquipmentSlot]*Equipment),
		Cooldowns:  make(map[int64]*CooldownEntry),
		Gold:       0,
		Capacity:   DefaultInventorySize,
	}
}

func (inv *Inventory) AddItem(itemID int64, count int32) error {
	config, ok := GetItemConfig(itemID)
	if !ok {
		return errors.New("item config not found")
	}

	inv.mu.Lock()
	defer inv.mu.Unlock()

	if config.Category == ItemCategoryCurrency {
		if config.SubType == SubTypeGold {
			inv.Gold += int64(count)
			return nil
		}
	}

	for _, item := range inv.Items {
		if item.ItemID == itemID && item.Count < config.MaxStack {
			item.Count += count
			return nil
		}
	}

	for slot := int32(0); slot < inv.Capacity; slot++ {
		if _, ok := inv.Items[slot]; !ok {
			inv.Items[slot] = &InventoryItem{
				ItemID: itemID,
				Slot:   slot,
				Count:  count,
				UID:    generateUID(),
			}
			return nil
		}
	}

	return errors.New("inventory is full")
}

func (inv *Inventory) RemoveItem(slot int32, count int32) error {
	inv.mu.Lock()
	defer inv.mu.Unlock()

	item, ok := inv.Items[slot]
	if !ok {
		return errors.New("item not found")
	}

	if item.Count < count {
		return errors.New("not enough items")
	}

	item.Count -= count
	if item.Count <= 0 {
		delete(inv.Items, slot)
	}

	return nil
}

func (inv *Inventory) GetItem(slot int32) (*InventoryItem, bool) {
	inv.mu.RLock()
	defer inv.mu.RUnlock()

	item, ok := inv.Items[slot]
	return item, ok
}

func (inv *Inventory) GetAllItems() map[int32]*InventoryItem {
	inv.mu.RLock()
	defer inv.mu.RUnlock()

	result := make(map[int32]*InventoryItem)
	for slot, item := range inv.Items {
		result[slot] = item
	}
	return result
}

func (inv *Inventory) EquipItem(slot int32) error {
	item, ok := inv.GetItem(slot)
	if !ok {
		return errors.New("item not found")
	}

	config, ok := GetItemConfig(item.ItemID)
	if !ok {
		return errors.New("item config not found")
	}

	if config.Category != ItemCategoryEquipment {
		return errors.New("not an equipment item")
	}

	inv.mu.Lock()
	defer inv.mu.Unlock()

	currentEquip, exists := inv.Equipments[config.EquipmentSlot]
	if exists {
		inv.Items[slot] = &InventoryItem{
			ItemID: currentEquip.ItemID,
			Slot:   slot,
			Count:  1,
			Level:  currentEquip.Level,
			UID:    generateUID(),
		}
	} else {
		delete(inv.Items, slot)
	}

	inv.Equipments[config.EquipmentSlot] = &Equipment{
		Slot:   config.EquipmentSlot,
		ItemID: item.ItemID,
		Level:  item.Level,
	}

	return nil
}

func (inv *Inventory) UnequipItem(slot EquipmentSlot) error {
	inv.mu.Lock()
	defer inv.mu.Unlock()

	equip, ok := inv.Equipments[slot]
	if !ok {
		return errors.New("no equipment in this slot")
	}

	for emptySlot := int32(0); emptySlot < inv.Capacity; emptySlot++ {
		if _, ok := inv.Items[emptySlot]; !ok {
			inv.Items[emptySlot] = &InventoryItem{
				ItemID: equip.ItemID,
				Slot:   emptySlot,
				Count:  1,
				Level:  equip.Level,
				UID:    generateUID(),
			}
			delete(inv.Equipments, slot)
			return nil
		}
	}

	return errors.New("inventory is full")
}

func (inv *Inventory) GetEquipment(slot EquipmentSlot) (*Equipment, bool) {
	inv.mu.RLock()
	defer inv.mu.RUnlock()

	equip, ok := inv.Equipments[slot]
	return equip, ok
}

func (inv *Inventory) GetAllEquipments() map[EquipmentSlot]*Equipment {
	inv.mu.RLock()
	defer inv.mu.RUnlock()

	result := make(map[EquipmentSlot]*Equipment)
	for slot, equip := range inv.Equipments {
		result[slot] = equip
	}
	return result
}

func (inv *Inventory) IsOnCooldown(itemID int64) bool {
	inv.mu.RLock()
	defer inv.mu.RUnlock()

	cooldown, ok := inv.Cooldowns[itemID]
	if !ok {
		return false
	}

	return cooldown.EndTime > time.Now().Unix()
}

func (inv *Inventory) SetCooldown(itemID int64, duration int32) {
	inv.mu.Lock()
	defer inv.mu.Unlock()

	inv.Cooldowns[itemID] = &CooldownEntry{
		ItemID:  itemID,
		EndTime: time.Now().Unix() + int64(duration),
	}
}

func (inv *Inventory) AddGold(amount int64) {
	inv.mu.Lock()
	inv.Gold += amount
	inv.mu.Unlock()
}

func (inv *Inventory) RemoveGold(amount int64) error {
	inv.mu.Lock()
	defer inv.mu.Unlock()

	if inv.Gold < amount {
		return errors.New("not enough gold")
	}

	inv.Gold -= amount
	return nil
}

func (inv *Inventory) GetGold() int64 {
	inv.mu.RLock()
	defer inv.mu.RUnlock()
	return inv.Gold
}

func generateUID() string {
	return time.Now().Format("20060102150405") + "_" + string(rune(time.Now().UnixNano()%10000))
}
