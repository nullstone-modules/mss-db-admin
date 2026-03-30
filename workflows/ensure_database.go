package workflows

import (
	"database/sql"
	"fmt"
	"github.com/nullstone-modules/mss-db-admin/sqlserver"
	"log"
)

func EnsureDatabase(db *sql.DB, newDatabase sqlserver.Database) error {
	log.Printf("ensuring database %q\n", newDatabase.Name)

	return newDatabase.Ensure(db)
}

// EnsureDatabaseWithInfo is available if callers need DbInfo for other purposes.
func EnsureDatabaseWithInfo(db *sql.DB, newDatabase sqlserver.Database) (*sqlserver.DbInfo, error) {
	log.Printf("ensuring database %q\n", newDatabase.Name)

	dbInfo, err := sqlserver.CalcDbConnectionInfo(db)
	if err != nil {
		return nil, fmt.Errorf("error introspecting sql server instance: %w", err)
	}

	if err := newDatabase.Ensure(db); err != nil {
		return nil, err
	}
	return dbInfo, nil
}
