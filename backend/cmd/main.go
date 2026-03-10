package main

import (
	"log"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"

	"yandex2sound/backend/internal/handlers"
	"yandex2sound/backend/internal/services"
	"yandex2sound/backend/internal/session"
)

func main() {
	_ = godotenv.Load()
	frontendURL := os.Getenv("FRONTEND_URL")

	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOriginsFunc: func(origin string) bool {
			if origin == "" {
				return false
			}
			if frontendURL != "" && origin == frontendURL {
				return true
			}
			return strings.HasPrefix(origin, "http://localhost:") ||
				strings.HasPrefix(origin, "http://127.0.0.1:") ||
				strings.HasPrefix(origin, "https://localhost:") ||
				strings.HasPrefix(origin, "https://127.0.0.1:")
		},
		AllowCredentials: true,
	}))

	app.Use(func(c *fiber.Ctx) error {
		session.Get(c)
		return c.Next()
	})

	yandexService := services.NewYandexService()
	soundCloudService := services.NewSoundCloudService()
	transferService := services.NewTransferService(soundCloudService)

	yandexHandler := handlers.NewYandexHandler(yandexService)
	soundHandler := handlers.NewSoundCloudHandler(soundCloudService)
	transferHandler := handlers.NewTransferHandler(yandexService, transferService)

	api := app.Group("/api")
	api.Get("/yandex/auth/url", yandexHandler.AuthURL)
	api.Get("/yandex/auth/callback", yandexHandler.AuthCallback)
	api.Post("/yandex/auth/validate", yandexHandler.ValidateToken)
	api.Get("/yandex/playlists", yandexHandler.GetPlaylists)
	api.Get("/yandex/playlist/:id/tracks", yandexHandler.GetPlaylistTracks)

	api.Get("/soundcloud/auth/url", soundHandler.AuthURL)
	api.Get("/soundcloud/auth", soundHandler.AuthStart)
	api.Get("/soundcloud/auth/callback", soundHandler.AuthCallback)
	api.Get("/soundcloud/status", soundHandler.Status)

	api.Post("/transfer", transferHandler.StartTransfer)
	api.Get("/transfer/progress/:job_id", transferHandler.ProgressSSE)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Fatal(app.Listen(":" + port))
}
