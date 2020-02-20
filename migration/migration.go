package migration

import (
	"context"
	"errors"
	"fmt"

	"github.com/archy-bold/mongo-go-helper/base"
	"github.com/archy-bold/mongo-go-helper/migration/schema"
	multierror "github.com/hashicorp/go-multierror"
)

type migrator struct {
	helper base.MongoHelper
	seeder schema.Seeder
}

func (m *migrator) Run(ctx context.Context, tasks map[string]schema.TaskContract) error {
	var errs *multierror.Error

	for tn, t := range tasks {
		var err error
		switch task := interface{}(t).(type) {
		case *schema.CreateIndexTask:
			err = m.helper.AddIndexIfNotExists(ctx, task.Collection, task.IndexName, task.Keys)
		case *schema.SeedTableTask:
			err = m.seeder.SeedData(ctx, task)
		default:
			errStr := fmt.Sprintf("could not run migration task '%s': unknown type '%s'", tn, t.GetType())
			err = errors.New(errStr)
		}

		if err != nil {
			errs = multierror.Append(err)
		}
	}

	return errs.ErrorOrNil()
}

// NewMigrator returns a new migrator instance for the given tasks
func NewMigrator(helper base.MongoHelper) (m schema.Migrator) {
	seeder := &seeder{helper}
	m = &migrator{helper, seeder}
	return
}
