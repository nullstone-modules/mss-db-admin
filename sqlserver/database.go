package sqlserver

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
)

type Database struct {
	Name      string `json:"name"`
	Collation string `json:"collation,omitempty"`
}

func (d Database) Key() string {
	return d.Name
}

type Databases struct {
	DbOpener DbOpener
}

func (d *Databases) Create(obj Database) (*Database, error) {
	db, err := d.DbOpener.OpenDatabase("")
	if err != nil {
		return nil, err
	}

	sq := fmt.Sprintf("CREATE DATABASE %s", QuoteIdentifier(obj.Name))
	if obj.Collation != "" {
		sq += fmt.Sprintf(" COLLATE %s", obj.Collation)
	}

	log.Printf("Creating database %q\n", obj.Name)
	if _, err := db.Exec(sq); err != nil {
		return nil, fmt.Errorf("error creating database %q: %w", obj.Name, err)
	}
	return d.Read(obj.Name)
}

func (d *Databases) Read(key string) (*Database, error) {
	db, err := d.DbOpener.OpenDatabase("")
	if err != nil {
		return nil, err
	}

	var name, collation string
	row := db.QueryRow(`SELECT name, collation_name FROM sys.databases WHERE name = @p1`, key)
	if err := row.Scan(&name, &collation); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &Database{Name: name, Collation: collation}, nil
}

func (d *Databases) Update(key string, obj Database) (*Database, error) { return d.Read(key) }
func (d *Databases) Drop(key string) (bool, error)                      { return true, nil }
