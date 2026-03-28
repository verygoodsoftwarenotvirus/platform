package tableaccess

import (
	"context"
	"database/sql"

	"github.com/verygoodsoftwarenotvirus/platform/v4/database"
	"github.com/verygoodsoftwarenotvirus/platform/v4/errors"
)

// ErrNotSupported is returned for operations that SQLite does not support.
// SQLite has no concept of users, roles, permissions, or multiple databases.
var ErrNotSupported = errors.New("operation not supported by SQLite")

type manager struct {
	db *sql.DB
}

func NewManager(db *sql.DB) database.Manager {
	return &manager{db: db}
}

func (m *manager) CreateUser(_ context.Context, _, _ string) error {
	return ErrNotSupported
}

func (m *manager) DeleteUser(_ context.Context, _ string) error {
	return ErrNotSupported
}

func (m *manager) CreateDatabase(_ context.Context, _, _ string) error {
	return ErrNotSupported
}

func (m *manager) DeleteDatabase(_ context.Context, _ string) error {
	return ErrNotSupported
}

func (m *manager) UserExists(_ context.Context, _ string) (bool, error) {
	return false, ErrNotSupported
}

func (m *manager) DatabaseExists(_ context.Context, _ string) (bool, error) {
	return false, ErrNotSupported
}

func (m *manager) GrantUserAccessToTable(_ context.Context, _, _, _, _ string) error {
	return ErrNotSupported
}

func (m *manager) UserCanAccessDatabase(_ context.Context, _, _ string) (bool, error) {
	return false, ErrNotSupported
}
