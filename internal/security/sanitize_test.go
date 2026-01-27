package security

import (
	"testing"
)

func TestSanitizeString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		opts     SanitizeOptions
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			opts:     DefaultSanitizeOptions(),
			expected: "",
		},
		{
			name:     "simple text",
			input:    "Hello World",
			opts:     DefaultSanitizeOptions(),
			expected: "Hello World",
		},
		{
			name:     "HTML tags stripped",
			input:    "<script>alert('xss')</script>Hello",
			opts:     DefaultSanitizeOptions(),
			expected: "alert(&#39;xss&#39;)Hello",
		},
		{
			name:     "excessive whitespace normalized",
			input:    "Hello    World",
			opts:     DefaultSanitizeOptions(),
			expected: "Hello World",
		},
		{
			name:     "trim whitespace",
			input:    "   Hello World   ",
			opts:     DefaultSanitizeOptions(),
			expected: "Hello World",
		},
		{
			name:     "null byte removed",
			input:    "Hello\x00World",
			opts:     DefaultSanitizeOptions(),
			expected: "HelloWorld",
		},
		{
			name:     "control characters removed",
			input:    "Hello\x01\x02World",
			opts:     DefaultSanitizeOptions(),
			expected: "HelloWorld",
		},
		{
			name:     "HTML entities escaped",
			input:    "Hello <b>bold</b> & 'quotes'",
			opts:     SanitizeOptions{AllowHTML: true, EscapeHTML: true, TrimWhitespace: true},
			expected: "Hello &lt;b&gt;bold&lt;/b&gt; &amp; &#39;quotes&#39;",
		},
		{
			name:     "max length truncation",
			input:    "Hello World",
			opts:     SanitizeOptions{MaxLength: 5, TrimWhitespace: true},
			expected: "Hello",
		},
		{
			name:     "newlines preserved when allowed",
			input:    "Hello\nWorld",
			opts:     SanitizeOptions{AllowNewlines: true, TrimWhitespace: true},
			expected: "Hello\nWorld",
		},
		{
			name:     "newlines removed when not allowed",
			input:    "Hello\nWorld",
			opts:     SanitizeOptions{AllowNewlines: false, TrimWhitespace: true, NormalizeWhitespace: true},
			expected: "Hello World",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeString(tt.input, tt.opts)
			if result != tt.expected {
				t.Errorf("SanitizeString(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSanitizeEmail(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal email",
			input:    "test@example.com",
			expected: "test@example.com",
		},
		{
			name:     "uppercase email",
			input:    "Test@Example.COM",
			expected: "test@example.com",
		},
		{
			name:     "email with whitespace",
			input:    "  test@example.com  ",
			expected: "test@example.com",
		},
		{
			name:     "email with null byte",
			input:    "test\x00@example.com",
			expected: "test@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeEmail(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeEmail(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestDetectSQLInjection(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "normal input",
			input:    "Hello World",
			expected: false,
		},
		{
			name:     "SQL comment",
			input:    "test--comment",
			expected: true,
		},
		{
			name:     "UNION SELECT",
			input:    "1 UNION SELECT * FROM users",
			expected: true,
		},
		{
			name:     "DROP TABLE",
			input:    "'; DROP TABLE users; --",
			expected: true,
		},
		{
			name:     "INSERT INTO",
			input:    "test'; INSERT INTO users VALUES",
			expected: true,
		},
		{
			name:     "OR 1=1",
			input:    "admin' OR 1=1 --",
			expected: true,
		},
		{
			name:     "hex encoding",
			input:    "0x1234abcd",
			expected: true,
		},
		{
			name:     "normal number",
			input:    "12345",
			expected: false,
		},
		{
			name:     "email address",
			input:    "test@example.com",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectSQLInjection(tt.input)
			if result != tt.expected {
				t.Errorf("DetectSQLInjection(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestDetectXSS(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "normal input",
			input:    "Hello World",
			expected: false,
		},
		{
			name:     "script tag",
			input:    "<script>alert('xss')</script>",
			expected: true,
		},
		{
			name:     "javascript protocol",
			input:    "javascript:alert('xss')",
			expected: true,
		},
		{
			name:     "onclick event",
			input:    "<img onclick=alert('xss')>",
			expected: true,
		},
		{
			name:     "iframe",
			input:    "<iframe src='evil.com'>",
			expected: true,
		},
		{
			name:     "normal HTML allowed text",
			input:    "Hello <b>World</b>",
			expected: false,
		},
		{
			name:     "data URL",
			input:    "data:text/html,<script>alert(1)</script>",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectXSS(tt.input)
			if result != tt.expected {
				t.Errorf("DetectXSS(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestDetectPathTraversal(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "normal path",
			input:    "/home/user/file.txt",
			expected: false,
		},
		{
			name:     "parent directory",
			input:    "../etc/passwd",
			expected: true,
		},
		{
			name:     "encoded parent directory",
			input:    "%2e%2e%2fetc/passwd",
			expected: true,
		},
		{
			name:     "windows path traversal",
			input:    "..\\windows\\system32",
			expected: true,
		},
		{
			name:     "normal filename",
			input:    "document.pdf",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DetectPathTraversal(tt.input)
			if result != tt.expected {
				t.Errorf("DetectPathTraversal(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSanitizeURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal URL",
			input:    "https://example.com/page",
			expected: "https://example.com/page",
		},
		{
			name:     "URL with whitespace",
			input:    "  https://example.com/page  ",
			expected: "https://example.com/page",
		},
		{
			name:     "path traversal attack",
			input:    "https://example.com/../etc/passwd",
			expected: "",
		},
		{
			name:     "javascript URL",
			input:    "javascript:alert('xss')",
			expected: "",
		},
		{
			name:     "null byte in URL",
			input:    "https://example.com/\x00page",
			expected: "https://example.com/page",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeURL(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeURL(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal filename",
			input:    "document.pdf",
			expected: "document.pdf",
		},
		{
			name:     "path traversal",
			input:    "../secret.txt",
			expected: "secret.txt",
		},
		{
			name:     "path separator",
			input:    "path/to/file.txt",
			expected: "path_to_file.txt",
		},
		{
			name:     "null byte",
			input:    "file\x00.txt",
			expected: "file.txt",
		},
		{
			name:     "control characters",
			input:    "file\x01name.txt",
			expected: "filename.txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SanitizeFilename(tt.input)
			if result != tt.expected {
				t.Errorf("SanitizeFilename(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestValidateInputLength(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		minLen   int
		maxLen   int
		expected bool
	}{
		{
			name:     "valid length",
			input:    "hello",
			minLen:   3,
			maxLen:   10,
			expected: true,
		},
		{
			name:     "too short",
			input:    "hi",
			minLen:   3,
			maxLen:   10,
			expected: false,
		},
		{
			name:     "too long",
			input:    "hello world!",
			minLen:   3,
			maxLen:   10,
			expected: false,
		},
		{
			name:     "no min constraint",
			input:    "",
			minLen:   0,
			maxLen:   10,
			expected: true,
		},
		{
			name:     "no max constraint",
			input:    "very long string here",
			minLen:   3,
			maxLen:   0,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateInputLength(tt.input, tt.minLen, tt.maxLen)
			if result != tt.expected {
				t.Errorf("ValidateInputLength(%q, %d, %d) = %v, want %v", tt.input, tt.minLen, tt.maxLen, result, tt.expected)
			}
		})
	}
}

func TestStripHTMLTags(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "no tags",
			input:    "Hello World",
			expected: "Hello World",
		},
		{
			name:     "simple tags",
			input:    "<p>Hello</p> <b>World</b>",
			expected: "Hello World",
		},
		{
			name:     "script tag",
			input:    "<script>alert('xss')</script>Hello",
			expected: "alert('xss')Hello",
		},
		{
			name:     "nested tags",
			input:    "<div><p>Hello</p></div>",
			expected: "Hello",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := StripHTMLTags(tt.input)
			if result != tt.expected {
				t.Errorf("StripHTMLTags(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestIsPrintable(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "printable string",
			input:    "Hello World 123!",
			expected: true,
		},
		{
			name:     "with newline",
			input:    "Hello\nWorld",
			expected: true,
		},
		{
			name:     "with tab",
			input:    "Hello\tWorld",
			expected: true,
		},
		{
			name:     "with control char",
			input:    "Hello\x01World",
			expected: false,
		},
		{
			name:     "with null byte",
			input:    "Hello\x00World",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsPrintable(tt.input)
			if result != tt.expected {
				t.Errorf("IsPrintable(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
