package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/ipoool/jastipin/backend/internal/domain"
	"github.com/ipoool/jastipin/backend/internal/http/response"
)

// maxUploadBytes membatasi ukuran unggahan. 8 MB cukup untuk foto struk dan
// tangkapan layar bukti transfer dari ponsel.
const maxUploadBytes = 8 << 20

// allowedUploadTypes membatasi jenis berkas yang boleh masuk. Daftar putih
// dipakai (bukan daftar hitam) supaya jenis berkas berbahaya tidak lolos hanya
// karena belum terpikirkan.
var allowedUploadTypes = map[string]string{
	"image/jpeg":      ".jpg",
	"image/png":       ".png",
	"image/webp":      ".webp",
	"application/pdf": ".pdf",
}

type UploadHandler struct {
	uploadDir string
	baseURL   string
}

func NewUploadHandler(uploadDir, baseURL string) *UploadHandler {
	return &UploadHandler{uploadDir: uploadDir, baseURL: strings.TrimSuffix(baseURL, "/")}
}

// Upload menerima satu berkas dan mengembalikan URL yang bisa disimpan pada
// kolom proof_url / receipt_url / image_url.
func (h *UploadHandler) Upload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes)

	if err := r.ParseMultipartForm(maxUploadBytes); err != nil {
		response.Error(w, r, domain.Validationf("berkas terlalu besar, maksimal %d MB", maxUploadBytes>>20))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		response.Error(w, r, domain.Validation("berkas tidak ditemukan", map[string]string{
			"file": "kirim berkas pada field bernama file",
		}))
		return
	}
	defer file.Close()

	// Tipe konten dideteksi dari isi berkas, bukan dari header yang dikirim
	// klien, supaya nama atau header yang dipalsukan tidak menentukan hasilnya.
	sniff := make([]byte, 512)
	n, _ := io.ReadFull(file, sniff)
	contentType := http.DetectContentType(sniff[:n])
	if idx := strings.IndexByte(contentType, ';'); idx >= 0 {
		contentType = strings.TrimSpace(contentType[:idx])
	}

	ext, allowed := allowedUploadTypes[contentType]
	if !allowed {
		response.Error(w, r, domain.Validation("jenis berkas tidak didukung", map[string]string{
			"file": "hanya JPG, PNG, WEBP, atau PDF yang bisa diunggah",
		}))
		return
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		response.Error(w, r, domain.Internal(err))
		return
	}

	// Berkas dikelompokkan per bulan supaya satu direktori tidak menampung
	// puluhan ribu berkas setelah beberapa tahun beroperasi.
	subdir := time.Now().Format("2006-01")
	targetDir := filepath.Join(h.uploadDir, subdir)
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		response.Error(w, r, domain.Internal(err))
		return
	}

	// Nama berkas dibuat ulang dari UUID: nama asli dari klien tidak pernah
	// dipakai sebagai path, sehingga tidak ada celah path traversal.
	filename := uuid.NewString() + ext
	targetPath := filepath.Join(targetDir, filename)

	destination, err := os.Create(targetPath)
	if err != nil {
		response.Error(w, r, domain.Internal(err))
		return
	}
	defer destination.Close()

	if _, err := io.Copy(destination, file); err != nil {
		_ = os.Remove(targetPath)
		response.Error(w, r, domain.Internal(err))
		return
	}

	response.Created(w, map[string]any{
		"url":           fmt.Sprintf("%s/uploads/%s/%s", h.baseURL, subdir, filename),
		"path":          filepath.Join(subdir, filename),
		"content_type":  contentType,
		"size":          header.Size,
		"original_name": filepath.Base(header.Filename),
	})
}
