package xerrors

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Msg is the standard JSON response envelope used by 3x-ui controllers.
// It mirrors entity.Msg but lives here so xerrors can produce it without
// importing the entity package (avoiding a potential import cycle).
type Msg struct {
	Success bool   `json:"success"`
	Msg     string `json:"msg"`
	Obj     any    `json:"obj,omitempty"`
}

// AbortWithError aborts the Gin request chain with the correct HTTP status
// code and JSON body derived from err. If err is nil it panics — callers
// must only pass non-nil errors:
//
//	if err != nil {
//	    xerrors.AbortWithError(c, err)
//	    return
//	}
//
// The response body uses the standard {success, msg, obj} envelope so the
// frontend's existing error-handling logic (which inspects success==false)
// continues to work unchanged. The only difference from the old jsonMsg
// helper is that the HTTP status code now reflects the error kind.
func AbortWithError(c *gin.Context, err error) {
	status := HTTPStatus(err)
	AbortWithHTTPError(c, status, err)
}

// AbortWithHTTPError aborts the Gin request chain with the given HTTP status
// code and a JSON error body derived from err. Use this when you need to
// override the automatic status mapping from AbortWithError.
//
// The msg field is set to the human-readable Message from an *Error, or the
// error text for unrecognized errors. The obj field is always nil on error
// responses.
func AbortWithHTTPError(c *gin.Context, status int, err error) {
	if err == nil {
		return
	}
	var msg string
	if e, ok := err.(*Error); ok {
		msg = e.Message
		if msg == "" {
			msg = err.Error()
		}
	} else {
		msg = err.Error()
	}
	c.AbortWithStatusJSON(status, Msg{
		Success: false,
		Msg:     msg,
	})
}

// AbortWithValidation is a convenience wrapper for validation errors. It
// sends a 400 Bad Request with the formatted message.
func AbortWithValidation(c *gin.Context, format string, a ...any) {
	AbortWithHTTPError(c, http.StatusBadRequest, ValidationError(format, a...))
}

// AbortWithNotFound is a convenience wrapper for not-found errors (404).
func AbortWithNotFound(c *gin.Context, format string, a ...any) {
	AbortWithHTTPError(c, http.StatusNotFound, NotFoundError(format, a...))
}

// AbortWithConflict is a convenience wrapper for conflict errors (409).
func AbortWithConflict(c *gin.Context, format string, a ...any) {
	AbortWithHTTPError(c, http.StatusConflict, ConflictError(format, a...))
}

// AbortWithUnauthenticated is a convenience wrapper for auth errors (401).
func AbortWithUnauthenticated(c *gin.Context, format string, a ...any) {
	AbortWithHTTPError(c, http.StatusUnauthorized, UnauthenticatedError(format, a...))
}

// AbortWithForbidden is a convenience wrapper for permission errors (403).
func AbortWithForbidden(c *gin.Context, format string, a ...any) {
	AbortWithHTTPError(c, http.StatusForbidden, ForbiddenError(format, a...))
}

// AbortWithInternal aborts with a generic 500 Internal Server Error. The
// msg is logged server-side but the response only says "internal error" to
// avoid leaking implementation details to the client.
func AbortWithInternal(c *gin.Context, err error) {
	if err == nil {
		return
	}
	// Log the real error server-side.
	c.Error(err)
	// Return a sanitized message to the client.
	AbortWithHTTPError(c, http.StatusInternalServerError, InternalError("internal error"))
}

// JSON sends a success response with the given msg and obj. It is the
// success-path counterpart of AbortWithError.
func JSON(c *gin.Context, msg string, obj any) {
	c.JSON(http.StatusOK, Msg{
		Success: true,
		Msg:     msg,
		Obj:     obj,
	})
}
