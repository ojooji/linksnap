package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/ojooji/linksnap/internal/repository"
)

type Handler struct {
	repo    repository.Repository
	baseURL string
}

func New(repo repository.Repository, baseURL string) *Handler {
	return &Handler{repo: repo, baseURL: strings.TrimRight(baseURL, "/")}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/shorten", h.shorten)
	mux.HandleFunc("DELETE /api/{code}", h.delete)
	mux.HandleFunc("GET /{code}", h.redirect)
}

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	Code     string `json:"code"`
	ShortURL string `json:"short_url"`
}

func (h *Handler) shorten(w http.ResponseWriter, r *http.Request) {
	var req shortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	target := strings.TrimSpace(req.URL)
	parsed, err := url.Parse(target)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		http.Error(w, "url must be http or https", http.StatusBadRequest)
		return
	}

	code, err := h.repo.CreateURL(r.Context(), parsed.String())
	if err != nil {
		http.Error(w, "could not shorten url", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, shortenResponse{
		Code:     code,
		ShortURL: h.baseURL + "/" + code,
	})
}

func (h *Handler) redirect(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	original, err := h.repo.GetURL(r.Context(), code)
	if errors.Is(err, repository.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, "lookup failed", http.StatusInternalServerError)
		return
	}

	go h.recordClick(code, clientIP(r), r.UserAgent())

	http.Redirect(w, r, original, http.StatusFound)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	err := h.repo.DeleteURL(r.Context(), code)
	if errors.Is(err, repository.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		http.Error(w, "delete failed", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) recordClick(code, ip, userAgent string) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = h.repo.RecordClick(ctx, code, ip, userAgent)
}

func clientIP(r *http.Request) string {
	if xf := r.Header.Get("X-Forwarded-For"); xf != "" {
		if i := strings.Index(xf, ","); i >= 0 {
			return strings.TrimSpace(xf[:i])
		}
		return strings.TrimSpace(xf)
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
