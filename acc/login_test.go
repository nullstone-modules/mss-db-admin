package acc

import (
	"github.com/nullstone-modules/mss-db-admin/sqlserver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestLogin(t *testing.T) {
	if os.Getenv("ACC") != "1" {
		t.Skip("Set ACC=1 to run e2e tests")
	}

	store := createStore(t)

	login := sqlserver.Login{
		Name:     "login-test-user",
		Password: "L0gin!TestPass",
	}
	result, err := store.Logins.Create(login)
	require.NoError(t, err, "create login")
	require.NotNil(t, result)
	assert.Equal(t, "login-test-user", result.Name)

	found, err := store.Logins.Read("login-test-user")
	require.NoError(t, err, "read login")
	require.NotNil(t, found)
	assert.Equal(t, "login-test-user", found.Name)

	updated, err := store.Logins.Update("login-test-user", sqlserver.Login{
		Name:     "login-test-user",
		Password: "NewL0gin!Pass",
	})
	require.NoError(t, err, "update login password")
	require.NotNil(t, updated)
}
