package api

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"resource-station/internal/database"
	"resource-station/internal/models"
	"resource-station/internal/storage"
	"resource-station/internal/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ResourceHandler 资源处理器
type ResourceHandler struct {
	db      *gorm.DB
	storage storage.FileStorage
}

// NewResourceHandler 创建资源处理器
func NewResourceHandler() (*ResourceHandler, error) {
	storageInstance, err := storage.NewLocalFileStorage()
	if err != nil {
		return nil, err
	}

	return &ResourceHandler{
		db:      database.GetDB(),
		storage: storageInstance,
	}, nil
}

// CreateResource 创建资源（初始化上传）
func (h *ResourceHandler) CreateResource(c *gin.Context) {
	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

		var req models.ResourceCreateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			// 特别处理MD5缺失错误
			if strings.Contains(err.Error(), "ResourceCreateRequest.Hash") {
					c.JSON(http.StatusBadRequest, gin.H{
						"error": "缺少文件SHA256哈希值",
						"details": "请在上传前计算文件SHA256并包含在请求中",
					})
				return
			}
			c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
			return
		}

	// 验证文件类型
	if !storage.ValidateFileType(req.FileType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "不支持的文件类型"})
		return
	}

	// 验证文件大小
	config := utils.GetConfig().Storage
	if req.FileSize > config.MaxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "文件大小超过限制"})
		return
	}

		// 检查SHA256哈希是否已存在
	var existingResource models.Resource
	if err := h.db.Where("hash = ?", req.Hash).First(&existingResource).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"error":                "文件已存在",
			"details":              "相同内容的文件已上传",
			"existing_resource_id": existingResource.ID,
		})
		return
	}

	// 生成资源ID
	resourceID := storage.GenerateResourceID()

	// 创建资源记录
	resource := models.Resource{
		ID:           resourceID,
		UserID:       userID.(uint),
		Filename:     req.Filename,
		OriginalName: req.Filename,
		FileType:     req.FileType,
		FileSize:     req.FileSize,
		StoragePath:  "", // 上传完成后设置
		Hash:         "", // 上传完成后计算
		Status:       "uploading",
		Description:  req.Description,
		Tags:         req.Tags,
		IsPublic:     req.IsPublic,
		ChunkCount:   req.ChunkCount,
		ChunkSize:    req.ChunkSize,
		UploadID:     resourceID, // 使用资源ID作为上传会话ID
	}

	if err := h.db.Create(&resource).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建资源失败"})
		return
	}

	// 如果是分片上传，创建分片记录
	if req.ChunkCount > 1 {
		for i := 0; i < req.ChunkCount; i++ {
			chunk := models.ResourceChunk{
				ResourceID: resourceID,
				ChunkIndex: i,
				ChunkSize:  0,
				Status:     "pending",
			}
			if err := h.db.Create(&chunk).Error; err != nil {
				log.Printf("创建分片记录失败: %v", err)
			}
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":    "资源创建成功",
		"resource":   resource.ToResponse(""),
		"upload_id":  resourceID,
		"chunk_size": req.ChunkSize,
	})
}

