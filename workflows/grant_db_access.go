package workflows

import (
	"database/sql"
	"fmt"
	"github.com/nullstone-modules/mss-db-admin/sqlserver"
	"log"
)

func GrantDbAccess(db *sql.DB, appDb *sql.DB, user sqlserver.Role, database sqlserver.Database) error {
	log.Printf("Granting user %q db access to %q\n", user.Name, database.Name)
	return grantAllPrivileges(appDb, user)
}

func grantAllPrivileges(appDb *sql.DB, user sqlserver.Role) error {
	// Create a database user mapped to the server login
	createUserSql := fmt.Sprintf(
		"IF NOT EXISTS (SELECT * FROM sys.database_principals WHERE name = %s) CREATE USER %s FOR LOGIN %s",
		sqlserver.QuoteLiteral(user.Name),
		sqlserver.QuoteIdentifier(user.Name),
		sqlserver.QuoteIdentifier(user.Name),
	)
	if _, err := appDb.Exec(createUserSql); err != nil {
		return fmt.Errorf("error creating database user %q: %w", user.Name, err)
	}

	// Add user to built-in database roles for full access
	roles := []string{"db_datareader", "db_datawriter", "db_ddladmin"}
	for _, role := range roles {
		sq := fmt.Sprintf("ALTER ROLE %s ADD MEMBER %s", sqlserver.QuoteIdentifier(role), sqlserver.QuoteIdentifier(user.Name))
		if _, err := appDb.Exec(sq); err != nil {
			return fmt.Errorf("error adding %q to role %q: %w", user.Name, role, err)
		}
	}

	// Grant execute on the dbo schema so the user can call stored procedures and functions
	sq := fmt.Sprintf("GRANT EXECUTE ON SCHEMA::dbo TO %s", sqlserver.QuoteIdentifier(user.Name))
	if _, err := appDb.Exec(sq); err != nil {
		return fmt.Errorf("error granting execute on dbo to %q: %w", user.Name, err)
	}

	return nil
}
