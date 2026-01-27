package security

import (
	"html"
	"regexp"
	"strings"
	"unicode"
)

var (
	// Patterns for common attack vectors
	sqlInjectionPattern    = regexp.MustCompile(`(?i)(--|;|'|"|\\|/\*|\*/|xp_|sp_|0x|union\s+select|select\s+.*\s+from|insert\s+into|delete\s+from|drop\s+table|update\s+.*\s+set|exec\s*\(|execute\s*\()`)
	xssPattern             = regexp.MustCompile(`(?i)<script[^>]*>|</script>|javascript:|on\w+\s*=|<iframe|<object|<embed|<form|<input|<button|data:text/html|vbscript:`)
	pathTraversalPattern   = regexp.MustCompile(`(?i)\.\.\/|\.\.\\|%2e%2e%2f|%2e%2e\/|\.\.%2f|%2e%2e%5c`)
	nullBytePattern        = regexp.MustCompile(`\x00|%00`)
	excessiveWhitespace    = regexp.MustCompile(`[\s\p{Zs}]{2,}`)
	htmlTagPattern         = regexp.MustCompile(`<[^>]*>`)
	multipleNewlines       = regexp.MustCompile(`\n{3,}`)
	controlCharsPattern    = regexp.MustCompile(`[\x00-\x08\x0B\x0C\x0E-\x1F\x7F]`)

	// URL validation
	safeURLSchemes = map[string]bool{
		"http":  true,
		"https": true,
		"mailto": true,
	}
)

// SanitizeOptions configures sanitization behavior
type SanitizeOptions struct {
	// AllowHTML preserves HTML tags (use with caution)
	AllowHTML bool
	// TrimWhitespace removes leading/trailing whitespace
	TrimWhitespace bool
	// NormalizeWhitespace collapses multiple spaces to single space
	NormalizeWhitespace bool
	// MaxLength truncates string to maximum length (0 = no limit)
	MaxLength int
	// AllowNewlines preserves newline characters
	AllowNewlines bool
	// StripControlChars removes non-printable control characters
	StripControlChars bool
	// EscapeHTML converts HTML special characters to entities
	EscapeHTML bool
}

// DefaultSanitizeOptions returns safe default options
func DefaultSanitizeOptions() SanitizeOptions {
	return SanitizeOptions{
		AllowHTML:           false,
		TrimWhitespace:      true,
		NormalizeWhitespace: true,
		MaxLength:           0,
		AllowNewlines:       true,
		StripControlChars:   true,
		EscapeHTML:          true,
	}
}

// StrictSanitizeOptions returns very restrictive options
func StrictSanitizeOptions() SanitizeOptions {
	return SanitizeOptions{
		AllowHTML:           false,
		TrimWhitespace:      true,
		NormalizeWhitespace: true,
		MaxLength:           10000,
		AllowNewlines:       false,
		StripControlChars:   true,
		EscapeHTML:          true,
	}
}

// SanitizeString sanitizes a string input with the given options
func SanitizeString(input string, opts SanitizeOptions) string {
	if input == "" {
		return ""
	}

	result := input

	// Remove null bytes first (critical security issue)
	result = nullBytePattern.ReplaceAllString(result, "")

	// Strip control characters
	if opts.StripControlChars {
		result = controlCharsPattern.ReplaceAllString(result, "")
	}

	// Handle HTML
	if !opts.AllowHTML {
		result = htmlTagPattern.ReplaceAllString(result, "")
	}

	// Escape HTML entities
	if opts.EscapeHTML {
		result = html.EscapeString(result)
	}

	// Handle whitespace
	if opts.TrimWhitespace {
		result = strings.TrimSpace(result)
	}

	if opts.NormalizeWhitespace {
		if opts.AllowNewlines {
			// Normalize spaces but preserve newlines
			lines := strings.Split(result, "\n")
			for i, line := range lines {
				lines[i] = excessiveWhitespace.ReplaceAllString(line, " ")
				lines[i] = strings.TrimSpace(lines[i])
			}
			result = strings.Join(lines, "\n")
			// Collapse excessive newlines
			result = multipleNewlines.ReplaceAllString(result, "\n\n")
		} else {
			result = excessiveWhitespace.ReplaceAllString(result, " ")
		}
	}

	// Handle newlines
	if !opts.AllowNewlines {
		result = strings.ReplaceAll(result, "\n", " ")
		result = strings.ReplaceAll(result, "\r", "")
	}

	// Truncate if necessary
	if opts.MaxLength > 0 && len(result) > opts.MaxLength {
		result = result[:opts.MaxLength]
	}

	return result
}

// SanitizeEmail sanitizes and validates an email address
func SanitizeEmail(email string) string {
	email = strings.TrimSpace(email)
	email = strings.ToLower(email)

	// Remove any null bytes or control characters
	email = nullBytePattern.ReplaceAllString(email, "")
	email = controlCharsPattern.ReplaceAllString(email, "")

	return email
}

