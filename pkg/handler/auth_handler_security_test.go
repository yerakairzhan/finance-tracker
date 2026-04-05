package handler

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSafeAuthResponse_NoRefreshOrSessionLeak(t *testing.T) {
	resp := safeAuthResponse("access-token", gin.H{"id": 1, "email": "a@b.com"})
	b, err := json.Marshal(resp)
	if err != nil {
		t.Fatal(err)
	}
	s := strings.ToLower(string(b))
	if strings.Contains(s, "refresh") || strings.Contains(s, "session") || strings.Contains(s, "jti") {
		t.Fatalf("sensitive fields leaked: %s", s)
	}
}

func TestReadRefreshTokenFromCookie_UsesCookieOnly(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// body token ignored; missing cookie must fail
	req := httptest.NewRequest("POST", "/api/v1/auth/refresh", strings.NewReader(`{"refresh_token":"bad"}`))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	if _, err := readRefreshTokenFromCookie(c); err == nil {
		t.Fatal("expected missing cookie error")
	}
}
