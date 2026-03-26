package noop_test

import (
	"context"
	"fmt"

	"github.com/verygoodsoftwarenotvirus/platform/v4/featureflags/noop"
)

func ExampleNewFeatureFlagManager() {
	mgr := noop.NewFeatureFlagManager()
	defer mgr.Close()

	canUse, err := mgr.CanUseFeature(context.Background(), "user-1", "dark-mode")
	if err != nil {
		panic(err)
	}

	fmt.Println(canUse)
	// Output: false
}
