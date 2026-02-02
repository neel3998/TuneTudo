package controllers

import (
	"strconv"
	"tunetudo/logger"
	"tunetudo/middleware"
	"tunetudo/models"
	"tunetudo/services"
	"strings"
	"github.com/gofiber/fiber/v2"
)

// AuthController handles authentication endpoints
type AuthController struct {
	authService *services.AuthService
}

func NewAuthController(authService *services.AuthService) *AuthController {
	return &AuthController{authService: authService}
}

func (ctrl *AuthController) Register(c *fiber.Ctx) error {
	var req models.RegisterRequest
	if err := c.BodyParser(&req); err != nil {
		ip := c.IP()
		logger.ValidationFailure("anonymous", ip, "request_body", "Invalid JSON format")
		// "Limit error information sent back to user"
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "invalid request data",
		})
	}

	ip := c.IP()
	user, err := ctrl.authService.RegisterUser(req, ip)
	if err != nil {
		// Error already logged in service layer
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(), // Service returns safe messages
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"error":   false,
		"message": "user registered successfully",
		"data":    user,
	})
}

func (ctrl *AuthController) Login(c *fiber.Ctx) error {
	var req models.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		ip := c.IP()
		logger.ValidationFailure("anonymous", ip, "request_body", "Invalid JSON format")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "invalid request data",
		})
	}

	ip := c.IP()
	token, user, err := ctrl.authService.LoginUser(req, ip)
	if err != nil {
		// Service returns "authorization failed" - not "no such user" or "password incorrect"
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"error":   false,
		"message": "login successful",
		"data": fiber.Map{
			"token": token,
			"user":  user,
		},
	})
}

func (ctrl *AuthController) Logout(c *fiber.Ctx) error {
	username := c.Locals("username")
	ip := c.IP()
	
	if username != nil {
		logger.Security("LOGOUT", logger.HashIdentifier(username.(string)), logger.MaskIP(ip), "User logged out")
	}
	
	return c.JSON(fiber.Map{
		"error":   false,
		"message": "logout successful",
	})
}

func (ctrl *AuthController) GetProfile(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return err
	}

	user, err := ctrl.authService.GetUserByID(userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   true,
			"message": "user not found",
		})
	}

	return c.JSON(fiber.Map{
		"error": false,
		"data":  user,
	})
}

// SearchController handles search endpoints
type SearchController struct {
	searchService *services.SearchService
}

func NewSearchController(searchService *services.SearchService) *SearchController {
	return &SearchController{searchService: searchService}
}

func (ctrl *SearchController) Search(c *fiber.Ctx) error {
	query := c.Query("q")
	if query == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "search query required",
		})
	}

	results, err := ctrl.searchService.FullTextSearch(query)
	if err != nil {
		logger.Error(logger.CategoryAPI, "Search failed", err)
		// Generic message to user
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "search temporarily unavailable",
		})
	}

	return c.JSON(fiber.Map{
		"error": false,
		"data":  results,
	})
}

func (ctrl *SearchController) GetCategories(c *fiber.Ctx) error {
	categories, err := ctrl.searchService.GetAllCategories()
	if err != nil {
		logger.Error(logger.CategoryDB, "Failed to fetch categories", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "failed to fetch categories",
		})
	}

	return c.JSON(fiber.Map{
		"error": false,
		"data":  categories,
	})
}

func (ctrl *SearchController) GetSongsByCategory(c *fiber.Ctx) error {
	categoryID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "invalid category ID",
		})
	}

	songs, err := ctrl.searchService.GetSongsByCategory(categoryID)
	if err != nil {
		logger.Error(logger.CategoryDB, "Failed to fetch songs by category", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "failed to fetch songs",
		})
	}

	if len(songs) == 0 {
		return c.JSON(fiber.Map{
			"error":   false,
			"message": "no tracks available",
			"data":    []models.Song{},
		})
	}

	return c.JSON(fiber.Map{
		"error": false,
		"data":  songs,
	})
}

// PlaylistController handles playlist endpoints
type PlaylistController struct {
	playlistService *services.PlaylistService
}

func NewPlaylistController(playlistService *services.PlaylistService) *PlaylistController {
	return &PlaylistController{playlistService: playlistService}
}

func (ctrl *PlaylistController) CreatePlaylist(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return err
	}

	var req models.CreatePlaylistRequest
	if err := c.BodyParser(&req); err != nil {
		ip := c.IP()
		username := c.Locals("username").(string)
		logger.ValidationFailure(username, ip, "request_body", "Invalid JSON format")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "invalid request data",
		})
	}

	playlist, err := ctrl.playlistService.CreatePlaylist(userID, req)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}

	logger.Info(logger.CategoryPlaylist, "Playlist created: ID=%d by user_id=%d", playlist.ID, userID)

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"error":   false,
		"message": "playlist created successfully",
		"data":    playlist,
	})
}

