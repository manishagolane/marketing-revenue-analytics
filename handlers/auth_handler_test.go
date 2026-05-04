package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/assert/v2"
)

func TestRegister_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{
		"name":"Manisha",
		"email":"new@test.com",
		"password":"password123",
		"phone":"9876543210",
		"role":"marketer"
	}`

	c.Request = httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h := setupAuthHandler()
	h.Register(c)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestRegister_AdminForbidden(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{
		"name":"Admin",
		"email":"admin@test.com",
		"password":"password123",
		"phone":"9876543210",
		"role":"admin"
	}`

	c.Request = httptest.NewRequest(http.MethodPost, "/register", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h := setupAuthHandler()
	h.Register(c)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestLogin_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	body := `{
		"email":"exists@test.com",
		"password":"password123"
	}`

	c.Request = httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h := setupAuthHandler()
	h.Login(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestLogin_InvalidPassword(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// user exists but wrong password
	body := `{
		"email":"exists@test.com",
		"password":"wrongpassword"
	}`

	c.Request = httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")

	h := setupAuthHandler()
	h.Login(c)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}
