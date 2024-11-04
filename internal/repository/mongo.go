package repository

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func ConnectToMongoDB(ctx context.Context, mongoURI string) (*mongo.Client, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		return nil, errors.Join(errors.New("connecting to MongoDB"), err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		return nil, errors.Join(errors.New("pinging MongoDB"), err)
	}

	return client, nil
}
