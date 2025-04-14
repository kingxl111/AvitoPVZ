package apihandler

import (
	httpfunc2 "AvitoPVZ/internal/infra/httpfunc"
	"AvitoPVZ/internal/infra/httpfunc/middleware/requestidmw"
	"AvitoPVZ/pkg/api/oapigen/pvzops"
	"context"
	"log/slog"
	"net/http"
)

const internalErrorMessage = "internal server error"

var _ pvzops.ServerInterface = (*Handler)(nil)

type Config struct {
	BodyLimitBytes int64
	FilterMaxLimit int64
}

var (
	defaultBodyLimitBytes int64 = 1 << 20 // 1 MB
	defaultFilterMaxLimit int64 = 30
)

func New(
	logger *slog.Logger,
	userStory userStory,
	pvzStory pvzStory,
	productStory productStory,
	receiptStory receiptStory,
) *Handler {
	return &Handler{
		logger: logger,
		vmxClientDrainRequestID: func(ctx context.Context) func(ctx context.Context, req *http.Request) error {
			return func(ctx context.Context, req *http.Request) error {
				req.Header.Set("Content-Type", "application/json")
				req.Header.Set("X-Request-ID", requestidmw.RequestIDFromContext(ctx))
				return nil
			}
		},
		unmarshalBody:   httpfunc2.RequestBodyUnmarshalJSON(httpfunc2.WithMaxBodyBytes(defaultBodyLimitBytes)),
		marshalResponse: httpfunc2.ResponseBodyMarshalJSON(logger),
		userStory:       userStory,
		pvzStory:        pvzStory,
		productStory:    productStory,
		receiptStory:    receiptStory,
	}
}

type Handler struct {
	logger                  *slog.Logger
	vmxClientDrainRequestID func(ctx context.Context) func(ctx context.Context, req *http.Request) error
	unmarshalBody           func(r *http.Request, v any) error
	marshalResponse         func(w http.ResponseWriter, status int, response interface{})

	userStory    userStory
	pvzStory     pvzStory
	receiptStory receiptStory
	productStory productStory
}
