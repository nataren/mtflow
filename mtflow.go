package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/bernerdschaefer/eventsource"
	"io/ioutil"
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

func executeCommand(user string, msg FlowdockMessage, client *http.Client, prsURL *url.URL, prsApiKey string, prsConfig []byte) {

	// Catch the panicking go routine
	defer func() {
		if err := recover(); err != nil {
			log.Printf("An error ocurred: %s", err)
		}
	}()
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
			log.Println("I will start processing of pull requests")
			startService := &http.Request{}
			startService.Method = "POST"
			q := prsURL.Query()
			q.Set("apikey", prsApiKey)
			startService.URL = &url.URL{
				Host:     prsURL.Host,
				Scheme:   prsURL.Scheme,
				Opaque:   "/host/services",
				RawQuery: q.Encode(),
			}
			startService.Header = map[string][]string{
				"Content-Type": {"application/xml"},
			}
			startService.Body = ioutil.NopCloser(bytes.NewReader(prsConfig))
			startService.ContentLength = int64(len(prsConfig))
			resp, err := client.Do(startService)
			if err != nil {
				log.Panic(err)
			}
			defer resp.Body.Close()
			statusCode := resp.StatusCode
			if statusCode >= 200 && statusCode < 300 {
				log.Println("Successfully started processing pull requests")
			} else {
				log.Printf("Failed to start processing pull requests: %v\n", resp.Status)
			}
		}

	case "stop":
		switch modifier {
		case "pr":
			log.Println("I will handle 'stop pr' command")
			stopService := &http.Request{}
			stopService.Method = "DELETE"
			q := prsURL.Query()
			q.Set("apikey", prsApiKey)
			prsURL.RawQuery = q.Encode()
			stopService.URL = prsURL
			resp, err := client.Do(stopService)
			if err != nil {
				log.Panic(err)
			}
			defer resp.Body.Close()
			statusCode := resp.StatusCode
			if statusCode >= 200 && statusCode < 300 {
				log.Println("Successfully stopped processing pull requests")
			} else {
				log.Printf("Failed to stop processing pull requests: %v\n", resp.Status)
			}
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
	var prsConfigFile string
	flag.StringVar(&prsConfigFile, "prsconfigfile", "", "Path to the configuration file for PullRequestService")
	flag.Parse()

	// Validation
	if accessToken == "" {
		log.Fatalf("The '%s' environment variable is required", FLOWDOCK_API_TOKEN)
	}
	if prsApiKey == "" {
		log.Fatalf("The '%s' environment variable is required", PRS_API_KEY)
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
		log.Fatal("'prsurl' is a required parameter")
	}
	if prsConfigFile == "" {
		log.Fatal("'prsconfigfile' is a required parameter")
	}
	prsConfig, err := ioutil.ReadFile(prsConfigFile)
	if err != nil {
		log.Fatal(err)
	}
	prsParsedURL, err := url.Parse(prsURL)
	if err != nil {
		log.Fatal(err)
	}

	// Build the HTTP request to flowdock
	streamURL := fmt.Sprintf("https://%s@stream.flowdock.com/flows/%s/%s?user=1", accessToken, organization, flow)
	log.Printf("I will stream from: organization='%s' flow='%s' user='%s' prsURL='%s' prsconfigfile='%s'", organization, flow, user, prsURL, prsConfigFile)
	flowdock, err := http.NewRequest("GET", streamURL, nil)
	if err != nil {
		log.Fatal(err)
	}
	flowdock.Header = map[string][]string{
		"Content-Type": {"text/event-stream"},
	}

	// The shared http client
	client := &http.Client{
		Timeout: time.Duration(5 * time.Second),
	}

	// Build the event source
	source := eventsource.New(flowdock, 3*time.Second)
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
		go executeCommand(user, msg, client, prsParsedURL, prsApiKey, prsConfig)
	}
}
