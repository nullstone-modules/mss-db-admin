package sqlserver

import (
	"database/sql"
	"fmt"
	"log"
)

type Role struct {
	Name     string
	Password string
}

func (r Role) Ensure(db *sql.DB) error {
	if exists, err := r.Exists(db); exists {
		log.Printf("Login %q already exists\n", r.Name)
		return nil
	} else if err != nil {
		return fmt.Errorf("error checking for login %q: %w", r.Name, err)
	}
	return r.Create(db)
}

func (r Role) Create(db *sql.DB) error {
	log.Printf("Creating login %q\n", r.Name)
	sq := fmt.Sprintf("CREATE LOGIN %s WITH PASSWORD = %s", QuoteIdentifier(r.Name), QuoteLiteral(r.Password))
	if _, err := db.Exec(sq); err != nil {
		return fmt.Errorf("error creating login %q: %w", r.Name, err)
	}
	return nil
}

func (r Role) Exists(db *sql.DB) (bool, error) {
	var name string
	err := db.QueryRow(`SELECT name FROM sys.server_principals WHERE name = @p1`, r.Name).Scan(&name)
	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func (r *Role) Read(db *sql.DB) error {
	var name string
	err := db.QueryRow(`SELECT name FROM sys.server_principals WHERE name = @p1`, r.Name).Scan(&name)
	if err != nil {
		return err
	}
	r.Name = name
	return nil
}
