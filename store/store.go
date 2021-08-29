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
	store gokv.Store
}

type Item struct {
	ID        int64
	WriteTime time.Time
}

func New() (*Store, error) {
	if m := os.Getenv("MYSQL_DB"); m != "" {
		opts := mysql.Options{DataSourceName: m}
		client, err := mysql.NewClient(opts)
		if err != nil {
			return nil, errors.Wrap(err, "could not connect to mysql server")
		}
		return &Store{client}, nil
	}
	// If no MYSQL_DB is provided, use an in-memory store. This should only
	// be used in development mode.
	return &Store{syncmap.NewStore(syncmap.DefaultOptions)}, nil
}

func (s *Store) Get(key string) (*Item, error) {
	i := new(Item)
	found, err := s.store.Get(key, i)
	if err != nil {
		return nil, errors.Wrap(err, "could not get item")
	}
	if !found {
		return nil, nil
	}
	return i, nil
}

func (s *Store) Put(key string, id int64) (*Item, error) {
	i := Item{
		ID:        id,
		WriteTime: time.Now(),
	}
	return &i, s.store.Set(key, i)
}
