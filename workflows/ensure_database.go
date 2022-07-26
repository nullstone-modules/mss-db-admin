package workflows

import (
	"database/sql"
	"fmt"
	"github.com/nullstone-modules/mss-db-admin/sqlserver"
	"log"
)

func EnsureDatabase(db *sql.DB, newDatabase sqlserver.Database) error {
	log.Printf("ensuring database %q\n", newDatabase.Name)

	dbInfo, err := sqlserver.CalcDbConnectionInfo(db)
	if err != nil {
		return fmt.Errorf("error introspecting postgres cluster: %w", err)
	}

	// Create a role with the same name as the database to give ownership
	if err := (sqlserver.Role{Name: newDatabase.Name}).Ensure(db); err != nil {
		return fmt.Errorf("error ensuring database owner role: %w", err)
	}
	return newDatabase.Ensure(db, *dbInfo)
}
