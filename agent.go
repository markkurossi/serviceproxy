/*
 * agent.go
 */

package authorizer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/markkurossi/authorizer/api"
)

var (
	reAgentPath = regexp.MustCompilePOSIX(`^/agents/([a-zA-Z][a-zA-Z0-9]+)$`)
)

func Agents(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s: %s\n", r.Method, r.URL.Path)

	ctx := context.Background()

	projectID, err := GetProjectID()
	if err != nil {
		Error500f(w, "google.FindDefaultCredentials: %s", err)
		return
	}
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
		result := &api.ServerConnectResult{
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
	log.Printf("%s: %s\n", r.Method, r.URL.Path)

	m := reAgentPath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		Errorf(w, http.StatusBadRequest, "Invalid agent URL path")
		return
	}
	agentID := m[1]

	ctx := context.Background()

	projectID, err := GetProjectID()
	if err != nil {
		Error500f(w, "google.FindDefaultCredentials: %s", err)
		return
	}

	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		Error500f(w, "NewClient failed: %s", err)
		return
	}

	sub := client.Subscription(agentID)

	switch r.Method {
	case "GET":
		var response *pubsub.Message

		cctx, cancel := context.WithTimeout(ctx, 60*time.Second)
		defer cancel()
		err = sub.Receive(cctx, func(ctx context.Context, m *pubsub.Message) {
			fmt.Printf("m: ID=%s, Data=%q, Attributes=%q\n",
				m.ID, m.Data, m.Attributes)
			response = m
			m.Ack()
			cancel()
		})
		if err != nil {
			Error500f(w, "sub.Receive: %s", err)
			return
		}
		if response == nil {
			w.WriteHeader(http.StatusNoContent)
		} else {
			w.Write(response.Data)
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
