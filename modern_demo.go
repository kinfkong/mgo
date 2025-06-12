// modern_demo.go - Working demonstration of MongoDB modern driver compatibility wrapper
// This shows how to maintain the mgo API while using the official MongoDB driver

package mgo

import (
	"context"
	"strings"
	"time"

	"github.com/globalsign/mgo/bson"
	officialBson "go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	mongodrv "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// ModernMGO provides the mgo API using the official MongoDB driver
type ModernMGO struct {
	client *mongodrv.Client
	dbName string
}

// ModernDB wraps the modern database
type ModernDB struct {
	mgoDB *mongodrv.Database
	name  string
}

// ModernColl wraps the modern collection
type ModernColl struct {
	mgoColl *mongodrv.Collection
	name    string
}

// ModernQ wraps query state
type ModernQ struct {
	coll   *ModernColl
	filter interface{}
	sort   interface{}
	skip   int64
	limit  int64
}

// ModernIt wraps cursor iteration
type ModernIt struct {
	cursor *mongodrv.Cursor
	ctx    context.Context
	err    error
}

// DialModernMGO creates a new modern MGO session
func DialModernMGO(url string) (*ModernMGO, error) {
	return DialModernMGOWithTimeout(url, 10*time.Second)
}

// DialModernMGOWithTimeout creates a new modern MGO session with timeout
func DialModernMGOWithTimeout(url string, timeout time.Duration) (*ModernMGO, error) {
	clientOpts := options.Client().ApplyURI(url)
	clientOpts.SetConnectTimeout(timeout)
	clientOpts.SetServerSelectionTimeout(timeout)

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	client, err := mongodrv.Connect(ctx, clientOpts)
	if err != nil {
		return nil, err
	}

	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		client.Disconnect(ctx)
		return nil, err
	}

	// Extract database name from URL
	dbName := "test"
	if strings.Contains(url, "/") {
		parts := strings.Split(url, "/")
		if len(parts) > 3 {
			dbPart := parts[3]
			if idx := strings.Index(dbPart, "?"); idx >= 0 {
				dbPart = dbPart[:idx]
			}
			if dbPart != "" {
				dbName = dbPart
			}
		}
	}

	return &ModernMGO{
		client: client,
		dbName: dbName,
	}, nil
}

// Close closes the modern MGO session
func (m *ModernMGO) Close() {
	if m.client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		m.client.Disconnect(ctx)
	}
}

// DB returns a database handle
func (m *ModernMGO) DB(name string) *ModernDB {
	if name == "" {
		name = m.dbName
	}
	return &ModernDB{
		mgoDB: m.client.Database(name),
		name:  name,
	}
}

// Ping tests the connection
func (m *ModernMGO) Ping() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return m.client.Ping(ctx, readpref.Primary())
}

// BuildInfo gets server build information (mgo API compatible)
func (m *ModernMGO) BuildInfo() (BuildInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db := m.client.Database("admin")

	var result struct {
		Version        string `bson:"version"`
		GitVersion     string `bson:"gitVersion"`
		SysInfo        string `bson:"sysInfo"`
		Bits           int    `bson:"bits"`
		Debug          bool   `bson:"debug"`
		MaxObjectSize  int    `bson:"maxBsonObjectSize"`
		VersionArray   []int  `bson:"versionArray"`
		OpenSSLVersion string `bson:"OpenSSLVersion"`
	}

	err := db.RunCommand(ctx, officialBson.D{{Key: "buildInfo", Value: 1}}).Decode(&result)
	if err != nil {
		return BuildInfo{}, err
	}

	return BuildInfo{
		Version:        result.Version,
		GitVersion:     result.GitVersion,
		SysInfo:        result.SysInfo,
		Bits:           result.Bits,
		Debug:          result.Debug,
		MaxObjectSize:  result.MaxObjectSize,
		VersionArray:   result.VersionArray,
		OpenSSLVersion: result.OpenSSLVersion,
	}, nil
}

// C returns a collection handle
func (db *ModernDB) C(name string) *ModernColl {
	return &ModernColl{
		mgoColl: db.mgoDB.Collection(name),
		name:    name,
	}
}

