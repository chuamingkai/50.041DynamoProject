package bolt

import (
	"encoding/json"
	"fmt"

	"github.com/chuamingkai/50.041DynamoProject/internal/models"
	bolt "go.etcd.io/bbolt"
)

type DB struct {
	DB *bolt.DB
}

// Setup database
func ConnectDB(id int) (DB, error) {
	dbName := fmt.Sprintf("store/node%vstore.db", id)
	db, err := bolt.Open(dbName, 0600, nil)
	return DB{DB: db}, err
}

// Create bucket
func (db *DB) CreateBucket(bucketName string) error {
	return db.DB.Update(func(tx *bolt.Tx) error {
		_, bucketErr := tx.CreateBucketIfNotExists([]byte(bucketName))
		return bucketErr
	})
}

// Insert into bucket
func (db *DB) Put(bucketName string, object models.Object) error {
	return db.DB.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		newentry, err := json.Marshal(object)

		b.Put([]byte(object.Key), newentry)
		return err
	})
}

// Read value at key in bucket
func (db *DB) Get(bucketName, key string) (models.Object, error) {
	value := make([]byte, 1000)
	var newupd models.Object

	err := db.DB.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(bucketName))
		copy(value, b.Get([]byte(key)))
		err := json.Unmarshal(b.Get([]byte(key)), &newupd)
		return err
	})

	if err != nil {
		return newupd, err
	}

	return newupd, nil

}
