package model

import (
	"time"
)

type News struct {
	ID          int64     `db:"id" json:"id"`
	Title       string    `db:"title" json:"title"`
	Summary     string    `db:"summary" json:"summary"`
	Source      string    `db:"source" json:"source"`
	PublishTime time.Time `db:"publish_time" json:"publish_time"`
	Lat         float64   `db:"lat" json:"latitude"`
	Lng         float64   `db:"lng" json:"longitude"`
	GeoLevel    string    `db:"geo_level" json:"geo_level"`
	IsBreaking  bool      `db:"is_breaking" json:"is_breaking"`
	Priority    int       `db:"priority" json:"priority"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
}

type NewsRequest struct {
	Title       string  `json:"title" validate:"required,max=255"`
	Summary     string  `json:"summary" validate:"required"`
	Source      string  `json:"source" validate:"required,max=50"`
	PublishTime string  `json:"publish_time" validate:"required"`
	GeoLevel    string  `json:"geo_level" validate:"required,oneof=world continent country city"`
	Latitude    float64 `json:"latitude" validate:"required,min=-90,max=90"`
	Longitude   float64 `json:"longitude" validate:"required,min=-180,max=180"`
	IsBreaking  bool    `json:"is_breaking"`
	Priority    int     `json:"priority" validate:"min=1,max=5"`
}

type NewsQueryParams struct {
	Level string
	Code  string
	Start time.Time
	End   time.Time
	Limit int
	Page  int
}
