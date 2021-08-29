package store

import (
	"os"
	"time"

	"github.com/philippgille/gokv"
	"github.com/philippgille/gokv/mysql"
	"github.com/philippgille/gokv/syncmap"
	"github.com/pkg/errors"
)

type Store struct {
	inMemoryStore gokv.Store
}

type Item struct {
	ID        int64
	WriteTime time.Time
}

func New() (*Store, error) {
	// An in-memory store that is used if no MYSQL_DB is provided.
	// This should only be used in development mode.
	s := &Store{syncmap.NewStore(syncmap.DefaultOptions)}

	client, err := s.connect()
	defer client.Close()
	if err != nil {
		return nil, errors.Wrap(err, "could not connect to mysql server")
	}
	return s, nil
}

func (s *Store) connect() (gokv.Store, error) {
	if m := os.Getenv("MYSQL_DB"); m != "" {
		return mysql.NewClient(mysql.Options{DataSourceName: m})
	}
	return s.inMemoryStore, nil
}

func (s *Store) Get(key string) (*Item, error) {
	client, err := s.connect()
	if err != nil {
		return nil, errors.Wrap(err, "could not connect to database")
	}
	defer client.Close()

	i := new(Item)
	found, err := client.Get(key, i)
	if err != nil {
		return nil, errors.Wrap(err, "could not get item")
	}
	if !found {
		return nil, nil
	}
	return i, nil
}

func (s *Store) Put(key string, id int64) (*Item, error) {
	client, err := s.connect()
	if err != nil {
		return nil, errors.Wrap(err, "could not connect to database")
	}
	defer client.Close()

	i := Item{
		ID:        id,
		WriteTime: time.Now(),
	}
	return &i, client.Set(key, i)
}
