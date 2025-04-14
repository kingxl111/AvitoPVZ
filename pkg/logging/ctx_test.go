package logging

import (
	"context"
	"log/slog"
	"sync/atomic"
	"testing"
	"time"
)

type mockCtxHandler struct {
	handleCalled    bool
	withAttrsCalled bool
	withGroupCalled bool
	attrs           []slog.Attr
	groupName       string
}

func (th *mockCtxHandler) Enabled(ctx context.Context, rec slog.Level) bool {
	return true
}

func (th *mockCtxHandler) Handle(ctx context.Context, rec slog.Record) error {
	th.handleCalled = true
	return nil
}

func (th *mockCtxHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	th.withAttrsCalled = true
	th.attrs = attrs
	return th
}

func (th *mockCtxHandler) WithGroup(name string) slog.Handler {
	th.withGroupCalled = true
	th.groupName = name
	return th
}

func extractorFunc1(ctx context.Context) []slog.Attr {
	val, ok := ctx.Value("key1").(string)
	if !ok {
		return nil
	}
	return []slog.Attr{slog.String("key1", val)}
}

func extractorFunc2(ctx context.Context) []slog.Attr {
	val, ok := ctx.Value("key2").(string)
	if !ok {
		return nil
	}
	return []slog.Attr{slog.String("key2", val)}
}

func testComputeCalls(v *int64, v1 ExtractAttrsFunc) func(ctx context.Context) []slog.Attr {
	return func(ctx context.Context) []slog.Attr {
		atomic.AddInt64(v, 1)
		return v1(ctx)
	}
}

func TestCtxHandler_Handle(t *testing.T) {
	t.Parallel()

	th := &mockCtxHandler{}
	var calls int64

	handler := NewCtxHandler(th, testComputeCalls(&calls, extractorFunc1), testComputeCalls(&calls, extractorFunc2))

	ctx := context.Background()
	ctx = context.WithValue(ctx, "key1", "val1") // nolint
	ctx = context.WithValue(ctx, "key2", "val2") // nolint
	rec := slog.NewRecord(time.Now(), slog.LevelInfo, "test message", 0)

	if err := handler.Handle(ctx, rec); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !th.handleCalled {
		t.Error("expected Handle to be called on next handler")
	}

	if atomic.LoadInt64(&calls) != 2 {
		t.Errorf("expected 2 attributes, got %d", rec.NumAttrs())
	}
}

func TestCtxHandler_WithAttrs(t *testing.T) {
	t.Parallel()

	th := &mockCtxHandler{}
	handler := NewCtxHandler(th)

	attrs := []slog.Attr{slog.String("test", "value")}
	newHandler := handler.WithAttrs(attrs)

	th2, ok := newHandler.(*ctxHandler)
	if !ok {
		t.Fatalf("expected new handler of type *ctxHandler, got %T", newHandler)
	}

	inner := th2.next
	inner.WithAttrs(nil)

	if !th.withAttrsCalled {
		t.Error("expected WithAttrs to be called on next handler")
	}
}

func TestCtxHandler_WithGroup(t *testing.T) {
	t.Parallel()

	th := &mockCtxHandler{}
	handler := NewCtxHandler(th)
	groupName := "group1"
	newHandler := handler.WithGroup(groupName)

	th2, ok := newHandler.(*ctxHandler)
	if !ok {
		t.Fatalf("expected new handler of type *ctxHandler, got %T", newHandler)
	}

	inner := th2.next
	inner.WithGroup("")

	if !th.withGroupCalled {
		t.Error("expected WithGroup to be called on next handler")
	}
}
