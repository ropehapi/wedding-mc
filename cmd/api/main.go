package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/ropehapi/wedding-mc/internal/config"
	"github.com/ropehapi/wedding-mc/internal/handler"
	"github.com/ropehapi/wedding-mc/internal/middleware"
	"github.com/ropehapi/wedding-mc/internal/repository"
	"github.com/ropehapi/wedding-mc/internal/service"
)

func main() {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr}).With().Timestamp().Logger()
	log.Logger = logger

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	if err := config.RunMigrations(cfg); err != nil {
		log.Fatal().Err(err).Msg("failed to run migrations")
	}

	db, err := config.NewDB(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	defer db.Close()

	// --- Repositories ---
	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewRefreshTokenRepository(db)
	weddingRepo := repository.NewWeddingRepository(db)

	// --- Services ---
	authSvc := service.NewAuthService(userRepo, tokenRepo, cfg.JWTSecret, cfg.JWTExpiry, cfg.RefreshExpiry)

	baseURL := fmt.Sprintf("http://localhost:%s/uploads", cfg.Port)
	storageSvc := service.NewLocalStorage(cfg.LocalStoragePath, baseURL)
	weddingSvc := service.NewWeddingService(weddingRepo, storageSvc)

	// --- Handlers ---
	authHandler := handler.NewAuthHandler(authSvc)
	weddingHandler := handler.NewWeddingHandler(weddingSvc)

	// --- Router ---
	r := chi.NewRouter()

	r.Use(chiMiddleware.RealIP)
	r.Use(middleware.Logger(logger))
	r.Use(middleware.Recoverer(logger))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.AllowedOrigins},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	r.Route("/v1", func(r chi.Router) {
		// Auth — public
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
			r.Post("/refresh", authHandler.Refresh)
			r.With(middleware.Auth(cfg.JWTSecret)).Post("/logout", authHandler.Logout)
		})

		// Wedding — protected
		r.Route("/wedding", func(r chi.Router) {
			r.Use(middleware.Auth(cfg.JWTSecret))
			r.Get("/", weddingHandler.Get)
			r.Post("/", weddingHandler.Create)
			r.Patch("/", weddingHandler.Update)
			r.Post("/photos", weddingHandler.UploadPhoto)
			r.Delete("/photos/{photoID}", weddingHandler.DeletePhoto)
		})
	})

	// Serve uploaded files (local storage only)
	if cfg.StorageDriver == "local" {
		r.Handle("/uploads/*", http.StripPrefix("/uploads/",
			http.FileServer(http.Dir(cfg.LocalStoragePath))))
	}

	addr := ":" + cfg.Port
	log.Info().Str("addr", addr).Msg("server starting")
	if err := http.ListenAndServe(addr, r); err != nil {
		log.Fatal().Err(err).Msg("server error")
	}
}
