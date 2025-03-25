package migrate

import (
	"fmt"

	"github.com/flexer2006/case-back-restaurant-go/common"
	"github.com/golang-migrate/migrate/v4"
)

type Adapter struct {
	m *migrate.Migrate
}

func NewAdapter(m *migrate.Migrate) *Adapter {
	return &Adapter{m: m}
}

func (a *Adapter) Up() error {
	err := a.m.Up()
	if err != nil {
		return fmt.Errorf("%s: %w", common.ErrMigrateApply, err)
	}

	return nil
}

func (a *Adapter) Down() error {
	err := a.m.Down()
	if err != nil {
		return fmt.Errorf("%s: %w", common.ErrMigrateDown, err)
	}

	return nil
}

func (a *Adapter) Version() (uint, bool, error) {
	version, dirty, err := a.m.Version()
	if err != nil {
		return 0, false, fmt.Errorf("%s: %w", common.ErrMigrateVersion, err)
	}

	return version, dirty, nil
}

func (a *Adapter) Close() (source error, database error) {
	src, db := a.m.Close()

	if src != nil {
		src = fmt.Errorf("%s: %w", common.ErrCloseMigrationSource, src)
	}

	if db != nil {
		db = fmt.Errorf("%s: %w", common.ErrCloseDBConnection, db)
	}

	return src, db
}

func (a *Adapter) MigrateTo(version uint) error {
	err := a.m.Migrate(version)
	if err != nil {
		return fmt.Errorf("%s to version %d: %w", common.ErrMigrateToVersion, version, err)
	}

	return nil
}

type Handler struct{}

type HandlerFactoryFunc func() MigrationHandler

func newHandlerImpl() MigrationHandler {
	return &Handler{}
}

var NewHandlerFunc HandlerFactoryFunc = newHandlerImpl

func NewHandler() MigrationHandler {
	return NewHandlerFunc()
}

func (h *Handler) Migrate(source, dsn string) (Migrator, error) {
	m, err := migrate.New(source, dsn)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", common.ErrMigrateInstanceCreation, err)
	}

	return NewAdapter(m), nil
}
