// modern_collection.go - Collection operations for modern MongoDB driver compatibility wrapper

package mgo

import (
	"context"
	"strings"
	"time"

	"github.com/globalsign/mgo/bson"
	officialBson "go.mongodb.org/mongo-driver/bson"
	mongodrv "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

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