// UploadChunk 上传文件分片
func (h *ResourceHandler) UploadChunk(c *gin.Context) {
	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	resourceID := c.Param("id")
	chunkIndexStr := c.Param("chunkIndex")
	chunkIndex, err := strconv.Atoi(chunkIndexStr)
	if err != nil || chunkIndex < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的分片索引"})
		return
	}

	// 验证资源所有权
	var resource models.Resource
	if err := h.db.Where("id = ? AND user_id = ?", resourceID, userID).First(&resource).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "资源不存在或无权访问"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询资源失败"})
		return
	}

	// 检查资源状态
	if resource.Status != "uploading" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "资源不在上传状态"})
		return
	}

	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "获取上传文件失败", "details": err.Error()})
		return
	}

	// 保存文件分片
	filePath, written, err := h.storage.SaveFile(file, resourceID, chunkIndex)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件分片失败", "details": err.Error()})
		return
	}

	// 更新分片记录
	var chunk models.ResourceChunk
	if err := h.db.Where("resource_id = ? AND chunk_index = ?", resourceID, chunkIndex).First(&chunk).Error; err != nil {
		// 如果分片记录不存在，创建新的
		chunk = models.ResourceChunk{
			ResourceID: resourceID,
			ChunkIndex: chunkIndex,
			ChunkSize:  written,
			Status:     "uploaded",
			UploadedAt: time.Now(),
		}
		if err := h.db.Create(&chunk).Error; err != nil {
			log.Printf("创建分片记录失败: %v", err)
		}
	} else {
		// 更新现有分片记录
		chunk.ChunkSize = written
		chunk.Status = "uploaded"
		chunk.UploadedAt = time.Now()
		if err := h.db.Save(&chunk).Error; err != nil {
			log.Printf("更新分片记录失败: %v", err)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "分片上传成功",
		"resource_id": resourceID,
		"chunk_index": chunkIndex,
		"chunk_size":  written,
		"file_path":   filePath,
	})
}

// CompleteUpload 完成上传（合并分片）
func (h *ResourceHandler) CompleteUpload(c *gin.Context) {
	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	resourceID := c.Param("id")

	// 验证资源所有权
	var resource models.Resource
	if err := h.db.Where("id = ? AND user_id = ?", resourceID, userID).First(&resource).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "资源不存在或无权访问"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询资源失败"})
		return
	}

	// 检查资源状态
	if resource.Status != "uploading" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "资源不在上传状态"})
		return
	}

	// 检查所有分片是否已上传
	var pendingChunks int64
	if err := h.db.Model(&models.ResourceChunk{}).
		Where("resource_id = ? AND status != ?", resourceID, "uploaded").
		Count(&pendingChunks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "检查分片状态失败"})
		return
	}

	if pendingChunks > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "还有未上传的分片",
			"pending": pendingChunks,
		})
		return
	}

	// 获取临时文件路径用于MD5计算
	var err error
	if resource.ChunkCount > 1 {
		// 分片上传，使用第一个分片路径计算MD5
		_, err = h.storage.GetTempFilePath(resourceID, 0)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取临时文件路径失败", "details": err.Error()})
			return
		}
	} else {
		// 单文件上传，使用临时文件路径
		_, err = h.storage.GetTempFilePath(resourceID, -1)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "获取临时文件路径失败", "details": err.Error()})
			return
		}
	}

	// 合并分片或移动单文件
	if resource.ChunkCount > 1 {
		if err := h.storage.MergeChunks(resourceID, resource.ChunkCount); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "合并分片失败", "details": err.Error()})
			return
		}
	} else {
		// 单文件上传，需要将文件从temp目录移动到最终目录
		if err := h.storage.MoveSingleFile(resourceID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "移动文件失败", "details": err.Error()})
			return
		}
	}

	// 获取最终文件路径和计算哈希
	filePath, err := h.storage.GetFilePath(resourceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取文件路径失败", "details": err.Error()})
		return
	}

	hash, err := h.storage.CalculateFileHash(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "计算文件哈希失败", "details": err.Error()})
		return
	}

	// 更新资源记录
	resource.Status = "completed"
	resource.StoragePath = filePath
	resource.Hash = hash
	resource.UpdatedAt = time.Now()

	if err := h.db.Save(&resource).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新资源状态失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "上传完成",
		"resource": resource.ToResponse(c.Request.Host),
	})
}

