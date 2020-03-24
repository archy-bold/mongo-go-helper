package base

import (
	"context"
	"testing"

	"github.com/archy-bold/mongo-go-helper/pagination"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type exampleSubStruct struct {
	TypeID string `bson:"type_id"`
}

type exampleStruct struct {
	ID  primitive.ObjectID `bson:"_id, omitempty"`
	Str string             `bson:"str"`
	Num int                `bson:"num"`
	Sub exampleSubStruct   `bson:"sub"`
}

var (
	sampleObjectID       = "5bb1deea71615727e4ae5e26"
	testDB               = "db-test"
	defaultURI           = "mongodb://localhost:27017"
	sampleObjectIDObj, _ = primitive.ObjectIDFromHex(sampleObjectID)
)

var newClientTests = map[string]struct {
	uri string
	err string
}{
	"success": {defaultURI, ""},
	// "with credentials": {"mongodb://user:pass@localhost:27017", ""},
	"error": {"bad-uri", "error parsing uri: scheme must be \"mongodb\" or \"mongodb+srv\""},
}

func Test_NewClient(t *testing.T) {
	h := &helper{nil}

	for tn, tt := range newClientTests {
		client, err := h.NewClient(tt.uri)

		// Assert everything is as expected.
		if tt.err == "" {
			assert.NotNilf(t, client, "Expected not nil client for NewClient on test '%s'", tn)
			assert.Nilf(t, err, "Expected nil err for NewClient on test '%s'", tn)
		} else {
			assert.Nilf(t, client, "Expected nil client for NewClient on test '%s'", tn)
			assert.NotNilf(t, err, "Expected not nil err for NewClient on test '%s'", tn)
			assert.Equalf(t, tt.err, err.Error(), "Expected err to match expected on test '%s'", tn)
		}
	}
}

var findOneTests = map[string]struct {
	filter  interface{}
	table   string
	matches bool
}{
	"by _id - matches":   {nil, "test", true},
	"by str - matches":   {&bson.M{"num": 999}, "test", true},
	"by random objectid": {&bson.M{"_id": primitive.NewObjectID()}, "test", false},
	"error":              {nil, "!$%^", false},
}

func Test_FindOne(t *testing.T) {
	ctx := context.Background()
	db, _ := setupTests(ctx)
	defer cleanTests(ctx, db)
	// Create the helper
	var h MongoHelper
	h = &helper{db}
	// Insert one row into test
	oid := primitive.NewObjectID()
	h.InsertOne(ctx, "test", exampleStruct{oid, "test string", 999, exampleSubStruct{}})

	for tn, tt := range findOneTests {
		filter := tt.filter
		// Set the filter to find by ID if required.
		if filter == nil {
			filter = bson.M{"_id": oid}
		}
		var item exampleStruct
		h.FindOne(ctx, tt.table, filter, &item)

		// Assert everything is as expected.
		if tt.matches {
			assert.Equalf(t, oid, item.ID, "Expected result to match for FindOne on test '%s'", tn)
			assert.Equalf(t, 999, item.Num, "Expected result to match for FindOne on test '%s'", tn)
			assert.Equalf(t, "test string", item.Str, "Expected result to match for FindOne on test '%s'", tn)
		} else {
			assert.Equalf(t, primitive.NilObjectID, item.ID, "Expected nil objectID for FindOne on test '%s'", tn)
		}
	}
}

var findTests = map[string]struct {
	filter  interface{}
	table   string
	opts    FindOptions
	matches []int
	res     pagination.Result
	err     string
}{
	"by _id - matches 1": {"objectid", "test", FindOptions{}, []int{0}, pagination.Result{Total: 1}, ""},
	"by str - matches 1": {&bson.M{"str": "test"}, "test", FindOptions{}, []int{1}, pagination.Result{Total: 1}, ""},
	"by num - matches 2": {&bson.M{"num": 999}, "test", FindOptions{}, []int{0, 1}, pagination.Result{Total: 2}, ""},
	"by sub - matches 0": {
		&bson.D{{Key: "sub.type_id", Value: bson.M{"$in": bson.A{"type_3"}}}},
		"test", FindOptions{}, []int{}, pagination.Result{}, "",
	},
	"by sub - matches 1": {
		&bson.D{{Key: "sub.type_id", Value: bson.M{"$in": bson.A{"type_2"}}}},
		"test", FindOptions{}, []int{1}, pagination.Result{Total: 1}, "",
	},
	"by sub - matches 2": {
		&bson.D{{Key: "sub.type_id", Value: bson.M{"$in": bson.A{"type_2", "type_1"}}}},
		"test", FindOptions{}, []int{0, 1}, pagination.Result{Total: 2}, "",
	},
	"empty doc - matches all": {&bson.M{}, "test", FindOptions{}, []int{0, 1}, pagination.Result{Total: 2}, ""},
	"paginate - first":        {&bson.M{}, "test", FindOptions{PageSize: 1}, []int{0}, pagination.Result{CurrentPage: 1, PageSize: 1, NumberOfPages: 2, Total: 2}, ""},
	"paginate - second":       {&bson.M{}, "test", FindOptions{Page: 2, PageSize: 1}, []int{1}, pagination.Result{CurrentPage: 2, PageSize: 1, NumberOfPages: 2, Total: 2}, ""},
	"paginate - none":         {&bson.M{}, "test", FindOptions{Page: 3, PageSize: 1}, []int{}, pagination.Result{CurrentPage: 3, PageSize: 1, NumberOfPages: 2, Total: 2}, ""},
	"paginate - all":          {&bson.M{}, "test", FindOptions{PageSize: 2}, []int{0, 1}, pagination.Result{CurrentPage: 1, PageSize: 2, NumberOfPages: 1, Total: 2}, ""},
	"paginate - no matches":   {&bson.M{"str": "blah"}, "test", FindOptions{PageSize: 10}, []int{}, pagination.Result{CurrentPage: 1, PageSize: 10, NumberOfPages: 0, Total: 0}, ""},
	"by random objectid":      {&bson.M{"_id": primitive.NewObjectID()}, "test", FindOptions{}, []int{}, pagination.Result{}, ""},
	"sorted":                  {&bson.M{}, "test", FindOptions{Sorting: map[string]interface{}{"sub.type_id": -1}}, []int{1, 0}, pagination.Result{Total: 2}, ""},
	"error":                   {"filter", "test", FindOptions{}, nil, pagination.Result{}, "failed count on Find: cannot transform type string to a BSON Document: WriteString can only write while positioned on a Element or Value but is positioned on a TopLevel"},
}

func Test_Find(t *testing.T) {
	ctx := context.Background()
	db, _ := setupTests(ctx)
	defer cleanTests(ctx, db)
	// Create the helper
	var h MongoHelper
	h = &helper{db}
	// Insert two rows into test
	item1 := exampleStruct{primitive.NewObjectID(), "test string", 999, exampleSubStruct{"type_1"}}
	h.InsertOne(ctx, "test", item1)
	item2 := exampleStruct{primitive.NewObjectID(), "test", 999, exampleSubStruct{"type_2"}}
	h.InsertOne(ctx, "test", item2)

	for tn, tt := range findTests {
		filter := tt.filter
		// Set the filter to find by ID if required.
		if filter == "objectid" {
			filter = &bson.M{"_id": item1.ID}
		}
		var item exampleStruct
		res, err := h.Find(ctx, tt.table, filter, item, tt.opts)

		// Assert everything is as expected.
		if tt.matches != nil {
			assert.Nilf(t, err, "Expected nil err for Find on test '%s'", tn)
			assert.Lenf(t, res.Items, len(tt.matches), "Expected length of items to match the expected length for Find on test '%s'", tn)
			for i, r := range res.Items {
				item := r.(*exampleStruct)
				// Assert the ID
				assert.Falsef(t, item.ID.IsZero(), "Expected item %d to have an ID for Find on test '%s'", i, tn)

				// And the whole item
				var expected *exampleStruct
				if tt.matches[i] == 0 {
					expected = &item1
				} else {
					expected = &item2
				}
				assert.Equalf(t, expected, item, "Expected item %d to match for Find on test '%s'", i, tn)
			}
			assert.Equalf(t, tt.res.Total, res.Total, "Expected Total to match on res for Find on test '%s'", tn)
			assert.Equalf(t, tt.res.CurrentPage, res.CurrentPage, "Expected CurrentPage to match on res for Find on test '%s'", tn)
			assert.Equalf(t, tt.res.PageSize, res.PageSize, "Expected PageSize to match on res for Find on test '%s'", tn)
			assert.Equalf(t, tt.res.NumberOfPages, res.NumberOfPages, "Expected NumberOfPages to match on res for Find on test '%s'", tn)
		} else {
			assert.NotNilf(t, err, "Expected not nil err for Find on test '%s'", tn)
			assert.Equalf(t, tt.err, err.Error(), "Expected err to match expected on test '%s'", tn)
		}
	}
}

var insertOneTests = map[string]struct {
	item  interface{}
	table string
	err   string
}{
	"empty struct":  {exampleStruct{}, "test", ""},
	"filled struct": {exampleStruct{primitive.NewObjectID(), "string", 1000, exampleSubStruct{}}, "test", ""},
	"nil":           {"string", "test", "cannot transform type string to a BSON Document: WriteString can only write while positioned on a Element or Value but is positioned on a TopLevel"},
}

func Test_InsertOne(t *testing.T) {
	ctx := context.Background()
	db, _ := setupTests(ctx)
	defer cleanTests(ctx, db)
	// Create the helper
	var h MongoHelper
	h = &helper{db}

	for tn, tt := range insertOneTests {
		var itemArg exampleStruct
		var ok bool
		var oid primitive.ObjectID
		var err error
		if itemArg, ok = tt.item.(exampleStruct); ok && itemArg.ID.IsZero() {
			itemArg.ID = primitive.NewObjectID()
			oid, err = h.InsertOne(ctx, tt.table, itemArg)
		} else {
			oid, err = h.InsertOne(ctx, tt.table, tt.item)
		}

		// Try to get the inserted item for assertions
		filter := &bson.M{"_id": oid}
		var item exampleStruct
		c := db.Collection(tt.table)
		res := c.FindOne(ctx, filter)
		res.Decode(&item)

		// Assert everything is as expected.
		if tt.err == "" {
			assert.Falsef(t, oid.IsZero(), "Expected item ID to not be zero for InsertOne on test '%s'", tn)
			assert.Equalf(t, itemArg.ID, oid, "Expected item IDs to match for InsertOne on test '%s'", tn)

			assert.Nilf(t, err, "Expected nil err for InsertOne on test '%s'", tn)
			assert.Equalf(t, oid.String(), itemArg.ID.String(), "Expected item ID to match for InsertOne on test '%s'", tn)
			assert.Equalf(t, itemArg.Str, item.Str, tn, "Expected property 'str' to match for InsertOne on test '%s'", tn)
			assert.Equalf(t, itemArg.Num, item.Num, tn, "Expected property 'num' to match for InsertOne on test '%s'", tn)
		} else {
			assert.NotNilf(t, err, "Expected not nil err for InsertOne on test '%s'", tn)
			assert.Equalf(t, tt.err, err.Error(), "Expected err to match expected on test '%s'", tn)
			assert.NotNilf(t, res.Decode(&item), "Expected collection to be empty on test '%s'", tn)
		}
	}
}

var getIDFromInsertOneResultTests = map[string]struct {
	res interface{}
	id  string
	err error
}{
	"success":            {sampleObjectID, sampleObjectID, nil},
	"not objectid":       {&bson.E{Key: "_id", Value: "string id"}, "", nil},
	"error empty result": {nil, "", ErrUnexpectedInsertResult},
}

func Test_GetIDFromInsertOneResult(t *testing.T) {
	// Create the helper
	var h MongoHelper
	h = &helper{}

	for tn, tt := range getIDFromInsertOneResultTests {
		// Set up the result
		res := &mongo.InsertOneResult{}
		if id, ok := tt.res.(string); ok {
			oid, _ := primitive.ObjectIDFromHex(id)
			res = &mongo.InsertOneResult{InsertedID: oid}
		} else if el, ok := tt.res.(*bson.E); ok {
			res = &mongo.InsertOneResult{InsertedID: el}
		}

		id, err := h.GetIDFromInsertOneResult(res)

		// Do the assertions
		if tt.err == nil {
			expected, _ := primitive.ObjectIDFromHex(tt.id)
			assert.Nilf(t, err, "Expected nil err for GetIDFromInsertOneResult on test '%s'", tn)
			assert.Equalf(t, expected, id, "Expected IDs to match for GetIDFromInsertOneResult on test '%s'", tn)
		} else {
			assert.NotNilf(t, err, "Expected not nil err for GetIDFromInsertOneResult on test '%s'", tn)
			assert.Equalf(t, tt.err, err, "Expected errors to be equal for GetIDFromInsertOneResult on test '%s'", tn)
			assert.Truef(t, id.IsZero(), "Expected id to be zero on test '%s'", tn)
		}
	}
}

var updateOneTests = map[string]struct {
	item    interface{}
	table   string
	matches bool
	err     string
}{
	"filled struct":           {exampleStruct{primitive.NewObjectID(), "string", 1000, exampleSubStruct{}}, "test", true, ""},
	"empty struct":            {exampleStruct{}, "test", true, ""},
	"filled struct, no match": {exampleStruct{primitive.NewObjectID(), "string", 1000, exampleSubStruct{}}, "test", false, ErrNoMatches.Error()},
	"error":                   {"string", "test", true, "cannot transform type string to a BSON Document: WriteString can only write while positioned on a Element or Value but is positioned on a TopLevel"},
}

func Test_UpdateOne(t *testing.T) {
	ctx := context.Background()
	db, _ := setupTests(ctx)
	defer cleanTests(ctx, db)
	// Create the helper
	var h MongoHelper
	h = &helper{db}

	// Insert one row into test
	oid := primitive.NewObjectID()
	h.InsertOne(ctx, "test", exampleStruct{oid, "test string", 999, exampleSubStruct{}})

	for tn, tt := range updateOneTests {
		var itemArg exampleStruct
		var ok bool
		var err error
		filter := &bson.M{"_id": primitive.NilObjectID}
		if itemArg, ok = tt.item.(exampleStruct); ok {
			filter = &bson.M{"_id": itemArg.ID}
			if tt.matches {
				filter = &bson.M{"_id": oid}
				itemArg.ID = oid
			}
			err = h.UpdateOne(ctx, tt.table, filter, itemArg)
		} else {
			err = h.UpdateOne(ctx, tt.table, filter, tt.item)
		}

		var item exampleStruct
		h.FindOne(ctx, tt.table, filter, &item)

		// Assert everything is as expected.
		if tt.err == "" {
			assert.Falsef(t, item.ID.IsZero(), "Expected item ID to not be zero for UpdateOne on test '%s'", tn)

			assert.Nilf(t, err, "Expected nil err for UpdateOne on test '%s'", tn)
			assert.Equalf(t, oid.String(), item.ID.String(), "Expected item ID to match for UpdateOne on test '%s'", tn)
			assert.Equalf(t, itemArg.Str, item.Str, tn, "Expected property 'str' to match for UpdateOne on test '%s'", tn)
			assert.Equalf(t, itemArg.Num, item.Num, tn, "Expected property 'num' to match for UpdateOne on test '%s'", tn)
		} else {
			assert.NotNilf(t, err, "Expected not nil err for UpdateOne on test '%s'", tn)
			assert.Equalf(t, tt.err, err.Error(), "Expected err to match expected on test '%s'", tn)
			assert.Truef(t, item.ID.IsZero(), "Expected item ID to be zero for UpdateOne on test '%s'", tn)
		}
	}
}

var getIndexTests = map[string]struct {
	uri         string
	collection  string
	index       string
	expectedKey *bson.M
	err         string
}{
	"success":            {"default", "test", "test_index", &bson.M{"foo": int32(-1)}, ""},
	"non-existent index": {"default", "test", "not_there", (*bson.M)(nil), ""},
	"non-existent table": {"default", "tests", "test_index", (*bson.M)(nil), ""},
	"error":              {"mongodb://127.0.0.1:27018", "test", "test_index", (*bson.M)(nil), "(InvalidNamespace) Invalid collection name specified 'db-test.$\\\"'"},
}

func Test_GetIndex(t *testing.T) {
	ctx := context.Background()

	for tn, tt := range getIndexTests {
		// Set up the client
		var db MongoDB
		if tt.uri == "default" {
			db, _ = setupTests(ctx)
		} else {
			opt := options.Client()
			opt.ApplyURI(tt.uri)
			client, _ := mongo.NewClient(opt)
			db = client.Database(testDB)
		}
		defer cleanTests(ctx, db)
		// Create the helper
		var h MongoHelper
		h = &helper{db}

		res, err := h.GetIndex(ctx, tt.collection, tt.index)

		// Do the assertions
		if tt.err == "" {
			assert.Nilf(t, err, "Expected nil err for GetIndex on test '%s'", tn)
			if tt.expectedKey == nil {
				assert.Nilf(t, res, "Expected nil index for GetIndex on test '%s'", tn)
			} else {
				key := (*res)["key"]
				name := (*res)["name"]
				assert.Equalf(t, *tt.expectedKey, key, "Expected key to equal expected on test '%s'", tn)
				assert.Equalf(t, tt.index, name, "Expected name to equal expected on test '%s'", tn)
			}
		} else {
			assert.Errorf(t, err, "Expected not nil err for GetIndex on test '%s'", tn)
			// assert.NotNilf(t, err, "Expected not nil err for GetIndex on test '%s'", tn)
			// assert.Equalf(t, tt.err, err.Error(), "Expected err ('%s') to equal '%s' on test '%s'", err.Error(), tt.err, tn)
			assert.Equalf(t, tt.expectedKey, res, "Expected not nil repo for GetIndex on test '%s'", tn)
		}
	}
}

var hasIndexTests = map[string]struct {
	uri        string
	collection string
	index      string
	expected   bool
	err        string
}{
	"success":            {"default", "test", "test_index", true, ""},
	"non-existent index": {"default", "test", "not_there", false, ""},
	"non-existent table": {"default", "testw", "test_index", false, ""},
	"error":              {"mongodb://127.0.0.1:27018", "test", "test_index", false, "client is disconnected"},
}

func Test_HasIndex(t *testing.T) {
	ctx := context.Background()

	for tn, tt := range hasIndexTests {
		// Set up the client
		var db MongoDB
		if tt.uri == "default" {
			db, _ = setupTests(ctx)
		} else {
			opt := options.Client()
			opt.ApplyURI(tt.uri)
			client, _ := mongo.NewClient(opt)
			db = client.Database(testDB)
		}
		defer cleanTests(ctx, db)
		// Create the helper
		var h MongoHelper
		h = &helper{db}

		res, err := h.HasIndex(ctx, tt.collection, tt.index)

		// Do the assertions
		if tt.err == "" {
			assert.Nilf(t, err, "Expected nil err for HasIndex on test '%s'", tn)
			assert.Equalf(t, tt.expected, res, "Expected not nil repo for HasIndex on test '%s'", tn)
		} else {
			assert.EqualErrorf(t, err, tt.err, "Expected err ('%s') to equal '%s' on test '%s'", err.Error(), tt.err, tn)
			// assert.NotNilf(t, err, "Expected not nil err for HasIndex on test '%s'", tn)
			// assert.Equalf(t, tt.err, err.Error(), "Expected err ('%s') to equal '%s' on test '%s'", err.Error(), tt.err, tn)
			assert.Equalf(t, tt.expected, res, "Expected not nil repo for HasIndex on test '%s'", tn)
		}
	}
}

var addIndexIfNotExistsTests = []struct {
	test          string
	uri           string
	index         string
	err           string
	existsAlready bool
}{
	{"index exists already", "default", "test_index", "", true},
	{"new index added", "default", "new_index", "", false},
	{"new index exists", "default", "new_index", "", true},
	{"error", "mongodb://127.0.0.1:27018", "test_index", "client is disconnected", false},
}

func Test_AddIndexIfNotExists(t *testing.T) {
	ctx := context.Background()
	defaultDb, _ := setupTests(ctx)
	defer cleanTests(ctx, defaultDb)

	for _, tt := range addIndexIfNotExistsTests {
		// Set up the client
		var db MongoDB
		if tt.uri == "default" {
			db = defaultDb
		} else {
			opt := options.Client()
			opt.ApplyURI(tt.uri)
			client, _ := mongo.NewClient(opt)
			db = client.Database(testDB)
		}
		defer cleanTests(ctx, db)
		// Create the helper
		var h MongoHelper
		h = &helper{db}

		if tt.existsAlready {
			HasIndex, _ := h.HasIndex(ctx, "test", tt.index)
			assert.Truef(t, HasIndex, "Expected index '%s' to already exist for test '%s'", tt.index, tt.test)
		}

		err := h.AddIndexIfNotExists(ctx, "test", tt.index, &bson.M{"bar": -1})

		// Do the assertions
		if tt.err == "" {
			assert.Nilf(t, err, "Expected nil err for HasIndex on test '%s'", tt.test)
			HasIndex, _ := h.HasIndex(ctx, "test", tt.index)
			assert.Truef(t, HasIndex, "Expected index '%s' to exist for test '%s'", tt.index, tt.test)
		} else {
			assert.EqualErrorf(t, err, tt.err, "Expected err ('%s') to equal '%s' on test '%s'", err.Error(), tt.err, tt.test)
			// assert.NotNilf(t, err, "Expected not nil err for HasIndex on test '%s'", tn)
			// assert.Equalf(t, tt.err, err.Error(), "Expected err ('%s') to equal '%s' on test '%s'", err.Error(), tt.err, tn)
		}
	}
}

var aggregateTests = map[string]struct {
	pipeline mongo.Pipeline
	expected []bson.M
	err      string
}{
	"successful calle to aggregate": {
		pipeline: mongo.Pipeline{
			bson.D{{Key: "$match", Value: bson.D{{Key: "num", Value: 999}}}},
		},
		expected: []bson.M{{"_id": sampleObjectIDObj, "str": "test string", "num": int32(999), "sub": bson.M{"type_id": ""}}},
	},
	"aggregate error": {
		pipeline: mongo.Pipeline{
			bson.D{{Key: "$notafunc", Value: bson.D{}}},
		},
		err: "(Location40324) Unrecognized pipeline stage name: '$notafunc'",
	},
}

func Test_Aggregate(t *testing.T) {
	ctx := context.Background()
	db, _ := setupTests(ctx)
	defer cleanTests(ctx, db)
	// Create the helper
	var h MongoHelper
	h = &helper{db}
	// Insert one row into test
	h.InsertOne(ctx, "test", exampleStruct{sampleObjectIDObj, "test string", 999, exampleSubStruct{}})

	for tn, tt := range aggregateTests {
		ret, err := h.Aggregate(ctx, "test", tt.pipeline)

		if tt.err == "" {
			assert.Nilf(t, err, "Expected nil err for Aggregate on test '%s'", tn)
			assert.Equalf(t, tt.expected, ret, "Expected ret to match for test '%s'", tn)
		} else {
			assert.EqualErrorf(t, err, tt.err, "Expected err ('%s') to equal '%s' on test '%s'", err.Error(), tt.err, tn)
			assert.Nilf(t, ret, "Expected err ('%s') to equal '%s' on test '%s'", err.Error(), tt.err, tn)
		}
	}
}

func Test_NewHelper(t *testing.T) {
	db := &mongo.Database{}
	h := NewHelper(db)

	assert.NotNil(t, h)
	assert.IsType(t, &helper{}, h)
	assert.Equal(t, db, h.(*helper).db)
}

func setupTests(ctx context.Context) (MongoDB, *mongo.Client) {
	// Connect to the mongo database
	cop := options.Client()
	cop.ApplyURI(defaultURI)
	client, _ := mongo.NewClient(cop)
	var db MongoDB
	db = client.Database(testDB)
	client.Connect(ctx)
	// Create an index
	c := db.Collection("test")
	iv := c.Indexes()
	iop := options.Index()
	iop.SetName("test_index")
	iv.CreateOne(
		ctx,
		mongo.IndexModel{
			Keys: bson.M{
				"foo": -1,
			},
			Options: iop,
		},
	)
	return db, client
}

func cleanTests(ctx context.Context, db MongoDB) {
	c := db.Collection("test")
	c.DeleteMany(ctx, &bson.M{})
	iv := c.Indexes()
	iv.DropOne(ctx, "test_index")
	iv.DropOne(ctx, "not_there")
	iv.DropAll(ctx)
}
