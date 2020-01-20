//
// client.go
//
// Copyright (c) 2020 Markku Rossi
//
// All rights reserved.
//

package authorizer

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/markkurossi/cloudsdk/api/auth"
)

var (
	rePath = regexp.MustCompilePOSIX(`^/clients/([a-f0-9]{16,32})$`)
)

// Clients hand REST calls to the "/clients" URI.
func Clients(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("%s: %s\n", r.Method, r.URL.Path)

	token := auth.Authorize(w, r, REALM, tokenVerifier, nil)
	if token == nil {
		return
	}

	ctx := context.Background()

	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		Error500f(w, "NewClient failed: %s", err)
		return
	}

	switch r.Method {
	case "POST":
		// Register new client.
		id, err := NewID()
		if err != nil {
			Error500f(w, "NewID: %s", err)
			return
		}
		// Create a topic for response.
		respTopic, err := client.CreateTopic(ctx, id.Topic())
		if err != nil {
			Error500f(w, "client.CreateTopic: %s", err)
			return
		}
		// Subscribe for response.
		_, err = client.CreateSubscription(ctx, id.Subscription(),
			pubsub.SubscriptionConfig{
				Topic:            respTopic,
				AckDeadline:      10 * time.Second,
				ExpirationPolicy: 25 * time.Hour,
			})
		if err != nil {
			Error500f(w, "client.CreateSubscription: %s", err)
			return
		}
		result := &ClientConnectResult{
			URL: "/clients/" + id.String(),
			ID:  id.String(),
		}
		data, err := json.Marshal(result)
		if err != nil {
			Error500f(w, "json.Marshal: %s", err)
			return
		}
		w.Write(data)

	default:
		Errorf(w, http.StatusBadRequest, "Unsupported method %s", r.Method)
	}
}

// Clients hand REST calls to the "/clients/{ID}" URI.
func Client(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("%s: %s\n", r.Method, r.URL.Path)

	token := auth.Authorize(w, r, REALM, tokenVerifier, nil)
	if token == nil {
		return
	}

	m := rePath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		Errorf(w, http.StatusBadRequest, "Invalid client URL path")
		return
	}
	clientID := m[1]

	ctx := context.Background()

	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		Error500f(w, "NewClient failed: %s", err)
		return
	}
	id, err := ParseID(clientID)
	if err != nil {
		Error500f(w, "Invalid client ID: %s", err)
		return
	}

	switch r.Method {
	case "POST":
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			Error500f(w, "ioutil.ReadAll: %s", err)
			return
		}
		msg := new(Message)
		err = json.Unmarshal(data, msg)
		if err != nil {
			Errorf(w, http.StatusBadRequest, "Invalid message data: %s", err)
			return
		}
		payload, err := msg.Bytes()
		if err != nil {
			Errorf(w, http.StatusBadRequest, "Invalid message payload: %s", err)
			return
		}
		id, err := ParseID(msg.From)
		if err != nil {
			Errorf(w, http.StatusBadRequest, "Invalid from ID '%s': %s",
				msg.From, err)
			return
		}

		// Send request.
		reqTopic := client.Topic(TOPIC_AUTHORIZER)
		result := reqTopic.Publish(ctx, &pubsub.Message{
			Data: payload,
			Attributes: map[string]string{
				ATTR_RESPONSE: id.String(),
			},
		})
		_, err = result.Get(ctx)
		reqTopic.Stop()
		if err != nil {
			Error500f(w, "reqTopic.Publish: %s", err)
			return
		}
		fallthrough

	case "GET":
		// Receive response.

		var response *pubsub.Message

		var timeout time.Duration
		if r.Method == "POST" {
			timeout = 30 * time.Second
		} else {
			timeout = 15 * time.Second
		}

		cctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		sub := client.Subscription(id.Subscription())
		err = sub.Receive(cctx, func(ctx context.Context, m *pubsub.Message) {
			if false {
				fmt.Printf("m: ID=%s, Data=%q, Attributes=%q\n",
					m.ID, m.Data, m.Attributes)
			}
			response = m
			m.Ack()
			cancel()
		})
		if err != nil {
			Error500f(w, "sub.Receive: %s", err)
			return
		}
		if response == nil {
			if r.Method == "POST" {
				w.WriteHeader(http.StatusAccepted)
			} else {
				w.WriteHeader(http.StatusRequestTimeout)
			}
			return
		}

		msg := new(Message)
		msg.SetBytes(response.Data)

		data, err := json.Marshal(msg)
		if err != nil {
			Error500f(w, "json.Marshal: %s", err)
			return
		}
		w.Write(data)

	case "DELETE":
		var msg string

		err := client.Subscription(id.Subscription()).Delete(ctx)
		if err != nil {
			msg = fmt.Sprintf("Subscription: %s", err)
		}
		err = client.Topic(id.Topic()).Delete(ctx)
		if err != nil {
			if len(msg) > 0 {
				msg += ", "
			}
			msg += fmt.Sprintf("Topic: %s", err)
		}
		if len(msg) > 0 {
			Error500f(w, "%s", msg)
		} else {
			w.WriteHeader(http.StatusOK)
		}

	default:
		Errorf(w, http.StatusBadRequest, "Unsupported method %s", r.Method)
	}
}
