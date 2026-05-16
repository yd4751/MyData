package service

import (
	"time"

	"mediastation/internal/model"
	"mediastation/internal/repository"
)

type HistoryService interface {
	GetPlayHistory(userID uint) ([]model.PlayHistory, error)
	GetProgress(userID, mediaID uint) (int, error)
	SaveProgress(userID, mediaID uint, progress int) error
	RemoveFromHistory(userID, mediaID uint) error
	ClearHistory(userID uint) error
}

type historyService struct {
	repo repository.HistoryRepository
}

func NewHistoryService(repo repository.HistoryRepository) HistoryService {
	return &historyService{repo: repo}
}

func (s *historyService) GetPlayHistory(userID uint) ([]model.PlayHistory, error) {
	return s.repo.GetByUserID(userID)
}

func (s *historyService) GetProgress(userID, mediaID uint) (int, error) {
	history, err := s.repo.GetByUserAndMedia(userID, mediaID)
	if err != nil {
		return 0, err
	}
	return history.Progress, nil
}

func (s *historyService) SaveProgress(userID, mediaID uint, progress int) error {
	history, err := s.repo.GetByUserAndMedia(userID, mediaID)
	if err != nil {
		newHistory := &model.PlayHistory{
			UserID:   userID,
			MediaID:  mediaID,
			Progress: progress,
			PlayedAt: time.Now(),
		}
		return s.repo.Create(newHistory)
	}

	history.Progress = progress
	history.UpdatedAt = time.Now()
	return s.repo.Update(history)
}

func (s *historyService) RemoveFromHistory(userID, mediaID uint) error {
	return s.repo.Delete(userID, mediaID)
}

func (s *historyService) ClearHistory(userID uint) error {
	return s.repo.DeleteAll(userID)
}
