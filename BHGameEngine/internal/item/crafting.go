package item

import (
	"errors"
	"sync"
)

type Recipe struct {
	ID            int64
	Name          string
	ResultItem    int64
	ResultCount   int32
	Materials     map[int64]int32
	RequiredSkill int32
}

var (
	recipes   = make(map[int64]*Recipe)
	recipesMu sync.RWMutex
)

func RegisterRecipe(recipe *Recipe) {
	recipesMu.Lock()
	recipes[recipe.ID] = recipe
	recipesMu.Unlock()
}

func GetRecipe(recipeID int64) (*Recipe, bool) {
	recipesMu.RLock()
	recipe, ok := recipes[recipeID]
	recipesMu.RUnlock()
	return recipe, ok
}

func GetAllRecipes() []*Recipe {
	recipesMu.RLock()
	defer recipesMu.RUnlock()

	result := make([]*Recipe, 0, len(recipes))
	for _, recipe := range recipes {
		result = append(result, recipe)
	}
	return result
}

func CanCraft(inv *Inventory, recipeID int64) (bool, error) {
	recipe, ok := GetRecipe(recipeID)
	if !ok {
		return false, errors.New("recipe not found")
	}

	inv.mu.RLock()
	defer inv.mu.RUnlock()

	for materialID, requiredCount := range recipe.Materials {
		found := int32(0)
		for _, item := range inv.Items {
			if item.ItemID == materialID {
				found += item.Count
			}
		}
		if found < requiredCount {
			return false, nil
		}
	}

	return true, nil
}

func Craft(inv *Inventory, recipeID int64) (bool, error) {
	recipe, ok := GetRecipe(recipeID)
	if !ok {
		return false, errors.New("recipe not found")
	}

	canCraft, err := CanCraft(inv, recipeID)
	if err != nil {
		return false, err
	}
	if !canCraft {
		return false, errors.New("not enough materials")
	}

	inv.mu.Lock()
	defer inv.mu.Unlock()

	for materialID, count := range recipe.Materials {
		remaining := count
		for _, item := range inv.Items {
			if item.ItemID == materialID && remaining > 0 {
				if item.Count >= remaining {
					item.Count -= remaining
					if item.Count <= 0 {
						delete(inv.Items, item.Slot)
					}
					remaining = 0
				} else {
					remaining -= item.Count
					delete(inv.Items, item.Slot)
				}
			}
		}
	}

	err = inv.AddItem(recipe.ResultItem, recipe.ResultCount)
	if err != nil {
		return false, err
	}

	return true, nil
}

func LoadDefaultRecipes() {
	RegisterRecipe(&Recipe{
		ID:          10001,
		Name:        "铁剑制作",
		ResultItem:  4001,
		ResultCount: 1,
		Materials: map[int64]int32{
			5001: 5,
			5002: 2,
		},
		RequiredSkill: 1,
	})

	RegisterRecipe(&Recipe{
		ID:          10002,
		Name:        "精钢剑制作",
		ResultItem:  4002,
		ResultCount: 1,
		Materials: map[int64]int32{
			5001: 10,
			5003: 2,
		},
		RequiredSkill: 10,
	})

	RegisterRecipe(&Recipe{
		ID:          10003,
		Name:        "小型生命药水制作",
		ResultItem:  1001,
		ResultCount: 5,
		Materials: map[int64]int32{
			5002: 3,
		},
		RequiredSkill: 1,
	})

	RegisterRecipe(&Recipe{
		ID:          10004,
		Name:        "力量卷轴制作",
		ResultItem:  3001,
		ResultCount: 1,
		Materials: map[int64]int32{
			5002: 2,
			5003: 1,
		},
		RequiredSkill: 5,
	})
}