func (ctrl *PlaylistController) GetUserPlaylists(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return err
	}

	playlists, err := ctrl.playlistService.GetUserPlaylists(userID)
	if err != nil {
		logger.Error(logger.CategoryDB, "Failed to fetch playlists", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "failed to fetch playlists",
		})
	}

	return c.JSON(fiber.Map{
		"error": false,
		"data":  playlists,
	})
}

func (ctrl *PlaylistController) GetPlaylistDetails(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return err
	}

	playlistID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "invalid playlist ID",
		})
	}

	playlist, err := ctrl.playlistService.GetPlaylistByID(playlistID, userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}

	songs, _ := ctrl.playlistService.GetPlaylistSongs(playlistID)

	return c.JSON(fiber.Map{
		"error": false,
		"data": fiber.Map{
			"playlist": playlist,
			"songs":    songs,
		},
	})
}

func (ctrl *PlaylistController) AddSongToPlaylist(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return err
	}

	playlistID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "invalid playlist ID",
		})
	}

	var req struct {
		SongID int `json:"song_id"`
	}
	if err := c.BodyParser(&req); err != nil {
		ip := c.IP()
		username := c.Locals("username").(string)
		logger.ValidationFailure(username, ip, "request_body", "Invalid JSON format")
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "invalid request data",
		})
	}

	err = ctrl.playlistService.AddSong(playlistID, req.SongID, userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}

	logger.Info(logger.CategoryPlaylist, "Song added to playlist: playlist_id=%d song_id=%d", playlistID, req.SongID)

	return c.JSON(fiber.Map{
		"error":   false,
		"message": "song added to playlist",
	})
}

func (ctrl *PlaylistController) RemoveSongFromPlaylist(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return err
	}

	playlistID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "invalid playlist ID",
		})
	}

	songID, err := strconv.Atoi(c.Params("songId"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "invalid song ID",
		})
	}

	err = ctrl.playlistService.RemoveSong(playlistID, songID, userID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}

	logger.Info(logger.CategoryPlaylist, "Song removed from playlist: playlist_id=%d song_id=%d", playlistID, songID)

	return c.JSON(fiber.Map{
		"error":   false,
		"message": "song removed from playlist",
	})
}

func (ctrl *PlaylistController) DeletePlaylist(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return err
	}

	playlistID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "invalid playlist ID",
		})
	}

	err = ctrl.playlistService.DeletePlaylist(playlistID, userID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}

	logger.Info(logger.CategoryPlaylist, "Playlist deleted: playlist_id=%d by user_id=%d", playlistID, userID)

	return c.JSON(fiber.Map{
		"error":   false,
		"message": "playlist deleted successfully",
	})
}

// PlaybackController handles song playback endpoints
type PlaybackController struct {
	playbackService *services.PlaybackService
}

func NewPlaybackController(playbackService *services.PlaybackService) *PlaybackController {
	return &PlaybackController{playbackService: playbackService}
}

func (ctrl *PlaybackController) GetSong(c *fiber.Ctx) error {
	songID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "invalid song ID",
		})
	}

	song, err := ctrl.playbackService.GetSongByID(songID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"error": false,
		"data":  song,
	})
}

func (ctrl *PlaybackController) StreamSong(c *fiber.Ctx) error {
	songID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "invalid song ID",
		})
	}

	filePath, err := ctrl.playbackService.AuthorizeStream(songID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}

	return c.SendFile(filePath)
}

func (ctrl *PlaybackController) GetRecentSongs(c *fiber.Ctx) error {
	limit := 20
	if l := c.Query("limit"); l != "" {
		if parsedLimit, err := strconv.Atoi(l); err == nil {
			limit = parsedLimit
		}
	}

	songs, err := ctrl.playbackService.GetRecentSongs(limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "failed to fetch songs",
		})
	}

	return c.JSON(fiber.Map{
		"error": false,
		"data":  songs,
	})
}

// UserController handles user-specific endpoints
type UserController struct {
	userService *services.UserService
}

func NewUserController(userService *services.UserService) *UserController {
	return &UserController{userService: userService}
}

func (ctrl *UserController) UploadProfileImage(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return err
	}

	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "no file provided",
		})
	}

	err = ctrl.userService.UploadProfileImage(userID, file)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"error":   false,
		"message": "profile picture updated successfully",
	})
}

