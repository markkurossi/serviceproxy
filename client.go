/*
 * client.go
 */

package authorizer

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"time"

	"cloud.google.com/go/pubsub"
)

var (
	rePath = regexp.MustCompilePOSIX(`^/client/([a-f0-9]{32,64})$`)
)

func Client(w http.ResponseWriter, r *http.Request) {
	log.Printf("Client: path=%s\n", r.URL.Path)

	m := rePath.FindStringSubmatch(r.URL.Path)
	if m == nil {
		http.Error(w, "Invalid client URL Path", http.StatusBadRequest)
		return
	}
	clientID := m[1]
	fmt.Printf("ClientID: %s\n", clientID)

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

	// Create a topic for response.
	respID := fmt.Sprintf("clnt%s", clientID)
	respTopic := client.Topic(respID)
	ok, err := respTopic.Exists(ctx)
	if err != nil {
		Error500f(w, "topic.Exists: %s", err)
		return
	}
	if !ok {
		respTopic, err = client.CreateTopic(ctx, respID)
		if err != nil {
			Error500f(w, "client.CreateTopic: %s", err)
			return
		}
	}

	// Subscribe for response.
	sub := client.Subscription(respID)
	ok, err = sub.Exists(ctx)
	if err != nil {
		Error500f(w, "sub.Exists: %s", err)
		return
	}
	if !ok {
		sub, err = client.CreateSubscription(ctx, respID,
			pubsub.SubscriptionConfig{
				Topic:            respTopic,
				AckDeadline:      10 * time.Second,
				ExpirationPolicy: 25 * time.Hour,
			})
		if err != nil {
			Error500f(w, "client.CreateSubscription: %s", err)
			return
		}
	}

	switch r.Method {
	case "POST":
		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			Error500f(w, "ioutil.ReadAll: %s", err)
			return
		}
		// Send request.
		reqTopic := client.Topic(TOPIC_AUTHORIZER)
		defer reqTopic.Stop()
		result := reqTopic.Publish(ctx, &pubsub.Message{
			Data: data,
			Attributes: map[string]string{
				"response": respID,
			},
		})
		reqID, err := result.Get(ctx)
		if err != nil {
			Error500f(w, "reqTopic.Publish: %s", err)
			return
		}
		fmt.Fprintf(w, "request ID %s\n", reqID)
		fallthrough

	case "GET":
		// Receive response.

		var response *pubsub.Message

		cctx, cancel := context.WithCancel(ctx)
		err = sub.Receive(cctx, func(ctx context.Context, m *pubsub.Message) {
			fmt.Fprintf(w, "m: ID=%s, Data=%q, Attributes=%q\n",
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
			w.WriteHeader(http.StatusAccepted)
		} else {
			w.Write(response.Data)
		}
	}
}
