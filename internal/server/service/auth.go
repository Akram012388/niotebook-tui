package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/Akram012388/niotebook-tui/internal/models"
	"github.com/Akram012388/niotebook-tui/internal/server/store"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	users      store.UserStore
	tokens     store.RefreshTokenStore
	jwtSecret  []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewAuthService(users store.UserStore, tokens store.RefreshTokenStore, jwtSecret string) *AuthService {
	return &AuthService{
		users:      users,
		tokens:     tokens,
		jwtSecret:  []byte(jwtSecret),
		accessTTL:  24 * time.Hour,
		refreshTTL: 7 * 24 * time.Hour,
	}
}

func (s *AuthService) Register(ctx context.Context, req *models.RegisterRequest) (*models.AuthResponse, error) {
	if err := ValidateUsername(req.Username); err != nil {
		return nil, err
	}
	if err := ValidateEmail(req.Email); err != nil {
		return nil, err
	}
	if err := ValidatePassword(req.Password); err != nil {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user, err := s.users.CreateUser(ctx, req.Username, req.Email, string(hash), req.Username)
	if err != nil {
		return nil, err
	}

	tokens, err := s.generateTokenPair(ctx, user)
	if err != nil {
		return nil, err
	}

	return &models.AuthResponse{User: user, Tokens: tokens}, nil
}

func (s *AuthService) Login(ctx context.Context, req *models.LoginRequest) (*models.AuthResponse, error) {
	user, hash, err := s.users.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
		return nil, &models.APIError{Code: models.ErrCodeUnauthorized, Message: "invalid email or password"}
	}

	tokens, err := s.generateTokenPair(ctx, user)
	if err != nil {
		return nil, err
	}

	return &models.AuthResponse{User: user, Tokens: tokens}, nil
}

func (s *AuthService) Refresh(ctx context.Context, rawToken string) (*models.TokenPair, error) {
	tokenHash := hashRefreshToken(rawToken)

	id, userID, expiresAt, err := s.tokens.GetByHash(ctx, tokenHash)
	if err != nil {
		return nil, &models.APIError{Code: models.ErrCodeTokenExpired, Message: "refresh token has expired"}
	}
	_ = id

	if time.Now().After(expiresAt) {
		s.tokens.DeleteByHash(ctx, tokenHash)
		return nil, &models.APIError{Code: models.ErrCodeTokenExpired, Message: "refresh token has expired"}
	}

	// Single-use: delete the consumed token
	if err := s.tokens.DeleteByHash(ctx, tokenHash); err != nil {
		return nil, fmt.Errorf("delete consumed token: %w", err)
	}

	user, err := s.users.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return s.generateTokenPair(ctx, user)
}

func (s *AuthService) generateTokenPair(ctx context.Context, user *models.User) (*models.TokenPair, error) {
	now := time.Now()
	expiresAt := now.Add(s.accessTTL)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":      user.ID,
		"username": user.Username,
		"iat":      now.Unix(),
		"exp":      expiresAt.Unix(),
	})

	accessToken, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("sign access token: %w", err)
	}

	rawRefresh, err := generateRefreshToken()
	if err != nil {
		return nil, err
	}

	refreshHash := hashRefreshToken(rawRefresh)
	refreshExpiry := now.Add(s.refreshTTL)

	if err := s.tokens.StoreToken(ctx, user.ID, refreshHash, refreshExpiry); err != nil {
		return nil, fmt.Errorf("store refresh token: %w", err)
	}

	return &models.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
		ExpiresAt:    expiresAt,
	}, nil
}

func generateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate refresh token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func hashRefreshToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", h)
}
