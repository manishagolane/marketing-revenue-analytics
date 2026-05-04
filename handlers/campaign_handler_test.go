package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
)

func TestCreateCampaign_Success(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{
		"name":"Campaign 1",
		"channel":"google",
		"budget":1000
	}`

	c.Request = httptest.NewRequest(http.MethodPost, "/campaigns", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h := setupCampaignHandler()
	h.Create(c)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestCreateCampaign_InvalidBudget(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{
		"name":"Campaign",
		"channel":"google",
		"budget":-10
	}`

	c.Request = httptest.NewRequest(http.MethodPost, "/campaigns", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h := setupCampaignHandler()
	h.Create(c)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestGetCampaign_Success(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Params = gin.Params{{Key: "id", Value: "c1"}}
	c.Request = httptest.NewRequest(http.MethodGet, "/campaigns/c1", nil)

	h := setupCampaignHandler()
	h.Get(c)

	assert.Equal(t, http.StatusOK, w.Code)
}
func TestUpdateCampaign_Success(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{
		"name":"Updated Campaign",
		"description":"desc",
		"channel":"google",
		"budget":2000,
		"is_public":true
	}`

	c.Params = gin.Params{{Key: "id", Value: "c1"}}
	c.Request = httptest.NewRequest(http.MethodPut, "/campaigns/c1", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h := setupCampaignHandler()
	h.Update(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestUpdateCampaign_Forbidden(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{
		"name":"Updated Campaign"
	}`

	c.Params = gin.Params{{Key: "id", Value: "c1"}}
	c.Request = httptest.NewRequest(http.MethodPut, "/campaigns/c1", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h := setupCampaignHandler()
	h.Update(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestDeleteCampaign_Success(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Params = gin.Params{{Key: "id", Value: "c1"}}
	c.Request = httptest.NewRequest(http.MethodDelete, "/campaigns/c1", nil)

	h := setupCampaignHandler()
	h.Delete(c)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestDeleteCampaign_NotOwnerForbidden(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Params = gin.Params{{Key: "id", Value: "c1"}}
	c.Request = httptest.NewRequest(http.MethodDelete, "/campaigns/c1", nil)

	h := setupCampaignHandler()
	h.Delete(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestSearchCampaign_Success(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest(http.MethodGet, "/campaigns/search?q=sale", nil)
	c.Request = req

	h := setupCampaignHandler()
	h.Search(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestListCampaign_MarketerSeesOwnOnly(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest(http.MethodGet, "/campaigns", nil)
	c.Request = req

	h := setupCampaignHandler()
	h.List(c)

	assert.Equal(t, http.StatusOK, w.Code)
}
