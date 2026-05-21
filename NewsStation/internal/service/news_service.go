package service

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/pkg/errors"

	"geonews/internal/model"
	"geonews/internal/repository"
)

type NewsService interface {
	CreateNews(req *model.NewsRequest) error
	GetAllNews() ([]model.News, error)
	GetNewsByGeoLevel(level, code string) ([]model.News, error)
	GetBreakingNews() ([]model.News, error)
	GetHistoryNews(start, end time.Time, limit, page int) ([]model.News, int64, error)
	GetNewsByBounds(bounds string) ([]model.News, error)
	GeoCode(address string) (float64, float64, error)
	GetLatestBreakingNews(after time.Time) ([]model.News, error)
}

type newsService struct {
	repo repository.NewsRepository
}

func NewNewsService(repo repository.NewsRepository) NewsService {
	return &newsService{repo: repo}
}

func (s *newsService) CreateNews(req *model.NewsRequest) error {
	publishTime, err := time.Parse(time.RFC3339, req.PublishTime)
	if err != nil {
		return errors.Wrap(err, "invalid publish time format")
	}

	isDuplicate, err := s.repo.CheckDuplicate(req.Title, publishTime)
	if err != nil {
		return err
	}

	if isDuplicate {
		return errors.New("duplicate news")
	}

	news := &model.News{
		Title:       req.Title,
		Summary:     req.Summary,
		Source:      req.Source,
		PublishTime: publishTime,
		Lat:         req.Latitude,
		Lng:         req.Longitude,
		GeoLevel:    req.GeoLevel,
		IsBreaking:  req.IsBreaking,
		Priority:    req.Priority,
		CreatedAt:   time.Now(),
	}

	return s.repo.CreateNews(news)
}

func (s *newsService) GetAllNews() ([]model.News, error) {
	return s.repo.GetAllNews()
}

func (s *newsService) GetNewsByGeoLevel(level, code string) ([]model.News, error) {
	return s.repo.GetNewsByGeoLevel(level, code)
}

func (s *newsService) GetBreakingNews() ([]model.News, error) {
	return s.repo.GetBreakingNews()
}

func (s *newsService) GetHistoryNews(start, end time.Time, limit, page int) ([]model.News, int64, error) {
	offset := (page - 1) * limit
	newsList, err := s.repo.GetHistoryNews(start, end, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.repo.CountHistoryNews(start, end)
	if err != nil {
		return nil, 0, err
	}

	return newsList, total, nil
}

func (s *newsService) GetNewsByBounds(bounds string) ([]model.News, error) {
	parts := parseBounds(bounds)
	if len(parts) != 4 {
		return nil, errors.New("invalid bounds format")
	}

	swLng, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return nil, errors.Wrap(err, "invalid sw_lng")
	}

	swLat, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return nil, errors.Wrap(err, "invalid sw_lat")
	}

	neLng, err := strconv.ParseFloat(parts[2], 64)
	if err != nil {
		return nil, errors.Wrap(err, "invalid ne_lng")
	}

	neLat, err := strconv.ParseFloat(parts[3], 64)
	if err != nil {
		return nil, errors.Wrap(err, "invalid ne_lat")
	}

	return s.repo.GetNewsByBounds(swLng, swLat, neLng, neLat)
}

func parseBounds(bounds string) []string {
	var result []string
	current := ""
	for _, c := range bounds {
		if c == ',' {
			result = append(result, current)
			current = ""
		} else {
			current += string(c)
		}
	}
	if current != "" {
		result = append(result, current)
	}
	return result
}

func (s *newsService) GeoCode(address string) (float64, float64, error) {
	url := fmt.Sprintf("https://restapi.amap.com/v3/geocode/geo?address=%s&key=5f6d7e8d9c3b2a1e0f5c7d9b8a7c6e5d", address)
	resp, err := http.Get(url)
	if err != nil {
		return 0, 0, errors.Wrap(err, "failed to call geocode API")
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return 0, 0, errors.Wrap(err, "failed to read geocode response")
	}

	var result struct {
		Status   string `json:"status"`
		Geocodes []struct {
			Location string `json:"location"`
		} `json:"geocodes"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return 0, 0, errors.Wrap(err, "failed to parse geocode response")
	}

	if result.Status != "1" || len(result.Geocodes) == 0 {
		return 0, 0, errors.New("geocode failed")
	}

	location := result.Geocodes[0].Location
	parts := parseBounds(location)
	if len(parts) != 2 {
		return 0, 0, errors.New("invalid location format")
	}

	lng, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return 0, 0, errors.Wrap(err, "invalid longitude")
	}

	lat, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return 0, 0, errors.Wrap(err, "invalid latitude")
	}

	return lat, lng, nil
}

func (s *newsService) GetLatestBreakingNews(after time.Time) ([]model.News, error) {
	return s.repo.GetLatestBreakingNews(after)
}
