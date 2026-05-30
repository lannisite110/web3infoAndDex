package db

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

const connectTimeout = 30 * time.Second

// Mongo wraps a MongoDB client and database handle.
type Mongo struct {
	Client   *mongo.Client
	Database *mongo.Database
}

// Connect establishes a MongoDB connection and verifies with a ping.
func Connect(ctx context.Context, uri, dbName string) (*Mongo, error) {
	ctx, cancel := context.WithTimeout(ctx, connectTimeout)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().
		ApplyURI(uri).
		SetServerSelectionTimeout(connectTimeout).
		SetTLSConfig(&tls.Config{MinVersion: tls.VersionTLS12}),
	)
	if err != nil {
		return nil, fmt.Errorf("mongo connect: %w", err)
	}

	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		_ = client.Disconnect(context.Background())
		return nil, fmt.Errorf("mongo ping: %w", err)
	}

	return &Mongo{
		Client:   client,
		Database: client.Database(dbName),
	}, nil
}

// Ping checks database connectivity.
func (m *Mongo) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return m.Client.Ping(ctx, readpref.Primary())
}

// Close disconnects the client.
func (m *Mongo) Close(ctx context.Context) error {
	if m == nil || m.Client == nil {
		return nil
	}
	return m.Client.Disconnect(ctx)
}

// Collection returns a named collection.
func (m *Mongo) Collection(name string) *mongo.Collection {
	return m.Database.Collection(name)
}
