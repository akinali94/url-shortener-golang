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
		mdb.Close(context.TODO())
	}()

	repo := repository.NewRepository[URLMapping](mdb.Collection)

	service := NewService(repo)

	handler := NewHandler(service)

	//Rate Limiting Settings
	redisClient := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	defer func() {
		redisClient.Close()
	}()

	shortenRateLimiterConfig := &ratelimiter.RateLimiterConfig{
		Extractor:   ratelimiter.NewHTTPHeadersExtractor(),
		Strategy:    ratelimiter.NewSortedSetCounterStrategy(redisClient, func() time.Time { return time.Now() }),
		Expiration:  1 * time.Minute,
		MaxRequests: 5,
	}

	redirectRateLimiterConfig := &ratelimiter.RateLimiterConfig{
		Extractor:   ratelimiter.NewHTTPHeadersExtractor(),
		Strategy:    ratelimiter.NewSortedSetCounterStrategy(redisClient, func() time.Time { return time.Now() }),
		Expiration:  1 * time.Minute,
		MaxRequests: 3,
	}

	mux := http.NewServeMux()

	shortenHandler := http.HandlerFunc(handler.shortenUrlHandler)
	shortenRateLimitedHandler := ratelimiter.NewHTTPRateLimiterHandler(shortenHandler, shortenRateLimiterConfig)

	redirectHandler := http.HandlerFunc(handler.redirectUrlHandler)
	redirectRateLimitedHandler := ratelimiter.NewHTTPRateLimiterHandler(redirectHandler, redirectRateLimiterConfig)

	mux.Handle("/shorten", shortenRateLimitedHandler)
	mux.Handle("/", redirectRateLimitedHandler)

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
