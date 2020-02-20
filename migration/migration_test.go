package migration

import (
	"context"
	"errors"
	"testing"

	"github.com/archy-bold/mongo-go-helper/base"
	"github.com/archy-bold/mongo-go-helper/migration/schema"
	bMocks "github.com/archy-bold/mongo-go-helper/mocks/base"
	mMocks "github.com/archy-bold/mongo-go-helper/mocks/migration"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	sampleDocument = &bson.M{"num": 999}
	errExample     = errors.New("error")
)

type exampleStruct struct {
	ID  primitive.ObjectID `bson:"_id, omitempty"`
	Str string             `bson:"str"`
	Num int                `bson:"num"`
}

type indexCall struct {
	collection string
	index      string
	document   *bson.M
	err        error
}

type seedCall struct {
	task string
	err  error
}

var runMigrationsTests = map[string]struct {
	tasks      map[string]schema.TaskContract
	indexCalls []indexCall
	seedCalls  []seedCall
	expected   error
}{
	"nil tasks": {
		nil, nil, nil, nil,
	},
	"empty map": {
		map[string]schema.TaskContract{}, nil, nil, nil,
	},
	"bad type": {
		map[string]schema.TaskContract{
			"blank": &schema.Task{},
		},
		nil,
		nil,
		buildMultiError([]error{
			errors.New("could not run migration task 'blank': unknown type '*schema.Task'"),
		}),
	},
	"index success": {
		map[string]schema.TaskContract{
			"index": &schema.CreateIndexTask{Task: schema.Task{Collection: "test-coll"}, IndexName: "index_1", Keys: sampleDocument},
		},
		[]indexCall{indexCall{"test-coll", "index_1", sampleDocument, nil}},
		nil,
		nil,
	},
	"index error": {
		map[string]schema.TaskContract{
			"index": &schema.CreateIndexTask{Task: schema.Task{Collection: "test-coll"}, IndexName: "index_1", Keys: sampleDocument},
		},
		[]indexCall{indexCall{"test-coll", "index_1", sampleDocument, errExample}},
		nil,
		buildMultiError([]error{errExample}),
	},
	"seed success": {
		map[string]schema.TaskContract{
			"seed": &schema.SeedTableTask{},
		},
		nil,
		[]seedCall{seedCall{"seed", nil}},
		nil,
	},
	"seed error": {
		map[string]schema.TaskContract{
			"seed": &schema.SeedTableTask{},
		},
		nil,
		[]seedCall{seedCall{"seed", errExample}},
		buildMultiError([]error{errExample}),
	},
}

func Test_RunMigrations(t *testing.T) {
	ctx := context.Background()

	for tn, tt := range runMigrationsTests {
		// Set up the mock
		helper := &bMocks.MongoHelper{}
		seeder := &mMocks.Seeder{}
		m := &migrator{helper, seeder}
		numSeedCalls := len(tt.seedCalls)
		if len(tt.indexCalls) > 0 {
			helper.On("AddIndexIfNotExists", ctx, mock.AnythingOfType("string"), mock.AnythingOfType("string"), mock.AnythingOfType("*primitive.M")).
				Return(tt.indexCalls[0].err).
				Times(len(tt.indexCalls))
		}
		if numSeedCalls > 0 {
			seeder.On("SeedData", ctx, mock.AnythingOfType("*schema.SeedTableTask")).
				Return(tt.seedCalls[0].err).
				Times(numSeedCalls)
		}

		err := m.Run(ctx, tt.tasks)

		// The assertions
		for _, indexCall := range tt.indexCalls {
			helper.AssertCalled(t, "AddIndexIfNotExists", ctx, indexCall.collection, indexCall.index, indexCall.document)
		}
		if numSeedCalls > 0 {
			seeder.AssertNumberOfCalls(t, "SeedData", numSeedCalls)
		} else {
			seeder.AssertNotCalled(t, "SeedData")
		}
		if tt.expected == nil {
			assert.Nilf(t, err, "Expected err to be nil on RunMigration test '%s'", tn)
		} else {
			assert.NotNilf(t, err, "Expected err to not be nil on RunMigration test '%s'", tn)
			assert.Equalf(t, tt.expected, err, "Expected err to match on RunMigration test '%s'", tn)
		}
	}
}

var newMigratorTests = map[string]struct {
	helper base.MongoHelper
}{
	"nil values":  {nil},
	"empty tasks": {&bMocks.MongoHelper{}},
	"has tasks":   {&bMocks.MongoHelper{}},
}

func Test_NewMigrator(t *testing.T) {
	for tn, tt := range newMigratorTests {
		m := NewMigrator(tt.helper)

		assert.NotNilf(t, m, "Expected migrator not to be nil on NewMigrator test '%s'", tn)
		assert.IsType(t, &migrator{}, m, "Expected migrator to be of expected type on NewMigrator test '%s'", tn)
		assert.Equal(t, tt.helper, m.(*migrator).helper, "Expected helper to match on NewMigrator test '%s'", tn)
	}
}

func buildMultiError(errs []error) *multierror.Error {
	return &multierror.Error{Errors: errs}
}
