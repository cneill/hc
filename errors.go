package hc

import (
	"errors"
	"fmt"
)

var (
	ErrMaxRedirects         = errors.New("hit maximum number of allowed redirects")
	ErrDowngradeRedirect    = errors.New("rejecting HTTPS to HTTP downgrade redirect")
	ErrNonHTTPSConnection   = errors.New("paranoid: rejecting non-HTTPS connection")
	ErrSchemeSwitchRedirect = errors.New("paranoid: rejecting scheme switch redirect")
	ErrCrossDomainRedirect  = errors.New("paranoid: rejecting cross-domain redirect")
	ErrInvalidPort          = errors.New("paranoid: invalid port")
	ErrPortMismatch         = errors.New("paranoid: port mismatch")
)

type StatusCodeError struct {
	StatusCode int
	Status     string
}

func (s StatusCodeError) Error() string {
	return fmt.Sprintf("status code: %d, message: %q", s.StatusCode, s.Status)
}

type MaxContentLengthError struct {
	MaxLength     int64
	ContentLength int64
}

func (m MaxContentLengthError) Error() string {
	return fmt.Sprintf("content length %d exceeded max length %d", m.ContentLength, m.MaxLength)
}
