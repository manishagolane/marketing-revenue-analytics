package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
)

func TestDailyMetrics_Success(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest("GET", "/analytics/daily", nil)

	h := setupAnalyticsHandler()
	h.Daily(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestPublicSummary_Success(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Params = gin.Params{{Key: "id", Value: "c1"}}
	c.Request = httptest.NewRequest("GET", "/analytics/summary", nil)

	h := setupAnalyticsHandler()
	h.PublicSummary(c)

	assert.Equal(t, http.StatusOK, w.Code)
}
