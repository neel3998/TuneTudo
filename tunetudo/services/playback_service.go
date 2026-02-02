package services

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"tunetudo/logger"
	"tunetudo/models"
)

type PlaybackService struct {
	db          *sql.DB
	storagePath string
}

func NewPlaybackService(db *sql.DB, storagePath string) *PlaybackService {
	return &PlaybackService{
		db:          db,
		storagePath: storagePath,
	}
}

// GetSongByID retrieves song metadata
func (s *PlaybackService) GetSongByID(songID int) (*models.Song, error) {
	var song models.Song
	var artistName, albumTitle, categoryName sql.NullString

	err := s.db.QueryRow(`
		SELECT s.id, s.title, s.artist_id, s.album_id, s.category_id,
			   s.duration_seconds, s.file_path, s.format, s.uploaded_by_user_id,
			   s.created_at, a.name, al.title, c.name
		FROM songs s
		LEFT JOIN artists a ON s.artist_id = a.id
		LEFT JOIN albums al ON s.album_id = al.id
		LEFT JOIN categories c ON s.category_id = c.id
		WHERE s.id = ?
	`, songID).Scan(
		&song.ID, &song.Title, &song.ArtistID, &song.AlbumID, &song.CategoryID,
		&song.DurationSeconds, &song.FilePath, &song.Format, &song.UploadedByUserID,
		&song.CreatedAt, &artistName, &albumTitle, &categoryName,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			// Don't reveal internal details to user
			return nil, errors.New("track not found")
		}
		// Log internal error without exposing to user
		logger.Error(logger.CategoryDB, "Failed to retrieve song", err)
		return nil, errors.New("track not found")
	}

	if artistName.Valid {
		song.Artist = &models.Artist{ID: song.ArtistID, Name: artistName.String}
	}
	if albumTitle.Valid && song.AlbumID != nil {
		song.Album = &models.Album{ID: *song.AlbumID, Title: albumTitle.String}
	}
	if categoryName.Valid && song.CategoryID != nil {
		song.Category = &models.Category{ID: *song.CategoryID, Name: categoryName.String}
	}

	return &song, nil
}

// AuthorizeStream validates that a song can be streamed
func (s *PlaybackService) AuthorizeStream(songID int) (string, error) {
	var filePath string
	err := s.db.QueryRow(`SELECT file_path FROM songs WHERE id = ?`, songID).Scan(&filePath)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", errors.New("track not found")
		}
		// Log error without exposing details
		logger.Error(logger.CategoryDB, "Failed to authorize stream", err)
		return "", errors.New("track not found")
	}

	// Check if file exists
	fullPath := filepath.Join(s.storagePath, filePath)
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		// Log the issue for debugging but don't expose file paths to user
		logger.Warning(logger.CategoryFile, "Song file not found on disk: song_id=%d", songID)
		return "", errors.New("track not found")
	}

	// Log file access
	logger.Info(logger.CategoryFile, "Song stream authorized: song_id=%d", songID)

	return fullPath, nil
}

// GetRecentSongs retrieves recently added songs (excluding personal uploads)
func (s *PlaybackService) GetRecentSongs(limit int) ([]models.Song, error) {
	if limit <= 0 {
		limit = 20
	}

	rows, err := s.db.Query(`
		SELECT s.id, s.title, s.artist_id, s.duration_seconds, s.file_path,
			   s.format, s.created_at, a.name as artist_name
		FROM songs s
		LEFT JOIN artists a ON s.artist_id = a.id
		WHERE s.uploaded_by_user_id IS NULL
		ORDER BY s.created_at DESC
		LIMIT ?
	`, limit)

	if err != nil {
		logger.Error(logger.CategoryDB, "Failed to retrieve recent songs", err)
		return nil, err
	}
	defer rows.Close()

	var songs []models.Song
	for rows.Next() {
		var song models.Song
		var artistName sql.NullString

		err := rows.Scan(
			&song.ID, &song.Title, &song.ArtistID, &song.DurationSeconds,
			&song.FilePath, &song.Format, &song.CreatedAt, &artistName,
		)
		if err != nil {
			logger.Warning(logger.CategoryDB, "Failed to scan song row")
			continue
		}

		if artistName.Valid {
			song.Artist = &models.Artist{Name: artistName.String}
		}

		songs = append(songs, song)
	}

	logger.Info(logger.CategoryAPI, "Retrieved %d recent songs", len(songs))

	return songs, nil
}