package acc

import (
	"github.com/nullstone-modules/mss-db-admin/sqlserver"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestDatabase(t *testing.T) {
	if os.Getenv("ACC") != "1" {
		t.Skip("Set ACC=1 to run e2e tests")
	}

	db := createDb(t)
	defer db.Close()

	database := sqlserver.Database{Name: "database-test-database"}
	require.NoError(t, database.Create(db), "unexpected error")

	find := &sqlserver.Database{Name: "database-test-database"}
	require.NoError(t, find.Read(db), "read database")
	require.Equal(t, "database-test-database", find.Name)
}
