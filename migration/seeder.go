package migration

import (
	"context"

	"github.com/archy-bold/mongo-go-helper/base"
	"github.com/archy-bold/mongo-go-helper/migration/schema"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/jinzhu/copier"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type seeder struct {
	helper base.MongoHelper
}

func (s *seeder) SeedData(ctx context.Context, task *schema.SeedTableTask) error {
	var errs *multierror.Error

	// Check everything is set on the task
	if task == nil {
		return nil
	}
	if task.Collection == "" {
		return ErrNoCollection
	}
	if task.FindFilterFn == nil {
		return ErrNoFindFilterFn
	}
	if task.Model == nil {
		return ErrNoModel
	}

	for _, it := range task.Items {
		// Get a copy to avoid issues when referencing the slice element
		var item, existing base.ModelInterface
		copier.Copy(&item, &it)
		copier.Copy(&existing, &task.Model)

		// Find if it already exists
		filter, err := task.FindFilterFn(item)

		if err == nil {
			if filter != nil {
				s.helper.FindOne(ctx, task.Collection, filter, existing)
			}

			if existing != nil && !existing.Exists() {
				if _, ok := existing.GetID().(primitive.ObjectID); ok {
					item.SetID(primitive.NewObjectID())
				}
				_, err = s.helper.InsertOne(ctx, task.Collection, item)
			} else {
				if _, ok := existing.GetID().(primitive.ObjectID); ok {
					item.SetID(existing.GetID())
				}
				err = s.helper.UpdateOne(ctx, task.Collection, filter, item)
			}
		}

		if err != nil {
			errs = multierror.Append(errs, err)
		} else if task.Callback != nil {
			task.Callback(item)
		}
	}
	return errs.ErrorOrNil()
}
