package model

import "time"

type MediaType string

const (
	MediaTypeVideo MediaType = "video"
	MediaTypeAudio MediaType = "audio"
	MediaTypeImage MediaType = "image"
	MediaTypeNovel MediaType = "novel"
)

type Media struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Title       string    `gorm:"size:255;not null" json:"title"`
	Description string    `gorm:"text" json:"description"`
	Type        MediaType `gorm:"size:20;not null" json:"type"`
	FilePath    string    `gorm:"size:500;not null" json:"file_path"`
	Thumbnail   string    `gorm:"size:500" json:"thumbnail"`
	Duration    int       `gorm:"default:0" json:"duration"`
	Width       int       `gorm:"default:0" json:"width"`
	Height      int       `gorm:"default:0" json:"height"`
	Season      int       `gorm:"default:0" json:"season"`
	Episode     int       `gorm:"default:0" json:"episode"`
	SeriesID    uint      `gorm:"default:0;index" json:"series_id"`
	IsVertical  bool      `gorm:"default:false" json:"is_vertical"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type MediaSeries struct {
	ID            uint      `gorm:"primaryKey"`
	Title         string    `gorm:"size:255;not null"`
	Description   string    `gorm:"text"`
	Thumbnail     string    `gorm:"size:500"`
	MediaType     MediaType `gorm:"size:20;not null"`
	TotalEpisodes int       `gorm:"default:0"`
	CreatedAt     time.Time `gorm:"autoCreateTime"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime"`
}