func (ctrl *UserController) UploadSong(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return err
	}

	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "no file provided",
		})
	}

	upload, err := ctrl.userService.UploadSong(userID, file)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"error":   false,
		"message": "track uploaded successfully",
		"data":    upload,
	})
}

func (ctrl *UserController) GetUserUploads(c *fiber.Ctx) error {
	userID, err := middleware.GetUserID(c)
	if err != nil {
		return err
	}

	songs, err := ctrl.userService.GetUserUploads(userID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "failed to fetch uploads",
		})
	}

	return c.JSON(fiber.Map{
		"error": false,
		"data":  songs,
	})
}

// AdminController handles admin endpoints
type AdminController struct {
	adminService *services.AdminService
}

func NewAdminController(adminService *services.AdminService) *AdminController {
	return &AdminController{adminService: adminService}
}

func (ctrl *AdminController) UploadSong(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "no file provided",
		})
	}

	title := c.FormValue("title")
	artistName := c.FormValue("artist")
	albumTitle := c.FormValue("album")
	categoryID, _ := strconv.Atoi(c.FormValue("category_id"))
	durationSeconds, _ := strconv.Atoi(c.FormValue("duration"))

	song, err := ctrl.adminService.UploadSong(file, title, artistName, albumTitle, categoryID, durationSeconds)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"error":   false,
		"message": "song uploaded successfully",
		"data":    song,
	})
}

func (ctrl *AdminController) DeleteSong(c *fiber.Ctx) error {
	songID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "invalid song ID",
		})
	}

	err = ctrl.adminService.DeleteSong(songID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"error":   false,
		"message": "song deleted successfully",
	})
}

func (ctrl *AdminController) GetAllUsers(c *fiber.Ctx) error {
	users, err := ctrl.adminService.GetAllUsers()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "failed to fetch users",
		})
	}

	return c.JSON(fiber.Map{
		"error": false,
		"data":  users,
	})
}

func (ctrl *AdminController) GetAllSongs(c *fiber.Ctx) error {
	limit, _ := strconv.Atoi(c.Query("limit", "50"))
	offset, _ := strconv.Atoi(c.Query("offset", "0"))

	songs, err := ctrl.adminService.GetAllSongs(limit, offset)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   true,
			"message": "failed to fetch songs",
		})
	}

	return c.JSON(fiber.Map{
		"error": false,
		"data":  songs,
	})
}

// ForgotPassword handles password reset request
func (ctrl *AuthController) ForgotPassword(c *fiber.Ctx) error {
	var req struct {
		Email string `json:"email"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid request format",
		})
	}

	// Validate email format
	if req.Email == "" || !isValidEmail(req.Email) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Valid email address is required",
		})
	}

	// Process password reset (always return success to prevent email enumeration)
	err := ctrl.authService.RequestPasswordReset(req.Email)
	if err != nil {
		logger.Error(logger.CategoryAuth, "Password reset request failed", err)
	}

	// Always return success message (security best practice)
	return c.JSON(fiber.Map{
		"error":   false,
		"message": "If your email is registered, you will receive a password reset link shortly.",
	})
}

// ValidateResetToken validates reset token
func (ctrl *AuthController) ValidateResetToken(c *fiber.Ctx) error {
	token := c.Query("token")
	
	if token == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Reset token is required",
		})
	}

	email, err := ctrl.authService.ValidateResetToken(token)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid or expired reset token",
		})
	}

	return c.JSON(fiber.Map{
		"error": false,
		"email": email,
	})
}

// ResetPassword handles password reset with token
func (ctrl *AuthController) ResetPassword(c *fiber.Ctx) error {
	var req struct {
		Token           string `json:"token"`
		NewPassword     string `json:"new_password"`
		ConfirmPassword string `json:"confirm_password"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Invalid request format",
		})
	}

	// Validate inputs
	if req.Token == "" || req.NewPassword == "" || req.ConfirmPassword == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "All fields are required",
		})
	}

	// Check password match
	if req.NewPassword != req.ConfirmPassword {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Passwords do not match",
		})
	}

	// Validate password strength
	if len(req.NewPassword) < 8 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": "Password must be at least 8 characters long",
		})
	}

	// Reset password
	err := ctrl.authService.ResetPassword(req.Token, req.NewPassword)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error":   true,
			"message": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"error":   false,
		"message": "Password reset successful. Please login with your new password.",
	})
}

// Helper function to validate email format
func isValidEmail(email string) bool {
	// Basic email validation
	return len(email) > 3 && 
		   len(email) < 255 && 
		   strings.Contains(email, "@") && 
		   strings.Contains(email, ".")
}