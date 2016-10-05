package httphelper

import (
	"context"
	"encoding/json"
	"net/http"

	"dmitryfrank.com/geekmarks/server/middleware"

	"github.com/golang/glog"
	"github.com/juju/errors"
)

var (
	internalServerError error
	unauthorizedError   error
	forbiddenError      error
)

func init() {
	internalServerError = errors.New("internal server error")
	unauthorizedError = errors.New("unauthorized")
	forbiddenError = errors.New("forbidden")
}

type ErrorResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
}

const (
	DesiredContentTypeKey = "desiredContentType"
)

func RespondWithError(w http.ResponseWriter, r *http.Request, errResp error) {
	httpErrorCode := getHTTPErrorCode(errResp)

	desiredContentType := "text/html"

	v := r.Context().Value(DesiredContentTypeKey)
	if v != nil {
		var ok bool
		desiredContentType, ok = v.(string)
		if !ok {
			glog.Errorf("wrong type of desiredContentType: %T (%v)",
				desiredContentType, desiredContentType)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}

	switch desiredContentType {
	case "application/json":
		resp := ErrorResponse{
			Status:  httpErrorCode,
			Message: errResp.Error(),
		}
		d, err := json.Marshal(resp)
		if err != nil {
			panic(err)
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(httpErrorCode)
		_, err = w.Write(d)
		if err != nil {
			panic(err)
		}
	case "text/html":
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(httpErrorCode)
		_, err := w.Write([]byte("Error: " + errResp.Error()))
		if err != nil {
			panic(err)
		}
	default:
		glog.Errorf("wrong desiredContentType: %q", desiredContentType)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

func MakeAPIHandler(
	f func(r *http.Request) (resp interface{}, err error),
) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := f(r)
		if err != nil {
			RespondWithError(w, r, err)
			return
		}

		d, err := json.Marshal(resp)
		if err != nil {
			RespondWithError(w, r, MakeInternalServerError(err, "marshalling resp"))
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.Write(d)
		if err != nil {
			panic(err)
		}
	}
}

// MakeInternalServerError logs the given error and returns internalServerError
// annotated with the message, which does NOT wrap the original error, since we
// don't want internal server error details to percolate to clients.
func MakeInternalServerError(err error, message string) error {
	glog.Errorf("%s: %s", message, errors.Trace(err))
	if errors.Cause(err) != internalServerError {
		err = errors.Annotatef(internalServerError, message)
	}
	return err
}

func MakeUnauthorizedError() error {
	return unauthorizedError
}

func MakeForbiddenError() error {
	return forbiddenError
}

func getHTTPErrorCode(err error) int {
	status := http.StatusBadRequest

	switch errors.Cause(err) {
	case internalServerError:
		status = http.StatusInternalServerError
	case unauthorizedError:
		status = http.StatusUnauthorized
	case forbiddenError:
		status = http.StatusForbidden
	}

	return status
}

func MakeDesiredContentTypeMiddleware(
	contentType string,
) func(inner http.Handler) http.Handler {
	return func(inner http.Handler) http.Handler {
		mw := func(w http.ResponseWriter, r *http.Request) {
			// Process request
			inner.ServeHTTP(w, r.WithContext(context.WithValue(
				r.Context(), DesiredContentTypeKey, contentType,
			)))
		}
		return middleware.MkMiddleware(mw)
	}
}
