package idgenerator

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

type Handler struct {
	l  *log.Logger
	sf *Snowflake
}

func NewHandler(logging *log.Logger, snowflake *Snowflake) *Handler {
	return &Handler{
		l:  logging,
		sf: snowflake,
	}
}

func (gh *Handler) GenerateId(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	id, err := gh.sf.NextID()
	if err != nil {
		gh.l.Println(err)
	}

	resp := Response{
		ID: id,
	}
	fmt.Println("Id created on Generator: ", resp)

	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
