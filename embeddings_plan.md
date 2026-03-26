# Plan: `embeddings` package for `github.com/verygoodsoftwarenotvirus/platform/v4`

## Reference

This plan is modelled on the `llm` package at
`github.com/verygoodsoftwarenotvirus/platform/v4/llm`, with the deliberate addition of
`logging.Logger` and `tracing.Tracer` injection into every provider constructor — an
improvement over `llm` that will be backfilled there separately.

---

## Relevant import paths

```go
"github.com/verygoodsoftwarenotvirus/platform/v4/observability/logging"
"github.com/verygoodsoftwarenotvirus/platform/v4/observability/tracing"
```

Key types:

- `logging.Logger` — use `logger.Error(description, err)`, `logger.WithValue(k, v)`, etc.
- `logging.NewNoopLogger()` — use in tests
- `logging.EnsureLogger(logger)` — guards against a nil logger being passed in
- `tracing.Tracer` — interface with `StartSpan(ctx) (context.Context, tracing.Span)` and `StartCustomSpan(ctx, name, ...opts)`
- `tracing.NewTracerForTest(name)` — use in tests
- `tracing.AttachErrorToSpan(span, description, err)` — standard span error attachment helper

---

## Package Layout

```
platform/v3/
└── embeddings/
    ├── doc.go               # package doc comment
    ├── embeddings.go        # Embedder interface, Input, Embedding, NewNoopEmbedder
    ├── config/
    │   ├── config.go        # Config struct + ValidateWithContext + ProvideEmbedder method
    │   ├── providers.go     # top-level ProvideEmbedder function
    │   └── do.go            # RegisterEmbedder(i do.Injector)
    ├── mock/
    │   └── mock.go          # testify/mock-based Embedder mock
    ├── openai/
    │   ├── config.go        # openai.Config + ValidateWithContext
    │   └── openai.go        # NewEmbedder(cfg, logger, tracer) (embeddings.Embedder, error)
    ├── ollama/
    │   ├── config.go        # ollama.Config + ValidateWithContext
    │   └── ollama.go        # NewEmbedder(cfg, logger, tracer) (embeddings.Embedder, error)
    └── cohere/
        ├── config.go        # cohere.Config + ValidateWithContext
        └── cohere.go        # NewEmbedder(cfg, logger, tracer) (embeddings.Embedder, error)
```

---

## Root Package (`embeddings/embeddings.go`)

### `Input`

```go
package embeddings

// Input is the content to be embedded.
type Input struct {
    // Content is the text to embed.
    Content string

    // Model optionally overrides the provider's configured DefaultModel.
    // Leave empty to use the default from the provider's Config.
    Model string
}
```

### `Embedding`

```go
// Embedding is the result of embedding a single piece of content.
// It carries provenance alongside the vector so that re-embedding
// and ETL pipelines can be driven from the stored result alone.
type Embedding struct {
    // Vector is the embedding itself.
    Vector []float32

    // SourceText is the text that was embedded.
    SourceText string

    // Model is the specific model used, e.g. "text-embedding-3-small".
    Model string

    // Provider is the backend used, e.g. "openai", "ollama", "cohere".
    Provider string

    // Dimensions is len(Vector). Stored explicitly for storage schema
    // validation without deserializing the full vector.
    Dimensions int

    // GeneratedAt is when the embedding was produced. Useful for
    // detecting staleness when a provider silently updates a model.
    GeneratedAt time.Time
}
```

### `Embedder` interface

```go
// Embedder generates vector embeddings for text.
type Embedder interface {
    GenerateEmbedding(ctx context.Context, input *Input) (*Embedding, error)
}
```

### No-op (also in `embeddings.go`)

```go
type noopEmbedder struct{}

// NewNoopEmbedder returns an Embedder that returns an empty vector and no error.
// Intended for tests and local development.
func NewNoopEmbedder() Embedder {
    return &noopEmbedder{}
}

func (n *noopEmbedder) GenerateEmbedding(_ context.Context, input *Input) (*Embedding, error) {
    return &Embedding{
        Vector:      []float32{},
        SourceText:  input.Content,
        Model:       "noop",
        Provider:    "noop",
        Dimensions:  0,
        GeneratedAt: time.Now(),
    }, nil
}
```

### `doc.go`

```go
// Package embeddings provides a vector embedding interface with implementations
// for OpenAI, Ollama, and Cohere providers.
package embeddings
```

---

## Config Package (`embeddings/config/`)

Three-file layout identical to `llm/config`. Package name is `embeddingscfg`.

### `config.go`

