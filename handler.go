package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os/exec"
	"runtime"
	"time"
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
func RunCommandHandler(commandChannel <-chan Command, resultChannel chan Result) {
	if resultChannel == nil {
		panic("Must supply a channel for command results")
	}
	if commandChannel == nil {
		panic("Must supply a channel for commands to arrive on")
	}

	// handle commands until the end of time
	for {
		newCommand := <-commandChannel
		go func() {

			// This is a buffered channel so that a goroutine that has timed out
			// has a place to respond to
			dedicatedResultChan := make(chan string, 1)

			// Fire off handling of command
			go handleCommand(newCommand, dedicatedResultChan)
			select {
			case res := <-dedicatedResultChan:
				resultChannel <- Result{Message: res, ThreadId: newCommand.ThreadId}
			case <-time.After(30 * time.Second):
				resultChannel <- Result{Message: "The operation took more than 30 seconds and timed out, sorry :(", ThreadId: newCommand.ThreadId}
			}
		}()
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
			q.Set("apikey", _prsAPIKey)
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
				log.Println(err)
				return
			}
			defer func() {
				if resp != nil {
					resp.Body.Close()
				}
			}()
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
			msg := fmt.Sprintf("The modifier '%v' is not handled\n", command.Target)
			log.Println(msg)
			resultChannel <- msg
		}
	case CommandStop:
		switch command.Target {
		case CommandTargetPR:
			log.Println("I will handle 'stop pr' command")
			stopService := &http.Request{}
			stopService.Method = "DELETE"
			q := _prsURL.Query()
			q.Set("apikey", _prsAPIKey)
			_prsURL.RawQuery = q.Encode()
			stopService.URL = _prsURL
			resp, err := _client.Do(stopService)
			if err != nil {
				log.Println(err)
				return
			}
			defer func() {
				if resp != nil {
					resp.Body.Close()
				}
			}()
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
			msg := fmt.Sprintf("The modifier '%v' is not handled\n", command.Target)
			log.Println(msg)
			resultChannel <- msg
		}
	case CommandStatus:
		switch command.Target {
		case CommandTargetMtFlow:
			log.Println("I will handle 'status mtflow' command")

			// get some memory statistics
			var memStats runtime.MemStats
			runtime.ReadMemStats(&memStats)
			resultChannel <- fmt.Sprintf("I am chugging along, thanks for asking.\n\n# of Goroutines: %v\n# of CPU: %v\nTotal Memory: %v", runtime.NumGoroutine(), runtime.NumCPU(), memStats.Alloc)
		}
	case CommandFortune:
		log.Println("I will handle the 'fortune' command")
		fortuneCookie, err := exec.Command("fortune").Output()
		if err != nil {
			log.Println(err)
			resultChannel <- "No cookie for you!"
			return
		}
		fortuneCookieString := string(fortuneCookie)
		cowsaying, err := exec.Command("cowsay", fortuneCookieString).Output()
		if err != nil {
			log.Println(err)
			resultChannel <- fortuneCookieString
			return
		}
		resultChannel <- fmt.Sprintf("```\n%s\n```", string(cowsaying))
	default:
		log.Printf("The command '%+v' is not handled\n", command)
	}
}
