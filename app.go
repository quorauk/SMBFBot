package main

import (
  "bytes"
  "fmt"
  "os"
  "encoding/json"
  "log"
  "net/http"

  "github.com/gorilla/mux"
)

const (
  FACEBOOK_API = "https://graph.facebook.com/v2.6/me/messages?access_token=%s"
  IMAGE        = "http://cdn.discordapp.com/attachments/254378086659784704/380863202365538305/image.png"
)


type Callback struct {
  Object string `json:"object,omitempty"`
  Entry  []struct {
    ID        string      `json:"id,omitempty"`
    Time      int         `json:"time,omitempty"`
    Messaging []Messaging `json:"messaging,omitempty"`
  } `json:"entry,omitempty"`
}

type Messaging struct {
  Sender    User    `json:"sender,omitempty"`
  Recipient User    `json:"recipient,omitempty"`
  Timestamp int     `json:"timestamp,omitempty"`
  Message   Message `json:"message,omitempty"`
}

type User struct {
  ID string `json:"id,omitempty"`
}

type Message struct {
  MID        string `json:"mid,omitempty"`
  Text       string `json:"text,omitempty"`
  QuickReply *struct {
    Payload string `json:"payload,omitempty"`
  } `json:"quick_reply,omitempty"`
  Attachments *[]Attachment `json:"attachments,omitempty"`
  Attachment  *Attachment   `json:"attachment,omitempty"`
}

type Attachment struct {
  Type    string  `json:"type,omitempty"`
  Payload Payload `json:"payload,omitempty"`
}

type Response struct {
  Recipient User    `json:"recipient,omitempty"`
  Message   Message `json:"message,omitempty"`
}

type Payload struct {
  URL string `json:"url,omitempty"`
}


func HomeEndpoint(w http.ResponseWriter, r *http.Request) {
  fmt.Fprintln(w, "Hello from mlabouardy :)")
}

func VerificationEndpoint(w http.ResponseWriter, r *http.Request) {
  challenge := r.URL.Query().Get("hub.challenge")
  token := r.URL.Query().Get("hub.verify_token")

  if token == os.Getenv("VERIFY_TOKEN") {
    w.WriteHeader(200)
    w.Write([]byte(challenge))
  } else {
    w.WriteHeader(404)
    w.Write([]byte("Error, wrong validation token"))
  }
}

func MessagesEndpoint(w http.ResponseWriter, r *http.Request) {
  var callback Callback
  json.NewDecoder(r.Body).Decode(&callback)
  if callback.Object == "page" {
    for _, entry := range callback.Entry {
      for _, event := range entry.Messaging {
        ProcessMessage(event)
      }
    }
    w.WriteHeader(200)
    w.Write([]byte("Got Your Message"))
  } else {
    w.WriteHeader(404)
    w.Write([]byte("Message not supported"))
  }
}

func ProcessMessage(event Messaging) {
  client := &http.Client{}
  response := Response{
    Recipient: User {
      ID: event.Sender.ID,
    },
    Message: Message{
      Attachment: &Attachment{
	Type: "image",
	Payload: Payload{
	  URL: IMAGE,
	},
      },
    },
  }
  body := new(bytes.Buffer)
  json.NewEncoder(body).Encode(&response)
  url := fmt.Sprintf(FACEBOOK_API, os.Getenv("PAGE_ACCESS_TOKEN"))
  req, err := http.NewRequest("POST", url, body)
  req.Header.Add("Content-Type", "application/json")
  if err != nil {
    log.Fatal(err)
  }

  resp, err := client.Do(req)
  if err != nil {
    log.Fatal(err)
  }
  defer resp.Body.Close()
}

func main() {
  r := mux.NewRouter()
  r.HandleFunc("/webhook", VerificationEndpoint).Methods("GET")
  r.HandleFunc("/webhook", MessagesEndpoint).Methods("POST")

  if err := http.ListenAndServe("0.0.0.0:8080", r); err != nil {
    log.Fatal(err)
  }
}
