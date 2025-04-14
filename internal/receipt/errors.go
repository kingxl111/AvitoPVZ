package receipt

import (
	"github.com/pkg/errors"
)

var (
	ErrLastReceiptNotClosed = errors.New("last receipt not closed")
	ErrLastReceiptClosed    = errors.New("last receipt closed")
	ErrReceiptNotFound      = errors.New("receipt not found")
	ErrReceiptEmpty         = errors.New("receipt empty")
)
