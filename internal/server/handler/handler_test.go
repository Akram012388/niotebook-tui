package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/Akram012388/niotebook-tui/internal/models"
	"github.com/Akram012388/niotebook-tui/internal/server/handler"
	"github.com/Akram012388/niotebook-tui/internal/server/middleware"
	"github.com/Akram012388/niotebook-tui/internal/server/service"
	"github.com/Akram012388/niotebook-tui/internal/server/store"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

const testJWTSecret = "test-secret-32-bytes-long-xxxxx"

func testDBURL() string {
	if url := os.Getenv("NIOTEBOOK_TEST_DB_URL"); url != "" {
		return url
	}
	return "postgres://localhost/niotebook_test?sslmode=disable"
}

func setupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dbURL := testDBURL()

	m, err := migrate.New("file://../../../migrations", dbURL)
	if err != nil {
		t.Fatalf("migrate new: %v", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("migrate up: %v", err)
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		t.Fatalf("pool: %v", err)
	}

	// Clean before test
	_, _ = pool.Exec(context.Background(),
		"TRUNCATE users, posts, refresh_tokens CASCADE")

	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(),
			"TRUNCATE users, posts, refresh_tokens CASCADE")
		pool.Close()
	})

	return pool
}

type testServer struct {
	mux     *http.ServeMux
	handler http.Handler
	authSvc *service.AuthService
	postSvc *service.PostService
	userSvc *service.UserService
}

func setupTestServer(t *testing.T) *testServer {
	t.Helper()

	pool := setupTestDB(t)

	userStore := store.NewUserStore(pool)
	postStore := store.NewPostStore(pool)
	tokenStore := store.NewRefreshTokenStore(pool)

	authSvc := service.NewAuthService(userStore, tokenStore, testJWTSecret)
	postSvc := service.NewPostService(postStore)
	userSvc := service.NewUserService(userStore)

	mux := http.NewServeMux()

	// Auth routes
	mux.HandleFunc("POST /api/v1/auth/register", handler.HandleRegister(authSvc))
	mux.HandleFunc("POST /api/v1/auth/login", handler.HandleLogin(authSvc))
	mux.HandleFunc("POST /api/v1/auth/refresh", handler.HandleRefresh(authSvc))

	// Post routes
	mux.HandleFunc("POST /api/v1/posts", handler.HandleCreatePost(postSvc))
	mux.HandleFunc("GET /api/v1/posts/{id}", handler.HandleGetPost(postSvc))

	// Timeline
	mux.HandleFunc("GET /api/v1/timeline", handler.HandleTimeline(postSvc))

	// User routes
	mux.HandleFunc("GET /api/v1/users/{id}", handler.HandleGetUser(userSvc))
	mux.HandleFunc("GET /api/v1/users/{id}/posts", handler.HandleGetUserPosts(postSvc))
	mux.HandleFunc("PATCH /api/v1/users/me", handler.HandleUpdateUser(userSvc))

	// Health
	mux.HandleFunc("GET /health", handler.HandleHealth(pool))

	// Apply auth middleware
	h := middleware.Auth(testJWTSecret)(mux)

	return &testServer{
		mux:     mux,
		handler: h,
		authSvc: authSvc,
		postSvc: postSvc,
		userSvc: userSvc,
	}
}

func (ts *testServer) do(method, path string, body interface{}, token string) *httptest.ResponseRecorder {
	var reqBody *bytes.Buffer
	if body != nil {
		b, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(b)
	} else {
		reqBody = &bytes.Buffer{}
	}

	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	rec := httptest.NewRecorder()
	ts.handler.ServeHTTP(rec, req)
	return rec
}

func parseJSON(t *testing.T, rec *httptest.ResponseRecorder, v interface{}) {
	t.Helper()
	if err := json.NewDecoder(rec.Body).Decode(v); err != nil {
		t.Fatalf("decode response: %v\nbody: %s", err, rec.Body.String())
	}
}

// --- Integration Tests ---

