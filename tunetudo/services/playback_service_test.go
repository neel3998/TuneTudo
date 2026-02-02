package services

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestPlaybackService(t *testing.T) (*PlaybackService, func()) {
	db := setupTestDB(t)
	seedTestData(t, db)
	
	// Create temp storage directory
	storageDir := "./test_storage_" + t.Name()
	os.MkdirAll(filepath.Join(storageDir, "test"), 0755)
	
	// Create a dummy audio file
	testFile := filepath.Join(storageDir, "test", "song.mp3")
	os.WriteFile(testFile, []byte("dummy audio content"), 0644)
	
	service := NewPlaybackService(db, storageDir)
	
	cleanup := func() {
		db.Close()
		os.RemoveAll(storageDir)
	}
	
	return service, cleanup
}

func TestGetSongByID(t *testing.T) {
	service, cleanup := setupTestPlaybackService(t)
	defer cleanup()

	tests := []struct {
		name        string
		songID      int
		expectError bool
	}{
		{
			name:        "Existing song",
			songID:      1,
			expectError: false,
		},
		{
			name:        "Non-existent song",
			songID:      99999,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			song, err := service.GetSongByID(tt.songID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, song)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, song)
				assert.Equal(t, tt.songID, song.ID)
				assert.NotEmpty(t, song.Title)
				assert.NotNil(t, song.Artist)
			}
		})
	}
}

func TestAuthorizeStream(t *testing.T) {
	service, cleanup := setupTestPlaybackService(t)
	defer cleanup()

	tests := []struct {
		name        string
		songID      int
		expectError bool
		errorMsg    string
	}{
		{
			name:        "Valid song with existing file",
			songID:      1,
			expectError: false,
		},
		{
			name:        "Non-existent song",
			songID:      99999,
			expectError: true,
			errorMsg:    "track not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filePath, err := service.AuthorizeStream(tt.songID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, filePath)
				if tt.errorMsg != "" {
					assert.Contains(t, err.Error(), tt.errorMsg)
				}
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, filePath)
			}
		})
	}
}

func TestGetRecentSongs(t *testing.T) {
	service, cleanup := setupTestPlaybackService(t)
	defer cleanup()

	tests := []struct {
		name         string
		limit        int
		expectedLen  int
	}{
		{
			name:        "Default limit",
			limit:       20,
			expectedLen: 3, // We seeded 3 songs
		},
		{
			name:        "Small limit",
			limit:       2,
			expectedLen: 2,
		},
		{
			name:        "Zero limit uses default",
			limit:       0,
			expectedLen: 3,
		},
		{
			name:        "Large limit",
			limit:       100,
			expectedLen: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			songs, err := service.GetRecentSongs(tt.limit)
			require.NoError(t, err)
			assert.Len(t, songs, tt.expectedLen)
			
			// Verify songs are ordered by creation date (newest first)
			for i := 0; i < len(songs)-1; i++ {
				assert.True(t, songs[i].CreatedAt.After(songs[i+1].CreatedAt) ||
					songs[i].CreatedAt.Equal(songs[i+1].CreatedAt))
			}
		})
	}
}

func TestGetRecentSongsExcludesUserUploads(t *testing.T) {
	service, cleanup := setupTestPlaybackService(t)
	defer cleanup()

	db := service.db
	
	// Insert a user and user-uploaded song
	db.Exec(`INSERT INTO users (username, email, password_hash) VALUES (?, ?, ?)`,
		"testuser", "test@test.com", "hash")
	
	var userID int
	db.QueryRow("SELECT id FROM users WHERE username = ?", "testuser").Scan(&userID)
	
	db.Exec(`INSERT INTO songs (title, artist_id, file_path, format, uploaded_by_user_id)
		VALUES (?, ?, ?, ?, ?)`,
		"User Upload Song", 1, "/test/user.mp3", "mp3", userID)

	// Recent songs should NOT include user uploads
	songs, err := service.GetRecentSongs(20)
	require.NoError(t, err)
	
	for _, song := range songs {
		assert.Nil(t, song.UploadedByUserID, "User uploads should not appear in recent songs")
	}
}