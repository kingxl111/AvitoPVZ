package middleware

import (
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func BenchmarkMaskFunc(b *testing.B) {
	headers := http.Header{
		"Authorization": []string{"Bearer token"},
		"Content-Type":  []string{"application/json"},
		"User-Agent":    []string{"Go-http-client/1.1"},
	}

	replacers := []ReplaceFunc{
		func(key, value string) string {
			if key == "Authorization" {
				return "*****"
			}
			return value
		},
		func(key, value string) string {
			if key == "User-Agent" {
				return strings.ToUpper(value)
			}
			return value
		},
	}

	mask := MaskHeadersFunc(replacers)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_ = mask(headers)
	}
}

func TestPathSkipper(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		paths    []string
		testPath string
		expected bool
	}{
		{
			name:     "test_with_no_skip",
			paths:    []string{},
			testPath: "/test",
			expected: false,
		},
		{
			name:     "test_with_one_skipped",
			paths:    []string{"/skip"},
			testPath: "/skip",
			expected: true,
		},
		{
			name:     "test_with_not_match_pth",
			paths:    []string{"/skip"},
			testPath: "/test",
			expected: false,
		},
		{
			name:     "test_with_one_of_skipped",
			paths:    []string{"/skip1", "/skip2", "/skip3"},
			testPath: "/skip2",
			expected: true,
		},
		{
			name:     "test_with_not_match_pth",
			paths:    []string{"/skip1", "/skip2", "/skip3"},
			testPath: "/test",
			expected: false,
		},
		{
			name:     "test_with_root_pth_skipped",
			paths:    []string{"/"},
			testPath: "/",
			expected: true,
		},
	}

	for _, tc := range tests {
		tc := tc // nolint
		t.Run(
			tc.name, func(t *testing.T) {
				t.Parallel()

				ctx := context.Background()
				req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://example.com"+tc.testPath, nil)
				require.NoError(t, err)

				skipper := PathSkipper(tc.paths...)

				got := skipper(ctx, req)
				assert.Equal(t, tc.expected, got, "PathSkipper(%v) with path %s", tc.paths, tc.testPath)
			},
		)
	}
}
