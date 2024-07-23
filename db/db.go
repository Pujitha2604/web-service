package db

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ConnectMongoDB connects to the MongoDB database using the provided URI and timeout duration.
func ConnectMongoDB(uri string, timeout time.Duration) (*mongo.Client, error) {
	client, err := mongo.NewClient(options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	err = client.Connect(ctx)
	if err != nil {
		return nil, err
	}
	return client, nil
}

// DisconnectMongoDB disconnects from the MongoDB database.
func DisconnectMongoDB(client *mongo.Client, ctx context.Context) error {
	return client.Disconnect(ctx)
}
