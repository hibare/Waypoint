package handlers

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClientIP_CloudflareHeader(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Cf-Connecting-Ip", "203.0.113.1")
	r.RemoteAddr = "10.42.2.50:57056"

	assert.Equal(t, "203.0.113.1", clientIP(r))
}

func TestClientIP_XForwardedFor(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-Forwarded-For", "198.51.100.2")
	r.RemoteAddr = "10.42.2.50:57056"

	assert.Equal(t, "198.51.100.2", clientIP(r))
}

func TestClientIP_PrefersCloudflareOverXForwardedFor(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("Cf-Connecting-Ip", "203.0.113.1")
	r.Header.Set("X-Forwarded-For", "198.51.100.2")
	r.RemoteAddr = "10.42.2.50:57056"

	assert.Equal(t, "203.0.113.1", clientIP(r))
}

func TestClientIP_FallsBackToRemoteAddr(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "203.0.113.1:8080"

	assert.Equal(t, "203.0.113.1", clientIP(r))
}

func TestClientIP_RemoteAddrNoPort(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r.RemoteAddr = "203.0.113.1"

	assert.Equal(t, "203.0.113.1", clientIP(r))
}

func TestClientIP_XForwardedForMultiple(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)
	r.Header.Set("X-Forwarded-For", "203.0.113.1, 198.51.100.2, 10.0.0.1")
	r.RemoteAddr = "10.42.2.50:57056"

	assert.Equal(t, "203.0.113.1", clientIP(r))
}

func TestClientIP_NoHeaders(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/", nil)

	assert.Equal(t, "", clientIP(r))
}
