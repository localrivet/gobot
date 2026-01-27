package auth

import (
	"strings"
	"testing"
)

func TestHashPassword(t *testing.T) {
	password := "SecureP@ss123!"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	if hash == "" {
		t.Error("Hash is empty")
	}

	if hash == password {
		t.Error("Hash should not equal plain password")
	}

	// Hash should start with bcrypt identifier
	if !strings.HasPrefix(hash, "$2a$") && !strings.HasPrefix(hash, "$2b$") {
		t.Error("Hash should be a bcrypt hash")
	}
}

func TestVerifyPassword(t *testing.T) {
	password := "SecureP@ss123!"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	// Correct password should verify
	if !VerifyPassword(password, hash) {
		t.Error("VerifyPassword should return true for correct password")
	}

	// Wrong password should not verify
	if VerifyPassword("wrong-password", hash) {
		t.Error("VerifyPassword should return false for wrong password")
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  error
	}{
		{
			name:     "valid password",
			password: "SecureP@ss123!",
			wantErr:  nil,
		},
		{
			name:     "too short",
			password: "Sh0rt!",
			wantErr:  ErrPasswordTooShort,
		},
		{
			name:     "no uppercase",
			password: "lowercase123!",
			wantErr:  ErrPasswordNoUpper,
		},
		{
			name:     "no lowercase",
			password: "UPPERCASE123!",
			wantErr:  ErrPasswordNoLower,
		},
		{
			name:     "no digit",
			password: "NoDigitsHere!",
			wantErr:  ErrPasswordNoDigit,
		},
		{
			name:     "no special character",
			password: "NoSpecial123",
			wantErr:  ErrPasswordNoSpecial,
		},
		{
			name:     "exactly minimum length",
			password: "Short1@a",
			wantErr:  nil,
		},
		{
			name:     "too long (73 chars)",
			password: strings.Repeat("Aa1!", 18) + "X", // 73 characters
			wantErr:  ErrPasswordTooLong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password)
			if err != tt.wantErr {
				t.Errorf("ValidatePassword() = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr error
	}{
		{
			name:    "valid email",
			email:   "user@example.com",
			wantErr: nil,
		},
		{
			name:    "valid email with subdomain",
			email:   "user@mail.example.com",
			wantErr: nil,
		},
		{
			name:    "valid email with plus",
			email:   "user+tag@example.com",
			wantErr: nil,
		},
		{
			name:    "valid email with dots",
			email:   "first.last@example.com",
			wantErr: nil,
		},
		{
			name:    "missing @",
			email:   "userexample.com",
			wantErr: ErrInvalidEmail,
		},
		{
			name:    "missing domain",
			email:   "user@",
			wantErr: ErrInvalidEmail,
		},
		{
			name:    "missing TLD",
			email:   "user@example",
			wantErr: ErrInvalidEmail,
		},
		{
			name:    "missing local part",
			email:   "@example.com",
			wantErr: ErrInvalidEmail,
		},
		{
			name:    "empty string",
			email:   "",
			wantErr: ErrInvalidEmail,
		},
		{
			name:    "with whitespace",
			email:   "  user@example.com  ",
			wantErr: nil, // Trimming should handle this
		},
		{
			name:    "too long",
			email:   strings.Repeat("a", 250) + "@example.com",
			wantErr: ErrEmailTooLong,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email)
			if err != tt.wantErr {
				t.Errorf("ValidateEmail(%q) = %v, want %v", tt.email, err, tt.wantErr)
			}
		})
	}
}

func TestValidateName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr error
	}{
		{
			name:    "valid name",
			input:   "John Doe",
			wantErr: nil,
		},
		{
			name:    "single character",
			input:   "A",
			wantErr: nil,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: ErrNameTooShort,
		},
		{
			name:    "only whitespace",
			input:   "   ",
			wantErr: ErrNameTooShort,
		},
		{
			name:    "too long",
			input:   strings.Repeat("a", 101),
			wantErr: ErrNameTooLong,
		},
		{
			name:    "exactly 100 characters",
			input:   strings.Repeat("a", 100),
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateName(tt.input)
			if err != tt.wantErr {
				t.Errorf("ValidateName(%q) = %v, want %v", tt.input, err, tt.wantErr)
			}
		})
	}
}

func TestNormalizeEmail(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"User@Example.com", "user@example.com"},
		{"  user@example.com  ", "user@example.com"},
		{"USER@EXAMPLE.COM", "user@example.com"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := NormalizeEmail(tt.input)
			if result != tt.expected {
				t.Errorf("NormalizeEmail(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func BenchmarkHashPassword(b *testing.B) {
	password := "SecureP@ss123!"
	for i := 0; i < b.N; i++ {
		HashPassword(password)
	}
}

func BenchmarkVerifyPassword(b *testing.B) {
	password := "SecureP@ss123!"
	hash, _ := HashPassword(password)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		VerifyPassword(password, hash)
	}
}
