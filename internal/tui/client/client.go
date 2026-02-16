package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/Akram012388/niotebook-tui/internal/models"
)

// Client wraps all API calls to the niotebook server.
type Client struct {
	baseURL      string
	httpClient   *http.Client
	accessToken  string
	refreshToken string
	mu           sync.Mutex
	onRefresh    func(accessToken, refreshToken string)
}

// New creates a new API client pointing at the given base URL.
func New(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SetToken sets the access token for authenticated requests.
func (c *Client) SetToken(token string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.accessToken = token
}

// SetRefreshToken sets the refresh token used for automatic token renewal.
func (c *Client) SetRefreshToken(token string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.refreshToken = token
}

// OnTokenRefresh registers a callback invoked when tokens are refreshed.
func (c *Client) OnTokenRefresh(fn func(accessToken, refreshToken string)) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onRefresh = fn
}

// Login authenticates with email and password.
func (c *Client) Login(email, password string) (*models.AuthResponse, error) {
	body := models.LoginRequest{Email: email, Password: password}
	var resp models.AuthResponse
	if err := c.doJSON("POST", "/api/v1/auth/login", body, &resp, false); err != nil {
		return nil, err
	}
	if resp.Tokens != nil {
		c.SetToken(resp.Tokens.AccessToken)
		c.SetRefreshToken(resp.Tokens.RefreshToken)
	}
	return &resp, nil
}

// Register creates a new account.
func (c *Client) Register(username, email, password string) (*models.AuthResponse, error) {
	body := models.RegisterRequest{Username: username, Email: email, Password: password}
	var resp models.AuthResponse
	if err := c.doJSON("POST", "/api/v1/auth/register", body, &resp, false); err != nil {
		return nil, err
	}
	if resp.Tokens != nil {
		c.SetToken(resp.Tokens.AccessToken)
		c.SetRefreshToken(resp.Tokens.RefreshToken)
	}
	return &resp, nil
}

// Refresh exchanges a refresh token for new tokens.
func (c *Client) Refresh() (*models.TokenPair, error) {
	c.mu.Lock()
	rt := c.refreshToken
	c.mu.Unlock()

	body := models.RefreshRequest{RefreshToken: rt}
	var wrapper struct {
		Tokens models.TokenPair `json:"tokens"`
	}
	if err := c.doJSON("POST", "/api/v1/auth/refresh", body, &wrapper, false); err != nil {
		return nil, err
	}

	c.mu.Lock()
	c.accessToken = wrapper.Tokens.AccessToken
	c.refreshToken = wrapper.Tokens.RefreshToken
	cb := c.onRefresh
	c.mu.Unlock()

	if cb != nil {
		cb(wrapper.Tokens.AccessToken, wrapper.Tokens.RefreshToken)
	}
	return &wrapper.Tokens, nil
}

// GetTimeline fetches the global timeline with cursor-based pagination.
func (c *Client) GetTimeline(cursor string, limit int) (*models.TimelineResponse, error) {
	q := url.Values{}
	if cursor != "" {
		q.Set("cursor", cursor)
	}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	path := "/api/v1/timeline"
	if encoded := q.Encode(); encoded != "" {
		path += "?" + encoded
	}

	var resp models.TimelineResponse
	if err := c.doJSON("GET", path, nil, &resp, true); err != nil {
		return nil, err
	}
	return &resp, nil
}

// CreatePost publishes a new post with the given content.
func (c *Client) CreatePost(content string) (*models.Post, error) {
	body := struct {
		Content string `json:"content"`
	}{Content: content}

	var wrapper struct {
		Post models.Post `json:"post"`
	}
	if err := c.doJSON("POST", "/api/v1/posts", body, &wrapper, true); err != nil {
		return nil, err
	}
	return &wrapper.Post, nil
}

// GetPost retrieves a single post by ID.
func (c *Client) GetPost(id string) (*models.Post, error) {
	var wrapper struct {
		Post models.Post `json:"post"`
	}
	if err := c.doJSON("GET", "/api/v1/posts/"+id, nil, &wrapper, true); err != nil {
		return nil, err
	}
	return &wrapper.Post, nil
}

