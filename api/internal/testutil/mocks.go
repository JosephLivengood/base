package testutil

import (
	"context"

	"base/api/internal/domain/user"
	"base/api/internal/session"
)

// MockUserRepository is a mock implementation of user repository for testing.
type MockUserRepository struct {
	GetByIDFunc    func(ctx context.Context, id string) (*user.User, error)
	GetByEmailFunc func(ctx context.Context, email string) (*user.User, error)
	UpsertFunc     func(ctx context.Context, email, name, picture, googleID string) (*user.User, error)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id string) (*user.User, error) {
	if m.GetByIDFunc != nil {
		return m.GetByIDFunc(ctx, id)
	}
	return nil, user.ErrNotFound
}

func (m *MockUserRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	if m.GetByEmailFunc != nil {
		return m.GetByEmailFunc(ctx, email)
	}
	return nil, user.ErrNotFound
}

func (m *MockUserRepository) Upsert(ctx context.Context, email, name, picture, googleID string) (*user.User, error) {
	if m.UpsertFunc != nil {
		return m.UpsertFunc(ctx, email, name, picture, googleID)
	}
	return NewTestUser(), nil
}

// MockSessionStore is a mock implementation of session store for testing.
type MockSessionStore struct {
	CreateFunc       func(ctx context.Context, userID string) (*session.Session, error)
	GetFunc          func(ctx context.Context, sessionID string) (*session.Session, error)
	DeleteFunc       func(ctx context.Context, sessionID string) error
	RefreshFunc      func(ctx context.Context, sessionID string) (*session.Session, error)
	SetActiveOrgFunc func(ctx context.Context, sessionID, orgID string) error
}

func (m *MockSessionStore) Create(ctx context.Context, userID string) (*session.Session, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, userID)
	}
	return NewTestSession(userID), nil
}

func (m *MockSessionStore) Get(ctx context.Context, sessionID string) (*session.Session, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, sessionID)
	}
	return nil, session.ErrSessionNotFound
}

func (m *MockSessionStore) Delete(ctx context.Context, sessionID string) error {
	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, sessionID)
	}
	return nil
}

func (m *MockSessionStore) Refresh(ctx context.Context, sessionID string) (*session.Session, error) {
	if m.RefreshFunc != nil {
		return m.RefreshFunc(ctx, sessionID)
	}
	return nil, session.ErrSessionNotFound
}

func (m *MockSessionStore) SetActiveOrg(ctx context.Context, sessionID, orgID string) error {
	if m.SetActiveOrgFunc != nil {
		return m.SetActiveOrgFunc(ctx, sessionID, orgID)
	}
	return nil
}
