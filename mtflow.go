package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"github.com/wm/go-flowdock/flowdock"
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

func release(coordinator chan bool) {
	coordinator <- true
}

func writeMessage(flowId string, client *flowdock.Client) func(msg string) {
	return func(msg string) {
		_, _, err := client.Messages.Create(&flowdock.MessagesCreateOptions{
			FlowID:  flowId,
			Content: msg,
			Event:   "message",
		})
		if err != nil {
			log.Println(err)
		}
	}
}

func executeCommand(
	write func(msg string),
	user string,
	msg json.RawMessage,
	client *http.Client,
	prsURL *url.URL,
	prsApiKey string,
	prsConfig []byte,
	coordinator chan bool) {

	// Catch the panicking go routine
	defer func() {
		if err := recover(); err != nil {
			log.Printf("An error ocurred: %s", err)
		}
		release(coordinator)
	}()
	command := strings.Trim(strings.ToLower(string(msg[:])), "\"")
	log.Printf("The received command: %s", command)
	prefix := "@" + user
	if !strings.HasPrefix(command, strings.ToLower(prefix)) {
		log.Printf("The command '%s' does not have the prefix '%s', will skip it\n", command, prefix)
		return
	}
	parts := strings.Split(command, " ")
	if len(parts) < 3 {
		log.Printf("The command '%s' has the incorrect command format", command)
		return
	}
	cmd := parts[1]
	modifier := parts[2]
	switch cmd {
	case "start":
		switch modifier {
		case "pr":
			<-coordinator
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
				msg := "Successfully started processing pull requests"
				log.Println(msg)
				write(msg)
			} else {
				msg := "Failed to start processing pull requests: " + resp.Status
				write(msg)
			}
		default:
			log.Printf("The modifier '%s' is not handled\n", modifier)
		}

	case "stop":
		switch modifier {
		case "pr":
			<-coordinator
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
				msg := "Successfully stopped processing pull requests"
				log.Println(msg)
				write(msg)
			} else {
				msg := "Failed to stop processing pull requests: " + resp.Status
				log.Println(msg)
				write(msg)
			}
		default:
			log.Println("The modifier '%s' is not handled", modifier)
		}
	default:
		log.Println("The command '%s' is not handled", cmd)
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

	// Setup the Flowdock REST client
	flowdockClient := flowdock.NewClientWithToken(&http.Client{}, accessToken)
	flows, _, flowsErr := flowdockClient.Flows.List(true, &flowdock.FlowsListOptions{User: false})
	if flowsErr != nil {
		log.Println(flowsErr)
	}

	// Figure out the flowId from the requested flow
	flowId := ""
	for _, f := range flows {
		if strings.ToLower(*f.Name) == strings.ToLower(flow) {
			flowId = *f.Id
			break
		}
	}
	if flowId == "" {
		log.Fatalf("Could not find the flow '%s' which you requested to listen from", flow)
	}

	// Say hello to the flow
	write := writeMessage(flowId, flowdockClient)
	write("Hello, I am ready to accept commands")

	// Build the streaming HTTP request to flowdock
	log.Printf("I will stream from: organization='%s' flow='%s' user='%s' prsURL='%s' prsconfigfile='%s'", organization, flow, user, prsURL, prsConfigFile)

	coordinator := make(chan bool)
	go release(coordinator)

	// Build the event source
	messages, _, err := flowdockClient.Messages.Stream(accessToken, organization, flow)
	if err != nil {
		log.Fatal(err)
	}
	httpClient := &http.Client{
		Timeout: time.Duration(5 * time.Second),
	}
	for {
		message := <-messages
		go executeCommand(write, user, *message.RawContent, httpClient, prsParsedURL, prsApiKey, prsConfig, coordinator)
	}
}
