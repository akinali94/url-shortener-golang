package main

import (
	"encoding/json"
	"log"
	"net/http"
)


type GeneratorHandler struct{
	l *log.Logger
	sf *Snowflake
}

func NewGeneratorHandler(logging *log.Logger, snowflake *Snowflake) *GeneratorHandler{
	return &GeneratorHandler{
		l: logging,
		sf: snowflake,
	}
}

func (gh *GeneratorHandler) GenerateId(w http.ResponseWriter, r *http.Request){

	if r.Method != http.MethodGet{
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	

	id, err := gh.sf.NextID()
	if err != nil{
		gh.l.Println(err)
	}

	resp := Response{
		ID: id,
	}

	w.Header().Add("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}