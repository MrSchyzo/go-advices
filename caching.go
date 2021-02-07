package main

import (
	j "encoding/json"
	"time"

	"github.com/tidwall/buntdb"
)

// CacheForAdvices interface is a trivial cache for string slices
type CacheForAdvices interface {
	Get(key string) ([]string, error)
	Put(key string, advices []string) ([]string, error)
}

// InMemoryCacheForAdvices struct is an implementation dependent upon buntDB
type InMemoryCacheForAdvices struct {
	db *buntdb.DB
}

// Get function
func (c *InMemoryCacheForAdvices) Get(key string) ([]string, error) {
	var cachedString string
	var cached []string

	err := c.db.View(func(tx *buntdb.Tx) error {
		val, err := tx.Get(key, false)
		if err != nil {
			return err
		}
		cachedString = val
		return nil
	})

	if err != nil {
		return nil, err
	}

	j.Unmarshal([]byte(cachedString), &cached)
	return cached, nil
}

// Put function
func (c *InMemoryCacheForAdvices) Put(key string, advices []string) ([]string, error) {
	value, failure := j.Marshal(advices)
	if failure != nil {
		return nil, failure
	}

	err := c.db.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(key, string(value), &buntdb.SetOptions{Expires: true, TTL: time.Minute * 5})
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return advices, nil
}
