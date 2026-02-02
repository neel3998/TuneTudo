package services

import (
	"database/sql"
	"strings"
	"tunetudo/models"
)

type SearchService struct {
	db *sql.DB
}

func NewSearchService(db *sql.DB) *SearchService {
	return &SearchService{db: db}
}

// FullTextSearch performs comprehensive search across songs, artists, and albums
func (s *SearchService) FullTextSearch(query string) (*models.SearchResult, error) {
	result := &models.SearchResult{
		Songs:     []models.Song{},
		Artists:   []models.Artist{},
		Albums:    []models.Album{},
		Playlists: []models.Playlist{},
	}

	if query == "" {
		return result, nil
	}

	searchTerm := "%" + strings.ToLower(query) + "%"

	// Search songs
	songs, err := s.searchSongs(searchTerm)
	if err == nil {
		result.Songs = songs
	}

	// Search artists
	artists, err := s.searchArtists(searchTerm)
	if err == nil {
		result.Artists = artists
	}

	// Search albums
	albums, err := s.searchAlbums(searchTerm)
	if err == nil {
		result.Albums = albums
	}

	return result, nil
}

func (s *SearchService) searchSongs(searchTerm string) ([]models.Song, error) {
	rows, err := s.db.Query(`
		SELECT s.id, s.title, s.artist_id, s.album_id, s.category_id, 
			   s.duration_seconds, s.file_path, s.format, s.uploaded_by_user_id, s.created_at,
			   a.name as artist_name, al.title as album_title, c.name as category_name
		FROM songs s
		LEFT JOIN artists a ON s.artist_id = a.id
		LEFT JOIN albums al ON s.album_id = al.id
		LEFT JOIN categories c ON s.category_id = c.id
		WHERE (LOWER(s.title) LIKE ? OR LOWER(a.name) LIKE ? OR LOWER(al.title) LIKE ?)
		AND s.uploaded_by_user_id IS NULL
		LIMIT 50
	`, searchTerm, searchTerm, searchTerm)

	if err != nil {
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

func (s *SearchService) searchArtists(searchTerm string) ([]models.Artist, error) {
	rows, err := s.db.Query(`
		SELECT id, name, description, created_at
		FROM artists
		WHERE LOWER(name) LIKE ?
		LIMIT 20
	`, searchTerm)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var artists []models.Artist
	for rows.Next() {
		var artist models.Artist
		err := rows.Scan(&artist.ID, &artist.Name, &artist.Description, &artist.CreatedAt)
		if err != nil {
			continue
		}
		artists = append(artists, artist)
	}

	return artists, nil
}

func (s *SearchService) searchAlbums(searchTerm string) ([]models.Album, error) {
	rows, err := s.db.Query(`
		SELECT a.id, a.title, a.artist_id, a.cover_image_path, a.release_date,
			   ar.name as artist_name
		FROM albums a
		LEFT JOIN artists ar ON a.artist_id = ar.id
		WHERE LOWER(a.title) LIKE ? OR LOWER(ar.name) LIKE ?
		LIMIT 20
	`, searchTerm, searchTerm)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var albums []models.Album
	for rows.Next() {
		var album models.Album
		var artistName sql.NullString

		err := rows.Scan(
			&album.ID, &album.Title, &album.ArtistID, &album.CoverImagePath,
			&album.ReleaseDate, &artistName,
		)
		if err != nil {
			continue
		}

		if artistName.Valid {
			album.Artist = &models.Artist{Name: artistName.String}
		}

		albums = append(albums, album)
	}

	return albums, nil
}

// GetSongsByCategory retrieves songs filtered by category
func (s *SearchService) GetSongsByCategory(categoryID int) ([]models.Song, error) {
	rows, err := s.db.Query(`
		SELECT s.id, s.title, s.artist_id, s.album_id, s.category_id,
			   s.duration_seconds, s.file_path, s.format, s.uploaded_by_user_id, s.created_at,
			   a.name as artist_name
		FROM songs s
		LEFT JOIN artists a ON s.artist_id = a.id
		WHERE s.category_id = ? AND s.uploaded_by_user_id IS NULL
		ORDER BY s.created_at DESC
		LIMIT 100
	`, categoryID)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var songs []models.Song
	for rows.Next() {
		var song models.Song
		var artistName sql.NullString

		err := rows.Scan(
			&song.ID, &song.Title, &song.ArtistID, &song.AlbumID, &song.CategoryID,
			&song.DurationSeconds, &song.FilePath, &song.Format, &song.UploadedByUserID,
			&song.CreatedAt, &artistName,
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

// GetAllCategories retrieves all categories
func (s *SearchService) GetAllCategories() ([]models.Category, error) {
	rows, err := s.db.Query(`SELECT id, name, description FROM categories ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []models.Category
	for rows.Next() {
		var cat models.Category
		err := rows.Scan(&cat.ID, &cat.Name, &cat.Description)
		if err != nil {
			continue
		}
		categories = append(categories, cat)
	}

	return categories, nil
}