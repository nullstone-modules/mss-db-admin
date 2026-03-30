package acc

import (
	"database/sql"
	"fmt"
	_ "github.com/microsoft/go-mssqldb"
	"github.com/nullstone-modules/mss-db-admin/sqlserver"
	"github.com/nullstone-modules/mss-db-admin/workflows"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/url"
	"os"
	"strings"
	"testing"
)

// TestFull tests the entire workflow of create-database, create-user, create-db-access
func TestFull(t *testing.T) {
	if os.Getenv("ACC") != "1" {
		t.Skip("Set ACC=1 to run e2e tests")
	}

	connUrl := "sqlserver://sa:YourStr0ng!Pass@localhost:1433?database=master"
	db, err := sql.Open("sqlserver", connUrl)
	require.NoError(t, err, "connecting to sql server")
	defer db.Close()

	newDatabase := sqlserver.Database{
		Name: "test-database",
	}
	newUser := sqlserver.Role{
		Name:     "test-user",
		Password: "Test!Passw0rd",
	}

	// Create the database
	require.NoError(t, workflows.EnsureDatabase(db, newDatabase), "ensure database")

	// Create the login
	require.NoError(t, workflows.EnsureUser(db, newUser), "ensure user")

	// Connect to the app database as admin to grant access
	appConnUrl := changeDatabase(t, connUrl, newDatabase.Name)
	appDb, err := sql.Open("sqlserver", appConnUrl)
	require.NoError(t, err, "connecting to app db as admin")
	defer appDb.Close()

	require.NoError(t, workflows.GrantDbAccess(db, appDb, newUser, newDatabase), "grant db access")

	// Connect as the new user to the app database
	userConnUrl := fmt.Sprintf("sqlserver://%s:%s@localhost:1433?database=%s",
		url.PathEscape(newUser.Name),
		url.PathEscape(newUser.Password),
		url.PathEscape(newDatabase.Name),
	)
	userDb, err := sql.Open("sqlserver", userConnUrl)
	require.NoError(t, err, "connecting as app user")
	defer userDb.Close()

	// Attempt to create schema objects
	_, err = userDb.Exec("CREATE TABLE todos ( id INT IDENTITY(1,1) NOT NULL, name varchar(255) );")
	require.NoError(t, err, "create table")

	// Attempt to insert records
	sq := strings.Join([]string{
		`INSERT INTO todos (name) VALUES ('item1');`,
		`INSERT INTO todos (name) VALUES ('item2');`,
		`INSERT INTO todos (name) VALUES ('item3');`,
	}, " ")
	_, err = userDb.Exec(sq)
	require.NoError(t, err, "insert todos")

	// Attempt to retrieve them
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

func changeDatabase(t *testing.T, connUrl string, databaseName string) string {
	u, err := url.Parse(connUrl)
	require.NoError(t, err, "parsing connection url")
	q := u.Query()
	q.Set("database", databaseName)
	u.RawQuery = q.Encode()
	u.Path = ""
	return u.String()
}
