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

		entry := New().WithFields(logrus.Fields{
			"path":   req.URL.Path,
			"method": req.Method,
			"query":  req.URL.Query(),
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
			"host":            req.Host,
			"statusCode":      recorder.status,
			"durationSeconds": time.Now().Sub(begin).Seconds(),
			"begin":           begin.Format(time.RFC3339),
			"agent":           req.UserAgent(),
			"referer":         req.Referer(),
		}).Info("HTTP Request")
	})
}

type responseRecorder struct {
	status int
	http.ResponseWriter
}

func (rr *responseRecorder) WriteHeader(status int) {
	rr.ResponseWriter.WriteHeader(status)
}

func statusCodeFamily(status int) string {
	return fmt.Sprintf("%d", status)[0:1] + "XX"
}
