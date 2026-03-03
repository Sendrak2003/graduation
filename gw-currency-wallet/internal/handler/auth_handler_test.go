package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"gw-currency-wallet/internal/models"
	"gw-currency-wallet/internal/service"
	"gw-currency-wallet/internal/utils/auth"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type mockUserRepository struct {
	createUserFunc    func(ctx context.Context, userID, username, email, passwordHash string) error
	getByUsernameFunc func(ctx context.Context, username string) (*models.User, error)
	getByIDFunc       func(ctx context.Context, userID string) (*models.User, error)
}

func (m *mockUserRepository) CreateUser(ctx context.Context, userID, username, email, passwordHash string) error {
	if m.createUserFunc != nil {
		return m.createUserFunc(ctx, userID, username, email, passwordHash)
	}
	return nil
}

func (m *mockUserRepository) GetByUsername(ctx context.Context, username string) (*models.User, error) {
	if m.getByUsernameFunc != nil {
		return m.getByUsernameFunc(ctx, username)
	}
	return nil, errors.New("user not found")
}

func (m *mockUserRepository) GetByID(ctx context.Context, userID string) (*models.User, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, userID)
	}
	return nil, errors.New("user not found")
}

func TestRegister(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    interface{}
		mockCreate     func(ctx context.Context, userID, username, email, passwordHash string) error
		wantStatusCode int
		wantErrSubstr  string
		checkResponse  func(t *testing.T, body map[string]interface{})
	}{
		{
			name: "valid registration",
			requestBody: RegisterRequest{
				Username: "testuser",
				Password: "password123",
				Email:    "test@example.com",
			},
			mockCreate: func(ctx context.Context, userID, username, email, passwordHash string) error {
				return nil
			},
			wantStatusCode: http.StatusCreated,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				if _, ok := body["access_token"]; !ok {
					t.Fatal("expected access_token in response")
				}
				if msg, ok := body["message"].(string); !ok || msg != "User registered successfully" {
					t.Fatalf("unexpected message: %v", body["message"])
				}
			},
		},
		{
			name: "username too short",
			requestBody: RegisterRequest{
				Username: "ab",
				Password: "password123",
				Email:    "test@example.com",
			},
			wantStatusCode: http.StatusBadRequest,
			wantErrSubstr:  "min",
		},
		{
			name: "username not alphanum",
			requestBody: RegisterRequest{
				Username: "test@user",
				Password: "password123",
				Email:    "test@example.com",
			},
			wantStatusCode: http.StatusBadRequest,
			wantErrSubstr:  "alphanum",
		},
		{
			name: "password too short",
			requestBody: RegisterRequest{
				Username: "testuser",
				Password: "12345",
				Email:    "test@example.com",
			},
			wantStatusCode: http.StatusBadRequest,
			wantErrSubstr:  "min",
		},
		{
			name: "invalid email",
			requestBody: RegisterRequest{
				Username: "testuser",
				Password: "password123",
				Email:    "invalid-email",
			},
			wantStatusCode: http.StatusBadRequest,
			wantErrSubstr:  "email",
		},
		{
			name: "missing username",
			requestBody: map[string]interface{}{
				"password": "password123",
				"email":    "test@example.com",
			},
			wantStatusCode: http.StatusBadRequest,
			wantErrSubstr:  "required",
		},
		{
			name: "user already exists",
			requestBody: RegisterRequest{
				Username: "existinguser",
				Password: "password123",
				Email:    "test@example.com",
			},
			mockCreate: func(ctx context.Context, userID, username, email, passwordHash string) error {
				return errors.New("user already exists")
			},
			wantStatusCode: http.StatusBadRequest,
			wantErrSubstr:  "user already exists",
		},
		{
			name:           "invalid json",
			requestBody:    "invalid json",
			wantStatusCode: http.StatusBadRequest,
			wantErrSubstr:  "error",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := &mockUserRepository{
				createUserFunc: tt.mockCreate,
			}

			userService := service.NewUserService(mockRepo)
			jwtManager := auth.New("test-secret", time.Hour, time.Hour*24)
			handler := NewAuthHandler(userService, jwtManager)

			router := gin.New()
			router.POST("/register", handler.Register)

			var body []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				if err != nil {
					t.Fatalf("failed to marshal request: %v", err)
				}
			}

			req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatusCode {
				t.Fatalf("unexpected status code: got %d, want %d", w.Code, tt.wantStatusCode)
			}

			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			if tt.wantErrSubstr != "" {
				errMsg, ok := response["error"].(string)
				if !ok {
					t.Fatal("expected error field in response")
				}
				if !strings.Contains(errMsg, tt.wantErrSubstr) {
					t.Fatalf("expected error containing %q, got %q", tt.wantErrSubstr, errMsg)
				}
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, response)
			}
		})
	}
}

