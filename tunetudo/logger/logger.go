package logger

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type LogLevel string

const (
	LogLevelInfo     LogLevel = "INFO"
	LogLevelWarning  LogLevel = "WARNING"
	LogLevelError    LogLevel = "ERROR"
	LogLevelSecurity LogLevel = "SECURITY"
	LogLevelDebug    LogLevel = "DEBUG"
)

type Logger struct {
	infoLogger     *log.Logger
	warningLogger  *log.Logger
	errorLogger    *log.Logger
	securityLogger *log.Logger
	debugLogger    *log.Logger
}

var defaultLogger *Logger

// Category constants for filtering logs
const (
	CategoryAuth       = "[AUTH]"
	CategoryFile       = "[FILE]"
	CategoryDB         = "[DATABASE]"
	CategoryAPI        = "[API]"
	CategoryPlaylist   = "[PLAYLIST]"
	CategoryUpload     = "[UPLOAD]"
	CategoryAdmin      = "[ADMIN]"
	CategoryValidation = "[VALIDATION]"
)

func InitLogger(logPath string) error {
	// Create logs directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(logPath), 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Open log file with restricted permissions (0600)
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	// Create security log file with restricted permissions
	securityLogPath := filepath.Join(filepath.Dir(logPath), "security.log")
	securityFile, err := os.OpenFile(securityLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return fmt.Errorf("failed to open security log file: %w", err)
	}

	// Create debug log file
	debugLogPath := filepath.Join(filepath.Dir(logPath), "debug.log")
	debugFile, err := os.OpenFile(debugLogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return fmt.Errorf("failed to open debug log file: %w", err)
	}

	defaultLogger = &Logger{
		infoLogger:     log.New(logFile, "[INFO] ", log.Ldate|log.Ltime|log.Lshortfile),
		warningLogger:  log.New(logFile, "[WARNING] ", log.Ldate|log.Ltime|log.Lshortfile),
		errorLogger:    log.New(logFile, "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile),
		securityLogger: log.New(securityFile, "[SECURITY] ", log.Ldate|log.Ltime|log.Lshortfile),
		debugLogger:    log.New(debugFile, "[DEBUG] ", log.Ldate|log.Ltime|log.Lshortfile),
	}

	return nil
}

// HashIdentifier creates a hash of sensitive identifiers (username, email, etc.)
func HashIdentifier(identifier string) string {
	if identifier == "" || identifier == "anonymous" {
		return "anonymous"
	}
	hash := sha256.Sum256([]byte(identifier))
	return "user_" + hex.EncodeToString(hash[:])[:8]
}

// MaskIP partially masks an IP address for privacy
func MaskIP(ip string) string {
	if ip == "" {
		return "unknown"
	}
	parts := strings.Split(ip, ".")
	if len(parts) == 4 {
		// Mask last two octets: 192.168.x.x
		return parts[0] + "." + parts[1] + ".x.x"
	}
	// For IPv6 or invalid format
	return "ip_present"
}

// RemoveCarriageReturns removes CRLF sequences to prevent log injection
func RemoveCarriageReturns(input string) string {
	input = strings.ReplaceAll(input, "\r", "")
	input = strings.ReplaceAll(input, "\n", "")
	input = strings.ReplaceAll(input, "\r\n", "")
	return input
}

// Categorized logging functions

func Info(category, msg string, args ...interface{}) {
	safeMsg := RemoveCarriageReturns(msg)
	fullMsg := category + " " + safeMsg
	if defaultLogger != nil {
		defaultLogger.infoLogger.Printf(fullMsg, args...)
	}
	log.Printf("[INFO] "+fullMsg, args...)
}

func Warning(category, msg string, args ...interface{}) {
	safeMsg := RemoveCarriageReturns(msg)
	fullMsg := category + " " + safeMsg
	if defaultLogger != nil {
		defaultLogger.warningLogger.Printf(fullMsg, args...)
	}
	log.Printf("[WARNING] "+fullMsg, args...)
}

func Error(category, msg string, err error) {
	safeMsg := RemoveCarriageReturns(msg)
	fullMsg := category + " " + safeMsg
	if err != nil {
		// Log error type without exposing sensitive details
		fullMsg = fmt.Sprintf("%s: error occurred", fullMsg)
	}
	if defaultLogger != nil {
		defaultLogger.errorLogger.Println(fullMsg)
	}
	log.Printf("[ERROR] " + fullMsg)
}

func Debug(category, msg string, args ...interface{}) {
	safeMsg := RemoveCarriageReturns(msg)
	fullMsg := category + " " + safeMsg
	if defaultLogger != nil {
		defaultLogger.debugLogger.Printf(fullMsg, args...)
	}
}

// Security logs authentication and authorization events with PII protection
func Security(eventType, userHash, maskedIP, details string) {
	safeDetails := RemoveCarriageReturns(details)
	logMsg := fmt.Sprintf("Event: %s | UserHash: %s | IP: %s | Details: %s | Time: %s",
		eventType, userHash, maskedIP, safeDetails, time.Now().Format(time.RFC3339))

	if defaultLogger != nil {
		defaultLogger.securityLogger.Println(logMsg)
	}
	log.Printf("[SECURITY] %s", logMsg)
}

// AuthAttempt logs authentication attempts (login, logout, registration)
func AuthAttempt(username, ipAddress string, success bool, reason string) {
	status := "SUCCESS"
	if !success {
		status = "FAILED"
	}

	userHash := HashIdentifier(username)
	maskedIP := MaskIP(ipAddress)
	sanitizedReason := RemoveCarriageReturns(reason)

	Security(fmt.Sprintf("AUTH_%s", status), userHash, maskedIP, sanitizedReason)
}

// AccessDenied logs unauthorized access attempts
func AccessDenied(username, ipAddress, resource, reason string) {
	userHash := HashIdentifier(username)
	maskedIP := MaskIP(ipAddress)

	sanitizedResource := sanitizeResourcePath(resource)
	sanitizedReason := RemoveCarriageReturns(reason)

	details := fmt.Sprintf("Resource: %s | Reason: %s", sanitizedResource, sanitizedReason)
	Security("ACCESS_DENIED", userHash, maskedIP, details)
}

// AdminAction logs administrative actions
func AdminAction(username, ipAddress, action, details string) {
	userHash := HashIdentifier(username)
	maskedIP := MaskIP(ipAddress)

	sanitizedAction := RemoveCarriageReturns(action)
	sanitizedDetails := RemoveCarriageReturns(details)

	Security("ADMIN_ACTION", userHash, maskedIP, fmt.Sprintf("Action: %s | Details: %s", sanitizedAction, sanitizedDetails))
}

// ValidationFailure logs input validation failures without exposing input values
func ValidationFailure(username, ipAddress, field, reason string) {
	userHash := HashIdentifier(username)
	maskedIP := MaskIP(ipAddress)

	sanitizedField := RemoveCarriageReturns(field)
	sanitizedReason := RemoveCarriageReturns(reason)

	details := fmt.Sprintf("Field: %s | Reason: %s", sanitizedField, sanitizedReason)
	Security("VALIDATION_FAILURE", userHash, maskedIP, details)
}

// DataAccess logs access to sensitive data
func DataAccess(username, ipAddress, resource string) {
	userHash := HashIdentifier(username)
	maskedIP := MaskIP(ipAddress)

	sanitizedResource := sanitizeResourcePath(resource)
	Security("DATA_ACCESS", userHash, maskedIP, fmt.Sprintf("Resource: %s", sanitizedResource))
}

// SessionCreated logs new session creation
func SessionCreated(username, ipAddress string) {
	userHash := HashIdentifier(username)
	maskedIP := MaskIP(ipAddress)

	Security("SESSION_CREATED", userHash, maskedIP, "New session established")
}

// SessionExpired logs session expiration
func SessionExpired(username string) {
	userHash := HashIdentifier(username)

	Security("SESSION_EXPIRED", userHash, "system", "Session expired")
}

// sanitizeResourcePath removes sensitive parts of resource paths
func sanitizeResourcePath(resource string) string {
	// Keep only the general endpoint category
	parts := strings.Split(resource, "/")
	if len(parts) > 3 {
		// Keep /api/category format, hide specific IDs
		return strings.Join(parts[:3], "/") + "/*"
	}
	return resource
}

// SanitizeResourcePath is the exported version for use in other packages
func SanitizeResourcePath(resource string) string {
	return sanitizeResourcePath(resource)
}