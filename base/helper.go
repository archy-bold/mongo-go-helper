package base

import (
	"context"
	"math"
	"reflect"

	"github.com/archy-bold/mongo-go-helper/pagination"
	"github.com/pkg/errors"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ErrUnexpectedInsertResult indicates that a result set is not a valid insert result
var ErrUnexpectedInsertResult = errors.New("unexpected insert result")

// ErrNoMatches indicates that no matches were found for a query
var ErrNoMatches = errors.New("no matches found for query")

// FindOptions represents Find function options
type FindOptions struct {
	PageSize int64
	Page     int64
	Sorting  map[string]interface{}
}

// MongoHelper is used for helper functions
type MongoHelper interface {
	NewClient(uri string) (MongoClient, error)
	Find(ctx context.Context, coll string, filter interface{}, item interface{}, opts FindOptions) (pagination.Result, error)
	FindOne(ctx context.Context, coll string, filter interface{}, item interface{})
	InsertOne(ctx context.Context, coll string, item interface{}) (primitive.ObjectID, error)
	GetIDFromInsertOneResult(*mongo.InsertOneResult) (primitive.ObjectID, error)
	UpdateOne(ctx context.Context, coll string, filter interface{}, item interface{}) error
	GetIndex(ctx context.Context, coll string, index string) (*bson.M, error)
	HasIndex(ctx context.Context, coll string, index string) (bool, error)
	AddIndexIfNotExists(ctx context.Context, coll string, name string, keys interface{}) error
}

type helper struct {
	db MongoDB
}

func (h *helper) NewClient(uri string) (MongoClient, error) {
	opt := options.Client()
	opt.ApplyURI(uri)
	return mongo.NewClient(opt)
}

func (h *helper) FindOne(ctx context.Context, coll string, filter interface{}, item interface{}) {
	c := h.db.Collection(coll)
	doc := c.FindOne(ctx, filter)
	doc.Decode(item)
}

func (h *helper) Find(ctx context.Context, coll string, filter interface{}, item interface{}, opts FindOptions) (res pagination.Result, err error) {
	c := h.db.Collection(coll)

	// Get the count for the result
	var count int64
	count, err = c.CountDocuments(ctx, filter)

	if err != nil {
		err = errors.Wrap(err, "failed count on Find")
		return
	}
	res.Total = int32(count)

	// Set the result based on the pagination
	if opts.PageSize > 0 {
		res.PageSize = int32(opts.PageSize)
		res.NumberOfPages = int32(math.Ceil(float64(count) / float64(opts.PageSize)))
		res.CurrentPage = 1
		if opts.Page > 1 {
			res.CurrentPage = int32(opts.Page)
		}
	}
	mOpts := h.convertFindOpts(opts)
	cur, err := c.Find(ctx, filter, mOpts)

	if err != nil {
		err = errors.Wrap(err, "failed on Find")
		return
	}

	defer cur.Close(ctx)

	res.Items = make([]interface{}, 0)
	for cur.Next(ctx) {
		vp := reflect.New(reflect.TypeOf(item))
		newItem := vp.Interface()
		err = cur.Decode(newItem)
		res.Items = append(res.Items, newItem)
	}

	return
}

func (h *helper) InsertOne(ctx context.Context, coll string, item interface{}) (primitive.ObjectID, error) {
	var oid primitive.ObjectID

	c := h.db.Collection(coll)
	res, err := c.InsertOne(ctx, item)

	if err == nil {
		oid, err = h.GetIDFromInsertOneResult(res)
	}

	return oid, err
}

func (h *helper) GetIDFromInsertOneResult(res *mongo.InsertOneResult) (primitive.ObjectID, error) {
	if oid, ok := res.InsertedID.(primitive.ObjectID); ok {
		return oid, nil
	} else if res.InsertedID != nil {
		return primitive.NilObjectID, nil
	}
	return primitive.NilObjectID, ErrUnexpectedInsertResult
}

func (h *helper) UpdateOne(ctx context.Context, coll string, filter interface{}, item interface{}) error {
	c := h.db.Collection(coll)
	res, err := c.ReplaceOne(ctx, filter, item)

	if err == nil && res.MatchedCount == 0 {
		return ErrNoMatches
	}

	return err
}

func (h *helper) GetIndex(ctx context.Context, coll string, index string) (*bson.M, error) {
	var cur *mongo.Cursor
	var err error

	c := h.db.Collection(coll)
	iv := c.Indexes()

	// Get the current indexes
	cur, err = iv.List(ctx)

	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		elem := &bson.M{}
		err = cur.Decode(elem)

		// Check the value of the name element
		if ni, ok := (*elem)["name"]; ok {
			if name, ok := ni.(string); ok && name == index {
				return elem, err
			}
		}
	}
	return nil, err
}

func (h *helper) HasIndex(ctx context.Context, coll string, index string) (bool, error) {
	i, err := h.GetIndex(ctx, coll, index)
	if err != nil {
		return false, err
	}
	return i != nil, nil
}

func (h *helper) AddIndexIfNotExists(ctx context.Context, coll string, name string, keys interface{}) error {
	res, err := h.HasIndex(ctx, coll, name)

	if err != nil {
		return err
	}

	if !res {
		c := h.db.Collection(coll)
		iv := c.Indexes()

		opt := options.Index()
		opt.SetName(name)
		iv.CreateOne(ctx, mongo.IndexModel{
			Keys:    keys,
			Options: opt,
		})
	}
	return nil
}

func (h *helper) convertFindOpts(opts FindOptions) *options.FindOptions {
	fo := options.Find()
	// Pagination options
	if opts.PageSize > 0 {
		fo.SetLimit(opts.PageSize)
		if opts.Page > 1 {
			fo.SetSkip((opts.Page - 1) * opts.PageSize)
		}
	}
	// Sorting options
	if len(opts.Sorting) > 0 {
		fo.SetSort(opts.Sorting)
	}
	return fo
}

// NewHelper get an implementation of a MongoHelper
func NewHelper(db MongoDB) MongoHelper {
	return &helper{db}
}
