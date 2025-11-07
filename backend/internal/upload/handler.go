package upload

import (
	"encoding/json"
	"net/http"

	"github.com/rishabh-sonic/orbit/internal/middleware"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// POST /api/upload
func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		middleware.BadRequest(w, "failed to parse form")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		middleware.BadRequest(w, "file field required")
		return
	}
	defer file.Close()

	url, err := h.svc.UploadFile(r.Context(), file, header)
	if err != nil {
		middleware.BadRequest(w, err.Error())
		return
	}
	middleware.Ok(w, map[string]string{"url": url})
}

// POST /api/upload/url
func (h *Handler) UploadFromURL(w http.ResponseWriter, r *http.Request) {
	var body struct {
		URL string `json:"url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.URL == "" {
		middleware.BadRequest(w, "url required")
		return
	}
	url, err := h.svc.UploadFromURL(r.Context(), body.URL)
	if err != nil {
		middleware.BadRequest(w, err.Error())
		return
	}
	middleware.Ok(w, map[string]string{"url": url})
}

// GET /api/upload/presign?filename=photo.jpg
func (h *Handler) Presign(w http.ResponseWriter, r *http.Request) {
	filename := r.URL.Query().Get("filename")
	if filename == "" {
		middleware.BadRequest(w, "filename required")
		return
	}
	presignedURL, publicURL, err := h.svc.PresignedURL(r.Context(), filename)
	if err != nil {
		middleware.BadRequest(w, err.Error())
		return
	}
	middleware.Ok(w, map[string]string{
		"upload_url": presignedURL,
		"public_url": publicURL,
	})
}
