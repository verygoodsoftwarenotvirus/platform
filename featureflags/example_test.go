package featureflags_test

import (
	"context"
	"fmt"

	"github.com/verygoodsoftwarenotvirus/platform/v2/featureflags"
)

func ExampleNewNoopFeatureFlagManager() {
	mgr := featureflags.NewNoopFeatureFlagManager()
	defer mgr.Close()

	canUse, err := mgr.CanUseFeature(context.Background(), "user-1", "dark-mode")
	if err != nil {
		panic(err)
	}

	fmt.Println(canUse)
	// Output: false
}
