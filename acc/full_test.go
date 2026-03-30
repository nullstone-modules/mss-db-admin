package acc

import (
	"fmt"
	_ "github.com/microsoft/go-mssqldb"
	"github.com/nullstone-modules/mss-db-admin/sqlserver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/url"
	"os"
	"strings"
	"testing"
)

// TestFull tests the entire CRUD workflow:
// create database -> create login -> grant access -> connect as user -> create table -> insert -> query
func TestFull(t *testing.T) {
	if os.Getenv("ACC") != "1" {
		t.Skip("Set ACC=1 to run e2e tests")
	}

	store := createStore(t)

	dbName := "full-test-db"
	loginName := "full-test-user"
	loginPass := "Full!Test0Pass"

	// Create the database
	_, err := store.Databases.Create(sqlserver.Database{Name: dbName})
	require.NoError(t, err, "create database")

	// Create the login
	_, err = store.Logins.Create(sqlserver.Login{Name: loginName, Password: loginPass})
	require.NoError(t, err, "create login")

	// Grant access
	_, err = store.DatabaseAccess.Create(sqlserver.DatabaseAccess{Database: dbName, Login: loginName})
	require.NoError(t, err, "grant database access")

	// Connect as the new user to the app database
	userConnUrl := fmt.Sprintf("sqlserver://%s:%s@localhost:1433?database=%s",
		url.PathEscape(loginName),
		url.PathEscape(loginPass),
		url.PathEscape(dbName),
	)
	userDb, err := sqlserver.OpenDatabase(userConnUrl, "")
	require.NoError(t, err, "connecting as app user")
	defer userDb.Close()

	// Create a table
	_, err = userDb.Exec("CREATE TABLE todos ( id INT IDENTITY(1,1) NOT NULL, name varchar(255) );")
	require.NoError(t, err, "create table")

	// Insert records
	sq := strings.Join([]string{
		`INSERT INTO todos (name) VALUES ('item1');`,
		`INSERT INTO todos (name) VALUES ('item2');`,
		`INSERT INTO todos (name) VALUES ('item3');`,
	}, " ")
	_, err = userDb.Exec(sq)
	require.NoError(t, err, "insert todos")

	// Query records
	results := make([]string, 0)
	rows, err := userDb.Query(`SELECT id, name FROM todos ORDER BY id`)
	require.NoError(t, err, "query todos")
	defer rows.Close()
	for rows.Next() {
		var id int
		var name string
		require.NoError(t, rows.Scan(&id, &name), "scan record")
		results = append(results, name)
	}
	assert.Equal(t, []string{"item1", "item2", "item3"}, results)
}
