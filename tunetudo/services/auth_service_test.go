package services

import (
	"testing"
	"tunetudo/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestAuthService(t *testing.T) (*AuthService, func()) {
	db := setupTestDB(t)
	service := NewAuthService(db, "test-secret-key")
	
	cleanup := func() {
		db.Close()
	}
	
	return service, cleanup
}

func TestRegisterUser(t *testing.T) {
	service, cleanup := setupTestAuthService(t)
	defer cleanup()

	tests := []struct {
		name        string
		req         models.RegisterRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid registration",
			req: models.RegisterRequest{
				Username: "testuser",
				Email:    "test@example.com",
				Password: "password123",
			},
			expectError: false,
		},
		{
			name: "Password too short",
			req: models.RegisterRequest{
				Username: "testuser2",
				Email:    "test2@example.com",
				Password: "short",
			},
			expectError: true,
			errorMsg:    "password must be at least 8 characters",
		},
		{
			name: "Duplicate username",
			req: models.RegisterRequest{
				Username: "testuser",
				Email:    "another@example.com",
				Password: "password123",
			},
			expectError: true,
			errorMsg:    "username or email already exists",
		},
		{
			name: "Duplicate email",
			req: models.RegisterRequest{
				Username: "anotheruser",
				Email:    "test@example.com",
				Password: "password123",
			},
			expectError: true,
			errorMsg:    "username or email already exists",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := "127.0.0.1"
			user, err := service.RegisterUser(tt.req, ip)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, user)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.req.Username, user.Username)
				assert.Equal(t, tt.req.Email, user.Email)
				assert.False(t, user.IsAdmin)
				assert.Greater(t, user.ID, 0)
			}
		})
	}
}

func TestLoginUser(t *testing.T) {
	service, cleanup := setupTestAuthService(t)
	defer cleanup()

	// Register a test user first
	regReq := models.RegisterRequest{
		Username: "logintest",
		Email:    "login@example.com",
		Password: "password123",
	}
	ip := "127.0.0.1"
	_, err := service.RegisterUser(regReq, ip)
	require.NoError(t, err)

	tests := []struct {
		name        string
		req         models.LoginRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid login with username",
			req: models.LoginRequest{
				Username: "logintest",
				Password: "password123",
			},
			expectError: false,
		},
		{
			name: "Invalid password",
			req: models.LoginRequest{
				Username: "logintest",
				Password: "wrongpassword",
			},
			expectError: true,
			errorMsg:    "authorization failed",
		},
		{
			name: "Non-existent user",
			req: models.LoginRequest{
				Username: "nonexistent",
				Password: "password123",
			},
			expectError: true,
			errorMsg:    "authorization failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip := "127.0.0.1"
			token, user, err := service.LoginUser(tt.req, ip)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Empty(t, token)
				assert.Nil(t, user)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, token)
				assert.NotNil(t, user)
				assert.Equal(t, "logintest", user.Username)
			}
		})
	}
}

func TestGenerateToken(t *testing.T) {
	service, cleanup := setupTestAuthService(t)
	defer cleanup()

	user := &models.User{
		ID:       1,
		Username: "testuser",
		Email:    "test@example.com",
		IsAdmin:  false,
	}

	token, err := service.GenerateToken(user)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Validate the token
	claims, err := service.ValidateToken(token)
	require.NoError(t, err)
	assert.Equal(t, float64(user.ID), claims["user_id"])
	assert.Equal(t, user.Username, claims["username"])
	assert.Equal(t, user.IsAdmin, claims["is_admin"])
}

func TestValidateToken(t *testing.T) {
	service, cleanup := setupTestAuthService(t)
	defer cleanup()

	user := &models.User{
		ID:       1,
		Username: "testuser",
		IsAdmin:  true,
	}

	token, err := service.GenerateToken(user)
	require.NoError(t, err)

	tests := []struct {
		name        string
		token       string
		expectError bool
	}{
		{
			name:        "Valid token",
			token:       token,
			expectError: false,
		},
		{
			name:        "Invalid token",
			token:       "invalid.token.here",
			expectError: true,
		},
		{
			name:        "Empty token",
			token:       "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claims, err := service.ValidateToken(tt.token)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, claims)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, claims)
			}
		})
	}
}

func TestGetUserByID(t *testing.T) {
	service, cleanup := setupTestAuthService(t)
	defer cleanup()

	// Register a test user
	regReq := models.RegisterRequest{
		Username: "gettest",
		Email:    "get@example.com",
		Password: "password123",
	}
	ip := "127.0.0.1"
	registeredUser, err := service.RegisterUser(regReq, ip)
	require.NoError(t, err)

	tests := []struct {
		name        string
		userID      int
		expectError bool
	}{
		{
			name:        "Existing user",
			userID:      registeredUser.ID,
			expectError: false,
		},
		{
			name:        "Non-existent user",
			userID:      99999,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := service.GetUserByID(tt.userID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, user)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, user)
				assert.Equal(t, tt.userID, user.ID)
				assert.Equal(t, "gettest", user.Username)
			}
		})
	}
}

func TestCheckPasswordPolicy(t *testing.T) {
	service, cleanup := setupTestAuthService(t)
	defer cleanup()

	tests := []struct {
		name        string
		password    string
		expectError bool
	}{
		{
			name:        "Valid password",
			password:    "password123",
			expectError: false,
		},
		{
			name:        "Password too short",
			password:    "short",
			expectError: true,
		},
		{
			name:        "Exactly 8 characters",
			password:    "12345678",
			expectError: false,
		},
		{
			name:        "Empty password",
			password:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.CheckPasswordPolicy(tt.password)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}