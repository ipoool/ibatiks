// Package response menstandarkan bentuk semua response JSON API.
//
// Sukses : {"data": ..., "meta": {...}}
// Gagal  : {"error": {"code": "...", "message": "...", "fields": {...}}}
package response

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5"

	"github.com/ipoool/jastipin/backend/internal/domain"
)

type Envelope struct {
	Data  any           `json:"data,omitempty"`
	Meta  *Meta         `json:"meta,omitempty"`
	Error *ErrorPayload `json:"error,omitempty"`
}

type Meta struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

type ErrorPayload struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Fields  map[string]string `json:"fields,omitempty"`
}

func JSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if payload == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		slog.Error("gagal menulis response JSON", "error", err)
	}
}

func OK(w http.ResponseWriter, data any) {
	JSON(w, http.StatusOK, Envelope{Data: data})
}

func Created(w http.ResponseWriter, data any) {
	JSON(w, http.StatusCreated, Envelope{Data: data})
}

func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// Paginated membungkus daftar hasil beserta metadata halaman.
func Paginated(w http.ResponseWriter, data any, page, perPage int, total int64) {
	totalPages := 0
	if perPage > 0 {
		totalPages = int((total + int64(perPage) - 1) / int64(perPage))
	}
	JSON(w, http.StatusOK, Envelope{
		Data: data,
		Meta: &Meta{Page: page, PerPage: perPage, Total: total, TotalPages: totalPages},
	})
}

// Error memetakan error apa pun menjadi response JSON yang konsisten.
// Detail error internal hanya masuk log, tidak pernah dikirim ke client.
func Error(w http.ResponseWriter, r *http.Request, err error) {
	// pgx.ErrNoRows yang lolos dari repository diperlakukan sebagai 404 agar
	// kelalaian di satu repo tidak berubah jadi 500 yang membingungkan.
	if errors.Is(err, pgx.ErrNoRows) {
		err = domain.NotFound("data")
	}

	domainErr, ok := domain.AsError(err)
	if !ok {
		domainErr = domain.Internal(err)
	}

	status := statusFor(domainErr.Code)
	if status >= http.StatusInternalServerError {
		slog.ErrorContext(r.Context(), "request gagal",
			"method", r.Method, "path", r.URL.Path, "error", err)
	}

	JSON(w, status, Envelope{Error: &ErrorPayload{
		Code:    string(domainErr.Code),
		Message: domainErr.Message,
		Fields:  domainErr.Fields,
	}})
}

func statusFor(code domain.ErrorCode) int {
	switch code {
	case domain.CodeNotFound:
		return http.StatusNotFound
	case domain.CodeConflict:
		return http.StatusConflict
	case domain.CodeValidation:
		return http.StatusUnprocessableEntity
	case domain.CodeUnauthorized:
		return http.StatusUnauthorized
	case domain.CodeForbidden:
		return http.StatusForbidden
	case domain.CodeInvalidState:
		// 409: permintaannya valid, tapi status entitas saat ini menolaknya.
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
