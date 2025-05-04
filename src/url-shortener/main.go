package urlshortener

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/akinali94/url-shortener-golang/pkg/ratelimiter"
	"github.com/akinali94/url-shortener-golang/pkg/repository"
	"github.com/redis/go-redis/v9"
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

	// Initialize the rate limiter
	limiter, err := ratelimiter.NewRateLimiter(ratelimiter.Options{
		Redis: &redis.Options{
			Addr: "localhost:6379", // Change to your Redis address
		},
		KeyPrefix: "urlshortener:",
		DefaultRate: ratelimiter.Rate{
			Limit:  100,         // Allow 100 requests
			Window: time.Minute, // per minute
		},
	})
	if err != nil {
		fmt.Println("Cannot initialize rate limiter, err:" + err.Error())
		return err
	}
	defer limiter.Close()

	mux := http.NewServeMux()

	// Create separate handlers with different rate limits
	shortenHandler := limiter.Middleware(
		ratelimiter.IPKeyFunc(true),
		ratelimiter.Rate{Limit: 10, Window: time.Minute}, //Limit is higher for url creation
	)(http.HandlerFunc(handler.shortenUrlHandler))

	redirectHandler := limiter.Middleware(
		ratelimiter.IPKeyFunc(true),
		ratelimiter.Rate{Limit: 200, Window: time.Minute}, //Lower limit for redirect
	)(http.HandlerFunc(handler.redirectUrlHandler))

	mux.HandleFunc("/shorten", shortenHandler)
	mux.HandleFunc("/", redirectHandler)

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