// GetResources 获取资源列表
func (h *ResourceHandler) GetResources(c *gin.Context) {
	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	var req models.ResourceQueryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	// 设置默认值
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}

	// 构建查询
	query := h.db.Model(&models.Resource{}).Where("user_id = ?", userID)

	// 添加过滤条件
	if req.Keyword != "" {
		keyword := "%" + req.Keyword + "%"
		query = query.Where("filename LIKE ? OR description LIKE ? OR tags LIKE ?", keyword, keyword, keyword)
	}

	if req.FileType != "" {
		query = query.Where("file_type LIKE ?", req.FileType+"%")
	}

	if req.StartDate != "" {
		query = query.Where("created_at >= ?", req.StartDate)
	}

	if req.EndDate != "" {
		query = query.Where("created_at <= ?", req.EndDate)
	}

	if req.IsPublic != nil {
		query = query.Where("is_public = ?", *req.IsPublic)
	}

	// 获取总数
	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询资源总数失败"})
		return
	}

	// 获取分页数据
	var resources []models.Resource
	offset := (req.Page - 1) * req.PageSize
	if err := query.Preload("User").
		Order("created_at DESC").
		Offset(offset).
		Limit(req.PageSize).
		Find(&resources).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询资源列表失败"})
		return
	}

	// 转换为响应格式
	var response []models.ResourceResponse
	for _, resource := range resources {
		response = append(response, *resource.ToResponse(c.Request.Host))
	}

	c.JSON(http.StatusOK, gin.H{
		"resources": response,
		"total":     total,
		"page":      req.Page,
		"page_size": req.PageSize,
		"pages":     (total + int64(req.PageSize) - 1) / int64(req.PageSize),
	})
}

// GetResource 获取单个资源详情
func (h *ResourceHandler) GetResource(c *gin.Context) {
	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	resourceID := c.Param("id")

	// 查询资源
	var resource models.Resource
	if err := h.db.Preload("User").
		Where("id = ?", resourceID).
		First(&resource).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "资源不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询资源失败"})
		return
	}

	// 检查权限（用户只能查看自己的资源或公开资源）
	if resource.UserID != userID.(uint) && !resource.IsPublic {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权访问该资源"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"resource": resource.ToResponse(c.Request.Host),
	})
}

// DownloadResource 下载资源
func (h *ResourceHandler) DownloadResource(c *gin.Context) {
	resourceID := c.Param("id")

	// 查询资源
	var resource models.Resource
	if err := h.db.Where("id = ?", resourceID).First(&resource).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "资源不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询资源失败"})
		return
	}

	// 检查资源状态
	if resource.Status != "completed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "资源未就绪"})
		return
	}

	// 检查权限（公开资源或用户自己的资源）
	userID, exists := c.Get("user_id")
	if !exists && !resource.IsPublic {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	if exists && resource.UserID != userID.(uint) && !resource.IsPublic {
		c.JSON(http.StatusForbidden, gin.H{"error": "无权访问该资源"})
		return
	}

	// 获取文件路径
	filePath, err := h.storage.GetFilePath(resourceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取文件失败", "details": err.Error()})
		return
	}

	// 设置下载头
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", resource.OriginalName))
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Length", fmt.Sprintf("%d", resource.FileSize))

	// 提供文件下载
	c.File(filePath)
}

// DeleteResource 删除资源
func (h *ResourceHandler) DeleteResource(c *gin.Context) {
	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	resourceID := c.Param("id")

	// 验证资源所有权
	var resource models.Resource
	if err := h.db.Where("id = ? AND user_id = ?", resourceID, userID).First(&resource).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "资源不存在或无权访问"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询资源失败"})
		return
	}

	// 删除物理文件
	if err := h.storage.DeleteFile(resourceID); err != nil {
		log.Printf("删除物理文件失败: %v", err)
	}

	// 删除相关的分片记录（硬删除，因为ResourceChunk没有DeletedAt字段）
	if err := h.db.Where("resource_id = ?", resourceID).Delete(&models.ResourceChunk{}).Error; err != nil {
		log.Printf("删除分片记录失败: %v", err)
	}

	// 删除数据库记录（硬删除）
	if err := h.db.Unscoped().Delete(&resource).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除资源失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "资源删除成功",
	})
}

