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

// TestDatabaseAccess_RepairsDrift verifies that when access is manually damaged out-of-band
// (a role membership is dropped), Read reports the access as not provisioned so Terraform
// detects drift, and that re-running Create restores full access.
func TestDatabaseAccess_RepairsDrift(t *testing.T) {
	if os.Getenv("ACC") != "1" {
		t.Skip("Set ACC=1 to run e2e tests")
	}

	store := createStore(t)

	dbName := "drift-test-db"
	loginName := "drift-test-user"

	_, err := store.Databases.Create(sqlserver.Database{Name: dbName})
	require.NoError(t, err, "create database")

	_, err = store.Logins.Create(sqlserver.Login{Name: loginName, Password: "Dr1ft!TestPass"})
	require.NoError(t, err, "create login")

	key := sqlserver.DatabaseAccessKey{Database: dbName, Login: loginName}
	access := sqlserver.DatabaseAccess{Database: dbName, Login: loginName}

	_, err = store.DatabaseAccess.Create(access)
	require.NoError(t, err, "grant database access")

	// Access is fully provisioned: Read finds it.
	found, err := store.DatabaseAccess.Read(key)
	require.NoError(t, err, "read after grant")
	require.NotNil(t, found, "access should be reported as provisioned after grant")

	// Damage access out-of-band: drop one role membership directly.
	db, err := store.OpenDatabase(dbName)
	require.NoError(t, err, "open database")
	_, err = db.Exec("ALTER ROLE " + sqlserver.QuoteIdentifier("db_datawriter") +
		" DROP MEMBER " + sqlserver.QuoteIdentifier(loginName))
	require.NoError(t, err, "manually drop role membership")

	// Read now reports drift (not provisioned) because a required role is missing.
	found, err = store.DatabaseAccess.Read(key)
	require.NoError(t, err, "read after damage")
	assert.Nil(t, found, "damaged access should be reported as not provisioned")

	// Re-running Create repairs the access.
	_, err = store.DatabaseAccess.Create(access)
	require.NoError(t, err, "repair database access")

	found, err = store.DatabaseAccess.Read(key)
	require.NoError(t, err, "read after repair")
	require.NotNil(t, found, "access should be reported as provisioned after repair")

	// Cleanup
	ok, err := store.DatabaseAccess.Drop(key)
	require.NoError(t, err, "revoke database access")
	assert.True(t, ok)
}
