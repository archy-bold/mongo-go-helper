package schema

import (
	"context"
	"reflect"

	"github.com/archy-bold/mongo-go-helper/base"
)

// Migrator is used for running migrations
type Migrator interface {
	Run(ctx context.Context, tasks map[string]TaskContract) error
}

// TaskContract represents a migration task
type TaskContract interface {
	GetType() string
}

// Task represents a migration task instance
type Task struct {
	Collection string
}

// GetType get the type of task
func (t *Task) GetType() string {
	return reflect.TypeOf(t).String()
}

// CreateIndexTask is a migration task for creating an index
type CreateIndexTask struct {
	Task
	IndexName string
	Keys      interface{}
}

// FindFilterFunction represents a function to get the filter to find an item
type FindFilterFunction func(item interface{}) (interface{}, error)

// SeededCallbackFunction represents a function that's called when an item is seeded
type SeededCallbackFunction func(item base.ModelInterface)
