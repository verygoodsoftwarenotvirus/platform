package healthcheck

import (
	"context"

	"github.com/verygoodsoftwarenotvirus/platform/v2/database"
	"github.com/verygoodsoftwarenotvirus/platform/v2/errors"
)

// DatabaseReadyChecker checks if a database client is ready.
type DatabaseReadyChecker interface {
	IsReady(ctx context.Context) bool
}

// NewDatabaseChecker returns a Checker that uses the given client's IsReady method.
func NewDatabaseChecker(name string, client DatabaseReadyChecker) Checker {
	return &databaseChecker{name: name, client: client}
}

type databaseChecker struct {
	client DatabaseReadyChecker
	name   string
}

func (d *databaseChecker) Name() string {
	return d.name
}

func (d *databaseChecker) Check(ctx context.Context) error {
	if d.client == nil {
		return errors.New("database client is nil")
	}
	if !d.client.IsReady(ctx) {
		return database.ErrDatabaseNotReady
	}
	return nil
}
