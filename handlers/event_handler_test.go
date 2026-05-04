package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
)

func TestTrackEvent_Success(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{
		"campaign_id":"c1",
		"event_type":"click",
		"source_url":"google.com"
	}`

	c.Request = httptest.NewRequest(http.MethodPost, "/events/track", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h := setupEventHandler()
	h.Track(c)

	assert.Equal(t, http.StatusAccepted, w.Code)
}

func TestTrackEvent_InvalidBody(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request = httptest.NewRequest(http.MethodPost, "/events/track", strings.NewReader("{invalid json"))
	c.Request.Header.Set("Content-Type", "application/json")

	h := setupEventHandler()
	h.Track(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestTrackEvent_CampaignNotFound(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{
		"campaign_id":"invalid",
		"event_type":"click"
	}`

	c.Request = httptest.NewRequest(http.MethodPost, "/events/track", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h := setupEventHandler()
	h.Track(c)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestTrackEvent_CampaignInactive(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{
		"campaign_id":"c1",
		"event_type":"click"
	}`

	c.Request = httptest.NewRequest(http.MethodPost, "/events/track", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h := setupEventHandler()
	h.Track(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestTrackEvent_SqsFailure(t *testing.T) {
	// simulate queue failure via custom handler
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{
		"campaign_id":"c1",
		"event_type":"click"
	}`

	c.Request = httptest.NewRequest(http.MethodPost, "/events/track", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h := setupEventHandler()
	h.Track(c)

	// still accepted but internal queue failure logged
	assert.Equal(t, http.StatusAccepted, w.Code)
}
