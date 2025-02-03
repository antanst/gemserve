package common

import (
	"net/url"
	"reflect"
	"testing"
)

func TestParseURL(t *testing.T) {
	t.Parallel()
	input := "gemini://caolan.uk/cgi-bin/weather.py/wxfcs/3162"
	parsed, err := ParseURL(input, "", true)
	value, _ := parsed.Value()
	if err != nil || !(value == "gemini://caolan.uk:1965/cgi-bin/weather.py/wxfcs/3162") {
		t.Errorf("fail: %s", parsed)
	}
}

func TestDeriveAbsoluteURL_abs_url_input(t *testing.T) {
	t.Parallel()
	currentURL := URL{
		Protocol: "gemini",
		Hostname: "smol.gr",
		Port:     1965,
		Path:     "/a/b",
		Descr:    "Nothing",
		Full:     "gemini://smol.gr:1965/a/b",
	}
	input := "gemini://a.b/c"
	output, err := DeriveAbsoluteURL(currentURL, input)
	if err != nil {
		t.Errorf("fail: %v", err)
	}
	expected := &URL{
		Protocol: "gemini",
		Hostname: "a.b",
		Port:     1965,
		Path:     "/c",
		Descr:    "",
		Full:     "gemini://a.b:1965/c",
	}
	pass := reflect.DeepEqual(output, expected)
	if !pass {
		t.Errorf("fail: %#v != %#v", output, expected)
	}
}

func TestDeriveAbsoluteURL_abs_path_input(t *testing.T) {
	t.Parallel()
	currentURL := URL{
		Protocol: "gemini",
		Hostname: "smol.gr",
		Port:     1965,
		Path:     "/a/b",
		Descr:    "Nothing",
		Full:     "gemini://smol.gr:1965/a/b",
	}
	input := "/c"
	output, err := DeriveAbsoluteURL(currentURL, input)
	if err != nil {
		t.Errorf("fail: %v", err)
	}
	expected := &URL{
		Protocol: "gemini",
		Hostname: "smol.gr",
		Port:     1965,
		Path:     "/c",
		Descr:    "",
		Full:     "gemini://smol.gr:1965/c",
	}
	pass := reflect.DeepEqual(output, expected)
	if !pass {
		t.Errorf("fail: %#v != %#v", output, expected)
	}
}

func TestDeriveAbsoluteURL_rel_path_input(t *testing.T) {
	t.Parallel()
	currentURL := URL{
		Protocol: "gemini",
		Hostname: "smol.gr",
		Port:     1965,
		Path:     "/a/b",
		Descr:    "Nothing",
		Full:     "gemini://smol.gr:1965/a/b",
	}
	input := "c/d"
	output, err := DeriveAbsoluteURL(currentURL, input)
	if err != nil {
		t.Errorf("fail: %v", err)
	}
	expected := &URL{
		Protocol: "gemini",
		Hostname: "smol.gr",
		Port:     1965,
		Path:     "/a/b/c/d",
		Descr:    "",
		Full:     "gemini://smol.gr:1965/a/b/c/d",
	}
	pass := reflect.DeepEqual(output, expected)
	if !pass {
		t.Errorf("fail: %#v != %#v", output, expected)
	}
}

func TestNormalizeURLSlash(t *testing.T) {
	t.Parallel()
	input := "gemini://uscoffings.net/retro-computing/magazines/"
	normalized, _ := NormalizeURL(input)
	output := normalized.String()
	expected := input
	pass := reflect.DeepEqual(output, expected)
	if !pass {
		t.Errorf("fail: %#v != %#v", output, expected)
	}
}

func TestNormalizeURLNoSlash(t *testing.T) {
	t.Parallel()
	input := "gemini://uscoffings.net/retro-computing/magazines"
	normalized, _ := NormalizeURL(input)
	output := normalized.String()
	expected := input
	pass := reflect.DeepEqual(output, expected)
	if !pass {
		t.Errorf("fail: %#v != %#v", output, expected)
	}
}

func TestNormalizeMultiSlash(t *testing.T) {
	t.Parallel()
	input := "gemini://uscoffings.net/retro-computing/////////a///magazines"
	normalized, _ := NormalizeURL(input)
	output := normalized.String()
	expected := "gemini://uscoffings.net/retro-computing/a/magazines"
	pass := reflect.DeepEqual(output, expected)
	if !pass {
		t.Errorf("fail: %#v != %#v", output, expected)
	}
}

func TestNormalizeTrailingSlash(t *testing.T) {
	t.Parallel()
	input := "gemini://uscoffings.net/"
	normalized, _ := NormalizeURL(input)
	output := normalized.String()
	expected := "gemini://uscoffings.net/"
	pass := reflect.DeepEqual(output, expected)
	if !pass {
		t.Errorf("fail: %#v != %#v", output, expected)
	}
}

func TestNormalizeNoTrailingSlash(t *testing.T) {
	t.Parallel()
	input := "gemini://uscoffings.net"
	normalized, _ := NormalizeURL(input)
	output := normalized.String()
	expected := "gemini://uscoffings.net"
	pass := reflect.DeepEqual(output, expected)
	if !pass {
		t.Errorf("fail: %#v != %#v", output, expected)
	}
}

func TestNormalizeTrailingSlashPath(t *testing.T) {
	t.Parallel()
	input := "gemini://uscoffings.net/a/"
	normalized, _ := NormalizeURL(input)
	output := normalized.String()
	expected := "gemini://uscoffings.net/a/"
	pass := reflect.DeepEqual(output, expected)
	if !pass {
		t.Errorf("fail: %#v != %#v", output, expected)
	}
}

