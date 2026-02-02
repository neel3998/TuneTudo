package main

import (
	"log"
	"net/http"
	"path/filepath"
	"os"
	"time"
	"tunetudo/config"
	"tunetudo/database"
	"tunetudo/logger"
	"tunetudo/middleware"
	"fmt"
	"github.com/joho/godotenv"
	"tunetudo/routes"
	"github.com/gofiber/fiber/v2/middleware/helmet"
	"github.com/gofiber/fiber/v2"
	fiberlogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/limiter"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: No .env file found or error loading it:", err)
	}

	// Initialize logger
	// "Centralize all logging/debugging, use consistently"
	logPath := filepath.Join(".", "logs", "app.log")
	if err := logger.InitLogger(logPath); err != nil {
		log.Fatal("Failed to initialize logger:", err)
	}
	logger.Info(logger.CategoryAPI, "Logger initialized successfully")

	// Initialize database
	absPath, _ := filepath.Abs("./tunetudo.db")
	db, err := database.InitDB(absPath)
	if err != nil {
		logger.Error(logger.CategoryDB, "Failed to initialize database", err)
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()
	logger.Info(logger.CategoryDB, "Database initialized successfully")

	// Run migrations
	if err := database.RunMigrations(db); err != nil {
		logger.Error(logger.CategoryDB, "Failed to run migrations", err)
		log.Fatal("Failed to run migrations:", err)
	}
	logger.Info(logger.CategoryDB, "Database migrations completed")

	// Initialize Fiber app with custom error handler
	app := fiber.New(fiber.Config{
		BodyLimit:     50 * 1024 * 1024, // 50MB for file uploads
		Prefork:       false,
		StrictRouting: false,
		CaseSensitive: false,
		ErrorHandler:  middleware.ErrorHandler, 
		ServerHeader:  "",                      // Don't expose server info
		AppName:       "TuneTudo v1.0",
	})

	// Security Middleware - Applied globally
	app.Use(recover.New(recover.Config{
		EnableStackTrace: false, // Don't expose stack traces
	}))

	app.Use(helmet.New())
	// Rate limiting to prevent abuse
	app.Use(limiter.New(limiter.Config{
		Max:        50,             // 50 requests
		Expiration: 1 * time.Minute, // per minute
		LimitReached: func(c *fiber.Ctx) error {
			ip := c.IP()
			logger.Security("RATE_LIMIT_EXCEEDED", "anonymous", logger.MaskIP(ip), "Rate limit exceeded")
			return c.Status(429).JSON(fiber.Map{
				"error":   true,
				"message": "Too many requests. Please try again later.",
			})
		},
	}))

	// Request logging
	app.Use(fiberlogger.New(fiberlogger.Config{
		Format:     "[${time}] ${status} - ${method} ${path} - ${latency}\n",
		TimeFormat: "2006-01-02 15:04:05",
		TimeZone:   "Local",
	}))

	// Security logger middleware
	app.Use(middleware.SecurityLogger())

	// Request validator middleware
	// Appropriately filter or quote CRLF sequences in user-controlled input
	app.Use(middleware.RequestValidator())

	// Setup routes
	routes.SetupRoutes(app, db)

	// Start server
	// "Categorize messages so operators can configure what gets logged"
	logger.Info(logger.CategoryAPI, "🎵 TuneTudo Server starting")

	
	// TLS configuration
	certicateFile := os.Getenv("TLS_CERT_FILE")
	keyFile := os.Getenv("TLS_KEY_FILE")
	log.Printf("Using TLS cert: %s and key: %s", certicateFile, keyFile)
	if certicateFile == "" || keyFile == "" {
		logger.Error(logger.CategoryAPI, "TLS certificate or key file not specified in environment variables", nil)
		log.Fatal("TLS certificate or key file not specified in environment variables")
	}

	go func() {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			target := "https://" + r.Host + r.URL.RequestURI()
			// redirect to HTTPS
			if r.Host == "localhost:2701" {
				target = "https://localhost:2701" + r.URL.RequestURI()
			}
			http.Redirect(w, r, target, http.StatusMovedPermanently)
		})
		if err := http.ListenAndServe(":" + cfg.Port, nil); err != nil {
			fmt.Printf("HTTP redirect server stopped due to: %v\n", err)
		}
	}()

	if err := app.ListenTLS(":" + cfg.Port, certicateFile, keyFile); err != nil {
		logger.Error(logger.CategoryAPI, "Server failed to start", err)
		log.Fatalf("Failed to start the TLS server: %v", err)
	}

}