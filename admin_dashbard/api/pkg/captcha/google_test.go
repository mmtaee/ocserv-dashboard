package captcha

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGoogleCaptcha_Verify_Success(t *testing.T) {
	// Mock CAPTCHA server
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `{"success": true}`)
	}))
	defer mockServer.Close()

	// Inject mock HTTP client with overridden URL
	client := &http.Client{
		Transport: rewriteURLTransport(mockServer.URL),
	}
	verifier := NewGoogleVerifier()
	verifier.SetSecretKey("dummy-secret")
	verifier.Client = client

	verifier.Verify("dummy-token")
	result := verifier.IsValid()

	assert.True(t, result, "Expected CAPTCHA verification to succeed")
}

func TestGoogleCaptcha_Verify_Failure(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `{"success": false}`)
	}))
	defer mockServer.Close()

	client := &http.Client{
		Transport: rewriteURLTransport(mockServer.URL),
	}

	verifier := NewGoogleVerifier()

	verifier.SetSecretKey("dummy-secret")
	verifier.Client = client

	verifier.Verify("invalid-token")
	result := verifier.IsValid()
	assert.False(t, result, "Expected CAPTCHA verification to fail")
}

// Utility to intercept and rewrite external CAPTCHA request
func rewriteURLTransport(mockURL string) http.RoundTripper {
	return roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		req.URL.Scheme = "http"
		req.URL.Host = strings.TrimPrefix(mockURL, "http://")
		req.URL.Path = "/"
		return http.DefaultTransport.RoundTrip(req)
	})
}

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}
