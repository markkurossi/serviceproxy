/*
 * client.go
 */

package authorizer

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/http"

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

	fmt.Fprint(w, "Goodbye, Client!\n")
}
