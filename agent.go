//
// agent.go
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
	reAgentPath = regexp.MustCompilePOSIX(`^/agents/([a-zA-Z][a-zA-Z0-9]+)$`)
)

func Agents(w http.ResponseWriter, r *http.Request) {
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
		// Register new agent.
		topic := client.Topic(TOPIC_AUTHORIZER)
		ok, err := topic.Exists(ctx)
		if err != nil {
			Error500f(w, "topic.Exists: %s", err)
			return
		}
		if !ok {
			topic, err = client.CreateTopic(ctx, TOPIC_AUTHORIZER)
			if err != nil {
				Error500f(w, "client.CreateTopic: %s", err)
				return
			}
		}

		sub := client.Subscription(SUB_REQUESTS)
		ok, err = sub.Exists(ctx)
		if err != nil {
			Error500f(w, "sub.Exists: %s", err)
			return
		}
		if !ok {
			sub, err = client.CreateSubscription(ctx, SUB_REQUESTS,
				pubsub.SubscriptionConfig{
					Topic:            topic,
					AckDeadline:      10 * time.Second,
					ExpirationPolicy: 25 * time.Hour,
				})
			if err != nil {
				Error500f(w, "client.CreateSubscription: %s", err)
				return
			}
		}
		result := &ServerConnectResult{
			URL: "/agents/" + SUB_REQUESTS,
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

func Agent(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("%s: %s\n", r.Method, r.URL.Path)

	token := auth.Authorize(w, r, REALM, tokenVerifier, nil)
	if token == nil {
		return
	}

	m := reAgentPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		Errorf(w, http.StatusBadRequest, "Invalid agent URL path")
		return
	}
	agentID := m[1]

	ctx := context.Background()

	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		Error500f(w, "NewClient failed: %s", err)
		return
	}

	sub := client.Subscription(agentID)

	switch r.Method {
	case "GET":
		var request *pubsub.Message

		cctx, cancel := context.WithTimeout(ctx, 30*time.Second)
		defer cancel()

		// XXX Request processing must be moved inside receive
		// function. We must ACK the message only if we correctly
		// passed it to our caller.
		err = sub.Receive(cctx, func(ctx context.Context, m *pubsub.Message) {
			if false {
				fmt.Printf("m: ID=%s, Data=%q, Attributes=%q\n",
					m.ID, m.Data, m.Attributes)
			}
			request = m
			m.Ack()
			cancel()
		})
		if err != nil {
			Error500f(w, "sub.Receive: %s", err)
			return
		}
		if request == nil {
			w.WriteHeader(http.StatusRequestTimeout)
			return
		}

		from, ok := request.Attributes[ATTR_RESPONSE]
		if !ok {
			Errorf(w, http.StatusBadRequest, "No sender ID in message")
			return
		}

		msg := &Message{
			From: from,
		}
		msg.SetBytes(request.Data)
		data, err := json.Marshal(msg)
		if err != nil {
			Error500f(w, "json.Marshal: %s", err)
			return
		}
		w.Write(data)

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
		id, err := ParseID(msg.To)
		if err != nil {
			Errorf(w, http.StatusBadRequest, "Invalid destination ID '%s': %s",
				msg.To, err)
			return
		}

		topic := client.Topic(id.Topic())
		result := topic.Publish(ctx, &pubsub.Message{
			Data: payload,
		})
		_, err = result.Get(ctx)
		topic.Stop()
		if err != nil {
			Error500f(w, "topic.Publish: %s", err)
			return
		}

	default:
		Errorf(w, http.StatusBadRequest, "Unsupported method %s", r.Method)
	}
}

func handleRequest(client *pubsub.Client, req *pubsub.Message) error {
	topicID, ok := req.Attributes["response"]
	if !ok {
		return fmt.Errorf("No response ID in request")
	}
	ctx := context.Background()

	topic := client.Topic(topicID)
	defer topic.Stop()
	result := topic.Publish(ctx, &pubsub.Message{
		Data: []byte("Hello, Client!"),
	})
	_, err := result.Get(ctx)
	return err
}
