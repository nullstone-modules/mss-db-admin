package workflows

import (
	"database/sql"
	"fmt"
	"github.com/lib/pq"
	"github.com/nullstone-modules/mss-db-admin/sqlserver"
	"log"
	"strings"
)

func GrantDbAccess(db *sql.DB, appDb *sql.DB, user sqlserver.Role, database sqlserver.Database) error {
	log.Printf("Granting user %q db access to %q\n", user.Name, database.Name)
	if err := grantRole(db, user, database); err != nil {
		return err
	}
	return grantAllPrivileges(appDb, user, database)
}

func grantRole(db *sql.DB, user sqlserver.Role, database sqlserver.Database) error {
	if err := database.Read(db); err != nil {
		return fmt.Errorf("unable to read database %q: %w", database.Name, err)
	}

	newRoleGrant := sqlserver.RoleGrant{
		Member: user.Name,
		Target: database.Owner,
	}
	if err := newRoleGrant.Ensure(db); err != nil {
		return fmt.Errorf("error granting role grant for %q to %q: %w", newRoleGrant.Member, newRoleGrant.Target, err)
	}
	return nil
}

func grantAllPrivileges(db *sql.DB, user sqlserver.Role, database sqlserver.Database) error {
	sq := strings.Join([]string{
		// CREATE | USAGE
		fmt.Sprintf(`GRANT ALL PRIVILEGES ON SCHEMA public TO %s;`, pq.QuoteIdentifier(user.Name)),
		// CREATE | CONNECT | TEMPORARY | TEMP
		fmt.Sprintf(`GRANT ALL PRIVILEGES ON DATABASE %s TO %s;`, pq.QuoteIdentifier(database.Name), pq.QuoteIdentifier(user.Name)),
	}, " ")

	if _, err := db.Exec(sq); err != nil {
		return fmt.Errorf("error granting privileges: %w", err)
	}
	return nil
}
