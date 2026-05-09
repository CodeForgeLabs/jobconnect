package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	authv1 "jobconnect/auth/gen/auth/v1"
	"jobconnect/gateway/internal/challenge"
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
	ChallengeID    string `json:"challenge_id" binding:"required"`
	RecaptchaToken string `json:"recaptcha_token" binding:"required"`
}

type AuthErrorResponse struct {
	Error string `json:"error"`
}

type RegisterResponse struct {
	UserID  string `json:"user_id"`
	OtpSent bool   `json:"otp_sent"`
}

type VerifyEmailOTPResponse struct {
	Verified bool `json:"verified"`
}

type LoginResponse struct {
	AccessToken                 string `json:"access_token"`
	AccessTokenExpiresInSeconds int64  `json:"access_token_expires_in_seconds"`
}

type RefreshResponse struct {
	AccessToken                 string `json:"access_token"`
	AccessTokenExpiresInSeconds int64  `json:"access_token_expires_in_seconds"`
}

type LogoutEverywhereResponse struct {
	OK bool `json:"ok"`
}

type ForgotPasswordResponse struct {
	Accepted bool `json:"accepted"`
}

type ResetPasswordResponse struct {
	OK bool `json:"ok"`
}

type RequestEmailChangeResponse struct {
	OtpSent bool `json:"otp_sent"`
}

type ConfirmEmailChangeResponse struct {
	OK bool `json:"ok"`
}

type OAuthCallbackResponse struct {
	AccessToken                 string `json:"access_token"`
	AccessTokenExpiresInSeconds int64  `json:"access_token_expires_in_seconds"`
	IsNewUser                   bool   `json:"is_new_user"`
}

type ListSessionsResponse struct {
	Sessions any `json:"sessions"`
}

type RevokeSessionResponse struct {
	OK bool `json:"ok"`
}

type ChallengeResponse struct {
	ChallengePassed bool    `json:"challenge_passed"`
	ChallengeProof  string  `json:"challenge_proof,omitempty"`
	ChallengeID     string  `json:"challenge_id,omitempty"`
	Score           float64 `json:"score"`
	ExpiresAt       string  `json:"expires_at,omitempty"`
}

type CookieErrorResponse struct {
	Error string `json:"error"`
}

type GenericOKResponse struct {
	OK bool `json:"ok"`
}

type GenericBooleanResponse struct {
	Accepted bool `json:"accepted,omitempty"`
	Verified bool `json:"verified,omitempty"`
}

// Register godoc
// @Summary Register a new account
// @Description Creates a new account and sends an email OTP for verification.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body registerRequest true "Register payload"
// @Success 200 {object} RegisterResponse
// @Failure 400 {object} AuthErrorResponse
// @Failure 500 {object} AuthErrorResponse
// @Router /api/v1/auth/register [post]
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

// VerifyEmailOTP godoc
// @Summary Verify email OTP
// @Description Verifies the email OTP sent during registration.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body verifyEmailOTPRequest true "Email verification payload"
// @Success 200 {object} VerifyEmailOTPResponse
// @Failure 400 {object} AuthErrorResponse
// @Failure 500 {object} AuthErrorResponse
// @Router /api/v1/auth/verify-email-otp [post]
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

// Login godoc
// @Summary Login
// @Description Authenticates a user and sets the refresh cookie.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body loginRequest true "Login payload"
// @Success 200 {object} LoginResponse
// @Failure 400 {object} AuthErrorResponse
// @Failure 500 {object} AuthErrorResponse
// @Router /api/v1/auth/login [post]
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

// Refresh godoc
// @Summary Refresh access token
// @Description Exchanges the refresh cookie for a new access token.
// @Tags Auth
// @Produce json
// @Success 200 {object} RefreshResponse
// @Failure 401 {object} AuthErrorResponse
// @Failure 500 {object} AuthErrorResponse
// @Router /api/v1/auth/refresh [post]
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

