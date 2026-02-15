package service_test

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/Akram012388/niotebook-tui/internal/models"
)

// mockUserStore implements store.UserStore with in-memory maps
type mockUserStore struct {
	mu       sync.Mutex
	users    map[string]*models.User
	emails   map[string]string // email -> user ID
	hashes   map[string]string // user ID -> password hash
	nextID   int
}

func newMockUserStore() *mockUserStore {
	return &mockUserStore{
		users:  make(map[string]*models.User),
		emails: make(map[string]string),
		hashes: make(map[string]string),
	}
}

func (m *mockUserStore) CreateUser(_ context.Context, username, email, passwordHash, displayName string) (*models.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.emails[email]; exists {
		return nil, &models.APIError{Code: models.ErrCodeConflict, Message: "email already exists"}
	}
	for _, u := range m.users {
		if u.Username == username {
			return nil, &models.APIError{Code: models.ErrCodeConflict, Message: "username already exists"}
		}
	}

	m.nextID++
	id := fmt.Sprintf("user-%d", m.nextID)
	user := &models.User{
		ID:          id,
		Username:    username,
		DisplayName: displayName,
		CreatedAt:   time.Now(),
	}
	m.users[id] = user
	m.emails[email] = id
	m.hashes[id] = passwordHash
	return user, nil
}

func (m *mockUserStore) GetUserByEmail(_ context.Context, email string) (*models.User, string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	id, exists := m.emails[email]
	if !exists {
		return nil, "", &models.APIError{Code: models.ErrCodeNotFound, Message: "user not found"}
	}
	return m.users[id], m.hashes[id], nil
}

func (m *mockUserStore) GetUserByID(_ context.Context, id string) (*models.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	user, exists := m.users[id]
	if !exists {
		return nil, &models.APIError{Code: models.ErrCodeNotFound, Message: "user not found"}
	}
	return user, nil
}

func (m *mockUserStore) GetUserByUsername(_ context.Context, username string) (*models.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, u := range m.users {
		if u.Username == username {
			return u, nil
		}
	}
	return nil, &models.APIError{Code: models.ErrCodeNotFound, Message: "user not found"}
}

func (m *mockUserStore) UpdateUser(_ context.Context, id string, updates *models.UserUpdate) (*models.User, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	user, exists := m.users[id]
	if !exists {
		return nil, &models.APIError{Code: models.ErrCodeNotFound, Message: "user not found"}
	}
	if updates.DisplayName != nil {
		user.DisplayName = *updates.DisplayName
	}
	if updates.Bio != nil {
		user.Bio = *updates.Bio
	}
	return user, nil
}

// mockRefreshTokenStore implements store.RefreshTokenStore with in-memory maps
type mockRefreshTokenStore struct {
	mu     sync.Mutex
	tokens map[string]refreshTokenEntry // hash -> entry
	nextID int
}

type refreshTokenEntry struct {
	id        string
	userID    string
	expiresAt time.Time
}

func newMockRefreshTokenStore() *mockRefreshTokenStore {
	return &mockRefreshTokenStore{
		tokens: make(map[string]refreshTokenEntry),
	}
}

func (m *mockRefreshTokenStore) StoreToken(_ context.Context, userID, tokenHash string, expiresAt time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.nextID++
	m.tokens[tokenHash] = refreshTokenEntry{
		id:        fmt.Sprintf("token-%d", m.nextID),
		userID:    userID,
		expiresAt: expiresAt,
	}
	return nil
}

func (m *mockRefreshTokenStore) GetByHash(_ context.Context, tokenHash string) (id, userID string, expiresAt time.Time, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	entry, exists := m.tokens[tokenHash]
	if !exists {
		return "", "", time.Time{}, &models.APIError{Code: models.ErrCodeTokenExpired, Message: "token not found"}
	}
	return entry.id, entry.userID, entry.expiresAt, nil
}

func (m *mockRefreshTokenStore) DeleteByHash(_ context.Context, tokenHash string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.tokens, tokenHash)
	return nil
}

func (m *mockRefreshTokenStore) DeleteAllForUser(_ context.Context, userID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	for hash, entry := range m.tokens {
		if entry.userID == userID {
			delete(m.tokens, hash)
		}
	}
	return nil
}

func (m *mockRefreshTokenStore) DeleteExpired(_ context.Context) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var count int64
	now := time.Now()
	for hash, entry := range m.tokens {
		if now.After(entry.expiresAt) {
			delete(m.tokens, hash)
			count++
		}
	}
	return count, nil
}
