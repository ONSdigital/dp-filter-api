package mongo

import (
	"context"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	testMongoContainer "github.com/testcontainers/testcontainers-go/modules/mongodb"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// getTestMongoClient starts a MongoDB container and returns a client connected to it.
func getTestMongoClient(ctx context.Context, t *testing.T) *mongo.Client {
	t.Helper()

	mongoContainer, err := testMongoContainer.Run(ctx, "mongo:4.4.8")
	if err != nil {
		t.Fatalf("failed to start MongoDB container: %v", err)
	}
	t.Cleanup(func() {
		testcontainers.CleanupContainer(t, mongoContainer)
	})

	connectionString, err := mongoContainer.ConnectionString(ctx)
	if err != nil {
		t.Fatalf("failed to get MongoDB connection string: %v", err)
	}

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(connectionString))
	if err != nil {
		t.Fatalf("failed to connect to MongoDB container: %v", err)
	}
	t.Cleanup(func() {
		if err := client.Disconnect(ctx); err != nil {
			t.Errorf("failed to disconnect MongoDB client: %v", err)
		}
	})

	return client
}
