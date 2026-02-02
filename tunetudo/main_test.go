package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"fmt"
	"tunetudo/database"
	"tunetudo/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestApp(t *testing.T) (*fiber.App, func()) {
	// Create test database
	dbPath := "./test_integration.db"
	os.Remove(dbPath)
	
	db, err := database.InitDB(dbPath)
	require.NoError(t, err)
	
	err = database.RunMigrations(db)
	require.NoError(t, err)
	
	// Create test app
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   true,
				"message": err.Error(),
			})
		},
	})
	
	routes.SetupRoutes(app, db)
	
	cleanup := func() {
		db.Close()
		os.Remove(dbPath)
	}
	
	return app, cleanup
}

func TestAuthFlow(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	// Test registration
	t.Run("Register", func(t *testing.T) {
		body := map[string]string{
			"username": "testuser",
			"email":    "test@example.com",
			"password": "password123",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.False(t, result["error"].(bool))
		assert.Contains(t, result, "data")
	})

	// Test login
	var authToken string
	t.Run("Login", func(t *testing.T) {
		body := map[string]string{
			"username": "testuser",
			"password": "password123",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.False(t, result["error"].(bool))
		
		data := result["data"].(map[string]interface{})
		authToken = data["token"].(string)
		assert.NotEmpty(t, authToken)
	})

	// Test protected route
	t.Run("Access Protected Route", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/profile", nil)
		req.Header.Set("Authorization", "Bearer "+authToken)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// Test unauthorized access
	t.Run("Unauthorized Access", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/profile", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

func TestPlaylistFlow(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	// Register and login first
	registerBody := map[string]string{
		"username": "playlistuser",
		"email":    "playlist@example.com",
		"password": "password123",
	}
	jsonBody, _ := json.Marshal(registerBody)
	req := httptest.NewRequest("POST", "/api/auth/register", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	app.Test(req)

	loginBody := map[string]string{
		"username": "playlistuser",
		"password": "password123",
	}
	jsonBody, _ = json.Marshal(loginBody)
	req = httptest.NewRequest("POST", "/api/auth/login", bytes.NewReader(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	var loginResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&loginResult)
	token := loginResult["data"].(map[string]interface{})["token"].(string)

	var playlistID float64

	// Create playlist
	t.Run("Create Playlist", func(t *testing.T) {
		body := map[string]string{
			"name":        "My Test Playlist",
			"description": "Test description",
		}
		jsonBody, _ := json.Marshal(body)

		req := httptest.NewRequest("POST", "/api/playlists", bytes.NewReader(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.False(t, result["error"].(bool))
		
		data := result["data"].(map[string]interface{})
		playlistID = data["id"].(float64)
		assert.Greater(t, playlistID, 0.0)
	})

	// Get playlists
	t.Run("Get Playlists", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/playlists", nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.False(t, result["error"].(bool))
		
		playlists := result["data"].([]interface{})
		assert.Len(t, playlists, 1)
	})

	// Delete playlist
	t.Run("Delete Playlist", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/api/playlists/"+fmt.Sprint(int(playlistID)), nil)
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestSearchFlow(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	// Test search (public endpoint)
	t.Run("Search Songs", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/search?q=test", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.False(t, result["error"].(bool))
		assert.Contains(t, result, "data")
	})

	// Test categories
	t.Run("Get Categories", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/categories", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		assert.False(t, result["error"].(bool))
		
		categories := result["data"].([]interface{})
		assert.Greater(t, len(categories), 0)
	})

	// Test recent songs
	t.Run("Get Recent Songs", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/songs/recent", nil)

		resp, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestHealthCheck(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	req := httptest.NewRequest("GET", "/health", nil)
	resp, err := app.Test(req)
	
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	assert.Equal(t, "ok", result["status"])
}

func TestInvalidRoutes(t *testing.T) {
	app, cleanup := setupTestApp(t)
	defer cleanup()

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{"Invalid API endpoint", "GET", "/api/invalid"},
		{"Invalid method", "PUT", "/api/search"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			_, err := app.Test(req)		
			require.NoError(t, err)
		})
	}
}
