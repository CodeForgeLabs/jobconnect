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

func New(cfg config.Config, authHandler *handlers.AuthHandler, verificationHandler *handlers.VerificationHandler, userHandler *handlers.UserHandler) *gin.Engine {
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

	userRoutes := api.Group("/users")
	userRoutes.Use(middleware.RequireAuth(jwtParser))
	userRoutes.GET("/me", userHandler.GetMeUser)
	userRoutes.GET("/me/profile", userHandler.GetMe)
	userRoutes.PATCH("/me/profile", userHandler.UpdateMeProfile)
	userRoutes.DELETE("/me/profile", userHandler.DeleteMeProfile)
	userRoutes.GET("/me/onboarding-status", userHandler.GetMeOnboardingStatus)
	userRoutes.GET("/me/settings", userHandler.GetMeAccountSettings)
	userRoutes.PATCH("/me/settings", userHandler.UpdateMeAccountSettings)
	userRoutes.GET("/me/settings/privacy", userHandler.GetMePrivacySettings)
	userRoutes.PATCH("/me/settings/privacy", userHandler.UpdateMePrivacySettings)
	userRoutes.GET("/me/settings/notifications", userHandler.GetMeNotificationSettings)
	userRoutes.PATCH("/me/settings/notifications", userHandler.UpdateMeNotificationSettings)
	userRoutes.POST("/me/avatar", userHandler.UploadMeAvatar)
	userRoutes.GET("/me/avatar", userHandler.GetMeAvatar)
	userRoutes.DELETE("/me/avatar", userHandler.RemoveMeAvatar)
	userRoutes.POST("/me/portfolio", userHandler.CreateMePortfolioItem)
	userRoutes.GET("/me/portfolio", userHandler.ListMyPortfolioItems)
	userRoutes.PATCH("/me/portfolio/:itemId", userHandler.UpdateMePortfolioItem)
	userRoutes.DELETE("/me/portfolio/:itemId", userHandler.DeleteMePortfolioItem)
	userRoutes.PUT("/me/portfolio:reorder", userHandler.ReorderMePortfolioItems)
	userRoutes.POST("/me/employment", userHandler.CreateMeEmployment)
	userRoutes.GET("/me/employment", userHandler.ListMyEmployment)
	userRoutes.PATCH("/me/employment/:employmentId", userHandler.UpdateMeEmployment)
	userRoutes.DELETE("/me/employment/:employmentId", userHandler.DeleteMeEmployment)
	userRoutes.POST("/me/education", userHandler.CreateMeEducation)
	userRoutes.GET("/me/education", userHandler.ListMyEducation)
	userRoutes.PATCH("/me/education/:educationId", userHandler.UpdateMeEducation)
	userRoutes.DELETE("/me/education/:educationId", userHandler.DeleteMeEducation)
	userRoutes.POST("/me/certifications", userHandler.CreateMeCertification)
	userRoutes.GET("/me/certifications", userHandler.ListMyCertifications)
	userRoutes.PATCH("/me/certifications/:certificationId", userHandler.UpdateMeCertification)
	userRoutes.DELETE("/me/certifications/:certificationId", userHandler.DeleteMeCertification)
	userRoutes.PUT("/me/languages", userHandler.UpsertMeLanguages)
	userRoutes.GET("/me/languages", userHandler.GetMeLanguages)

	adminUserRoutes := api.Group("/admin/users")
	adminUserRoutes.Use(middleware.RequireAuth(jwtParser), middleware.RequireRoles("admin"))
	adminUserRoutes.GET("/:userId", userHandler.GetUser)
	adminUserRoutes.GET("/:userId/profile", userHandler.GetProfile)
	adminUserRoutes.PATCH("/:userId/account-status", userHandler.UpdateAccountStatus)

	publicRoutes := api.Group("/public")
	publicRoutes.GET("/users/:userId/profile", userHandler.GetPublicProfile)
	publicRoutes.GET("/users/:userId/portfolio", userHandler.ListPublicPortfolioItems)
	publicRoutes.GET("/users/:userId/employment", userHandler.ListPublicEmployment)
	publicRoutes.GET("/users/:userId/education", userHandler.ListPublicEducation)
	publicRoutes.GET("/users/:userId/certifications", userHandler.ListPublicCertifications)
	publicRoutes.GET("/users/:userId/languages", userHandler.GetPublicLanguages)

	adminVerificationRoutes := api.Group("/admin/verifications")
	adminVerificationRoutes.Use(middleware.RequireAuth(jwtParser), middleware.RequireRoles("admin"))
	adminVerificationRoutes.GET("/pending", verificationHandler.ListPending)
	adminVerificationRoutes.GET("/:requestId", verificationHandler.GetByID)
	adminVerificationRoutes.POST("/:requestId/review", sensitiveLimiter.Middleware(), verificationHandler.Review)
	adminVerificationRoutes.POST("/reverification", sensitiveLimiter.Middleware(), verificationHandler.RequestReverification)

	return engine
}
