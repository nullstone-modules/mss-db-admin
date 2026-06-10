package sqlserver

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
)

// requiredRoles is the set of built-in database roles a login is granted for full access.
// Update grants every role in this list and Read verifies every role in this list, so the
// granted set and the verified set can never drift apart.
var requiredRoles = []string{"db_datareader", "db_datawriter", "db_ddladmin"}

type DatabaseAccess struct {
	Database string `json:"database"`
	Login    string `json:"login"`
}

func (a DatabaseAccess) Key() DatabaseAccessKey {
	return DatabaseAccessKey{
		Database: a.Database,
		Login:    a.Login,
	}
}

type DatabaseAccessKey struct {
	Database string
	Login    string
}

type DatabaseAccesses struct {
	DbOpener DbOpener
}

func (r *DatabaseAccesses) Create(obj DatabaseAccess) (*DatabaseAccess, error) {
	return r.Update(obj.Key(), obj)
}

// Read verifies the complete access state for a login on a database: the user must exist,
// be a member of every required role, and hold EXECUTE on the dbo schema. If any piece is
// missing (e.g. someone manually revoked a role), Read reports the access as not provisioned
// (nil, nil) so Terraform detects drift and repairs it via the idempotent Create/Update.
func (r *DatabaseAccesses) Read(key DatabaseAccessKey) (*DatabaseAccess, error) {
	db, err := r.DbOpener.OpenDatabase(key.Database)
	if err != nil {
		return nil, err
	}

	// 1. The database user must exist.
	var name string
	row := db.QueryRow(`SELECT name FROM sys.database_principals WHERE name = @p1 AND type IN ('S', 'U')`, key.Login)
	if err := row.Scan(&name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	// 2. All required role memberships must be present.
	actualRoles, err := r.readRoles(db, key.Login)
	if err != nil {
		return nil, err
	}
	for _, want := range requiredRoles {
		if !actualRoles[want] {
			log.Printf("[Read] login %q missing role %q on database %q -> reporting drift\n", key.Login, want, key.Database)
			return nil, nil
		}
	}

	// 3. EXECUTE on schema::dbo must be granted.
	if ok, err := r.hasExecuteOnDbo(db, key.Login); err != nil {
		return nil, err
	} else if !ok {
		log.Printf("[Read] login %q missing EXECUTE on dbo in database %q -> reporting drift\n", key.Login, key.Database)
		return nil, nil
	}

	return &DatabaseAccess{Database: key.Database, Login: name}, nil
}

// readRoles returns the set of database role names the login is currently a member of.
func (r *DatabaseAccesses) readRoles(db *sql.DB, login string) (map[string]bool, error) {
	rows, err := db.Query(`
		SELECT r.name
		FROM sys.database_role_members drm
		JOIN sys.database_principals r ON r.principal_id = drm.role_principal_id
		JOIN sys.database_principals m ON m.principal_id = drm.member_principal_id
		WHERE m.name = @p1`, login)
	if err != nil {
		return nil, fmt.Errorf("error reading role memberships for %q: %w", login, err)
	}
	defer rows.Close()

	out := map[string]bool{}
	for rows.Next() {
		var role string
		if err := rows.Scan(&role); err != nil {
			return nil, err
		}
		out[role] = true
	}
	return out, rows.Err()
}

// hasExecuteOnDbo reports whether the login has been granted EXECUTE on the dbo schema.
func (r *DatabaseAccesses) hasExecuteOnDbo(db *sql.DB, login string) (bool, error) {
	var count int
	row := db.QueryRow(`
		SELECT COUNT(*)
		FROM sys.database_permissions p
		JOIN sys.database_principals u ON u.principal_id = p.grantee_principal_id
		WHERE u.name = @p1
		  AND p.permission_name = 'EXECUTE'
		  AND p.state = 'G'
		  AND p.class = 3              -- 3 = schema
		  AND p.major_id = SCHEMA_ID('dbo')`, login)
	if err := row.Scan(&count); err != nil {
		return false, fmt.Errorf("error reading EXECUTE grant for %q: %w", login, err)
	}
	return count > 0, nil
}

func (r *DatabaseAccesses) Update(key DatabaseAccessKey, obj DatabaseAccess) (*DatabaseAccess, error) {
	db, err := r.DbOpener.OpenDatabase(obj.Database)
	if err != nil {
		return nil, err
	}

	log.Printf("Granting login %q access to database %q\n", obj.Login, obj.Database)

	// Create database user mapped to the server login
	createUserSql := fmt.Sprintf(
		"IF NOT EXISTS (SELECT * FROM sys.database_principals WHERE name = %s) CREATE USER %s FOR LOGIN %s",
		QuoteLiteral(obj.Login),
		QuoteIdentifier(obj.Login),
		QuoteIdentifier(obj.Login),
	)
	if _, err := db.Exec(createUserSql); err != nil {
		return nil, fmt.Errorf("error creating database user %q: %w", obj.Login, err)
	}

	// Add user to built-in database roles for full access (idempotent: a no-op if already a member)
	for _, role := range requiredRoles {
		sq := fmt.Sprintf("ALTER ROLE %s ADD MEMBER %s", QuoteIdentifier(role), QuoteIdentifier(obj.Login))
		if _, err := db.Exec(sq); err != nil {
			return nil, fmt.Errorf("error adding %q to role %q: %w", obj.Login, role, err)
		}
	}

	// Grant execute on dbo schema for stored procedures and functions
	sq := fmt.Sprintf("GRANT EXECUTE ON SCHEMA::dbo TO %s", QuoteIdentifier(obj.Login))
	if _, err := db.Exec(sq); err != nil {
		return nil, fmt.Errorf("error granting execute on dbo to %q: %w", obj.Login, err)
	}

	return &obj, nil
}

func (r *DatabaseAccesses) Drop(key DatabaseAccessKey) (bool, error) {
	db, err := r.DbOpener.OpenDatabase(key.Database)
	if err != nil {
		return false, err
	}

	log.Printf("Revoking login %q access from database %q\n", key.Login, key.Database)

	// Remove from database roles
	for _, role := range requiredRoles {
		sq := fmt.Sprintf("ALTER ROLE %s DROP MEMBER %s", QuoteIdentifier(role), QuoteIdentifier(key.Login))
		if _, err := db.Exec(sq); err != nil {
			return false, fmt.Errorf("error removing %q from role %q: %w", key.Login, role, err)
		}
	}

	// Drop the database user
	sq := fmt.Sprintf("DROP USER %s", QuoteIdentifier(key.Login))
	if _, err := db.Exec(sq); err != nil {
		return false, fmt.Errorf("error dropping database user %q: %w", key.Login, err)
	}

	return true, nil
}
