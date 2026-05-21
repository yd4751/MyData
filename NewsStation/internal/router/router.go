package router

import (
	"github.com/gin-gonic/gin"

	"geonews/internal/controller"
)

func SetupRouter(newsController *controller.NewsController) *gin.Engine {
	r := gin.Default()

	v1 := r.Group("/api/v1")
	{
		v1.POST("/news", newsController.CreateNews)
		v1.GET("/news", newsController.GetNewsForMap)
		v1.GET("/news/map", newsController.GetNewsForMap)
		v1.GET("/news/breaking", newsController.GetBreakingNews)
		v1.GET("/news/history", newsController.GetHistoryNews)
		v1.GET("/stream", newsController.StreamBreakingNews)
		v1.GET("/geocode", newsController.GeoCode)
	}

	return r
}
