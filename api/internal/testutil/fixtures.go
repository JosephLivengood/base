package testutil

import (
	"time"

	"base/api/internal/domain/user"
	"base/api/internal/session"
)

// NewTestUser creates a test user with sensible defaults.
// Optionally override fields using the returned pointer.
func NewTestUser() *user.User {
	return &user.User{
		ID:        "user-123",
		Email:     "test@example.com",
		Name:      "Test User",
		Picture:   "https://example.com/picture.jpg",
		GoogleID:  "google-123",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

// NewTestSession creates a test session.
func NewTestSession(userID string) *session.Session {
	return &session.Session{
		ID:        "session-123",
		UserID:    userID,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}
}

// NewTestSessionWithOrg creates a test session with an active organization.
func NewTestSessionWithOrg(userID, orgID string) *session.Session {
	sess := NewTestSession(userID)
	sess.ActiveOrgID = orgID
	return sess
}
