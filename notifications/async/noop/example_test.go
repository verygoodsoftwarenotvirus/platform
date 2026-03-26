package noop_test

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/verygoodsoftwarenotvirus/platform/v4/notifications/async"
	"github.com/verygoodsoftwarenotvirus/platform/v4/notifications/async/noop"
)

func ExampleNewAsyncNotifier() {
	notifier, err := noop.NewAsyncNotifier()
	if err != nil {
		panic(err)
	}

	err = notifier.Publish(context.Background(), "my-channel", &async.Event{
		Type: "greeting",
		Data: json.RawMessage(`{"message":"hello"}`),
	})

	fmt.Println(err)
	// Output: <nil>
}
