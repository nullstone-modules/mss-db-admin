package sqlserver

import (
	"database/sql"
	"fmt"
	"log"
)

type Database struct {
	Name      string
	Collation string
}

func (d Database) Create(db *sql.DB) error {
	sq := fmt.Sprintf("CREATE DATABASE %s", QuoteIdentifier(d.Name))
	if d.Collation != "" {
		sq += fmt.Sprintf(" COLLATE %s", d.Collation)
	}

	log.Printf("Creating database %q\n", d.Name)
	if _, err := db.Exec(sq); err != nil {
		return fmt.Errorf("error creating database %q: %w", d.Name, err)
	}
	return nil
}

func (d Database) Ensure(db *sql.DB) error {
	if exists, err := d.Exists(db); exists {
		log.Printf("database %q already exists\n", d.Name)
		return nil
	} else if err != nil {
		return fmt.Errorf("error checking for database %q: %w", d.Name, err)
	}
	return d.Create(db)
}

func (d Database) Exists(db *sql.DB) (bool, error) {
	var name string
	err := db.QueryRow(`SELECT name FROM sys.databases WHERE name = @p1`, d.Name).Scan(&name)
	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return true, nil
}

func (d *Database) Read(db *sql.DB) error {
	var name string
	err := db.QueryRow(`SELECT name FROM sys.databases WHERE name = @p1`, d.Name).Scan(&name)
	if err != nil {
		return err
	}
	d.Name = name
	return nil
}
