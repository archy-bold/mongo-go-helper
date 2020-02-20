// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import (
	context "context"

	base "github.com/archy-bold/mongo-go-helper/base"

	mock "github.com/stretchr/testify/mock"

	mongo "go.mongodb.org/mongo-driver/mongo"

	pagination "github.com/archy-bold/mongo-go-helper/pagination"

	primitive "go.mongodb.org/mongo-driver/bson/primitive"
)

// MongoHelper is an autogenerated mock type for the MongoHelper type
type MongoHelper struct {
	mock.Mock
}

// AddIndexIfNotExists provides a mock function with given fields: ctx, coll, name, keys
func (_m *MongoHelper) AddIndexIfNotExists(ctx context.Context, coll string, name string, keys interface{}) error {
	ret := _m.Called(ctx, coll, name, keys)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, interface{}) error); ok {
		r0 = rf(ctx, coll, name, keys)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Find provides a mock function with given fields: ctx, coll, filter, item, opts
func (_m *MongoHelper) Find(ctx context.Context, coll string, filter interface{}, item interface{}, opts base.FindOptions) (pagination.Result, error) {
	ret := _m.Called(ctx, coll, filter, item, opts)

	var r0 pagination.Result
	if rf, ok := ret.Get(0).(func(context.Context, string, interface{}, interface{}, base.FindOptions) pagination.Result); ok {
		r0 = rf(ctx, coll, filter, item, opts)
	} else {
		r0 = ret.Get(0).(pagination.Result)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, interface{}, interface{}, base.FindOptions) error); ok {
		r1 = rf(ctx, coll, filter, item, opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FindOne provides a mock function with given fields: ctx, coll, filter, item
func (_m *MongoHelper) FindOne(ctx context.Context, coll string, filter interface{}, item interface{}) {
	_m.Called(ctx, coll, filter, item)
}

// GetIDFromInsertOneResult provides a mock function with given fields: _a0
func (_m *MongoHelper) GetIDFromInsertOneResult(_a0 *mongo.InsertOneResult) (primitive.ObjectID, error) {
	ret := _m.Called(_a0)

	var r0 primitive.ObjectID
	if rf, ok := ret.Get(0).(func(*mongo.InsertOneResult) primitive.ObjectID); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(primitive.ObjectID)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*mongo.InsertOneResult) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetIndex provides a mock function with given fields: ctx, coll, index
func (_m *MongoHelper) GetIndex(ctx context.Context, coll string, index string) (*primitive.M, error) {
	ret := _m.Called(ctx, coll, index)

	var r0 *primitive.M
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *primitive.M); ok {
		r0 = rf(ctx, coll, index)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*primitive.M)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, coll, index)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// HasIndex provides a mock function with given fields: ctx, coll, index
func (_m *MongoHelper) HasIndex(ctx context.Context, coll string, index string) (bool, error) {
	ret := _m.Called(ctx, coll, index)

	var r0 bool
	if rf, ok := ret.Get(0).(func(context.Context, string, string) bool); ok {
		r0 = rf(ctx, coll, index)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, coll, index)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// InsertOne provides a mock function with given fields: ctx, coll, item
func (_m *MongoHelper) InsertOne(ctx context.Context, coll string, item interface{}) (primitive.ObjectID, error) {
	ret := _m.Called(ctx, coll, item)

	var r0 primitive.ObjectID
	if rf, ok := ret.Get(0).(func(context.Context, string, interface{}) primitive.ObjectID); ok {
		r0 = rf(ctx, coll, item)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(primitive.ObjectID)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, interface{}) error); ok {
		r1 = rf(ctx, coll, item)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewClient provides a mock function with given fields: uri
func (_m *MongoHelper) NewClient(uri string) (base.MongoClient, error) {
	ret := _m.Called(uri)

	var r0 base.MongoClient
	if rf, ok := ret.Get(0).(func(string) base.MongoClient); ok {
		r0 = rf(uri)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(base.MongoClient)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(uri)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateOne provides a mock function with given fields: ctx, coll, filter, item
func (_m *MongoHelper) UpdateOne(ctx context.Context, coll string, filter interface{}, item interface{}) error {
	ret := _m.Called(ctx, coll, filter, item)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, interface{}, interface{}) error); ok {
		r0 = rf(ctx, coll, filter, item)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
