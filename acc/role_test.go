package acc

import (
	"github.com/nullstone-modules/mss-db-admin/sqlserver"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestRole(t *testing.T) {
	if os.Getenv("ACC") != "1" {
		t.Skip("Set ACC=1 to run e2e tests")
	}

	db := createDb(t)
	defer db.Close()

	role := sqlserver.Role{
		Name:     "role-test-user",
		Password: "role-test-password",
	}
	require.NoError(t, role.Create(db), "unexpected error")

	find := &sqlserver.Role{Name: "role-test-user"}
	require.NoError(t, find.Read(db), "read user")
}
