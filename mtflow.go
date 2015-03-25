package main

import (
	"flag"
	"fmt"
	"github.com/bernerdschaefer/eventsource"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	FLOWDOCK_API_TOKEN = "FLOWDOCK_API_TOKEN"
)

func main() {

	// Environment variables definitions
	accessToken := os.Getenv(FLOWDOCK_API_TOKEN)

	// Command-line arguments definition
	var organization string
	flag.StringVar(&organization, "organization", "", "The organization name")
	var flow string
	flag.StringVar(&flow, "flow", "", "The flow to stream from")
	flag.Parse()

	// Validation
	if accessToken == "" {
		log.Fatalf("'%s' environment variable not found", FLOWDOCK_API_TOKEN)
	}
	if organization == "" {
		log.Fatal("'organization' is a required parameter")
	}
	if flow == "" {
		log.Fatal("'flow' is a required parameter")
	}

	// Build the HTTP request
	streamURL := fmt.Sprintf("https://%s@stream.flowdock.com/flows/%s/%s?user=1", accessToken, organization, flow)
	log.Printf("Will stream from organization '%s' flow '%s'", organization, flow)
	request, err := http.NewRequest("GET", streamURL, nil)
	if err != nil {
		log.Fatal(err)
	}
	request.Header = map[string][]string{
		"Content-Type": {"text/event-stream"},
	}

	// Build the event source
	source := eventsource.New(request, 3*time.Second)
	for {
		event, err := source.Read()
		if err != nil {
			log.Println(err)
			time.Sleep(5 * time.Second)
		}
		log.Printf("event.ID '%s', event.Type '%s', event.Data '%s'\n", event.ID, event.Type, event.Data)
	}
}
