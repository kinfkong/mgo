// modern_demo.go - Working demonstration of MongoDB modern driver compatibility wrapper
// This shows how to maintain the mgo API while using the official MongoDB driver

package mgo

import (
	"context"
	"reflect"
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
	client     *mongodrv.Client
	dbName     string
	mode       Mode
	safe       *Safe
	isOriginal bool // Track if this is the original session or a copy
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
	coll       *ModernColl
	filter     interface{}
	sort       interface{}
	skip       int64
	limit      int64
	projection interface{}
}

// ModernIt wraps cursor iteration
type ModernIt struct {
	cursor *mongodrv.Cursor
	ctx    context.Context
	err    error
}

// ModernPipe wraps aggregation pipeline state
type ModernPipe struct {
	collection *ModernColl
	pipeline   interface{}
	allowDisk  bool
	batchSize  int32
	maxTimeMS  int64
	collation  *options.Collation
}

// ModernBulk provides bulk operations using the official MongoDB driver
type ModernBulk struct {
	collection *ModernColl
	operations []mongodrv.WriteModel
	ordered    bool
	opcount    int
}

// DialModernMGO connects to MongoDB using the official driver but provides mgo API (mgo API compatible)
func DialModernMGO(url string) (*ModernMGO, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongodrv.Connect(ctx, options.Client().ApplyURI(url))
	if err != nil {
		return nil, err
	}

	// Parse database name from URL - simple parsing for common cases
	dbName := "test" // Default database name

	return &ModernMGO{
		client: client,
		dbName: dbName,
		mode:   Primary,
		safe: &Safe{
			W:        1,
			WTimeout: 0,
			FSync:    false,
			J:        false,
		},
		isOriginal: true, // Mark as original session
	}, nil
}

