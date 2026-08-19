// Package domain berisi entitas bisnis, enum, dan tipe error yang dipakai
// lintas layer. Package ini sengaja tidak mengimpor apa pun dari layer HTTP
// maupun database agar aturan bisnis tetap bisa diuji tanpa keduanya.
package domain

import (
	"errors"
	"fmt"
)

type ErrorCode string

const (
	CodeNotFound     ErrorCode = "NOT_FOUND"
	CodeConflict     ErrorCode = "CONFLICT"
	CodeValidation   ErrorCode = "VALIDATION_ERROR"
	CodeUnauthorized ErrorCode = "UNAUTHORIZED"
	CodeForbidden    ErrorCode = "FORBIDDEN"
	CodeInvalidState ErrorCode = "INVALID_STATE"
	CodeTooMany      ErrorCode = "TOO_MANY_REQUESTS"
	CodeInternal     ErrorCode = "INTERNAL_ERROR"
)

// Error adalah error bisnis yang membawa kode agar layer HTTP bisa memetakannya
// ke status code tanpa perlu tahu detail domain.
type Error struct {
	Code    ErrorCode
	Message string
	// Fields diisi untuk error validasi: nama field -> pesan.
	Fields map[string]string
	cause  error
}

func (e *Error) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *Error) Unwrap() error { return e.cause }

// WithCause menempelkan error asli untuk keperluan logging tanpa membocorkannya
// ke response HTTP.
func (e *Error) WithCause(err error) *Error {
	e.cause = err
	return e
}

func NotFound(entity string) *Error {
	return &Error{Code: CodeNotFound, Message: fmt.Sprintf("%s tidak ditemukan", entity)}
}

func Conflict(format string, args ...any) *Error {
	return &Error{Code: CodeConflict, Message: fmt.Sprintf(format, args...)}
}

func Validation(message string, fields map[string]string) *Error {
	return &Error{Code: CodeValidation, Message: message, Fields: fields}
}

func Validationf(format string, args ...any) *Error {
	return &Error{Code: CodeValidation, Message: fmt.Sprintf(format, args...)}
}

// InvalidState dipakai saat operasi ditolak karena status entitas tidak
// mengizinkannya, misalnya mengedit order yang sudah dikirim.
func InvalidState(format string, args ...any) *Error {
	return &Error{Code: CodeInvalidState, Message: fmt.Sprintf(format, args...)}
}

func Unauthorized(message string) *Error {
	return &Error{Code: CodeUnauthorized, Message: message}
}

func Forbidden(message string) *Error {
	return &Error{Code: CodeForbidden, Message: message}
}

// TooMany dipakai saat permintaan ditolak karena terlalu sering dicoba, bukan
// karena isinya salah. Dipisah dari Unauthorized supaya frontend bisa
// membedakan "passwordmu salah" dari "tunggu dulu".
func TooMany(format string, args ...any) *Error {
	return &Error{Code: CodeTooMany, Message: fmt.Sprintf(format, args...)}
}

func Internal(err error) *Error {
	return &Error{Code: CodeInternal, Message: "terjadi kesalahan pada server", cause: err}
}

// AsError mengekstrak *Error dari rantai error, jika ada.
func AsError(err error) (*Error, bool) {
	var domainErr *Error
	if errors.As(err, &domainErr) {
		return domainErr, true
	}
	return nil, false
}
