package bolt

import (
	"fmt"

	bolt "go.etcd.io/bbolt"
)

type DB struct {
	DB *bolt.DB
}

// Setup database
func ConnectDB(id int) (*DB, error) {
	dbName := fmt.Sprintf("store/node%vstore.db", id)
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

func (db *DB) Iterate(bucketname string, handle_kv func(k, v []byte) error) error {
	err := db.DB.Update(func(tx *bolt.Tx) error {
		// Assume bucket exists and has keys
		b := tx.Bucket([]byte(bucketname))
		return b.ForEach(handle_kv)
	})
	return err
}

func (db *DB) DeleteKey(bucketname string, key string) error {
	err := db.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketname))
		return b.Delete([]byte(key))
	})
	return err
}
