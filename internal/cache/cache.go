package cache

import (
	"encoding/json"
	"errors"
	"github.com/dgraph-io/badger/v3"
)

// Cache is a simple struct to manage the badger database
type Cache struct {
	db *badger.DB
}

// NewCache initializes and returns a Cache instance
func NewCache(dbPath string) (*Cache, error) {
	opts := badger.DefaultOptions(dbPath)
	opts = opts.WithLogger(nil) // Disable default logger if needed

	db, err := badger.Open(opts)
	if err != nil {
		return nil, err
	}

	return &Cache{db: db}, nil
}

// Set stores the object for a given key
func (c *Cache) Set(key string, object interface{}) error {
	data, err := json.Marshal(object)
	if err != nil {
		return err
	}

	return c.db.Update(func(txn *badger.Txn) error {
		return txn.Set([]byte(key), data)
	})
}

// Get retrieves the object for a given key
func (c *Cache) Get(key string, target interface{}) (hit bool, err error) {
	err = c.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if errors.Is(err, badger.ErrKeyNotFound) {
			return nil
		} else if err != nil {
			return err
		}

		return item.Value(func(val []byte) error {
			hit = true
			return json.Unmarshal(val, target)
		})
	})
	return hit, err
}

// Close closes the badger database
func (c *Cache) Close() error {
	return c.db.Close()
}
