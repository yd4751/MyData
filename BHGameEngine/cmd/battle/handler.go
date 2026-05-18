package main

import (
	"github.com/openworld-server/internal/network"
	"github.com/openworld-server/pkg/logger"
)

type BattleHandler struct{}

func NewBattleHandler() *BattleHandler {
	return &BattleHandler{}
}

func (h *BattleHandler) Handle(msg *network.Message) {
	logger.Info("Battle server received message from ", msg.Session.RemoteAddr())
}
