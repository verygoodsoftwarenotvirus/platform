package llmcfg

import (
	"github.com/verygoodsoftwarenotvirus/platform/v3/llm"

	"github.com/samber/do/v2"
)

// RegisterLLMProvider registers an llm.Provider with the injector.
func RegisterLLMProvider(i do.Injector) {
	do.Provide[llm.Provider](i, func(i do.Injector) (llm.Provider, error) {
		return ProvideLLMProvider(do.MustInvoke[*Config](i))
	})
}
