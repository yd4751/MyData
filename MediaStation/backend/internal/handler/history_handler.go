package handler

import (
	"encoding/json"
	"net/http"
	"strconv"

	"mediastation/internal/service"
)

type HistoryHandler struct {
	historyService service.HistoryService
}

func NewHistoryHandler(historyService service.HistoryService) *HistoryHandler {
	return &HistoryHandler{historyService: historyService}
}

func (h *HistoryHandler) GetHistory(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseUint(r.URL.Query().Get("user_id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	history, err := h.historyService.GetPlayHistory(uint(userID))
	if err != nil {
		http.Error(w, "Failed to get history", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

func (h *HistoryHandler) GetProgress(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseUint(r.URL.Query().Get("user_id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	mediaID, err := strconv.ParseUint(r.URL.Query().Get("media_id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid media ID", http.StatusBadRequest)
		return
	}

	progress, err := h.historyService.GetProgress(uint(userID), uint(mediaID))
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]int{"progress": 0})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"progress": progress})
}

func (h *HistoryHandler) SaveProgress(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID   uint `json:"user_id"`
		MediaID  uint `json:"media_id"`
		Progress int  `json:"progress"`
	}

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err = h.historyService.SaveProgress(req.UserID, req.MediaID, req.Progress)
	if err != nil {
		http.Error(w, "Failed to save progress", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *HistoryHandler) RemoveFromHistory(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseUint(r.URL.Query().Get("user_id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	mediaID, err := strconv.ParseUint(r.URL.Query().Get("media_id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid media ID", http.StatusBadRequest)
		return
	}

	err = h.historyService.RemoveFromHistory(uint(userID), uint(mediaID))
	if err != nil {
		http.Error(w, "Failed to remove from history", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *HistoryHandler) ClearHistory(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseUint(r.URL.Query().Get("user_id"), 10, 64)
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	err = h.historyService.ClearHistory(uint(userID))
	if err != nil {
		http.Error(w, "Failed to clear history", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
