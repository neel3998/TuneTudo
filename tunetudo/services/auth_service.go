package services

import (
	"database/sql"
	"errors"
	"time"
	"tunetudo/logger"
	"tunetudo/models"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/smtp"
	"os"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	db        *sql.DB
	jwtSecret []byte
}

func NewAuthService(db *sql.DB, jwtSecret string) *AuthService {
	return &AuthService{
		db:        db,
		jwtSecret: []byte(jwtSecret),
	}
}

// Add to existing AuthService struct
type PasswordResetToken struct {
	Token     string
	Email     string
	ExpiresAt time.Time
}

var passwordResetStore = make(map[string]PasswordResetToken)

// GenerateSecureToken generates a cryptographically secure random token
func GenerateSecureToken() (string, error) {
	bytes := make([]byte, 32) // 256 bits
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// SendPasswordResetEmail sends password reset email with token
func SendPasswordResetEmail(toEmail, token string) error {
	smtpHost := os.Getenv("SMTP_HOST")
	smtpPort := os.Getenv("SMTP_PORT")
	smtpUser := os.Getenv("SMTP_USER")
	smtpPass := os.Getenv("SMTP_PASS")
	fromEmail := os.Getenv("FROM_EMAIL")

	if smtpHost == "" || smtpPort == "" || smtpUser == "" || smtpPass == "" {
		return fmt.Errorf("email configuration missing in .env file")
	}

	resetLink := fmt.Sprintf("https://localhost:2701/reset-password.html?token=%s", token)

	subject := "Password Reset Request - TuneTudo"
	body := fmt.Sprintf(`Hello,

You requested a password reset for your TuneTudo account.

Click the link below to reset your password:
%s

This link will expire in 15 minutes.

If you didn't request this, please ignore this email and your password will remain unchanged.

Best regards,
TuneTudo Team`, resetLink)

	message := []byte(fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", fromEmail, toEmail, subject, body))

	auth := smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
	addr := smtpHost + ":" + smtpPort

	err := smtp.SendMail(addr, auth, fromEmail, []string{toEmail}, message)
	if err != nil {
		logger.Error(logger.CategoryAuth, "Failed to send password reset email", err)
		return err
	}

	logger.Info(logger.CategoryAuth, fmt.Sprintf("Password reset email sent to: %s", toEmail))
	return nil
}

// RequestPasswordReset initiates password reset flow
func (s *AuthService) RequestPasswordReset(email string) error {
	// Check if user exists
	var user models.User
	err := s.db.QueryRow("SELECT id, email, username FROM users WHERE email = ?", email).
		Scan(&user.ID, &user.Email, &user.Username)
	
	if err == sql.ErrNoRows {
		// Don't reveal if email exists or not (security best practice)
		logger.Info(logger.CategoryAuth, fmt.Sprintf("Password reset requested for non-existent email: %s", email))
		return nil
	}
	
	if err != nil {
		logger.Error(logger.CategoryAuth, "Database error during password reset request", err)
		return fmt.Errorf("failed to process request")
	}

	// Generate secure token
	token, err := GenerateSecureToken()
	if err != nil {
		logger.Error(logger.CategoryAuth, "Failed to generate reset token", err)
		return fmt.Errorf("failed to generate reset token")
	}

	// Store token with 15-minute expiration
	expiresAt := time.Now().Add(15 * time.Minute)
	passwordResetStore[email] = PasswordResetToken{
		Token:     token,
		Email:     email,
		ExpiresAt: expiresAt,
	}

	logger.Security("PASSWORD_RESET_REQUESTED", user.Username, email, 
		fmt.Sprintf("Password reset token generated (expires: %s)", expiresAt))

	// Send email
	if err := SendPasswordResetEmail(email, token); err != nil {
		delete(passwordResetStore, email)
		return fmt.Errorf("failed to send reset email")
	}

	return nil
}

// ValidateResetToken validates the reset token
func (s *AuthService) ValidateResetToken(token string) (string, error) {
	// Find token in store
	for email, resetData := range passwordResetStore {
		if resetData.Token == token {
			// Check expiration
			if time.Now().After(resetData.ExpiresAt) {
				delete(passwordResetStore, email)
				logger.Security("PASSWORD_RESET_TOKEN_EXPIRED", "anonymous", email, "Expired token used")
				return "", fmt.Errorf("reset token has expired")
			}
			return email, nil
		}
	}

	logger.Security("PASSWORD_RESET_INVALID_TOKEN", "anonymous", "unknown", "Invalid reset token used")
	return "", fmt.Errorf("invalid reset token")
}

// ResetPassword resets user password with token
func (s *AuthService) ResetPassword(token, newPassword string) error {
	// Validate token and get email
	email, err := s.ValidateResetToken(token)
	if err != nil {
		return err
	}

	// Validate password policy
	if len(newPassword) < 8 {
		return fmt.Errorf("password must be at least 8 characters")
	}

	// Get user
	var user models.User
	err = s.db.QueryRow("SELECT id, username FROM users WHERE email = ?", email).
		Scan(&user.ID, &user.Username)
	
	if err != nil {
		logger.Error(logger.CategoryAuth, "User not found during password reset", err)
		return fmt.Errorf("user not found")
	}

	// Hash new password with bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), 12)
	if err != nil {
		logger.Error(logger.CategoryAuth, "Failed to hash password", err)
		return fmt.Errorf("failed to process password")
	}

	// Update password in database
	_, err = s.db.Exec("UPDATE users SET password_hash = ? WHERE id = ?", 
		string(hashedPassword), user.ID)
	
	if err != nil {
		logger.Error(logger.CategoryAuth, "Failed to update password", err)
		return fmt.Errorf("failed to update password")
	}

	// Remove token from store
	delete(passwordResetStore, email)

	logger.Security("PASSWORD_RESET_SUCCESS", user.Username, email, "Password reset completed")

	return nil
}

// RegisterUser creates a new user account
func (s *AuthService) RegisterUser(req models.RegisterRequest, ipAddress string) (*models.User, error) {
	// Validate password strength
	if len(req.Password) < 8 {
		logger.ValidationFailure(req.Username, ipAddress, "password", "Password too short")
		return nil, errors.New("password must be at least 8 characters")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.Error(logger.CategoryAuth, "Password hashing failed", err)
		return nil, errors.New("failed to process registration")
	}

	// Insert user
	result, err := s.db.Exec(
		`INSERT INTO users (username, email, password_hash) VALUES (?, ?, ?)`,
		req.Username, req.Email, string(hashedPassword),
	)
	if err != nil {
		// Log without exposing email/username - don't reveal "no such user"
		logger.AuthAttempt(req.Username, ipAddress, false, "Registration failed - duplicate")
		return nil, errors.New("username or email already exists")
	}

	id, _ := result.LastInsertId()

	user := &models.User{
		ID:        int(id),
		Username:  req.Username,
		Email:     req.Email,
		IsAdmin:   false,
		CreatedAt: time.Now(),
	}

	logger.AuthAttempt(req.Username, ipAddress, true, "User registered successfully")
	logger.Info(logger.CategoryAuth, "New user registered: ID=%d", user.ID)

	return user, nil
}

// LoginUser authenticates user credentials
// Following "Just tell them 'authorization failed' – not 'no such user' or 'password incorrect'"
func (s *AuthService) LoginUser(req models.LoginRequest, ipAddress string) (string, *models.User, error) {
	var user models.User
	var passwordHash string

	err := s.db.QueryRow(
		`SELECT id, username, email, password_hash, is_admin, profile_image_path, created_at 
		FROM users WHERE username = ? OR email = ?`,
		req.Username, req.Username,
	).Scan(&user.ID, &user.Username, &user.Email, &passwordHash, &user.IsAdmin,
		&user.ProfileImagePath, &user.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			// Don't say "no such user" - just log it internally
			logger.AuthAttempt("attempted_user", ipAddress, false, "Account not found")
		} else {
			// Log system error without user details
			logger.Error(logger.CategoryAuth, "Login query failed", err)
		}
		// Return same error message for both cases (timing attack prevention)
		return "", nil, errors.New("authorization failed")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		// Don't say "password incorrect" - use generic message
		logger.AuthAttempt(user.Username, ipAddress, false, "Invalid credentials")
		return "", nil, errors.New("authorization failed")
	}

	// Update last login
	_, err = s.db.Exec(`UPDATE users SET last_login = CURRENT_TIMESTAMP WHERE id = ?`, user.ID)
	if err != nil {
		logger.Warning(logger.CategoryAuth, "Failed to update last login time for user_id=%d", user.ID)
	}

	// Generate JWT token
	token, err := s.GenerateToken(&user)
	if err != nil {
		logger.Error(logger.CategoryAuth, "Token generation failed", err)
		return "", nil, errors.New("failed to generate authentication token")
	}

	// Log successful login - username will be hashed by logger
	logger.AuthAttempt(user.Username, ipAddress, true, "Login successful")
	logger.SessionCreated(user.Username, ipAddress)
	logger.Info(logger.CategoryAuth, "User login: user_id=%d from IP=%s", user.ID, logger.MaskIP(ipAddress))

	return token, &user, nil
}

// GenerateToken creates a JWT token for the user
func (s *AuthService) GenerateToken(user *models.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"is_admin": user.IsAdmin,
		"exp":      time.Now().Add(time.Hour * 24 * 7).Unix(), // 7 days
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken validates and parses a JWT token
func (s *AuthService) ValidateToken(tokenString string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid token signing method")
		}
		return s.jwtSecret, nil
	})

	if err != nil {
		// Don't log token details - no sensitive data in logs
		logger.Warning(logger.CategoryAuth, "Token validation failed")
		return nil, errors.New("invalid or expired token")
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		// Check token expiration
		if exp, ok := claims["exp"].(float64); ok {
			if time.Now().Unix() > int64(exp) {
				logger.Warning(logger.CategoryAuth, "Expired token used")
				if username, ok := claims["username"].(string); ok {
					logger.SessionExpired(username)
				}
				return nil, errors.New("token has expired")
			}
		}
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// GetUserByID retrieves a user by ID
func (s *AuthService) GetUserByID(userID int) (*models.User, error) {
	var user models.User
	err := s.db.QueryRow(
		`SELECT id, username, email, is_admin, profile_image_path, created_at, last_login 
		FROM users WHERE id = ?`,
		userID,
	).Scan(&user.ID, &user.Username, &user.Email, &user.IsAdmin,
		&user.ProfileImagePath, &user.CreatedAt, &user.LastLogin)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("user not found")
		}
		// Don't log user_id in error to avoid correlation
		logger.Error(logger.CategoryDB, "Failed to retrieve user", err)
		return nil, errors.New("failed to retrieve user information")
	}

	return &user, nil
}

// CheckPasswordPolicy validates password strength
func (s *AuthService) CheckPasswordPolicy(password string) error {
	if len(password) < 8 {
		return errors.New("password must be at least 8 characters")
	}
	// Add more password policy checks as needed
	return nil
}