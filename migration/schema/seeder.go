package schema

import (
	"context"

	"github.com/archy-bold/mongo-go-helper/base"
)

// SeedTableTask is a migration task for seeding a table
type SeedTableTask struct {
	Task
	Model        base.ModelInterface
	FindFilterFn FindFilterFunction
	Callback     SeededCallbackFunction
	Items        []base.ModelInterface
	Result       interface{}
}

// Seeder represents a seeder
type Seeder interface {
	SeedData(ctx context.Context, task *SeedTableTask) error
}