```go
package embeddingscfg

const (
    ProviderOpenAI  = "openai"
    ProviderOllama  = "ollama"
    ProviderCohere  = "cohere"
)

// Config is the configuration for the embeddings provider.
type Config struct {
    OpenAI   *openai.Config `env:"init" envPrefix:"OPENAI_" json:"openai"`
    Ollama   *ollama.Config `env:"init" envPrefix:"OLLAMA_" json:"ollama"`
    Cohere   *cohere.Config `env:"init" envPrefix:"COHERE_" json:"cohere"`
    Provider string         `env:"PROVIDER"                 json:"provider"`
}

// ValidateWithContext validates the config.
func (c *Config) ValidateWithContext(ctx context.Context) error { ... }

// ProvideEmbedder provides an Embedder based on config.
func (c *Config) ProvideEmbedder(ctx context.Context, logger logging.Logger, tracer tracing.Tracer) (embeddings.Embedder, error) {
    return ProvideEmbedder(c, logger, tracer)
}
```

### `providers.go`

```go
package embeddingscfg

// ProvideEmbedder provides an Embedder from config.
func ProvideEmbedder(c *Config, logger logging.Logger, tracer tracing.Tracer) (embeddings.Embedder, error) {
    switch c.Provider {
    case ProviderOpenAI:
        return openai.NewEmbedder(c.OpenAI, logger, tracer)
    case ProviderOllama:
        return ollama.NewEmbedder(c.Ollama, logger, tracer)
    case ProviderCohere:
        return cohere.NewEmbedder(c.Cohere, logger, tracer)
    default:
        return embeddings.NewNoopEmbedder(), nil
    }
}
```

### `do.go`

```go
package embeddingscfg

// RegisterEmbedder registers an embeddings.Embedder with the injector.
func RegisterEmbedder(i do.Injector) {
    do.Provide(i, func(i do.Injector) (embeddings.Embedder, error) {
        cfg    := do.MustInvoke[*Config](i)
        logger := do.MustInvoke[logging.Logger](i)
        tracer := do.MustInvoke[tracing.Tracer](i)
        return ProvideEmbedder(cfg, logger, tracer)
    })
}
```

---

## Provider Packages

Each provider has a `config.go` and an implementation file. The constructor signature for
all three is:

```go
func NewEmbedder(cfg *Config, logger logging.Logger, tracer tracing.Tracer) (embeddings.Embedder, error)
```

The concrete struct stored by each provider implementation should look like:

```go
type embedder struct {
    logger logging.Logger
    tracer tracing.Tracer
    client *http.Client
    cfg    *Config
}
```

### How to use logger and tracer inside `GenerateEmbedding`

```go
func (e *embedder) GenerateEmbedding(ctx context.Context, input *embeddings.Input) (*embeddings.Embedding, error) {
    ctx, span := e.tracer.StartSpan(ctx)
    defer span.End()

    // ... build and execute request ...

    if err != nil {
        tracing.AttachErrorToSpan(span, "generating embedding", err)
        e.logger.Error("generating embedding", err)
        return nil, err
    }

    return result, nil
}
```

### OpenAI (`embeddings/openai/`)

**`config.go`**

```go
package openai

// Config configures the OpenAI embeddings provider.
type Config struct {
    APIKey       string        `env:"API_KEY"       json:"apiKey,omitempty"`
    BaseURL      string        `env:"BASE_URL"      json:"baseURL,omitempty"`
    DefaultModel string        `env:"DEFAULT_MODEL" json:"defaultModel,omitempty"`
    Timeout      time.Duration `env:"TIMEOUT"       json:"timeout"`
}

func (c *Config) ValidateWithContext(ctx context.Context) error {
    // return error if APIKey is empty
}
```

**`openai.go`**

```go
func NewEmbedder(cfg *Config, logger logging.Logger, tracer tracing.Tracer) (embeddings.Embedder, error)
```

- Endpoint: `https://api.openai.com/v1/embeddings` (overridable via `BaseURL`)
- Auth: `Authorization: Bearer <APIKey>` header
- Default model: `text-embedding-3-small`
- Request body: `{"input": "<text>", "model": "<model>", "encoding_format": "float"}`
- Response: extract `data[0].embedding` (`[]float64` → convert to `[]float32`)
- Set `Embedding.Provider` to `"openai"`

### Ollama (`embeddings/ollama/`)

**`config.go`**

```go
package ollama

// Config configures the Ollama embeddings provider.
type Config struct {
    BaseURL      string        `env:"BASE_URL"      json:"baseURL,omitempty"`
    DefaultModel string        `env:"DEFAULT_MODEL" json:"defaultModel,omitempty"`
    Timeout      time.Duration `env:"TIMEOUT"       json:"timeout"`
}

func (c *Config) ValidateWithContext(ctx context.Context) error {
    // no required fields; BaseURL defaults in NewEmbedder
}
```

