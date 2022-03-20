package main

import (
	"log"

	"github.com/chuamingkai/50.041DynamoProject/internal/bolt"
	"github.com/chuamingkai/50.041DynamoProject/internal/models"
)

func main() {
	id := 55 // TODO: Dynamically assign node IDs

	testEntry := models.Object{
		Key: "S1234567A",
		Value: "23:23:23:23 NW",
	}

	// Open database
	db, err := bolt.ConnectDB(id)
	if err != nil {
		log.Fatalf("Error opening database: %s", err)
	}
	defer db.DB.Close()

	// Create bucket
	err = db.CreateBucket("testBucket")
	if err != nil {
		log.Fatalf("Error creating bucket: %s", err)
	}

	// Insert test value into bucket
	err = db.Put("testBucket", testEntry)
	if err != nil {
		log.Fatalf("Error inserting into bucket: %s", err)
	}

	// Read from bucket
	// getResult := db.Get("testBucket", testEntry.Key)
	// fmt.Println(getResult)
}
