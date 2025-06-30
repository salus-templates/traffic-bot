package main

import (
	"io"
	"log"
	"math/rand/v2"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// callEndpoint makes an HTTP GET request to the given URL and logs the response.
// It uses a WaitGroup to signal completion.
func callEndpoint(url string, wg *sync.WaitGroup) {
	// Ensure wg.Done() is called when the goroutine finishes, even if errors occur.
	defer wg.Done()

	log.Printf("Calling endpoint: %s\n", url)

	now := time.Now()

	// Make the HTTP GET request.
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Error calling %s: %v\n", url, err)
		return
	}
	// Ensure the response body is closed to prevent resource leaks.
	defer resp.Body.Close()

	duration := time.Since(now)

	// Read the response body.
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response from %s: %v\n", url, err)
		return
	}

	// Log the status code and the response body.
	log.Printf("Response from %s - Status: %s, Body size: %d Bytes, Duration: %v\n", url, resp.Status, len(body), duration)
}

const defaultIntervalSeconds = 30

func main() {
	// --- Configuration via Environment Variables ---

	// Get the interval from environment variable. Default to 5 seconds if not set or invalid.
	intervalSecondsStr := os.Getenv("INTERVAL_SECONDS")
	intervalSeconds, err := strconv.Atoi(intervalSecondsStr)
	if err != nil || intervalSeconds <= 0 {
		log.Printf("Invalid or missing INTERVAL_SECONDS environment variable. Defaulting to %d seconds. Error: %v\n", defaultIntervalSeconds, err)
		intervalSeconds = defaultIntervalSeconds
	}
	interval := time.Duration(intervalSeconds) * time.Second

	// Get the endpoints from environment variable. Default to example URLs if not set.
	endpointsStr, ok := os.LookupEnv("ENDPOINTS")
	if !ok {
		log.Println("no endpoints configured")
		log.Println("set the ENDPOINTS env var")
		os.Exit(1)
	}

	var endpoints []string

	// Split the comma-separated string into a slice of URLs.
	for _, ep := range strings.Split(endpointsStr, ",") {
		// Trim whitespace from each endpoint.
		ep = strings.TrimSpace(ep)

		// insert scheme if missing
		if !strings.HasPrefix(ep, "http://") || !strings.HasPrefix(ep, "https://") {
			ep = "http://" + ep
		}

		endpoints = append(endpoints, ep)
	}

	log.Printf("Configured Interval: %s\n", interval)
	log.Printf("Configured Endpoints: %v\n", endpoints)

	// handle health endpoint
	go handleHealth()

	// --- Main application loop ---
	for {
		log.Println("--- Starting new round of API calls ---")

		var wg sync.WaitGroup // Declare a WaitGroup for this round of calls.

		// Iterate over the configured endpoints.
		for _, endpoint := range endpoints {
			// Increment the WaitGroup counter for each goroutine launched.
			wg.Add(1)
			// Launch a new goroutine for each API call.
			go callEndpoint(endpoint, &wg)
		}

		// Wait for all goroutines in this round to complete.
		wg.Wait()

		log.Println("--- All API calls for this round completed ---")

		// Wait for the configured interval before the next round of calls.
		currentInterval := time.Duration(rand.Int64N(int64(interval)))

		log.Printf("--- Waiting for a randomized interval of %v ---\n", currentInterval)

		time.Sleep(currentInterval)
	}
}

func handleHealth() {
	http.ListenAndServe(":8080",
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Healthy"))
		}))
}
