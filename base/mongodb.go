package base

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoClient encapsulates the MongoDB Client struct
type MongoClient interface {
	Connect(context.Context) error
	Database(string, ...*options.DatabaseOptions) *mongo.Database
}

// MongoDB encapsulates the MongoDB DB struct
type MongoDB interface {
	Collection(string, ...*options.CollectionOptions) *mongo.Collection
}

// ModelInterface defines common mongo models
type ModelInterface interface {
	Exists() bool
	GetID() interface{}
	SetID(interface{})
}
