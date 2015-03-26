package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/bernerdschaefer/eventsource"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	FLOWDOCK_API_TOKEN = "FLOWDOCK_API_TOKEN"
	PRS_API_KEY        = "PRS_API_KEY"
)

type FlowdockMessage struct {
	App         string    `json:"app,omitempty"`
	Attachments []string  `json:"attachments"`
	Content     string    `json:"content"`
	CreatedAt   time.Time `json:"created_at"`
	Event       string    `json:"event"`
	Flow        string    `json:"flow"`
	Id          uint32    `json:"id"`
	Persist     bool      `json:"persist"`
	Sent        uint64    `json:"sent"`
	Tags        []string  `json:"tags"`
	User        string    `json:"user"`
	Uuid        string    `json:"uuid,omitempty"`
}

func execute(user string, msg FlowdockMessage, client *http.Client, prsURL *url.URL, prsApiKey string) {
	message := strings.ToLower(msg.Content)
	prefix := "@" + user
	if !strings.HasPrefix(message, strings.ToLower(prefix)) {
		return
	}
	parts := strings.Split(message, " ")
	if len(parts) < 3 {
		log.Println("Incorrect cmd format: %s", message)
		return
	}
	cmd := parts[1]
	modifier := parts[2]
	switch cmd {
	case "start":
		switch modifier {
		case "pr":
			go func() {
				log.Println("I will start processing of pull requests")
			}()
		}

	case "stop":
		switch modifier {
		case "pr":
			go func() {
				log.Println("I will handle 'stop pr' command")
				stopService := &http.Request{}
				stopService.Method = "DELETE"
				q := prsURL.Query()
				q.Set("apikey", prsApiKey)
				prsURL.RawQuery = q.Encode()
				stopService.URL = prsURL
				resp, err := client.Do(stopService)
				if err != nil {
					log.Fatal(err)
				}
				statusCode := resp.StatusCode
				if statusCode >= 200 && statusCode < 300 {
					log.Println("Successfully stopped processing pull requests")
				} else {
					log.Printf("Failed to stop processing pull requests: %v\n", resp.Status)
				}
			}()
		}
	}
}

func main() {

	// Environment variables definitions
	accessToken := os.Getenv(FLOWDOCK_API_TOKEN)
	prsApiKey := os.Getenv(PRS_API_KEY)

	// Command-line arguments definition
	var organization string
	flag.StringVar(&organization, "organization", "", "The organization name")
	var flow string
	flag.StringVar(&flow, "flow", "", "The flow to stream from")
	var user string
	flag.StringVar(&user, "user", "", "The name of the user which commands are being directed to")
	var prsURL string
	flag.StringVar(&prsURL, "prsurl", "", "The URL where we can talk to the PullRequestService")
	flag.Parse()

	// Validation
	if accessToken == "" {
		log.Fatalf("'%s' environment variable not found", FLOWDOCK_API_TOKEN)
	}
	if prsApiKey == "" {
		log.Fatal("The environment variable 'PRS_API_KEY' is required to be set")
	}
	if organization == "" {
		log.Fatal("'organization' is a required parameter")
	}
	if flow == "" {
		log.Fatal("'flow' is a required parameter")
	}
	if user == "" {
		log.Fatal("'user' is a required parameter")
	}
	if prsURL == "" {
		log.Fatal("The 'pullrequestservice' URL is required")
	}
	prsParsedURL, err := url.Parse(prsURL)
	if err != nil {
		log.Fatal(err)
	}

	// Build the HTTP request
	streamURL := fmt.Sprintf("https://%s@stream.flowdock.com/flows/%s/%s?user=1", accessToken, organization, flow)
	log.Printf("I will stream from organization '%s', flow '%s', user '%s', prsURL '%s'", organization, flow, user, prsURL)
	request, err := http.NewRequest("GET", streamURL, nil)
	if err != nil {
		log.Fatal(err)
	}
	request.Header = map[string][]string{
		"Content-Type": {"text/event-stream"},
	}

	// The shared http client
	client := &http.Client{}
	client.Timeout = 5 * time.Second

	// Build the event source
	source := eventsource.New(request, 3*time.Second)
	for {
		event, err := source.Read()
		if err != nil {
			log.Println("Error parsing unsupported Flowdock event")
			continue
		}

		// Interpret the commands
		var msg FlowdockMessage
		unmarshalErr := json.Unmarshal(event.Data, &msg)
		if unmarshalErr != nil {
			continue
		}
		execute(user, msg, client, prsParsedURL, prsApiKey)
	}
}
