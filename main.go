package main

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	//"github.com/DistributedClocks/GoVector/govec/vclock"
)

type Context struct {
	//vectorClock      vclock.VClock
	isConflicting    bool   `bson:"isConflicting,omitempty"`
	writeCoordinator string `bson:"writeCoordinator,omitempty"`
}

type Object struct {
	ID         primitive.ObjectID `bson:"_id,omitempty"`
	Objectname string             `bson:"objectname,omitempty"`
	Objectcont string             `bson:"objectcontent,omitempty"`
}

func close(client *mongo.Client, ctx context.Context,
	cancel context.CancelFunc) {

	defer cancel()

	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()
}

func connect(uri string) (*mongo.Client, context.Context, context.CancelFunc, error) {

	ctx, cancel := context.WithTimeout(context.Background(),
		30*time.Second)
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	return client, ctx, cancel, err
}

func insertOne(client *mongo.Client, ctx context.Context, dataBase, col string, doc interface{}) (*mongo.InsertOneResult, error) {
	collection := client.Database(dataBase).Collection(col)
	result, err := collection.InsertOne(ctx, doc)
	return result, err
}

func insertMany(client *mongo.Client, ctx context.Context, dataBase, col string, docs []interface{}) (*mongo.InsertManyResult, error) {

	collection := client.Database(dataBase).Collection(col)

	result, err := collection.InsertMany(ctx, docs)
	return result, err
}

func main() {
	client, ctx, cancel, err := connect("mongodb+srv://pw:b6O6Uoe62He5R766@cluster0.8apga.mongodb.net/myFirstDatabase?retryWrites=true&w=majority")

	if err != nil {
		panic(err)
	}

	defer close(client, ctx, cancel)

	object := Object{
		Objectname: "S1234567A",
		Objectcont: "23:23:23:23 NW",
	}

	//test insert
	insertOneResult, err := insertOne(client, ctx, "nodes", "node1", object)

	// handle the error
	if err != nil {
		panic(err)
	}

	fmt.Println("Result of InsertOne")
	fmt.Println(insertOneResult.InsertedID)

}
