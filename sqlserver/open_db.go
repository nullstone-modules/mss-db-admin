package sqlserver

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"time"
)

var (
	dbOpenConnTimeout = 3 * time.Second
)

func OpenDatabase(connUrl string, databaseName string) (*sql.DB, error) {
	if databaseName != "" {
		u, err := url.Parse(connUrl)
		if err != nil {
			return nil, fmt.Errorf("invalid connection url: %w", err)
		}
		q := u.Query()
		q.Set("database", databaseName)
		u.RawQuery = q.Encode()
		u.Path = ""
		log.Printf("Opening sql server connection to %s\n", u.Host)
		connUrl = u.String()
	}

	db, err := sql.Open("sqlserver", connUrl)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), dbOpenConnTimeout)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("error establishing connection to sql server: %w", err)
	}
	log.Println("SQL Server connection established")
	return db, nil
}
