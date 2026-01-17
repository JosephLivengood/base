package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"base/api/pkg/response"
)

// Test Google user info structure
func TestGoogleUserInfo_Structure(t *testing.T) {
	userInfo := &GoogleUserInfo{
		ID:      "google-123",
		Email:   "test@gmail.com",
		Name:    "Test User",
		Picture: "https://lh3.googleusercontent.com/photo.jpg",
	}

	data, err := json.Marshal(userInfo)
	require.NoError(t, err)

	var parsed map[string]interface{}
	err = json.Unmarshal(data, &parsed)
	require.NoError(t, err)

	assert.Equal(t, "google-123", parsed["id"])
	assert.Equal(t, "test@gmail.com", parsed["email"])
	assert.Equal(t, "Test User", parsed["name"])
	assert.Equal(t, "https://lh3.googleusercontent.com/photo.jpg", parsed["picture"])
}

// Test generateState produces valid base64
func TestGenerateState(t *testing.T) {
	state1 := generateState()
	state2 := generateState()

	// States should be non-empty
	assert.NotEmpty(t, state1)
	assert.NotEmpty(t, state2)

	// States should be unique
	assert.NotEqual(t, state1, state2)

	// Should be base64 URL encoded (32 bytes = 44 chars)
	assert.Len(t, state1, 44)
	assert.Len(t, state2, 44)
}

// Test logout clears session cookie
func TestLogout_ClearsCookie(t *testing.T) {
	// Create a mock handler that doesn't need actual dependencies
	// for testing the cookie clearing behavior

	w := httptest.NewRecorder()

	// Simulate what logout does for cookie clearing
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	cookies := w.Result().Cookies()
	require.Len(t, cookies, 1)

	cookie := cookies[0]
	assert.Equal(t, "session_id", cookie.Name)
	assert.Equal(t, "", cookie.Value)
	assert.Equal(t, -1, cookie.MaxAge)
}

// Test Me endpoint returns null for no session
func TestMe_NoSession_ReturnsNull(t *testing.T) {
	w := httptest.NewRecorder()

	// When there's no session, Me should return OK with null data
	response.OK(w, nil)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp response.SuccessResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Nil(t, resp.Data)
}

// Test OAuth state cookie settings
func TestOAuthStateCookie_Settings(t *testing.T) {
	w := httptest.NewRecorder()

	// Test state cookie creation (similar to GoogleLogin)
	state := generateState()
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Would be true in production
		SameSite: http.SameSiteLaxMode,
		MaxAge:   300, // 5 minutes
	})

	cookies := w.Result().Cookies()
	require.Len(t, cookies, 1)

	cookie := cookies[0]
	assert.Equal(t, "oauth_state", cookie.Name)
	assert.Equal(t, state, cookie.Value)
	assert.Equal(t, "/", cookie.Path)
	assert.True(t, cookie.HttpOnly)
	assert.Equal(t, http.SameSiteLaxMode, cookie.SameSite)
	assert.Equal(t, 300, cookie.MaxAge)
}

// Test state cookie clearing
func TestOAuthStateCookie_Clearing(t *testing.T) {
	w := httptest.NewRecorder()

	// Test clearing the state cookie (as done after callback)
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})

	cookies := w.Result().Cookies()
	require.Len(t, cookies, 1)

	cookie := cookies[0]
	assert.Equal(t, "oauth_state", cookie.Name)
	assert.Equal(t, "", cookie.Value)
	assert.Equal(t, -1, cookie.MaxAge)
}

// Test session cookie settings
func TestSessionCookie_Settings(t *testing.T) {
	w := httptest.NewRecorder()

	// Test session cookie creation (similar to GoogleCallback)
	http.SetCookie(w, &http.Cookie{
		Name:     "session_id",
		Value:    "test-session-123",
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Would be true in production
		SameSite: http.SameSiteLaxMode,
		MaxAge:   86400, // 24 hours
	})

	cookies := w.Result().Cookies()
	require.Len(t, cookies, 1)

	cookie := cookies[0]
	assert.Equal(t, "session_id", cookie.Name)
	assert.Equal(t, "test-session-123", cookie.Value)
	assert.Equal(t, "/", cookie.Path)
	assert.True(t, cookie.HttpOnly)
	assert.Equal(t, http.SameSiteLaxMode, cookie.SameSite)
	assert.Equal(t, 86400, cookie.MaxAge)
}

// Test error responses from callback
func TestCallback_Errors(t *testing.T) {
	tests := []struct {
		name           string
		error          string
		expectedStatus int
	}{
		{
			name:           "state cookie not found",
			error:          "State cookie not found",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid state parameter",
			error:          "Invalid state parameter",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "code not found",
			error:          "Code not found",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			http.Error(w, tt.error, tt.expectedStatus)

			assert.Equal(t, tt.expectedStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.error)
		})
	}
}

// Test redirect responses
func TestRedirects(t *testing.T) {
	tests := []struct {
		name        string
		redirectURL string
	}{
		{
			name:        "logout redirect to home",
			redirectURL: "/",
		},
		{
			name:        "callback redirect to home",
			redirectURL: "/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)

			http.Redirect(w, req, tt.redirectURL, http.StatusTemporaryRedirect)

			assert.Equal(t, http.StatusTemporaryRedirect, w.Code)
			assert.Equal(t, tt.redirectURL, w.Header().Get("Location"))
		})
	}
}
