package server

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"marketing-revenue-analytics/auth"
	"marketing-revenue-analytics/cache"
	"marketing-revenue-analytics/constants"
	"marketing-revenue-analytics/handlers"
	"marketing-revenue-analytics/models"
	"marketing-revenue-analytics/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

func ginLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		c.Next()

		logger := utils.GetCtxLogger(c.Request.Context())
		cost := time.Since(start)
		logger.Info(path,
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.String("user-agent", c.Request.UserAgent()),
			zap.String("errors", c.Errors.ByType(gin.ErrorTypePrivate).String()),
			zap.Duration("cost", cost),
		)
	}
}

func ginRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				httpRequest, _ := httputil.DumpRequest(c.Request, true)
				logger := utils.GetCtxLogger(c.Request.Context())
				if brokenPipe {
					logger.Sugar().Error(c.Request.URL.Path,
						zap.Any("error", err),
						zap.String("requestId", c.Request.Context().Value(constants.REQUEST_ID).(string)))
					logger.Sugar().Error(string(httpRequest))
					c.Error(err.(error))
					c.Abort()
					return
				}
				logger.Sugar().Error(err)
				logger.Sugar().Error(string(debug.Stack()))
				logger.Sugar().Error("[raw http request] ", string(httpRequest))
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"message": "Something went wrong, please try again later",
				})
			}
		}()
		c.Next()
	}
}

func requestIdInterceptor() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestId := uuid.NewString()
		r := c.Request.Context()
		updatedContext := context.WithValue(r, constants.REQUEST_ID, requestId)
		c.Request = c.Request.WithContext(updatedContext)
		c.Writer.Header().Add("x-request-id", requestId)
		c.Next()
	}
}

func authMiddleware(cache *cache.Cache, jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		tokenStr, err := handlers.GetTokenFromRequest(c, constants.API_TOKEN)
		if err != nil {
			handlers.SendUnauthorizedError(c, errors.New("missing token"))
			c.Abort()
			return
		}
		parts := strings.Split(tokenStr, ".")
		if len(parts) != 3 {
			handlers.SendUnauthorizedError(c, errors.New("invalid token"))
			c.Abort()
			return
		}

		sig := parts[2]
		key := fmt.Sprintf("jwt_sig:%s", sig)

		_, err = cache.Get(ctx, key)
		if err != nil {
			handlers.SendUnauthorizedError(c, errors.New("session expired or logged out"))
			c.Abort()
			return
		}

		claims, err := jwtManager.ParseToken(tokenStr)
		if err != nil {
			handlers.SendUnauthorizedError(c, errors.New("invalid token"))
			c.Abort()
			return
		}

		_ = cache.Set(ctx, key, claims.UserID, time.Until(time.Unix(claims.ExpiresAt, 0)))

		c.Set("userID", claims.UserID)
		c.Set("roleID", claims.Role)

		c.Next()
	}
}

func rbacMiddleware(queries *models.Queries, cache *cache.Cache, requiredObject string, requiredAction string) gin.HandlerFunc {
	_ = queries
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		logger := utils.GetCtxLogger(ctx).With(zap.String("source", "rbacmiddleware"))

		userID, exists := c.Get("userID")
		roleID, roleExists := c.Get("roleID")

		if !exists || !roleExists {
			logger.Error("user_id or role_id doesn't exist")
			handlers.SendUnauthorizedError(c, errors.New("unauthorized"))
			c.Abort()
			return
		}

		userIDStr, ok := userID.(string)
		if !ok {
			logger.Error("invalid userID type")
			handlers.SendUnauthorizedError(c, errors.New("unauthorized"))
			c.Abort()
			return

		}
		roleIDStr, ok := roleID.(string)
		if !ok {
			logger.Error("invalid roleID type")
			handlers.SendUnauthorizedError(c, errors.New("unauthorized"))
			c.Abort()
			return

		}

		// Cache key specific to (roleID, Object, Action)
		cacheKey := fmt.Sprintf("permission:%s:%s:%s", roleIDStr, requiredObject, requiredAction)

		// Try getting permission from cache
		cachedPermission, err := cache.GetBool(ctx, cacheKey)
		if err == nil && cachedPermission {
			c.Next()
			return
		}

		hasPermission := hasRolePermission(roleIDStr, requiredObject, requiredAction, userIDStr)

		cache.SetBool(ctx, cacheKey, hasPermission, time.Minute*10)

		// Allow or deny access
		if hasPermission {
			c.Next()
			return
		}

		handlers.SendForbiddenError(c, errors.New("permission denied"))
		c.Abort()
		return

	}
}

func hasRolePermission(roleID, object, action, userID string) bool {
	_ = userID
	permission := fmt.Sprintf("%s:%s", object, action)
	rolePermissions := map[string]map[string]bool{
		"admin": {
			"campaign:create": true, "campaign:read": true, "campaign:update": true, "campaign:delete": true,
			"stats:read": true, "export:create": true, "profile:update": true, "profile:read": true,
		},
		"marketer": {
			"campaign:create": true, "campaign:read": true, "campaign:update": true, "campaign:delete": true,
			"stats:read": true, "export:create": true, "profile:update": true, "profile:read": true,
		},
		"analyst": {
			"campaign:read": true, "stats:read": true, "export:create": true, "profile:read": true,
		},
	}
	perms, ok := rolePermissions[roleID]
	if !ok {
		return false
	}
	return perms[permission]
}

func apiKeyMiddleware(cache *cache.Cache, queries *models.Queries) gin.HandlerFunc {
	_ = cache
	_ = queries
	return func(c *gin.Context) {
		apiKey := c.GetHeader("x-api-key")
		signature := c.GetHeader("x-signature")
		ts := c.GetHeader("x-timestamp")
		if apiKey == "" || signature == "" || ts == "" {
			handlers.SendUnauthorizedError(c, errors.New("missing api key signature headers"))
			c.Abort()
			return
		}
		c.Next()
	}
}

func rateLimitMiddleware(cache *cache.Cache, limit int) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// Use user ID if authenticated, IP if not
		identifier := c.ClientIP()
		if userID, exists := c.Get("userID"); exists {
			identifier = fmt.Sprintf("user:%v", userID)
		}

		key := fmt.Sprintf("rl:%s", identifier)
		cmd := cache.GetClient().B().Incr().Key(key).Build()
		count, err := cache.GetClient().Do(ctx, cmd).AsInt64()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"message": "rate limiter unavailable",
			})
			return
		}
		if count == 1 {
			expire := cache.GetClient().B().Expire().Key(key).Seconds(60).Build()
			cache.GetClient().Do(ctx, expire)
		}

		c.Header("X-RateLimit-Limit", strconv.Itoa(limit))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(max(0, limit-int(count))))

		if int(count) > limit {
			c.Header("Retry-After", "60")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"status":  "failure",
				"message": "rate limit exceeded, try again in 60 seconds",
			})
			return
		}
		c.Next()
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
