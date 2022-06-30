package upload

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// mongodb struct as concete implementation of datastore interface
type mongodb struct {
	db      *mongo.Database
	context context.Context
}

// Create new instance of mongodb with specific parameters
func NewDb(name string, app string, url string, timeout time.Duration) DataStore {
	client, err := mongo.NewClient(options.Client().ApplyURI(url), options.Client().SetAppName(app))
	if err != nil {
		fmt.Println("Cannot set up Database:", err)
		return nil
	}

	ctx, _ := context.WithTimeout(context.Background(), timeout)
	err = client.Connect(ctx)
	if err != nil {
		fmt.Println("Database connection error:", err)
		return nil
	}
	if err = client.Ping(ctx, readpref.Primary()); err != nil {
		fmt.Println("Cannot ping the database:", err)
		return nil
	}

	return &mongodb{
		db:      client.Database(name),
		context: ctx,
	}

}

// Push documents into datastore
func (ds *mongodb) Set(doc map[string][]interface{}) error {

	if doc == nil {
		errors.New("No doc provided")
	}
	for fn, flow := range doc {
		col := ds.db.Collection(fn)
		// insert specific record into specific collection
		result, err := col.InsertMany(ds.context, flow)
		if err != nil {
			return err
		}
		fmt.Printf("Collection: %+v - Objects : %+v\n", fn, result)

		col = ds.db.Collection("All")
		result, err = col.InsertMany(ds.context, flow)

		if err != nil {
			return err
		}

		// insert specific record into combined collection
		fmt.Printf("Collection: %+v - Objects : %+v\n", "All", result)

	}
	return nil
}
