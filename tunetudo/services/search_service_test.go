package services

import (
	"testing"
	"log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestSearchService(t *testing.T) (*SearchService, func()) {
	db := setupTestDB(t)
	seedTestData(t, db)
	service := NewSearchService(db)
	
	cleanup := func() {
		db.Close()
	}
	
	return service, cleanup
}

func TestFullTextSearch(t *testing.T) {
	service, cleanup := setupTestSearchService(t)
	defer cleanup()

	tests := []struct {
		name          string
		query         string
		expectSongs   bool
		expectArtists bool
		expectAlbums  bool
	}{
		{
			name:          "Search by song title",
			query:         "Test Song",
			expectSongs:   true,
			expectArtists: false,
			expectAlbums:  false,
		},
		{
			name:          "Search by album",
			query:         "Test Album",
			expectSongs:   true,
			expectArtists: false,
			expectAlbums:  true,
		},
		{
			name:          "Empty query",
			query:         "",
			expectSongs:   false,
			expectArtists: false,
			expectAlbums:  false,
		},
		{
			name:          "No results",
			query:         "NonExistent",
			expectSongs:   false,
			expectArtists: false,
			expectAlbums:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := service.FullTextSearch(tt.query)
			log.Printf("Search results for query '%s': %+v", tt.query, result)
			require.NoError(t, err)
			assert.NotNil(t, result)

			if tt.expectSongs {
				assert.Greater(t, len(result.Songs), 0)
			} else {
				assert.Len(t, result.Songs, 0)
			}
			if tt.expectAlbums {
				assert.Greater(t, len(result.Albums), 0)
			} else {
				assert.Len(t, result.Albums, 0)
			}
		})
	}
}

func TestFullTextSearchMalicious(t *testing.T) {
	service, cleanup := setupTestSearchService(t)
	defer cleanup()

	tests := []struct {
		name          string
		query         string
		expectSongs   bool
		expectArtists bool
		expectAlbums  bool
	}{
		{
			name:          "Malicious input",
			query:         "test' OR 1=1;",
			expectSongs:   true,
			expectArtists: false,
			expectAlbums:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _ := service.FullTextSearch(tt.query)
			log.Printf("Search results for query '%s': %+v", tt.query, result)
			assert.NotNil(t, result)

			if tt.expectSongs {
				assert.Len(t, result.Songs, 0)
			} else {
				assert.Len(t, result.Songs, 0)
			}
			if tt.expectAlbums {
				assert.Greater(t, len(result.Albums), 0)
			} else {
				assert.Len(t, result.Albums, 0)
			}
		})
	}
}


func TestGetSongsByCategory(t *testing.T) {
	service, cleanup := setupTestSearchService(t)
	defer cleanup()

	tests := []struct {
		name        string
		categoryID  int
		expectSongs bool
	}{
		{
			name:        "Valid category with songs",
			categoryID:  1,
			expectSongs: true,
		},
		{
			name:        "Valid category without songs",
			categoryID:  2,
			expectSongs: false,
		},
		{
			name:        "Non-existent category",
			categoryID:  99999,
			expectSongs: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			songs, err := service.GetSongsByCategory(tt.categoryID)
			require.NoError(t, err)

			if tt.expectSongs {
				assert.Greater(t, len(songs), 0)
				// Verify all songs belong to the category
				for _, song := range songs {
					assert.NotNil(t, song.CategoryID)
					assert.Equal(t, tt.categoryID, *song.CategoryID)
				}
			} else {
				assert.Len(t, songs, 0)
			}
		})
	}
}

func TestGetAllCategories(t *testing.T) {
	service, cleanup := setupTestSearchService(t)
	defer cleanup()

	categories, err := service.GetAllCategories()
	require.NoError(t, err)
	assert.Greater(t, len(categories), 0)
	
	// Verify categories are sorted by name
	for i := 0; i < len(categories)-1; i++ {
		assert.LessOrEqual(t, categories[i].Name, categories[i+1].Name)
	}
	
	// Verify each category has required fields
	for _, cat := range categories {
		assert.Greater(t, cat.ID, 0)
		assert.NotEmpty(t, cat.Name)
	}
}

func TestSearchExcludesUserUploads(t *testing.T) {
	service, cleanup := setupTestSearchService(t)
	defer cleanup()

	db := service.db
	
	// Insert a user-uploaded song
	db.Exec(`INSERT INTO users (username, email, password_hash) VALUES (?, ?, ?)`,
		"testuser", "test@test.com", "hash")
	
	var userID int
	db.QueryRow("SELECT id FROM users WHERE username = ?", "testuser").Scan(&userID)
	
	db.Exec(`INSERT INTO songs (title, artist_id, file_path, format, uploaded_by_user_id)
		VALUES (?, ?, ?, ?, ?)`,
		"User Upload Song", 1, "/test/user.mp3", "mp3", userID)

	// Search should NOT return user uploads
	result, err := service.FullTextSearch("User Upload")
	require.NoError(t, err)
	assert.Len(t, result.Songs, 0, "User uploads should not appear in search")
}

func TestCategorySearchExcludesUserUploads(t *testing.T) {
	service, cleanup := setupTestSearchService(t)
	defer cleanup()

	db := service.db
	
	// Insert a user-uploaded song with category
	db.Exec(`INSERT INTO users (username, email, password_hash) VALUES (?, ?, ?)`,
		"testuser", "test@test.com", "hash")
	
	var userID int
	db.QueryRow("SELECT id FROM users WHERE username = ?", "testuser").Scan(&userID)
	
	db.Exec(`INSERT INTO songs (title, artist_id, category_id, file_path, format, uploaded_by_user_id)
		VALUES (?, ?, ?, ?, ?, ?)`,
		"User Upload Song", 1, 1, "/test/user.mp3", "mp3", userID)

	// Category search should NOT return user uploads
	songs, err := service.GetSongsByCategory(1)
	require.NoError(t, err)
	
	// Check that user uploads are not in results
	for _, song := range songs {
		assert.Nil(t, song.UploadedByUserID, "User uploads should not appear in category search")
	}
}