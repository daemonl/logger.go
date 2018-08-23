package logger

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
)

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		begin := time.Now()

		entry := FromContext(req.Context()).
			WithField("serving", map[string]interface{}{
				"path":   req.URL.Path,
				"method": req.Method,
				"query":  req.URL.Query(),
				"host":   req.Host,
			})

		recorder := &responseRecorder{
			status:         200,
			ResponseWriter: rw,
		}

		req = req.WithContext(WithEntry(req.Context(), entry))
		next.ServeHTTP(recorder, req)

		if req.URL.Path == "/up" {
			return
		}

		entry.WithFields(logrus.Fields{
			"statusCode":      recorder.status,
			"durationSeconds": time.Now().Sub(begin).Seconds(),
			"begin":           begin.Format(time.RFC3339),
			"agent":           req.UserAgent(),
			"referer":         req.Referer(),
		}).Info("HTTP Served Request")
	})
}

type responseRecorder struct {
	status int
	http.ResponseWriter
}

func (rr *responseRecorder) WriteHeader(status int) {
	rr.status = status
	rr.ResponseWriter.WriteHeader(status)
}

func statusCodeFamily(status int) string {
	return fmt.Sprintf("%d", status)[0:1] + "XX"
}