// UpdateResource 更新资源信息
func (h *ResourceHandler) UpdateResource(c *gin.Context) {
	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	resourceID := c.Param("id")

	// 验证资源所有权
	var resource models.Resource
	if err := h.db.Where("id = ? AND user_id = ?", resourceID, userID).First(&resource).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "资源不存在或无权访问"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询资源失败"})
		return
	}

	var req models.ResourceUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	// 更新资源信息
	updates := make(map[string]interface{})
	if req.Filename != "" {
		updates["filename"] = req.Filename
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Tags != "" {
		updates["tags"] = req.Tags
	}
	updates["is_public"] = req.IsPublic
	updates["updated_at"] = time.Now()

	if len(updates) > 0 {
		if err := h.db.Model(&resource).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "更新资源失败"})
			return
		}
	}

	// 重新获取资源信息
	h.db.First(&resource, resourceID)

	c.JSON(http.StatusOK, gin.H{
		"message":  "资源更新成功",
		"resource": resource.ToResponse(c.Request.Host),
	})
}

// GetUploadProgress 获取上传进度
func (h *ResourceHandler) GetUploadProgress(c *gin.Context) {
	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	resourceID := c.Param("id")

	// 验证资源所有权
	var resource models.Resource
	if err := h.db.Where("id = ? AND user_id = ?", resourceID, userID).First(&resource).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "资源不存在或无权访问"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询资源失败"})
		return
	}

	// 获取已上传的分片
	var uploadedChunks []models.ResourceChunk
	if err := h.db.Where("resource_id = ? AND status = ?", resourceID, "uploaded").
		Order("chunk_index").
		Find(&uploadedChunks).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "查询上传进度失败"})
		return
	}

	// 计算上传进度
	var uploadedSize int64
	for _, chunk := range uploadedChunks {
		uploadedSize += chunk.ChunkSize
	}

	progress := float64(uploadedSize) / float64(resource.FileSize) * 100
	if progress > 100 {
		progress = 100
	}

	c.JSON(http.StatusOK, gin.H{
		"resource_id":     resourceID,
		"total_size":      resource.FileSize,
		"uploaded_size":   uploadedSize,
		"progress":        fmt.Sprintf("%.2f", progress),
		"chunk_count":     resource.ChunkCount,
		"uploaded_chunks": len(uploadedChunks),
		"status":          resource.Status,
	})
}

// BatchUpload 批量上传（创建多个资源）
func (h *ResourceHandler) BatchUpload(c *gin.Context) {
	// 获取用户ID
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "未认证"})
		return
	}

	var requests []models.ResourceCreateRequest
	if err := c.ShouldBindJSON(&requests); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数错误", "details": err.Error()})
		return
	}

	var responses []gin.H
	for _, req := range requests {
		// 验证文件类型
		if !storage.ValidateFileType(req.FileType) {
			responses = append(responses, gin.H{
				"filename": req.Filename,
				"error":    "不支持的文件类型",
			})
			continue
		}

		// 验证文件大小
		config := utils.GetConfig().Storage
		if req.FileSize > config.MaxFileSize {
			responses = append(responses, gin.H{
				"filename": req.Filename,
				"error":    "文件大小超过限制",
			})
			continue
		}

		// 生成资源ID
		resourceID := storage.GenerateResourceID()

		// 创建资源记录
		resource := models.Resource{
			ID:           resourceID,
			UserID:       userID.(uint),
			Filename:     req.Filename,
			OriginalName: req.Filename,
			FileType:     req.FileType,
			FileSize:     req.FileSize,
			StoragePath:  "",
			Hash:         "",
			Status:       "uploading",
			Description:  req.Description,
			Tags:         req.Tags,
			IsPublic:     req.IsPublic,
			ChunkCount:   req.ChunkCount,
			ChunkSize:    req.ChunkSize,
			UploadID:     resourceID,
		}

		if err := h.db.Create(&resource).Error; err != nil {
			responses = append(responses, gin.H{
				"filename": req.Filename,
				"error":    "创建资源失败",
			})
			continue
		}

		responses = append(responses, gin.H{
			"filename":    req.Filename,
			"resource_id": resourceID,
			"upload_id":   resourceID,
			"chunk_size":  req.ChunkSize,
			"status":      "uploading",
		})
	}

	c.JSON(http.StatusMultiStatus, gin.H{
		"results": responses,
	})
}
