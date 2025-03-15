package idgenerator

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

var server *http.Server

func Start() error {

	datacenterID, err := GetDatacenterID()
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}

	machineID, err := GetMachineID()
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}

	snowflake, err := NewSnowflake(datacenterID, machineID)
	if err != nil {
		fmt.Println("Error:", err)
		return err
	}

	l := log.New(os.Stdout, "id-generator-api", log.LstdFlags)

	gh := NewHandler(l, snowflake)

	mux := http.NewServeMux()

	mux.HandleFunc("/getid", gh.GenerateId)

	server = &http.Server{
		Addr:         ":8081",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Println("ID Generator listening on :8081")
	return server.ListenAndServe()
}

// Shutdown gracefully stops the URL Shortener service
func Shutdown(ctx context.Context) error {
	if server == nil {
		return nil
	}
	return server.Shutdown(ctx)
}
