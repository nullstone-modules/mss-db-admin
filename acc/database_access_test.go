package acc

import (
	"github.com/nullstone-modules/mss-db-admin/sqlserver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestDatabaseAccess(t *testing.T) {
	if os.Getenv("ACC") != "1" {
		t.Skip("Set ACC=1 to run e2e tests")
	}

	store := createStore(t)

	// Create database
	_, err := store.Databases.Create(sqlserver.Database{Name: "access-test-db"})
	require.NoError(t, err, "create database")

	// Create login
	_, err = store.Logins.Create(sqlserver.Login{
		Name:     "access-test-user",
		Password: "Acc3ss!TestPass",
	})
	require.NoError(t, err, "create login")

	// Grant access
	access := sqlserver.DatabaseAccess{
		Database: "access-test-db",
		Login:    "access-test-user",
	}
	result, err := store.DatabaseAccess.Create(access)
	require.NoError(t, err, "grant database access")
	require.NotNil(t, result)
	assert.Equal(t, "access-test-db", result.Database)
	assert.Equal(t, "access-test-user", result.Login)

	// Read access
	found, err := store.DatabaseAccess.Read(sqlserver.DatabaseAccessKey{
		Database: "access-test-db",
		Login:    "access-test-user",
	})
	require.NoError(t, err, "read database access")
	require.NotNil(t, found)

	// Revoke access
	ok, err := store.DatabaseAccess.Drop(sqlserver.DatabaseAccessKey{
		Database: "access-test-db",
		Login:    "access-test-user",
	})
	require.NoError(t, err, "revoke database access")
	assert.True(t, ok)
}
