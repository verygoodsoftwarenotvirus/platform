package noop_test

import (
	"context"
	"fmt"

	"github.com/verygoodsoftwarenotvirus/platform/v5/notifications/mobile"
	"github.com/verygoodsoftwarenotvirus/platform/v5/notifications/mobile/noop"
)

func Example_pushNotificationSender() {
	sender := noop.NewPushNotificationSender()

	err := sender.SendPush(context.Background(), "ios", "device-token-abc", mobile.PushMessage{
		Title: "New Message",
		Body:  "You have a new message!",
	})

	fmt.Println(err)
	// Output: <nil>
}
