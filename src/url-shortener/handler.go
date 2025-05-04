package urlshortener

import (
	"encoding/json"
	"fmt"
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
	fmt.Print("LongUrl is: ", longUrl)
	redir := "https://" + longUrl

	http.Redirect(w, r, redir, http.StatusFound)
}

func (h *Handler) shortenUrlHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var requestBody struct {
		URL string `json:"url"`
	}

	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		http.Error(w, "Invalid JSON request body", http.StatusBadRequest)
		return
	}

	longUrl := strings.TrimSpace(requestBody.URL)
	longUrl = strings.Trim(longUrl, "\"'")

	if longUrl == "" {
		http.Error(w, "URL cannot be empty", http.StatusBadRequest)
		return
	}

	shortUrl, err := h.service.generateShortUrl(longUrl)
	if err != nil {
		http.Error(w, "error on generate Short Url", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated) // Set header before writing response

	response := map[string]string{"shortUrl": "http://localhost:8080/" + shortUrl}
	json.NewEncoder(w).Encode(response)

	//w.WriteHeader(http.StatusCreated) --> This causing an error: "http: superfluous response.WriteHeader". json.NewEncoder... line is already return with 200 code.
}