// LogoutEverywhere godoc
// @Summary Logout from all sessions
// @Description Revokes all refresh sessions for the current account.
// @Tags Auth
// @Produce json
// @Success 200 {object} LogoutEverywhereResponse
// @Failure 401 {object} AuthErrorResponse
// @Failure 500 {object} AuthErrorResponse
// @Router /api/v1/auth/logout-everywhere [post]
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

// ForgotPassword godoc
// @Summary Start password reset
// @Description Sends a password reset OTP to the account email.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body forgotPasswordRequest true "Forgot password payload"
// @Success 200 {object} ForgotPasswordResponse
// @Failure 400 {object} AuthErrorResponse
// @Failure 500 {object} AuthErrorResponse
// @Router /api/v1/auth/forgot-password [post]
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

// ResetPassword godoc
// @Summary Reset password
// @Description Resets the password using the OTP sent to email.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body resetPasswordRequest true "Reset password payload"
// @Success 200 {object} ResetPasswordResponse
// @Failure 400 {object} AuthErrorResponse
// @Failure 500 {object} AuthErrorResponse
// @Router /api/v1/auth/reset-password [post]
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

// RequestEmailChange godoc
// @Summary Request email change
// @Description Sends an OTP to the new email address for verification.
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body requestEmailChangeRequest true "Email change request"
// @Success 200 {object} RequestEmailChangeResponse
// @Failure 400 {object} AuthErrorResponse
// @Failure 401 {object} AuthErrorResponse
// @Failure 500 {object} AuthErrorResponse
// @Router /api/v1/auth/email-change/request [post]
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

// ConfirmEmailChange godoc
// @Summary Confirm email change
// @Description Confirms the email change using the OTP sent to the new address.
// @Tags Auth
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body confirmEmailChangeRequest true "Email change confirmation"
// @Success 200 {object} ConfirmEmailChangeResponse
// @Failure 400 {object} AuthErrorResponse
// @Failure 401 {object} AuthErrorResponse
// @Failure 500 {object} AuthErrorResponse
// @Router /api/v1/auth/email-change/confirm [post]
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

// OAuthStart godoc
// @Summary Start OAuth login
// @Description Redirects the client to the OAuth provider authorization page.
// @Tags Auth
// @Produce json
// @Param provider path string true "OAuth provider" Enums(google,github)
// @Param role query string false "Desired role"
// @Success 302 {string} string "Redirect to provider authorization URL"
// @Failure 400 {object} AuthErrorResponse
// @Failure 500 {object} AuthErrorResponse
// @Router /api/v1/auth/oauth/{provider}/start [get]
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

// OAuthCallback godoc
// @Summary Complete OAuth login
// @Description Handles the OAuth callback and issues access/refresh tokens.
// @Tags Auth
// @Produce json
// @Param provider path string true "OAuth provider" Enums(google,github)
// @Param state query string true "OAuth state"
// @Param code query string true "OAuth authorization code"
// @Success 200 {object} OAuthCallbackResponse
// @Failure 400 {object} AuthErrorResponse
// @Failure 401 {object} AuthErrorResponse
// @Failure 500 {object} AuthErrorResponse
// @Router /api/v1/auth/oauth/{provider}/callback [get]
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

	if strings.TrimSpace(h.cfg.OAuthFrontendRedirectURL) != "" {
		redirectURL, parseErr := url.Parse(h.cfg.OAuthFrontendRedirectURL)
		if parseErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid oauth frontend redirect url"})
			return
		}
		q := redirectURL.Query()
		q.Set("access_token", resp.GetAccessToken())
		q.Set("access_token_expires_in_seconds", fmt.Sprintf("%d", resp.GetAccessTokenExpiresInSeconds()))
		q.Set("is_new_user", fmt.Sprintf("%t", resp.GetIsNewUser()))
		q.Set("provider", provider)
		redirectURL.RawQuery = q.Encode()
		c.Redirect(http.StatusFound, redirectURL.String())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token":                    resp.GetAccessToken(),
		"access_token_expires_in_seconds": resp.GetAccessTokenExpiresInSeconds(),
		"is_new_user":                     resp.GetIsNewUser(),
	})
}

