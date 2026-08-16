// Package request berisi helper untuk membaca dan memvalidasi input HTTP.
package request

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/ipoool/jastipin/backend/internal/domain"
	"github.com/ipoool/jastipin/backend/internal/pkg/validate"
)

// maxBodyBytes membatasi ukuran body supaya request bermasalah tidak menghabiskan
// memori server. 1 MB jauh lebih dari cukup untuk seluruh form di aplikasi ini.
const maxBodyBytes = 1 << 20

// DecodeJSON membaca body JSON ke dst lalu menjalankan validasi tag.
func DecodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

	decoder := json.NewDecoder(r.Body)
	// Field asing ditolak supaya salah ketik nama field terdeteksi saat
	// pengembangan, bukan diam-diam terabaikan.
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return decodeError(err)
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return domain.Validationf("body hanya boleh berisi satu objek JSON")
	}

	return validate.Struct(dst)
}

func decodeError(err error) error {
	var syntaxErr *json.SyntaxError
	var typeErr *json.UnmarshalTypeError
	var maxBytesErr *http.MaxBytesError

	switch {
	case errors.As(err, &syntaxErr):
		return domain.Validationf("format JSON tidak valid pada posisi %d", syntaxErr.Offset)

	case errors.As(err, &typeErr):
		if typeErr.Field != "" {
			return domain.Validation("tipe data tidak sesuai", map[string]string{
				typeErr.Field: fmt.Sprintf("harus bertipe %s", typeErr.Type.String()),
			})
		}
		return domain.Validationf("tipe data pada body tidak sesuai")

	case errors.As(err, &maxBytesErr):
		return domain.Validationf("ukuran body melebihi batas %d byte", maxBytesErr.Limit)

	case errors.Is(err, io.EOF):
		return domain.Validationf("body request kosong")

	case strings.HasPrefix(err.Error(), "json: unknown field "):
		field := strings.Trim(strings.TrimPrefix(err.Error(), "json: unknown field "), `"`)
		return domain.Validation("ada field yang tidak dikenal", map[string]string{
			field: "field ini tidak dikenali",
		})

	default:
		return domain.Validationf("body request tidak bisa dibaca")
	}
}

// UUIDParam membaca parameter path bertipe UUID.
func UUIDParam(r *http.Request, name string) (uuid.UUID, error) {
	raw := chi.URLParam(r, name)
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, domain.Validationf("parameter %s bukan UUID yang valid", name)
	}
	return id, nil
}

// UUIDQuery membaca query string bertipe UUID opsional.
func UUIDQuery(r *http.Request, name string) (*uuid.UUID, error) {
	raw := strings.TrimSpace(r.URL.Query().Get(name))
	if raw == "" {
		return nil, nil
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return nil, domain.Validationf("parameter %s bukan UUID yang valid", name)
	}
	return &id, nil
}

// BoolQuery membaca flag opsional seperti ?active_only=true.
func BoolQuery(r *http.Request, name string) bool {
	value, err := strconv.ParseBool(strings.TrimSpace(r.URL.Query().Get(name)))
	return err == nil && value
}

func IntQuery(r *http.Request, name string, fallback int) int {
	value, err := strconv.Atoi(strings.TrimSpace(r.URL.Query().Get(name)))
	if err != nil {
		return fallback
	}
	return value
}

// DateQuery membaca tanggal opsional berformat YYYY-MM-DD.
func DateQuery(r *http.Request, name string) (*time.Time, error) {
	raw := strings.TrimSpace(r.URL.Query().Get(name))
	if raw == "" {
		return nil, nil
	}
	parsed, err := time.Parse(dateLayout, raw)
	if err != nil {
		return nil, domain.Validationf("parameter %s harus berformat YYYY-MM-DD", name)
	}
	return &parsed, nil
}

const dateLayout = "2006-01-02"

// Date adalah tanggal tanpa jam yang dikirim sebagai "YYYY-MM-DD" pada JSON.
// Tipe khusus ini dipakai supaya form tanggal di frontend tidak perlu mengurus
// zona waktu, yang jadi sumber bug klasik pada tanggal berangkat/pulang trip.
type Date struct {
	time.Time
}

func (d *Date) UnmarshalJSON(data []byte) error {
	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("tanggal harus berupa string berformat YYYY-MM-DD")
	}
	if raw == "" {
		d.Time = time.Time{}
		return nil
	}

	parsed, err := time.Parse(dateLayout, raw)
	if err != nil {
		// Terima juga timestamp penuh, karena beberapa date picker mengirimkannya.
		parsed, err = time.Parse(time.RFC3339, raw)
		if err != nil {
			return fmt.Errorf("tanggal %q harus berformat YYYY-MM-DD", raw)
		}
	}

	d.Time = time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, time.UTC)
	return nil
}

func (d Date) MarshalJSON() ([]byte, error) {
	if d.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(d.Format(dateLayout))
}

// OrNow mengembalikan tanggal hari ini kalau field tanggal dikosongkan.
func (d Date) OrNow() time.Time {
	if d.IsZero() {
		return time.Now()
	}
	return d.Time
}

// Ptr mengembalikan pointer waktu, atau nil kalau tanggalnya kosong.
func (d Date) Ptr() *time.Time {
	if d.IsZero() {
		return nil
	}
	t := d.Time
	return &t
}
