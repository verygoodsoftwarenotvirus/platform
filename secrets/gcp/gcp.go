package gcp

import (
	"context"
	"fmt"
	"strings"

	"github.com/verygoodsoftwarenotvirus/platform/secrets"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
)

const (
	secretVersionLatest = "latest"
	projectsPrefix      = "projects/"
)

// SecretVersionAccessor abstracts AccessSecretVersion for testability.
type SecretVersionAccessor interface {
	AccessSecretVersion(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest) (*secretmanagerpb.AccessSecretVersionResponse, error)
	Close() error
}

type gcpSecretSource struct {
	client    SecretVersionAccessor
	projectID string
}

// NewGCPSecretSource creates a SecretSource backed by GCP Secret Manager.
// If client is nil, a new client is created using Application Default Credentials.
func NewGCPSecretSource(ctx context.Context, cfg *Config, client SecretVersionAccessor) (secrets.SecretSource, error) {
	if cfg == nil {
		return nil, fmt.Errorf("gcp secret source: config is required")
	}
	if err := cfg.ValidateWithContext(ctx); err != nil {
		return nil, fmt.Errorf("gcp secret source: %w", err)
	}

	if client != nil {
		return &gcpSecretSource{
			client:    client,
			projectID: cfg.ProjectID,
		}, nil
	}

	smClient, err := secretmanager.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("gcp secret source: creating client: %w", err)
	}

	return &gcpSecretSource{
		client:    &secretManagerClientAdapter{Client: smClient},
		projectID: cfg.ProjectID,
	}, nil
}

// secretManagerClientAdapter adapts *secretmanager.Client to SecretVersionAccessor.
type secretManagerClientAdapter struct {
	*secretmanager.Client
}

func (a *secretManagerClientAdapter) AccessSecretVersion(ctx context.Context, req *secretmanagerpb.AccessSecretVersionRequest) (*secretmanagerpb.AccessSecretVersionResponse, error) {
	return a.Client.AccessSecretVersion(ctx, req)
}

func (g *gcpSecretSource) GetSecret(ctx context.Context, name string) (string, error) {
	resourceName := g.resolveName(name)
	req := &secretmanagerpb.AccessSecretVersionRequest{
		Name: resourceName,
	}

	resp, err := g.client.AccessSecretVersion(ctx, req)
	if err != nil {
		return "", fmt.Errorf("accessing secret %q: %w", name, err)
	}
	if resp.Payload == nil || resp.Payload.Data == nil {
		return "", nil
	}
	return string(resp.Payload.Data), nil
}

func (g *gcpSecretSource) Close() error {
	return g.client.Close()
}

func (g *gcpSecretSource) resolveName(name string) string {
	if strings.HasPrefix(name, projectsPrefix) {
		return name
	}
	return fmt.Sprintf("projects/%s/secrets/%s/versions/%s", g.projectID, name, secretVersionLatest)
}
