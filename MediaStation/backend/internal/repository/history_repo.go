package repository

import (
	"mediastation/internal/model"
	"mediastation/pkg/database"
)

type HistoryRepository interface {
	GetByUserID(userID uint) ([]model.PlayHistory, error)
	GetByUserAndMedia(userID, mediaID uint) (*model.PlayHistory, error)
	Create(history *model.PlayHistory) error
	Update(history *model.PlayHistory) error
	Delete(userID, mediaID uint) error
	DeleteAll(userID uint) error
}

type historyRepository struct{}

func NewHistoryRepository() HistoryRepository {
	return &historyRepository{}
}

func (r *historyRepository) GetByUserID(userID uint) ([]model.PlayHistory, error) {
	var historyList []model.PlayHistory
	err := database.DB.Where("user_id = ?", userID).Order("played_at desc").Find(&historyList).Error
	return historyList, err
}

func (r *historyRepository) GetByUserAndMedia(userID, mediaID uint) (*model.PlayHistory, error) {
	var history model.PlayHistory
	err := database.DB.Where("user_id = ? AND media_id = ?", userID, mediaID).First(&history).Error
	return &history, err
}

func (r *historyRepository) Create(history *model.PlayHistory) error {
	return database.DB.Create(history).Error
}

func (r *historyRepository) Update(history *model.PlayHistory) error {
	return database.DB.Save(history).Error
}

func (r *historyRepository) Delete(userID, mediaID uint) error {
	return database.DB.Where("user_id = ? AND media_id = ?", userID, mediaID).Delete(&model.PlayHistory{}).Error
}

func (r *historyRepository) DeleteAll(userID uint) error {
	return database.DB.Where("user_id = ?", userID).Delete(&model.PlayHistory{}).Error
}