func TestLogin(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	validPasswordHash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	tests := []struct {
		name              string
		requestBody       interface{}
		mockGetByUsername func(ctx context.Context, username string) (*models.User, error)
		wantStatusCode    int
		wantErrSubstr     string
		checkResponse     func(t *testing.T, body map[string]interface{})
	}{
		{
			name: "valid login",
			requestBody: LoginRequest{
				Username: "testuser",
				Password: "password123",
			},
			mockGetByUsername: func(ctx context.Context, username string) (*models.User, error) {
				return &models.User{
					ID:           "user-123",
					Username:     "testuser",
					Email:        "test@example.com",
					PasswordHash: string(validPasswordHash),
				}, nil
			},
			wantStatusCode: http.StatusOK,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				if _, ok := body["access_token"]; !ok {
					t.Fatal("expected access_token in response")
				}
			},
		},
		{
			name: "user not found",
			requestBody: LoginRequest{
				Username: "nonexistent",
				Password: "password123",
			},
			mockGetByUsername: func(ctx context.Context, username string) (*models.User, error) {
				return nil, errors.New("user not found")
			},
			wantStatusCode: http.StatusUnauthorized,
			wantErrSubstr:  "invalid credentials",
		},
		{
			name: "wrong password",
			requestBody: LoginRequest{
				Username: "testuser",
				Password: "wrongpassword",
			},
			mockGetByUsername: func(ctx context.Context, username string) (*models.User, error) {
				return &models.User{
					ID:           "user-123",
					Username:     "testuser",
					Email:        "test@example.com",
					PasswordHash: string(validPasswordHash),
				}, nil
			},
			wantStatusCode: http.StatusUnauthorized,
			wantErrSubstr:  "invalid credentials",
		},
		{
			name: "missing username",
			requestBody: map[string]interface{}{
				"password": "password123",
			},
			wantStatusCode: http.StatusBadRequest,
			wantErrSubstr:  "required",
		},
		{
			name: "missing password",
			requestBody: map[string]interface{}{
				"username": "testuser",
			},
			wantStatusCode: http.StatusBadRequest,
			wantErrSubstr:  "required",
		},
		{
			name:           "invalid json",
			requestBody:    "invalid json",
			wantStatusCode: http.StatusBadRequest,
			wantErrSubstr:  "error",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := &mockUserRepository{
				getByUsernameFunc: tt.mockGetByUsername,
			}

			userService := service.NewUserService(mockRepo)
			jwtManager := auth.New("test-secret", time.Hour, time.Hour*24)
			handler := NewAuthHandler(userService, jwtManager)

			router := gin.New()
			router.POST("/login", handler.Login)

			var body []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				if err != nil {
					t.Fatalf("failed to marshal request: %v", err)
				}
			}

			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatusCode {
				t.Fatalf("unexpected status code: got %d, want %d", w.Code, tt.wantStatusCode)
			}

			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			if tt.wantErrSubstr != "" {
				errMsg, ok := response["error"].(string)
				if !ok {
					t.Fatal("expected error field in response")
				}
				if !strings.Contains(errMsg, tt.wantErrSubstr) {
					t.Fatalf("expected error containing %q, got %q", tt.wantErrSubstr, errMsg)
				}
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, response)
			}
		})
	}
}

func TestRefresh(t *testing.T) {
	t.Parallel()

	gin.SetMode(gin.TestMode)

	jwtManager := auth.New("test-secret", time.Hour, time.Hour*24)
	validRefreshToken, _ := jwtManager.GenerateRefresh("user-123")

	tests := []struct {
		name           string
		requestBody    interface{}
		wantStatusCode int
		wantErrSubstr  string
		checkResponse  func(t *testing.T, body map[string]interface{})
	}{
		{
			name: "valid refresh token",
			requestBody: RefreshRequest{
				RefreshToken: validRefreshToken,
			},
			wantStatusCode: http.StatusOK,
			checkResponse: func(t *testing.T, body map[string]interface{}) {
				if _, ok := body["access_token"]; !ok {
					t.Fatal("expected access_token in response")
				}
			},
		},
		{
			name: "invalid refresh token",
			requestBody: RefreshRequest{
				RefreshToken: "invalid.token.here",
			},
			wantStatusCode: http.StatusUnauthorized,
			wantErrSubstr:  "invalid refresh token",
		},
		{
			name: "missing refresh token",
			requestBody: map[string]interface{}{
				"other_field": "value",
			},
			wantStatusCode: http.StatusBadRequest,
			wantErrSubstr:  "required",
		},
		{
			name:           "invalid json",
			requestBody:    "invalid json",
			wantStatusCode: http.StatusBadRequest,
			wantErrSubstr:  "error",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockRepo := &mockUserRepository{}
			userService := service.NewUserService(mockRepo)
			handler := NewAuthHandler(userService, jwtManager)

			router := gin.New()
			router.POST("/refresh", handler.Refresh)

			var body []byte
			var err error
			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				if err != nil {
					t.Fatalf("failed to marshal request: %v", err)
				}
			}

			req := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatusCode {
				t.Fatalf("unexpected status code: got %d, want %d", w.Code, tt.wantStatusCode)
			}

			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			if tt.wantErrSubstr != "" {
				errMsg, ok := response["error"].(string)
				if !ok {
					t.Fatal("expected error field in response")
				}
				if !strings.Contains(errMsg, tt.wantErrSubstr) {
					t.Fatalf("expected error containing %q, got %q", tt.wantErrSubstr, errMsg)
				}
			}

			if tt.checkResponse != nil {
				tt.checkResponse(t, response)
			}
		})
	}
}