**`ollama.go`**

```go
func NewEmbedder(cfg *Config, logger logging.Logger, tracer tracing.Tracer) (embeddings.Embedder, error)
```

- Default `BaseURL`: `http://localhost:11434` (set in `NewEmbedder` if `cfg.BaseURL` is empty)
- No auth header
- Default model: `nomic-embed-text`
- Endpoint: `POST <BaseURL>/api/embed`
- Request body: `{"model": "<model>", "input": "<text>"}`
- Response: extract `embeddings[0]` (`[]float64` → convert to `[]float32`)
- Set `Embedding.Provider` to `"ollama"`

### Cohere (`embeddings/cohere/`)

**`config.go`**

```go
package cohere

// Config configures the Cohere embeddings provider.
type Config struct {
    APIKey       string        `env:"API_KEY"       json:"apiKey,omitempty"`
    BaseURL      string        `env:"BASE_URL"      json:"baseURL,omitempty"`
    DefaultModel string        `env:"DEFAULT_MODEL" json:"defaultModel,omitempty"`
    Timeout      time.Duration `env:"TIMEOUT"       json:"timeout"`
}

func (c *Config) ValidateWithContext(ctx context.Context) error {
    // return error if APIKey is empty
}
```

**`cohere.go`**

```go
func NewEmbedder(cfg *Config, logger logging.Logger, tracer tracing.Tracer) (embeddings.Embedder, error)
```

- Endpoint: `https://api.cohere.com/v2/embed` (overridable via `BaseURL`)
- Auth: `Authorization: Bearer <APIKey>` header
- Default model: `embed-english-v3.0`
- Request body:
  ```json
  {
    "texts": ["<text>"],
    "model": "<model>",
    "input_type": "search_document",
    "embedding_types": ["float"]
  }
  ```
- Response: extract `embeddings.float[0]` (`[]float64` → convert to `[]float32`)
- Set `Embedding.Provider` to `"cohere"`

### Shared implementation requirements for all providers

- `NewEmbedder` returns an error (does not panic) if `cfg` is nil
- Call `logging.EnsureLogger(logger)` at the top of `NewEmbedder` to guard against nil logger
- Call `cfg.ValidateWithContext(ctx)` and propagate any error
- Use `cfg.DefaultModel` as the default; use `input.Model` when non-empty
- Wrap each `GenerateEmbedding` call in a span via `e.tracer.StartSpan(ctx)`
- On error, call both `tracing.AttachErrorToSpan(span, description, err)` and `e.logger.Error(description, err)`
- Set `Embedding.GeneratedAt = time.Now()` after a successful response
- Set `Embedding.Dimensions = len(vector)` before returning
- Respect `cfg.Timeout` via `http.Client.Timeout`
- Use only `net/http` and `encoding/json` — do not introduce new dependencies

---

## Mock (`embeddings/mock/mock.go`)

Uses `testify/mock.Mock`, consistent with `llm/mock`:

```go
package mock

// Embedder is a mock embeddings.Embedder for use in tests.
type Embedder struct {
    mock.Mock
}

// GenerateEmbedding satisfies the embeddings.Embedder interface.
func (m *Embedder) GenerateEmbedding(ctx context.Context, input *embeddings.Input) (*embeddings.Embedding, error) {
    args := m.Called(ctx, input)
    return args.Get(0).(*embeddings.Embedding), args.Error(1)
}
```

---

## Testing Requirements

- Every provider must have a `_test.go` file
- Use `httptest.NewServer` to mock HTTP responses — no real network calls in tests
- Use `logging.NewNoopLogger()` and `tracing.NewTracerForTest("test")` when constructing
  providers in tests
- Test the happy path: valid response → correctly populated `*Embedding`
  - Verify `SourceText`, `Model`, `Provider`, `Dimensions`, and `GeneratedAt` are all set
- Test error cases: nil config, missing APIKey (where required), non-200 response, malformed JSON
- Test that `input.Model` overrides `cfg.DefaultModel` when non-empty
- The noop must have a test (trivial but required for blanket compliance)
- The mock must have a test confirming it delegates to `mock.Called`

---

## What to Avoid

- No `init()` functions
- No package-level variables holding provider instances
- No global default Embedder
- No `GenerateEmbeddings` batch method — callers manage their own loops
- Do not put a `Config` in the root `embeddings` package — config lives in
  `embeddings/config` with package name `embeddingscfg`
- Do not use a hand-rolled function-field mock — use `testify/mock.Mock` as in `llm/mock`
- Do not introduce dependencies beyond those already present in the module's `go.mod`