// ListSessions godoc
// @Summary List sessions
// @Description Returns active login sessions for the authenticated user.
// @Tags Auth
// @Produce json
// @Security BearerAuth
// @Success 200 {object} ListSessionsResponse
// @Failure 401 {object} AuthErrorResponse
// @Failure 500 {object} AuthErrorResponse
// @Router /api/v1/auth/sessions [get]
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

// RevokeSession godoc
// @Summary Revoke session
// @Description Revokes a specific session for the authenticated user.
// @Tags Auth
// @Produce json
// @Security BearerAuth
// @Param sessionId path string true "Session ID"
// @Success 200 {object} RevokeSessionResponse
// @Failure 400 {object} AuthErrorResponse
// @Failure 401 {object} AuthErrorResponse
// @Failure 500 {object} AuthErrorResponse
// @Router /api/v1/auth/sessions/{sessionId} [delete]
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

// Challenge godoc
// @Summary Create challenge proof
// @Description Verifies the CAPTCHA response and returns a short-lived challenge proof.
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body challengeRequest true "Challenge payload"
// @Success 200 {object} ChallengeResponse
// @Failure 400 {object} AuthErrorResponse
// @Failure 401 {object} AuthErrorResponse
// @Failure 503 {object} AuthErrorResponse
// @Failure 500 {object} AuthErrorResponse
// @Router /api/v1/auth/challenge [post]
func (h *AuthHandler) Challenge(c *gin.Context) {
	var req challengeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	passed := false
	score := 0.0
	if h.cfg.RecaptchaDevBypass && strings.TrimSpace(req.RecaptchaToken) == h.cfg.RecaptchaBypassToken {
		passed = true
		score = 1.0
	} else {
		if strings.TrimSpace(h.cfg.RecaptchaSecretKey) == "" {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "recaptcha is not configured"})
			return
		}

		var err error
		passed, score, err = h.verifyRecaptcha(c.Request.Context(), req.RecaptchaToken, c.ClientIP())
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "challenge verification failed"})
			return
		}
	}
	if !passed {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":            "challenge failed",
			"challenge_passed": false,
			"score":            score,
		})
		return
	}

	proof, expiresAt, err := challenge.IssueProof(h.cfg.ChallengeProofSecret, c.ClientIP(), h.cfg.ChallengeProofTTL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to issue challenge proof"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"challenge_passed": true,
		"challenge_proof":  proof,
		"challenge_id":     req.ChallengeID,
		"score":            score,
		"expires_at":       expiresAt.Format(time.RFC3339),
	})
}

func (h *AuthHandler) verifyRecaptcha(ctx context.Context, token string, remoteIP string) (bool, float64, error) {
	form := url.Values{}
	form.Set("secret", h.cfg.RecaptchaSecretKey)
	form.Set("response", strings.TrimSpace(token))
	if remoteIP != "" {
		form.Set("remoteip", remoteIP)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://www.google.com/recaptcha/api/siteverify", strings.NewReader(form.Encode()))
	if err != nil {
		return false, 0, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := h.http.Do(req)
	if err != nil {
		return false, 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return false, 0, fmt.Errorf("recaptcha verify http %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, 0, err
	}
	var out struct {
		Success bool     `json:"success"`
		Score   float64  `json:"score"`
		Errors  []string `json:"error-codes"`
	}
	if err := json.Unmarshal(body, &out); err != nil {
		return false, 0, err
	}
	if !out.Success {
		return false, out.Score, fmt.Errorf("recaptcha rejected")
	}
	if out.Score > 0 && out.Score < h.cfg.RecaptchaMinScore {
		return false, out.Score, nil
	}
	return true, out.Score, nil
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
