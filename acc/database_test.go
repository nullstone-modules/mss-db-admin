package acc

import (
	"os"
	"testing"

	"github.com/nullstone-modules/mss-db-admin/sqlserver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatabase(t *testing.T) {
	if os.Getenv("ACC") != "1" {
		t.Skip("Set ACC=1 to run e2e tests")
	}

	store := createStore(t)

	database := sqlserver.Database{Name: "database-test-db"}
	result, err := store.Databases.Create(database)
	require.NoError(t, err, "create database")
	require.NotNil(t, result)
	assert.Equal(t, "database-test-db", result.Name)

	found, err := store.Databases.Read("database-test-db")
	require.NoError(t, err, "read database")
	require.NotNil(t, found)
	assert.Equal(t, "database-test-db", found.Name)
}
