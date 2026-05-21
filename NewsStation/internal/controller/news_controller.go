package controller

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"geonews/internal/model"
	"geonews/internal/service"
)

type NewsController struct {
	service  service.NewsService
	validate *validator.Validate
}

func NewNewsController(service service.NewsService) *NewsController {
	return &NewsController{
		service:  service,
		validate: validator.New(),
	}
}

func (c *NewsController) CreateNews(ctx *gin.Context) {
	var req model.NewsRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.validate.Struct(req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.service.CreateNews(&req); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "News created successfully"})
}

func (c *NewsController) GetNewsForMap(ctx *gin.Context) {
	bounds := ctx.Query("bounds")

	if bounds != "" {
		c.GetNewsByBounds(ctx)
		return
	}

	level := ctx.Query("level")
	code := ctx.Query("code")

	if level == "" && code == "" {
		newsList, err := c.service.GetAllNews()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"data": newsList})
		return
	}

	if level == "" {
		level = "country"
	}

	newsList, err := c.service.GetNewsByGeoLevel(level, code)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": newsList})
}

func (c *NewsController) GetNewsByBounds(ctx *gin.Context) {
	bounds := ctx.Query("bounds")
	if bounds == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "bounds parameter is required"})
		return
	}

	newsList, err := c.service.GetNewsByBounds(bounds)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": newsList})
}

func (c *NewsController) GetBreakingNews(ctx *gin.Context) {
	newsList, err := c.service.GetBreakingNews()
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"data": newsList})
}

func (c *NewsController) GetHistoryNews(ctx *gin.Context) {
	startStr := ctx.Query("start")
	endStr := ctx.Query("end")
	pageStr := ctx.DefaultQuery("page", "1")
	limitStr := ctx.DefaultQuery("limit", "20")

	start, err := time.Parse("2006-01-02", startStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid start date format"})
		return
	}

	end, err := time.Parse("2006-01-02", endStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "invalid end date format"})
		return
	}

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 20
	}

	newsList, total, err := c.service.GetHistoryNews(start, end, limit, page)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":  newsList,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (c *NewsController) StreamBreakingNews(ctx *gin.Context) {
	ctx.Header("Content-Type", "text/event-stream")
	ctx.Header("Cache-Control", "no-cache")
	ctx.Header("Connection", "keep-alive")
	ctx.Header("Access-Control-Allow-Origin", "*")

	lastTime := time.Now()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			newsList, err := c.service.GetLatestBreakingNews(lastTime)
			if err != nil {
				continue
			}

			if len(newsList) > 0 {
				lastTime = time.Now()

				for _, news := range newsList {
					data, _ := json.Marshal(news)
					ctx.Writer.WriteString("data: " + string(data) + "\n\n")
					ctx.Writer.Flush()
				}
			}
		case <-ctx.Request.Context().Done():
			return
		}
	}
}

func (c *NewsController) GeoCode(ctx *gin.Context) {
	address := ctx.Query("address")
	if address == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "address parameter is required"})
		return
	}

	lat, lng, err := c.service.GeoCode(address)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"latitude":  lat,
		"longitude": lng,
	})
}
