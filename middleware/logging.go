package middleware

import (
	"log"
	"net/http"
	"time"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func NewLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func Logging() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t1 := time.Now()
			lrw := NewLoggingResponseWriter(w)
			next.ServeHTTP(lrw, r)
			t2 := time.Now()
			statusCode := lrw.statusCode
			log.Printf("[%s] %q %v %d", r.Method, r.URL.String(), t2.Sub(t1), statusCode)
		})
	}
}
