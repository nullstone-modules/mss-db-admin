package sqlserver

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
)

type Login struct {
	Name     string `json:"name"`
	Password string `json:"password,omitempty"`
}

func (l Login) Key() string {
	return l.Name
}

type Logins struct {
	DbOpener DbOpener
}

func (r *Logins) Create(obj Login) (*Login, error) {
	db, err := r.DbOpener.OpenDatabase("")
	if err != nil {
		return nil, err
	}

	log.Printf("Creating login %q\n", obj.Name)
	sq := fmt.Sprintf("CREATE LOGIN %s WITH PASSWORD = %s", QuoteIdentifier(obj.Name), QuoteLiteral(obj.Password))
	if _, err := db.Exec(sq); err != nil {
		return nil, fmt.Errorf("error creating login %q: %w", obj.Name, err)
	}
	return r.Read(obj.Name)
}

func (r *Logins) Read(key string) (*Login, error) {
	db, err := r.DbOpener.OpenDatabase("")
	if err != nil {
		return nil, err
	}

	var name string
	row := db.QueryRow(`SELECT name FROM sys.server_principals WHERE name = @p1`, key)
	if err := row.Scan(&name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &Login{Name: name}, nil
}

func (r *Logins) Update(key string, obj Login) (*Login, error) {
	if obj.Password == "" {
		return r.Read(key)
	}

	db, err := r.DbOpener.OpenDatabase("")
	if err != nil {
		return nil, err
	}

	log.Printf("Updating login %q password\n", key)
	sq := fmt.Sprintf("ALTER LOGIN %s WITH PASSWORD = %s", QuoteIdentifier(key), QuoteLiteral(obj.Password))
	if _, err := db.Exec(sq); err != nil {
		return nil, fmt.Errorf("error updating login %q: %w", key, err)
	}
	return r.Read(key)
}

func (r *Logins) Drop(key string) (bool, error) {
	return true, nil
}
