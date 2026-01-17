package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"base/api/internal/database"
	"base/api/internal/domain/user"
	"base/api/internal/session"
	"base/api/pkg/response"
)

// setupTestDependencies creates mock dependencies for testing
func setupTestDependencies(t *testing.T) (*session.Store, *user.Repository, redismock.ClientMock) {
	client, mock := redismock.NewClientMock()
	redisDB := &database.RedisDB{Client: client}
	sessionStore := session.NewStore(redisDB, "test-secret")

	// We can't easily mock the user repository since it uses a real postgres connection
	// For now, we'll test what we can and note that full integration tests would be needed
	return sessionStore, nil, mock
}

func TestRequireAuth_NoCookie(t *testing.T) {
	sessionStore, _, _ := setupTestDependencies(t)

	handler := RequireAuth(sessionStore, nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called when no cookie")
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp response.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "unauthorized", resp.Error)
	assert.Contains(t, resp.Message, "authentication required")
}

func TestRequireAuth_InvalidSession(t *testing.T) {
	sessionStore, _, mock := setupTestDependencies(t)

	mock.ExpectGet("session:invalid-session-id").RedisNil()

	handler := RequireAuth(sessionStore, nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called with invalid session")
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{
		Name:  "session_id",
		Value: "invalid-session-id",
	})
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var resp response.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	assert.Equal(t, "unauthorized", resp.Error)
	assert.Contains(t, resp.Message, "invalid or expired session")

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserFromContext_WithUser(t *testing.T) {
	testUser := &user.User{
		ID:        "user-123",
		Email:     "test@example.com",
		Name:      "Test User",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	ctx := context.WithValue(context.Background(), UserContextKey, testUser)
	result := GetUserFromContext(ctx)

	require.NotNil(t, result)
	assert.Equal(t, testUser.ID, result.ID)
	assert.Equal(t, testUser.Email, result.Email)
	assert.Equal(t, testUser.Name, result.Name)
}

func TestGetUserFromContext_NoUser(t *testing.T) {
	ctx := context.Background()
	result := GetUserFromContext(ctx)

	assert.Nil(t, result)
}

func TestGetUserFromContext_WrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), UserContextKey, "not a user")
	result := GetUserFromContext(ctx)

	assert.Nil(t, result)
}

func TestGetSessionFromContext_WithSession(t *testing.T) {
	testSession := &session.Session{
		ID:          "session-123",
		UserID:      "user-123",
		ActiveOrgID: "org-456",
		CreatedAt:   time.Now(),
		ExpiresAt:   time.Now().Add(24 * time.Hour),
	}

	ctx := context.WithValue(context.Background(), SessionContextKey, testSession)
	result := GetSessionFromContext(ctx)

	require.NotNil(t, result)
	assert.Equal(t, testSession.ID, result.ID)
	assert.Equal(t, testSession.UserID, result.UserID)
	assert.Equal(t, testSession.ActiveOrgID, result.ActiveOrgID)
}

func TestGetSessionFromContext_NoSession(t *testing.T) {
	ctx := context.Background()
	result := GetSessionFromContext(ctx)

	assert.Nil(t, result)
}

func TestGetSessionFromContext_WrongType(t *testing.T) {
	ctx := context.WithValue(context.Background(), SessionContextKey, "not a session")
	result := GetSessionFromContext(ctx)

	assert.Nil(t, result)
}

func TestContextKeys_AreDistinct(t *testing.T) {
	// Ensure context keys are distinct
	assert.NotEqual(t, UserContextKey, SessionContextKey)
}
