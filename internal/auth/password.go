package auth

import (
	"errors"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/crypto/bcrypt"
)

const (
	// bcrypt cost factor - higher is more secure but slower
	bcryptCost = 12

	// Minimum password length
	minPasswordLength = 8

	// Maximum password length (bcrypt has a 72 byte limit)
	maxPasswordLength = 72
)

var (
	ErrPasswordTooShort  = errors.New("password must be at least 8 characters")
	ErrPasswordTooLong   = errors.New("password must be at most 72 characters")
	ErrPasswordNoUpper   = errors.New("password must contain at least one uppercase letter")
	ErrPasswordNoLower   = errors.New("password must contain at least one lowercase letter")
	ErrPasswordNoDigit   = errors.New("password must contain at least one digit")
	ErrPasswordNoSpecial = errors.New("password must contain at least one special character")
	ErrInvalidEmail      = errors.New("invalid email format")
	ErrEmailTooLong      = errors.New("email must be at most 255 characters")
	ErrNameTooShort      = errors.New("name must be at least 1 character")
	ErrNameTooLong       = errors.New("name must be at most 100 characters")
)

// emailRegex is a basic but practical email validation regex
var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// HashPassword creates a bcrypt hash of the password
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// VerifyPassword checks if the provided password matches the hash
func VerifyPassword(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// ValidatePassword checks if the password meets security requirements
func ValidatePassword(password string) error {
	if len(password) < minPasswordLength {
		return ErrPasswordTooShort
	}
	if len(password) > maxPasswordLength {
		return ErrPasswordTooLong
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return ErrPasswordNoUpper
	}
	if !hasLower {
		return ErrPasswordNoLower
	}
	if !hasDigit {
		return ErrPasswordNoDigit
	}
	if !hasSpecial {
		return ErrPasswordNoSpecial
	}

	return nil
}

// ValidateEmail checks if the email format is valid
func ValidateEmail(email string) error {
	email = strings.TrimSpace(email)
	if len(email) > 255 {
		return ErrEmailTooLong
	}
	if !emailRegex.MatchString(email) {
		return ErrInvalidEmail
	}
	return nil
}

// ValidateName checks if the name meets requirements
func ValidateName(name string) error {
	name = strings.TrimSpace(name)
	if len(name) < 1 {
		return ErrNameTooShort
	}
	if len(name) > 100 {
		return ErrNameTooLong
	}
	return nil
}

// NormalizeEmail lowercases and trims the email
func NormalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}
