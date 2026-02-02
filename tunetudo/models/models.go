package models

import (
	"time"
)

// User represents a user account
type User struct {
	ID               int       `json:"id"`
	Username         string    `json:"username"`
	Email            string    `json:"email"`
	PasswordHash     string    `json:"-"`
	IsAdmin          bool      `json:"is_admin"`
	ProfileImagePath *string   `json:"profile_image_path"`
	CreatedAt        time.Time `json:"created_at"`
	LastLogin        *time.Time `json:"last_login"`
}

// Artist represents a music artist
type Artist struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// Album represents a music album
type Album struct {
	ID             int       `json:"id"`
	Title          string    `json:"title"`
	ArtistID       int       `json:"artist_id"`
	CoverImagePath *string   `json:"cover_image_path"`
	ReleaseDate    *string   `json:"release_date"`
	Artist         *Artist   `json:"artist,omitempty"`
}

// Category represents a genre/category
type Category struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

// Song represents a music track
type Song struct {
	ID               int       `json:"id"`
	Title            string    `json:"title"`
	ArtistID         int       `json:"artist_id"`
	AlbumID          *int      `json:"album_id"`
	CategoryID       *int      `json:"category_id"`
	DurationSeconds  int       `json:"duration_seconds"`
	FilePath         string    `json:"file_path"`
	Format           string    `json:"format"`
	UploadedByUserID *int      `json:"uploaded_by_user_id"`
	CreatedAt        time.Time `json:"created_at"`
	Artist           *Artist   `json:"artist,omitempty"`
	Album            *Album    `json:"album,omitempty"`
	Category         *Category `json:"category,omitempty"`
}

// Playlist represents a user's playlist
type Playlist struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	Name        string    `json:"name"`
	Description *string   `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	SongCount   int       `json:"song_count,omitempty"`
}

// PlaylistSong represents a song in a playlist
type PlaylistSong struct {
	ID          int       `json:"id"`
	PlaylistID  int       `json:"playlist_id"`
	SongID      int       `json:"song_id"`
	QueueNumber int       `json:"queue_number"`
	AddedAt     time.Time `json:"added_at"`
	Song        *Song     `json:"song,omitempty"`
}

// Upload represents a user file upload
type Upload struct {
	ID               int       `json:"id"`
	UserID           int       `json:"user_id"`
	OriginalFilename string    `json:"original_filename"`
	StoredPath       string    `json:"stored_path"`
	FileSizeBytes    int64     `json:"file_size_bytes"`
	ErrorMessage     *string   `json:"error_message"`
	CreatedAt        time.Time `json:"created_at"`
}

// SearchResult represents combined search results
type SearchResult struct {
	Songs     []Song     `json:"songs"`
	Artists   []Artist   `json:"artists"`
	Albums    []Album    `json:"albums"`
	Playlists []Playlist `json:"playlists"`
}

// LoginRequest represents login credentials
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// RegisterRequest represents registration data
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// CreatePlaylistRequest represents playlist creation data
type CreatePlaylistRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

// PasswordResetRequest represents password reset request
type PasswordResetRequest struct {
	Email string `json:"email"`
}

// PasswordResetConfirm represents password reset confirmation
type PasswordResetConfirm struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}