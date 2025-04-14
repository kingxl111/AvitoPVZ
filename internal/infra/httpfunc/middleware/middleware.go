package middleware

import (
	"context"
	"net/http"
)

type (
	SkipMiddlewareFunc func(ctx context.Context, r *http.Request) bool
	ReplaceFunc        func(key, oldValue string) (modified string)
)

var _ ReplaceFunc = BasicAuthReplacer

// BasicAuthReplacer - replace basic auth value.
func BasicAuthReplacer(key string, oldValue string) string {
	if key == "Authorization" {
		return "*****"
	}

	return oldValue
}

// PathSkipper - skip middleware by path.
func PathSkipper(pths ...string) SkipMiddlewareFunc {
	return func(ctx context.Context, r *http.Request) bool {
		for _, pth := range pths {
			if r.URL.Path == pth {
				return true
			}
		}
		return false
	}
}

// SkipperFunc - configure skippers for middleware.
func SkipperFunc(skippers []SkipMiddlewareFunc) func(childCtx context.Context, r *http.Request) bool {
	return func(childCtx context.Context, r *http.Request) bool {
		for _, f := range skippers {
			if f(childCtx, r) {
				return true
			}
		}
		return false
	}
}

// MaskHeadersFunc - configure maskers for middleware.
func MaskHeadersFunc(replacers []ReplaceFunc) func(header http.Header) http.Header {
	return func(header http.Header) http.Header {
		cloned := header.Clone()

		for key, values := range header {
			for idx := range values {
				val := values[idx]
				for idx1 := range replacers {
					val = replacers[idx1](key, values[idx])
				}
				cloned[key][idx] = val
			}
		}

		return cloned
	}
}
