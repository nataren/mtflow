package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/bernerdschaefer/eventsource"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	FLOWDOCK_API_TOKEN = "FLOWDOCK_API_TOKEN"
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

func execute(user string, msg FlowdockMessage) {
	message := msg.Content
	prefix := "@" + user
	if !strings.HasPrefix(message, prefix) {
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
				log.Println("Will start processing of pull requests")
			}()
		}

	case "stop":
		switch modifier {
		case "pr":
			go func() {
				log.Println("Will stop processing of pull requests")
			}()
		}
	}
}

func main() {

	// Environment variables definitions
	accessToken := os.Getenv(FLOWDOCK_API_TOKEN)

	// Command-line arguments definition
	var organization string
	flag.StringVar(&organization, "organization", "", "The organization name")
	var flow string
	flag.StringVar(&flow, "flow", "", "The flow to stream from")
	var user string
	flag.StringVar(&user, "user", "", "The name of the user which commands are being directed to")
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
	if user == "" {
		log.Fatal("'user' is a required parameter")
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
			continue
		}
		log.Printf("eventType '%s', data '%s'", event.Type, event.Data)

		// Interpret the commands
		var msg FlowdockMessage
		unmarshalErr := json.Unmarshal(event.Data, &msg)
		if unmarshalErr != nil {
			continue
		}
		execute(user, msg)
	}
}
