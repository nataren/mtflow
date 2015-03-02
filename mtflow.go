package main

import(
  "github.com/wm/go-flowdock/flowdock"
  "code.google.com/p/goauth2/oauth"
  "os"
  "log"
)

const (
  FLOWDOCK_API_TOKEN = "FLOWDOCK_API_TOKEN"
)

func main() {

  // Config
  accessToken := os.Getenv(FLOWDOCK_API_TOKEN)
  if accessToken == "" {
    log.Fatalf("%s environment variable not found", FLOWDOCK_API_TOKEN)
  }

  // Auth
  t := &oauth.Transport{
    Token: &oauth.Token { AccessToken: accessToken },
  }
  client := flowdock.NewClient(t.Client())

  // list all flows the authenticated user is a member of or can join
  flows, _, err := client.Flows.List(true, nil)
  if err != nil {
    log.Fatal(err)
  }
  log.Println(flows)
}
