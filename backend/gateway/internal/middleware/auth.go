package middleware

import (
	"net/http"
	"strings"

	"jobconnect/gateway/internal/auth"

	"github.com/gin-gonic/gin"
)

const (
	ContextUserID  = "caller_user_id"
	ContextRole    = "caller_role"
	ContextTokenID = "caller_token_id"
)

type AuthParser interface {
	ParseAccessToken(tokenString string) (auth.Caller, error)
}

func OptionalAuth(parser AuthParser) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := bearerTokenFromHeader(c.GetHeader("Authorization"))
		if !ok {
			c.Next()
			return
		}
		caller, err := parser.ParseAccessToken(token)
		if err != nil {
			c.Next()
			return
		}
		setCaller(c, caller)
		c.Next()
	}
}

func RequireAuth(parser AuthParser) gin.HandlerFunc {
	return func(c *gin.Context) {
		token, ok := bearerTokenFromHeader(c.GetHeader("Authorization"))
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing authorization header",
			})
			return
		}

		caller, err := parser.ParseAccessToken(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid access token",
			})
			return
		}

		setCaller(c, caller)
		c.Next()
	}
}

func RequireRoles(roles ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(roles))
	for _, role := range roles {
		allowed[strings.ToLower(strings.TrimSpace(role))] = struct{}{}
	}

	return func(c *gin.Context) {
		v, ok := c.Get(ContextRole)
		if !ok {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		role, _ := v.(string)
		if _, exists := allowed[strings.ToLower(strings.TrimSpace(role))]; !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient role"})
			return
		}
		c.Next()
	}
}

func bearerTokenFromHeader(raw string) (string, bool) {
	parts := strings.SplitN(strings.TrimSpace(raw), " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return "", false
	}
	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", false
	}
	return token, true
}

func setCaller(c *gin.Context, caller auth.Caller) {
	c.Set(ContextUserID, caller.UserID.String())
	c.Set(ContextRole, caller.Role)
	c.Set(ContextTokenID, caller.TokenID)
}
