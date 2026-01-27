package security

import (
	"testing"
)

func TestSQLValidator_ValidateInput(t *testing.T) {
	validator := NewSQLValidator()

	tests := []struct {
		name        string
		input       string
		expectError bool
	}{
		{
			name:        "normal text",
			input:       "Hello World",
			expectError: false,
		},
		{
			name:        "normal email",
			input:       "test@example.com",
			expectError: false,
		},
		{
			name:        "normal number",
			input:       "12345",
			expectError: false,
		},
		{
			name:        "SQL comment double dash",
			input:       "test--comment",
			expectError: true,
		},
		{
			name:        "SQL comment block",
			input:       "test/*comment*/",
			expectError: true,
		},
		{
			name:        "UNION SELECT attack",
			input:       "1 UNION SELECT * FROM users",
			expectError: true,
		},
		{
			name:        "SELECT FROM attack",
			input:       "'; SELECT password FROM users",
			expectError: true,
		},
		{
			name:        "INSERT INTO attack",
			input:       "test'; INSERT INTO admin VALUES",
			expectError: true,
		},
		{
			name:        "DELETE FROM attack",
			input:       "'; DELETE FROM users WHERE 1=1",
			expectError: true,
		},
		{
			name:        "DROP TABLE attack",
			input:       "'; DROP TABLE users; --",
			expectError: true,
		},
		{
			name:        "UPDATE SET attack",
			input:       "test'; UPDATE users SET admin=1",
			expectError: true,
		},
		{
			name:        "EXEC attack",
			input:       "'; EXEC(master.dbo.xp_cmdshell)",
			expectError: true,
		},
		{
			name:        "OR 1=1 attack",
			input:       "admin' OR 1=1 --",
			expectError: true,
		},
		{
			name:        "string equality attack",
			input:       "admin' OR 'x'='x' --",
			expectError: true,
		},
		{
			name:        "AND 1=1 attack",
			input:       "test' AND 1=1 --",
			expectError: true,
		},
		{
			name:        "waitfor delay attack",
			input:       "'; WAITFOR DELAY '0:0:5'",
			expectError: true,
		},
		{
			name:        "sleep attack",
			input:       "test' AND SLEEP(5)",
			expectError: true,
		},
		{
			name:        "benchmark attack",
			input:       "test' AND BENCHMARK(10000000,SHA1('test'))",
			expectError: true,
		},
		{
			name:        "information_schema access",
			input:       "SELECT * FROM information_schema.tables",
			expectError: true,
		},
		{
			name:        "hex encoding",
			input:       "0x48656c6c6f",
			expectError: true,
		},
		{
			name:        "null byte",
			input:       "test\x00value",
			expectError: true,
		},
		{
			name:        "percent encoded null",
			input:       "test%00value",
			expectError: true,
		},
		{
			name:        "char function",
			input:       "CHAR(65)+CHAR(66)",
			expectError: true,
		},
		{
			name:        "stacked query",
			input:       "1; SELECT * FROM users",
			expectError: true,
		},
		{
			name:        "stored procedure prefix",
			input:       "xp_cmdshell",
			expectError: true,
		},
		{
			name:        "system stored procedure",
			input:       "sp_executesql",
			expectError: true,
		},
		{
			name:        "concat function",
			input:       "CONCAT('admin','--')",
			expectError: true,
		},
		{
			name:        "pipe concat",
			input:       "admin'||'password",
			expectError: true,
		},
		{
			name:        "truncate table",
			input:       "'; TRUNCATE TABLE users",
			expectError: true,
		},
		{
			name:        "drop database",
			input:       "'; DROP DATABASE production",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateInput(tt.input)
			if tt.expectError && err == nil {
				t.Errorf("ValidateInput(%q) expected error, got nil", tt.input)
			}
			if !tt.expectError && err != nil {
				t.Errorf("ValidateInput(%q) unexpected error: %v", tt.input, err)
			}
		})
	}
}

func TestSQLValidator_IsSafe(t *testing.T) {
	validator := NewSQLValidator()

	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "safe input",
			input:    "Hello World",
			expected: true,
		},
		{
			name:     "unsafe input",
			input:    "'; DROP TABLE users; --",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.IsSafe(tt.input)
			if result != tt.expected {
				t.Errorf("IsSafe(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSQLValidator_ValidateInputs(t *testing.T) {
	validator := NewSQLValidator()

	// All safe inputs
	err := validator.ValidateInputs("hello", "world", "test123")
	if err != nil {
		t.Errorf("ValidateInputs with safe inputs returned error: %v", err)
	}

	// One unsafe input
	err = validator.ValidateInputs("hello", "'; DROP TABLE users", "test123")
	if err == nil {
		t.Error("ValidateInputs with unsafe input should return error")
	}
}

func TestIdentifierValidator(t *testing.T) {
	validator := NewIdentifierValidator()

	tests := []struct {
		name        string
		identifier  string
		expectError bool
	}{
		{
			name:        "valid identifier",
			identifier:  "users",
			expectError: false,
		},
		{
			name:        "valid with underscore",
			identifier:  "user_profiles",
			expectError: false,
		},
		{
			name:        "valid with number",
			identifier:  "user2",
			expectError: false,
		},
		{
			name:        "starts with underscore",
			identifier:  "_internal",
			expectError: false,
		},
		{
			name:        "empty string",
			identifier:  "",
			expectError: true,
		},
		{
			name:        "starts with number",
			identifier:  "123users",
			expectError: true,
		},
		{
			name:        "contains hyphen",
			identifier:  "user-profiles",
			expectError: true,
		},
		{
			name:        "contains space",
			identifier:  "user profiles",
			expectError: true,
		},
		{
			name:        "reserved word SELECT",
			identifier:  "select",
			expectError: true,
		},
		{
			name:        "reserved word DROP",
			identifier:  "drop",
			expectError: true,
		},
		{
			name:        "reserved word TABLE",
			identifier:  "table",
			expectError: true,
		},
		{
			name:        "contains SQL injection",
			identifier:  "users; DROP TABLE",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateIdentifier(tt.identifier)
			if tt.expectError && err == nil {
				t.Errorf("ValidateIdentifier(%q) expected error, got nil", tt.identifier)
			}
			if !tt.expectError && err != nil {
				t.Errorf("ValidateIdentifier(%q) unexpected error: %v", tt.identifier, err)
			}
		})
	}
}

func TestValidateOrderByColumn(t *testing.T) {
	allowedColumns := []string{"id", "name", "email", "created_at"}

	tests := []struct {
		name     string
		column   string
		expected bool
	}{
		{
			name:     "allowed column",
			column:   "name",
			expected: true,
		},
		{
			name:     "allowed column uppercase",
			column:   "NAME",
			expected: true,
		},
		{
			name:     "not allowed column",
			column:   "password",
			expected: false,
		},
		{
			name:     "empty column",
			column:   "",
			expected: false,
		},
		{
			name:     "sql injection attempt",
			column:   "name; DROP TABLE users",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateOrderByColumn(tt.column, allowedColumns)
			if result != tt.expected {
				t.Errorf("ValidateOrderByColumn(%q) = %v, want %v", tt.column, result, tt.expected)
			}
		})
	}
}

