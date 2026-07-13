// Package xerrors provides structured domain error types for the 3x-ui panel.
//
// It replaces the existing pattern of wrapping plain errors with human-readable
// messages (common.NewError / common.NewErrorf) with typed errors that carry a
// machine-readable Kind, a human-readable Message, and an optional wrapped inner
// error. This lets callers:
//
//   - Switch on the error Kind at API boundaries to return the right HTTP status.
//   - Use errors.Is / errors.As to test for error categories.
//   - Preserve the full error chain for logging and debugging.
//
// Sentinel values (ErrNotFound, ErrConflict, etc.) support errors.Is matching
// without importing this package's error constructors. Wrap or Wrapf are the
// primary way to create values from an existing error.
//
// Usage:
//
//	if err != nil {
//	    return xerrors.Wrapf(err, xerrors.KindNotFound,
//	        "inbound %d not found", id)
//	}
//
//	// At the API boundary:
//	var xe *xerrors.Error
//	if errors.As(err, &xe) {
//	    status := xerrors.HTTPStatus(xe.Kind)
//	    c.JSON(status, jsonMsg(c, xe.Message, err))
//	}
package xerrors

import (
	"errors"
	"fmt"
)

// Kind classifies errors for machine-readable handling at API boundaries,
// in monitoring, and for client-side decision making.
type Kind string

const (
	// KindValidation means the request data is invalid (bad format, missing
	// required fields, out-of-range values).
	KindValidation Kind = "validation"
	// KindNotFound means the requested resource does not exist.
	KindNotFound Kind = "not_found"
	// KindConflict means the request conflicts with current state (duplicate
	// email, tag already in use).
	KindConflict Kind = "conflict"
	// KindUnauthenticated means the request lacks valid authentication
	// (missing or expired session, bad credentials).
	KindUnauthenticated Kind = "unauthenticated"
	// KindForbidden means the authenticated identity is not permitted to
	// perform the requested action.
	KindForbidden Kind = "forbidden"
	// KindQuotaExceeded means a resource limit was reached (traffic cap,
	// client count ceiling, IP limit).
	KindQuotaExceeded Kind = "quota_exceeded"
	// KindRateLimited means the client has sent too many requests.
	KindRateLimited Kind = "rate_limited"
	// KindUnavailable means an external dependency is unreachable or
	// unhealthy (xray-core not running, database down, remote node
	// unreachable, provider API timeout).
	KindUnavailable Kind = "unavailable"
	// KindInternal means an unexpected error that does not fit any other
	// category (should never be shown to end users without context).
	KindInternal Kind = "internal"
)

// Error is a structured domain error. It carries a machine-readable Kind, a
// human-readable Message, and an optional wrapped inner error (Err). It
// implements the error interface and supports errors.Is / errors.As via its
// Is method and Unwrap method.
type Error struct {
	// Kind classifies the error for API status mapping and client logic.
	Kind Kind
	// Message is a human-readable description. It should be safe to return
	// to the frontend but may contain dynamic values (resource names, IDs).
	Message string
	// Err is the optional wrapped cause. Set by Wrap / Wrapf or manually.
	Err error
}

// Error satisfies the error interface.
func (e *Error) Error() string {
	if e == nil {
		return "<nil>"
	}
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Kind, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Kind, e.Message)
}

// Unwrap returns the wrapped inner error so errors.Is and errors.As can
// traverse the chain.
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

// Is enables sentinel matching via errors.Is. An Error matches a target
// *Error when their Kind values are equal, regardless of Message or Err.
// This lets callers write:
//
//	if errors.Is(err, xerrors.ErrNotFound) { ... }
//
// without comparing the full error value.
func (e *Error) Is(target error) bool {
	if e == nil || target == nil {
		return e == target
	}
	t, ok := target.(*Error)
	if !ok {
		return false
	}
	return e.Kind == t.Kind
}

