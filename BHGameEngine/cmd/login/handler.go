package main

import (
	"encoding/json"

	"github.com/openworld-server/internal/msg"
	"github.com/openworld-server/internal/network"
	"github.com/openworld-server/pkg/logger"
)

type LoginHandler struct {
}

func NewLoginHandler() *LoginHandler {
	return &LoginHandler{}
}

func (h *LoginHandler) Handle(msgObj *network.Message) {
	logger.Info("Login received message - MsgID:", msgObj.ID, "(", msg.GetMsgName(msgObj.ID), ")")

	switch msgObj.ID {
	case msg.MSG_LOGIN_REQ:
		h.handleLoginRequest(msgObj)
	case msg.MSG_REGISTER_REQ:
		h.handleRegisterRequest(msgObj)
	}
}

func (h *LoginHandler) handleLoginRequest(msgObj *network.Message) {
	logger.Info("Handling login request")

	req := &msg.LoginRequest{}
	if err := json.Unmarshal(msgObj.Data, req); err != nil {
		logger.Error("Failed to unmarshal login request:", err)
		return
	}

	logger.Info("Login attempt - Account:", req.Account)

	res := &msg.LoginResponse{
		Result:     0,
		Message:    "success",
		SessionID:  "session_" + req.Account,
		PlayerID:   10001,
		PlayerName: req.Account,
		Level:      1,
		Health:     100,
	}

	data, err := json.Marshal(res)
	if err != nil {
		logger.Error("Failed to marshal login response:", err)
		return
	}

	msgObj.Session.Send(msg.MSG_LOGIN_RES, msg.NodeTypeLogin, data)
	logger.Info("Login success - Account:", req.Account, "PlayerID:", res.PlayerID)
}

func (h *LoginHandler) handleRegisterRequest(msgObj *network.Message) {
	logger.Info("Handling register request")

	req := &msg.RegisterRequest{}
	if err := json.Unmarshal(msgObj.Data, req); err != nil {
		logger.Error("Failed to unmarshal register request:", err)
		return
	}

	logger.Info("Register attempt - Account:", req.Account)

	res := &msg.RegisterResponse{
		Result:  0,
		Message: "success",
	}

	data, err := json.Marshal(res)
	if err != nil {
		logger.Error("Failed to marshal register response:", err)
		return
	}

	msgObj.Session.Send(msg.MSG_REGISTER_RES, msg.NodeTypeLogin, data)
	logger.Info("Register success - Account:", req.Account)
}
