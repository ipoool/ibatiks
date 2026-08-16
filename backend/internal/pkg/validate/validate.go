// Package validate membungkus go-playground/validator dan menerjemahkan
// error-nya menjadi pesan berbahasa Indonesia per field.
package validate

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/ipoool/jastipin/backend/internal/domain"
)

var instance *validator.Validate

func init() {
	instance = validator.New(validator.WithRequiredStructEnabled())

	// Pakai nama pada tag json sebagai nama field di pesan error, supaya nama
	// yang muncul di UI sama persis dengan yang dikirim client.
	instance.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "" || name == "-" {
			return fld.Name
		}
		return name
	})
}

// Struct memvalidasi payload dan mengembalikan domain.Error berisi peta
// field -> pesan bila ada yang tidak lolos.
func Struct(payload any) error {
	if err := instance.Struct(payload); err != nil {
		var invalid *validator.InvalidValidationError
		if errors.As(err, &invalid) {
			return domain.Internal(err)
		}

		var validationErrs validator.ValidationErrors
		if errors.As(err, &validationErrs) {
			fields := make(map[string]string, len(validationErrs))
			for _, fe := range validationErrs {
				fields[fe.Field()] = message(fe)
			}
			return domain.Validation("data yang dikirim tidak valid", fields)
		}
		return domain.Internal(err)
	}
	return nil
}

func message(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "wajib diisi"
	case "email":
		return "format email tidak valid"
	case "min":
		if fe.Kind() == reflect.String {
			return fmt.Sprintf("minimal %s karakter", fe.Param())
		}
		return fmt.Sprintf("nilai minimal %s", fe.Param())
	case "max":
		if fe.Kind() == reflect.String {
			return fmt.Sprintf("maksimal %s karakter", fe.Param())
		}
		return fmt.Sprintf("nilai maksimal %s", fe.Param())
	case "gt":
		return fmt.Sprintf("harus lebih besar dari %s", fe.Param())
	case "gte":
		return fmt.Sprintf("minimal %s", fe.Param())
	case "lte":
		return fmt.Sprintf("maksimal %s", fe.Param())
	case "oneof":
		return fmt.Sprintf("harus salah satu dari: %s", strings.ReplaceAll(fe.Param(), " ", ", "))
	case "uuid", "uuid4":
		return "harus berupa UUID yang valid"
	case "url":
		return "harus berupa URL yang valid"
	case "len":
		return fmt.Sprintf("panjangnya harus %s", fe.Param())
	case "numeric":
		return "harus berupa angka"
	case "required_if", "required_with":
		return "wajib diisi untuk kombinasi data ini"
	case "datetime":
		return fmt.Sprintf("format tanggal harus %s", fe.Param())
	case "eqfield":
		return fmt.Sprintf("harus sama dengan %s", fe.Param())
	case "gtefield":
		return fmt.Sprintf("tidak boleh lebih kecil dari %s", fe.Param())
	default:
		return fmt.Sprintf("tidak memenuhi aturan %s", fe.Tag())
	}
}
