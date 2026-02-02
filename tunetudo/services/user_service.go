package services

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"tunetudo/models"

	"github.com/google/uuid"
)

type UserService struct {
	db          *sql.DB
	storagePath string
}

func NewUserService(db *sql.DB, storagePath string) *UserService {
	return &UserService{
		db:          db,
		storagePath: storagePath,
	}
}

// UploadProfileImage uploads a user's profile picture
func (s *UserService) UploadProfileImage(userID int, file *multipart.FileHeader) error {
	// Validate file size (5MB limit)
	if file.Size > 5*1024*1024 {
		return errors.New("file too large. Maximum size is 5MB")
	}

	// Validate file type
	ext := filepath.Ext(file.Filename)
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		return errors.New("unsupported file type. Only JPG and PNG allowed")
	}

	// Create storage directory
	profileDir := filepath.Join(s.storagePath, "images", "profiles", fmt.Sprintf("%d", userID))
	if err := os.MkdirAll(profileDir, 0755); err != nil {
		return err
	}

	// Generate unique filename
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	filePath := filepath.Join(profileDir, filename)

	// Save file
	src, err := file.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return err
	}

	// Update database
	relativePath := filepath.Join("images", "profiles", fmt.Sprintf("%d", userID), filename)
	_, err = s.db.Exec(
		`UPDATE users SET profile_image_path = ? WHERE id = ?`,
		relativePath, userID,
	)

	return err
}

// UploadSong uploads a song to user's personal library
func (s *UserService) UploadSong(userID int, file *multipart.FileHeader) (*models.Upload, error) {
	// Validate file size (50MB limit)
	if file.Size > 50*1024*1024 {
		return nil, errors.New("file too large. Maximum size is 50MB")
	}

	// Validate file type
	ext := filepath.Ext(file.Filename)
	if ext != ".mp4" && ext != ".wav" && ext != ".mp3" {
		return nil, errors.New("unsupported file format. Only MP4, WAV, and MP3 allowed")
	}

	// Create storage directory
	uploadDir := filepath.Join(s.storagePath, "media", "uploads", fmt.Sprintf("%d", userID))
	if err := os.MkdirAll(uploadDir, 0755); err != nil {
		return nil, err
	}

	// Generate unique filename
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	filePath := filepath.Join(uploadDir, filename)

	// Save file
	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	dst, err := os.Create(filePath)
	if err != nil {
		return nil, err
	}
	defer dst.Close()

	if _, err = io.Copy(dst, src); err != nil {
		return nil, err
	}

	// Store upload record
	relativePath := filepath.Join("media", "uploads", fmt.Sprintf("%d", userID), filename)
	result, err := s.db.Exec(
		`INSERT INTO uploads (user_id, original_filename, stored_path, file_size_bytes) 
		VALUES (?, ?, ?, ?)`,
		userID, file.Filename, relativePath, file.Size,
	)
	if err != nil {
		return nil, err
	}

	uploadID, _ := result.LastInsertId()

	// Create a song entry for this upload
	// Extract title from filename (remove extension)
	title := file.Filename[:len(file.Filename)-len(ext)]
	
	// Get or create "Unknown Artist"
	var artistID int
	err = s.db.QueryRow(`SELECT id FROM artists WHERE name = ?`, "Unknown Artist").Scan(&artistID)
	if err != nil {
		result, _ := s.db.Exec(`INSERT INTO artists (name, description) VALUES (?, ?)`, 
			"Unknown Artist", "User uploaded content")
		aid, _ := result.LastInsertId()
		artistID = int(aid)
	}

	_, err = s.db.Exec(
		`INSERT INTO songs (title, artist_id, file_path, format, uploaded_by_user_id, duration_seconds) 
		VALUES (?, ?, ?, ?, ?, ?)`,
		title, artistID, relativePath, ext[1:], userID, 0,
	)

	upload := &models.Upload{
		ID:               int(uploadID),
		UserID:           userID,
		OriginalFilename: file.Filename,
		StoredPath:       relativePath,
		FileSizeBytes:    file.Size,
	}

	return upload, nil
}

// GetUserUploads retrieves all uploads for a user
func (s *UserService) GetUserUploads(userID int) ([]models.Song, error) {
	rows, err := s.db.Query(`
		SELECT s.id, s.title, s.artist_id, s.duration_seconds, s.file_path, 
			   s.format, s.created_at, a.name as artist_name
		FROM songs s
		LEFT JOIN artists a ON s.artist_id = a.id
		WHERE s.uploaded_by_user_id = ?
		ORDER BY s.created_at DESC
	`, userID)

	if err != nil {
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
			continue
		}

		if artistName.Valid {
			song.Artist = &models.Artist{Name: artistName.String}
		}

		songs = append(songs, song)
	}

	return songs, nil
}

// GetProfile retrieves user profile information
func (s *UserService) GetProfile(userID int) (*models.User, error) {
	var user models.User
	err := s.db.QueryRow(`
		SELECT id, username, email, is_admin, profile_image_path, created_at, last_login
		FROM users WHERE id = ?
	`, userID).Scan(
		&user.ID, &user.Username, &user.Email, &user.IsAdmin,
		&user.ProfileImagePath, &user.CreatedAt, &user.LastLogin,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}