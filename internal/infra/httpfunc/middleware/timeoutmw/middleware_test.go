package timeoutmw

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func timeoutHandleFunc(sleep time.Duration, statusCode int) http.HandlerFunc { // nolint
	if statusCode == 0 {
		statusCode = http.StatusOK
	}

	return func(w http.ResponseWriter, r *http.Request) {
		timer := time.NewTimer(sleep)
		defer timer.Stop()

		select {
		case <-r.Context().Done():
			time.Sleep(50 * time.Millisecond)
			w.WriteHeader(http.StatusInternalServerError)
			return
		case <-timer.C:
		}

		w.WriteHeader(statusCode)
	}
}

func TestTimeoutMiddleware_Success(t *testing.T) {
	t.Parallel()

	middleware := TimeoutMiddleware(WithTimeout(5 * time.Second))
	server := httptest.NewServer(middleware(timeoutHandleFunc(5*time.Nanosecond, http.StatusOK)))
	defer server.Close()

	resp, err := http.Get(server.URL) // nolint
	require.NoError(t, err)

	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestTimeoutMiddleware_Timeout(t *testing.T) {
	t.Parallel()

	middleware := TimeoutMiddleware(WithTimeout(300 * time.Millisecond))
	server := httptest.NewServer(middleware(timeoutHandleFunc(1*time.Second, http.StatusOK)))
	defer server.Close()

	resp, err := http.Get(server.URL) // nolint
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusGatewayTimeout, resp.StatusCode)
}

func TestTimeoutMiddleware_ClientCancel(t *testing.T) {
	t.Parallel()

	middleware := TimeoutMiddleware(WithTimeout(5 * time.Second))
	server := httptest.NewServer(middleware(timeoutHandleFunc(2*time.Second, http.StatusOK)))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, server.URL, nil)
	require.NoError(t, err)

	time.AfterFunc(200*time.Millisecond, cancel)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil && !errors.Is(err, context.Canceled) {
		t.Fatalf("Request failed: %v", err)
	}
	if err == nil {
		defer resp.Body.Close()
	}

	assert.ErrorIs(t, err, context.Canceled)
}

func TestTimeoutMiddleware_ConcurrentWrite(t *testing.T) {
	t.Parallel()

	middleware := TimeoutMiddleware(WithTimeout(1 * time.Second))
	server := httptest.NewServer(middleware(timeoutHandleFunc(2*time.Second, http.StatusOK)))
	defer server.Close()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		resp, err := http.Get(server.URL) // nolint
		require.NoError(t, err)
		defer resp.Body.Close()
		t.Logf("Request 1 status code: %d", resp.StatusCode)
	}()

	go func() {
		defer wg.Done()
		resp, err := http.Get(server.URL) // nolint
		require.NoError(t, err)
		defer resp.Body.Close()
		t.Logf("Request 2 status code: %d", resp.StatusCode)
	}()

	wg.Wait()
}