// Insert inserts documents (mgo API compatible)
func (c *ModernColl) Insert(docs ...interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	convertedDocs := make([]interface{}, len(docs))
	for i, doc := range docs {
		convertedDocs[i] = convertMGOToOfficial(doc)
	}

	if len(convertedDocs) == 1 {
		_, err := c.mgoColl.InsertOne(ctx, convertedDocs[0])
		return err
	}
	_, err := c.mgoColl.InsertMany(ctx, convertedDocs)
	return err
}

// Find creates a query (mgo API compatible)
func (c *ModernColl) Find(query interface{}) *ModernQ {
	return &ModernQ{
		coll:   c,
		filter: convertMGOToOfficial(query),
		skip:   0,
		limit:  0,
	}
}

// Count counts documents
func (c *ModernColl) Count() (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	count, err := c.mgoColl.CountDocuments(ctx, officialBson.D{})
	return int(count), err
}

// Remove removes a document
func (c *ModernColl) Remove(selector interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := convertMGOToOfficial(selector)
	_, err := c.mgoColl.DeleteOne(ctx, filter)
	return err
}

// Update updates a document
func (c *ModernColl) Update(selector, update interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := convertMGOToOfficial(selector)
	updateDoc := convertMGOToOfficial(update)

	_, err := c.mgoColl.UpdateOne(ctx, filter, updateDoc)
	return err
}

// EnsureIndex creates an index (mgo API compatible)
func (c *ModernColl) EnsureIndex(index Index) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	keys := officialBson.D{}
	for _, key := range index.Key {
		order := 1
		fieldName := key
		if strings.HasPrefix(key, "-") {
			order = -1
			fieldName = key[1:]
		}
		keys = append(keys, officialBson.E{Key: fieldName, Value: order})
	}

	indexModel := mongodrv.IndexModel{
		Keys: keys,
		Options: &options.IndexOptions{
			Name:       &index.Name,
			Unique:     &index.Unique,
			Background: &index.Background,
			Sparse:     &index.Sparse,
		},
	}

	if index.ExpireAfter > 0 {
		expireAfterSeconds := int32(index.ExpireAfter.Seconds())
		indexModel.Options.ExpireAfterSeconds = &expireAfterSeconds
	}

	_, err := c.mgoColl.Indexes().CreateOne(ctx, indexModel)
	return err
}

// DropCollection drops the collection
func (c *ModernColl) DropCollection() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return c.mgoColl.Drop(ctx)
}

// One finds one document (mgo API compatible)
func (q *ModernQ) One(result interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	findOpts := &options.FindOneOptions{}
	if q.sort != nil {
		findOpts.Sort = q.sort
	}
	if q.skip > 0 {
		findOpts.Skip = &q.skip
	}

	singleResult := q.coll.mgoColl.FindOne(ctx, q.filter, findOpts)
	if singleResult.Err() != nil {
		if singleResult.Err() == mongodrv.ErrNoDocuments {
			return ErrNotFound
		}
		return singleResult.Err()
	}

	var doc officialBson.M
	err := singleResult.Decode(&doc)
	if err != nil {
		return err
	}

	converted := convertOfficialToMGO(doc)
	return mapStructToInterface(converted, result)
}

// All finds all documents
func (q *ModernQ) All(result interface{}) error {
	iter := q.Iter()
	defer iter.Close()
	return iter.All(result)
}

// Count counts query results
func (q *ModernQ) Count() (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	opts := &options.CountOptions{}
	if q.skip > 0 {
		opts.Skip = &q.skip
	}
	if q.limit > 0 {
		opts.Limit = &q.limit
	}

	count, err := q.coll.mgoColl.CountDocuments(ctx, q.filter, opts)
	return int(count), err
}

// Iter returns an iterator
func (q *ModernQ) Iter() *ModernIt {
	ctx := context.Background()

	findOpts := &options.FindOptions{}
	if q.sort != nil {
		findOpts.Sort = q.sort
	}
	if q.skip > 0 {
		findOpts.Skip = &q.skip
	}
	if q.limit > 0 {
		findOpts.Limit = &q.limit
	}

	cursor, err := q.coll.mgoColl.Find(ctx, q.filter, findOpts)

	return &ModernIt{
		cursor: cursor,
		ctx:    ctx,
		err:    err,
	}
}

