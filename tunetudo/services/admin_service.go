package services

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"tunetudo/logger"
	"tunetudo/models"

	"github.com/google/uuid"
)

type AdminService struct {
	db          *sql.DB
	storagePath string
}

func NewAdminService(db *sql.DB, storagePath string) *AdminService {
	return &AdminService{
		db:          db,
		storagePath: storagePath,
	}
}

// UploadSong uploads a new song to the catalog (admin only)
func (s *AdminService) UploadSong(
	file *multipart.FileHeader,
	title, artistName, albumTitle string,
	categoryID, durationSeconds int,
) (*models.Song, error) {
	logger.Info(logger.CategoryFile, "Admin song upload initiated: title=%s, artist=%s", title, artistName)

	// Validate required fields
	if title == "" || artistName == "" {
		logger.Warning(logger.CategoryFile, "Song upload failed: missing required metadata")
		return nil, errors.New("missing metadata. Title and artist are required")
	}

	// Validate file type
	ext := filepath.Ext(file.Filename)
	if ext != ".mp4" && ext != ".wav" && ext != ".mp3" {
		logger.Warning(logger.CategoryFile, "Song upload failed: invalid file format %s", ext)
		return nil, errors.New("invalid format. Only MP4, WAV, and MP3 allowed")
	}

	// Validate file size (50MB)
	if file.Size > 50*1024*1024 {
		logger.Warning(logger.CategoryFile, "Song upload failed: file too large (%d bytes)", file.Size)
		return nil, errors.New("file too large. Maximum size is 50MB")
	}

	// Check for duplicate song
	var existingID int
	err := s.db.QueryRow(`
		SELECT s.id FROM songs s
		JOIN artists a ON s.artist_id = a.id
		WHERE LOWER(s.title) = LOWER(?) AND LOWER(a.name) = LOWER(?)
	`, title, artistName).Scan(&existingID)

	if err == nil {
		logger.Warning(logger.CategoryDB, "Duplicate song detected: song_id=%d, title=%s", existingID, title)
		return nil, fmt.Errorf("duplicate song detected. Song ID %d already exists with this title and artist", existingID)
	}

	// Get or create artist
	artistID, err := s.getOrCreateArtist(artistName)
	if err != nil {
		logger.Error(logger.CategoryDB, "Failed to get or create artist", err)
		return nil, err
	}

	// Get or create album if provided
	var albumID *int
	if albumTitle != "" {
		aid, err := s.getOrCreateAlbum(albumTitle, artistID)
		if err == nil {
			albumID = &aid
		} else {
			logger.Warning(logger.CategoryDB, "Failed to create album: %s", albumTitle)
		}
	}

	// Create storage directory
	songDir := filepath.Join(s.storagePath, "media", "songs")
	if err := os.MkdirAll(songDir, 0755); err != nil {
		logger.Error(logger.CategoryFile, "Failed to create storage directory", err)
		return nil, err
	}

	// Generate unique filename
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	filePath := filepath.Join(songDir, filename)

	// Save file
	src, err := file.Open()
	if err != nil {
		logger.Error(logger.CategoryFile, "Failed to open uploaded file", err)
		return nil, err
	}
	defer src.Close()

	dst, err := os.Create(filePath)
	if err != nil {
		logger.Error(logger.CategoryFile, "Failed to create destination file", err)
		return nil, err
	}
	defer dst.Close()

	bytesWritten, err := io.Copy(dst, src)
	if err != nil {
		logger.Error(logger.CategoryFile, "Failed to write file to disk", err)
		return nil, err
	}

	logger.Info(logger.CategoryFile, "File saved successfully: %d bytes written to %s", bytesWritten, filename)

	// Store song record
	relativePath := filepath.Join("media", "songs", filename)
	var catID *int
	if categoryID > 0 {
		catID = &categoryID
	}

	result, err := s.db.Exec(`
		INSERT INTO songs (title, artist_id, album_id, category_id, duration_seconds, file_path, format)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, title, artistID, albumID, catID, durationSeconds, relativePath, ext[1:])

	if err != nil {
		logger.Error(logger.CategoryDB, "Failed to insert song record", err)
		// Clean up uploaded file
		os.Remove(filePath)
		return nil, err
	}

	songID, _ := result.LastInsertId()

	// Update FTS index
	s.updateFTSIndex(int(songID), title, artistName, albumTitle, categoryID)

	song := &models.Song{
		ID:              int(songID),
		Title:           title,
		ArtistID:        artistID,
		AlbumID:         albumID,
		CategoryID:      catID,
		DurationSeconds: durationSeconds,
		FilePath:        relativePath,
		Format:          ext[1:],
	}

	logger.Info(logger.CategoryDB, "Song uploaded successfully: song_id=%d, title=%s, artist=%s", songID, title, artistName)
	return song, nil
}

func (s *AdminService) getOrCreateArtist(name string) (int, error) {
	var artistID int
	err := s.db.QueryRow(`SELECT id FROM artists WHERE LOWER(name) = LOWER(?)`, name).Scan(&artistID)
	if err == nil {
		logger.Info(logger.CategoryDB, "Found existing artist: %s (id=%d)", name, artistID)
		return artistID, nil
	}

	result, err := s.db.Exec(`INSERT INTO artists (name) VALUES (?)`, name)
	if err != nil {
		logger.Error(logger.CategoryDB, "Failed to create artist", err)
		return 0, err
	}

	id, _ := result.LastInsertId()
	logger.Info(logger.CategoryDB, "Created new artist: %s (id=%d)", name, int(id))
	return int(id), nil
}

func (s *AdminService) getOrCreateAlbum(title string, artistID int) (int, error) {
	var albumID int
	err := s.db.QueryRow(`
		SELECT id FROM albums WHERE LOWER(title) = LOWER(?) AND artist_id = ?
	`, title, artistID).Scan(&albumID)
	if err == nil {
		logger.Info(logger.CategoryDB, "Found existing album: %s (id=%d)", title, albumID)
		return albumID, nil
	}

	result, err := s.db.Exec(`INSERT INTO albums (title, artist_id) VALUES (?, ?)`, title, artistID)
	if err != nil {
		logger.Error(logger.CategoryDB, "Failed to create album", err)
		return 0, err
	}

	id, _ := result.LastInsertId()
	logger.Info(logger.CategoryDB, "Created new album: %s (id=%d)", title, int(id))
	return int(id), nil
}

func (s *AdminService) updateFTSIndex(songID int, title, artistName, albumTitle string, categoryID int) {
	var categoryName string
	if categoryID > 0 {
		err := s.db.QueryRow(`SELECT name FROM categories WHERE id = ?`, categoryID).Scan(&categoryName)
		if err != nil {
			logger.Warning(logger.CategoryDB, "Failed to get category name for FTS index")
		}
	}

	_, err := s.db.Exec(`
		INSERT INTO songs_fts (song_id, title, artist_name, album_title, category_name)
		VALUES (?, ?, ?, ?, ?)
	`, songID, title, artistName, albumTitle, categoryName)
	
	if err != nil {
		logger.Warning(logger.CategoryDB, "Failed to update FTS index for song_id=%d", songID)
	} else {
		logger.Info(logger.CategoryDB, "FTS index updated for song_id=%d", songID)
	}
}

// DeleteSong removes a song from the catalog
func (s *AdminService) DeleteSong(songID int) error {
	logger.Info(logger.CategoryDB, "Attempting to delete song: song_id=%d", songID)

	// Get file path and title before deleting
	var filePath, title string
	err := s.db.QueryRow(`SELECT file_path, title FROM songs WHERE id = ?`, songID).Scan(&filePath, &title)
	if err != nil {
		if err == sql.ErrNoRows {
			logger.Warning(logger.CategoryDB, "Delete failed: song not found (song_id=%d)", songID)
			return errors.New("song not found")
		}
		logger.Error(logger.CategoryDB, "Database error during song deletion", err)
		return errors.New("song not found")
	}

	// Delete from database (cascades to playlist_songs)
	_, err = s.db.Exec(`DELETE FROM songs WHERE id = ?`, songID)
	if err != nil {
		logger.Error(logger.CategoryDB, "Failed to delete song from database", err)
		return err
	}

	// Delete file
	fullPath := filepath.Join(s.storagePath, filePath)
	err = os.Remove(fullPath)
	if err != nil {
		logger.Warning(logger.CategoryFile, "Failed to delete song file: %s", fullPath)
	} else {
		logger.Info(logger.CategoryFile, "Song file deleted: %s", fullPath)
	}

	// Delete from FTS
	_, err = s.db.Exec(`DELETE FROM songs_fts WHERE song_id = ?`, songID)
	if err != nil {
		logger.Warning(logger.CategoryDB, "Failed to delete from FTS index")
	}

	logger.Info(logger.CategoryDB, "Song deleted successfully: song_id=%d, title=%s", songID, title)
	return nil
}

// GetAllUsers retrieves all users (admin view)
func (s *AdminService) GetAllUsers() ([]models.User, error) {
	logger.Info(logger.CategoryDB, "Retrieving all users (admin view)")

	rows, err := s.db.Query(`
		SELECT id, username, email, is_admin, profile_image_path, created_at, last_login
		FROM users
		ORDER BY created_at DESC
	`)
	if err != nil {
		logger.Error(logger.CategoryDB, "Failed to retrieve users", err)
		return nil, err
	}
	defer rows.Close()

	var users []models.User
	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.ID, &user.Username, &user.Email, &user.IsAdmin,
			&user.ProfileImagePath, &user.CreatedAt, &user.LastLogin,
		)
		if err != nil {
			logger.Warning(logger.CategoryDB, "Failed to scan user row")
			continue
		}
		users = append(users, user)
	}

	logger.Info(logger.CategoryDB, "Retrieved %d users", len(users))
	return users, nil
}

// GetAllSongs retrieves all songs (admin view)
func (s *AdminService) GetAllSongs(limit, offset int) ([]models.Song, error) {
	logger.Info(logger.CategoryDB, "Retrieving all songs (admin view): limit=%d, offset=%d", limit, offset)

	rows, err := s.db.Query(`
		SELECT s.id, s.title, s.artist_id, s.album_id, s.category_id,
			   s.duration_seconds, s.file_path, s.format, s.uploaded_by_user_id,
			   s.created_at, a.name, al.title, c.name
		FROM songs s
		LEFT JOIN artists a ON s.artist_id = a.id
		LEFT JOIN albums al ON s.album_id = al.id
		LEFT JOIN categories c ON s.category_id = c.id
		ORDER BY s.created_at DESC
		LIMIT ? OFFSET ?
	`, limit, offset)

	if err != nil {
		logger.Error(logger.CategoryDB, "Failed to retrieve songs", err)
		return nil, err
	}
	defer rows.Close()

	var songs []models.Song
	for rows.Next() {
		var song models.Song
		var artistName, albumTitle, categoryName sql.NullString

		err := rows.Scan(
			&song.ID, &song.Title, &song.ArtistID, &song.AlbumID, &song.CategoryID,
			&song.DurationSeconds, &song.FilePath, &song.Format, &song.UploadedByUserID,
			&song.CreatedAt, &artistName, &albumTitle, &categoryName,
		)
		if err != nil {
			logger.Warning(logger.CategoryDB, "Failed to scan song row")
			continue
		}

		if artistName.Valid {
			song.Artist = &models.Artist{Name: artistName.String}
		}
		if albumTitle.Valid {
			song.Album = &models.Album{Title: albumTitle.String}
		}
		if categoryName.Valid {
			song.Category = &models.Category{Name: categoryName.String}
			}

		songs = append(songs, song)
	}

	return songs, nil
}