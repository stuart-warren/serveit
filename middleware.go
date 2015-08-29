package main

import (
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
)

type Middleware func(http.Handler) http.Handler

func Decorate(h http.Handler, m ...Middleware) http.Handler {
	decorated := h
	// decorate is a function
	for _, decorate := range m {
		decorated = decorate(decorated)
	}
	return decorated
}

type Decorator func(Middleware) Middleware

func Logging(log *logrus.Logger) Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			wr := NewResponseWriter(w)
			h.ServeHTTP(wr, r)

			remoteAddr := r.RemoteAddr
			if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
				remoteAddr = realIP
			}

			status := wr.Status()

			latency := time.Since(start)
			entry := log.WithFields(logrus.Fields{
				"request":    r.RequestURI,
				"method":     r.Method,
				"remote":     remoteAddr,
				"status":     status,
				"latency":    latency,
				"latency_us": latency.Nanoseconds() / 1e3,
			})

			if reqID := r.Header.Get("X-Request-ID"); reqID != "" {
				entry = entry.WithField("request_id", reqID)
			}

			switch {
			case status >= 500:
				entry.Error("request")
			case status >= 400:
				entry.Warn("request")
			default:
				entry.Info("request")
			}
		})
	}
}

//func Logging(o io.Writer) Middleware {
//	return func(h http.Handler) http.Handler {
//		return handlers.CombinedLoggingHandler(o, h)
//	}
//}

// func Logging(l *log.Logger) Middleware {
// 	return func(h http.Handler) http.Handler {
// 		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 			l.Printf("%s %s %s\n", r.Proto, r.Method, r.URL)
// 			h.ServeHTTP(w, r)
// 		})
// 	}
// }
