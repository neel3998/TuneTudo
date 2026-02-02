package database

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

func InitDB(dbPath string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

func RunMigrations(db *sql.DB) error {
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			is_admin INTEGER DEFAULT 0,
			profile_image_path TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			last_login DATETIME
		)`,
		
		`CREATE TABLE IF NOT EXISTS artists (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			description TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		
		`CREATE TABLE IF NOT EXISTS albums (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			artist_id INTEGER,
			cover_image_path TEXT,
			release_date DATE,
			FOREIGN KEY(artist_id) REFERENCES artists(id) ON DELETE CASCADE
		)`,
		
		`CREATE TABLE IF NOT EXISTS categories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			description TEXT
		)`,
		
		`CREATE TABLE IF NOT EXISTS songs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			artist_id INTEGER,
			album_id INTEGER,
			category_id INTEGER,
			duration_seconds INTEGER,
			file_path TEXT NOT NULL,
			format TEXT NOT NULL,
			uploaded_by_user_id INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(artist_id) REFERENCES artists(id) ON DELETE CASCADE,
			FOREIGN KEY(album_id) REFERENCES albums(id) ON DELETE SET NULL,
			FOREIGN KEY(category_id) REFERENCES categories(id) ON DELETE SET NULL,
			FOREIGN KEY(uploaded_by_user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		
		`CREATE TABLE IF NOT EXISTS playlists (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			description TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, name),
			FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		
		`CREATE TABLE IF NOT EXISTS playlist_songs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			playlist_id INTEGER NOT NULL,
			song_id INTEGER NOT NULL,
			queue_number INTEGER DEFAULT 0,
			added_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(playlist_id) REFERENCES playlists(id) ON DELETE CASCADE,
			FOREIGN KEY(song_id) REFERENCES songs(id) ON DELETE CASCADE,
			UNIQUE(playlist_id, song_id)
		)`,
		
		`CREATE TABLE IF NOT EXISTS uploads (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			original_filename TEXT,
			stored_path TEXT,
			file_size_bytes INTEGER,
			error_message TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE
		)`,
		
		// FTS5 Virtual Table for search
		// Indexes
		`CREATE INDEX IF NOT EXISTS idx_songs_artist ON songs(artist_id)`,
		`CREATE INDEX IF NOT EXISTS idx_songs_album ON songs(album_id)`,
		`CREATE INDEX IF NOT EXISTS idx_songs_category ON songs(category_id)`,
		`CREATE INDEX IF NOT EXISTS idx_playlists_user ON playlists(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_playlist_songs_playlist ON playlist_songs(playlist_id)`,
		`CREATE INDEX IF NOT EXISTS idx_uploads_user ON uploads(user_id)`,
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("migration failed: %v", err)
		}
	}

	return seedDefaultData(db)
}

func seedDefaultData(db *sql.DB) error {
	// Check if categories exist
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM categories").Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		categories := []struct {
			Name        string
			Description string
		}{
			{"Pop", "Popular music"},
			{"Rock", "Rock and roll"},
			{"Jazz", "Jazz music"},
			{"Classical", "Classical music"},
			{"Hip Hop", "Hip hop and rap"},
			{"Electronic", "Electronic music"},
			{"Country", "Country music"},
			{"R&B", "Rhythm and blues"},
			{"Love", "Romantic songs"},
			{"Workout", "Energetic workout tunes"},
			{"Chill", "Relaxing and chill music"},
			{"Party", "Upbeat party tracks"},
			{"Indie", "Independent music"},
			{"Metal", "Heavy metal music"},
			{"Folk", "Folk and acoustic"},
		}

		for _, cat := range categories {
			_, err := db.Exec("INSERT INTO categories (name, description) VALUES (?, ?)", 
				cat.Name, cat.Description)
			if err != nil {
				return err
			}
		}
	}

	return nil
}