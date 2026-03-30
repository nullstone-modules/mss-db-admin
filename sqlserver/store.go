package sqlserver

import (
	"database/sql"
	"sync"
)

type Store struct {
	Databases      *Databases
	Logins         *Logins
	DatabaseAccess *DatabaseAccesses

	connUrl       string
	connsByDbName map[string]*sql.DB
	sync.Mutex
}

type DbOpener interface {
	OpenDatabase(dbName string) (*sql.DB, error)
}

func NewStore(connUrl string) *Store {
	store := &Store{connUrl: connUrl, connsByDbName: map[string]*sql.DB{}}
	store.Databases = &Databases{DbOpener: store}
	store.Logins = &Logins{DbOpener: store}
	store.DatabaseAccess = &DatabaseAccesses{DbOpener: store}
	return store
}

func (s *Store) ConnectionUrl() string {
	return s.connUrl
}

func (s *Store) Close() {
	s.Lock()
	defer s.Unlock()

	for name, conn := range s.connsByDbName {
		conn.Close()
		delete(s.connsByDbName, name)
	}
}

func (s *Store) OpenDatabase(dbName string) (*sql.DB, error) {
	s.Lock()
	defer s.Unlock()

	existing, ok := s.connsByDbName[dbName]
	if ok {
		return existing, nil
	}

	db, err := OpenDatabase(s.connUrl, dbName)
	if err != nil {
		return nil, err
	}
	s.connsByDbName[dbName] = db
	return db, nil
}
