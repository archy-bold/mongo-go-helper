package migration

import (
	"context"
	"testing"

	"github.com/archy-bold/mongo-go-helper/base"
	"github.com/archy-bold/mongo-go-helper/migration/schema"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	testDB     = "wata-delivery-test"
	defaultURI = "mongodb://localhost:27017"
	cMock      = &callbackMock{}
)

type callbackMock struct {
	mock.Mock
}

func (m *callbackMock) Callback(item base.ModelInterface) {
	m.Called(item)
	return
}

type exampleModelStrID struct {
	ID  string `bson:"_id"`
	Num int    `bson:"num"`
}

func (m exampleModelStrID) Exists() bool {
	return m.ID != ""
}

func (m exampleModelStrID) GetID() interface{} {
	return m.ID
}

func (m *exampleModelStrID) SetID(id interface{}) {
	m.ID = id.(string)
}

type exampleModel struct {
	ID  primitive.ObjectID `bson:"_id, omitempty"`
	Str string             `bson:"str"`
	Num int                `bson:"num"`
}

func (m exampleModel) Exists() bool {
	return !m.ID.IsZero()
}

func (m exampleModel) GetID() interface{} {
	return m.ID
}

func (m *exampleModel) SetID(id interface{}) {
	m.ID = id.(primitive.ObjectID)
}

func findFilterFn(item interface{}) (interface{}, error) {
	em := item.(*exampleModel)
	return &bson.M{"num": int32(em.Num)}, nil
}

func findFilterFn2(item interface{}) (interface{}, error) {
	em := item.(*exampleModelStrID)
	return &bson.M{"_id": em.ID}, nil
}

func findFilterFnErr(item interface{}) (interface{}, error) {
	return nil, errExample
}

var seedDataTests = []struct {
	tn       string
	task     *schema.SeedTableTask
	clean    bool
	err      error
	expected []interface{}
}{
	{"nil task", nil, true, nil, []interface{}{}},
	{"empty task", &schema.SeedTableTask{}, true, ErrNoCollection, nil},
	{"no filter",
		&schema.SeedTableTask{Task: schema.Task{"test"}},
		true,
		ErrNoFindFilterFn,
		nil,
	},
	{"no model",
		&schema.SeedTableTask{Task: schema.Task{"test"}, FindFilterFn: findFilterFn},
		true,
		ErrNoModel,
		nil,
	},
	{"no data",
		&schema.SeedTableTask{Task: schema.Task{"test"}, FindFilterFn: findFilterFn, Model: &exampleModel{}},
		true,
		nil,
		[]interface{}{},
	},
	{"seeds new data",
		&schema.SeedTableTask{
			Task:         schema.Task{"test"},
			Items:        []base.ModelInterface{&exampleModel{Str: "test1", Num: 999}, &exampleModel{Str: "test2", Num: 1000}},
			FindFilterFn: findFilterFn,
			Model:        &exampleModel{},
		},
		false,
		nil,
		[]interface{}{exampleModel{Str: "test1", Num: 999}, exampleModel{Str: "test2", Num: 1000}},
	},
	{"updates data",
		&schema.SeedTableTask{
			Task:         schema.Task{"test"},
			Items:        []base.ModelInterface{&exampleModel{Str: "testing", Num: 999}, &exampleModel{Str: "testing2", Num: 1000}},
			FindFilterFn: findFilterFn,
			Model:        &exampleModel{},
		},
		true,
		nil,
		[]interface{}{exampleModel{Str: "testing", Num: 999}, exampleModel{Str: "testing2", Num: 1000}},
	},
	{"seeds new with callback",
		&schema.SeedTableTask{
			Task:         schema.Task{"test"},
			Items:        []base.ModelInterface{&exampleModel{Str: "test1", Num: 999}, &exampleModel{Str: "test2", Num: 1000}},
			FindFilterFn: findFilterFn,
			Model:        &exampleModel{},
			Callback:     cMock.Callback,
		},
		true,
		nil,
		[]interface{}{exampleModel{Str: "test1", Num: 999}, exampleModel{Str: "test2", Num: 1000}},
	},
	{"seeds new - error",
		&schema.SeedTableTask{
			Task:         schema.Task{"!"},
			Items:        []base.ModelInterface{&exampleModel{Str: "test1", Num: 999}},
			FindFilterFn: findFilterFnErr,
			Model:        &exampleModel{},
		},
		true,
		&multierror.Error{Errors: []error{errExample}},
		[]interface{}{},
	},
	{"seeds new data - string ID",
		&schema.SeedTableTask{
			Task:         schema.Task{"test2"},
			Items:        []base.ModelInterface{&exampleModelStrID{ID: "test1", Num: 999}, &exampleModelStrID{ID: "test2", Num: 1000}},
			FindFilterFn: findFilterFn2,
			Model:        &exampleModelStrID{},
		},
		false,
		nil,
		[]interface{}{exampleModelStrID{ID: "test1", Num: 999}, exampleModelStrID{ID: "test2", Num: 1000}},
	},
	{"updates data - string ID",
		&schema.SeedTableTask{
			Task:         schema.Task{"test2"},
			Items:        []base.ModelInterface{&exampleModelStrID{ID: "test1", Num: 1}, &exampleModelStrID{ID: "test2", Num: 2}},
			FindFilterFn: findFilterFn2,
			Model:        &exampleModelStrID{},
		},
		true,
		nil,
		[]interface{}{exampleModelStrID{ID: "test1", Num: 1}, exampleModelStrID{ID: "test2", Num: 2}},
	},
}

