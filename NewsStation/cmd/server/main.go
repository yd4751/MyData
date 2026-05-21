package main

import (
	"fmt"

	"github.com/spf13/viper"

	"geonews/internal/config"
	"geonews/internal/controller"
	"geonews/internal/database"
	"geonews/internal/repository"
	"geonews/internal/router"
	"geonews/internal/service"
)

func main() {
	_, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		return
	}

	db, err := database.NewDB()
	if err != nil {
		fmt.Printf("Failed to connect to database: %v\n", err)
		return
	}
	defer db.Close()

	newsRepo := repository.NewNewsRepository(db)
	newsService := service.NewNewsService(newsRepo)
	newsController := controller.NewNewsController(newsService)

	r := router.SetupRouter(newsController)

	port := viper.GetString("server.port")
	if port == "" {
		port = "8080"
	}

	fmt.Printf("Server starting on port %s...\n", port)
	if err := r.Run(":" + port); err != nil {
		fmt.Printf("Failed to start server: %v\n", err)
	}
}
