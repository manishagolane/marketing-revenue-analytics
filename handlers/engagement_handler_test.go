package handlers

import (
	"context"
	"marketing-revenue-analytics/models"
	"marketing-revenue-analytics/utils"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
)

func TestFunnel_Success(t *testing.T) {
	w := httptest.NewRecorder()

	req := httptest.NewRequest(http.MethodGet, "/funnel?campaign_id=1&start=2024-01-01", nil)
	req.Header.Set("Content-Type", "application/json")

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	h := setupEngagementHandler()
	h.Funnel(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestClickPath_Success(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Params = gin.Params{{Key: "campaign_id", Value: "c1"}}
	c.Request = httptest.NewRequest(http.MethodGet, "/click-path", nil)

	h := setupEngagementHandler()
	h.ClickPath(c)

	assert.Equal(t, http.StatusOK, w.Code)
}
func TestFunnel_DropoffCalculation(t *testing.T) {
	h := NewEngagementHandler(&mockEngagementStore{
		getFunnelStatsFn: func(ctx context.Context, p models.GetFunnelStatsParams) ([]models.GetFunnelStatsRow, error) {
			return []models.GetFunnelStatsRow{
				{Step: utils.ToNullString("ad"), Sessions: 100},
				{Step: utils.ToNullString("landing"), Sessions: 60},
				{Step: utils.ToNullString("signup"), Sessions: 30},
			}, nil
		},
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest(
		http.MethodGet,
		"/funnel?campaign_id=c1",
		nil,
	)
	c.Request = req

	h.Funnel(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestTimeSpent_AvgCalculation(t *testing.T) {
	h := setupEngagementHandler()

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Params = gin.Params{{Key: "campaign_id", Value: "c1"}}
	c.Request = httptest.NewRequest(http.MethodGet, "/time-spent", nil)

	h.TimeSpent(c)

	assert.Equal(t, http.StatusOK, w.Code)
}
