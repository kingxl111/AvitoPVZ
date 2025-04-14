package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
)

// newMockResponseWriter returns a new mockResponseWriter.
func newMockResponseWriter() *mockResponseWriter {
	return &mockResponseWriter{
		headers: make(http.Header),
	}
}

// mockResponseWriter implements http.ResponseWriter and http.Flusher.
type mockResponseWriter struct {
	headers    http.Header
	body       bytes.Buffer
	statusCode int
	flushed    bool
}

// Header returns the mockResponseWriter's headers.
func (m *mockResponseWriter) Header() http.Header {
	return m.headers
}

// Write writes the mockResponseWriter's body.
func (m *mockResponseWriter) Write(b []byte) (int, error) {
	return m.body.Write(b)
}

// WriteHeader sets the mockResponseWriter's status code.
func (m *mockResponseWriter) WriteHeader(statusCode int) {
	m.statusCode = statusCode
}

// Flush sets the mockResponseWriter's flushed flag.
func (m *mockResponseWriter) Flush() {
	m.flushed = true
}

// TestWrapResponseWriter tests the WrapResponseWriter.
func TestWrapResponseWriter_Write(t *testing.T) {
	t.Parallel()

	type fields struct {
		data    []byte
		discard bool
	}

	testCases := []struct {
		name   string
		before func() (*WrapResponseWriter, fields, func())
	}{
		{
			name: "test_wrapper_write_rw",
			before: func() (*WrapResponseWriter, fields, func()) {
				recorder := httptest.NewRecorder()
				rw := NewWrapResponseWriter(recorder)
				data := []byte("hello world")
				return rw, fields{
						data:    data,
						discard: false,
					}, func() {
						if got := recorder.Body.Bytes(); !bytes.Equal(got, data) {
							t.Errorf("Bytes() = %v, expected %v", got, data)
						}
					}
			},
		},
		{
			name: "test_wrapper_with_empty_data",
			before: func() (*WrapResponseWriter, fields, func()) {
				recorder := httptest.NewRecorder()
				rw := NewWrapResponseWriter(recorder)
				data := []byte(nil)
				return rw, fields{
						data:    data,
						discard: false,
					}, func() {
						if got := recorder.Body.Bytes(); !bytes.Equal(got, data) {
							t.Errorf("Bytes() = %v, expected %v", got, data)
						}
					}
			},
		},
		{
			name: "test_wrapper_with_discard",
			before: func() (*WrapResponseWriter, fields, func()) {
				recorder := httptest.NewRecorder()
				rw := NewWrapResponseWriter(recorder)
				data := []byte(nil)
				return rw, fields{
						data:    data,
						discard: true,
					}, func() {
						if got := recorder.Body.Bytes(); got != nil {
							t.Errorf("Bytes() = %v, expected %v", got, nil)
						}
					}
			},
		},
	}

	for _, tc := range testCases {
		tc := tc // nolint
		t.Run(
			tc.name, func(t *testing.T) {
				t.Parallel()

				rw, fields, after := tc.before()
				if fields.discard {
					rw.Discard()
				}
				_, err := rw.Write(fields.data)
				if err != nil {
					t.Fatalf("Write() error = %v", err)
				}
				after()
			},
		)
	}
}

// TestWrapResponseWriter_Status tests the Status() method.
func TestWrapResponseWriter_Status(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		writeCode  int
		wantStatus int
	}{
		{
			name:       "test_explicit_status",
			writeCode:  http.StatusNotFound,
			wantStatus: http.StatusNotFound,
		},
		{
			name:       "test_implicit_status",
			wantStatus: 0,
		},
	}

	for _, tt := range tests {
		t.Run(
			tt.name, func(t *testing.T) {
				t.Parallel()

				rw := NewWrapResponseWriter(httptest.NewRecorder())

				if tt.writeCode != 0 {
					rw.WriteHeader(tt.writeCode)
				}

				if got := rw.Status(); got != tt.wantStatus {
					t.Errorf("Status() = %v, expected %v", got, tt.wantStatus)
				}
			},
		)
	}
}

func TestWrapResponseWriter_Flush(t *testing.T) {
	t.Parallel()

	mock := newMockResponseWriter()
	rw := NewWrapResponseWriter(mock)
	rw.Flush()

	if !mock.flushed {
		t.Error("Flush() did not call underlying Flush()")
	}

	if atomic.CompareAndSwapInt32(&rw.wroteHeader, 0, 1) {
		t.Error("Flush() did not set wroteHeader")
	}
}

func TestWrapResponseWriter_Size(t *testing.T) {
	t.Parallel()

	rw := NewWrapResponseWriter(httptest.NewRecorder())

	data := []byte("hello world")
	n, err := rw.Write(data)
	if err != nil {
		t.Fatalf("Write() error = %v", err)
	}

	if got := rw.Size(); got != n {
		t.Errorf("Size() = %v, expected %v", got, len(data))
	}

	if got := rw.Size(); got != len(data) {
		t.Errorf("Size() = %v, expected %v", got, len(data))
	}
}

func TestWrapResponseWriter_Bytes(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name   string
		before func() (*WrapResponseWriter, []byte, func(b []byte))
	}{
		{
			name: "test_wrapper_with_tee",
			before: func() (*WrapResponseWriter, []byte, func(b []byte)) {
				rw := NewWrapResponseWriter(httptest.NewRecorder())
				rw.Tee(new(bytes.Buffer))
				data := []byte("hello world")
				return rw, data, func(b []byte) {
					if got := rw.Bytes(); !bytes.Equal(got, data) {
						t.Errorf("Bytes() = %v, expected %v", got, data)
					}
				}
			},
		},
		{
			name: "test_wrapper_with_tee_with_empty_bytes",
			before: func() (*WrapResponseWriter, []byte, func(b []byte)) {
				rw := NewWrapResponseWriter(httptest.NewRecorder())
				rw.Tee(new(bytes.Buffer))
				data := []byte("")
				return rw, data, func(b []byte) {
					if got := rw.Bytes(); !bytes.Equal(got, data) {
						t.Errorf("Bytes() = %v, expected %v", got, data)
					}
				}
			},
		},
		{
			name: "test_wrapper_with_tee_with_nil_bytes",
			before: func() (*WrapResponseWriter, []byte, func(b []byte)) {
				rw := NewWrapResponseWriter(httptest.NewRecorder())
				rw.Tee(new(bytes.Buffer))
				data := []byte(nil)
				return rw, data, func(b []byte) {
					if got := rw.Bytes(); !bytes.Equal(got, data) {
						t.Errorf("Bytes() = %v, expected %v", got, data)
					}
				}
			},
		},
		{
			name: "test_wrapper_without_tee",
			before: func() (*WrapResponseWriter, []byte, func(b []byte)) {
				rw := NewWrapResponseWriter(httptest.NewRecorder())
				data := []byte("hello world")
				return rw, data, func(b []byte) {
					if got := rw.Bytes(); got != nil {
						t.Errorf("Bytes() = %v, expected %v", got, data)
					}
				}
			},
		},
	}

	for _, tc := range testCases {
		tc := tc // nolint
		t.Run(
			tc.name, func(t *testing.T) {
				t.Parallel()

				rw, data, after := tc.before()
				_, err := rw.Write(data)
				if err != nil {
					t.Fatalf("Write() error = %v", err)
				}
				after(data)
			},
		)
	}
}
