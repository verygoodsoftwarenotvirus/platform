package mobilenotifications_test

import (
	"context"
	"fmt"

	"github.com/verygoodsoftwarenotvirus/platform/v3/mobilenotifications"
)

func Example_noopPushNotificationSender() {
	sender := &mobilenotifications.NoopPushNotificationSender{}

	err := sender.SendPush(context.Background(), "ios", "device-token-abc", mobilenotifications.PushMessage{
		Title: "New Message",
		Body:  "You have a new message!",
	})

	fmt.Println(err)
	// Output: <nil>
}
