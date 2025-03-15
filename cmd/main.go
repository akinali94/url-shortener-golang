package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	idgenerator "github.com/akinali94/url-shortener-golang/src/id-generator"
	urlshortener "github.com/akinali94/url-shortener-golang/src/url-shortener"
)

func main() {
	serviceFlag := flag.String("service", "", "Service to run: 'urlshortener', 'idgenerator', or leave empty to run specified services")
	allFlag := flag.Bool("all", false, "Run all services")
	flag.Parse()

	if *serviceFlag == "" && !*allFlag {
		fmt.Println("Please specify a service (--service=urlshortener or --service=idgenerator) or use --all to run all services")
		flag.PrintDefaults()
		os.Exit(1)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)

	var wg sync.WaitGroup

	if *serviceFlag == "urlshortener" || *allFlag {
		wg.Add(1)
		go func() {
			defer wg.Done()
			log.Println("Starting URL Shortener service...")
			err := urlshortener.Start()
			if err != nil {
				log.Printf("URL Shortener service failed: %v", err)
			}
		}()
	}

	if *serviceFlag == "idgenerator" || *allFlag {
		wg.Add(1)
		go func() {
			defer wg.Done()
			log.Println("Starting ID Generator service...")
			err := idgenerator.Start()
			if err != nil {
				log.Printf("ID Generator service failed: %v", err)
			}
		}()
	}

	// Wait for interrupt signal
	go func() {
		<-signalChan
		log.Println("\nReceived interrupt signal. Shutting down...")

		// Implement graceful shutdown for your services here
		// For example: urlshortener.Shutdown() and idgenerator.Shutdown()

		// Force exit after a timeout if graceful shutdown takes too long
		go func() {
			<-time.After(2 * time.Second)
			log.Println("Forced shutdown after timeout")
			os.Exit(1)
		}()
	}()

	// Wait for all services to complete
	wg.Wait()
	log.Println("All services have stopped. Exiting.")
	// Wait for interrupt signal
	go func() {
		<-signalChan
		log.Println("\nReceived interrupt signal. Shutting down...")

		// Implement graceful shutdown for your services here
		// For example: urlshortener.Shutdown() and idgenerator.Shutdown()

		// Force exit after a timeout if graceful shutdown takes too long
		go func() {
			<-time.After(2 * time.Second)
			log.Println("Forced shutdown after timeout")
			os.Exit(1)
		}()
	}()

	// Wait for all services to complete
	wg.Wait()
	log.Println("All services have stopped. Exiting.")

}