func Test_DeliveriesRepository_SeedData(t *testing.T) {
	// Connect to the mongo database
	var db base.MongoDB
	ctx := context.Background()
	opt := options.Client()
	opt.ApplyURI(defaultURI)
	client, _ := mongo.NewClient(opt)
	db = client.Database(testDB)
	helper := base.NewHelper(db)
	client.Connect(ctx)
	defer cleanTests(ctx, db)

	// Create the seeder too
	s := &seeder{helper}

	for _, tt := range seedDataTests {
		// Expect the callback to be mocked
		cback := false
		if tt.task != nil && tt.task.Callback != nil {
			cMock.On("Callback", mock.AnythingOfType("*migration.exampleModel")).
				Times(len(tt.expected))
			cback = true
		}

		err := s.SeedData(ctx, tt.task)

		// The assertions
		if tt.err == nil {
			assert.Nilf(t, err, "Expected nil err for SeedData on test '%s'", tt.tn)
			if cback {
				cMock.AssertNumberOfCalls(t, "Callback", len(tt.expected))
			}
			// Check that all the items exist
			var item exampleModel
			coll := "test"
			if tt.task != nil {
				coll = tt.task.Collection
			}
			res, _ := helper.Find(ctx, coll, nil, item, base.FindOptions{})

			assert.Lenf(t, res.Items, len(tt.expected), "Expected collection to match the expected len for SeedData on test '%s'", tt.tn)
			for i, ex := range tt.expected {
				if examp, ok := ex.(*exampleModel); ok {
					item := findExampleModel(res.Items, examp.Num)
					assert.NotNilf(t, item, "Expected non-nil result for SeedData on test '%s', i %d", tt.tn, i)
					assert.NotEqualf(t, primitive.NilObjectID, item.ID, "Expected non-zero ID for SeedData on test '%s', i %d", tt.tn, i)
					assert.Equalf(t, examp.Str, item.Str, "Expected Str to match the expected for SeedData on test '%s', i %d", tt.tn, i)
					assert.Equalf(t, examp.Num, item.Num, "Expected Num to match the expected for SeedData on test '%s', i %d", tt.tn, i)
				}
				if examp, ok := ex.(*exampleModelStrID); ok {
					item := findExampleModelStrID(res.Items, examp.ID)
					assert.NotNilf(t, item, "Expected non-nil result for SeedData on test '%s', i %d", tt.tn, i)
					assert.Equalf(t, examp.ID, item.ID, "Expected ID to match the expected for SeedData on test '%s', i %d", tt.tn, i)
					assert.Equalf(t, examp.Num, item.Num, "Expected Num to match the expected for SeedData on test '%s', i %d", tt.tn, i)
				}
			}
		} else {
			cMock.AssertNotCalled(t, "Callback")
			assert.NotNilf(t, err, "Expected not nil err for SeedData on test '%s'", tt.tn)
			assert.Equalf(t, tt.err, err, "Expected err to match for SeedData on test '%s'", tt.tn)
		}

		if tt.clean {
			cleanTests(ctx, db)
		}
		// Reset the cMock calls
		cMock.Calls = make([]mock.Call, 0)
		cMock.ExpectedCalls = make([]*mock.Call, 0)
	}
}

func findExampleModel(arr []interface{}, num int) *exampleModel {
	for _, i := range arr {
		if item, ok := i.(*exampleModel); ok {
			if item.Num == num {
				return item
			}
		}
	}
	return nil
}

func findExampleModelStrID(arr []interface{}, id string) *exampleModelStrID {
	for _, i := range arr {
		if item, ok := i.(*exampleModelStrID); ok {
			if item.ID == id {
				return item
			}
		}
	}
	return nil
}

func cleanTests(ctx context.Context, db base.MongoDB) {
	c := db.Collection("test")
	c.DeleteMany(ctx, nil)
	c = db.Collection("test2")
	c.DeleteMany(ctx, nil)
}