func TestNormalizeNoTrailingSlashPath(t *testing.T) {
	t.Parallel()
	input := "gemini://uscoffings.net/a"
	normalized, _ := NormalizeURL(input)
	output := normalized.String()
	expected := "gemini://uscoffings.net/a"
	pass := reflect.DeepEqual(output, expected)
	if !pass {
		t.Errorf("fail: %#v != %#v", output, expected)
	}
}

func TestNormalizeDot(t *testing.T) {
	t.Parallel()
	input := "gemini://uscoffings.net/retro-computing/./././////a///magazines"
	normalized, _ := NormalizeURL(input)
	output := normalized.String()
	expected := "gemini://uscoffings.net/retro-computing/a/magazines"
	pass := reflect.DeepEqual(output, expected)
	if !pass {
		t.Errorf("fail: %#v != %#v", output, expected)
	}
}

func TestNormalizePort(t *testing.T) {
	t.Parallel()
	input := "gemini://uscoffings.net:1965/a"
	normalized, _ := NormalizeURL(input)
	output := normalized.String()
	expected := "gemini://uscoffings.net/a"
	pass := reflect.DeepEqual(output, expected)
	if !pass {
		t.Errorf("fail: %#v != %#v", output, expected)
	}
}

func TestNormalizeURL(t *testing.T) {
	t.Parallel()
	input := "gemini://chat.gemini.lehmann.cx:11965/"
	normalized, _ := NormalizeURL(input)
	output := normalized.String()
	expected := "gemini://chat.gemini.lehmann.cx:11965/"
	pass := reflect.DeepEqual(output, expected)
	if !pass {
		t.Errorf("fail: %#v != %#v", output, expected)
	}

	input = "gemini://chat.gemini.lehmann.cx:11965/index?a=1&b=c"
	normalized, _ = NormalizeURL(input)
	output = normalized.String()
	expected = "gemini://chat.gemini.lehmann.cx:11965/index?a=1&b=c"
	pass = reflect.DeepEqual(output, expected)
	if !pass {
		t.Errorf("fail: %#v != %#v", output, expected)
	}

	input = "gemini://chat.gemini.lehmann.cx:11965/index#1"
	normalized, _ = NormalizeURL(input)
	output = normalized.String()
	expected = "gemini://chat.gemini.lehmann.cx:11965/index#1"
	pass = reflect.DeepEqual(output, expected)
	if !pass {
		t.Errorf("fail: %#v != %#v", output, expected)
	}

	input = "gemini://gemi.dev/cgi-bin/xkcd.cgi?1494"
	normalized, _ = NormalizeURL(input)
	output = normalized.String()
	expected = "gemini://gemi.dev/cgi-bin/xkcd.cgi?1494"
	pass = reflect.DeepEqual(output, expected)
	if !pass {
		t.Errorf("fail: %#v != %#v", output, expected)
	}
}

func TestNormalizePath(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    string // URL string to parse
		expected string // Expected normalized path
	}{
		// Basic cases
		{
			name:     "empty_path",
			input:    "http://example.com",
			expected: "/",
		},
		{
			name:     "root_path",
			input:    "http://example.com/",
			expected: "/",
		},
		{
			name:     "single_trailing_slash",
			input:    "http://example.com/test/",
			expected: "/test",
		},
		{
			name:     "no_trailing_slash",
			input:    "http://example.com/test",
			expected: "/test",
		},

		// Edge cases with slashes
		{
			name:     "multiple_trailing_slashes",
			input:    "http://example.com/test//",
			expected: "/test/",
		},
		{
			name:     "multiple_consecutive_slashes",
			input:    "http://example.com//test//",
			expected: "//test/",
		},
		{
			name:     "only_slashes",
			input:    "http://example.com////",
			expected: "///",
		},
		{
			name:     "single_slash",
			input:    "/",
			expected: "/",
		},

		// Encoded characters
		{
			name:     "encoded_spaces",
			input:    "http://example.com/foo%20bar/",
			expected: "/foo%20bar",
		},
		{
			name:     "encoded_special_chars",
			input:    "http://example.com/foo%2Fbar/",
			expected: "/foo%2Fbar",
		},

		// Query parameters and fragments
		{
			name:     "with_query_parameters",
			input:    "http://example.com/path?query=param",
			expected: "/path",
		},
		{
			name:     "with_fragment",
			input:    "http://example.com/path#fragment",
			expected: "/path",
		},
		{
			name:     "with_both_query_and_fragment",
			input:    "http://example.com/path?query=param#fragment",
			expected: "/path",
		},

		// Relative URLs
		{
			name:     "relative_path",
			input:    "/just/a/path/",
			expected: "/just/a/path",
		},

		// Unicode paths
		{
			name:     "unicode_characters",
			input:    "http://example.com/über/path/",
			expected: "/%C3%BCber/path",
		},
		{
			name:     "unicode_encoded",
			input:    "http://example.com/%C3%BCber/path/",
			expected: "/%C3%BCber/path",
		},

		// Weird but valid cases
		{
			name:     "dot_in_path",
			input:    "http://example.com/./path/",
			expected: "/./path",
		},
		{
			name:     "double_dot_in_path",
			input:    "http://example.com/../path/",
			expected: "/../path",
		},
		{
			name:     "mixed_case",
			input:    "http://example.com/PaTh/",
			expected: "/PaTh",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := url.Parse(tt.input)
			if err != nil {
				t.Fatalf("Failed to parse URL %q: %v", tt.input, err)
			}

			result := TrimTrailingPathSlash(u.EscapedPath())
			if result != tt.expected {
				t.Errorf("Input: %s\nExpected: %q\nGot: %q",
					u.Path, tt.expected, result)
			}
		})
	}
}
