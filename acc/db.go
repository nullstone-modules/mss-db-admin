package acc

import (
	"database/sql"
	"testing"

	_ "github.com/microsoft/go-mssqldb"
)

func createDb(t *testing.T) *sql.DB {
	connUrl := "sqlserver://sa:YourStr0ng!Pass@localhost:1433?database=master"
	db, err := sql.Open("sqlserver", connUrl)
	if err != nil {
		t.Fatalf("error connecting to sql server: %s", err)
	}
	return db
}
