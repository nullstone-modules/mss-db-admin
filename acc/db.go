package acc

import (
	"github.com/nullstone-modules/mss-db-admin/sqlserver"
	"testing"

	_ "github.com/microsoft/go-mssqldb"
)

const connUrl = "sqlserver://sa:YourStr0ng!Pass@localhost:1433?database=master"

func createStore(t *testing.T) *sqlserver.Store {
	store := sqlserver.NewStore(connUrl)
	t.Cleanup(func() { store.Close() })
	return store
}
