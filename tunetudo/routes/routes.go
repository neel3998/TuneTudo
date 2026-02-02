package routes

import (
	"database/sql"
	"tunetudo/config"
	"tunetudo/controllers"
	"tunetudo/middleware"
	"tunetudo/services"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(app *fiber.App, db *sql.DB) {
	cfg := config.LoadConfig()

	// Initialize services
	authService := services.NewAuthService(db, cfg.JWTSecret)
	searchService := services.NewSearchService(db)
	playlistService := services.NewPlaylistService(db)
	playbackService := services.NewPlaybackService(db, cfg.StoragePath)
	userService := services.NewUserService(db, cfg.StoragePath)
	adminService := services.NewAdminService(db, cfg.StoragePath)

	// Initialize controllers
	authCtrl := controllers.NewAuthController(authService)
	searchCtrl := controllers.NewSearchController(searchService)
	playlistCtrl := controllers.NewPlaylistController(playlistService)
	playbackCtrl := controllers.NewPlaybackController(playbackService)
	userCtrl := controllers.NewUserController(userService)
	adminCtrl := controllers.NewAdminController(adminService)

	// Health check - should be first
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"message": "TuneTudo API is running",
		})
	})

	// Serve static files - IMPORTANT: This must come before HTML routes
	app.Static("/static", "./static")
	app.Static("/storage", "./storage")

	// API routes
	api := app.Group("/api")

	// Public routes - Authentication
	auth := api.Group("/auth")
	auth.Post("/register", authCtrl.Register)
	auth.Post("/login", authCtrl.Login)
	auth.Post("/logout", authCtrl.Logout)
	// In the auth group section, add:
	auth.Post("/forgot-password", authCtrl.ForgotPassword)
	auth.Get("/validate-reset-token", authCtrl.ValidateResetToken)
	auth.Post("/reset-password", authCtrl.ResetPassword)

	// Public routes - Search and Browse
	api.Get("/search", searchCtrl.Search)
	api.Get("/categories", searchCtrl.GetCategories)
	api.Get("/categories/:id/songs", searchCtrl.GetSongsByCategory)
	api.Get("/songs/recent", playbackCtrl.GetRecentSongs)
	api.Get("/songs/:id", playbackCtrl.GetSong)
	api.Get("/songs/:id/stream", playbackCtrl.StreamSong)

	// Protected routes - require authentication
	protected := api.Group("", middleware.AuthMiddleware(authService))

	// User profile routes
	protected.Get("/profile", authCtrl.GetProfile)
	protected.Put("/profile/picture", userCtrl.UploadProfileImage)

	// Playlist routes
	protected.Get("/playlists", playlistCtrl.GetUserPlaylists)
	protected.Post("/playlists", playlistCtrl.CreatePlaylist)
	protected.Get("/playlists/:id", playlistCtrl.GetPlaylistDetails)
	protected.Post("/playlists/:id/songs", playlistCtrl.AddSongToPlaylist)
	protected.Delete("/playlists/:id/songs/:songId", playlistCtrl.RemoveSongFromPlaylist)
	protected.Delete("/playlists/:id", playlistCtrl.DeletePlaylist)

	// User upload routes
	protected.Post("/upload", userCtrl.UploadSong)
	protected.Get("/uploads", userCtrl.GetUserUploads)

	// Admin routes - require admin privileges
	admin := api.Group("/admin", middleware.AuthMiddleware(authService), middleware.AdminMiddleware())
	admin.Post("/songs", adminCtrl.UploadSong)
	admin.Delete("/songs/:id", adminCtrl.DeleteSong)
	admin.Get("/songs", adminCtrl.GetAllSongs)
	admin.Get("/users", adminCtrl.GetAllUsers)

	// Serve HTML pages - MUST BE LAST (after all /api routes)
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendFile("./static/index.html")
	})
	app.Get("/index.html", func(c *fiber.Ctx) error {
		return c.SendFile("./static/index.html")
	})
	app.Get("/playlists.html", func(c *fiber.Ctx) error {
		return c.SendFile("./static/playlists.html")
	})
	app.Get("/uploads.html", func(c *fiber.Ctx) error {
		return c.SendFile("./static/uploads.html")
	})
	app.Get("/profile.html", func(c *fiber.Ctx) error {
		return c.SendFile("./static/profile.html")
	})
	app.Get("/admin.html", func(c *fiber.Ctx) error {
		return c.SendFile("./static/admin.html")
	})
		// Add after other HTML routes
	app.Get("/forgot-password.html", func(c *fiber.Ctx) error {
		return c.SendFile("./static/forgot-password.html")
	})
	app.Get("/reset-password.html", func(c *fiber.Ctx) error {
		return c.SendFile("./static/reset-password.html")
	})
}
