package xerrors

import (
	"errors"
	"fmt"
	"testing"
)

func TestErrorImplementsError(t *testing.T) {
	var err error = &Error{Kind: KindNotFound, Message: "test"}
	if err.Error() == "" {
		t.Fatal("expected non-empty Error()")
	}
}

func TestErrorFormatting_noCause(t *testing.T) {
	err := &Error{Kind: KindNotFound, Message: "inbound not found"}
	want := "not_found: inbound not found"
	if got := err.Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestErrorFormatting_withCause(t *testing.T) {
	cause := errors.New("record not found")
	err := &Error{Kind: KindNotFound, Message: "inbound", Err: cause}
	want := "not_found: inbound: record not found"
	if got := err.Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestErrorFormatting_nil(t *testing.T) {
	var err *Error
	if got := err.Error(); got != "<nil>" {
		t.Errorf("Error() = %q, want %q", got, "<nil>")
	}
}

func TestErrorIs_matchesSameKind(t *testing.T) {
	err := &Error{Kind: KindNotFound, Message: "specific inbound"}
	if !errors.Is(err, ErrNotFound) {
		t.Fatal("expected errors.Is to match ErrNotFound")
	}
}

func TestErrorIs_notMatchDifferentKind(t *testing.T) {
	err := &Error{Kind: KindNotFound, Message: "inbound"}
	if errors.Is(err, ErrValidation) {
		t.Fatal("expected errors.Is to NOT match ErrValidation")
	}
}

func TestErrorIs_notMatchPlainError(t *testing.T) {
	err := &Error{Kind: KindInternal, Message: "db error"}
	plain := errors.New("plain error")
	if errors.Is(err, plain) {
		t.Fatal("expected errors.Is to NOT match a plain error")
	}
}

func TestErrorIs_matchesWrappedChain(t *testing.T) {
	inner := &Error{Kind: KindNotFound, Message: "inner"}
	outer := fmt.Errorf("outer: %w", inner)
	if !errors.Is(outer, ErrNotFound) {
		t.Fatal("expected errors.Is through a fmt.Errorf wrapper to match ErrNotFound")
	}
}

func TestErrorAs_extractsStructuredError(t *testing.T) {
	inner := &Error{Kind: KindValidation, Message: "port must be 1-65535"}
	outer := fmt.Errorf("wrap: %w", inner)

	var extracted *Error
	if !errors.As(outer, &extracted) {
		t.Fatal("expected errors.As to succeed")
	}
	if extracted.Kind != KindValidation {
		t.Errorf("Kind = %q, want %q", extracted.Kind, KindValidation)
	}
	if extracted.Message != "port must be 1-65535" {
		t.Errorf("Message = %q, want %q", extracted.Message, "port must be 1-65535")
	}
}

func TestErrorUnwrap(t *testing.T) {
	cause := errors.New("inner cause")
	err := &Error{Kind: KindUnavailable, Message: "xray not running", Err: cause}
	if err.Unwrap() != cause {
		t.Fatal("Unwrap() should return the inner error")
	}
}

func TestErrorUnwrap_nil(t *testing.T) {
	var err *Error
	if err.Unwrap() != nil {
		t.Fatal("nil Error should return nil from Unwrap")
	}
}

func TestConstructors(t *testing.T) {
	tests := []struct {
		name     string
		fn       func(string, ...any) *Error
		kind     Kind
		msg      string
	}{
		{"ValidationError", ValidationError, KindValidation, "bad port"},
		{"NotFoundError", NotFoundError, KindNotFound, "client not found"},
		{"ConflictError", ConflictError, KindConflict, "email exists"},
		{"UnauthenticatedError", UnauthenticatedError, KindUnauthenticated, "session expired"},
		{"ForbiddenError", ForbiddenError, KindForbidden, "not an admin"},
		{"QuotaExceededError", QuotaExceededError, KindQuotaExceeded, "traffic limit"},
		{"RateLimitedError", RateLimitedError, KindRateLimited, "too fast"},
		{"UnavailableError", UnavailableError, KindUnavailable, "node down"},
		{"InternalError", InternalError, KindInternal, "bug"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn(tt.msg)
			if err.Kind != tt.kind {
				t.Errorf("Kind = %q, want %q", err.Kind, tt.kind)
			}
			if err.Message != tt.msg {
				t.Errorf("Message = %q, want %q", err.Message, tt.msg)
			}
			if err.Err != nil {
				t.Errorf("Err = %v, want nil", err.Err)
			}
		})
	}
}

func TestConstructors_formatted(t *testing.T) {
	err := NotFoundError("client %q not found in inbound %d", "alice", 7)
	want := "client \"alice\" not found in inbound 7"
	if err.Message != want {
		t.Errorf("Message = %q, want %q", err.Message, want)
	}
}

func TestWrap_nilErr(t *testing.T) {
	if got := Wrap(nil, KindInternal, "should be nil"); got != nil {
		t.Fatal("Wrap(nil, ...) should return nil")
	}
}

func TestWrap_wrapsInner(t *testing.T) {
	cause := errors.New("disk full")
	err := Wrap(cause, KindUnavailable, "xray log write failed")
	if err == nil {
		t.Fatal("expected non-nil error")
	}
	if err.Kind != KindUnavailable {
		t.Errorf("Kind = %q, want %q", err.Kind, KindUnavailable)
	}
	if err.Err != cause {
		t.Fatal("Err should be the original cause")
	}
	if !errors.Is(err, ErrUnavailable) {
		t.Fatal("errors.Is should match ErrUnavailable")
	}
}

func TestWrapf_nilErr(t *testing.T) {
	if got := Wrapf(nil, KindInternal, "x"); got != nil {
		t.Fatal("Wrapf(nil, ...) should return nil")
	}
}

func TestWrapf_formatsMessage(t *testing.T) {
	cause := errors.New("closed pipe")
	err := Wrapf(cause, KindUnavailable, "node %d unreachable", 5)
	if err.Message != "node 5 unreachable" {
		t.Errorf("Message = %q, want %q", err.Message, "node 5 unreachable")
	}
}

func TestKindOf(t *testing.T) {
	if kind := KindOf(nil); kind != "" {
		t.Errorf("KindOf(nil) = %q, want empty", kind)
	}
	if kind := KindOf(errors.New("plain")); kind != KindInternal {
		t.Errorf("KindOf(plain) = %q, want %q", kind, KindInternal)
	}
	err := &Error{Kind: KindNotFound}
	if kind := KindOf(err); kind != KindNotFound {
		t.Errorf("KindOf(notFound) = %q, want %q", kind, KindNotFound)
	}
}

func TestMessageOf(t *testing.T) {
	if msg := MessageOf(nil); msg != "" {
		t.Errorf("MessageOf(nil) = %q, want empty", msg)
	}
	if msg := MessageOf(errors.New("plain")); msg != "plain" {
		t.Errorf("MessageOf(plain) = %q, want %q", msg, "plain")
	}
	err := &Error{Kind: KindValidation, Message: "invalid port"}
	if msg := MessageOf(err); msg != "invalid port" {
		t.Errorf("MessageOf = %q, want %q", msg, "invalid port")
	}
}

func TestSentinelValues_areDistinct(t *testing.T) {
	sentinels := []*Error{
		ErrValidation, ErrNotFound, ErrConflict, ErrUnauthenticated,
		ErrForbidden, ErrQuotaExceeded, ErrRateLimited, ErrUnavailable, ErrInternal,
	}
	seen := make(map[Kind]bool)
	for _, s := range sentinels {
		if seen[s.Kind] {
			t.Errorf("duplicate sentinel Kind: %q", s.Kind)
		}
		seen[s.Kind] = true
	}
}
