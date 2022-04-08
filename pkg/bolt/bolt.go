package bolt

import (
	"errors"
	"fmt"

	bolt "go.etcd.io/bbolt"
)

type DB struct {
	DB *bolt.DB
}

// Setup database
func ConnectDB(id int) (*DB, error) {
	dbName := fmt.Sprintf("store/boltdb/node%vstore.db", id)
	db, err := bolt.Open(dbName, 0600, nil)
	return &DB{DB: db}, err
}

// Create bucket
func (db *DB) CreateBucket(bucketName string) error {
	return db.DB.Update(func(tx *bolt.Tx) error {
		_, bucketErr := tx.CreateBucket([]byte(bucketName))
		return bucketErr
	})
}

// Check if bucket exists in database
func (db *DB) BucketExists(bucketName string) bool {
	var exists bool
	db.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		exists = (b != nil)
		return nil
	})
	return exists
}

// Insert into bucket
func (db *DB) Put(bucketName, key string, data []byte) error {
	return db.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		return b.Put([]byte(key), data)
	})
}

// Read value at key in bucket
func (db *DB) Get(bucketName, key string) ([]byte, error) {
	var numBytes int
	var value []byte
	err := db.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		v := b.Get([]byte(key))
		numBytes = len(v)
		if v != nil {
			value = make([]byte, numBytes)
			copy(value, v)
		}
		return nil
	})
	return value, err
}

// Execute function for each key-value pair in bucket
func (db *DB) Iterate(bucketname string, handle_kv func(k, v []byte) error) error {
	err := db.DB.Update(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte(bucketname))
		return b.ForEach(handle_kv)
	})
	return err
}

// Delete key from bucket
func (db *DB) DeleteKey(bucketname string, key string) error {
	err := db.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketname))
		return b.Delete([]byte(key))
	})
	return err
}

// Return list of all objects in a bucket
func (db *DB) GetAllObjects(bucketname string) ([][]byte, error) {	
	var bucketObjects [][]byte
	err := db.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketname))
		if b == nil {
			return errors.New("bucket does not exist")
		}
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			bucketObjects = append(bucketObjects, v)
		}
		return nil
	})

	if err != nil {
		return nil, err
	} else {
		return bucketObjects, nil
	}
}

// Return list of all buckets in the database
func (db *DB) GetAllBuckets() ([]string, error) {
	var bucketNames []string
	err := db.DB.View(func(tx *bolt.Tx) error {
		return tx.ForEach(func(name []byte, _ *bolt.Bucket) error {
			bucketNames = append(bucketNames, string(name))
			return nil
		})
	})
	return bucketNames, err
}