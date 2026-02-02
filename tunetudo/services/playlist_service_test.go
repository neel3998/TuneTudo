package services

import (
	"testing"
	"tunetudo/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestPlaylistService(t *testing.T) (*PlaylistService, *AuthService, int, func()) {
	db := setupTestDB(t)
	seedTestData(t, db)
	
	playlistService := NewPlaylistService(db)
	authService := NewAuthService(db, "test-secret")
	
	// Create a test user
	ip := "127.0.0.1"
	user, err := authService.RegisterUser(models.RegisterRequest{
		Username: "playlistuser",
		Email:    "playlist@test.com",
		Password: "password123",
	}, ip)
	require.NoError(t, err)
	
	cleanup := func() {
		db.Close()
	}
	
	return playlistService, authService, user.ID, cleanup
}

func TestCreatePlaylist(t *testing.T) {
	service, _, userID, cleanup := setupTestPlaylistService(t)
	defer cleanup()

	tests := []struct {
		name        string
		req         models.CreatePlaylistRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid playlist",
			req: models.CreatePlaylistRequest{
				Name:        "My Playlist",
				Description: stringPtr("Test description"),
			},
			expectError: false,
		},
		{
			name: "Empty name",
			req: models.CreatePlaylistRequest{
				Name: "",
			},
			expectError: true,
			errorMsg:    "enter valid playlist name",
		},
		{
			name: "Duplicate playlist name",
			req: models.CreatePlaylistRequest{
				Name: "My Playlist",
			},
			expectError: true,
			errorMsg:    "Playlist already exists",
		},
		{
			name: "Playlist without description",
			req: models.CreatePlaylistRequest{
				Name: "Another Playlist",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			playlist, err := service.CreatePlaylist(userID, tt.req)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, playlist)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, playlist)
				assert.Equal(t, tt.req.Name, playlist.Name)
				assert.Equal(t, userID, playlist.UserID)
				assert.Greater(t, playlist.ID, 0)
			}
		})
	}
}

func TestGetUserPlaylists(t *testing.T) {
	service, _, userID, cleanup := setupTestPlaylistService(t)
	defer cleanup()

	// Create test playlists
	service.CreatePlaylist(userID, models.CreatePlaylistRequest{Name: "Playlist 1"})
	service.CreatePlaylist(userID, models.CreatePlaylistRequest{Name: "Playlist 2"})

	playlists, err := service.GetUserPlaylists(userID)
	require.NoError(t, err)
	assert.Len(t, playlists, 2)
	assert.Equal(t, "Playlist 1", playlists[1].Name)
	assert.Equal(t, "Playlist 2", playlists[0].Name)
}

func TestGetPlaylistByID(t *testing.T) {
	service, _, userID, cleanup := setupTestPlaylistService(t)
	defer cleanup()

	// Create a playlist
	created, err := service.CreatePlaylist(userID, models.CreatePlaylistRequest{
		Name: "Test Playlist",
	})
	require.NoError(t, err)

	tests := []struct {
		name        string
		playlistID  int
		userID      int
		expectError bool
	}{
		{
			name:        "Valid playlist",
			playlistID:  created.ID,
			userID:      userID,
			expectError: false,
		},
		{
			name:        "Non-existent playlist",
			playlistID:  99999,
			userID:      userID,
			expectError: true,
		},
		{
			name:        "Wrong user",
			playlistID:  created.ID,
			userID:      99999,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			playlist, err := service.GetPlaylistByID(tt.playlistID, tt.userID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, playlist)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, playlist)
				assert.Equal(t, created.ID, playlist.ID)
			}
		})
	}
}

func TestAddSong(t *testing.T) {
	service, _, userID, cleanup := setupTestPlaylistService(t)
	defer cleanup()

	// Create a playlist
	playlist, err := service.CreatePlaylist(userID, models.CreatePlaylistRequest{
		Name: "Test Playlist",
	})
	require.NoError(t, err)

	tests := []struct {
		name        string
		playlistID  int
		songID      int
		userID      int
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Add song to playlist",
			playlistID:  playlist.ID,
			songID:      1,
			userID:      userID,
			expectError: false,
		},
		{
			name:        "Add duplicate song",
			playlistID:  playlist.ID,
			songID:      1,
			userID:      userID,
			expectError: true,
			errorMsg:    "Song already in the playlist",
		},
		{
			name:        "Wrong user",
			playlistID:  playlist.ID,
			songID:      2,
			userID:      99999,
			expectError: true,
			errorMsg:    "unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.AddSong(tt.playlistID, tt.songID, tt.userID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRemoveSong(t *testing.T) {
	service, _, userID, cleanup := setupTestPlaylistService(t)
	defer cleanup()

	// Create playlist and add song
	playlist, err := service.CreatePlaylist(userID, models.CreatePlaylistRequest{
		Name: "Test Playlist",
	})
	require.NoError(t, err)
	
	err = service.AddSong(playlist.ID, 1, userID)
	require.NoError(t, err)

	tests := []struct {
		name        string
		playlistID  int
		songID      int
		userID      int
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Remove existing song",
			playlistID:  playlist.ID,
			songID:      1,
			userID:      userID,
			expectError: false,
		},
		{
			name:        "Remove non-existent song",
			playlistID:  playlist.ID,
			songID:      99,
			userID:      userID,
			expectError: true,
			errorMsg:    "song not found in playlist",
		},
		{
			name:        "Wrong user",
			playlistID:  playlist.ID,
			songID:      1,
			userID:      99999,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.RemoveSong(tt.playlistID, tt.songID, tt.userID)

			if tt.expectError {
				assert.Error(t, err)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDeletePlaylist(t *testing.T) {
	service, _, userID, cleanup := setupTestPlaylistService(t)
	defer cleanup()

	// Create a playlist
	playlist, err := service.CreatePlaylist(userID, models.CreatePlaylistRequest{
		Name: "Test Playlist",
	})
	require.NoError(t, err)

	tests := []struct {
		name        string
		playlistID  int
		userID      int
		expectError bool
	}{
		{
			name:        "Delete own playlist",
			playlistID:  playlist.ID,
			userID:      userID,
			expectError: false,
		},
		{
			name:        "Delete non-existent playlist",
			playlistID:  99999,
			userID:      userID,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.DeletePlaylist(tt.playlistID, tt.userID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetPlaylistSongs(t *testing.T) {
	service, _, userID, cleanup := setupTestPlaylistService(t)
	defer cleanup()

	// Create playlist and add songs
	playlist, err := service.CreatePlaylist(userID, models.CreatePlaylistRequest{
		Name: "Test Playlist",
	})
	require.NoError(t, err)

	service.AddSong(playlist.ID, 1, userID)
	service.AddSong(playlist.ID, 2, userID)

	songs, err := service.GetPlaylistSongs(playlist.ID)
	require.NoError(t, err)
	assert.Len(t, songs, 2)
	assert.NotNil(t, songs[0].Song)
	assert.Equal(t, 1, songs[0].SongID)
}

// Helper function
func stringPtr(s string) *string {
	return &s
}