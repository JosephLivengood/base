package session

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"

	"base/api/internal/database"
)

const (
	sessionPrefix = "session:"
	sessionTTL    = 24 * time.Hour
)

var ErrSessionNotFound = errors.New("session not found")

type Session struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	ActiveOrgID string    `json:"active_org_id,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	ExpiresAt   time.Time `json:"expires_at"`
}

type Store struct {
	redis  *database.RedisDB
	secret string
}

func NewStore(redis *database.RedisDB, secret string) *Store {
	return &Store{
		redis:  redis,
		secret: secret,
	}
}

func (s *Store) Create(ctx context.Context, userID string) (*Session, error) {
	sessionID, err := generateSessionID()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	session := &Session{
		ID:        sessionID,
		UserID:    userID,
		CreatedAt: now,
		ExpiresAt: now.Add(sessionTTL),
	}

	data, err := json.Marshal(session)
	if err != nil {
		return nil, err
	}

	key := sessionPrefix + sessionID
	if err := s.redis.Client.Set(ctx, key, data, sessionTTL).Err(); err != nil {
		return nil, err
	}

	return session, nil
}

func (s *Store) Get(ctx context.Context, sessionID string) (*Session, error) {
	key := sessionPrefix + sessionID
	data, err := s.redis.Client.Get(ctx, key).Bytes()
	if errors.Is(err, redis.Nil) {
		return nil, ErrSessionNotFound
	}
	if err != nil {
		return nil, err
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, err
	}

	if time.Now().After(session.ExpiresAt) {
		s.Delete(ctx, sessionID)
		return nil, ErrSessionNotFound
	}

	return &session, nil
}

func (s *Store) Delete(ctx context.Context, sessionID string) error {
	key := sessionPrefix + sessionID
	return s.redis.Client.Del(ctx, key).Err()
}

func (s *Store) Refresh(ctx context.Context, sessionID string) (*Session, error) {
	session, err := s.Get(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	session.ExpiresAt = time.Now().Add(sessionTTL)

	data, err := json.Marshal(session)
	if err != nil {
		return nil, err
	}

	key := sessionPrefix + sessionID
	if err := s.redis.Client.Set(ctx, key, data, sessionTTL).Err(); err != nil {
		return nil, err
	}

	return session, nil
}

func (s *Store) SetActiveOrg(ctx context.Context, sessionID, orgID string) error {
	session, err := s.Get(ctx, sessionID)
	if err != nil {
		return err
	}

	session.ActiveOrgID = orgID

	data, err := json.Marshal(session)
	if err != nil {
		return err
	}

	key := sessionPrefix + sessionID
	ttl := time.Until(session.ExpiresAt)
	if ttl <= 0 {
		ttl = sessionTTL
	}
	return s.redis.Client.Set(ctx, key, data, ttl).Err()
}

func generateSessionID() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
