package trace

import (
	"context"
	"net/http"

	"github.com/pborman/uuid"
)

const (
	TRACE_HEADER = "X-Trace-ID"
)

type traceKeyType struct{}

var traceKey = traceKeyType{}

func WithTrace(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, traceKey, id)
}

func GetTrace(ctx context.Context) (string, bool) {
	if v, ok := ctx.Value(traceKey).(string); !ok {
		return "", false
	} else {
		return v, true
	}
}

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		tracingID := req.Header.Get(TRACE_HEADER)
		if tracingID == "" {
			tracingID = uuid.New()
		}
		req = req.WithContext(WithTrace(req.Context(), tracingID))
		next.ServeHTTP(rw, req)
	})
}

type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

type DoerFunc func(*http.Request) (*http.Response, error)

func (fn DoerFunc) Do(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func Outerware(next Doer) Doer {
	return DoerFunc(func(req *http.Request) (*http.Response, error) {
		if trace, ok := GetTrace(req.Context()); ok {
			req.Header.Add(TRACE_HEADER, trace)
		}
		return next.Do(req)
	})
}
