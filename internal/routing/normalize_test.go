package routing

import "testing"

func TestNormalizeRoutePath(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		arg      string
		expected string
		wantErr  bool
	}{
		{
			name:     "valid to valid, with trailing",
			arg:      "/some/",
			expected: "/some/",
		},
		{
			name:     "valid to valid, with trailing, multiple slashes and spaces",
			arg:      "      /some///  ",
			expected: "/some/",
		},
		{
			name:     "valid to valid, without trailing",
			arg:      "/some",
			expected: "/some",
		},
		{
			name:     "without prefix slash, without trailing",
			arg:      "some",
			expected: "/some",
		},
		{
			name:     "without prefix slash, with trailing",
			arg:      "some/",
			expected: "/some/",
		},
		{
			name:     "root to root",
			arg:      "/",
			expected: "/",
		},
		{
			name:     "root to root, with spaces",
			arg:      "       /  ",
			expected: "/",
		},
		{
			name:     "root to root, multiple slashes",
			arg:      "///",
			expected: "/",
		},
		{
			name:    "empty path",
			arg:     "",
			wantErr: true,
		},
		{
			name:    "spaces only",
			arg:     "    ",
			wantErr: true,
		},
		{
			name:    "space inside path",
			arg:     "/some /other",
			wantErr: true,
		},
		{
			name:    "multiple slashes inside path",
			arg:     "/some///other",
			wantErr: true,
		},
		{
			name:    "fragment rejected",
			arg:     "/some/other#skib",
			wantErr: true,
		},
		{
			name:    "query rejected",
			arg:     "/some/other?skib=67",
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			res, err := NormalizeRoutePath(test.arg)
			if err != nil && !test.wantErr {
				t.Errorf("Expected %s, but error occured: %s",
					test.expected,
					err,
				)
			}

			if err != nil && test.wantErr {
				return
			}

			if test.wantErr {
				t.Errorf("Expected error, but got: %s",
					res,
				)
			}

			if res != test.expected {
				t.Errorf("Expected %s, but got: %s",
					test.expected,
					res,
				)
			}
		})
	}
}

func TestNormalizeUrl(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		arg      string
		expected string
		wantErr  bool
	}{
		{
			name:     "valid http url stays same",
			arg:      "http://127.0.0.1:8080",
			expected: "http://127.0.0.1:8080",
		},
		{
			name:     "valid https url stays same",
			arg:      "https://api.local",
			expected: "https://api.local",
		},
		{
			name:     "spaces trimmed",
			arg:      "   http://127.0.0.1:8080   ",
			expected: "http://127.0.0.1:8080",
		},
		{
			name:     "trailing slash removed",
			arg:      "http://127.0.0.1:8080/",
			expected: "http://127.0.0.1:8080",
		},
		{
			name:     "multiple trailing slashes removed",
			arg:      "http://127.0.0.1:8080///",
			expected: "http://127.0.0.1:8080",
		},
		{
			name:     "scheme and host lowercased",
			arg:      "HTTP://API.LOCAL:8080",
			expected: "http://api.local:8080",
		},
		{
			name:     "query removed",
			arg:      "http://127.0.0.1:8080?foo=bar",
			expected: "http://127.0.0.1:8080",
		},
		{
			name:     "fragment removed",
			arg:      "http://127.0.0.1:8080#section",
			expected: "http://127.0.0.1:8080",
		},
		{
			name:     "path removed",
			arg:      "http://127.0.0.1:8080/api/users",
			expected: "http://127.0.0.1:8080",
		},
		{
			name:     "path query and fragment removed",
			arg:      "http://127.0.0.1:8080/api/users?foo=bar#section",
			expected: "http://127.0.0.1:8080",
		},
		{
			name:    "empty url",
			arg:     "",
			wantErr: true,
		},
		{
			name:    "spaces only",
			arg:     "     ",
			wantErr: true,
		},
		{
			name:    "missing scheme",
			arg:     "127.0.0.1:8080",
			wantErr: true,
		},
		{
			name:    "unsupported scheme",
			arg:     "ftp://127.0.0.1:8080",
			wantErr: true,
		},
		{
			name:    "empty host",
			arg:     "http://",
			wantErr: true,
		},
		{
			name:    "path without host",
			arg:     "http:///api",
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			res, err := NormalizeUrl(test.arg)
			if err != nil && !test.wantErr {
				t.Errorf("Expected %s, but error occurred: %s",
					test.expected,
					err,
				)
				return
			}

			if err != nil && test.wantErr {
				return
			}

			if test.wantErr {
				t.Errorf("Expected error, but got: %s",
					res.String(),
				)
				return
			}

			if res.String() != test.expected {
				t.Errorf("Expected %s, but got: %s",
					test.expected,
					res.String(),
				)
			}
		})
	}
}

// StripPrefix requires normalized path and prefix,
// therefore this test doesnt check it.
func TestStripPrefix(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		path     string
		prefix   string
		expected string
	}{
		{
			name:     "root prefix keeps path",
			path:     "/some",
			prefix:   "/",
			expected: "/some",
		},
		{
			name:     "same path becomes root",
			path:     "/some",
			prefix:   "/some",
			expected: "/",
		},
		{
			name:     "nested path with trailing slash",
			path:     "/some/other/",
			prefix:   "/some",
			expected: "/other/",
		},
		{
			name:     "nested path without trailing slash",
			path:     "/some/other",
			prefix:   "/some",
			expected: "/other",
		},
		{
			name:     "deep nested path",
			path:     "/some/other/deep",
			prefix:   "/some",
			expected: "/other/deep",
		},
		{
			name:     "prefix with trailing slash",
			path:     "/some/other",
			prefix:   "/some/",
			expected: "/other",
		},
		{
			name:     "root path with root prefix",
			path:     "/",
			prefix:   "/",
			expected: "/",
		},
		{
			name:     "non matching prefix keeps path",
			path:     "/something",
			prefix:   "/some",
			expected: "/something",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			res := StripPrefix(test.path, test.prefix)

			if res != test.expected {
				t.Errorf("Expected %s, but got: %s",
					test.expected,
					res,
				)
			}
		})
	}
}
