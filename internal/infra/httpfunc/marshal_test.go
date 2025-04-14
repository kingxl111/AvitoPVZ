package httpfunc

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"AvitoPVZ/pkg/logging"
)

func TestResponseBodyMarshalJSON(t *testing.T) {
	t.Parallel()

	logger := logging.Development()

	marshalResponse := ResponseBodyMarshalJSON(logger)

	type mockStruct struct {
		Message string `json:"message"`
		Embed   struct {
			Name string `json:"name"`
			Age  int    `json:"age"`
		} `json:"embed"`
	}
	type mockSlice []string
	type mockMaps map[string]string

	tests := []struct {
		name           string
		status         int
		response       interface{}
		expectedStatus int
		expectedBody   string
	}{
		{
			name:           "test_status_ok_slice",
			status:         http.StatusOK,
			response:       mockSlice{"hello", "world"},
			expectedStatus: http.StatusOK,
			expectedBody:   `["hello","world"]` + "\n",
		},
		{
			name:           "test_status_ok_map",
			status:         http.StatusOK,
			response:       mockMaps{"hello": "world"},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"hello":"world"}` + "\n",
		},
		{
			name:           "test_status_ok_nil",
			status:         http.StatusOK,
			response:       nil,
			expectedStatus: http.StatusOK,
			expectedBody:   "",
		},
		{
			name:           "test_status_ok_empty_struct",
			status:         http.StatusOK,
			response:       struct{}{},
			expectedStatus: http.StatusOK,
			expectedBody:   `{}` + "\n",
		},
		{
			name:   "test_status_ok_struct",
			status: http.StatusOK,
			response: mockStruct{
				Message: "hello",
				Embed: struct {
					Name string `json:"name"`
					Age  int    `json:"age"`
				}{Name: "John", Age: 30},
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"message":"hello","embed":{"name":"John","age":30}}` + "\n",
		},
		{
			name:           "test_status_internal_func",
			status:         http.StatusInternalServerError,
			response:       func() {},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   "",
		},
		{
			name:   "test_status_ok_nested_slice",
			status: http.StatusOK,
			response: []mockStruct{
				{
					Message: "hello",
					Embed: struct {
						Name string `json:"name"`
						Age  int    `json:"age"`
					}{Name: "Alice", Age: 25},
				},
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `[{"message":"hello","embed":{"name":"Alice","age":25}}]` + "\n",
		},
		{
			name:   "test_status_ok_nested_map",
			status: http.StatusOK,
			response: map[string]mockStruct{
				"user": {
					Message: "hello", Embed: struct {
						Name string `json:"name"`
						Age  int    `json:"age"`
					}{Name: "Bob", Age: 35},
				},
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"user":{"message":"hello","embed":{"name":"Bob","age":35}}}` + "\n",
		},
		{
			name:           "test_status_ok_large_number",
			status:         http.StatusOK,
			response:       map[string]interface{}{"largeNumber": 9223372036854775807},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"largeNumber":9223372036854775807}` + "\n",
		},
		{
			name:           "test_status_ok_float",
			status:         http.StatusOK,
			response:       map[string]interface{}{"pi": 3.14159},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"pi":3.14159}` + "\n",
		},
		{
			name:           "test_status_ok_bool",
			status:         http.StatusOK,
			response:       map[string]interface{}{"isActive": true},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"isActive":true}` + "\n",
		},
		{
			name:           "test_status_ok_empty_slice",
			status:         http.StatusOK,
			response:       []int{},
			expectedStatus: http.StatusOK,
			expectedBody:   `[]` + "\n",
		},
		{
			name:           "test_status_ok_empty_map",
			status:         http.StatusOK,
			response:       map[string]int{},
			expectedStatus: http.StatusOK,
			expectedBody:   `{}` + "\n",
		},
		{
			name:   "test_status_ok_mixed_types",
			status: http.StatusOK,
			response: map[string]interface{}{
				"string": "hello", "number": 42, "bool": true, "slice": []int{1, 2, 3},
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"bool":true,"number":42,"slice":[1,2,3],"string":"hello"}` + "\n",
		},
		{
			name:           "test_status_ok_time",
			status:         http.StatusOK,
			response:       map[string]interface{}{"timestamp": "2023-07-25T12:00:00Z"},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"timestamp":"2023-07-25T12:00:00Z"}` + "\n",
		},
		{
			name:           "test_status_ok_null",
			status:         http.StatusOK,
			response:       map[string]interface{}{"nullValue": nil},
			expectedStatus: http.StatusOK,
			expectedBody:   `{"nullValue":null}` + "\n",
		},
	}

	for _, tt := range tests {
		tt := tt // nolint
		t.Run(
			tt.name, func(t *testing.T) {
				t.Parallel()

				rr := httptest.NewRecorder()

				marshalResponse(rr, tt.status, tt.response)

				if status := rr.Code; status != tt.expectedStatus {
					t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
				}

				if tt.expectedBody != "" && !bytes.Equal(rr.Body.Bytes(), []byte(tt.expectedBody)) {
					t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), tt.expectedBody)
				}
			},
		)
	}
}
