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

	registerAuthRoutes(api, authHandler, jwtParser, sensitiveLimiter)
	registerVerificationRoutes(api, verificationHandler, jwtParser, sensitiveLimiter)
	registerUserRoutes(api, userHandler, jwtParser)
	registerAdminUserRoutes(api, userHandler, jwtParser)
	registerInternalUserRoutes(api, userHandler, jwtParser)
	registerPublicUserRoutes(api, userHandler)
	registerAdminVerificationRoutes(api, verificationHandler, jwtParser, sensitiveLimiter)

	return engine
}

func registerAuthRoutes(api *gin.RouterGroup, authHandler *handlers.AuthHandler, jwtParser *auth.JWTParser, sensitiveLimiter *middleware.InMemoryLimiter) {
	// Auth session and account lifecycle endpoints.
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
}

func registerVerificationRoutes(api *gin.RouterGroup, verificationHandler *handlers.VerificationHandler, jwtParser *auth.JWTParser, sensitiveLimiter *middleware.InMemoryLimiter) {
	// Verification endpoints for authenticated users.
	verificationRoutes := api.Group("/verifications")
	verificationRoutes.Use(middleware.RequireAuth(jwtParser))
	verificationRoutes.POST("/submit", sensitiveLimiter.Middleware(), verificationHandler.Submit)
	verificationRoutes.GET("/me", verificationHandler.GetMyStatus)
}

func registerUserRoutes(api *gin.RouterGroup, userHandler *handlers.UserHandler, jwtParser *auth.JWTParser) {
	// Self-service user endpoints for authenticated callers.
	userRoutes := api.Group("/users")
	userRoutes.Use(middleware.RequireAuth(jwtParser))

	// Profile core: identity and onboarding lifecycle.
	userRoutes.GET("/me", userHandler.GetMeUser)
	userRoutes.GET("/me/profile", userHandler.GetMe)
	userRoutes.PATCH("/me/profile", userHandler.UpdateMeProfile)
	userRoutes.DELETE("/me/profile", userHandler.DeleteMeProfile)
	userRoutes.GET("/me/onboarding-status", userHandler.GetMeOnboardingStatus)

	// Settings: account, privacy, and notification preferences.
	userRoutes.GET("/me/settings", userHandler.GetMeAccountSettings)
	userRoutes.PATCH("/me/settings", userHandler.UpdateMeAccountSettings)
	userRoutes.GET("/me/settings/privacy", userHandler.GetMePrivacySettings)
	userRoutes.PATCH("/me/settings/privacy", userHandler.UpdateMePrivacySettings)
	userRoutes.GET("/me/settings/notifications", userHandler.GetMeNotificationSettings)
	userRoutes.PATCH("/me/settings/notifications", userHandler.UpdateMeNotificationSettings)

	// Avatar: upload, read, and delete profile media.
	userRoutes.POST("/me/avatar", userHandler.UploadMeAvatar)
	userRoutes.GET("/me/avatar", userHandler.GetMeAvatar)
	userRoutes.DELETE("/me/avatar", userHandler.RemoveMeAvatar)

	// Portfolio: CRUD and ordering for showcase projects.
	userRoutes.POST("/me/portfolio", userHandler.CreateMePortfolioItem)
	userRoutes.GET("/me/portfolio", userHandler.ListMyPortfolioItems)
	userRoutes.PATCH("/me/portfolio/:itemId", userHandler.UpdateMePortfolioItem)
	userRoutes.DELETE("/me/portfolio/:itemId", userHandler.DeleteMePortfolioItem)
	userRoutes.PUT("/me/portfolio:reorder", userHandler.ReorderMePortfolioItems)

	// Employment: work history timeline entries.
	userRoutes.POST("/me/employment", userHandler.CreateMeEmployment)
	userRoutes.GET("/me/employment", userHandler.ListMyEmployment)
	userRoutes.PATCH("/me/employment/:employmentId", userHandler.UpdateMeEmployment)
	userRoutes.DELETE("/me/employment/:employmentId", userHandler.DeleteMeEmployment)

	// Education: academic history entries.
	userRoutes.POST("/me/education", userHandler.CreateMeEducation)
	userRoutes.GET("/me/education", userHandler.ListMyEducation)
	userRoutes.PATCH("/me/education/:educationId", userHandler.UpdateMeEducation)
	userRoutes.DELETE("/me/education/:educationId", userHandler.DeleteMeEducation)

	// Certifications: credential records management.
	userRoutes.POST("/me/certifications", userHandler.CreateMeCertification)
	userRoutes.GET("/me/certifications", userHandler.ListMyCertifications)
	userRoutes.PATCH("/me/certifications/:certificationId", userHandler.UpdateMeCertification)
	userRoutes.DELETE("/me/certifications/:certificationId", userHandler.DeleteMeCertification)

	// Languages: proficiency list upsert and retrieval.
	userRoutes.PUT("/me/languages", userHandler.UpsertMeLanguages)
	userRoutes.GET("/me/languages", userHandler.GetMeLanguages)

	// Freelancer preferences: availability, rates, and work style.
	userRoutes.PUT("/me/availability", userHandler.SetMeAvailability)
	userRoutes.GET("/me/availability", userHandler.GetMeAvailability)
	userRoutes.PUT("/me/rates", userHandler.SetMeRates)
	userRoutes.GET("/me/rates", userHandler.GetMeRates)
	userRoutes.PUT("/me/work-preferences", userHandler.SetMeWorkPreferences)
	userRoutes.GET("/me/work-preferences", userHandler.GetMeWorkPreferences)

	// Client hiring: profile, company, and hiring controls.
	userRoutes.GET("/me/client-profile", userHandler.GetMeClientProfile)
	userRoutes.PATCH("/me/client-profile", userHandler.UpdateMeClientProfile)
	userRoutes.GET("/me/company", userHandler.GetMeCompany)
	userRoutes.PATCH("/me/company", userHandler.UpdateMeCompany)
	userRoutes.GET("/me/hiring-preferences", userHandler.GetMeHiringPreferences)
	userRoutes.PATCH("/me/hiring-preferences", userHandler.UpdateMeHiringPreferences)

	// Saved freelancers: shortlist and recruiter notes.
	userRoutes.POST("/me/saved-freelancers/:freelancerId", userHandler.SaveMeFreelancer)
	userRoutes.GET("/me/saved-freelancers", userHandler.ListMeSavedFreelancers)
	userRoutes.DELETE("/me/saved-freelancers/:freelancerId", userHandler.RemoveMeSavedFreelancer)
	userRoutes.PUT("/me/freelancer-notes/:freelancerId", userHandler.UpsertMeFreelancerNote)
	userRoutes.GET("/me/freelancer-notes/:freelancerId", userHandler.GetMeFreelancerNote)
}

