package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"geonews/internal/model"
)

type NewsRepository interface {
	CreateNews(news *model.News) error
	GetAllNews() ([]model.News, error)
	GetNewsByGeoLevel(level, code string) ([]model.News, error)
	GetBreakingNews() ([]model.News, error)
	GetHistoryNews(start, end time.Time, limit, offset int) ([]model.News, error)
	CountHistoryNews(start, end time.Time) (int64, error)
	CheckDuplicate(title string, publishTime time.Time) (bool, error)
	GetNewsByBounds(swLng, swLat, neLng, neLat float64) ([]model.News, error)
	GetLatestBreakingNews(after time.Time) ([]model.News, error)
}

type newsRepository struct {
	db *sqlx.DB
}

func NewNewsRepository(db *sqlx.DB) NewsRepository {
	return &newsRepository{db: db}
}

func (r *newsRepository) CreateNews(news *model.News) error {
	query := `INSERT INTO news (title, summary, source, publish_time, lat, lng, geo_level, is_breaking, priority, created_at)
			  VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := r.db.Exec(query,
		news.Title,
		news.Summary,
		news.Source,
		news.PublishTime,
		news.Lat,
		news.Lng,
		news.GeoLevel,
		news.IsBreaking,
		news.Priority,
		news.CreatedAt,
	)

	return errors.Wrap(err, "failed to create news")
}

func (r *newsRepository) GetAllNews() ([]model.News, error) {
	var newsList []model.News
	query := `SELECT * FROM news ORDER BY publish_time DESC`

	err := r.db.Select(&newsList, query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get all news")
	}

	return newsList, nil
}

func (r *newsRepository) GetNewsByGeoLevel(level, code string) ([]model.News, error) {
	var newsList []model.News
	query := `SELECT * FROM news WHERE geo_level = ? ORDER BY priority ASC, publish_time DESC LIMIT 20`

	if code != "" {
		query = fmt.Sprintf(`SELECT * FROM news WHERE geo_level = '%s' ORDER BY priority ASC, publish_time DESC LIMIT 20`, level)
	}

	err := r.db.Select(&newsList, query, level)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get news by geo level")
	}

	return newsList, nil
}

func (r *newsRepository) GetBreakingNews() ([]model.News, error) {
	var newsList []model.News
	query := `SELECT * FROM news WHERE is_breaking = 1 ORDER BY priority ASC, publish_time DESC LIMIT 10`

	err := r.db.Select(&newsList, query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get breaking news")
	}

	return newsList, nil
}

func (r *newsRepository) GetHistoryNews(start, end time.Time, limit, offset int) ([]model.News, error) {
	var newsList []model.News
	query := `SELECT * FROM news WHERE publish_time BETWEEN ? AND ? ORDER BY publish_time DESC LIMIT ? OFFSET ?`

	err := r.db.Select(&newsList, query, start, end, limit, offset)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get history news")
	}

	return newsList, nil
}

func (r *newsRepository) CountHistoryNews(start, end time.Time) (int64, error) {
	var count int64
	query := `SELECT COUNT(*) FROM news WHERE publish_time BETWEEN ? AND ?`

	err := r.db.Get(&count, query, start, end)
	if err != nil {
		return 0, errors.Wrap(err, "failed to count history news")
	}

	return count, nil
}

func (r *newsRepository) CheckDuplicate(title string, publishTime time.Time) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM news WHERE title = ? AND publish_time = ?)`

	err := r.db.Get(&exists, query, title, publishTime)
	if err != nil && err != sql.ErrNoRows {
		return false, errors.Wrap(err, "failed to check duplicate")
	}

	return exists, nil
}

func (r *newsRepository) GetNewsByBounds(swLng, swLat, neLng, neLat float64) ([]model.News, error) {
	var newsList []model.News
	query := `SELECT * FROM news WHERE lng BETWEEN ? AND ? AND lat BETWEEN ? AND ? ORDER BY publish_time DESC`

	err := r.db.Select(&newsList, query, swLng, neLng, swLat, neLat)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get news by bounds")
	}

	return newsList, nil
}

func (r *newsRepository) GetLatestBreakingNews(after time.Time) ([]model.News, error) {
	var newsList []model.News
	query := `SELECT * FROM news WHERE is_breaking = 1 AND publish_time > ? ORDER BY publish_time DESC`

	err := r.db.Select(&newsList, query, after)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get latest breaking news")
	}

	return newsList, nil
}
