package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"time"

	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/ipoool/jastipin/backend/internal/domain"
	"github.com/ipoool/jastipin/backend/internal/http/response"
)

// RequestLogger mencatat setiap request dengan durasi dan status responsnya.
func RequestLogger(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Health check dipanggil terus-menerus oleh Docker; mencatatnya
			// hanya membuat log membanjir tanpa informasi berguna.
			if r.URL.Path == "/health" || r.URL.Path == "/health/ready" {
				next.ServeHTTP(w, r)
				return
			}

			start := time.Now()
			ww := chimw.NewWrapResponseWriter(w, r.ProtoMajor)

			next.ServeHTTP(ww, r)

			attrs := []any{
				"method", r.Method,
				"path", r.URL.Path,
				"status", ww.Status(),
				"bytes", ww.BytesWritten(),
				"duration_ms", time.Since(start).Milliseconds(),
				"request_id", chimw.GetReqID(r.Context()),
			}
			if user, ok := UserFrom(r.Context()); ok {
				attrs = append(attrs, "user_id", user.ID.String())
			}

			switch {
			case ww.Status() >= 500:
				log.ErrorContext(r.Context(), "request selesai dengan error server", attrs...)
			case ww.Status() >= 400:
				log.WarnContext(r.Context(), "request ditolak", attrs...)
			default:
				log.InfoContext(r.Context(), "request selesai", attrs...)
			}
		})
	}
}

// Recoverer menangkap panic supaya satu handler bermasalah tidak menjatuhkan
// seluruh proses server.
func Recoverer(log *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				rec := recover()
				if rec == nil {
					return
				}
				// http.ErrAbortHandler adalah cara idiomatik menghentikan
				// handler; teruskan supaya net/http menanganinya sendiri.
				if rec == http.ErrAbortHandler {
					panic(rec)
				}

				log.ErrorContext(r.Context(), "panic pada handler",
					"method", r.Method,
					"path", r.URL.Path,
					"panic", rec,
					"stack", string(debug.Stack()),
				)
				response.Error(w, r, domain.Internal(nil))
			}()

			next.ServeHTTP(w, r)
		})
	}
}
