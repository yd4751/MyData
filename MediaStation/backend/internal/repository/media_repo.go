package repository

import (
	"mediastation/internal/model"
	"mediastation/pkg/database"
)

type MediaRepository interface {
	GetByID(id uint) (*model.Media, error)
	GetBySeriesID(seriesID uint) ([]model.Media, error)
	GetAllByType(mediaType model.MediaType) ([]model.Media, error)
	GetSeriesByID(id uint) (*model.MediaSeries, error)
	GetAllSeries(mediaType model.MediaType) ([]model.MediaSeries, error)
	Create(media *model.Media) error
	CreateSeries(series *model.MediaSeries) error
	Update(media *model.Media) error
	Delete(id uint) error
	Search(keyword string) ([]model.Media, error)
}

type mediaRepository struct{}

func NewMediaRepository() MediaRepository {
	return &mediaRepository{}
}

func (r *mediaRepository) GetByID(id uint) (*model.Media, error) {
	var media model.Media
	err := database.DB.First(&media, id).Error
	return &media, err
}

func (r *mediaRepository) GetBySeriesID(seriesID uint) ([]model.Media, error) {
	var mediaList []model.Media
	err := database.DB.Where("series_id = ?", seriesID).Order("episode asc").Find(&mediaList).Error
	return mediaList, err
}

func (r *mediaRepository) GetAllByType(mediaType model.MediaType) ([]model.Media, error) {
	var mediaList []model.Media
	err := database.DB.Where("type = ?", mediaType).Find(&mediaList).Error
	return mediaList, err
}

func (r *mediaRepository) GetSeriesByID(id uint) (*model.MediaSeries, error) {
	var series model.MediaSeries
	err := database.DB.First(&series, id).Error
	return &series, err
}

func (r *mediaRepository) GetAllSeries(mediaType model.MediaType) ([]model.MediaSeries, error) {
	var seriesList []model.MediaSeries
	err := database.DB.Where("media_type = ?", mediaType).Find(&seriesList).Error
	return seriesList, err
}

func (r *mediaRepository) Create(media *model.Media) error {
	return database.DB.Create(media).Error
}

func (r *mediaRepository) CreateSeries(series *model.MediaSeries) error {
	return database.DB.Create(series).Error
}

func (r *mediaRepository) Update(media *model.Media) error {
	return database.DB.Save(media).Error
}

func (r *mediaRepository) Delete(id uint) error {
	return database.DB.Delete(&model.Media{}, id).Error
}

func (r *mediaRepository) Search(keyword string) ([]model.Media, error) {
	var mediaList []model.Media
	err := database.DB.Where("title LIKE ?", "%"+keyword+"%").Find(&mediaList).Error
	return mediaList, err
}
