package security

import (
	"encoding/base64"
	"encoding/json"
	"html"
	"net/url"
	"strings"
	"unicode/utf8"
)

// OutputEncoder provides safe encoding for different output contexts
type OutputEncoder struct{}

// NewOutputEncoder creates a new output encoder
func NewOutputEncoder() *OutputEncoder {
	return &OutputEncoder{}
}

// HTMLEncode encodes a string for safe HTML output
// This prevents XSS by escaping HTML special characters
func (e *OutputEncoder) HTMLEncode(input string) string {
	return html.EscapeString(input)
}

// HTMLDecode decodes HTML entities back to their original characters
func (e *OutputEncoder) HTMLDecode(input string) string {
	return html.UnescapeString(input)
}

// HTMLAttributeEncode encodes a string for safe use in HTML attributes
// More restrictive than general HTML encoding
func (e *OutputEncoder) HTMLAttributeEncode(input string) string {
	var builder strings.Builder
	for _, r := range input {
		switch r {
		case '&':
			builder.WriteString("&amp;")
		case '<':
			builder.WriteString("&lt;")
		case '>':
			builder.WriteString("&gt;")
		case '"':
			builder.WriteString("&quot;")
		case '\'':
			builder.WriteString("&#x27;")
		case '/':
			builder.WriteString("&#x2F;")
		case '`':
			builder.WriteString("&#x60;")
		case '=':
			builder.WriteString("&#x3D;")
		default:
			builder.WriteRune(r)
		}
	}
	return builder.String()
}

// URLEncode encodes a string for safe use in URLs
func (e *OutputEncoder) URLEncode(input string) string {
	return url.QueryEscape(input)
}

// URLDecode decodes a URL-encoded string
func (e *OutputEncoder) URLDecode(input string) (string, error) {
	return url.QueryUnescape(input)
}

// URLPathEncode encodes a string for safe use in URL paths
func (e *OutputEncoder) URLPathEncode(input string) string {
	return url.PathEscape(input)
}

// URLPathDecode decodes a URL path-encoded string
func (e *OutputEncoder) URLPathDecode(input string) (string, error) {
	return url.PathUnescape(input)
}

// JavaScriptEncode encodes a string for safe use in JavaScript contexts
// This should be used when embedding data in inline JavaScript
func (e *OutputEncoder) JavaScriptEncode(input string) string {
	var builder strings.Builder
	for _, r := range input {
		switch r {
		case '\\':
			builder.WriteString("\\\\")
		case '\'':
			builder.WriteString("\\'")
		case '"':
			builder.WriteString("\\\"")
		case '\n':
			builder.WriteString("\\n")
		case '\r':
			builder.WriteString("\\r")
		case '\t':
			builder.WriteString("\\t")
		case '<':
			builder.WriteString("\\u003c")
		case '>':
			builder.WriteString("\\u003e")
		case '&':
			builder.WriteString("\\u0026")
		case '=':
			builder.WriteString("\\u003d")
		case '/':
			builder.WriteString("\\/")
		default:
			if r < 32 || r > 126 {
				// Encode non-ASCII and control characters
				builder.WriteString("\\u")
				builder.WriteString(strings.ToLower(strings.Repeat("0", 4-len([]rune(string(r)))+1)))
				for _, b := range []byte(string(r)) {
					builder.WriteString(strings.ToLower(string("0123456789abcdef"[b>>4])))
					builder.WriteString(strings.ToLower(string("0123456789abcdef"[b&0xf])))
				}
			} else {
				builder.WriteRune(r)
			}
		}
	}
	return builder.String()
}

// CSSEncode encodes a string for safe use in CSS contexts
func (e *OutputEncoder) CSSEncode(input string) string {
	var builder strings.Builder
	for _, r := range input {
		// Only allow alphanumeric characters without encoding
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') {
			builder.WriteRune(r)
		} else {
			// CSS escape: \HH (hexadecimal)
			builder.WriteString("\\")
			builder.WriteString(strings.ToLower(strings.Repeat("0", 6-len([]rune(string(r)))+1)))
			for _, b := range []byte(string(r)) {
				builder.WriteString(strings.ToLower(string("0123456789abcdef"[b>>4])))
				builder.WriteString(strings.ToLower(string("0123456789abcdef"[b&0xf])))
			}
			builder.WriteString(" ")
		}
	}
	return builder.String()
}