func registerAdminUserRoutes(api *gin.RouterGroup, userHandler *handlers.UserHandler, jwtParser *auth.JWTParser) {
	// Admin-only user management and audit endpoints.
	adminUserRoutes := api.Group("/admin/users")
	adminUserRoutes.Use(middleware.RequireAuth(jwtParser), middleware.RequireRoles("admin"))
	adminUserRoutes.GET("", userHandler.ListUsers)
	adminUserRoutes.GET("/:userId", userHandler.GetUser)
	adminUserRoutes.GET("/:userId/profile", userHandler.GetProfile)
	adminUserRoutes.PATCH("/:userId/account-status", userHandler.UpdateAccountStatus)
	adminUserRoutes.POST("/:userId/impersonation-token", userHandler.CreateImpersonationToken)
	adminUserRoutes.GET("/:userId/audit-summary", userHandler.GetUserAuditSummary)
}

func registerInternalUserRoutes(api *gin.RouterGroup, userHandler *handlers.UserHandler, jwtParser *auth.JWTParser) {
	// Internal admin read models for dependent services.
	internalUserRoutes := api.Group("/internal/users")
	internalUserRoutes.Use(middleware.RequireAuth(jwtParser), middleware.RequireRoles("admin"))
	internalUserRoutes.GET("/:userId/basic", userHandler.GetInternalUserBasic)
	internalUserRoutes.GET("/:userId/profile", userHandler.GetInternalUserProfile)
}

func registerPublicUserRoutes(api *gin.RouterGroup, userHandler *handlers.UserHandler) {
	// Public profile projections available without auth.
	publicRoutes := api.Group("/public")
	publicRoutes.GET("/users/:userId/profile", userHandler.GetPublicProfile)
	publicRoutes.GET("/users/:userId/portfolio", userHandler.ListPublicPortfolioItems)
	publicRoutes.GET("/users/:userId/employment", userHandler.ListPublicEmployment)
	publicRoutes.GET("/users/:userId/education", userHandler.ListPublicEducation)
	publicRoutes.GET("/users/:userId/certifications", userHandler.ListPublicCertifications)
	publicRoutes.GET("/users/:userId/languages", userHandler.GetPublicLanguages)
}

func registerAdminVerificationRoutes(api *gin.RouterGroup, verificationHandler *handlers.VerificationHandler, jwtParser *auth.JWTParser, sensitiveLimiter *middleware.InMemoryLimiter) {
	// Admin verification review and reverification controls.
	adminVerificationRoutes := api.Group("/admin/verifications")
	adminVerificationRoutes.Use(middleware.RequireAuth(jwtParser), middleware.RequireRoles("admin"))
	adminVerificationRoutes.GET("/pending", verificationHandler.ListPending)
	adminVerificationRoutes.GET("/:requestId", verificationHandler.GetByID)
	adminVerificationRoutes.POST("/:requestId/review", sensitiveLimiter.Middleware(), verificationHandler.Review)
	adminVerificationRoutes.POST("/reverification", sensitiveLimiter.Middleware(), verificationHandler.RequestReverification)
}
