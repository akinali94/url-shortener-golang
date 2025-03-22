package urlshortener

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Handler struct {
	service *Service
}

func NewHandler(s *Service) *Handler {
	return &Handler{
		service: s,
	}
}

func (h *Handler) mainPageHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, `
		<html>
		<head><title>URL Shortener</title></head>
		<body>
			<h1>URL Shortener</h1>
			<p>Welcome to the URL shortener service!</p>
			<!-- You could add a form here to create new short URLs -->
		</body>
		</html>
	`)
}

func (h *Handler) redirectUrlHandler(w http.ResponseWriter, r *http.Request) {

	path := r.URL.Path
	if path == "/" {
		h.mainPageHandler(w, r)
		return
	}

	shortUrl := strings.TrimPrefix(path, "/")

	longUrl, err := h.service.getLongUrl(shortUrl)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if longUrl == "" {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}
	redir := "https://" + longUrl

	http.Redirect(w, r, redir, http.StatusFound)
}

func (h *Handler) shortenUrlHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var req string
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	req = string(body)

	shortUrl, err := h.service.generateShortUrl(req)
	if err != nil {
		http.Error(w, "error on generate Short Url", http.StatusBadRequest)
		return
	}

	response := map[string]string{"shortUrl": "http://localhost:8080/" + shortUrl}

	json.NewEncoder(w).Encode(response)
	w.WriteHeader(http.StatusCreated)

}
