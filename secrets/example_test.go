package secrets_test

import (
	"context"
	"fmt"
	"os"

	"github.com/verygoodsoftwarenotvirus/platform/v2/secrets/env"
)

func Example_envSecretSource() {
	os.Setenv("EXAMPLE_SECRET", "s3cret")
	defer os.Unsetenv("EXAMPLE_SECRET")

	source := env.NewEnvSecretSource()
	defer source.Close()

	secret, err := source.GetSecret(context.Background(), "EXAMPLE_SECRET")
	if err != nil {
		panic(err)
	}

	fmt.Println(secret)
	// Output: s3cret
}
