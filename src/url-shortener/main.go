package urlshortener

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/akinali94/url-shortener-golang/pkg/ratelimiter"
	"github.com/akinali94/url-shortener-golang/pkg/repository"
	"github.com/redis/go-redis/v9"
)

var server *http.Server

func Start() error {

	baseDomain := os.Getenv("BASE_DOMAIN")
	mongoURI := os.Getenv("MONGO_URI")
	mongoDB := os.Getenv("MONGO_DB")
	mongoCollection := os.Getenv("MONGO_COLLECTION")
	redisAddr := os.Getenv("REDIS_ADDR")
	redisPassword := os.Getenv("REDIS_PASSWORD")
	idGeneratorDomain := os.Getenv("ID_GENERATOR_DOMAIN")

	if mongoURI == "" || mongoDB == "" || mongoCollection == "" || redisAddr == "" || redisPassword == "" {
		return errors.New("environment variables are empty")
	}

	mdb, err := repository.NewMongoDB(mongoURI, mongoDB, mongoCollection)
	if err != nil {
		fmt.Println("Cannot Connect to Database, err:" + err.Error())
		return err
	}
	defer func() {
		mdb.Close(context.TODO())
	}()

	repo := repository.NewRepository[URLMapping](mdb.Collection)

	service := NewService(repo, idGeneratorDomain)

	handler := NewHandler(service, baseDomain)

	//Rate Limiting Settings
	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       0,
	})

	defer func() {
		redisClient.Close()
	}()

	shortenRateLimiterConfig := &ratelimiter.RateLimiterConfig{
		Extractor:   ratelimiter.NewHTTPHeadersExtractor(),
		Strategy:    ratelimiter.NewSortedSetCounterStrategy(redisClient, func() time.Time { return time.Now() }),
		Expiration:  1 * time.Minute,
		MaxRequests: 20,
	}

	redirectRateLimiterConfig := &ratelimiter.RateLimiterConfig{
		Extractor:   ratelimiter.NewHTTPHeadersExtractor(),
		Strategy:    ratelimiter.NewSortedSetCounterStrategy(redisClient, func() time.Time { return time.Now() }),
		Expiration:  1 * time.Minute,
		MaxRequests: 75,
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
