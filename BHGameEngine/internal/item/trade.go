package item

import (
	"errors"
	"sync"
	"time"
)

type TradeState int32

const (
	TradeStatePending   TradeState = 1
	TradeStateAccepted  TradeState = 2
	TradeStateCompleted TradeState = 3
	TradeStateCancelled TradeState = 4
)

type TradeItem struct {
	ItemID int64
	Count  int32
}

type Trade struct {
	ID            int64
	Player1ID     int64
	Player2ID     int64
	Player1Items  []*TradeItem
	Player2Items  []*TradeItem
	Player1Gold   int64
	Player2Gold   int64
	Player1Accept bool
	Player2Accept bool
	State         TradeState
	CreatedAt     int64
	ExpireAt      int64
	mu            sync.RWMutex
}

type TradeManager struct {
	trades map[int64]*Trade
	mu     sync.RWMutex
}

var tradeManager *TradeManager

func init() {
	tradeManager = &TradeManager{
		trades: make(map[int64]*Trade),
	}
}

func GetTradeManager() *TradeManager {
	return tradeManager
}

func (tm *TradeManager) CreateTrade(player1ID, player2ID int64) (*Trade, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	now := time.Now().Unix()
	trade := &Trade{
		ID:            now,
		Player1ID:     player1ID,
		Player2ID:     player2ID,
		Player1Items:  make([]*TradeItem, 0),
		Player2Items:  make([]*TradeItem, 0),
		Player1Accept: false,
		Player2Accept: false,
		State:         TradeStatePending,
		CreatedAt:     now,
		ExpireAt:      now + 120,
	}

	tm.trades[trade.ID] = trade
	return trade, nil
}

func (tm *TradeManager) GetTrade(tradeID int64) (*Trade, bool) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	trade, ok := tm.trades[tradeID]
	return trade, ok
}

func (tm *TradeManager) RemoveTrade(tradeID int64) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	delete(tm.trades, tradeID)
}

func (t *Trade) AddItem(playerID int64, itemID int64, count int32) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.State != TradeStatePending {
		return errors.New("trade is not in pending state")
	}

	if playerID == t.Player1ID {
		t.Player1Items = append(t.Player1Items, &TradeItem{ItemID: itemID, Count: count})
	} else if playerID == t.Player2ID {
		t.Player2Items = append(t.Player2Items, &TradeItem{ItemID: itemID, Count: count})
	} else {
		return errors.New("player not in trade")
	}

	return nil
}

func (t *Trade) SetGold(playerID int64, gold int64) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.State != TradeStatePending {
		return errors.New("trade is not in pending state")
	}

	if playerID == t.Player1ID {
		t.Player1Gold = gold
	} else if playerID == t.Player2ID {
		t.Player2Gold = gold
	} else {
		return errors.New("player not in trade")
	}

	return nil
}

func (t *Trade) Accept(playerID int64) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.State != TradeStatePending {
		return errors.New("trade is not in pending state")
	}

	if playerID == t.Player1ID {
		t.Player1Accept = true
	} else if playerID == t.Player2ID {
		t.Player2Accept = true
	} else {
		return errors.New("player not in trade")
	}

	if t.Player1Accept && t.Player2Accept {
		t.State = TradeStateAccepted
	}

	return nil
}

func (t *Trade) Cancel(playerID int64) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if playerID != t.Player1ID && playerID != t.Player2ID {
		return errors.New("player not in trade")
	}

	t.State = TradeStateCancelled
	return nil
}

func (t *Trade) Complete(player1Inv, player2Inv *Inventory) error {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.State != TradeStateAccepted {
		return errors.New("trade not accepted")
	}

	for _, item := range t.Player1Items {
		err := player1Inv.RemoveItemByID(item.ItemID, item.Count)
		if err != nil {
			return err
		}
		err = player2Inv.AddItem(item.ItemID, item.Count)
		if err != nil {
			player1Inv.AddItem(item.ItemID, item.Count)
			return err
		}
	}

	for _, item := range t.Player2Items {
		err := player2Inv.RemoveItemByID(item.ItemID, item.Count)
		if err != nil {
			return err
		}
		err = player1Inv.AddItem(item.ItemID, item.Count)
		if err != nil {
			player2Inv.AddItem(item.ItemID, item.Count)
			return err
		}
	}

	if t.Player1Gold > 0 {
		err := player1Inv.RemoveGold(t.Player1Gold)
		if err != nil {
			return err
		}
		player2Inv.AddGold(t.Player1Gold)
	}

	if t.Player2Gold > 0 {
		err := player2Inv.RemoveGold(t.Player2Gold)
		if err != nil {
			return err
		}
		player1Inv.AddGold(t.Player2Gold)
	}

	t.State = TradeStateCompleted
	return nil
}

func (inv *Inventory) RemoveItemByID(itemID int64, count int32) error {
	inv.mu.Lock()
	defer inv.mu.Unlock()

	remaining := count
	for _, item := range inv.Items {
		if item.ItemID == itemID {
			if item.Count >= remaining {
				item.Count -= remaining
				if item.Count <= 0 {
					delete(inv.Items, item.Slot)
				}
				return nil
			} else {
				remaining -= item.Count
				delete(inv.Items, item.Slot)
			}
		}
	}

	return errors.New("not enough items")
}
