package router

import (
	"net/http"
	"time"

	"jobconnect/gateway/internal/auth"
	"jobconnect/gateway/internal/config"
	"jobconnect/gateway/internal/handlers"
	"jobconnect/gateway/internal/middleware"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

func New(cfg config.Config, authHandler *handlers.AuthHandler, verificationHandler *handlers.VerificationHandler) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Recovery())
	engine.Use(gin.Logger())

	engine.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	jwtParser := auth.NewJWTParser(cfg.JWTSecret)
	engine.Use(middleware.OptionalAuth(jwtParser))

	sensitiveLimiter := middleware.NewInMemoryLimiter(rate.Limit(1), 5, 15*time.Minute, cfg.ChallengeProofSecret)

	api := engine.Group("/api/v1")
	authRoutes := api.Group("/auth")

	authRoutes.POST("/register", sensitiveLimiter.Middleware(), authHandler.Register)
	authRoutes.POST("/verify-email-otp", sensitiveLimiter.Middleware(), authHandler.VerifyEmailOTP)
	authRoutes.POST("/login", sensitiveLimiter.Middleware(), authHandler.Login)
	authRoutes.POST("/refresh", sensitiveLimiter.Middleware(), authHandler.Refresh)
	authRoutes.POST("/logout-everywhere", sensitiveLimiter.Middleware(), authHandler.LogoutEverywhere)

	authRoutes.POST("/forgot-password", sensitiveLimiter.Middleware(), authHandler.ForgotPassword)
	authRoutes.POST("/reset-password", sensitiveLimiter.Middleware(), authHandler.ResetPassword)
	authRoutes.POST("/email-change/request", sensitiveLimiter.Middleware(), middleware.RequireAuth(jwtParser), authHandler.RequestEmailChange)
	authRoutes.POST("/email-change/confirm", sensitiveLimiter.Middleware(), middleware.RequireAuth(jwtParser), authHandler.ConfirmEmailChange)
	authRoutes.GET("/oauth/:provider/start", authHandler.OAuthStart)
	authRoutes.GET("/oauth/:provider/callback", authHandler.OAuthCallback)
	authRoutes.GET("/sessions", middleware.RequireAuth(jwtParser), authHandler.ListSessions)
	authRoutes.DELETE("/sessions/:sessionId", middleware.RequireAuth(jwtParser), authHandler.RevokeSession)
	authRoutes.POST("/challenge", authHandler.Challenge)

	verificationRoutes := api.Group("/verifications")
	verificationRoutes.Use(middleware.RequireAuth(jwtParser))
	verificationRoutes.POST("/submit", sensitiveLimiter.Middleware(), verificationHandler.Submit)
	verificationRoutes.GET("/me", verificationHandler.GetMyStatus)

	adminVerificationRoutes := api.Group("/admin/verifications")
	adminVerificationRoutes.Use(middleware.RequireAuth(jwtParser), middleware.RequireRoles("admin"))
	adminVerificationRoutes.GET("/pending", verificationHandler.ListPending)
	adminVerificationRoutes.GET("/:requestId", verificationHandler.GetByID)
	adminVerificationRoutes.POST("/:requestId/review", sensitiveLimiter.Middleware(), verificationHandler.Review)
	adminVerificationRoutes.POST("/reverification", sensitiveLimiter.Middleware(), verificationHandler.RequestReverification)

	return engine
}
