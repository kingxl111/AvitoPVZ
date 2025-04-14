package httpfunc

import (
	"bytes"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestUnmarshal(t *testing.T) {
	t.Parallel()

	testStruct := &struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}{}

	testCases := []struct {
		name          string
		contentType   string
		body          string
		maxBodyBytes  int64
		expectedErr   error
		expectedData  interface{}
		validateFuncs []RequestValidateFunc
	}{
		{
			name:         "test_valid_JSON",
			contentType:  "application/json",
			body:         `{"name":"John","age":30}`,
			maxBodyBytes: 1024,
			expectedErr:  nil,
			expectedData: &struct {
				Name string `json:"name"`
				Age  int    `json:"age"`
			}{Name: "John", Age: 30},
			validateFuncs: []RequestValidateFunc{
				ValidateJSONContentType,
			},
		},
		{
			name:         "test_invalid_content_type",
			contentType:  "text/plain",
			body:         `{"name":"John","age":30}`,
			maxBodyBytes: 1024,
			expectedErr:  ErrUnsupportedMediaType,
			validateFuncs: []RequestValidateFunc{
				ValidateJSONContentType,
			},
		},
		{
			name:        "test_validate_custom_content_type",
			contentType: "text/plain",
			body:        `{"name":"John","age":30}`,
			validateFuncs: []RequestValidateFunc{
				func(r *http.Request) error {
					if t := r.Header.Get("content-type"); len(t) < 10 || t[:10] != "text/plain" {
						return ErrUnsupportedMediaType
					}
					return nil
				},
			},
			maxBodyBytes: 1024,
			expectedErr:  ErrUnsupportedMediaType,
		},
		{
			name:         "test_malformed_JSON",
			contentType:  "application/json",
			body:         `{"name":"John","age":30`,
			maxBodyBytes: 1024,
			expectedErr:  ErrBadRequest,
			validateFuncs: []RequestValidateFunc{
				ValidateJSONContentType,
			},
		},
		{
			name:         "test_invalid_JSON_value",
			contentType:  "application/json",
			body:         `{"name":"John","age":"thirty"}`,
			maxBodyBytes: 1024,
			expectedErr:  ErrBadRequest,
			validateFuncs: []RequestValidateFunc{
				ValidateJSONContentType,
			},
		},
		{
			name:         "test_unknown_JSON_field",
			contentType:  "application/json",
			body:         `{"name":"John", "age":30, "city":"New York"}`,
			maxBodyBytes: 1024,
			expectedErr:  ErrBadRequest,
			validateFuncs: []RequestValidateFunc{
				ValidateJSONContentType,
			},
		},
		{
			name:         "test_empty_body",
			contentType:  "application/json",
			body:         ``,
			maxBodyBytes: 1024,
			expectedErr:  ErrBadRequest,
			validateFuncs: []RequestValidateFunc{
				ValidateJSONContentType,
			},
		},
		{
			name:         "test_body_too_large",
			contentType:  "application/json",
			body:         `{"name":"John","age":30}`,
			maxBodyBytes: 10,
			expectedErr:  ErrRequestEntityTooLarge,
			validateFuncs: []RequestValidateFunc{
				ValidateJSONContentType,
			},
		},
		{
			name:         "test_multiple_JSON_objects",
			contentType:  "application/json",
			body:         `{"name":"John","age":30}{"name":"Jane","age":25}`,
			maxBodyBytes: 1024,
			expectedErr:  ErrBadRequest,
			validateFuncs: []RequestValidateFunc{
				ValidateJSONContentType,
			},
		},
	}

	for _, tc := range testCases {
		t.Run(
			tc.name, func(t *testing.T) {
				t.Parallel()

				req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tc.body))
				req.Header.Set("Content-Type", tc.contentType)

				unmarshalFn := RequestBodyUnmarshalAny(
					func(r io.Reader) Decoder {
						return newDefaultJSONDecoder(r, false)
					}, WithMaxBodyBytes(tc.maxBodyBytes), WithValidateFuncs(tc.validateFuncs...),
				)

				if err := unmarshalFn(req, &testStruct); err != nil && !errors.Is(err, tc.expectedErr) {
					t.Errorf("expected error %q but got %q", tc.expectedErr, err.Error())
					return
				}

				if tc.expectedData != nil && !reflect.DeepEqual(testStruct, tc.expectedData) {
					t.Errorf("expected testStruct %v but got %v", tc.expectedData, testStruct)
				}
			},
		)
	}
}

func TestMakeMaxBytesReader(t *testing.T) {
	t.Parallel()

	expectedLimit := int64(1)
	br, _ := makeMaxBytesReader(bytes.NewReader([]byte("test")), expectedLimit)
	if br.N != expectedLimit {
		t.Errorf("expected limit %decoder but got %decoder", expectedLimit, br.N)
		return
	}

	brp.Put(br)

	br1, _ := makeMaxBytesReader(bytes.NewReader([]byte("test")), expectedLimit)
	if br1.N != expectedLimit {
		t.Errorf("expected limit %decoder but got %decoder", expectedLimit, br1.N)
		return
	}
}

func TestDropMaxBytesReader(t *testing.T) {
	t.Parallel()

	expectedCase1Limit := int64(1)
	br, drop := makeMaxBytesReader(bytes.NewReader([]byte("test")), expectedCase1Limit)
	if br.N != expectedCase1Limit {
		t.Errorf("expected limit %decoder but got %decoder", expectedCase1Limit, br.N)
		return
	}

	drop()

	expectedCase2Limit := int64(2)
	br1, drop1 := makeMaxBytesReader(bytes.NewReader([]byte("test")), expectedCase2Limit)
	if br1.N != expectedCase2Limit {
		t.Errorf("expected limit %decoder but got %decoder", expectedCase2Limit, br.N)
		return
	}

	br1 = nil
	_ = br1
	drop1()

	expectedCase3Limit := int64(3)
	br2, _ := makeMaxBytesReader(bytes.NewReader([]byte("test")), expectedCase3Limit)
	if br2 == nil {
		t.Errorf("expected limit %v but got %decoder", nil, br2)
		return
	}

	if br2.N != expectedCase3Limit {
		t.Errorf("expected limit %decoder but got %decoder", expectedCase3Limit, br2.N)
	}
}