// sentinel values for errors.Is matching. Each carries only a Kind; the
// Message is empty so callers can distinguish category from detail.
var (
	ErrValidation      = &Error{Kind: KindValidation}
	ErrNotFound        = &Error{Kind: KindNotFound}
	ErrConflict        = &Error{Kind: KindConflict}
	ErrUnauthenticated = &Error{Kind: KindUnauthenticated}
	ErrForbidden       = &Error{Kind: KindForbidden}
	ErrQuotaExceeded   = &Error{Kind: KindQuotaExceeded}
	ErrRateLimited     = &Error{Kind: KindRateLimited}
	ErrUnavailable     = &Error{Kind: KindUnavailable}
	ErrInternal        = &Error{Kind: KindInternal}
)

// --- Constructors for top-level errors (no underlying cause) ---

// ValidationError builds a KindValidation error with a formatted message.
func ValidationError(msg string, a ...any) *Error {
	return &Error{Kind: KindValidation, Message: fmt.Sprintf(msg, a...)}
}

// NotFoundError builds a KindNotFound error with a formatted message.
func NotFoundError(msg string, a ...any) *Error {
	return &Error{Kind: KindNotFound, Message: fmt.Sprintf(msg, a...)}
}

// ConflictError builds a KindConflict error with a formatted message.
func ConflictError(msg string, a ...any) *Error {
	return &Error{Kind: KindConflict, Message: fmt.Sprintf(msg, a...)}
}

// UnauthenticatedError builds a KindUnauthenticated error with a formatted message.
func UnauthenticatedError(msg string, a ...any) *Error {
	return &Error{Kind: KindUnauthenticated, Message: fmt.Sprintf(msg, a...)}
}

// ForbiddenError builds a KindForbidden error with a formatted message.
func ForbiddenError(msg string, a ...any) *Error {
	return &Error{Kind: KindForbidden, Message: fmt.Sprintf(msg, a...)}
}

// QuotaExceededError builds a KindQuotaExceeded error with a formatted message.
func QuotaExceededError(msg string, a ...any) *Error {
	return &Error{Kind: KindQuotaExceeded, Message: fmt.Sprintf(msg, a...)}
}

// RateLimitedError builds a KindRateLimited error with a formatted message.
func RateLimitedError(msg string, a ...any) *Error {
	return &Error{Kind: KindRateLimited, Message: fmt.Sprintf(msg, a...)}
}

// UnavailableError builds a KindUnavailable error with a formatted message.
func UnavailableError(msg string, a ...any) *Error {
	return &Error{Kind: KindUnavailable, Message: fmt.Sprintf(msg, a...)}
}

// InternalError builds a KindInternal error with a formatted message.
func InternalError(msg string, a ...any) *Error {
	return &Error{Kind: KindInternal, Message: fmt.Sprintf(msg, a...)}
}

// --- Wrappers (preserve an underlying cause) ---

// Wrap wraps an existing error with a kind and a contextual message. If err
// is nil, Wrap returns nil so callers can wrap unconditionally:
//
//	return xerrors.Wrap(err, xerrors.KindInternal, "flush failed")
func Wrap(err error, kind Kind, msg string) *Error {
	if err == nil {
		return nil
	}
	return &Error{Kind: kind, Message: msg, Err: err}
}

// Wrapf is like Wrap but accepts a format string.
func Wrapf(err error, kind Kind, format string, a ...any) *Error {
	if err == nil {
		return nil
	}
	return &Error{Kind: kind, Message: fmt.Sprintf(format, a...), Err: err}
}

// --- Query helpers ---

// KindOf extracts the Kind from an error, returning KindInternal for
// unrecognized errors. Use when you need the kind without a full errors.As.
func KindOf(err error) Kind {
	if err == nil {
		return ""
	}
	var e *Error
	if !errors.As(err, &e) {
		return KindInternal
	}
	return e.Kind
}

// MessageOf extracts the human-readable message from an Error, or the
// error text for unrecognized errors.
func MessageOf(err error) string {
	if err == nil {
		return ""
	}
	var e *Error
	if errors.As(err, &e) {
		return e.Message
	}
	return err.Error()
}
