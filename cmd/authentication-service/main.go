package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"github.com/VincentArjuna/RexiErp/internal/authentication/config"
	"github.com/VincentArjuna/RexiErp/internal/authentication/handler"
	"github.com/VincentArjuna/RexiErp/internal/authentication/model"
	"github.com/VincentArjuna/RexiErp/internal/authentication/repository"
	"github.com/VincentArjuna/RexiErp/internal/authentication/service"
	"github.com/VincentArjuna/RexiErp/internal/shared/cache"
	"github.com/VincentArjuna/RexiErp/internal/shared/database"
	"github.com/VincentArjuna/RexiErp/internal/shared/middleware"
)

func main() {
	// Initialize logger
	logger := logrus.New()
	logger.SetFormatter(&logrus.JSONFormatter{})

	// Set log level from environment
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		level = logrus.InfoLevel
	}
	logger.SetLevel(level)

	// Load configuration
	cfg, err := config.LoadAuthServiceConfig()
	if err != nil {
		logger.WithError(err).Fatal("Failed to load configuration")
	}

	logger.WithFields(logrus.Fields{
		"app_name":    cfg.App.Name,
		"version":     cfg.App.Version,
		"environment": cfg.App.Environment,
		"debug":       cfg.App.Debug,
	}).Info("Starting authentication service")

	// Set Gin mode based on environment
	if cfg.App.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	// Initialize database connection
	db, err := database.NewDatabase(&cfg.Database, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to database")
	}
	defer db.Close()

	// Initialize Redis cache
	redisCache, err := cache.NewRedisCache(&cfg.Redis, logger)
	if err != nil {
		logger.WithError(err).Fatal("Failed to connect to Redis")
	}
	defer redisCache.Close()

	// Run database migrations
	if err := model.AutoMigrate(db); err != nil {
		logger.WithError(err).Fatal("Failed to run database migrations")
	}
	logger.Info("Database migrations completed successfully")

	// Initialize repositories
	userRepo := repository.NewUserRepository(db, logger)
	sessionRepo := repository.NewSessionRepository(db, logger)
	activityRepo := repository.NewActivityRepository(db, logger)
	passwordResetRepo := repository.NewPasswordResetRepository(db, logger)

	// Initialize services
	authConfig := service.NewAuthConfig(cfg)
	jwtService := service.NewJWTService(
		cfg.JWT.Secret,
		cfg.JWT.Issuer,
		cfg.JWT.AccessTokenTTL,
		time.Duration(cfg.JWT.RefreshTokenDays)*24*time.Hour,
	)
	authService := service.NewAuthService(
		userRepo,
		sessionRepo,
		activityRepo,
		passwordResetRepo,
		redisCache,
		jwtService,
		logger,
		authConfig,
	)

	// Initialize handlers
	authHandler := handler.NewAuthHandler(authService, logger)

	// Initialize middleware
	jwtMiddleware := middleware.NewJWTMiddleware(authService, logger)
	rbacMiddleware := middleware.NewRBACMiddleware(jwtMiddleware, logger)

	// Create Gin router
	router := gin.New()

	// Add middleware
	router.Use(gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		logger.WithFields(logrus.Fields{
			"method":     param.Method,
			"path":       param.Path,
			"status":     param.StatusCode,
			"latency":    param.Latency,
			"client_ip":  param.ClientIP,
			"user_agent": param.Request.UserAgent(),
		}).Info("HTTP Request")
		return ""
	}))
	router.Use(gin.Recovery())

	// CORS middleware
	router.Use(func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if cfg.IsDevelopment() || origin == "https://rexi-erp.com" || origin == "https://www.rexi-erp.com" {
			c.Header("Access-Control-Allow-Origin", origin)
		}
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Header("Access-Control-Expose-Headers", "Content-Length")
		c.Header("Access-Control-Allow-Credentials", "true")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		// Check database health
		dbHealth := "healthy"
		if err := db.HealthCheck(); err != nil {
			dbHealth = "unhealthy"
		}

		// Check Redis health
		redisHealth := "healthy"
		if err := redisCache.HealthCheck(); err != nil {
			redisHealth = "unhealthy"
		}

		status := "healthy"
		if dbHealth != "healthy" || redisHealth != "healthy" {
			status = "unhealthy"
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":    status,
				"service":   "authentication-service",
				"timestamp": time.Now().UTC(),
				"version":   cfg.App.Version,
				"checks": gin.H{
					"database": dbHealth,
					"redis":    redisHealth,
				},
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status":    status,
			"service":   "authentication-service",
			"timestamp": time.Now().UTC(),
			"version":   cfg.App.Version,
			"checks": gin.H{
				"database": dbHealth,
				"redis":    redisHealth,
			},
		})
	})

	// API routes
	api := router.Group("/api/v1")
	{
		// Public routes (no authentication required)
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.RefreshToken)
			auth.POST("/password-reset", authHandler.RequestPasswordReset)
			auth.GET("/validate-reset-token", authHandler.ValidateResetToken)
			auth.POST("/reset-password", authHandler.ResetPassword)
		}

		// Protected routes (authentication required)
		protected := api.Group("/auth")
		protected.Use(jwtMiddleware.RequireAuth())
		{
			protected.POST("/logout", authHandler.Logout)
			protected.POST("/logout-all", authHandler.LogoutAll)
			protected.GET("/profile", authHandler.GetProfile)
			protected.PUT("/profile", authHandler.UpdateProfile)
			protected.POST("/change-password", authHandler.ChangePassword)
			protected.GET("/sessions", authHandler.GetSessions)
		}

		// RBAC protected routes (for API gateway integration example)
		admin := api.Group("/admin")
		admin.Use(jwtMiddleware.RequireAuth())
		admin.Use(rbacMiddleware.RequireRole("super_admin", "tenant_admin"))
		{
			admin.GET("/users", func(c *gin.Context) {
				userID := c.GetString("user_id")
				tenantID := c.GetString("tenant_id")
				userRole := c.GetString("user_role")

				c.JSON(http.StatusOK, gin.H{
					"message": "Admin access granted",
					"user_id": userID,
					"tenant_id": tenantID,
					"role": userRole,
					"endpoint": "GET /admin/users",
				})
			})
			admin.POST("/users", func(c *gin.Context) {
				userID := c.GetString("user_id")
				userRole := c.GetString("user_role")

				c.JSON(http.StatusOK, gin.H{
					"message": "User creation endpoint accessed",
					"created_by": userID,
					"role": userRole,
					"endpoint": "POST /admin/users",
				})
			})
		}

		// Permission-based protected routes
		orders := api.Group("/orders")
		orders.Use(jwtMiddleware.RequireAuth())
		orders.Use(rbacMiddleware.RequirePermission("orders", "read"))
		{
			orders.GET("", func(c *gin.Context) {
				userID := c.GetString("user_id")
				userRole := c.GetString("user_role")

				c.JSON(http.StatusOK, gin.H{
					"message": "Orders read access granted",
					"user_id": userID,
					"role": userRole,
					"endpoint": "GET /orders",
				})
			})
		}

		ordersWrite := api.Group("/orders")
		ordersWrite.Use(jwtMiddleware.RequireAuth())
		ordersWrite.Use(rbacMiddleware.RequirePermission("orders", "write"))
		{
			ordersWrite.POST("", func(c *gin.Context) {
				userID := c.GetString("user_id")
				userRole := c.GetString("user_role")

				c.JSON(http.StatusOK, gin.H{
					"message": "Orders write access granted",
					"user_id": userID,
					"role": userRole,
					"endpoint": "POST /orders",
				})
			})
		}

		// Service status
		api.GET("/auth/status", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"message": "Authentication service is running",
				"service": "authentication-service",
				"version": cfg.App.Version,
				"environment": cfg.App.Environment,
			})
		})
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.App.Port),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		logger.WithFields(logrus.Fields{
			"port":     cfg.App.Port,
			"service":  "authentication-service",
			"env":      cfg.App.Environment,
			"host":     cfg.App.Host,
		}).Info("Starting authentication service")

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Fatal("Failed to start server")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down authentication service...")

	// Give outstanding requests 30 seconds to complete
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.WithError(err).Fatal("Server forced to shutdown")
	}

	logger.Info("Authentication service stopped")
}