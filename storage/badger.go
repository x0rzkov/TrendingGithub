package storage

import (
	"time"
	"os"
	"fmt"
	"log"

	"github.com/golang/snappy"
	badger "github.com/dgraph-io/badger"
)

// BadgerStorage represents the in memory storage engine.
// This storage can be useful for debugging / development
type BadgerStorage struct{}

// BadgerPool is the pool of connections to your local memory ;)
type BadgerPool struct {
	cacheDir string
	storage *badger.DB
}

// BadgerConnection represents a in memory connection
type BadgerConnection struct {
	storage *badger.DB
}

// NewPool returns a pool to communicate with your in memory
func (ms *BadgerStorage) NewPool(dir, url, auth string) Pool {
	err := ensureDir(dir)
	if err != nil {
		log.Fatal(err)
	}
	store, err := badger.Open(badger.DefaultOptions(dir))
	if err != nil {
		log.Fatal(err)
	}

	return BadgerPool{
		cacheDir: dir,
		storage: store,
	}
}

// Close closes a in memory pool
func (mp BadgerPool) Close() error {
	return mp.storage.Close()
}

// Get returns you a connection to your in memory storage
func (mp BadgerPool) Get() Connection {
	return &BadgerConnection{
		storage: mp.storage,
	}
}

// Err will return an error once one occurred
func (mc *BadgerConnection) Err() error {
	return nil
}

// Close shuts down a in memory connection
func (mc *BadgerConnection) Close() error {
	return mc.storage.Close()
}

// MarkRepositoryAsTweeted marks a single projects as "already tweeted".
// This information will be stored as a hashmap with a TTL.
// The timestamp of the tweet will be used as value.
func (mc *BadgerConnection) MarkRepositoryAsTweeted(projectName, score string) (bool, error) {
	// Add grey listing to current time
	now := time.Now()
	future := now.Add(time.Second * BlackListTTL)
	err := mc.storage.Update(func(txn *badger.Txn) error {
		fmt.Println("indexing: ", projectName)
		cnt, err := compress([]byte(future.String()))
		if err != nil {
			return err
		}
		err = txn.Set([]byte(projectName), cnt)
		return err
	})
	return true, err
}

// IsRepositoryAlreadyTweeted checks if a project was already tweeted.
// If it is not available
//	a) the project was not tweeted yet
//	b) the project ttl expired and is ready to tweet again
func (mc *BadgerConnection) IsRepositoryAlreadyTweeted(projectName string) (bool, error) {
	err := mc.storage.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(projectName))
		if err != nil {
			return err
		}
		err = item.Value(func(val []byte) error {
			// This func with val would only be called if item.Value encounters no error.
			// Accessing val here is valid.
			valdo, err := decompress(val)
			if err != nil {
				return err
			}
			fmt.Printf("The answer is: %s\n", string(valdo))
			layout := "2006-01-02T15:04:05.000Z"
			t, err := time.Parse(layout, string(valdo))
			if err != nil {
				return err
			}
			fmt.Println(t)
			if res := t.Before(time.Now()); res == true {
				// delete(mc.storage, projectName)
				return nil
			}
			return nil
		})
		if err != nil {
			return err
		}
		return nil
	})
	return err == nil, err
}

func (mc *BadgerConnection) delete(projectName string) (bool, error) {
	return true, nil
}

func ensureDir(path string) error {
	d, err := os.Open(path)
	if err != nil {
		os.MkdirAll(path, os.FileMode(0755))
	} else {
		return err
	}
	d.Close()
	return nil
}

func compress(data []byte) ([]byte, error) {
	return snappy.Encode([]byte{}, data), nil
}

func decompress(data []byte) ([]byte, error) {
	return snappy.Decode([]byte{}, data)
}
