/*
 * function.go
 */

package authorizer

import (
	"context"
	"fmt"
	"net/http"

	"cloud.google.com/go/pubsub"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
)

const (
	TOPIC_AUTHORIZER = "Authorizer"
)

func Authorizer(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Hello, World!\n")

	ctx := context.Background()

	credentials, err := google.FindDefaultCredentials(ctx, compute.ComputeScope)
	if err != nil {
		fmt.Fprintf(w, "google.FindDefaultCredentials: %s\n", err)
		return
	}

	client, err := pubsub.NewClient(ctx, credentials.ProjectID)
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

	_ = topic // TODO: use the topic.

	fmt.Fprint(w, "Goodbye, World!\n")
}
