package email_test

import (
	"context"
	"fmt"

	"github.com/verygoodsoftwarenotvirus/platform/v3/email"
)

func ExampleNewNoopEmailer() {
	emailer, err := email.NewNoopEmailer()
	if err != nil {
		panic(err)
	}

	err = emailer.SendEmail(context.Background(), &email.OutboundEmailMessage{
		ToAddress:   "user@example.com",
		ToName:      "User",
		FromAddress: "noreply@example.com",
		FromName:    "App",
		Subject:     "Welcome!",
		HTMLContent: "<h1>Hello</h1>",
	})

	fmt.Println(err)
	// Output: <nil>
}
