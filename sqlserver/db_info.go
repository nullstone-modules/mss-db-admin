package sqlserver

import (
	"database/sql"
	"fmt"
)

type DbInfo struct {
	IsSuperuser bool
	CurrentUser string
}

func CalcDbConnectionInfo(db *sql.DB) (*DbInfo, error) {
	dci := &DbInfo{}

	var isSysadmin int
	if err := db.QueryRow(`SELECT IS_SRVROLEMEMBER('sysadmin')`).Scan(&isSysadmin); err != nil {
		return nil, fmt.Errorf("could not check if current user is sysadmin: %w", err)
	}
	dci.IsSuperuser = isSysadmin == 1

	var err error
	dci.CurrentUser, err = getCurrentUser(db)
	if err != nil {
		return nil, fmt.Errorf("error getting current user: %w", err)
	}

	return dci, nil
}

func getCurrentUser(db *sql.DB) (string, error) {
	var currentUser string
	err := db.QueryRow("SELECT SUSER_NAME()").Scan(&currentUser)
	switch {
	case err == sql.ErrNoRows:
		return "", fmt.Errorf("SELECT SUSER_NAME() returned no rows")
	case err != nil:
		return "", fmt.Errorf("error looking up current user: %w", err)
	}
	return currentUser, nil
}
