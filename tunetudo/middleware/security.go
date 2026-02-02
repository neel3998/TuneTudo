package middleware

import (
	"strings"
	"tunetudo/logger"

	"github.com/gofiber/fiber/v2"
)

// SecurityLogger logs security-relevant events with PII protection
// Following "Categorize messages so operators can configure what gets logged"
func SecurityLogger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get user info if authenticated
		username := c.Locals("username")
		
		// Get request info
		method := c.Method()
		path := c.Path()
		
		// Log security-relevant requests
		if isSensitiveEndpoint(path) {
			userStr := "anonymous"
			if username != nil {
				userStr = username.(string)
			}
			
			// DataAccess handles all sanitization internally
			logger.DataAccess(userStr, c.IP(), method+" "+path)
		}
		
		// Process request
		err := c.Next()
		
		// Log errors without exposing sensitive details
		// "Log instead of revealing problem details to users"
		if err != nil {
			// Don't log full error message - may contain sensitive data
			logger.Error(logger.CategoryAPI, "Request error occurred", err)
		}
		
		// Log failed authentication/authorization
		// "Record important successes & failures"
		status := c.Response().StatusCode()
		if status == 401 || status == 403 {
			userStr := "anonymous"
			if username != nil {
				userStr = username.(string)
			}
			reason := "Access denied"
			if status == 401 {
				reason = "Authentication required"
			} else {
				reason = "Insufficient permissions"
			}
			logger.AccessDenied(userStr, c.IP(), path, reason)
		}
		
		return err
	}
}

// isSensitiveEndpoint checks if an endpoint handles sensitive data
func isSensitiveEndpoint(path string) bool {
	sensitivePatterns := []string{
		"/api/auth/",
		"/api/user/",
		"/api/admin/",
		"/api/profile",
		"/api/upload",
		"/api/playlists",
	}
	
	for _, pattern := range sensitivePatterns {
		if strings.HasPrefix(path, pattern) {
			return true
		}
	}
	return false
}

// ErrorHandler is a custom error handler that logs errors and returns safe messages
func ErrorHandler(c *fiber.Ctx, err error) error {
	// Default to 500 Internal Server Error
	code := fiber.StatusInternalServerError
	message := "An internal error occurred"
	errorCode := "INTERNAL_ERROR"
	
	// Extract fiber error
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
		// "Do not reveal sensitive information from the error to the user"
		switch code {
		case 400:
			message = "Invalid request"
			errorCode = "BAD_REQUEST"
		case 401:
			message = "Authentication required"
			errorCode = "UNAUTHORIZED"
		case 403:
			message = "Access forbidden"
			errorCode = "FORBIDDEN"
		case 404:
			message = "Resource not found"
			errorCode = "NOT_FOUND"
		case 409:
			message = "Resource conflict"
			errorCode = "CONFLICT"
		case 429:
			message = "Too many requests"
			errorCode = "RATE_LIMIT"
		default:
			message = "An error occurred"
		}
	}
	
	// Log the error with minimal details
	// "Logs usually sent to separate system in operation"
	maskedIP := logger.MaskIP(c.IP())
	sanitizedPath := sanitizeResourcePath(c.Path())
	method := c.Method()
	
	logger.Error(logger.CategoryAPI,
		"Request error",
		nil, // Don't log the actual error - may contain sensitive data
	)
	
	logger.Debug(logger.CategoryAPI, "Error details: method=%s path=%s ip=%s status=%d", 
		method, sanitizedPath, maskedIP, code)
	
	// Return safe error response
	return c.Status(code).JSON(fiber.Map{
		"error":   true,
		"code":    errorCode,
		"message": message,
	})
}

// sanitizeResourcePath removes sensitive parts of resource paths
func sanitizeResourcePath(resource string) string {
	parts := strings.Split(resource, "/")
	if len(parts) > 3 {
		return strings.Join(parts[:3], "/") + "/*"
	}
	return resource
}

// RequestValidator validates common request parameters
// "Appropriately filter or quote CRLF sequences in user-controlled input"
func RequestValidator() fiber.Handler {
	return func(c *fiber.Ctx) error {
		username := c.Locals("username")
		userStr := "anonymous"
		if username != nil {
			userStr = username.(string)
		}
		
		// Check for suspicious patterns in query parameters
		queries := c.Queries()
		for _, value := range queries {
			// Remove CRLF to prevent log injection
			cleanValue := logger.RemoveCarriageReturns(value)
			if containsSuspiciousPattern(cleanValue) {
				// Log without exposing the actual suspicious value
				logger.ValidationFailure(userStr, c.IP(), cleanValue, "Suspicious pattern detected")
				return c.Status(400).JSON(fiber.Map{
					"error":   true,
					"message": "Invalid input detected",
				})
			}
			
			// Check for overly long inputs (potential DoS)
			if len(cleanValue) > 1000 {
				logger.ValidationFailure(userStr, c.IP(), cleanValue, "Input too long")
				return c.Status(400).JSON(fiber.Map{
					"error":   true,
					"message": "Input exceeds maximum length",
				})
			}
		}
		
		// Validate body size is reasonable
		if len(c.Body()) > 50*1024*1024 { 
			if !strings.Contains(c.Path(), "/upload") {
				logger.ValidationFailure(userStr, c.IP(), "body", "Request body too large")
				return c.Status(413).JSON(fiber.Map{
					"error":   true,
					"message": "Request body too large",
				})
			}
		}
		
		return c.Next()
	}
}

// containsSuspiciousPattern checks for common injection patterns
// Does NOT log the actual input value - only the pattern type
func containsSuspiciousPattern(input string) bool {
	suspicious := []string{
		"<script",
		"javascript:",
		"onerror=",
		"onclick=",
		"../",
		"..\\",
		"DROP TABLE",
		"DELETE FROM",
		"INSERT INTO",
		"UPDATE ",
		"UNION SELECT",
		"'; --",
		"OR 1=1",
	}
	
	lowerInput := strings.ToLower(input)
	for _, pattern := range suspicious {
		if strings.Contains(lowerInput, strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}