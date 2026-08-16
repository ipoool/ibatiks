package handler

import (
	"net/http"

	"github.com/ipoool/jastipin/backend/internal/http/response"
	"github.com/ipoool/jastipin/backend/internal/service"
)

type FXHandler struct {
	fx *service.FXService
}

func NewFXHandler(fx *service.FXService) *FXHandler {
	return &FXHandler{fx: fx}
}

// Rate mengembalikan kurs terkini sebuah mata uang terhadap rupiah, dipakai
// mengisi kolom kurs saat trip baru dibuat.
func (h *FXHandler) Rate(w http.ResponseWriter, r *http.Request) {
	rate, err := h.fx.Rate(r.Context(), r.URL.Query().Get("from"))
	if err != nil {
		response.Error(w, r, err)
		return
	}
	response.OK(w, rate)
}
