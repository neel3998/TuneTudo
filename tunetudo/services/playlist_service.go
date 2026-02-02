package services

import (
	"database/sql"
	"errors"
	"tunetudo/models"
)

type PlaylistService struct {
	db *sql.DB
}

func NewPlaylistService(db *sql.DB) *PlaylistService {
	return &PlaylistService{db: db}
}

// CreatePlaylist creates a new playlist for a user
func (s *PlaylistService) CreatePlaylist(userID int, req models.CreatePlaylistRequest) (*models.Playlist, error) {
	if req.Name == "" {
		return nil, errors.New("enter valid playlist name")
	}

	result, err := s.db.Exec(
		`INSERT INTO playlists (user_id, name, description) VALUES (?, ?, ?)`,
		userID, req.Name, req.Description,
	)
	if err != nil {
		return nil, errors.New("Playlist already exists")
	}

	id, _ := result.LastInsertId()

	playlist := &models.Playlist{
		ID:          int(id),
		UserID:      userID,
		Name:        req.Name,
		Description: req.Description,
	}

	return playlist, nil
}

// GetUserPlaylists retrieves all playlists for a user
func (s *PlaylistService) GetUserPlaylists(userID int) ([]models.Playlist, error) {
	rows, err := s.db.Query(`
		SELECT p.id, p.user_id, p.name, p.description, p.created_at,
			   COUNT(ps.id) as song_count
		FROM playlists p
		LEFT JOIN playlist_songs ps ON p.id = ps.playlist_id
		WHERE p.user_id = ?
		GROUP BY p.id
		ORDER BY p.created_at DESC
	`, userID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var playlists []models.Playlist
	for rows.Next() {
		var playlist models.Playlist
		err := rows.Scan(
			&playlist.ID, &playlist.UserID, &playlist.Name,
			&playlist.Description, &playlist.CreatedAt, &playlist.SongCount,
		)
		if err != nil {
			continue
		}
		playlists = append(playlists, playlist)
	}

	return playlists, nil
}

// GetPlaylistByID retrieves a specific playlist
func (s *PlaylistService) GetPlaylistByID(playlistID int, userID int) (*models.Playlist, error) {
	var playlist models.Playlist
	err := s.db.QueryRow(`
		SELECT id, user_id, name, description, created_at
		FROM playlists
		WHERE id = ? AND user_id = ?
	`, playlistID, userID).Scan(
		&playlist.ID, &playlist.UserID, &playlist.Name,
		&playlist.Description, &playlist.CreatedAt,
	)

	if err != nil {
		return nil, errors.New("no playlist found")
	}

	return &playlist, nil
}

// GetPlaylistSongs retrieves all songs in a playlist
func (s *PlaylistService) GetPlaylistSongs(playlistID int) ([]models.PlaylistSong, error) {
	rows, err := s.db.Query(`
		SELECT ps.id, ps.playlist_id, ps.song_id, ps.queue_number, ps.added_at,
			   s.title, s.artist_id, s.album_id, s.duration_seconds, s.file_path, s.format,
			   a.name as artist_name
		FROM playlist_songs ps
		JOIN songs s ON ps.song_id = s.id
		LEFT JOIN artists a ON s.artist_id = a.id
		WHERE ps.playlist_id = ?
		ORDER BY ps.queue_number, ps.added_at
	`, playlistID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var playlistSongs []models.PlaylistSong
	for rows.Next() {
		var ps models.PlaylistSong
		var song models.Song
		var artistName sql.NullString

		err := rows.Scan(
			&ps.ID, &ps.PlaylistID, &ps.SongID, &ps.QueueNumber, &ps.AddedAt,
			&song.Title, &song.ArtistID, &song.AlbumID, &song.DurationSeconds,
			&song.FilePath, &song.Format, &artistName,
		)
		if err != nil {
			continue
		}

		song.ID = ps.SongID
		if artistName.Valid {
			song.Artist = &models.Artist{Name: artistName.String}
		}
		ps.Song = &song

		playlistSongs = append(playlistSongs, ps)
	}

	return playlistSongs, nil
}

// AddSong adds a song to a playlist
func (s *PlaylistService) AddSong(playlistID, songID, userID int) error {
	// Verify playlist belongs to user
	var ownerID int
	err := s.db.QueryRow(`SELECT user_id FROM playlists WHERE id = ?`, playlistID).Scan(&ownerID)
	if err != nil {
		return errors.New("no playlist found")
	}
	if ownerID != userID {
		return errors.New("unauthorized")
	}

	// Check if song already in playlist
	var exists int
	err = s.db.QueryRow(
		`SELECT COUNT(*) FROM playlist_songs WHERE playlist_id = ? AND song_id = ?`,
		playlistID, songID,
	).Scan(&exists)

	if err == nil && exists > 0 {
		return errors.New("Song already in the playlist")
	}

	// Get next queue number
	var maxQueue sql.NullInt64
	s.db.QueryRow(
		`SELECT MAX(queue_number) FROM playlist_songs WHERE playlist_id = ?`,
		playlistID,
	).Scan(&maxQueue)

	queueNumber := 0
	if maxQueue.Valid {
		queueNumber = int(maxQueue.Int64) + 1
	}

	// Add song to playlist
	_, err = s.db.Exec(
		`INSERT INTO playlist_songs (playlist_id, song_id, queue_number) VALUES (?, ?, ?)`,
		playlistID, songID, queueNumber,
	)

	return err
}

// RemoveSong removes a song from a playlist
func (s *PlaylistService) RemoveSong(playlistID, songID, userID int) error {
	// Verify playlist belongs to user
	var ownerID int
	err := s.db.QueryRow(`SELECT user_id FROM playlists WHERE id = ?`, playlistID).Scan(&ownerID)
	if err != nil {
		return errors.New("no playlist found")
	}
	if ownerID != userID {
		return errors.New("unauthorized")
	}

	result, err := s.db.Exec(
		`DELETE FROM playlist_songs WHERE playlist_id = ? AND song_id = ?`,
		playlistID, songID,
	)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("song not found in playlist")
	}

	return nil
}

// DeletePlaylist deletes a playlist
func (s *PlaylistService) DeletePlaylist(playlistID, userID int) error {
	result, err := s.db.Exec(
		`DELETE FROM playlists WHERE id = ? AND user_id = ?`,
		playlistID, userID,
	)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return errors.New("no playlist found")
	}

	return nil
}