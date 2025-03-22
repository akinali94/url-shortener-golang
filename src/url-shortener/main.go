package urlshortener

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/akinali94/url-shortener-golang/pkg/repository"
)

var server *http.Server

func Start() error {

	mdb, err := repository.NewMongoDB("mongodb://localhost:27017", "denemedb", "urlshort")
	if err != nil {
		fmt.Println("Cannot Connect to Database, err:" + err.Error())
		return err
	}
	defer func() {
		fmt.Println("DEFER CALISTI")
		mdb.Close(context.TODO())
	}()

	repo := repository.NewRepository[URLMapping](mdb.Collection)

	service := NewService(repo)

	handler := NewHandler(service)

	mux := http.NewServeMux()

	mux.HandleFunc("/shorten", handler.shortenUrlHandler)
	mux.HandleFunc("/", handler.redirectUrlHandler)

	server = &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Println("URL Shortener listening on :8080")
	return server.ListenAndServe()
}

func Shutdown(ctx context.Context) error {
	if server == nil {
		return nil
	}
	return server.Shutdown(ctx)
}
