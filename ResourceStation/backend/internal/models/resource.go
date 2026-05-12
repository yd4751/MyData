package models

import (
	"time"

	"gorm.io/gorm"
)

// Resource 资源模型
type Resource struct {
	ID          string         `gorm:"primaryKey;size:36" json:"id"` // UUID
	UserID      uint           `gorm:"index;not null" json:"user_id"`
	Filename    string         `gorm:"size:255;not null" json:"filename"`
	OriginalName string        `gorm:"size:255;not null" json:"original_name"`
	FileType    string         `gorm:"size:100;not null" json:"file_type"` // MIME类型
	FileSize    int64          `gorm:"not null" json:"file_size"` // 字节
	StoragePath string         `gorm:"size:500;not null" json:"storage_path"`
		Hash        string         `gorm:"size:64;index" json:"hash"` // 文件哈希（SHA256）
	Status      string         `gorm:"size:20;default:'uploading'" json:"status"` // uploading, completed, deleted, error
	Description string         `gorm:"type:text" json:"description"`
	Tags        string         `gorm:"type:text" json:"tags"` // 逗号分隔的标签
	IsPublic    bool           `gorm:"default:false" json:"is_public"`
	ChunkCount  int            `gorm:"default:1" json:"chunk_count"` // 总分片数
	ChunkSize   int64          `gorm:"default:0" json:"chunk_size"` // 分片大小
	UploadID    string         `gorm:"size:50;index" json:"upload_id"` // 上传会话ID
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// 关联
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

// ResourceChunk 资源分片模型（用于断点续传）
type ResourceChunk struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	ResourceID string    `gorm:"size:36;index;not null" json:"resource_id"`
	ChunkIndex int       `gorm:"not null" json:"chunk_index"` // 分片索引，从0开始
	ChunkSize  int64     `gorm:"not null" json:"chunk_size"` // 当前分片大小
	Hash       string    `gorm:"size:64" json:"hash"` // 分片哈希
	Status     string    `gorm:"size:20;default:'pending'" json:"status"` // pending, uploaded, merged
	UploadedAt time.Time `json:"uploaded_at"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`

	// 关联
	Resource Resource `gorm:"foreignKey:ResourceID" json:"resource,omitempty"`
}

// ResourceCreateRequest 资源创建请求
type ResourceCreateRequest struct {
	Filename     string `json:"filename" binding:"required"`
	FileType     string `json:"file_type" binding:"required"`
	FileSize     int64  `json:"file_size" binding:"required,min=1"`
	ChunkCount   int    `json:"chunk_count" binding:"min=1"`
	ChunkSize    int64  `json:"chunk_size" binding:"min=0"`
		Hash         string `json:"hash" binding:"required,len=64"` // SHA256哈希长度为64
	Description  string `json:"description"`
	Tags         string `json:"tags"`
	IsPublic     bool   `json:"is_public"`
}

// ResourceUpdateRequest 资源更新请求
type ResourceUpdateRequest struct {
	Filename    string `json:"filename"`
	Description string `json:"description"`
	Tags        string `json:"tags"`
	IsPublic    bool   `json:"is_public"`
}

// ResourceUploadChunkRequest 资源上传分片请求
type ResourceUploadChunkRequest struct {
	ChunkIndex int    `json:"chunk_index" binding:"required,min=0"`
	ChunkSize  int64  `json:"chunk_size" binding:"required,min=1"`
		Hash       string    `json:"hash" binding:"required,len=64"` // SHA256哈希长度为64
}

// ResourceQueryRequest 资源查询请求
type ResourceQueryRequest struct {
	Page      int    `form:"page" binding:"min=1"`
	PageSize  int    `form:"page_size" binding:"min=1,max=100"`
	Keyword   string `form:"keyword"`
	FileType  string `form:"file_type"`
	StartDate string `form:"start_date"`
	EndDate   string `form:"end_date"`
	IsPublic  *bool  `form:"is_public"`
	UserID    uint   `form:"user_id"`
}

// ResourceResponse 资源响应
type ResourceResponse struct {
	ID           string    `json:"id"`
	UserID       uint      `json:"user_id"`
	Filename     string    `json:"filename"`
	OriginalName string    `json:"original_name"`
	FileType     string    `json:"file_type"`
	FileSize     int64     `json:"file_size"`
	StoragePath  string    `json:"storage_path"`
	Hash         string    `json:"hash"`
	Status       string    `json:"status"`
	Description  string    `json:"description"`
	Tags         string    `json:"tags"`
	IsPublic     bool      `json:"is_public"`
	ChunkCount   int       `json:"chunk_count"`
	ChunkSize    int64     `json:"chunk_size"`
	UploadID     string    `json:"upload_id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	DownloadURL  string    `json:"download_url,omitempty"`
	PreviewURL   string    `json:"preview_url,omitempty"`
	
	// 用户信息
	Username string `json:"username,omitempty"`
	Email    string `json:"email,omitempty"`
	Avatar   string `json:"avatar,omitempty"`
}

// ToResponse 将Resource转换为ResourceResponse
func (r *Resource) ToResponse(baseURL string) *ResourceResponse {
	resp := &ResourceResponse{
		ID:           r.ID,
		UserID:       r.UserID,
		Filename:     r.Filename,
		OriginalName: r.OriginalName,
		FileType:     r.FileType,
		FileSize:     r.FileSize,
		StoragePath:  r.StoragePath,
		Hash:         r.Hash,
		Status:       r.Status,
		Description:  r.Description,
		Tags:         r.Tags,
		IsPublic:     r.IsPublic,
		ChunkCount:   r.ChunkCount,
		ChunkSize:    r.ChunkSize,
		UploadID:     r.UploadID,
		CreatedAt:    r.CreatedAt,
		UpdatedAt:    r.UpdatedAt,
	}

	// 生成下载URL和预览URL
	if r.Status == "completed" {
		resp.DownloadURL = baseURL + "/api/v1/resources/" + r.ID + "/download"
		
		// 如果是图片，生成预览URL
		if len(r.FileType) >= 5 && r.FileType[:5] == "image" {
			resp.PreviewURL = baseURL + "/api/v1/resources/" + r.ID + "/preview"
		}
	}

	// 添加用户信息
	if r.User.ID != 0 {
		resp.Username = r.User.Username
		resp.Email = r.User.Email
		resp.Avatar = r.User.Avatar
	}

	return resp
}