// SanitizeName sanitizes a name field (first name, last name, etc.)
func SanitizeName(name string) string {
	opts := SanitizeOptions{
		AllowHTML:           false,
		TrimWhitespace:      true,
		NormalizeWhitespace: true,
		MaxLength:           100,
		AllowNewlines:       false,
		StripControlChars:   true,
		EscapeHTML:          true,
	}
	return SanitizeString(name, opts)
}

// SanitizeURL sanitizes a URL string
func SanitizeURL(url string) string {
	url = strings.TrimSpace(url)
	url = nullBytePattern.ReplaceAllString(url, "")
	url = controlCharsPattern.ReplaceAllString(url, "")

	// Check for path traversal attempts
	if pathTraversalPattern.MatchString(url) {
		return ""
	}

	// Check for JavaScript/XSS attempts in URL
	if xssPattern.MatchString(url) {
		return ""
	}

	return url
}

// SanitizeFilename sanitizes a filename for safe storage
func SanitizeFilename(filename string) string {
	filename = strings.TrimSpace(filename)

	// Remove null bytes
	filename = nullBytePattern.ReplaceAllString(filename, "")

	// Remove path traversal attempts
	filename = pathTraversalPattern.ReplaceAllString(filename, "")

	// Remove path separators
	filename = strings.ReplaceAll(filename, "/", "_")
	filename = strings.ReplaceAll(filename, "\\", "_")

	// Remove control characters
	filename = controlCharsPattern.ReplaceAllString(filename, "")

	// Limit length
	if len(filename) > 255 {
		filename = filename[:255]
	}

	return filename
}

// DetectSQLInjection checks for potential SQL injection patterns
func DetectSQLInjection(input string) bool {
	return sqlInjectionPattern.MatchString(input)
}

// DetectXSS checks for potential XSS patterns
func DetectXSS(input string) bool {
	return xssPattern.MatchString(input)
}

// DetectPathTraversal checks for path traversal attempts
func DetectPathTraversal(input string) bool {
	return pathTraversalPattern.MatchString(input)
}

// ContainsNullByte checks if input contains null bytes
func ContainsNullByte(input string) bool {
	return nullBytePattern.MatchString(input)
}

// IsValidURLScheme checks if the URL has a safe scheme
func IsValidURLScheme(url string) bool {
	url = strings.ToLower(strings.TrimSpace(url))
	for scheme := range safeURLSchemes {
		if strings.HasPrefix(url, scheme+"://") || strings.HasPrefix(url, scheme+":") {
			return true
		}
	}
	return false
}

// IsPrintable checks if all characters in the string are printable
func IsPrintable(s string) bool {
	for _, r := range s {
		if !unicode.IsPrint(r) && r != '\n' && r != '\r' && r != '\t' {
			return false
		}
	}
	return true
}

// StripHTMLTags removes all HTML tags from a string
func StripHTMLTags(input string) string {
	return htmlTagPattern.ReplaceAllString(input, "")
}

// EscapeHTML escapes HTML special characters
func EscapeHTML(input string) string {
	return html.EscapeString(input)
}

// UnescapeHTML unescapes HTML entities
func UnescapeHTML(input string) string {
	return html.UnescapeString(input)
}

// SanitizeForLog sanitizes a string for safe logging (prevents log injection)
func SanitizeForLog(input string) string {
	// Remove newlines to prevent log injection
	result := strings.ReplaceAll(input, "\n", "\\n")
	result = strings.ReplaceAll(result, "\r", "\\r")

	// Remove control characters
	result = controlCharsPattern.ReplaceAllString(result, "")

	// Truncate long strings
	if len(result) > 1000 {
		result = result[:1000] + "...[truncated]"
	}

	return result
}

// SanitizeJSON sanitizes a string for safe JSON output
func SanitizeJSON(input string) string {
	// Remove null bytes
	result := nullBytePattern.ReplaceAllString(input, "")

	// Remove control characters except common whitespace
	var builder strings.Builder
	for _, r := range result {
		if unicode.IsPrint(r) || r == '\n' || r == '\r' || r == '\t' {
			builder.WriteRune(r)
		}
	}

	return builder.String()
}

// ValidateInputLength checks if input length is within bounds
func ValidateInputLength(input string, minLen, maxLen int) bool {
	length := len(input)
	if minLen > 0 && length < minLen {
		return false
	}
	if maxLen > 0 && length > maxLen {
		return false
	}
	return true
}

// TruncateString safely truncates a string to max length
func TruncateString(input string, maxLen int) string {
	if len(input) <= maxLen {
		return input
	}
	return input[:maxLen]
}
