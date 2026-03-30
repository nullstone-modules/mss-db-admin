package sqlserver

import (
	"database/sql"
	"fmt"
	"log"
)

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

func (r *DatabaseAccesses) Read(key DatabaseAccessKey) (*DatabaseAccess, error) {
	db, err := r.DbOpener.OpenDatabase(key.Database)
	if err != nil {
		return nil, err
	}

	var name string
	row := db.QueryRow(`SELECT name FROM sys.database_principals WHERE name = @p1 AND type IN ('S', 'U')`, key.Login)
	if err := row.Scan(&name); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &DatabaseAccess{Database: key.Database, Login: name}, nil
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

	// Add user to built-in database roles for full access
	roles := []string{"db_datareader", "db_datawriter", "db_ddladmin"}
	for _, role := range roles {
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
	roles := []string{"db_datareader", "db_datawriter", "db_ddladmin"}
	for _, role := range roles {
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
