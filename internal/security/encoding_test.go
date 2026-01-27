package security

import (
	"testing"
)

func TestOutputEncoder_HTMLEncode(t *testing.T) {
	encoder := NewOutputEncoder()

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
			name:     "less than",
			input:    "a < b",
			expected: "a &lt; b",
		},
		{
			name:     "greater than",
			input:    "a > b",
			expected: "a &gt; b",
		},
		{
			name:     "ampersand",
			input:    "a & b",
			expected: "a &amp; b",
		},
		{
			name:     "double quote",
			input:    `a "b" c`,
			expected: "a &#34;b&#34; c",
		},
		{
			name:     "script tag",
			input:    "<script>alert('xss')</script>",
			expected: "&lt;script&gt;alert(&#39;xss&#39;)&lt;/script&gt;",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := encoder.HTMLEncode(tt.input)
			if result != tt.expected {
				t.Errorf("HTMLEncode(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestOutputEncoder_HTMLAttributeEncode(t *testing.T) {
	encoder := NewOutputEncoder()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal text",
			input:    "Hello",
			expected: "Hello",
		},
		{
			name:     "single quote",
			input:    "it's",
			expected: "it&#x27;s",
		},
		{
			name:     "double quote",
			input:    `say "hi"`,
			expected: "say &quot;hi&quot;",
		},
		{
			name:     "backtick",
			input:    "`test`",
			expected: "&#x60;test&#x60;",
		},
		{
			name:     "equals",
			input:    "a=b",
			expected: "a&#x3D;b",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := encoder.HTMLAttributeEncode(tt.input)
			if result != tt.expected {
				t.Errorf("HTMLAttributeEncode(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestOutputEncoder_URLEncode(t *testing.T) {
	encoder := NewOutputEncoder()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "normal text",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "space",
			input:    "hello world",
			expected: "hello+world",
		},
		{
			name:     "special characters",
			input:    "a=b&c=d",
			expected: "a%3Db%26c%3Dd",
		},
		{
			name:     "slash",
			input:    "path/to/file",
			expected: "path%2Fto%2Ffile",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := encoder.URLEncode(tt.input)
			if result != tt.expected {
				t.Errorf("URLEncode(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestOutputEncoder_URLDecode(t *testing.T) {
	encoder := NewOutputEncoder()

	tests := []struct {
		name      string
		input     string
		expected  string
		expectErr bool
	}{
		{
			name:     "normal text",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "encoded space",
			input:    "hello+world",
			expected: "hello world",
		},
		{
			name:     "percent encoded",
			input:    "hello%20world",
			expected: "hello world",
		},
		{
			name:      "invalid encoding",
			input:     "hello%",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := encoder.URLDecode(tt.input)
			if tt.expectErr {
				if err == nil {
					t.Errorf("URLDecode(%q) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("URLDecode(%q) unexpected error: %v", tt.input, err)
				return
			}
			if result != tt.expected {
				t.Errorf("URLDecode(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestOutputEncoder_Base64Encode(t *testing.T) {
	encoder := NewOutputEncoder()

	tests := []struct {
		name     string
		input    []byte
		expected string
	}{
		{
			name:     "simple text",
			input:    []byte("hello"),
			expected: "aGVsbG8=",
		},
		{
			name:     "binary data",
			input:    []byte{0x00, 0x01, 0x02},
			expected: "AAEC",
		},
		{
			name:     "empty",
			input:    []byte{},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := encoder.Base64Encode(tt.input)
			if result != tt.expected {
				t.Errorf("Base64Encode(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestOutputEncoder_Base64Decode(t *testing.T) {
	encoder := NewOutputEncoder()

	tests := []struct {
		name      string
		input     string
		expected  []byte
		expectErr bool
	}{
		{
			name:     "valid base64",
			input:    "aGVsbG8=",
			expected: []byte("hello"),
		},
		{
			name:      "invalid base64",
			input:     "!!!invalid!!!",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := encoder.Base64Decode(tt.input)
			if tt.expectErr {
				if err == nil {
					t.Errorf("Base64Decode(%q) expected error, got nil", tt.input)
				}
				return
			}
			if err != nil {
				t.Errorf("Base64Decode(%q) unexpected error: %v", tt.input, err)
				return
			}
			if string(result) != string(tt.expected) {
				t.Errorf("Base64Decode(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestValidateUTF8(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid ASCII",
			input:    "Hello World",
			expected: true,
		},
		{
			name:     "valid UTF-8",
			input:    "Hello 世界",
			expected: true,
		},
		{
			name:     "valid emoji",
			input:    "Hello 👋",
			expected: true,
		},
		{
			name:     "invalid UTF-8",
			input:    string([]byte{0xff, 0xfe}),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidateUTF8(tt.input)
			if result != tt.expected {
				t.Errorf("ValidateUTF8(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSanitizeUTF8(t *testing.T) {
	// Valid UTF-8 should pass through unchanged
	valid := "Hello World 世界"
	result := SanitizeUTF8(valid)
	if result != valid {
		t.Errorf("SanitizeUTF8(%q) = %q, want %q", valid, result, valid)
	}

	// Invalid UTF-8 should be fixed
	invalid := string([]byte{0xff, 0xfe})
	result = SanitizeUTF8(invalid)
	if len(result) == 0 {
		t.Error("SanitizeUTF8 should produce output for invalid UTF-8")
	}
}

func TestEncodeForContext(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		context  OutputContext
		expected string
	}{
		{
			name:     "HTML context",
			input:    "<script>",
			context:  ContextHTML,
			expected: "&lt;script&gt;",
		},
		{
			name:     "URL context",
			input:    "hello world",
			context:  ContextURL,
			expected: "hello+world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := EncodeForContext(tt.input, tt.context)
			if result != tt.expected {
				t.Errorf("EncodeForContext(%q, %v) = %q, want %q", tt.input, tt.context, result, tt.expected)
			}
		})
	}
}

func TestHTMLEncode_Convenience(t *testing.T) {
	input := "<script>"
	expected := "&lt;script&gt;"
	result := HTMLEncode(input)
	if result != expected {
		t.Errorf("HTMLEncode(%q) = %q, want %q", input, result, expected)
	}
}

func TestHTMLDecode_Convenience(t *testing.T) {
	input := "&lt;script&gt;"
	expected := "<script>"
	result := HTMLDecode(input)
	if result != expected {
		t.Errorf("HTMLDecode(%q) = %q, want %q", input, result, expected)
	}
}

func TestURLEncode_Convenience(t *testing.T) {
	input := "hello world"
	expected := "hello+world"
	result := URLEncode(input)
	if result != expected {
		t.Errorf("URLEncode(%q) = %q, want %q", input, result, expected)
	}
}

func TestURLDecode_Convenience(t *testing.T) {
	input := "hello+world"
	expected := "hello world"
	result, err := URLDecode(input)
	if err != nil {
		t.Errorf("URLDecode(%q) unexpected error: %v", input, err)
		return
	}
	if result != expected {
		t.Errorf("URLDecode(%q) = %q, want %q", input, result, expected)
	}
}
