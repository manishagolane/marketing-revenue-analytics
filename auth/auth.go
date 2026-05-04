package auth

import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"marketing-revenue-analytics/config"

	"github.com/cristalhq/jwt/v4"
	"go.uber.org/zap"
)

type JWTManager struct {
	signer   jwt.Signer
	verifier jwt.Verifier
	logger   *zap.Logger
}

type Claims struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	TokenType string `json:"token_type"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
}

// NewJWTManager initializes and returns a JWTManager
func NewJWTManager(logger *zap.Logger) *JWTManager {
	jwtSecret := config.GetString("authentication.jwt.user_secret")

	privateKeyBytes, err := hex.DecodeString(jwtSecret)
	if err != nil {
		logger.Fatal("invalid jwt private key", zap.Error(err))
		return nil
	}
	privateKey := ed25519.PrivateKey(privateKeyBytes)
	publicKey := privateKey.Public().(ed25519.PublicKey)

	signer, err := jwt.NewSignerEdDSA(privateKey)
	if err != nil {
		logger.Fatal("failed to create jwt signer", zap.Error(err))
		return nil
	}

	verifier, err := jwt.NewVerifierEdDSA(publicKey)
	if err != nil {
		logger.Fatal("failed to create jwt verifier", zap.Error(err))
		return nil
	}

	return &JWTManager{signer: signer, verifier: verifier, logger: logger}
}

func (jm *JWTManager) GenerateToken(userID, email, role, tokenType string, ttl time.Duration) (string, error) {
	now := time.Now().Unix()
	claims := Claims{
		UserID:    userID,
		Email:     email,
		Role:      role,
		TokenType: tokenType,
		IssuedAt:  now,
		ExpiresAt: time.Now().Add(ttl).Unix(),
	}

	builder := jwt.NewBuilder(jm.signer)
	token, err := builder.Build(claims)
	if err != nil {
		return "", fmt.Errorf("build jwt token: %w", err)
	}

	return token.String(), nil
}

func (jm *JWTManager) ParseToken(tokenString string) (Claims, error) {
	token, err := jwt.Parse([]byte(tokenString), jm.verifier)
	if err != nil {
		return Claims{}, errors.New("invalid token signature")
	}

	var claims Claims
	if err := json.Unmarshal(token.Claims(), &claims); err != nil {
		return Claims{}, errors.New("failed to parse claims")
	}

	if claims.UserID == "" || claims.Role == "" || claims.TokenType == "" {
		return Claims{}, errors.New("invalid token claims")
	}
	if claims.ExpiresAt <= time.Now().Unix() {
		return Claims{}, errors.New("token expired")
	}

	return claims, nil
}
