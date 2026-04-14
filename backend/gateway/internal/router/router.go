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

func New(cfg config.Config, authHandler *handlers.AuthHandler, verificationHandler *handlers.VerificationHandler, userHandler *handlers.UserHandler, jobHandler *handlers.JobHandler, proposalHandler *handlers.ProposalHandler) *gin.Engine {
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
	registerAdminVerificationRoutes(api, verificationHandler, jwtParser, sensitiveLimiter)
	registerJobRoutes(api, jobHandler, jwtParser)
	registerProposalRoutes(api, proposalHandler, jwtParser)

	return engine
}

func registerProposalRoutes(api *gin.RouterGroup, proposalHandler *handlers.ProposalHandler, jwtParser *auth.JWTParser) {
	proposalRoutes := api.Group("/proposals")
	proposalRoutes.Use(middleware.RequireAuth(jwtParser))
	proposalRoutes.GET("/:proposalId", proposalHandler.GetProposal)
	proposalRoutes.GET("/:proposalId/attachments/:attachmentId/download-url", proposalHandler.GetProposalAttachmentDownloadURL)

	freelancerRoutes := proposalRoutes.Group("")
	freelancerRoutes.Use(middleware.RequireRoles("freelancer"))
	freelancerRoutes.GET("/me/jobs/:jobId", proposalHandler.GetMyProposalForJob)
	freelancerRoutes.GET("/me/jobs/:jobId/has-applied", proposalHandler.HasAppliedToJob)
	freelancerRoutes.GET("/me", proposalHandler.ListMyProposals)
	freelancerRoutes.POST("/:proposalId/attachments/upload-url", proposalHandler.GetProposalAttachmentUploadURL)

	clientRoutes := proposalRoutes.Group("")
	clientRoutes.Use(middleware.RequireRoles("client"))
	clientRoutes.GET("/client", proposalHandler.ListClientProposals)
	clientRoutes.GET("/client/counts", proposalHandler.CountClientProposalInbox)
	clientRoutes.GET("/jobs/:jobId/counts", proposalHandler.CountProposalsByJob)
	clientRoutes.POST("/:proposalId/decision", proposalHandler.SetProposalDecision)
}

