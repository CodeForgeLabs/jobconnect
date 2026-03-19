package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	authv1 "jobconnect/auth/gen/auth/v1"
	"jobconnect/gateway/internal/config"
	"jobconnect/gateway/internal/middleware"
	"jobconnect/gateway/internal/oauth"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	cfg    config.Config
	client authv1.AuthServiceClient
	http   *http.Client
}

func NewAuthHandler(cfg config.Config, client authv1.AuthServiceClient) *AuthHandler {
	return &AuthHandler{
		cfg:    cfg,
		client: client,
		http:   &http.Client{Timeout: 15 * time.Second},
	}
}

type registerRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required"`
	FirstName   string `json:"first_name" binding:"required"`
	LastName    string `json:"last_name" binding:"required"`
	Role        string `json:"role" binding:"required"`
	AcceptTerms bool   `json:"accept_terms"`
}

type verifyEmailOTPRequest struct {
	Email string `json:"email" binding:"required,email"`
	OTP   string `json:"otp" binding:"required"`
}

type loginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type forgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type resetPasswordRequest struct {
	Email       string `json:"email" binding:"required,email"`
	OTP         string `json:"otp" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

type requestEmailChangeRequest struct {
	NewEmail string `json:"new_email" binding:"required,email"`
}

type confirmEmailChangeRequest struct {
	OTP string `json:"otp" binding:"required"`
}

type challengeRequest struct {
	ChallengeID string `json:"challenge_id" binding:"required"`
	Response    string `json:"response" binding:"required"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.client.Register(c.Request.Context(), &authv1.RegisterRequest{
		Email:       req.Email,
		Password:    req.Password,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		Role:        req.Role,
		AcceptTerms: req.AcceptTerms,
	})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user_id":  resp.GetUserId(),
		"otp_sent": resp.GetOtpSent(),
	})
}

func (h *AuthHandler) VerifyEmailOTP(c *gin.Context) {
	var req verifyEmailOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.client.VerifyEmailOTP(c.Request.Context(), &authv1.VerifyEmailOTPRequest{
		Email: req.Email,
		Otp:   req.OTP,
	})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"verified": resp.GetVerified()})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.client.Login(c.Request.Context(), &authv1.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	h.setRefreshCookie(c, resp.GetRefreshToken(), int(h.cfg.RefreshCookieMaxAge.Seconds()))

	c.JSON(http.StatusOK, gin.H{
		"access_token":                    resp.GetAccessToken(),
		"access_token_expires_in_seconds": resp.GetAccessTokenExpiresInSeconds(),
	})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	refreshToken, err := c.Cookie(h.cfg.RefreshCookieName)
	if err != nil || refreshToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing refresh cookie"})
		return
	}

	resp, rpcErr := h.client.Refresh(c.Request.Context(), &authv1.RefreshRequest{RefreshToken: refreshToken})
	if rpcErr != nil {
		writeGRPCError(c, rpcErr)
		return
	}

	h.setRefreshCookie(c, resp.GetRefreshToken(), int(h.cfg.RefreshCookieMaxAge.Seconds()))

	c.JSON(http.StatusOK, gin.H{
		"access_token":                    resp.GetAccessToken(),
		"access_token_expires_in_seconds": resp.GetAccessTokenExpiresInSeconds(),
	})
}

func (h *AuthHandler) LogoutEverywhere(c *gin.Context) {
	refreshToken, err := c.Cookie(h.cfg.RefreshCookieName)
	if err != nil || refreshToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing refresh cookie"})
		return
	}

	resp, rpcErr := h.client.LogoutEverywhere(c.Request.Context(), &authv1.LogoutEverywhereRequest{RefreshToken: refreshToken})
	if rpcErr != nil {
		writeGRPCError(c, rpcErr)
		return
	}

	h.clearRefreshCookie(c)
	c.JSON(http.StatusOK, gin.H{"ok": resp.GetOk()})
}

func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req forgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.client.ForgotPassword(c.Request.Context(), &authv1.ForgotPasswordRequest{Email: req.Email})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"accepted": resp.GetAccepted()})
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req resetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := h.client.ResetPassword(c.Request.Context(), &authv1.ResetPasswordRequest{
		Email:       req.Email,
		Otp:         req.OTP,
		NewPassword: req.NewPassword,
	})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": resp.GetOk()})
}

func (h *AuthHandler) RequestEmailChange(c *gin.Context) {
	userIDVal, hasUser := c.Get(middleware.ContextUserID)
	if !hasUser {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	userID, _ := userIDVal.(string)

	var req requestEmailChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.client.RequestEmailChange(c.Request.Context(), &authv1.RequestEmailChangeRequest{
		UserId:   userID,
		NewEmail: req.NewEmail,
	})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"otp_sent": resp.GetOtpSent()})
}

func (h *AuthHandler) ConfirmEmailChange(c *gin.Context) {
	userIDVal, hasUser := c.Get(middleware.ContextUserID)
	if !hasUser {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	userID, _ := userIDVal.(string)

	var req confirmEmailChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.client.ConfirmEmailChange(c.Request.Context(), &authv1.ConfirmEmailChangeRequest{
		UserId: userID,
		Otp:    req.OTP,
	})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": resp.GetOk()})
}

func (h *AuthHandler) OAuthStart(c *gin.Context) {
	provider := strings.ToLower(strings.TrimSpace(c.Param("provider")))
	if provider == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "provider is required"})
		return
	}

	providerCfg, err := h.providerConfig(provider)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	role := strings.TrimSpace(c.Query("role"))
	state, err := oauth.IssueState(h.cfg.OAuthStateSecret, provider, role, 10*time.Minute)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to issue oauth state"})
		return
	}
	authURL, err := oauth.AuthURL(provider, providerCfg, state)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Redirect(http.StatusFound, authURL)
}

