package httpfunc

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

// Encoder encode and serialize the request body.
type Encoder interface {
	Encode(v any) error
}

var _ Encoder = (*defaultJSONEncoder)(nil)

// defaultJSONEncoder implements Encoder.
type defaultJSONEncoder struct {
	encoder *json.Encoder
}

func newDefaultJSONEncoder(w io.Writer) *defaultJSONEncoder {
	return &defaultJSONEncoder{encoder: json.NewEncoder(w)}
}

func (d *defaultJSONEncoder) Encode(v any) error {
	if err := d.encoder.Encode(v); err != nil {
		return fmt.Errorf("default json encoder Encode: %w", err)
	}

	return nil
}

// EncoderFunc type func which create encoder.
type EncoderFunc func(w io.Writer) Encoder

// ResponseBodyMarshalFunc marshals response func.
type ResponseBodyMarshalFunc func(w http.ResponseWriter, status int, response interface{})

func ResponseBodyMarshalJSON(logger *slog.Logger) ResponseBodyMarshalFunc {
	return ResponseBodyMarshalAny(
		func(w io.Writer) Encoder {
			return newDefaultJSONEncoder(w)
		}, logger, map[string][]string{"Content-Type": {"application/json"}},
	)
}

// ResponseBodyMarshalAny marshals response as JSON.
func ResponseBodyMarshalAny(
	encoderFn EncoderFunc,
	logger *slog.Logger,
	headers map[string][]string,
) ResponseBodyMarshalFunc {
	return func(w http.ResponseWriter, status int, response interface{}) {
		header := w.Header()
		for key, h := range headers {
			for _, v := range h {
				header.Add(key, v)
			}
		}

		w.WriteHeader(status)

		if response == nil {
			return
		}

		encoder := encoderFn(w)
		if err := encoder.Encode(response); err != nil {
			logger.Error("json.Marshal", slog.Any("err", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}
