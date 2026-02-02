package services

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

// setupTestDB creates a temporary test database
func setupTestDB(t *testing.T) *sql.DB {
	// Create temp database file
	dbPath := "./test_" + t.Name() + ".db"
	
	// Remove if exists
	os.Remove(dbPath)
	
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	// Create tables
	createTables(t, db)
	
	// Register cleanup
	t.Cleanup(func() {
		db.Close()
		os.Remove(dbPath)
	})

	return db
}

func createTables(t *testing.T, db *sql.DB) {
	tables := []string{
		`CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT UNIQUE NOT NULL,
			email TEXT UNIQUE NOT NULL,
			password_hash TEXT NOT NULL,
			is_admin INTEGER DEFAULT 0,
			profile_image_path TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			last_login DATETIME
		)`,
		`CREATE TABLE artists (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL,
			description TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE albums (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			title TEXT NOT NULL,
			artist_id INTEGER,
			cover_image_path TEXT,
			release_date DATE,
			FOREIGN KEY(artist_id) REFERENCES artists(id)
		)`,
		`CREATE TABLE categories (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			description TEXT
		)`,
		`CREATE TABLE songs (
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
			FOREIGN KEY(artist_id) REFERENCES artists(id),
			FOREIGN KEY(album_id) REFERENCES albums(id),
			FOREIGN KEY(category_id) REFERENCES categories(id),
			FOREIGN KEY(uploaded_by_user_id) REFERENCES users(id)
		)`,
		`CREATE TABLE playlists (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			name TEXT NOT NULL,
			description TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user_id, name),
			FOREIGN KEY(user_id) REFERENCES users(id)
		)`,
		`CREATE TABLE playlist_songs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			playlist_id INTEGER NOT NULL,
			song_id INTEGER NOT NULL,
			queue_number INTEGER DEFAULT 0,
			added_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(playlist_id) REFERENCES playlists(id),
			FOREIGN KEY(song_id) REFERENCES songs(id),
			UNIQUE(playlist_id, song_id)
		)`,
		`CREATE TABLE uploads (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			original_filename TEXT,
			stored_path TEXT,
			file_size_bytes INTEGER,
			error_message TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY(user_id) REFERENCES users(id)
		)`,
	}

	for _, table := range tables {
		if _, err := db.Exec(table); err != nil {
			t.Fatalf("Failed to create table: %v", err)
		}
	}

	// Seed test categories
	categories := []string{"Pop", "Rock", "Jazz", "Classical"}
	for _, cat := range categories {
		db.Exec("INSERT INTO categories (name) VALUES (?)", cat)
	}
}

// seedTestData inserts sample data for testing
func seedTestData(t *testing.T, db *sql.DB) {
	// Insert test artist
	result, err := db.Exec("INSERT INTO artists (name, description) VALUES (?, ?)", 
		"Test Artist", "A test artist")
	if err != nil {
		t.Fatalf("Failed to insert test artist: %v", err)
	}
	artistID, _ := result.LastInsertId()

	// Insert test album
	result, err = db.Exec("INSERT INTO albums (title, artist_id) VALUES (?, ?)", 
		"Test Album", artistID)
	if err != nil {
		t.Fatalf("Failed to insert test album: %v", err)
	}
	albumID, _ := result.LastInsertId()

	// Insert test songs
	for i := 1; i <= 3; i++ {
		db.Exec(`INSERT INTO songs (title, artist_id, album_id, category_id, duration_seconds, file_path, format)
			VALUES (?, ?, ?, ?, ?, ?, ?)`,
			"Test Song "+string(rune(i+'0')), artistID, albumID, 1, 180, "/test/song.mp3", "mp3")
	}
}