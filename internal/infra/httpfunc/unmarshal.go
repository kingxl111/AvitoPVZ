package httpfunc

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
)

const defaultMaxBodyBytes = -1 // negative means unlimited

var (
	ErrBadRequest            = errors.New("bad request")
	ErrRequestEntityTooLarge = errors.New("request entity too large")
	ErrUnsupportedMediaType  = errors.New("unsupported media type")
	ErrUnknown               = errors.New("unknown")
)

// use cache sync.Pool for more control about MaxBytesReader allocations.
var brp = &sync.Pool{
	New: func() any {
		return &MaxBytesReader{}
	},
}

// makeMaxBytesReader returns a new MaxBytesReader with the given reader and limit.
func makeMaxBytesReader(r io.Reader, n int64) (br *MaxBytesReader, drop func()) {
	// destructor for MaxBytesReader with reinitialize the reader and limit
	drop = func() {
		// drop is a no-op if br is nil
		defer brp.Put(br)

		if br == nil {
			return
		}

		br.R = nil
		br.N = 0
	}

	br, _ = brp.Get().(*MaxBytesReader)
	if br == nil {
		return NewMaxBytesReader(r, n), drop
	}

	// reinitialize the reader and limit
	br.R = r
	br.N = n

	return br, drop
}

type RequestValidateFunc func(r *http.Request) error

// ValidateJSONContentType checks if the content type is application/json.
func ValidateJSONContentType(r *http.Request) error {
	if t := r.Header.Get("content-type"); len(t) < 16 || t[:16] != "application/json" {
		return ErrUnsupportedMediaType
	}

	return nil
}

type UnmarshalOptions struct {
	validateFuncs       []RequestValidateFunc
	maxBodyBytes        int64
	permitUnknownFields bool
	decoder             Decoder
}

type UnmarshalOption func(options *UnmarshalOptions)

// WithValidateFuncs append validate funcs.
func WithValidateFuncs(validateFuncs ...RequestValidateFunc) UnmarshalOption {
	return func(options *UnmarshalOptions) {
		options.validateFuncs = append(options.validateFuncs, validateFuncs...)
	}
}

// WithMaxBodyBytes set max number of bytes when reading request body.
func WithMaxBodyBytes(n int64) UnmarshalOption {
	return func(options *UnmarshalOptions) {
		options.maxBodyBytes = n
	}
}

// WithPermitUnknownFields set permit unknown fields for decoded json.
func WithPermitUnknownFields() UnmarshalOption {
	return func(options *UnmarshalOptions) {
		options.permitUnknownFields = true
	}
}

func WithCustomDecoder(d Decoder) UnmarshalOption {
	return func(options *UnmarshalOptions) {
		options.decoder = d
	}
}

// RequestBodyUnmarshalFunc unmarshal the request body.
type RequestBodyUnmarshalFunc func(r *http.Request, data interface{}) error

// Decoder decode and deserialize the request body.
type Decoder interface {
	Decode(data interface{}) error
}

// DecoderFunc type func which create decoder.
type DecoderFunc func(r io.Reader) Decoder

var _ Decoder = (*defaultJSONDecoder)(nil)

// DecoderFunc make decoder with io.Reader and permitUnknownFields.
func newDefaultJSONDecoder(r io.Reader, permitUnknownFields bool) *defaultJSONDecoder {
	return &defaultJSONDecoder{decoder: json.NewDecoder(r), permitUnknownFields: permitUnknownFields}
}

type defaultJSONDecoder struct {
	permitUnknownFields bool
	decoder             *json.Decoder
}

func (dec defaultJSONDecoder) Decode(v any) error {
	d := dec.decoder

	if dec.permitUnknownFields {
		d.DisallowUnknownFields()
	}

	if err := d.Decode(&v); err != nil {
		var syntaxErr *json.SyntaxError
		var unmarshalError *json.UnmarshalTypeError
		switch {
		case errors.As(err, &syntaxErr):
			return fmt.Errorf(
				"malformed json at position %decoder %v: %w", syntaxErr.Offset, err.Error(), ErrBadRequest,
			)
		case errors.Is(err, io.ErrUnexpectedEOF):
			return fmt.Errorf("malformed json: %w", ErrBadRequest)
		case errors.As(err, &unmarshalError):
			return fmt.Errorf(
				"invalid value %q at position %decoder: %w", unmarshalError.Field, unmarshalError.Offset, ErrBadRequest,
			)
		case strings.HasPrefix(err.Error(), "json: unknown field"):
			fieldName := strings.TrimPrefix(err.Error(), "json: unknown field ")
			return fmt.Errorf("unknown field %s: %w", fieldName, ErrBadRequest)
		case errors.Is(err, io.EOF):
			return fmt.Errorf("body must not be empty: %w", ErrBadRequest)
		case errors.Is(err, ErrMaxLimit) || err.Error() == "http: request body too large":
			return fmt.Errorf("%v: %w", err.Error(), ErrRequestEntityTooLarge)
		default:
			return fmt.Errorf("%v failed to decode json: %w", err.Error(), ErrUnknown)
		}
	}

	if d.More() {
		return fmt.Errorf("body must contain only one JSON object: %w", ErrBadRequest)
	}

	return nil
}

// RequestBodyUnmarshalJSON decode and deserialize the request body.
func RequestBodyUnmarshalJSON(opts ...UnmarshalOption) RequestBodyUnmarshalFunc {
	return RequestBodyUnmarshalAny(
		func(r io.Reader) Decoder {
			return newDefaultJSONDecoder(r, false)
		}, opts...,
	)
}

// RequestBodyUnmarshalAny unmarshal the request body.
func RequestBodyUnmarshalAny(decodeFn DecoderFunc, opts ...UnmarshalOption) RequestBodyUnmarshalFunc {
	opt := UnmarshalOptions{
		maxBodyBytes: defaultMaxBodyBytes,
	}

	for _, o := range opts {
		o(&opt)
	}

	return func(req *http.Request, data interface{}) error {
		for _, f := range opt.validateFuncs {
			if err := f(req); err != nil {
				return fmt.Errorf("request validate func: %w", err)
			}
		}

		var r io.Reader
		if opt.maxBodyBytes > defaultMaxBodyBytes {
			br, drop := makeMaxBytesReader(req.Body, opt.maxBodyBytes)
			defer drop()
			r = br
		}

		decoder := decodeFn(r)
		if err := decoder.Decode(data); err != nil {
			return fmt.Errorf("decoder.Decode: %w", err)
		}

		return nil
	}
}

// ErrToHTTPCode convert error to http code.
func ErrToHTTPCode(err error) int {
	switch {
	case errors.Is(err, ErrBadRequest):
		return http.StatusBadRequest
	case errors.Is(err, ErrRequestEntityTooLarge):
		return http.StatusRequestEntityTooLarge
	case errors.Is(err, ErrUnsupportedMediaType):
		return http.StatusUnsupportedMediaType
	case errors.Is(err, ErrUnknown):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
