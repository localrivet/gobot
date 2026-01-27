package security

import (
	"errors"
	"regexp"
	"strings"
)

var (
	// ErrSQLInjectionDetected is returned when SQL injection is detected
	ErrSQLInjectionDetected = errors.New("potential SQL injection detected")
	// ErrInvalidInput is returned when input validation fails
	ErrInvalidInput = errors.New("invalid input")
)

// SQLInjectionPatterns contains patterns that indicate potential SQL injection
var SQLInjectionPatterns = []string{
	// SQL comments
	`--`,
	`/\*`,
	`\*/`,
	`#`,
	// SQL keywords with context
	`(?i)\bunion\s+select\b`,
	`(?i)\bselect\s+.*\s+from\b`,
	`(?i)\binsert\s+into\b`,
	`(?i)\bdelete\s+from\b`,
	`(?i)\bdrop\s+table\b`,
	`(?i)\bdrop\s+database\b`,
	`(?i)\btruncate\s+table\b`,
	`(?i)\bupdate\s+.*\s+set\b`,
	`(?i)\bexec\s*\(`,
	`(?i)\bexecute\s*\(`,
	`(?i)\bxp_`,
	`(?i)\bsp_`,
	// String concatenation attacks
	`(?i)\bconcat\s*\(`,
	`\|\|`,
	// Boolean-based injection
	`(?i)\bor\s+1\s*=\s*1\b`,
	`(?i)\bor\s+'.*'\s*=\s*'.*'\b`,
	`(?i)\band\s+1\s*=\s*1\b`,
	`(?i)\band\s+'.*'\s*=\s*'.*'\b`,
	// Time-based injection
	`(?i)\bwaitfor\s+delay\b`,
	`(?i)\bsleep\s*\(`,
	`(?i)\bbenchmark\s*\(`,
	// Information gathering
	`(?i)\binformation_schema\b`,
	`(?i)\bsys\.`,
	`(?i)\bsysobjects\b`,
	`(?i)\bsyscolumns\b`,
	// LIKE abuse
	`(?i)\blike\s+['"]%`,
	// Hex encoding
	`0x[0-9a-fA-F]+`,
	// Null byte injection
	`\x00`,
	`%00`,
	// Character encoding attacks
	`(?i)char\s*\(\s*\d+\s*\)`,
	// Stacked queries
	`;\s*(?i)(select|insert|update|delete|drop|truncate|exec|execute)`,
}

var sqlInjectionRegexes []*regexp.Regexp

func init() {
	sqlInjectionRegexes = make([]*regexp.Regexp, len(SQLInjectionPatterns))
	for i, pattern := range SQLInjectionPatterns {
		sqlInjectionRegexes[i] = regexp.MustCompile(pattern)
	}
}

// SQLValidator provides SQL injection validation
type SQLValidator struct {
	// AdditionalPatterns allows adding custom patterns
	AdditionalPatterns []*regexp.Regexp
	// AllowedPatterns are exceptions to the default rules
	AllowedPatterns []string
	// MaxInputLength is the maximum allowed input length
	MaxInputLength int
}

// NewSQLValidator creates a new SQL validator with default settings
func NewSQLValidator() *SQLValidator {
	return &SQLValidator{
		MaxInputLength: 10000,
	}
}

// ValidateInput checks if the input contains potential SQL injection patterns
func (v *SQLValidator) ValidateInput(input string) error {
	if input == "" {
		return nil
	}

	// Check length
	if v.MaxInputLength > 0 && len(input) > v.MaxInputLength {
		return ErrInvalidInput
	}

	// Check for null bytes
	if strings.Contains(input, "\x00") || strings.Contains(input, "%00") {
		return ErrSQLInjectionDetected
	}

	// Check default patterns
	for _, regex := range sqlInjectionRegexes {
		if regex.MatchString(input) {
			return ErrSQLInjectionDetected
		}
	}

	// Check additional patterns
	for _, regex := range v.AdditionalPatterns {
		if regex.MatchString(input) {
			return ErrSQLInjectionDetected
		}
	}

	return nil
}

// ValidateInputs validates multiple inputs
func (v *SQLValidator) ValidateInputs(inputs ...string) error {
	for _, input := range inputs {
		if err := v.ValidateInput(input); err != nil {
			return err
		}
	}
	return nil
}