func TestFullFlow(t *testing.T) {
	ts := setupTestServer(t)

	// 1. Register
	rec := ts.do("POST", "/api/v1/auth/register", models.RegisterRequest{
		Username: "testuser",
		Email:    "test@example.com",
		Password: "securepass123",
	}, "")

	if rec.Code != http.StatusCreated {
		t.Fatalf("register: status = %d, want %d\nbody: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	var authResp models.AuthResponse
	parseJSON(t, rec, &authResp)
	if authResp.User.Username != "testuser" {
		t.Errorf("register: username = %q, want %q", authResp.User.Username, "testuser")
	}
	if authResp.Tokens.AccessToken == "" {
		t.Error("register: access token is empty")
	}

	refreshToken := authResp.Tokens.RefreshToken
	userID := authResp.User.ID

	// 2. Login
	rec = ts.do("POST", "/api/v1/auth/login", models.LoginRequest{
		Email:    "test@example.com",
		Password: "securepass123",
	}, "")

	if rec.Code != http.StatusOK {
		t.Fatalf("login: status = %d, want %d\nbody: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var loginResp models.AuthResponse
	parseJSON(t, rec, &loginResp)
	if loginResp.User.Username != "testuser" {
		t.Errorf("login: username = %q, want %q", loginResp.User.Username, "testuser")
	}
	// Use tokens from login
	accessToken := loginResp.Tokens.AccessToken

	// 3. Create Post
	rec = ts.do("POST", "/api/v1/posts", map[string]string{
		"content": "Hello from tests!",
	}, accessToken)

	if rec.Code != http.StatusCreated {
		t.Fatalf("create post: status = %d, want %d\nbody: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	var postResp map[string]models.Post
	parseJSON(t, rec, &postResp)
	post := postResp["post"]
	if post.Content != "Hello from tests!" {
		t.Errorf("create post: content = %q, want %q", post.Content, "Hello from tests!")
	}
	if post.AuthorID != userID {
		t.Errorf("create post: author_id = %q, want %q", post.AuthorID, userID)
	}
	postID := post.ID

	// 4. Get Post
	rec = ts.do("GET", "/api/v1/posts/"+postID, nil, accessToken)

	if rec.Code != http.StatusOK {
		t.Fatalf("get post: status = %d, want %d\nbody: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var getPostResp map[string]models.Post
	parseJSON(t, rec, &getPostResp)
	if getPostResp["post"].Content != "Hello from tests!" {
		t.Errorf("get post: content = %q, want %q", getPostResp["post"].Content, "Hello from tests!")
	}

	// 5. Get Timeline
	rec = ts.do("GET", "/api/v1/timeline?limit=10", nil, accessToken)

	if rec.Code != http.StatusOK {
		t.Fatalf("timeline: status = %d, want %d\nbody: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var timelineResp models.TimelineResponse
	parseJSON(t, rec, &timelineResp)
	if len(timelineResp.Posts) != 1 {
		t.Errorf("timeline: got %d posts, want 1", len(timelineResp.Posts))
	}

	// 6. Get Profile (me)
	rec = ts.do("GET", "/api/v1/users/me", nil, accessToken)

	if rec.Code != http.StatusOK {
		t.Fatalf("get user me: status = %d, want %d\nbody: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var userResp map[string]models.User
	parseJSON(t, rec, &userResp)
	if userResp["user"].Username != "testuser" {
		t.Errorf("get user me: username = %q, want %q", userResp["user"].Username, "testuser")
	}

	// 7. Get Profile by ID
	rec = ts.do("GET", "/api/v1/users/"+userID, nil, accessToken)

	if rec.Code != http.StatusOK {
		t.Fatalf("get user by id: status = %d, want %d\nbody: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	// 8. Update Profile
	newName := "Test User Updated"
	newBio := "A test bio"
	rec = ts.do("PATCH", "/api/v1/users/me", models.UserUpdate{
		DisplayName: &newName,
		Bio:         &newBio,
	}, accessToken)

	if rec.Code != http.StatusOK {
		t.Fatalf("update user: status = %d, want %d\nbody: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var updatedUserResp map[string]models.User
	parseJSON(t, rec, &updatedUserResp)
	if updatedUserResp["user"].DisplayName != newName {
		t.Errorf("update user: display_name = %q, want %q", updatedUserResp["user"].DisplayName, newName)
	}
	if updatedUserResp["user"].Bio != newBio {
		t.Errorf("update user: bio = %q, want %q", updatedUserResp["user"].Bio, newBio)
	}

	// 9. Get User Posts
	rec = ts.do("GET", "/api/v1/users/"+userID+"/posts?limit=10", nil, accessToken)

	if rec.Code != http.StatusOK {
		t.Fatalf("user posts: status = %d, want %d\nbody: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var userPostsResp models.TimelineResponse
	parseJSON(t, rec, &userPostsResp)
	if len(userPostsResp.Posts) != 1 {
		t.Errorf("user posts: got %d posts, want 1", len(userPostsResp.Posts))
	}

	// 10. Refresh Token
	rec = ts.do("POST", "/api/v1/auth/refresh", models.RefreshRequest{
		RefreshToken: refreshToken,
	}, "")

	if rec.Code != http.StatusOK {
		t.Fatalf("refresh: status = %d, want %d\nbody: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var refreshResp map[string]models.TokenPair
	parseJSON(t, rec, &refreshResp)
	if refreshResp["tokens"].AccessToken == "" {
		t.Error("refresh: new access token is empty")
	}
}

func TestRegisterDuplicateUsername(t *testing.T) {
	ts := setupTestServer(t)

	// Register first user
	rec := ts.do("POST", "/api/v1/auth/register", models.RegisterRequest{
		Username: "dupuser",
		Email:    "dup1@example.com",
		Password: "securepass123",
	}, "")

	if rec.Code != http.StatusCreated {
		t.Fatalf("first register: status = %d, want %d\nbody: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	// Try to register with same username
	rec = ts.do("POST", "/api/v1/auth/register", models.RegisterRequest{
		Username: "dupuser",
		Email:    "dup2@example.com",
		Password: "securepass123",
	}, "")

	if rec.Code != http.StatusConflict {
		t.Fatalf("dup register: status = %d, want %d\nbody: %s", rec.Code, http.StatusConflict, rec.Body.String())
	}

	var errResp map[string]models.APIError
	parseJSON(t, rec, &errResp)
	if errResp["error"].Code != models.ErrCodeConflict {
		t.Errorf("dup register: error code = %q, want %q", errResp["error"].Code, models.ErrCodeConflict)
	}
}

func TestLoginInvalidPassword(t *testing.T) {
	ts := setupTestServer(t)

	// Register user
	ts.do("POST", "/api/v1/auth/register", models.RegisterRequest{
		Username: "pwuser",
		Email:    "pw@example.com",
		Password: "securepass123",
	}, "")

	// Login with wrong password
	rec := ts.do("POST", "/api/v1/auth/login", models.LoginRequest{
		Email:    "pw@example.com",
		Password: "wrongpassword",
	}, "")

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("bad login: status = %d, want %d\nbody: %s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
}

func TestRefreshExpiredToken(t *testing.T) {
	ts := setupTestServer(t)

	// Try to refresh with an invalid token
	rec := ts.do("POST", "/api/v1/auth/refresh", models.RefreshRequest{
		RefreshToken: "totally-invalid-token",
	}, "")

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("bad refresh: status = %d, want %d\nbody: %s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
}

func TestCreatePostTooLong(t *testing.T) {
	ts := setupTestServer(t)

	// Register and get token
	rec := ts.do("POST", "/api/v1/auth/register", models.RegisterRequest{
		Username: "longpost",
		Email:    "long@example.com",
		Password: "securepass123",
	}, "")

	var authResp models.AuthResponse
	parseJSON(t, rec, &authResp)
	token := authResp.Tokens.AccessToken

	// Try to create a post that's too long (>140 chars)
	longContent := strings.Repeat("a", 141)
	rec = ts.do("POST", "/api/v1/posts", map[string]string{
		"content": longContent,
	}, token)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("long post: status = %d, want %d\nbody: %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestGetPostNotFound(t *testing.T) {
	ts := setupTestServer(t)

	// Register and get token
	rec := ts.do("POST", "/api/v1/auth/register", models.RegisterRequest{
		Username: "notfound",
		Email:    "notfound@example.com",
		Password: "securepass123",
	}, "")

	var authResp models.AuthResponse
	parseJSON(t, rec, &authResp)
	token := authResp.Tokens.AccessToken

	// Try to get a post that doesn't exist
	rec = ts.do("GET", "/api/v1/posts/00000000-0000-0000-0000-000000000000", nil, token)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("not found: status = %d, want %d\nbody: %s", rec.Code, http.StatusNotFound, rec.Body.String())
	}
}

func TestUnauthorizedAccess(t *testing.T) {
	ts := setupTestServer(t)

	// Try to access protected endpoint without token
	rec := ts.do("GET", "/api/v1/timeline", nil, "")

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("unauth: status = %d, want %d\nbody: %s", rec.Code, http.StatusUnauthorized, rec.Body.String())
	}
}

func TestHealthEndpoint(t *testing.T) {
	ts := setupTestServer(t)

	rec := ts.do("GET", "/health", nil, "")

	if rec.Code != http.StatusOK {
		t.Fatalf("health: status = %d, want %d\nbody: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var resp map[string]string
	parseJSON(t, rec, &resp)
	if resp["status"] != "ok" {
		t.Errorf("health: status = %q, want %q", resp["status"], "ok")
	}
}

func TestEmptyTimeline(t *testing.T) {
	ts := setupTestServer(t)

	// Register and get token
	rec := ts.do("POST", "/api/v1/auth/register", models.RegisterRequest{
		Username: "emptyuser",
		Email:    "empty@example.com",
		Password: "securepass123",
	}, "")

	var authResp models.AuthResponse
	parseJSON(t, rec, &authResp)
	token := authResp.Tokens.AccessToken

	// Get empty timeline
	rec = ts.do("GET", "/api/v1/timeline", nil, token)

	if rec.Code != http.StatusOK {
		t.Fatalf("empty timeline: status = %d, want %d\nbody: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var timelineResp models.TimelineResponse
	parseJSON(t, rec, &timelineResp)
	if len(timelineResp.Posts) != 0 {
		t.Errorf("empty timeline: got %d posts, want 0", len(timelineResp.Posts))
	}
	if timelineResp.HasMore {
		t.Error("empty timeline: has_more should be false")
	}
	if timelineResp.NextCursor != nil {
		t.Error("empty timeline: next_cursor should be nil")
	}
}

func TestRegisterDuplicateEmail(t *testing.T) {
	ts := setupTestServer(t)

	// Register first user
	rec := ts.do("POST", "/api/v1/auth/register", models.RegisterRequest{
		Username: "emailuser1",
		Email:    "same@example.com",
		Password: "securepass123",
	}, "")

	if rec.Code != http.StatusCreated {
		t.Fatalf("first register: status = %d, want %d\nbody: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	// Try to register with same email but different username
	rec = ts.do("POST", "/api/v1/auth/register", models.RegisterRequest{
		Username: "emailuser2",
		Email:    "same@example.com",
		Password: "securepass123",
	}, "")

	if rec.Code != http.StatusConflict {
		t.Fatalf("dup email register: status = %d, want %d\nbody: %s", rec.Code, http.StatusConflict, rec.Body.String())
	}

	var errResp map[string]models.APIError
	parseJSON(t, rec, &errResp)
	if errResp["error"].Code != models.ErrCodeConflict {
		t.Errorf("dup email register: error code = %q, want %q", errResp["error"].Code, models.ErrCodeConflict)
	}
}

func TestCreatePostEmptyContent(t *testing.T) {
	ts := setupTestServer(t)

	// Register and get token
	rec := ts.do("POST", "/api/v1/auth/register", models.RegisterRequest{
		Username: "emptypost",
		Email:    "emptypost@example.com",
		Password: "securepass123",
	}, "")

	var authResp models.AuthResponse
	parseJSON(t, rec, &authResp)
	token := authResp.Tokens.AccessToken

	// Try to create a post with whitespace-only content
	rec = ts.do("POST", "/api/v1/posts", map[string]string{
		"content": "   ",
	}, token)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("empty post: status = %d, want %d\nbody: %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}

	var errResp map[string]models.APIError
	parseJSON(t, rec, &errResp)
	if errResp["error"].Code != models.ErrCodeValidation {
		t.Errorf("empty post: error code = %q, want %q", errResp["error"].Code, models.ErrCodeValidation)
	}
}

func TestTimelineCursorPagination(t *testing.T) {
	ts := setupTestServer(t)

	// Register and get token
	rec := ts.do("POST", "/api/v1/auth/register", models.RegisterRequest{
		Username: "paginuser",
		Email:    "pagin@example.com",
		Password: "securepass123",
	}, "")

	var authResp models.AuthResponse
	parseJSON(t, rec, &authResp)
	token := authResp.Tokens.AccessToken

	// Create 5 posts with small delays to ensure distinct created_at timestamps
	for i := 0; i < 5; i++ {
		rec = ts.do("POST", "/api/v1/posts", map[string]string{
			"content": fmt.Sprintf("Post number %d", i+1),
		}, token)
		if rec.Code != http.StatusCreated {
			t.Fatalf("create post %d: status = %d, want %d\nbody: %s", i+1, rec.Code, http.StatusCreated, rec.Body.String())
		}
		// Small delay to ensure distinct timestamps
		time.Sleep(10 * time.Millisecond)
	}

	// Fetch first page with limit=2
	rec = ts.do("GET", "/api/v1/timeline?limit=2", nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("page 1: status = %d, want %d\nbody: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var page1 models.TimelineResponse
	parseJSON(t, rec, &page1)
	if len(page1.Posts) != 2 {
		t.Fatalf("page 1: got %d posts, want 2", len(page1.Posts))
	}
	if !page1.HasMore {
		t.Error("page 1: has_more should be true")
	}
	if page1.NextCursor == nil {
		t.Fatal("page 1: next_cursor should not be nil")
	}

	// Fetch second page using the cursor (URL-encode to handle + in timezone)
	rec = ts.do("GET", "/api/v1/timeline?limit=2&cursor="+url.QueryEscape(*page1.NextCursor), nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("page 2: status = %d, want %d\nbody: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var page2 models.TimelineResponse
	parseJSON(t, rec, &page2)
	if len(page2.Posts) != 2 {
		t.Fatalf("page 2: got %d posts, want 2", len(page2.Posts))
	}

	// Verify no overlap between page 1 and page 2
	page1IDs := make(map[string]bool)
	for _, p := range page1.Posts {
		page1IDs[p.ID] = true
	}
	for _, p := range page2.Posts {
		if page1IDs[p.ID] {
			t.Errorf("page 2 contains post %s which was already in page 1", p.ID)
		}
	}

	// Fetch third page — should have 1 remaining post
	if page2.NextCursor == nil {
		t.Fatal("page 2: next_cursor should not be nil")
	}
	rec = ts.do("GET", "/api/v1/timeline?limit=2&cursor="+url.QueryEscape(*page2.NextCursor), nil, token)
	if rec.Code != http.StatusOK {
		t.Fatalf("page 3: status = %d, want %d\nbody: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var page3 models.TimelineResponse
	parseJSON(t, rec, &page3)
	if len(page3.Posts) != 1 {
		t.Fatalf("page 3: got %d posts, want 1", len(page3.Posts))
	}
	if page3.HasMore {
		t.Error("page 3: has_more should be false")
	}
}

func TestUpdateUserTooLongDisplayName(t *testing.T) {
	ts := setupTestServer(t)

	// Register and get token
	rec := ts.do("POST", "/api/v1/auth/register", models.RegisterRequest{
		Username: "longname",
		Email:    "longname@example.com",
		Password: "securepass123",
	}, "")

	var authResp models.AuthResponse
	parseJSON(t, rec, &authResp)
	token := authResp.Tokens.AccessToken

	// Try to update with a 51-char display name
	longName := strings.Repeat("a", 51)
	rec = ts.do("PATCH", "/api/v1/users/me", models.UserUpdate{
		DisplayName: &longName,
	}, token)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("long display name: status = %d, want %d\nbody: %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}

	var errResp map[string]models.APIError
	parseJSON(t, rec, &errResp)
	if errResp["error"].Code != models.ErrCodeValidation {
		t.Errorf("long display name: error code = %q, want %q", errResp["error"].Code, models.ErrCodeValidation)
	}
}

func TestUpdateUserValidBio(t *testing.T) {
	ts := setupTestServer(t)

	// Register and get token
	rec := ts.do("POST", "/api/v1/auth/register", models.RegisterRequest{
		Username: "biouser",
		Email:    "biouser@example.com",
		Password: "securepass123",
	}, "")

	var authResp models.AuthResponse
	parseJSON(t, rec, &authResp)
	token := authResp.Tokens.AccessToken

	// Update with a valid bio
	newBio := "This is my new bio. Building cool things!"
	rec = ts.do("PATCH", "/api/v1/users/me", models.UserUpdate{
		Bio: &newBio,
	}, token)

	if rec.Code != http.StatusOK {
		t.Fatalf("update bio: status = %d, want %d\nbody: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var userResp map[string]models.User
	parseJSON(t, rec, &userResp)
	if userResp["user"].Bio != newBio {
		t.Errorf("update bio: bio = %q, want %q", userResp["user"].Bio, newBio)
	}
}

func TestGetUserByUsername(t *testing.T) {
	ts := setupTestServer(t)

	// Register a user
	rec := ts.do("POST", "/api/v1/auth/register", models.RegisterRequest{
		Username: "lookupuser",
		Email:    "lookup@example.com",
		Password: "securepass123",
	}, "")

	if rec.Code != http.StatusCreated {
		t.Fatalf("register: status = %d, want %d\nbody: %s", rec.Code, http.StatusCreated, rec.Body.String())
	}

	var authResp models.AuthResponse
	parseJSON(t, rec, &authResp)
	token := authResp.Tokens.AccessToken
	userID := authResp.User.ID

	// Get user by ID via GET /api/v1/users/{id}
	rec = ts.do("GET", "/api/v1/users/"+userID, nil, token)

	if rec.Code != http.StatusOK {
		t.Fatalf("get user: status = %d, want %d\nbody: %s", rec.Code, http.StatusOK, rec.Body.String())
	}

	var userResp map[string]models.User
	parseJSON(t, rec, &userResp)
	if userResp["user"].Username != "lookupuser" {
		t.Errorf("get user: username = %q, want %q", userResp["user"].Username, "lookupuser")
	}
	if userResp["user"].ID != userID {
		t.Errorf("get user: id = %q, want %q", userResp["user"].ID, userID)
	}
}

func TestTimelineInvalidCursor(t *testing.T) {
	ts := setupTestServer(t)

	// Register and get token
	rec := ts.do("POST", "/api/v1/auth/register", models.RegisterRequest{
		Username: "cursoruser",
		Email:    "cursor@example.com",
		Password: "securepass123",
	}, "")

	var authResp models.AuthResponse
	parseJSON(t, rec, &authResp)
	token := authResp.Tokens.AccessToken

	// Send an invalid cursor format
	rec = ts.do("GET", "/api/v1/timeline?cursor=not-a-date", nil, token)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("invalid cursor: status = %d, want %d\nbody: %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestTimelineInvalidLimit(t *testing.T) {
	ts := setupTestServer(t)

	// Register and get token
	rec := ts.do("POST", "/api/v1/auth/register", models.RegisterRequest{
		Username: "limituser",
		Email:    "limit@example.com",
		Password: "securepass123",
	}, "")

	var authResp models.AuthResponse
	parseJSON(t, rec, &authResp)
	token := authResp.Tokens.AccessToken

	// Send an invalid limit (0)
	rec = ts.do("GET", "/api/v1/timeline?limit=0", nil, token)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("invalid limit: status = %d, want %d\nbody: %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestGetUserPostsInvalidCursor(t *testing.T) {
	ts := setupTestServer(t)

	// Register and get token
	rec := ts.do("POST", "/api/v1/auth/register", models.RegisterRequest{
		Username: "upcursor",
		Email:    "upcursor@example.com",
		Password: "securepass123",
	}, "")

	var authResp models.AuthResponse
	parseJSON(t, rec, &authResp)
	token := authResp.Tokens.AccessToken
	userID := authResp.User.ID

	// Send an invalid cursor format for user posts
	rec = ts.do("GET", "/api/v1/users/"+userID+"/posts?cursor=bad-cursor", nil, token)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("invalid user posts cursor: status = %d, want %d\nbody: %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}

func TestGetUserPostsInvalidLimit(t *testing.T) {
	ts := setupTestServer(t)

	// Register and get token
	rec := ts.do("POST", "/api/v1/auth/register", models.RegisterRequest{
		Username: "uplimit",
		Email:    "uplimit@example.com",
		Password: "securepass123",
	}, "")

	var authResp models.AuthResponse
	parseJSON(t, rec, &authResp)
	token := authResp.Tokens.AccessToken
	userID := authResp.User.ID

	// Send an invalid limit (101, over max)
	rec = ts.do("GET", "/api/v1/users/"+userID+"/posts?limit=101", nil, token)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("invalid user posts limit: status = %d, want %d\nbody: %s", rec.Code, http.StatusBadRequest, rec.Body.String())
	}
}
