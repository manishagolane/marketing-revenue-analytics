package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/oklog/ulid/v2"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"marketing-revenue-analytics/internal/dto"
	"marketing-revenue-analytics/models"
	"marketing-revenue-analytics/utils"
)

type AuthHandler struct {
	queries    AuthStore
	jwtManager TokenService
	cache      CacheStore
}

func NewAuthHandler(q AuthStore, j TokenService, c CacheStore) *AuthHandler {
	return &AuthHandler{queries: q, jwtManager: j, cache: c}
}

type TokenPair struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SendBadRequestError(c, err)
		return
	}

	ctx := c.Request.Context()
	logger := utils.GetCtxLogger(ctx)

	existing, err := h.queries.GetUserByEmail(ctx, req.Email)
	if err == nil && existing.ID != "" {
		SendConflictError(c, errors.New("email already registered"))
		return
	}

	// Prevent self-registration as admin
	if req.Role == "admin" {
		SendForbiddenError(c, errors.New("cannot self-register as admin"))
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error("bcrypt error", zap.Error(err))
		SendApplicationError(c, err)
		return
	}

	user, err := h.queries.CreateUser(ctx, models.CreateUserParams{
		ID:       ulid.Make().String(),
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hash),
		Phone:    req.Phone,
		Role:     req.Role,
	})
	if err != nil {
		logger.Error("create user", zap.Error(err))
		SendApplicationError(c, err)
		return
	}

	// Generate SINGLE access token (24h)
	token, err := h.issueToken(user.ID, user.Email, user.Role)
	if err != nil {
		SendApplicationError(c, err)
		return
	}

	// Store session in Redis
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		SendApplicationError(c, errors.New("invalid token format"))
		return
	}

	sig := parts[2]
	key := "jwt_sig:" + sig

	err = h.cache.Set(ctx, key, user.ID, 24*time.Hour)
	if err != nil {
		logger.Error("failed to store token in redis", zap.Error(err))
		SendApplicationError(c, err)
		return
	}

	// ✅ Response
	c.JSON(http.StatusCreated, dto.APIResponse{
		Status:  "success",
		Message: "registered successfully",
		Data: gin.H{
			"user":  safeUser(user),
			"token": token,
		},
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SendBadRequestError(c, err)
		return
	}

	ctx := c.Request.Context()

	user, err := h.queries.GetUserByEmail(ctx, req.Email)
	if err != nil {
		SendUnauthorizedError(c, errors.New("invalid credentials"))
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		SendUnauthorizedError(c, errors.New("invalid credentials"))
		return
	}

	token, err := h.issueToken(user.ID, user.Email, user.Role)
	if err != nil {
		SendApplicationError(c, err)
		return
	}

	// store in redis
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		SendApplicationError(c, errors.New("invalid token format"))
		return
	}
	sig := parts[2]
	key := fmt.Sprintf("jwt_sig:%s", sig)

	_ = h.cache.Set(ctx, key, user.ID, 24*time.Hour)

	c.JSON(http.StatusOK, dto.APIResponse{
		Status:  "success",
		Message: "logged in successfully",
		Data: gin.H{
			"user":  safeUser(user),
			"token": token,
		},
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	ctx := c.Request.Context()

	tokenStr, _ := GetTokenFromRequest(c, "Authorization")

	parts := strings.Split(tokenStr, ".")
	if len(parts) == 3 {
		sig := parts[2]
		_ = h.cache.Delete(ctx, "jwt_sig:"+sig)
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Status:  "success",
		Message: "logged out successfully",
	})
}

func (h *AuthHandler) GetProfile(c *gin.Context) {
	ctx := c.Request.Context()

	user, err := h.queries.GetUserByID(ctx, c.GetString("userID"))
	if err != nil {
		SendNotFoundError(c, errors.New("user not found"))
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Status:  "success",
		Message: "profile fetched",
		Data:    gin.H{"user": safeUser(user)},
	})
}

func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	var req dto.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		SendBadRequestError(c, err)
		return
	}

	ctx := c.Request.Context()

	user, err := h.queries.UpdateUserProfile(ctx, models.UpdateUserProfileParams{
		ID:      c.GetString("userID"),
		Name:    req.Name,
		Phone:   req.Phone,
		Bio:     utils.ToNullString(req.Bio),
		Picture: utils.ToNullString(req.Picture),
	})
	if err != nil {
		SendApplicationError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.APIResponse{
		Status:  "success",
		Message: "profile updated",
		Data:    gin.H{"user": safeUser(user)},
	})
}

func (h *AuthHandler) issueToken(userID, email, role string) (string, error) {
	return h.jwtManager.GenerateToken(userID, email, role, "access", 24*time.Hour)
}

type SafeUser struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Role      string `json:"role"`
	Bio       string `json:"bio"`
	Picture   string `json:"picture"`
	CreatedAt string `json:"created_at"`
}

func safeUser(u models.User) SafeUser {
	return SafeUser{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		Phone:     u.Phone,
		Role:      u.Role,
		Bio:       u.Bio.String,
		Picture:   u.Picture.String,
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
	}
}

func GetTokenFromRequest(c *gin.Context, authKey string) (string, error) {
	ctx := c.Request.Context()
	logger := utils.GetCtxLogger(ctx)
	authToken, err := c.Request.Cookie(authKey)
	if err != nil {
		authToken := c.Request.Header.Get(authKey)
		if authToken == "" {
			return "", err
		}
		return authToken, nil
	} else {
		logger.Info("authToken",
			zap.String("name", authToken.Name),
			zap.String("value", authToken.Value),
			zap.String("path", authToken.Path),
			zap.String("domain", authToken.Domain),
			zap.Time("expires", authToken.Expires),
			zap.Bool("secure", authToken.Secure),
			zap.Bool("httpOnly", authToken.HttpOnly),
		)
	}

	authTokenValue, err := url.QueryUnescape(authToken.Value)
	if err != nil {
		logger.Error("Failed to decode session key", zap.Error(err))
		return "", err
	}

	return authTokenValue, nil
}