func (h *AuthHandler) OAuthCallback(c *gin.Context) {
	provider := strings.ToLower(strings.TrimSpace(c.Param("provider")))
	if provider == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "provider is required"})
		return
	}

	state := c.Query("state")
	code := c.Query("code")
	if state == "" || code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing oauth code/state"})
		return
	}

	claims, err := oauth.ParseState(h.cfg.OAuthStateSecret, state, provider)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid oauth state"})
		return
	}

	providerCfg, err := h.providerConfig(provider)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	providerUser, err := oauth.ExchangeAndFetchUser(h.http, provider, providerCfg, code)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "oauth exchange failed"})
		return
	}

	resp, err := h.client.OAuthLogin(c.Request.Context(), &authv1.OAuthLoginRequest{
		Provider:       provider,
		ProviderUserId: providerUser.ProviderUserID,
		Email:          providerUser.Email,
		FirstName:      providerUser.FirstName,
		LastName:       providerUser.LastName,
		DisplayName:    providerUser.DisplayName,
		Role:           claims.Role,
	})
	if err != nil {
		writeGRPCError(c, err)
		return
	}

	h.setRefreshCookie(c, resp.GetRefreshToken(), int(h.cfg.RefreshCookieMaxAge.Seconds()))
	c.JSON(http.StatusOK, gin.H{
		"access_token":                    resp.GetAccessToken(),
		"access_token_expires_in_seconds": resp.GetAccessTokenExpiresInSeconds(),
		"is_new_user":                     resp.GetIsNewUser(),
	})
}

func (h *AuthHandler) ListSessions(c *gin.Context) {
	userIDVal, hasUser := c.Get(middleware.ContextUserID)
	if !hasUser {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}
	userID, _ := userIDVal.(string)
	resp, err := h.client.ListSessions(c.Request.Context(), &authv1.ListSessionsRequest{UserId: userID})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"sessions": resp.GetSessions()})
}

func (h *AuthHandler) RevokeSession(c *gin.Context) {
	userIDVal, hasUser := c.Get(middleware.ContextUserID)
	if !hasUser {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	sessionID := c.Param("sessionId")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "sessionId is required"})
		return
	}

	userID, _ := userIDVal.(string)
	resp, err := h.client.RevokeSession(c.Request.Context(), &authv1.RevokeSessionRequest{
		UserId:    userID,
		SessionId: sessionID,
	})
	if err != nil {
		writeGRPCError(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": resp.GetOk()})
}

func (h *AuthHandler) Challenge(c *gin.Context) {
	var req challengeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Response != "human" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":            "challenge failed",
			"challenge_passed": false,
		})
		return
	}

	expiresAt := time.Now().Add(2 * time.Minute).UTC()
	c.JSON(http.StatusOK, gin.H{
		"challenge_passed": true,
		"challenge_proof":  req.ChallengeID + ":ok",
		"expires_at":       expiresAt.Format(time.RFC3339),
	})
}

func (h *AuthHandler) setRefreshCookie(c *gin.Context, token string, maxAgeSeconds int) {
	if token == "" {
		return
	}
	c.SetSameSite(h.cfg.RefreshCookieSameSite)
	c.SetCookie(
		h.cfg.RefreshCookieName,
		token,
		maxAgeSeconds,
		h.cfg.RefreshCookiePath,
		h.cfg.RefreshCookieDomain,
		h.cfg.RefreshCookieSecure,
		h.cfg.RefreshCookieHTTPOnly,
	)
}

func (h *AuthHandler) clearRefreshCookie(c *gin.Context) {
	c.SetSameSite(h.cfg.RefreshCookieSameSite)
	c.SetCookie(
		h.cfg.RefreshCookieName,
		"",
		-1,
		h.cfg.RefreshCookiePath,
		h.cfg.RefreshCookieDomain,
		h.cfg.RefreshCookieSecure,
		h.cfg.RefreshCookieHTTPOnly,
	)
}

func notImplemented(c *gin.Context, endpoint string, reason string) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error":    "endpoint scaffolded but not wired",
		"endpoint": endpoint,
		"reason":   reason,
	})
}

func (h *AuthHandler) providerConfig(provider string) (oauth.ProviderConfig, error) {
	switch provider {
	case "google":
		if h.cfg.OAuthGoogleClientID == "" || h.cfg.OAuthGoogleClientSecret == "" || h.cfg.OAuthGoogleRedirectURI == "" {
			return oauth.ProviderConfig{}, fmt.Errorf("google oauth is not configured")
		}
		return oauth.ProviderConfig{
			ClientID:     h.cfg.OAuthGoogleClientID,
			ClientSecret: h.cfg.OAuthGoogleClientSecret,
			RedirectURI:  h.cfg.OAuthGoogleRedirectURI,
		}, nil
	case "github":
		if h.cfg.OAuthGitHubClientID == "" || h.cfg.OAuthGitHubClientSecret == "" || h.cfg.OAuthGitHubRedirectURI == "" {
			return oauth.ProviderConfig{}, fmt.Errorf("github oauth is not configured")
		}
		return oauth.ProviderConfig{
			ClientID:     h.cfg.OAuthGitHubClientID,
			ClientSecret: h.cfg.OAuthGitHubClientSecret,
			RedirectURI:  h.cfg.OAuthGitHubRedirectURI,
		}, nil
	default:
		return oauth.ProviderConfig{}, fmt.Errorf("unsupported provider")
	}
}
