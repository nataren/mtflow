package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/wm/go-flowdock/flowdock"
)

const (
	flowdockAPITokenEnvVar = "FLOWDOCK_API_TOKEN"
	prsAPIKeyEnvVar        = "PRS_API_KEY"
)

func writeMessage(flowID string, client *flowdock.Client) func(msg string) {
	return func(msg string) {
		_, _, err := client.Messages.Create(&flowdock.MessagesCreateOptions{
			FlowID:  flowID,
			Content: msg,
			Event:   "message",
		})
		if err != nil {
			log.Println(err)
		}
	}
}

func executeCommand(
	commandChannel chan<- Command,
	resultChannel chan<- string,
	user string,
	msg json.RawMessage) {
	commandStr := strings.Trim(string(msg[:]), "\"")
	command, err := ParseCommand(commandStr)
	log.Printf("The received command: %s", commandStr)
	if err != nil {
		log.Printf("Error parsing command: %v", err.Error())
		return
	}
	prefix := "@" + user
	containsUser := false
	for _, mention := range command.Mentions {
		if mention == prefix {
			containsUser = true
			break
		}
	}
	if !containsUser {
		log.Printf("The command does not have the mention '%s', instead it has mentions '%+v', will skip it\n", prefix, command.Mentions)
		return
	}
	if command.Type == COMMAND_NONE || command.Target == COMMAND_TARGET_NONE {
		log.Println("Unknown command: ", commandStr)

		//TODO(yurig): this should probably be the help menu
		resultChannel <- "huh? I don't know this command."
		return
	}
	commandChannel <- *command
}

var (

	// Environment variables definitions
	accessToken = os.Getenv(flowdockAPITokenEnvVar)
	prsAPIKey   = os.Getenv(prsAPIKeyEnvVar)

	// Command-line arguments definition
	organization  = flag.String("organization", "", "The organization name")
	flow          = flag.String("flow", "", "The flow to stream from")
	user          = flag.String("user", "", "The name of the user which commands are being directed to")
	prsURL        = flag.String("prsurl", "", "The URL where we can talk to the PullRequestService")
	prsConfigFile = flag.String("prsconfigfile", "", "Path to the configuration file for PullRequestService")
)

func main() {
	flag.Parse()

	// Validation
	if accessToken == "" {
		log.Fatalf("The '%s' environment variable is required", flowdockAPITokenEnvVar)
	}
	if prsAPIKey == "" {
		log.Fatalf("The '%s' environment variable is required", prsAPIKeyEnvVar)
	}
	if *organization == "" {
		log.Fatal("'organization' is a required parameter")
	}
	if *flow == "" {
		log.Fatal("'flow' is a required parameter")
	}
	if *user == "" {
		log.Fatal("'user' is a required parameter")
	}
	if *prsURL == "" {
		log.Fatal("'prsurl' is a required parameter")
	}
	if *prsConfigFile == "" {
		log.Fatal("'prsconfigfile' is a required parameter")
	}
	prsConfig, err := ioutil.ReadFile(*prsConfigFile)
	if err != nil {
		log.Fatal(err)
	}
	prsParsedURL, err := url.Parse(*prsURL)
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
	flowID := ""
	for _, f := range flows {
		if strings.ToLower(*f.ParameterizedName) == strings.ToLower(*flow) {
			flowID = *f.Id
			break
		}
	}
	if flowID == "" {
		log.Fatalf("Could not find the flow '%s' which you requested to listen from", *flow)
	}

	// Build the streaming HTTP request to flowdock
	log.Printf("I will stream from: organization='%s' flow='%s' user='%s' prsURL='%s' prsconfigfile='%s'", *organization, *flow, *user, *prsURL, *prsConfigFile)

	// Build the event source
	messages, _, err := flowdockClient.Messages.Stream(accessToken, *organization, *flow)
	if err != nil {
		log.Fatal(err)
	}
	httpClient := &http.Client{
		Timeout: time.Duration(5 * time.Second),
	}

	// Kick off the result handler
	write := writeMessage(flowID, flowdockClient)
	resultChannel := make(chan string)
	go func() {
		for {
			write(<-resultChannel)
		}
	}()

	// Kick off the command handler
	commandChannel := make(chan Command)
	InitCommandHandler(prsParsedURL, &prsConfig, prsAPIKey, httpClient)
	go RunCommandHandler(commandChannel, resultChannel)

	// When we get a new message fire off the handler
	for {
		message := <-messages
		if message.RawContent == nil {
			continue
		}
		executeCommand(commandChannel, resultChannel, *user, *message.RawContent)
	}
}
