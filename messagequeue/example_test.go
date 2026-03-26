package messagequeue_test

import (
	"context"
	"fmt"

	"github.com/verygoodsoftwarenotvirus/platform/v3/messagequeue"
)

func ExampleNewNoopPublisher() {
	pub := messagequeue.NewNoopPublisher()
	defer pub.Stop()

	err := pub.Publish(context.Background(), map[string]string{"event": "user.created"})
	fmt.Println(err)
	// Output: <nil>
}

func ExampleNewNoopPublisherProvider() {
	provider := messagequeue.NewNoopPublisherProvider()
	defer provider.Close()

	pub, err := provider.ProvidePublisher(context.Background(), "user-events")
	if err != nil {
		panic(err)
	}

	fmt.Println(pub != nil)
	// Output: true
}
