package workflows

import (
	"database/sql"
	"github.com/nullstone-modules/mss-db-admin/sqlserver"
	"log"
)

func EnsureUser(db *sql.DB, newUser sqlserver.Role) error {
	log.Printf("ensuring user %q\n", newUser.Name)

	return newUser.Ensure(db)
}
