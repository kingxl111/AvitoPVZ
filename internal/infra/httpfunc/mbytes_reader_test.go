package httpfunc

import (
	"bytes"
	"errors"
	"io"
	"testing"
)

func TestCopyNegative(t *testing.T) {
	t.Parallel()

	rb := new(bytes.Buffer)
	wb := new(bytes.Buffer)

	rb.WriteString("hello")
	if _, err := io.Copy(wb, NewMaxBytesReader(rb, -1)); err != nil {
		if !errors.Is(err, ErrMaxLimit) {
			t.Errorf("Copy error: got %v, want %v", err, ErrMaxLimit)
		}
	}
	if wb.String() != "" {
		t.Errorf("Copy on MaxBytesReader with N<0 copied data")
	}

	if _, err := io.CopyN(wb, rb, -1); err != nil {
		t.Fatalf("CopyN error: got %v, want nil", err)
	}

	if wb.String() != "" {
		t.Errorf("CopyN with N<0 copied data")
	}
}
