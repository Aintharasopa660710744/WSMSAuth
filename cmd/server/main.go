package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/yourorg/auth-service/internal/config"
	"github.com/yourorg/auth-service/internal/handler"
	"github.com/yourorg/auth-service/internal/middleware"
	"github.com/yourorg/auth-service/internal/repository"
	"github.com/yourorg/auth-service/internal/service"
	jwtpkg "github.com/yourorg/auth-service/pkg/jwt"
)

func main() {
	cfg := config.Load()
	gin.SetMode(cfg.Server.Mode)

	// ── Database ──────────────────────────────────────────────────────────────
	db, err := pgxpool.New(context.Background(), cfg.Database.DSN)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(context.Background()); err != nil {
		log.Fatalf("database ping failed: %v", err)
	}
	log.Println("✅ Database connected")

	// ── Dependencies ──────────────────────────────────────────────────────────
	jwtManager := jwtpkg.NewManager(
		cfg.JWT.AccessSecret,
		cfg.JWT.RefreshSecret,
		cfg.JWT.AccessTokenExpiry,
		cfg.JWT.RefreshTokenExpiry,
	)
	userRepo := repository.NewUserRepository(db)
	authSvc  := service.NewAuthService(userRepo, jwtManager)
	authH    := handler.NewAuthHandler(authSvc)

	// ── Router ────────────────────────────────────────────────────────────────
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	auth := r.Group("/auth")
	{
		// Public endpoints
		auth.POST("/register", authH.Register)
		auth.POST("/login", authH.Login)
		auth.POST("/refresh", authH.Refresh)

		// Protected endpoints (require valid access token)
		protected := auth.Group("")
		protected.Use(middleware.AuthMiddleware(jwtManager))
		{
			protected.GET("/me", authH.Me)
			protected.POST("/validate", authH.Validate) // for other services
		}
	}

	// ── HTTP Server with graceful shutdown ────────────────────────────────────
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("🚀 Auth service running on port %s", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}
	log.Println("Server exited cleanly")
}
