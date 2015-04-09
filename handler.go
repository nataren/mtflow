package main

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
)

var _prsURL *url.URL
var _prsConfig *[]byte
var _prsAPIKey string
var _client *http.Client

// InitCommandHandler initializes a PullRequestService command
// handler, you need to have a URL to the service, a configuration
// definition, its API key, and a HTTP client so you can talk to it.
func InitCommandHandler(prsURL *url.URL, prsConfig *[]byte, prsAPIKey string, client *http.Client) {
	_prsURL = prsURL
	_prsConfig = prsConfig
	_prsAPIKey = prsAPIKey
	_client = client
}

// RunCommandHandler It is the top level function that must be called
// as a goroutine to handle a command.
func RunCommandHandler(commandChannel <-chan Command, resultChannel chan string) {
	if resultChannel == nil {
		panic("Must supply a channel for command results")
	}
	if commandChannel == nil {
		panic("Must supply a channel for commands to arrive on")
	}

	// handle commands until the end of time
	for {

		// TODO(yurig): this is currently has no timeout mechanism
		go handleCommand(<-commandChannel, resultChannel)
	}
}

func handleCommand(command Command, resultChannel chan string) {
	switch command.Type {
	case CommandStart:
		switch command.Target {
		case CommandTargetPR:
			log.Println("I will start processing of pull requests")
			startService := &http.Request{}
			startService.Method = "POST"
			q := _prsURL.Query()
			q.Set("apikey", prsAPIKey)
			startService.URL = &url.URL{
				Host:     _prsURL.Host,
				Scheme:   _prsURL.Scheme,
				Opaque:   "/host/services",
				RawQuery: q.Encode(),
			}
			startService.Header = map[string][]string{
				"Content-Type": {"application/xml"},
			}
			startService.Body = ioutil.NopCloser(bytes.NewReader(*_prsConfig))
			startService.ContentLength = int64(len(*_prsConfig))
			resp, err := _client.Do(startService)
			if err != nil {
				log.Panic(err)
			}
			defer resp.Body.Close()
			statusCode := resp.StatusCode
			if statusCode >= 200 && statusCode < 300 {
				msg := "Successfully started processing pull requests"
				log.Println(msg)
				resultChannel <- msg
			} else {
				msg := "Failed to start processing pull requests: " + resp.Status
				resultChannel <- msg
			}
		default:
			log.Printf("The modifier '%s' is not handled\n", command.Target)
		}
	case CommandStop:
		switch command.Target {
		case CommandTargetPR:
			log.Println("I will handle 'stop pr' command")
			stopService := &http.Request{}
			stopService.Method = "DELETE"
			q := _prsURL.Query()
			q.Set("apikey", prsAPIKey)
			_prsURL.RawQuery = q.Encode()
			stopService.URL = _prsURL
			resp, err := _client.Do(stopService)
			if err != nil {
				log.Panic(err)
			}
			defer resp.Body.Close()
			statusCode := resp.StatusCode
			if statusCode >= 200 && statusCode < 300 {
				msg := "Successfully stopped processing pull requests"
				log.Println(msg)
				resultChannel <- msg
			} else {
				msg := "Failed to stop processing pull requests: " + resp.Status
				log.Println(msg)
				resultChannel <- msg
			}
		default:
			log.Printf("The modifier '%s' is not handled\n", command.Target)
		}
	default:
		log.Printf("The command '%+v' is not handled\n", command)
	}
}
