/*
 * function.go
 */

package authorizer

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"cloud.google.com/go/pubsub"
)

const (
	TOPIC_AUTHORIZER = "Authorizer"
	SUB_REQUESTS     = "Requests"
)

func Authorizer(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello, World!\n")

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

	topic := client.Topic(TOPIC_AUTHORIZER)
	ok, err := topic.Exists(ctx)
	if err != nil {
		fmt.Fprintf(w, "topic.Exists: %s\n", err)
		return
	}
	if !ok {
		fmt.Fprintf(w, "Creating topic %s\n", TOPIC_AUTHORIZER)
		topic, err = client.CreateTopic(ctx, TOPIC_AUTHORIZER)
		if err != nil {
			fmt.Fprintf(w, "client.CreateTopic: %s\n", err)
			return
		}
	}

	sub := client.Subscription(SUB_REQUESTS)
	ok, err = sub.Exists(ctx)
	if err != nil {
		fmt.Fprintf(w, "sub.Exists: %s\n", err)
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
			fmt.Fprintf(w, "client.CreateSubscription: %s\n", err)
			return
		}
	}

	cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	err = sub.Receive(cctx, func(ctx context.Context, m *pubsub.Message) {
		fmt.Fprintf(w, "m: %v\n", m)
		m.Ack()
		cancel()
	})
	if err != nil {
		fmt.Fprintf(w, "sub.Receive: %s\n", err)
		return
	}

	fmt.Fprint(w, "Goodbye, World!\n")
}
