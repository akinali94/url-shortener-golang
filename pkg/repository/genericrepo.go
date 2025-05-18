package repository

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type GenericMongoRepo[T any] struct {
	coll *mongo.Collection
}

func NewRepository[T any](mongoColl *mongo.Collection) *GenericMongoRepo[T] {
	return &GenericMongoRepo[T]{
		coll: mongoColl,
	}
}

func (c *GenericMongoRepo[T]) Add(item T) (interface{}, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := c.coll.InsertOne(ctx, item)

	if err != nil {
		fmt.Println(err.Error())
		return nil, err
	}

	//TODO: buraada ID'miz bizim generatedId olsun, tekrar database _id olusturmasin
	return result.InsertedID, nil
}

func (c *GenericMongoRepo[T]) Get(id string) (*T, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancel()

	var item T
	err := c.coll.FindOne(ctx, bson.M{"_id": id}).Decode(&item)

	if err != nil {
		fmt.Println(err.Error())
	}

	return &item, err
}

func (c *GenericMongoRepo[T]) GetByField(val string, field string) (*T, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var item T
	filter := bson.D{{"shortUrl", val}}
	err := c.coll.FindOne(ctx, filter).Decode(&item)
	if err != nil {
		fmt.Println("GetByField'da hata, err: " + err.Error())
	}

	return &item, err
}

func (c *GenericMongoRepo[T]) Delete(id string) (int64, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := c.coll.DeleteOne(ctx, bson.D{{Key: "_id", Value: id}})

	if err != nil {
		fmt.Println(err.Error())
	}

	return result.DeletedCount, err
}