// GetUser retrieves a user by ID. Use "me" for the authenticated user.
func (c *Client) GetUser(id string) (*models.User, error) {
	var wrapper struct {
		User models.User `json:"user"`
	}
	if err := c.doJSON("GET", "/api/v1/users/"+id, nil, &wrapper, true); err != nil {
		return nil, err
	}
	return &wrapper.User, nil
}

// GetUserPosts retrieves posts by a specific user with cursor-based pagination.
func (c *Client) GetUserPosts(userID, cursor string, limit int) (*models.TimelineResponse, error) {
	q := url.Values{}
	if cursor != "" {
		q.Set("cursor", cursor)
	}
	if limit > 0 {
		q.Set("limit", strconv.Itoa(limit))
	}
	path := "/api/v1/users/" + userID + "/posts"
	if encoded := q.Encode(); encoded != "" {
		path += "?" + encoded
	}

	var resp models.TimelineResponse
	if err := c.doJSON("GET", path, nil, &resp, true); err != nil {
		return nil, err
	}
	return &resp, nil
}

// UpdateUser updates the authenticated user's profile.
func (c *Client) UpdateUser(updates *models.UserUpdate) (*models.User, error) {
	var wrapper struct {
		User models.User `json:"user"`
	}
	if err := c.doJSON("PUT", "/api/v1/users/me", updates, &wrapper, true); err != nil {
		return nil, err
	}
	return &wrapper.User, nil
}

// doJSON performs an HTTP request, optionally with auth, and decodes the JSON response.
// If withAuth is true and a 401 with token_expired is received, it attempts a refresh and retries.
func (c *Client) doJSON(method, path string, body interface{}, dst interface{}, withAuth bool) error {
	resp, err := c.do(method, path, body, withAuth)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Handle 401 with transparent refresh
	if resp.StatusCode == http.StatusUnauthorized && withAuth {
		resp.Body.Close()
		if _, err := c.Refresh(); err != nil {
			return fmt.Errorf("token refresh failed: %w", err)
		}
		resp, err = c.do(method, path, body, withAuth)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
	}

	if resp.StatusCode >= 400 {
		var errResp struct {
			Error models.APIError `json:"error"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
			return fmt.Errorf("api error (status %d)", resp.StatusCode)
		}
		return &errResp.Error
	}

	if dst != nil {
		if err := json.NewDecoder(resp.Body).Decode(dst); err != nil {
			return fmt.Errorf("decoding response: %w", err)
		}
	}
	return nil
}

// do performs the raw HTTP request with retry on network errors.
func (c *Client) do(method, path string, body interface{}, withAuth bool) (*http.Response, error) {
	const maxRetries = 3

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(1<<(attempt-1)) * time.Second)
		}

		resp, err := c.doOnce(method, path, body, withAuth)
		if err == nil {
			return resp, nil
		}

		// Only retry on network errors, not on successful HTTP responses
		var netErr net.Error
		if errors.As(err, &netErr) || errors.Is(err, net.ErrClosed) {
			lastErr = err
			continue
		}
		// Non-network error — don't retry
		return nil, err
	}
	return nil, fmt.Errorf("request failed after %d attempts: %w", maxRetries, lastErr)
}

// doOnce performs a single HTTP request attempt.
func (c *Client) doOnce(method, path string, body interface{}, withAuth bool) (*http.Response, error) {
	var reqBody *bytes.Buffer
	if body != nil {
		reqBody = &bytes.Buffer{}
		if err := json.NewEncoder(reqBody).Encode(body); err != nil {
			return nil, fmt.Errorf("encoding request: %w", err)
		}
	}

	var req *http.Request
	var err error
	if reqBody != nil {
		req, err = http.NewRequest(method, c.baseURL+path, reqBody)
	} else {
		req, err = http.NewRequest(method, c.baseURL+path, nil)
	}
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if withAuth {
		c.mu.Lock()
		token := c.accessToken
		c.mu.Unlock()
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
	}

	return c.httpClient.Do(req)
}
