package main

import (
	"encoding/json"
	"net/http"
	"strings"
)



 
func redirectUrlHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet{
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	shortUrl := strings.TrimPrefix(r.URL.Path, "/")
	if shortCode == ""{
		http.Error(w, "Short code required", http.StatusBadRequest)
		return
	}


	longUrl, exists := getItemsFromDB() //fetch data
	if !exists{
		http.NotFound(w, r)
		return
	}

	http.Redirect(w, r, longUrl, http.StatusFound)
}


func shortenUrlHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost{
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	var req URLMapping
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil{
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	shortUrl := generateShortUrl(req.LongUrl)

	saveUrlMapping(shortCode, req.LongUrl)

	response := map[string]string{"shortUrl": "http://localhost:8080/" + shortUrl}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}