// Close closes the modern MGO session
func (m *ModernMGO) Close() {
	// Only close the client if this is the original session
	if m.isOriginal && m.client != nil {
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
	var filter interface{}
	if query == nil {
		filter = officialBson.D{} // Empty document for "find all"
	} else {
		filter = convertMGOToOfficial(query)
	}

	return &ModernQ{
		coll:   c,
		filter: filter,
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

// Pipe creates an aggregation pipeline (mgo API compatible)
func (c *ModernColl) Pipe(pipeline interface{}) *ModernPipe {
	return &ModernPipe{
		collection: c,
		pipeline:   pipeline,
		allowDisk:  false,
		batchSize:  101, // Default batch size
		maxTimeMS:  0,
		collation:  nil,
	}
}

// One finds one document (mgo API compatible)
func (q *ModernQ) One(result interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	findOpts := &options.FindOneOptions{}
	if q.projection != nil {
		findOpts.Projection = q.projection
	}
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
	if q.projection != nil {
		findOpts.Projection = q.projection
	}
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

// Select sets the fields to select (mgo API compatible)
func (q *ModernQ) Select(selector interface{}) *ModernQ {
	q.projection = convertMGOToOfficial(selector)
	return q
}

// Apply applies a change to a single document and returns the old or new document (mgo API compatible)
func (q *ModernQ) Apply(change Change, result interface{}) (*ChangeInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var updateDoc interface{}

	if change.Remove {
		// For remove operations, use FindOneAndDelete
		deleteOpts := options.FindOneAndDelete()

		singleResult := q.coll.mgoColl.FindOneAndDelete(ctx, q.filter, deleteOpts)
		if singleResult.Err() != nil {
			if singleResult.Err() == mongodrv.ErrNoDocuments {
				return &ChangeInfo{}, ErrNotFound
			}
			return nil, singleResult.Err()
		}

		if result != nil {
			var doc officialBson.M
			err := singleResult.Decode(&doc)
			if err != nil {
				return nil, err
			}
			converted := convertOfficialToMGO(doc)
			err = mapStructToInterface(converted, result)
			if err != nil {
				return nil, err
			}
		}

		return &ChangeInfo{Removed: 1}, nil
	}

	// For update/upsert operations
	updateDoc = convertMGOToOfficial(change.Update)
	updateOpts := options.FindOneAndUpdate()
	updateOpts.SetUpsert(change.Upsert)

	if change.ReturnNew {
		updateOpts.SetReturnDocument(options.After)
	} else {
		updateOpts.SetReturnDocument(options.Before)
	}

	singleResult := q.coll.mgoColl.FindOneAndUpdate(ctx, q.filter, updateDoc, updateOpts)
	if singleResult.Err() != nil {
		if singleResult.Err() == mongodrv.ErrNoDocuments {
			if change.Upsert {
				// Document was upserted but we need to return ChangeInfo
				return &ChangeInfo{Updated: 1}, nil
			}
			return &ChangeInfo{}, ErrNotFound
		}
		return nil, singleResult.Err()
	}

	if result != nil {
		var doc officialBson.M
		err := singleResult.Decode(&doc)
		if err != nil {
			return nil, err
		}
		converted := convertOfficialToMGO(doc)
		err = mapStructToInterface(converted, result)
		if err != nil {
			return nil, err
		}
	}

	return &ChangeInfo{Updated: 1}, nil
}

// Next gets next document from iterator
func (it *ModernIt) Next(result interface{}) bool {
	if it.err != nil {
		return false
	}

	if it.cursor == nil {
		it.err = ErrNotFound
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

	if it.cursor == nil {
		return ErrNotFound
	}

	// Use Next() in a loop to avoid BSON slice unmarshalling issues
	var docs []interface{}

	for {
		var doc bson.M
		if !it.Next(&doc) {
			break
		}
		if it.err != nil {
			return it.err
		}
		docs = append(docs, doc)
	}

	// Check for iteration errors (not end-of-cursor)
	if it.err != nil && it.err != ErrNotFound {
		return it.err
	}

	// Reset error since reaching end of cursor is expected
	it.err = nil

	return mapStructToInterface(docs, result)
}

// Modern implementations of Pipe methods

// Iter executes the aggregation pipeline and returns an iterator
func (p *ModernPipe) Iter() *ModernIt {
	ctx := context.Background()

	// Convert pipeline to the correct format for the official driver
	var pipeline interface{}

	// Handle different pipeline input types
	switch v := p.pipeline.(type) {
	case []interface{}:
		// Already converted, use as-is
		pipeline = v
	case []bson.M:
		// Convert []bson.M to []interface{}
		converted := make([]interface{}, len(v))
		for i, stage := range v {
			converted[i] = convertMGOToOfficial(stage)
		}
		pipeline = converted
	case []officialBson.M:
		// Already in official format
		pipeline = v
	default:
		// Try to convert single stage
		pipeline = []interface{}{convertMGOToOfficial(v)}
	}

	// Create aggregation options
	opts := &options.AggregateOptions{}
	if p.allowDisk {
		opts.AllowDiskUse = &p.allowDisk
	}
	if p.batchSize > 0 {
		opts.BatchSize = &p.batchSize
	}
	if p.maxTimeMS > 0 {
		maxTime := time.Duration(p.maxTimeMS) * time.Millisecond
		opts.MaxTime = &maxTime
	}
	if p.collation != nil {
		opts.Collation = p.collation
	}

	cursor, err := p.collection.mgoColl.Aggregate(ctx, pipeline, opts)

	return &ModernIt{
		cursor: cursor,
		ctx:    ctx,
		err:    err,
	}
}

// All executes the pipeline and returns all results
func (p *ModernPipe) All(result interface{}) error {
	iter := p.Iter()
	defer iter.Close()
	return iter.All(result)
}

// One executes the pipeline and returns the first result
func (p *ModernPipe) One(result interface{}) error {
	iter := p.Iter()
	defer iter.Close()

	if iter.Next(result) {
		return nil
	}
	if err := iter.err; err != nil {
		return err
	}
	return ErrNotFound
}

// Explain returns aggregation execution statistics
func (p *ModernPipe) Explain(result interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Convert pipeline to the correct format
	var pipeline []interface{}

	switch v := p.pipeline.(type) {
	case []interface{}:
		pipeline = v
	case []bson.M:
		pipeline = make([]interface{}, len(v))
		for i, stage := range v {
			pipeline[i] = convertMGOToOfficial(stage)
		}
	case []officialBson.M:
		pipeline = make([]interface{}, len(v))
		for i, stage := range v {
			pipeline[i] = stage
		}
	default:
		pipeline = []interface{}{convertMGOToOfficial(v)}
	}

	// Create explain command
	explainCmd := officialBson.D{
		{Key: "aggregate", Value: p.collection.name},
		{Key: "pipeline", Value: pipeline},
		{Key: "explain", Value: true},
	}

	db := p.collection.mgoColl.Database()
	singleResult := db.RunCommand(ctx, explainCmd)

	var doc officialBson.M
	err := singleResult.Decode(&doc)
	if err != nil {
		return err
	}

	converted := convertOfficialToMGO(doc)
	return mapStructToInterface(converted, result)
}

// AllowDiskUse enables writing to temporary files during aggregation
func (p *ModernPipe) AllowDiskUse() *ModernPipe {
	p.allowDisk = true
	return p
}

// Batch sets the batch size for the aggregation cursor
func (p *ModernPipe) Batch(n int) *ModernPipe {
	p.batchSize = int32(n)
	return p
}

// SetMaxTime sets the maximum execution time for the aggregation
func (p *ModernPipe) SetMaxTime(d time.Duration) *ModernPipe {
	p.maxTimeMS = int64(d / time.Millisecond)
	return p
}

// Collation sets the collation for the aggregation
func (p *ModernPipe) Collation(collation *Collation) *ModernPipe {
	if collation != nil {
		// Convert mgo Collation to official driver Collation
		p.collation = &options.Collation{
			Locale:          collation.Locale,
			CaseFirst:       collation.CaseFirst,
			Strength:        collation.Strength,
			Alternate:       collation.Alternate,
			MaxVariable:     collation.MaxVariable,
			Normalization:   collation.Normalization,
			CaseLevel:       collation.CaseLevel,
			NumericOrdering: collation.NumericOrdering,
			Backwards:       collation.Backwards,
		}
	}
	return p
}

// Run executes a database command on the collection's database (mgo API compatible)
func (c *ModernColl) Run(cmd, result interface{}) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	command := convertMGOToOfficial(cmd)
	singleResult := c.mgoColl.Database().RunCommand(ctx, command)

	var doc officialBson.M
	err := singleResult.Decode(&doc)
	if err != nil {
		return err
	}

	converted := convertOfficialToMGO(doc)
	return mapStructToInterface(converted, result)
}

// Bulk returns a bulk operation builder (mgo API compatible)
func (c *ModernColl) Bulk() *ModernBulk {
	return &ModernBulk{
		collection: c,
		operations: make([]mongodrv.WriteModel, 0),
		ordered:    true,
		opcount:    0,
	}
}

// Unordered puts the bulk operation in unordered mode (mgo API compatible)
func (b *ModernBulk) Unordered() {
	b.ordered = false
}

// Insert queues up documents for insertion (mgo API compatible)
func (b *ModernBulk) Insert(docs ...interface{}) {
	for _, doc := range docs {
		convertedDoc := convertMGOToOfficial(doc)
		insertModel := mongodrv.NewInsertOneModel().SetDocument(convertedDoc)
		b.operations = append(b.operations, insertModel)
		b.opcount++
	}
}

// Update queues up pairs of updating instructions (mgo API compatible)
// Each pair matches exactly one document for updating at most
func (b *ModernBulk) Update(pairs ...interface{}) {
	if len(pairs)%2 != 0 {
		panic("Bulk.Update requires an even number of parameters")
	}

	for i := 0; i < len(pairs); i += 2 {
		selector := pairs[i]
		update := pairs[i+1]

		if selector == nil {
			selector = bson.D{}
		}

		filter := convertMGOToOfficial(selector)
		updateDoc := convertMGOToOfficial(update)

		updateModel := mongodrv.NewUpdateOneModel().SetFilter(filter).SetUpdate(updateDoc)
		b.operations = append(b.operations, updateModel)
		b.opcount++
	}
}

// UpdateAll queues up pairs of updating instructions (mgo API compatible)
// Each pair updates all documents matching the selector
func (b *ModernBulk) UpdateAll(pairs ...interface{}) {
	if len(pairs)%2 != 0 {
		panic("Bulk.UpdateAll requires an even number of parameters")
	}

	for i := 0; i < len(pairs); i += 2 {
		selector := pairs[i]
		update := pairs[i+1]

		if selector == nil {
			selector = bson.D{}
		}

		filter := convertMGOToOfficial(selector)
		updateDoc := convertMGOToOfficial(update)

		updateModel := mongodrv.NewUpdateManyModel().SetFilter(filter).SetUpdate(updateDoc)
		b.operations = append(b.operations, updateModel)
		b.opcount++
	}
}

// Upsert queues up pairs of upserting instructions (mgo API compatible)
// Each pair matches exactly one document for updating at most
func (b *ModernBulk) Upsert(pairs ...interface{}) {
	if len(pairs)%2 != 0 {
		panic("Bulk.Upsert requires an even number of parameters")
	}

	for i := 0; i < len(pairs); i += 2 {
		selector := pairs[i]
		update := pairs[i+1]

		if selector == nil {
			selector = bson.D{}
		}

		filter := convertMGOToOfficial(selector)
		updateDoc := convertMGOToOfficial(update)

		upsert := true
		updateModel := mongodrv.NewUpdateOneModel().SetFilter(filter).SetUpdate(updateDoc).SetUpsert(upsert)
		b.operations = append(b.operations, updateModel)
		b.opcount++
	}
}

// Remove queues up selectors for removing matching documents (mgo API compatible)
// Each selector will remove only a single matching document
func (b *ModernBulk) Remove(selectors ...interface{}) {
	for _, selector := range selectors {
		if selector == nil {
			selector = bson.D{}
		}

		filter := convertMGOToOfficial(selector)
		deleteModel := mongodrv.NewDeleteOneModel().SetFilter(filter)
		b.operations = append(b.operations, deleteModel)
		b.opcount++
	}
}

// RemoveAll queues up selectors for removing all matching documents (mgo API compatible)
// Each selector will remove all matching documents
func (b *ModernBulk) RemoveAll(selectors ...interface{}) {
	for _, selector := range selectors {
		if selector == nil {
			selector = bson.D{}
		}

		filter := convertMGOToOfficial(selector)
		deleteModel := mongodrv.NewDeleteManyModel().SetFilter(filter)
		b.operations = append(b.operations, deleteModel)
		b.opcount++
	}
}

// Run executes all queued bulk operations (mgo API compatible)
func (b *ModernBulk) Run() (*BulkResult, error) {
	if len(b.operations) == 0 {
		return &BulkResult{}, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	opts := options.BulkWrite().SetOrdered(b.ordered)

	result, err := b.collection.mgoColl.BulkWrite(ctx, b.operations, opts)
	if err != nil {
		// Convert bulk write errors to mgo format
		if bulkErr, ok := err.(mongodrv.BulkWriteException); ok {
			return b.convertBulkError(result, &bulkErr)
		}
		return nil, err
	}

	return b.convertBulkResult(result), nil
}

// convertBulkResult converts official driver BulkWriteResult to mgo BulkResult
func (b *ModernBulk) convertBulkResult(result *mongodrv.BulkWriteResult) *BulkResult {
	if result == nil {
		return &BulkResult{}
	}

	// For delete operations, DeletedCount represents both matched and modified
	// For update operations, use MatchedCount and ModifiedCount
	matched := int(result.MatchedCount + result.DeletedCount)
	modified := int(result.ModifiedCount + result.DeletedCount + result.UpsertedCount)

	return &BulkResult{
		Matched:  matched,
		Modified: modified,
	}
}

// convertBulkError converts official driver BulkWriteException to mgo BulkError
func (b *ModernBulk) convertBulkError(result *mongodrv.BulkWriteResult, bulkErr *mongodrv.BulkWriteException) (*BulkResult, error) {
	// Convert write errors to BulkErrorCase format
	var ecases []BulkErrorCase

	for _, writeErr := range bulkErr.WriteErrors {
		ecase := BulkErrorCase{
			Index: writeErr.Index,
			Err: &QueryError{
				Code:    writeErr.Code,
				Message: writeErr.Message,
			},
		}
		ecases = append(ecases, ecase)
	}

	// Handle write concern error if present
	if bulkErr.WriteConcernError != nil {
		ecase := BulkErrorCase{
			Index: -1, // Write concern errors don't have specific indices
			Err: &QueryError{
				Code:    bulkErr.WriteConcernError.Code,
				Message: bulkErr.WriteConcernError.Message,
			},
		}
		ecases = append(ecases, ecase)
	}

	bulkResult := b.convertBulkResult(result)

	if len(ecases) > 0 {
		return bulkResult, &BulkError{ecases: ecases}
	}

	// If we have a bulk write exception but no specific errors, return the general error
	return bulkResult, &BulkError{
		ecases: []BulkErrorCase{{
			Index: -1,
			Err: &QueryError{
				Message: bulkErr.Error(),
			},
		}},
	}
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

// convertSliceWithReflect converts a slice of interfaces to a target slice type using reflection
func convertSliceWithReflect(srcSlice []interface{}, dst interface{}) error {
	dstValue := reflect.ValueOf(dst)
	if dstValue.Kind() != reflect.Ptr {
		return ErrNotFound
	}

	dstSlice := dstValue.Elem()
	if dstSlice.Kind() != reflect.Slice {
		return ErrNotFound
	}

	elementType := dstSlice.Type().Elem()
	newSlice := reflect.MakeSlice(dstSlice.Type(), 0, len(srcSlice))

	for _, item := range srcSlice {
		// Convert each item to the target element type
		newElement := reflect.New(elementType).Interface()
		err := mapStructToInterface(item, newElement)
		if err != nil {
			return err
		}
		newSlice = reflect.Append(newSlice, reflect.ValueOf(newElement).Elem())
	}

	dstSlice.Set(newSlice)
	return nil
}

func mapStructToInterface(src, dst interface{}) error {
	if src == nil {
		return ErrNotFound
	}

	// Handle slice conversion specifically
	if srcSlice, ok := src.([]interface{}); ok {
		// Use reflection to handle slice conversion properly
		return convertSliceWithReflect(srcSlice, dst)
	}

	// Handle single document conversion
	data, err := bson.Marshal(src)
	if err != nil {
		return err
	}
	return bson.Unmarshal(data, dst)
}

// Copy creates a copy of the session (mgo API compatible)
func (m *ModernMGO) Copy() *ModernMGO {
	return &ModernMGO{
		client:     m.client, // Reuse the same client connection
		dbName:     m.dbName,
		mode:       m.mode,
		safe:       m.safe,
		isOriginal: false, // Mark as copy
	}
}

// Clone creates a clone of the session (mgo API compatible)
func (m *ModernMGO) Clone() *ModernMGO {
	return m.Copy() // In our implementation, Clone behaves like Copy
}

// SetMode sets the session mode for read preference (mgo API compatible)
func (m *ModernMGO) SetMode(mode Mode, refresh bool) {
	m.mode = mode
	// Note: refresh parameter is for mgo compatibility but not used in modern driver
}

// Mode returns the current session mode
func (m *ModernMGO) Mode() Mode {
	return m.mode
}

// getReadPreference converts mgo Mode to official driver ReadPreference
func (m *ModernMGO) getReadPreference() *readpref.ReadPref {
	switch m.mode {
	case Primary:
		return readpref.Primary()
	case PrimaryPreferred:
		return readpref.PrimaryPreferred()
	case Secondary:
		return readpref.Secondary()
	case SecondaryPreferred:
		return readpref.SecondaryPreferred()
	case Nearest:
		return readpref.Nearest()
	default:
		return readpref.Primary()
	}
}

// FindId finds a document by its ID (mgo API compatible)
func (c *ModernColl) FindId(id interface{}) *ModernQ {
	filter := convertMGOToOfficial(bson.M{"_id": id})
	return &ModernQ{
		coll:   c,
		filter: filter,
		skip:   0,
		limit:  0,
	}
}

// UpdateId updates a document by its ID (mgo API compatible)
func (c *ModernColl) UpdateId(id, update interface{}) error {
	return c.Update(bson.M{"_id": id}, update)
}

// RemoveId removes a document by its ID (mgo API compatible)
func (c *ModernColl) RemoveId(id interface{}) error {
	return c.Remove(bson.M{"_id": id})
}

// RemoveAll removes all documents matching the selector (mgo API compatible)
func (c *ModernColl) RemoveAll(selector interface{}) (*ChangeInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := convertMGOToOfficial(selector)
	result, err := c.mgoColl.DeleteMany(ctx, filter)
	if err != nil {
		return nil, err
	}

	return &ChangeInfo{
		Removed: int(result.DeletedCount),
		Matched: int(result.DeletedCount),
	}, nil
}

// Upsert updates a document or inserts it if it doesn't exist (mgo API compatible)
func (c *ModernColl) Upsert(selector, update interface{}) (*ChangeInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := convertMGOToOfficial(selector)
	updateDoc := convertMGOToOfficial(update)

	opts := options.Update().SetUpsert(true)
	result, err := c.mgoColl.UpdateOne(ctx, filter, updateDoc, opts)
	if err != nil {
		return nil, err
	}

	changeInfo := &ChangeInfo{
		Updated: int(result.ModifiedCount),
		Matched: int(result.MatchedCount),
	}

	if result.UpsertedID != nil {
		changeInfo.UpsertedId = convertOfficialToMGO(result.UpsertedID)
	}

	return changeInfo, nil
}