// Base64Encode encodes data to base64
func (e *OutputEncoder) Base64Encode(input []byte) string {
	return base64.StdEncoding.EncodeToString(input)
}

// Base64Decode decodes base64 data
func (e *OutputEncoder) Base64Decode(input string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(input)
}

// Base64URLEncode encodes data to URL-safe base64
func (e *OutputEncoder) Base64URLEncode(input []byte) string {
	return base64.URLEncoding.EncodeToString(input)
}

// Base64URLDecode decodes URL-safe base64 data
func (e *OutputEncoder) Base64URLDecode(input string) ([]byte, error) {
	return base64.URLEncoding.DecodeString(input)
}

// SafeJSONResponse represents a response with proper JSON encoding
type SafeJSONResponse struct {
	encoder *OutputEncoder
}

// NewSafeJSONResponse creates a new safe JSON response builder
func NewSafeJSONResponse() *SafeJSONResponse {
	return &SafeJSONResponse{
		encoder: NewOutputEncoder(),
	}
}

// Encode encodes a value to safe JSON with HTML-safe escaping
func (s *SafeJSONResponse) Encode(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

// EncodeHTML encodes to JSON with HTML entities escaped for embedding in HTML
func (s *SafeJSONResponse) EncodeHTML(v interface{}) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	// The json.Marshal already escapes < > & for HTML safety
	return string(data), nil
}

// SafeResponseHeaders sets security headers for JSON API responses
type SafeResponseHeaders struct {
	// ContentType sets the Content-Type header
	ContentType string
	// NoSniff enables X-Content-Type-Options: nosniff
	NoSniff bool
	// NoCache enables cache prevention headers
	NoCache bool
}

// DefaultSafeResponseHeaders returns safe default headers for API responses
func DefaultSafeResponseHeaders() *SafeResponseHeaders {
	return &SafeResponseHeaders{
		ContentType: "application/json; charset=utf-8",
		NoSniff:     true,
		NoCache:     true,
	}
}

// ValidateUTF8 checks if a string is valid UTF-8
func ValidateUTF8(input string) bool {
	return utf8.ValidString(input)
}

// SanitizeUTF8 removes invalid UTF-8 sequences
func SanitizeUTF8(input string) string {
	if utf8.ValidString(input) {
		return input
	}

	// Replace invalid sequences with replacement character
	var builder strings.Builder
	for len(input) > 0 {
		r, size := utf8.DecodeRuneInString(input)
		if r == utf8.RuneError && size == 1 {
			builder.WriteRune(utf8.RuneError)
			input = input[1:]
		} else {
			builder.WriteRune(r)
			input = input[size:]
		}
	}
	return builder.String()
}

// Convenience functions for common encoding operations

// HTMLEncode is a convenience function for HTML encoding
func HTMLEncode(input string) string {
	return html.EscapeString(input)
}

// HTMLDecode is a convenience function for HTML decoding
func HTMLDecode(input string) string {
	return html.UnescapeString(input)
}

// URLEncode is a convenience function for URL encoding
func URLEncode(input string) string {
	return url.QueryEscape(input)
}

// URLDecode is a convenience function for URL decoding
func URLDecode(input string) (string, error) {
	return url.QueryUnescape(input)
}

// EncodeForContext encodes a string based on the output context
type OutputContext int

const (
	ContextHTML OutputContext = iota
	ContextHTMLAttribute
	ContextJavaScript
	ContextCSS
	ContextURL
	ContextURLPath
	ContextJSON
)

// EncodeForContext encodes input for the specified output context
func EncodeForContext(input string, ctx OutputContext) string {
	encoder := NewOutputEncoder()
	switch ctx {
	case ContextHTML:
		return encoder.HTMLEncode(input)
	case ContextHTMLAttribute:
		return encoder.HTMLAttributeEncode(input)
	case ContextJavaScript:
		return encoder.JavaScriptEncode(input)
	case ContextCSS:
		return encoder.CSSEncode(input)
	case ContextURL:
		return encoder.URLEncode(input)
	case ContextURLPath:
		return encoder.URLPathEncode(input)
	case ContextJSON:
		// For JSON, we rely on json.Marshal which handles escaping
		data, _ := json.Marshal(input)
		return string(data)
	default:
		return encoder.HTMLEncode(input) // Default to HTML encoding
	}
}
