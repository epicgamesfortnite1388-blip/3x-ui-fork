package xerrors

import (
	"errors"
	"net/http"
)

// HTTPStatus maps an error's Kind to an HTTP status code. For plain errors
// (not *xerrors.Error), it defaults to 500 Internal Server Error.
//
// Controllers use this at the API boundary to set the response code:
//
//	if err != nil {
//	    status := xerrors.HTTPStatus(err)
//	    c.JSON(status, gin.H{"success": false, "msg": err.Error()})
//	    return
//	}
func HTTPStatus(err error) int {
	if err == nil {
		return http.StatusOK
	}
	var e *Error
	if !errors.As(err, &e) {
		return http.StatusInternalServerError
	}
	switch e.Kind {
	case KindValidation:
		return http.StatusBadRequest
	case KindNotFound:
		return http.StatusNotFound
	case KindConflict:
		return http.StatusConflict
	case KindUnauthenticated:
		return http.StatusUnauthorized
	case KindForbidden:
		return http.StatusForbidden
	case KindQuotaExceeded:
		return http.StatusTooManyRequests
	case KindRateLimited:
		return http.StatusTooManyRequests
	case KindUnavailable:
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}
