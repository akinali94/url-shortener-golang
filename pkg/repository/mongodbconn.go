package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type MongoDB struct {
	Client     *mongo.Client
	Database   *mongo.Database
	Collection *mongo.Collection
}

func NewMongoDB(connString, dbName, collectionName string) (*MongoDB, error) {

	client, err := mongo.Connect(options.Client().ApplyURI(connString))
	if err != nil {
		return nil, err
	}

	database := client.Database(dbName)
	collection := database.Collection(collectionName)

	return &MongoDB{
		Client:     client,
		Database:   database,
		Collection: collection,
	}, nil
}

func (m *MongoDB) Close(ctx context.Context) error {
	return m.Client.Disconnect(ctx)
}
