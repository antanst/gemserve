package server

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestCalculateLocalPath(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		basePath    string
		want        string
		expectError bool
	}{
		// Basic path handling
		{
			name:        "Simple valid path",
			input:       "folder/file.txt",
			basePath:    "/base",
			want:        "/base/folder/file.txt",
			expectError: false,
		},
		{
			name:        "Empty path",
			input:       "",
			basePath:    "/base",
			want:        "/base",
			expectError: false,
		},
		{
			name:        "Current directory",
			input:       ".",
			basePath:    "/base",
			want:        "/base",
			expectError: false,
		},

		// Leading/trailing slash handling
		{
			name:        "Path with leading slash",
			input:       "/folder/file.txt",
			basePath:    "/base",
			want:        "/base/folder/file.txt",
			expectError: false,
		},
		{
			name:        "Path with trailing slash",
			input:       "folder/",
			basePath:    "/base",
			want:        "/base/folder",
			expectError: false,
		},
		{
			name:        "Path with both leading and trailing slashes",
			input:       "/folder/",
			basePath:    "/base",
			want:        "/base/folder",
			expectError: false,
		},

		// Path traversal attempts
		{
			name:        "Simple path traversal attempt",
			input:       "../file.txt",
			basePath:    "/base",
			want:        "",
			expectError: true,
		},
		{
			name:        "Complex path traversal attempt",
			input:       "folder/../../../etc/passwd",
			basePath:    "/base",
			want:        "",
			expectError: true,
		},
		{
			name:        "Encoded path traversal attempt",
			input:       "folder/..%2F..%2F..%2Fetc%2Fpasswd",
			basePath:    "/base",
			want:        "/base/folder/..%2F..%2F..%2Fetc%2Fpasswd",
			expectError: false,
		},
		{
			name:        "Double dot hidden in path",
			input:       "folder/.../.../etc/passwd",
			basePath:    "/base",
			want:        "/base/folder/.../.../etc/passwd",
			expectError: false,
		},

		// Edge cases
		{
			name:        "Multiple sequential slashes",
			input:       "folder///subfolder////file.txt",
			basePath:    "/base",
			want:        "",
			expectError: true,
		},
		{
			name:        "Unicode characters in path",
			input:       "фольдер/файл.txt",
			basePath:    "/base",
			want:        "/base/фольдер/файл.txt",
			expectError: false,
		},
		{
			name:        "Path with spaces and special characters",
			input:       "my folder/my file!@#$%.txt",
			basePath:    "/base",
			want:        "/base/my folder/my file!@#$%.txt",
			expectError: false,
		},
		{
			name:        "Very long path",
			input:       "a/b/c/d/e/f/g/h/i/j/k/l/m/n/o/p/q/r/s/t/u/v/w/x/y/z/file.txt",
			basePath:    "/base",
			want:        "/base/a/b/c/d/e/f/g/h/i/j/k/l/m/n/o/p/q/r/s/t/u/v/w/x/y/z/file.txt",
			expectError: false,
		},

		// Base path variations
		{
			name:        "Empty base path",
			input:       "file.txt",
			basePath:    "",
			want:        "file.txt",
			expectError: false,
		},
		{
			name:        "Relative base path",
			input:       "file.txt",
			basePath:    "base/folder",
			want:        "base/folder/file.txt",
			expectError: false,
		},
		{
			name:        "Base path with trailing slash",
			input:       "file.txt",
			basePath:    "/base/",
			want:        "/base/file.txt",
			expectError: false,
		},

		// Symbolic link-like paths (if supported)
		{
			name:        "Path with symbolic link-like components",
			input:       "folder/symlink/../file.txt",
			basePath:    "/base",
			want:        "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calculateLocalPath(tt.input, tt.basePath)

			// Check error expectation
			if (err != nil) != tt.expectError {
				t.Errorf("calculateLocalPath() error = %v, expectError = %v", err, tt.expectError)
				return
			}

			// If we expect an error, don't check the returned path
			if tt.expectError {
				return
			}

			// Check if the returned path matches expected
			if got != tt.want {
				t.Errorf("calculateLocalPath() = %v, want %v", got, tt.want)
			}

			// Additional security checks for non-error cases
			if !tt.expectError {
				// Verify the returned path is within base path
				if !isWithinBasePath(got, tt.basePath) {
					t.Errorf("Result path %v escapes base path %v", got, tt.basePath)
				}

				// Verify no '..' components in final path
				if containsParentRef(got) {
					t.Errorf("Result path %v contains parent references", got)
				}
			}
		})
	}
}

// Helper function to check if a path is contained within the base path
func isWithinBasePath(path, basePath string) bool {
	if basePath == "" {
		return true
	}

	absBase, err := filepath.Abs(basePath)
	if err != nil {
		return false
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	rel, err := filepath.Rel(absBase, absPath)
	if err != nil {
		return false
	}

	return !strings.HasPrefix(rel, "..")
}

// Helper function to check if a path contains parent directory references
func containsParentRef(path string) bool {
	parts := strings.Split(filepath.Clean(path), string(filepath.Separator))
	for _, part := range parts {
		if part == ".." {
			return true
		}
	}
	return false
}
