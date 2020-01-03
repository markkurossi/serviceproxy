/*
 * client.go
 */

package authorizer

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
	"time"

	"cloud.google.com/go/pubsub"
)

func Client(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello, Client\n")

	ctx := context.Background()

	projectID, err := GetProjectID()
	if err != nil {
		fmt.Fprintf(w, "google.FindDefaultCredentials: %s\n", err)
		return
	}

	client, err := pubsub.NewClient(ctx, projectID)
	if err != nil {
		fmt.Fprintf(w, "NewClient failed: %s\n", err)
		return
	}

	// Create a topic for response.
	var id [16]byte
	_, err = rand.Read(id[:])
	if err != nil {
		fmt.Fprintf(w, "rand.Read: %s\n", err)
		return
	}
	respID := fmt.Sprintf("clnt%x", id[:])
	respTopic, err := client.CreateTopic(ctx, respID)
	if err != nil {
		fmt.Fprintf(w, "client.CreateTopic: %s\n", err)
		return
	}
	defer respTopic.Delete(ctx)

	// Subscribe for response.
	sub, err := client.CreateSubscription(ctx, fmt.Sprintf("resp%x", id[:]),
		pubsub.SubscriptionConfig{
			Topic:            respTopic,
			AckDeadline:      10 * time.Second,
			ExpirationPolicy: 25 * time.Hour,
		})
	if err != nil {
		fmt.Fprintf(w, "client.CreateSubscription: %s\n", err)
		return
	}
	defer sub.Delete(ctx)

	// Send request.
	reqTopic := client.Topic(TOPIC_AUTHORIZER)
	defer reqTopic.Stop()
	result := reqTopic.Publish(ctx, &pubsub.Message{
		Data: []byte("Hello, Authorizer!"),
		Attributes: map[string]string{
			"response": respID,
		},
	})
	reqID, err := result.Get(ctx)
	if err != nil {
		fmt.Fprintf(w, "reqTopic.Publish: %s\n", err)
		return
	}
	fmt.Fprintf(w, "request ID %s\n", reqID)

	// Receive response.
	cctx, cancel := context.WithCancel(ctx)
	err = sub.Receive(cctx, func(ctx context.Context, m *pubsub.Message) {
		fmt.Fprintf(w, "m: ID=%s, Data=%q, Attributes=%q\n",
			m.ID, m.Data, m.Attributes)
		m.Ack()
		cancel()
	})
	if err != nil {
		fmt.Fprintf(w, "sub.Receive: %s\n", err)
		return
	}

	fmt.Fprint(w, "Goodbye, Client!\n")
}
