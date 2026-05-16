package main

import (
	"flag"
	"log"
	"net/http"

	"mediastation/config"
	"mediastation/internal/handler"
	"mediastation/internal/repository"
	"mediastation/internal/service"
	"mediastation/pkg/database"
)

func main() {
	configPath := flag.String("config", "config.json", "Path to the configuration file")
	flag.Parse()

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	err = database.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	mediaRepo := repository.NewMediaRepository()
	userRepo := repository.NewUserRepository()
	historyRepo := repository.NewHistoryRepository()

	mediaService := service.NewMediaService(mediaRepo, cfg)
	userService := service.NewUserService(userRepo)
	historyService := service.NewHistoryService(historyRepo)

	mediaHandler := handler.NewMediaHandler(mediaService, cfg)
	userHandler := handler.NewUserHandler(userService)
	historyHandler := handler.NewHistoryHandler(historyService)

	http.Handle("/stream", http.HandlerFunc(mediaHandler.StreamMedia))
	http.Handle("/api/media", http.HandlerFunc(mediaHandler.GetMedia))
	http.Handle("/api/media/list", http.HandlerFunc(mediaHandler.GetMediaList))
	http.Handle("/hls/playlist", http.HandlerFunc(mediaHandler.HLSPlaylist))
	http.Handle("/hls/segment", http.HandlerFunc(mediaHandler.HLSSegment))
	http.Handle("/api/media/search", http.HandlerFunc(mediaHandler.SearchMedia))
	http.Handle("/api/media/series", http.HandlerFunc(mediaHandler.GetMediaBySeries))
	http.Handle("/api/media/add", http.HandlerFunc(mediaHandler.AddMedia))
	http.Handle("/api/media/delete", http.HandlerFunc(mediaHandler.DeleteMedia))
	http.Handle("/api/series", http.HandlerFunc(mediaHandler.GetSeries))
	http.Handle("/api/series/list", http.HandlerFunc(mediaHandler.GetSeriesList))
	http.Handle("/api/series/add", http.HandlerFunc(mediaHandler.AddSeries))

	http.Handle("/api/user/register", http.HandlerFunc(userHandler.Register))
	http.Handle("/api/user/login", http.HandlerFunc(userHandler.Login))
	http.Handle("/api/user", http.HandlerFunc(userHandler.GetUser))

	http.Handle("/api/history", http.HandlerFunc(historyHandler.GetHistory))
	http.Handle("/api/history/progress", http.HandlerFunc(historyHandler.GetProgress))
	http.Handle("/api/history/save", http.HandlerFunc(historyHandler.SaveProgress))
	http.Handle("/api/history/remove", http.HandlerFunc(historyHandler.RemoveFromHistory))
	http.Handle("/api/history/clear", http.HandlerFunc(historyHandler.ClearHistory))

	http.Handle("/", http.FileServer(http.Dir(cfg.StaticDir)))
	http.Handle("/media/", http.StripPrefix("/media/", http.FileServer(http.Dir(cfg.MediaDir))))
	http.Handle("/thumbnails/", http.StripPrefix("/thumbnails/", http.FileServer(http.Dir(cfg.MediaDir+"/thumbnails"))))

	log.Printf("Server starting on port %s...", cfg.Port)
	log.Printf("Serving static files from: %s", cfg.StaticDir)
	log.Printf("Serving media files from: %s", cfg.MediaDir)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, nil))
}