func TestValidateOrderDirection(t *testing.T) {
	tests := []struct {
		name      string
		direction string
		expected  bool
	}{
		{
			name:      "ASC",
			direction: "ASC",
			expected:  true,
		},
		{
			name:      "DESC",
			direction: "DESC",
			expected:  true,
		},
		{
			name:      "lowercase asc",
			direction: "asc",
			expected:  true,
		},
		{
			name:      "lowercase desc",
			direction: "desc",
			expected:  true,
		},
		{
			name:      "empty",
			direction: "",
			expected:  true,
		},
		{
			name:      "with whitespace",
			direction: "  ASC  ",
			expected:  true,
		},
		{
			name:      "invalid",
			direction: "RANDOM",
			expected:  false,
		},
		{
			name:      "sql injection",
			direction: "ASC; DROP TABLE",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateOrderDirection(tt.direction)
			if result != tt.expected {
				t.Errorf("ValidateOrderDirection(%q) = %v, want %v", tt.direction, result, tt.expected)
			}
		})
	}
}

func TestValidatePagination(t *testing.T) {
	tests := []struct {
		name     string
		page     int
		pageSize int
		expected bool
	}{
		{
			name:     "valid pagination",
			page:     0,
			pageSize: 10,
			expected: true,
		},
		{
			name:     "page 1",
			page:     1,
			pageSize: 20,
			expected: true,
		},
		{
			name:     "max page size",
			page:     0,
			pageSize: 1000,
			expected: true,
		},
		{
			name:     "negative page",
			page:     -1,
			pageSize: 10,
			expected: false,
		},
		{
			name:     "zero page size",
			page:     0,
			pageSize: 0,
			expected: false,
		},
		{
			name:     "negative page size",
			page:     0,
			pageSize: -10,
			expected: false,
		},
		{
			name:     "page size too large",
			page:     0,
			pageSize: 1001,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidatePagination(tt.page, tt.pageSize)
			if result != tt.expected {
				t.Errorf("ValidatePagination(%d, %d) = %v, want %v", tt.page, tt.pageSize, result, tt.expected)
			}
		})
	}
}

func TestSafeLimit(t *testing.T) {
	tests := []struct {
		name       string
		requested  int
		maxAllowed int
		expected   int
	}{
		{
			name:       "within limit",
			requested:  50,
			maxAllowed: 100,
			expected:   50,
		},
		{
			name:       "exceeds limit",
			requested:  150,
			maxAllowed: 100,
			expected:   100,
		},
		{
			name:       "zero requested",
			requested:  0,
			maxAllowed: 100,
			expected:   10,
		},
		{
			name:       "negative requested",
			requested:  -5,
			maxAllowed: 100,
			expected:   10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SafeLimit(tt.requested, tt.maxAllowed)
			if result != tt.expected {
				t.Errorf("SafeLimit(%d, %d) = %d, want %d", tt.requested, tt.maxAllowed, result, tt.expected)
			}
		})
	}
}

func TestSafeOffset(t *testing.T) {
	tests := []struct {
		name     string
		offset   int
		expected int
	}{
		{
			name:     "positive offset",
			offset:   10,
			expected: 10,
		},
		{
			name:     "zero offset",
			offset:   0,
			expected: 0,
		},
		{
			name:     "negative offset",
			offset:   -5,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SafeOffset(tt.offset)
			if result != tt.expected {
				t.Errorf("SafeOffset(%d) = %d, want %d", tt.offset, result, tt.expected)
			}
		})
	}
}

func TestSanitizeForSQL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal text",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "single quote",
			input:    "O'Brien",
			expected: "O''Brien",
		},
		{
			name:     "null byte",
			input:    "test\x00value",
			expected: "testvalue",
		},
		{
			name:     "percent null",
			input:    "test%00value",
			expected: "testvalue",
		},
		{
			name:     "SQL comment",
			input:    "test--comment",
			expected: "testcomment",
		},
		{
			name:     "block comment",
			input:    "test/*comment*/value",
			expected: "testcommentvalue",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeForSQL(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeForSQL(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
