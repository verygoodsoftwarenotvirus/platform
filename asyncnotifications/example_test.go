package asyncnotifications_test

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/verygoodsoftwarenotvirus/platform/v3/asyncnotifications"
)

func ExampleNewNoopAsyncNotifier() {
	notifier, err := asyncnotifications.NewNoopAsyncNotifier()
	if err != nil {
		panic(err)
	}

	err = notifier.Publish(context.Background(), "my-channel", &asyncnotifications.Event{
		Type: "greeting",
		Data: json.RawMessage(`{"message":"hello"}`),
	})

	fmt.Println(err)
	// Output: <nil>
}
