package utils

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"marketing-revenue-analytics/constants"
	"marketing-revenue-analytics/logger"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgtype"
	ulid "github.com/oklog/ulid/v2"
	"go.uber.org/zap"
)

func GetLogger() *zap.Logger {
	return logger.GetLogger()
}

func GetCtxLogger(ctx context.Context) *zap.Logger {
	lgr := GetLogger()
	if requestID, ok := ctx.Value(constants.REQUEST_ID).(string); ok {
		lgr = lgr.With(zap.String("requestID", requestID))
	}
	if userID, ok := ctx.Value(constants.UserID).(string); ok {
		lgr = lgr.With(zap.String(string(constants.UserID), userID))
	}
	return lgr
}

// ToCamelCase lowercases the first letter of a string.
// Used for pretty validation error messages.
func ToCamelCase(value string) string {
	if value == "" {
		return ""
	}
	return fmt.Sprintf("%s%s", strings.ToLower(string(value[0])), value[1:])
}

func ToNumeric(v float64) pgtype.Numeric {
	var n pgtype.Numeric
	_ = n.Set(v)
	return n
}

func ToNullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{
		Time:  *t,
		Valid: true,
	}
}

func ToNullBool(b *bool) sql.NullBool {
	if b == nil {
		return sql.NullBool{Valid: false}
	}
	return sql.NullBool{Bool: *b, Valid: true}
}

// utils/utils.go
func ToJSONB(m map[string]interface{}) (pgtype.JSONB, error) {
	if m == nil {
		return pgtype.JSONB{Bytes: []byte("{}"), Status: pgtype.Present}, nil
	}
	b, err := json.Marshal(m)
	if err != nil {
		return pgtype.JSONB{}, err
	}
	return pgtype.JSONB{Bytes: b, Status: pgtype.Present}, nil
}

func ToNullStringPtr(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: *s, Valid: true}
}

func ToNullDate(s string) sql.NullTime {
	if s == "" {
		return sql.NullTime{Valid: false}
	}
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: t, Valid: true}
}

func ToNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

func GetCurrentTime() time.Time {
	location, _ := time.LoadLocation(constants.TIMEZONE)
	return time.Now().In(location)
}

// GetUlid generates a new ULID string.
func GetUlid() (string, error) {
	entropy := rand.New(rand.NewSource(time.Now().UnixNano()))
	ms := ulid.Timestamp(time.Now())
	id, err := ulid.New(ms, entropy)
	if err != nil {
		return "", errors.New("failed to generate ULID")
	}
	return id.String(), nil
}

// GetLoggedInUser extracts userID from gin context set by authMiddleware.
func GetLoggedInUser(c *gin.Context) (string, error) {
	val, exists := c.Get("userID")
	if !exists {
		return "", errors.New("userID not found in context")
	}
	userID, ok := val.(string)
	if !ok {
		return "", errors.New("invalid userID type")
	}
	return userID, nil
}

func GetEndOfDayTime() time.Time {
	location, _ := time.LoadLocation(constants.TIMEZONE)
	year, month, day := time.Now().In(location).Date()
	return time.Date(year, month, day, 23, 59, 59, 0, location)
}
