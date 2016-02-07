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

func main() {
	const (
		flowdockAPITokenEnvVar = "FLOWDOCK_API_TOKEN"
		prsAPIKeyEnvVar        = "PRS_API_KEY"

		orgFlagName           = "organization"
		flowFlagName          = "flow"
		userFlagName          = "user"
		prsURLFlagName        = "prsurl"
		prsConfigFileFlagName = "prsconfigfile"
	)
	var (

		// Environment variables definitions
		accessToken = os.Getenv(flowdockAPITokenEnvVar)
		prsAPIKey   = os.Getenv(prsAPIKeyEnvVar)

		// Command-line arguments definition
		org           = flag.String(orgFlagName, "", "The organization name (required)")
		flow          = flag.String(flowFlagName, "", "The flow to stream from (required)")
		user          = flag.String(userFlagName, "", "The name of the user which commands are being directed to (required)")
		prsURL        = flag.String(prsURLFlagName, "", "The URL where we can talk to the PullRequestService (required)")
		prsConfigFile = flag.String(prsConfigFileFlagName, "", "Path to the configuration file for PullRequestService (required)")
	)
	flag.Parse()

	// Validation
	assertNonEmptyEnvVar(accessToken, flowdockAPITokenEnvVar)
	assertNonEmptyEnvVar(prsAPIKey, prsAPIKeyEnvVar)
	assertNonEmptyFlag(org, orgFlagName)
	assertNonEmptyFlag(flow, flowFlagName)
	assertNonEmptyFlag(user, userFlagName)
	assertNonEmptyFlag(prsURL, prsURLFlagName)
	assertNonEmptyFlag(prsConfigFile, prsConfigFileFlagName)
	prsConfig, err := ioutil.ReadFile(*prsConfigFile)
	if err != nil {
		log.Fatal(err)
	}
	prsParsedURL, err := url.Parse(*prsURL)
	if err != nil {
		log.Fatal(err)
	}

	// Setup the Flowdock REST client
	flowdockClient := flowdock.NewClientWithToken(&http.Client{Timeout: time.Duration(5 * time.Second)}, accessToken)
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
	log.Printf("I will stream from: organization='%s' flow='%s' user='%s' prsURL='%s' prsconfigfile='%s'", *org, *flow, *user, *prsURL, *prsConfigFile)

	// Build the event source
	messages, _, err := flowdockClient.Messages.Stream(accessToken, *org, *flow)
	if err != nil {
		log.Fatal(err)
	}
	httpClient := &http.Client{
		Timeout: time.Duration(5 * time.Second),
	}

	// Kick off the result handler
	write := writeMessage(flowID, flowdockClient)
	resultChannel := make(chan Result)
	go func() {
		for {
			result := <-resultChannel
			write(result.Message, result.ThreadId)
		}
	}()
	searcher := &Searcher{}

	// Kick off the command handler
	commandChannel := make(chan Command)
	InitCommandHandler(prsParsedURL, &prsConfig, prsAPIKey, httpClient, searcher)
	go RunCommandHandler(commandChannel, resultChannel)

	// When we get a new message fire off the handler
	for {
		message := <-messages
		if message.RawContent == nil {
			continue
		}
		tmp := message.ThreadId
		var threadId string
		if tmp != nil {
			threadId = *tmp
		}
		executeCommand(commandChannel, resultChannel, *user, *message.RawContent, threadId)
	}
}

func executeCommand(
	commandChannel chan<- Command,
	resultChannel chan<- Result,
	user string,
	msg json.RawMessage,
	threadId string,
) {
	commandStr := strings.Trim(string(msg[:]), "\"")
	command, err := ParseCommand(commandStr, threadId)
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
	if command.Type == CommandNone || (command.Type != CommandFortune && command.Target == CommandTargetNone) {
		log.Println("Unknown command: ", commandStr)

		//TODO(yurig): this should probably be the help menu
		resultChannel <- Result{Message: "huh? I don't know this command.", ThreadId: threadId}
		return
	}
	commandChannel <- *command
}

func writeMessage(flowID string, client *flowdock.Client) func(msg string, threadId string) {
	return func(msg string, threadId string) {
		_, _, err := client.Messages.Create(&flowdock.MessagesCreateOptions{
			FlowID:   flowID,
			Content:  msg,
			Event:    "message",
			ThreadId: threadId,
		})
		if err != nil {
			log.Println(err)
		}
	}
}

func assertNonEmptyEnvVar(envVar string, envVarName string) {
	if envVar == "" {
		log.Fatalf("The '%s' environment variable is required", envVarName)
	}
}

func assertNonEmptyFlag(flag *string, flagName string) {
	if *flag == "" {
		log.Fatalf("'%s' is a required parameter", flagName)
	}
}

type Result struct {
	Message  string
	ThreadId string
}