// IsSafe checks if input is safe (no SQL injection patterns detected)
func (v *SQLValidator) IsSafe(input string) bool {
	return v.ValidateInput(input) == nil
}

// ValidateSQLInput validates input for SQL safety
func ValidateSQLInput(input string) error {
	validator := NewSQLValidator()
	return validator.ValidateInput(input)
}

// SanitizeForSQL removes potentially dangerous characters for SQL
// Note: This should be used in ADDITION to parameterized queries, not as a replacement
func SanitizeForSQL(input string) string {
	// Remove null bytes
	result := strings.ReplaceAll(input, "\x00", "")
	result = strings.ReplaceAll(result, "%00", "")

	// Remove SQL comment indicators
	result = strings.ReplaceAll(result, "--", "")
	result = strings.ReplaceAll(result, "/*", "")
	result = strings.ReplaceAll(result, "*/", "")

	// Escape single quotes by doubling them
	result = strings.ReplaceAll(result, "'", "''")

	return result
}

// IdentifierValidator validates SQL identifiers (table names, column names)
type IdentifierValidator struct {
	// AllowedChars is a regex pattern for allowed characters in identifiers
	AllowedChars *regexp.Regexp
	// MaxLength is the maximum length for identifiers
	MaxLength int
	// ReservedWords are SQL reserved words that cannot be used as identifiers
	ReservedWords map[string]bool
}

// NewIdentifierValidator creates a new identifier validator
func NewIdentifierValidator() *IdentifierValidator {
	reservedWords := map[string]bool{
		"select": true, "insert": true, "update": true, "delete": true,
		"drop": true, "create": true, "alter": true, "truncate": true,
		"from": true, "where": true, "and": true, "or": true,
		"not": true, "in": true, "like": true, "between": true,
		"join": true, "inner": true, "outer": true, "left": true,
		"right": true, "cross": true, "on": true, "group": true,
		"by": true, "order": true, "having": true, "limit": true,
		"offset": true, "union": true, "except": true, "intersect": true,
		"all": true, "distinct": true, "as": true, "null": true,
		"true": true, "false": true, "table": true, "index": true,
		"database": true, "schema": true, "grant": true, "revoke": true,
	}

	return &IdentifierValidator{
		AllowedChars:  regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`),
		MaxLength:     128,
		ReservedWords: reservedWords,
	}
}

// ValidateIdentifier validates a SQL identifier
func (v *IdentifierValidator) ValidateIdentifier(identifier string) error {
	if identifier == "" {
		return ErrInvalidInput
	}

	if len(identifier) > v.MaxLength {
		return ErrInvalidInput
	}

	if !v.AllowedChars.MatchString(identifier) {
		return ErrInvalidInput
	}

	if v.ReservedWords[strings.ToLower(identifier)] {
		return ErrInvalidInput
	}

	return nil
}

// IsValidIdentifier checks if an identifier is valid
func (v *IdentifierValidator) IsValidIdentifier(identifier string) bool {
	return v.ValidateIdentifier(identifier) == nil
}

// ValidateOrderByColumn validates a column name used in ORDER BY
// This is important as ORDER BY cannot be parameterized in most databases
func ValidateOrderByColumn(column string, allowedColumns []string) bool {
	if column == "" {
		return false
	}

	// Normalize the column name
	column = strings.ToLower(strings.TrimSpace(column))

	// Check against allowed list
	for _, allowed := range allowedColumns {
		if strings.ToLower(allowed) == column {
			return true
		}
	}

	return false
}

// ValidateOrderDirection validates ORDER BY direction
func ValidateOrderDirection(direction string) bool {
	direction = strings.ToUpper(strings.TrimSpace(direction))
	return direction == "ASC" || direction == "DESC" || direction == ""
}

// ValidatePagination validates pagination parameters
func ValidatePagination(page, pageSize int) bool {
	return page >= 0 && pageSize > 0 && pageSize <= 1000
}

// SafeLimit returns a safe limit value within bounds
func SafeLimit(requested, maxAllowed int) int {
	if requested <= 0 {
		return 10 // default
	}
	if requested > maxAllowed {
		return maxAllowed
	}
	return requested
}

// SafeOffset returns a safe offset value
func SafeOffset(offset int) int {
	if offset < 0 {
		return 0
	}
	return offset
}