func registerJobRoutes(api *gin.RouterGroup, jobHandler *handlers.JobHandler, jwtParser *auth.JWTParser) {
	// Public discovery and detail endpoints.
	publicJobs := api.Group("/jobs")
	publicJobs.GET("", jobHandler.ListOpenJobs)
	publicJobs.GET("/search-v2", jobHandler.SearchJobsV2)
	publicJobs.GET("/:jobId/public", jobHandler.GetPublicJobDetail)

	// Client-owned job operations.
	clientJobs := api.Group("/jobs")
	clientJobs.Use(middleware.RequireAuth(jwtParser), middleware.RequireRoles("client"))
	clientJobs.POST("", jobHandler.CreateJob)
	clientJobs.GET("/me", jobHandler.ListMyJobs)
	clientJobs.GET("/:jobId", jobHandler.GetJob)
	clientJobs.PATCH("/:jobId", jobHandler.UpdateJob)
	clientJobs.POST("/:jobId/close", jobHandler.CloseJob)
	clientJobs.POST("/:jobId/pause", jobHandler.PauseJob)
	clientJobs.POST("/:jobId/reopen", jobHandler.ReopenJob)
	clientJobs.POST("/:jobId/mark-filled", jobHandler.MarkJobFilled)
	clientJobs.POST("/:jobId/visibility", jobHandler.SetJobVisibility)
	clientJobs.POST("/:jobId/budget-range", jobHandler.SetJobBudgetRange)
	clientJobs.POST("/:jobId/invite", jobHandler.InviteFreelancerToJob)
	clientJobs.GET("/:jobId/applicants", jobHandler.ListJobApplicants)
	clientJobs.POST("/applicants/:proposalId/stage", jobHandler.SetApplicantStage)
	clientJobs.POST("/:jobId/reject-all", jobHandler.RejectAllApplicants)
	clientJobs.POST("/:jobId/reopen-hiring", jobHandler.ReopenHiringForJob)
	clientJobs.POST("/hire", jobHandler.HireApplicant)
	clientJobs.GET("/:jobId/stats", jobHandler.GetJobStats)
	clientJobs.POST("/:jobId/mark-completed", jobHandler.MarkJobCompleted)
	clientJobs.POST("/:jobId/cancel-with-settlement", jobHandler.CancelJobWithSettlementPolicy)
	clientJobs.POST("/:jobId/attachments", jobHandler.UploadJobAttachment)
	clientJobs.DELETE("/:jobId/attachments/:attachmentId", jobHandler.DeleteJobAttachment)
	clientJobs.GET("/:jobId/attachments", jobHandler.ListJobAttachments)
	clientJobs.GET("/:jobId/attachments/:attachmentId/download-url", jobHandler.GetJobAttachmentDownloadURL)

	// Freelancer interactions.
	freelancerJobs := api.Group("/jobs")
	freelancerJobs.Use(middleware.RequireAuth(jwtParser), middleware.RequireRoles("freelancer"))
	freelancerJobs.GET("/invited", jobHandler.ListInvitedJobs)
	freelancerJobs.POST("/:jobId/invite-response", jobHandler.RespondToJobInvite)
	freelancerJobs.POST("/:jobId/save", jobHandler.SaveJob)
	freelancerJobs.DELETE("/:jobId/save", jobHandler.UnsaveJob)
	freelancerJobs.GET("/saved", jobHandler.ListSavedJobs)
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
	userRoutes.GET("/me/profile", userHandler.GetMe)
	userRoutes.PATCH("/me/profile", userHandler.UpdateMeProfile)
	userRoutes.DELETE("/me/profile", userHandler.DeleteMeProfile)
	userRoutes.GET("/me/onboarding-status", userHandler.GetMeOnboardingStatus)

	// Settings: account, privacy, and notification preferences.
	userRoutes.GET("/me/settings", userHandler.GetMeAccountSettings)
	userRoutes.PATCH("/me/settings", userHandler.UpdateMeAccountSettings)

	// Avatar: upload, read, and delete profile media.
	userRoutes.POST("/me/avatar/upload-url", userHandler.GetMeAvatarUploadUrl)
	userRoutes.POST("/me/avatar", userHandler.UploadMeAvatar)
	userRoutes.GET("/me/avatar", userHandler.GetMeAvatar)
	userRoutes.DELETE("/me/avatar", userHandler.RemoveMeAvatar)

	// CV: upload, read URL, and delete profile document.
	userRoutes.POST("/me/cv/upload-url", userHandler.GetMeCVUploadUrl)
	userRoutes.POST("/me/cv", userHandler.UploadMeCV)
	userRoutes.GET("/me/cv", userHandler.GetMeCV)
	userRoutes.DELETE("/me/cv", userHandler.RemoveMeCV)

	// Portfolio: freelancer-only media upload reservation + CRUD for showcase projects.
	portfolioRoutes := userRoutes.Group("/me/portfolio")
	portfolioRoutes.Use(middleware.RequireRoles("freelancer"))
	portfolioRoutes.POST("/media/upload-url", userHandler.GetMePortfolioMediaUploadUrl)
	portfolioRoutes.POST("", userHandler.CreateMePortfolioItem)
	portfolioRoutes.GET("", userHandler.ListMePortfolioItems)
	portfolioRoutes.GET("/:itemId", userHandler.GetMePortfolioItem)
	portfolioRoutes.PUT("/:itemId", userHandler.UpdateMePortfolioItem)
	portfolioRoutes.DELETE("/:itemId", userHandler.DeleteMePortfolioItem)

	// Freelancer preferences: work style and client matching.
	userRoutes.PATCH("/me/work-preferences", userHandler.SetMeWorkPreferences)
	userRoutes.GET("/me/work-preferences", userHandler.GetMeWorkPreferences)

	// Client hiring: profile and hiring controls.
	userRoutes.GET("/me/hiring-preferences", userHandler.GetMeHiringPreferences)
	userRoutes.PATCH("/me/hiring-preferences", userHandler.UpdateMeHiringPreferences)

	// Saved freelancers: shortlist and recruiter notes.
	userRoutes.POST("/me/saved-freelancers/:freelancerId", userHandler.SaveMeFreelancer)
	userRoutes.GET("/me/saved-freelancers", userHandler.ListMeSavedFreelancers)
	userRoutes.DELETE("/me/saved-freelancers/:freelancerId", userHandler.RemoveMeSavedFreelancer)
	userRoutes.PUT("/me/freelancer-notes/:freelancerId", userHandler.UpsertMeFreelancerNote)
	userRoutes.GET("/me/freelancer-notes/:freelancerId", userHandler.GetMeFreelancerNote)
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
