// Package pagination menstandarkan parameter halaman untuk semua endpoint list.
package pagination

import (
	"net/http"
	"strconv"
)

const (
	DefaultPerPage = 20
	MaxPerPage     = 200
)

type Params struct {
	Page    int
	PerPage int
	Search  string
	Sort    string
	Order   string // asc | desc
}

func (p Params) Offset() int { return (p.Page - 1) * p.PerPage }
func (p Params) Limit() int  { return p.PerPage }

// FromRequest membaca ?page, ?per_page, ?q, ?sort, ?order dari query string.
// Nilai yang tidak masuk akal dinormalkan, bukan ditolak, supaya URL yang
// diketik manual tidak gampang bikin error.
func FromRequest(r *http.Request) Params {
	q := r.URL.Query()

	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}

	perPage, _ := strconv.Atoi(q.Get("per_page"))
	switch {
	case perPage < 1:
		perPage = DefaultPerPage
	case perPage > MaxPerPage:
		perPage = MaxPerPage
	}

	order := q.Get("order")
	if order != "asc" && order != "desc" {
		order = "desc"
	}

	return Params{
		Page:    page,
		PerPage: perPage,
		Search:  q.Get("q"),
		Sort:    q.Get("sort"),
		Order:   order,
	}
}

// SortColumn memetakan nama sort dari client ke nama kolom yang diizinkan.
// Ini mencegah SQL injection lewat parameter ?sort karena hanya kolom yang
// terdaftar di allowed yang bisa dipakai.
func SortColumn(sort string, allowed map[string]string, fallback string) string {
	if col, ok := allowed[sort]; ok {
		return col
	}
	return fallback
}