// Sort sets sort order
func (q *ModernQ) Sort(fields ...string) *ModernQ {
	sort := officialBson.D{}
	for _, field := range fields {
		order := 1
		if strings.HasPrefix(field, "-") {
			order = -1
			field = field[1:]
		}
		sort = append(sort, officialBson.E{Key: field, Value: order})
	}
	q.sort = sort
	return q
}

// Limit sets query limit
func (q *ModernQ) Limit(n int) *ModernQ {
	q.limit = int64(n)
	return q
}

// Skip sets query skip
func (q *ModernQ) Skip(n int) *ModernQ {
	q.skip = int64(n)
	return q
}

// Next gets next document from iterator
func (it *ModernIt) Next(result interface{}) bool {
	if it.err != nil {
		return false
	}

	if !it.cursor.Next(it.ctx) {
		// Check if there was an actual error, or just end of cursor
		it.err = it.cursor.Err()
		// Don't set ErrNotFound here - end of iteration is normal
		return false
	}

	var doc officialBson.M
	err := it.cursor.Decode(&doc)
	if err != nil {
		it.err = err
		return false
	}

	converted := convertOfficialToMGO(doc)
	it.err = mapStructToInterface(converted, result)
	return it.err == nil
}

// Close closes the iterator
func (it *ModernIt) Close() error {
	if it.cursor != nil {
		err := it.cursor.Close(it.ctx)
		if err != nil && it.err == nil {
			it.err = err
		}
	}
	return it.err
}

// All gets all documents from iterator
func (it *ModernIt) All(result interface{}) error {
	if it.err != nil {
		return it.err
	}

	var docs []officialBson.M
	err := it.cursor.All(it.ctx, &docs)
	if err != nil {
		it.err = err
		return err
	}

	converted := make([]interface{}, len(docs))
	for idx, doc := range docs {
		converted[idx] = convertOfficialToMGO(doc)
	}

	return mapStructToInterface(converted, result)
}

// Conversion helpers
func convertMGOToOfficial(input interface{}) interface{} {
	if input == nil {
		return nil
	}

	switch v := input.(type) {
	case bson.M:
		result := officialBson.M{}
		for key, value := range v {
			result[key] = convertMGOToOfficial(value)
		}
		return result
	case bson.D:
		result := officialBson.D{}
		for _, elem := range v {
			result = append(result, officialBson.E{
				Key:   elem.Name,
				Value: convertMGOToOfficial(elem.Value),
			})
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = convertMGOToOfficial(item)
		}
		return result
	case bson.ObjectId:
		if len(v) == 12 {
			objID := primitive.ObjectID{}
			copy(objID[:], []byte(v))
			return objID
		}
		return v
	default:
		return v
	}
}

func convertOfficialToMGO(input interface{}) interface{} {
	if input == nil {
		return nil
	}

	switch v := input.(type) {
	case officialBson.M:
		result := bson.M{}
		for key, value := range v {
			result[key] = convertOfficialToMGO(value)
		}
		return result
	case officialBson.D:
		result := bson.D{}
		for _, elem := range v {
			result = append(result, bson.DocElem{
				Name:  elem.Key,
				Value: convertOfficialToMGO(elem.Value),
			})
		}
		return result
	case []interface{}:
		result := make([]interface{}, len(v))
		for i, item := range v {
			result[i] = convertOfficialToMGO(item)
		}
		return result
	case primitive.ObjectID:
		return bson.ObjectId(v[:])
	default:
		return v
	}
}

func mapStructToInterface(src, dst interface{}) error {
	// Handle slice conversion specifically
	if srcSlice, ok := src.([]interface{}); ok {
		// This is for the All() method - convert slice of interfaces to proper slice
		data, err := bson.Marshal(srcSlice)
		if err != nil {
			return err
		}
		return bson.Unmarshal(data, dst)
	}

	// Handle single document conversion
	data, err := bson.Marshal(src)
	if err != nil {
		return err
	}
	return bson.Unmarshal(data, dst)
}
