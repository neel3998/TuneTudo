package middleware

import (
	"errors"
	"strings"
	"tunetudo/logger"
	"tunetudo/services"

	"github.com/gofiber/fiber/v2"
)

// AuthMiddleware validates JWT tokens
func AuthMiddleware(authService *services.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get token from Authorization header
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			ip := c.IP()
			logger.AccessDenied("anonymous", ip, c.Path(), "No authorization token provided")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   true,
				"message": "Authentication required",
			})
		}

		// Extract token (Bearer <token>)
		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			ip := c.IP()
			logger.AccessDenied("anonymous", ip, c.Path(), "Invalid authorization format")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   true,
				"message": "Invalid authorization format",
			})
		}

		token := tokenParts[1]

		// Validate token
		claims, err := authService.ValidateToken(token)
		if err != nil {
			ip := c.IP()
			logger.AccessDenied("anonymous", ip, c.Path(), "Invalid or expired token")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   true,
				"message": "Invalid or expired token",
			})
		}

		// Extract user information from claims
		userID, ok := claims["user_id"].(float64)
		if !ok {
			ip := c.IP()
			logger.AccessDenied("anonymous", ip, c.Path(), "Invalid token claims")
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   true,
				"message": "Invalid token",
			})
		}

		username, _ := claims["username"].(string)
		isAdmin, _ := claims["is_admin"].(bool)

		// Store user info in context
		c.Locals("user_id", int(userID))
		c.Locals("username", username)
		c.Locals("is_admin", isAdmin)

		return c.Next()
	}
}

// AdminMiddleware checks if user has admin privileges
func AdminMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		isAdmin, ok := c.Locals("is_admin").(bool)
		username := c.Locals("username")
		ip := c.IP()

		if !ok || !isAdmin {
			userStr := "anonymous"
			if username != nil {
				userStr = username.(string)
			}
			logger.AccessDenied(userStr, ip, c.Path(), "Admin access required")
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"error":   true,
				"message": "Admin access required",
			})
		}

		// Log admin action
		if username != nil {
			logger.AdminAction(username.(string), ip, c.Method()+" "+c.Path(), "Admin endpoint accessed")
		}

		return c.Next()
	}
}

// GetUserID extracts user ID from context
func GetUserID(c *fiber.Ctx) (int, error) {
	userID, ok := c.Locals("user_id").(int)
	if !ok {
		ip := c.IP()
		logger.AccessDenied("anonymous", ip, c.Path(), "User ID not found in context")
		return 0, c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error":   true,
			"message": "Authentication required",
		})
	}
	return userID, nil
}

// GetUsername extracts username from context
func GetUsername(c *fiber.Ctx) (string, error) {
	username, ok := c.Locals("username").(string)
	if !ok {
		return "", errors.New("username not found in context")
	}
	return username, nil
}

// IsAdmin checks if the current user is an admin
func IsAdmin(c *fiber.Ctx) bool {
	isAdmin, ok := c.Locals("is_admin").(bool)
	return ok && isAdmin
}