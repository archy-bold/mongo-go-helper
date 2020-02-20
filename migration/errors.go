package migration

import "errors"

// ErrNoCollection error indicating no collection was supplied
var ErrNoCollection = errors.New("you must specify the collection")

// ErrNoFindFilterFn error indicating no find filter function was supplied
var ErrNoFindFilterFn = errors.New("you must specify a FindFilterFn")

// ErrNoModel error indicating no model was supplied
var ErrNoModel = errors.New("you must specify a Model")
