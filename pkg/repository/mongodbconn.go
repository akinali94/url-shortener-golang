package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func DbConnection(connStr string, dbName string, dbColl string) *mongo.Collection {
	client, err := mongo.Connect(options.Client().ApplyURI(connStr))

	if err != nil {
		panic(err)
	}

	defer func() {
		if err := client.Disconnect(context.TODO()); err != nil {
			panic(err)
		}
	}()

	coll := client.Database(dbName).Collection(dbColl)

	return coll
}
