package notifications_test

import (
	"context"
	"fmt"

	"github.com/verygoodsoftwarenotvirus/platform/v2/notifications"
)

func Example_noopPushNotificationSender() {
	sender := &notifications.NoopPushNotificationSender{}

	err := sender.SendPush(context.Background(), "ios", "device-token-abc", notifications.PushMessage{
		Title: "New Message",
		Body:  "You have a new message!",
	})

	fmt.Println(err)
	// Output: <nil>
}
