package migrate

type Migrator interface {
	Up() error
	Down() error
	Version() (uint, bool, error)
	Close() (source error, database error)
	MigrateTo(version uint) error
}

type MigrationHandler interface {
	Migrate(source, dsn string) (Migrator, error)
}
