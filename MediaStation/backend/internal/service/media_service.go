package service

import (
	"os"
	"path/filepath"

	"mediastation/config"
	"mediastation/internal/model"
	"mediastation/internal/repository"
)

type MediaService interface {
	GetMedia(id uint) (*model.Media, error)
	GetMediaBySeries(seriesID uint) ([]model.Media, error)
	GetAllMedia(mediaType model.MediaType) ([]model.Media, error)
	GetSeries(id uint) (*model.MediaSeries, error)
	GetAllSeries(mediaType model.MediaType) ([]model.MediaSeries, error)
	CreateMedia(media *model.Media) error
	CreateSeries(series *model.MediaSeries) error
	UpdateMedia(media *model.Media) error
	DeleteMedia(id uint) error
	SearchMedia(keyword string) ([]model.Media, error)
	StreamMedia(id uint) (string, error)
}

type mediaService struct {
	repo   repository.MediaRepository
	config *config.Config
}

func NewMediaService(repo repository.MediaRepository, cfg *config.Config) MediaService {
	return &mediaService{repo: repo, config: cfg}
}

func (s *mediaService) GetMedia(id uint) (*model.Media, error) {
	return s.repo.GetByID(id)
}

func (s *mediaService) GetMediaBySeries(seriesID uint) ([]model.Media, error) {
	return s.repo.GetBySeriesID(seriesID)
}

func (s *mediaService) GetAllMedia(mediaType model.MediaType) ([]model.Media, error) {
	return s.repo.GetAllByType(mediaType)
}

func (s *mediaService) GetSeries(id uint) (*model.MediaSeries, error) {
	return s.repo.GetSeriesByID(id)
}

func (s *mediaService) GetAllSeries(mediaType model.MediaType) ([]model.MediaSeries, error) {
	return s.repo.GetAllSeries(mediaType)
}

func (s *mediaService) CreateMedia(media *model.Media) error {
	return s.repo.Create(media)
}

func (s *mediaService) CreateSeries(series *model.MediaSeries) error {
	return s.repo.CreateSeries(series)
}

func (s *mediaService) UpdateMedia(media *model.Media) error {
	return s.repo.Update(media)
}

func (s *mediaService) DeleteMedia(id uint) error {
	return s.repo.Delete(id)
}

func (s *mediaService) SearchMedia(keyword string) ([]model.Media, error) {
	return s.repo.Search(keyword)
}

func (s *mediaService) StreamMedia(id uint) (string, error) {
	media, err := s.repo.GetByID(id)
	if err != nil {
		return "", err
	}

	filePath := filepath.Join(s.config.MediaDir, media.FilePath)
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", os.ErrNotExist
	}

	return filePath, nil
}
