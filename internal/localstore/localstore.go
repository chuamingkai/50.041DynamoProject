package localstore

import (
	"github.com/chuamingkai/50.041DynamoProject/internal/bolt"
	"github.com/chuamingkai/50.041DynamoProject/internal/models"
)

type Store struct {
	boltDB bolt.DB
}

func (store *Store) InsertOne(bucketName string, object models.Object) error {
	return store.boltDB.Put(bucketName, object)
}

func (store *Store) GetOne(bucketName, key string) string {
	return store.boltDB.Get(bucketName, key)
}