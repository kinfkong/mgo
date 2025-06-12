package mgo_test

import (
	"time"

	mgo "github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	. "gopkg.in/check.v1"
)

func (s *S) TestFindApplyUpdate(c *C) {
	session, err := mgo.Dial("localhost:40001")
	c.Assert(err, IsNil)
	defer session.Close()

	coll := session.DB("mydb").C("mycoll")

	// Insert initial document
	doc := M{"_id": bson.NewObjectId(), "counter": 1, "name": "test", "taskPickedAt": nil}
	err = coll.Insert(doc)
	c.Assert(err, IsNil)

	// Test basic update with Apply
	change := mgo.Change{
		Update:    bson.M{"$inc": bson.M{"counter": 1}, "$set": bson.M{"taskPickedAt": time.Now()}},
		ReturnNew: true,
	}

	var result M
	info, err := coll.Find(bson.M{"_id": doc["_id"]}).Apply(change, &result)
	c.Assert(err, IsNil)
	c.Assert(info, NotNil)
	c.Assert(info.Updated, Equals, 1)
	c.Assert(result["counter"], Equals, 2)
	c.Assert(result["taskPickedAt"], NotNil)
}

func (s *S) TestFindApplyNotFound(c *C) {
	session, err := mgo.Dial("localhost:40001")
	c.Assert(err, IsNil)
	defer session.Close()

	coll := session.DB("mydb").C("mycoll")

	// Try to apply on non-existent document
	change := mgo.Change{
		Update:    bson.M{"$set": bson.M{"status": "updated"}},
		ReturnNew: true,
	}

	var result M
	_, err = coll.Find(bson.M{"_id": bson.NewObjectId()}).Apply(change, &result)
	c.Assert(err, Equals, mgo.ErrNotFound)
}

func (s *S) TestFindApplyUpsert(c *C) {
	session, err := mgo.Dial("localhost:40001")
	c.Assert(err, IsNil)
	defer session.Close()

	coll := session.DB("mydb").C("mycoll")

	// Test upsert functionality
	change := mgo.Change{
		Update:    bson.M{"$set": bson.M{"status": "created", "createdAt": time.Now()}},
		Upsert:    true,
		ReturnNew: true,
	}

	newId := bson.NewObjectId()
	var result M
	info, err := coll.Find(bson.M{"_id": newId}).Apply(change, &result)
	c.Assert(err, IsNil)
	c.Assert(info, NotNil)
	c.Assert(info.UpsertedId, Equals, newId)
	c.Assert(result["_id"], Equals, newId)
	c.Assert(result["status"], Equals, "created")
}

func (s *S) TestFindApplyRemove(c *C) {
	session, err := mgo.Dial("localhost:40001")
	c.Assert(err, IsNil)
	defer session.Close()

	coll := session.DB("mydb").C("mycoll")

	// Insert document to remove
	doc := M{"_id": bson.NewObjectId(), "name": "to_remove", "value": 42}
	err = coll.Insert(doc)
	c.Assert(err, IsNil)

	// Test remove with Apply
	change := mgo.Change{
		Remove: true,
	}

	var result M
	info, err := coll.Find(bson.M{"_id": doc["_id"]}).Apply(change, &result)
	c.Assert(err, IsNil)
	c.Assert(info, NotNil)
	c.Assert(info.Removed, Equals, 1)
	c.Assert(result["name"], Equals, "to_remove")
	c.Assert(result["value"], Equals, 42)

	// Verify document was actually removed
	count, err := coll.Find(bson.M{"_id": doc["_id"]}).Count()
	c.Assert(err, IsNil)
	c.Assert(count, Equals, 0)
}

func (s *S) TestFindApplyReturnOld(c *C) {
	session, err := mgo.Dial("localhost:40001")
	c.Assert(err, IsNil)
	defer session.Close()

	coll := session.DB("mydb").C("mycoll")

	// Insert initial document
	doc := M{"_id": bson.NewObjectId(), "counter": 10, "name": "original"}
	err = coll.Insert(doc)
	c.Assert(err, IsNil)

	// Test returning old document (default behavior)
	change := mgo.Change{
		Update:    bson.M{"$inc": bson.M{"counter": 5}, "$set": bson.M{"name": "updated"}},
		ReturnNew: false, // This is the default
	}

	var result M
	info, err := coll.Find(bson.M{"_id": doc["_id"]}).Apply(change, &result)
	c.Assert(err, IsNil)
	c.Assert(info, NotNil)
	c.Assert(info.Updated, Equals, 1)
	// Should return the original values
	c.Assert(result["counter"], Equals, 10)
	c.Assert(result["name"], Equals, "original")

	// Verify the document was actually updated
	var updated M
	err = coll.Find(bson.M{"_id": doc["_id"]}).One(&updated)
	c.Assert(err, IsNil)
	c.Assert(updated["counter"], Equals, 15)
	c.Assert(updated["name"], Equals, "updated")
}

func (s *S) TestFindApplyWithCondition(c *C) {
	session, err := mgo.Dial("localhost:40001")
	c.Assert(err, IsNil)
	defer session.Close()

	coll := session.DB("mydb").C("mycoll")

	// Insert multiple documents
	docs := []interface{}{
		M{"_id": bson.NewObjectId(), "status": "pending", "priority": 1, "taskPickedAt": nil},
		M{"_id": bson.NewObjectId(), "status": "pending", "priority": 2, "taskPickedAt": nil},
		M{"_id": bson.NewObjectId(), "status": "completed", "priority": 1, "taskPickedAt": time.Now()},
	}
	err = coll.Insert(docs...)
	c.Assert(err, IsNil)

	// Test applying with complex condition - similar to countdown service pattern
	condition := bson.M{
		"status":       "pending",
		"taskPickedAt": nil,
	}
	change := mgo.Change{
		Update:    bson.M{"$set": bson.M{"taskPickedAt": time.Now(), "status": "processing"}},
		ReturnNew: false,
	}

	var result M
	info, err := coll.Find(condition).Apply(change, &result)
	c.Assert(err, IsNil)
	c.Assert(info, NotNil)
	c.Assert(info.Updated, Equals, 1)
	// Should get one of the pending documents
	c.Assert(result["status"], Equals, "pending")
	c.Assert(result["taskPickedAt"], IsNil)

	// Verify only one document was updated
	count, err := coll.Find(bson.M{"status": "processing"}).Count()
	c.Assert(err, IsNil)
	c.Assert(count, Equals, 1)
}

func (s *S) TestFindApplySort(c *C) {
	session, err := mgo.Dial("localhost:40001")
	c.Assert(err, IsNil)
	defer session.Close()

	coll := session.DB("mydb").C("mycoll")

	// Insert documents with different priorities
	docs := []interface{}{
		M{"_id": bson.NewObjectId(), "priority": 3, "name": "low"},
		M{"_id": bson.NewObjectId(), "priority": 1, "name": "high"},
		M{"_id": bson.NewObjectId(), "priority": 2, "name": "medium"},
	}
	err = coll.Insert(docs...)
	c.Assert(err, IsNil)

	// Test that Apply respects sort order
	change := mgo.Change{
		Update:    bson.M{"$set": bson.M{"processed": true}},
		ReturnNew: false,
	}

	var result M
	// Should pick the document with highest priority (lowest number)
	info, err := coll.Find(nil).Sort("priority").Apply(change, &result)
	c.Assert(err, IsNil)
	c.Assert(info, NotNil)
	c.Assert(info.Updated, Equals, 1)
	c.Assert(result["name"], Equals, "high")
	c.Assert(result["priority"], Equals, 1)
}

func (s *S) TestFindApplyAtomicity(c *C) {
	session, err := mgo.Dial("localhost:40001")
	c.Assert(err, IsNil)
	defer session.Close()

	coll := session.DB("mydb").C("mycoll")

	// Insert a document
	doc := M{"_id": bson.NewObjectId(), "counter": 0, "available": true}
	err = coll.Insert(doc)
	c.Assert(err, IsNil)

	// Test atomicity - only one of concurrent operations should succeed
	change := mgo.Change{
		Update:    bson.M{"$inc": bson.M{"counter": 1}, "$set": bson.M{"available": false}},
		ReturnNew: true,
	}

	condition := bson.M{"_id": doc["_id"], "available": true}

	// First apply should succeed
	var result1 M
	info1, err1 := coll.Find(condition).Apply(change, &result1)
	c.Assert(err1, IsNil)
	c.Assert(info1.Updated, Equals, 1)
	c.Assert(result1["counter"], Equals, 1)
	c.Assert(result1["available"], Equals, false)

	// Second apply with same condition should fail (not found)
	var result2 M
	_, err2 := coll.Find(condition).Apply(change, &result2)
	c.Assert(err2, Equals, mgo.ErrNotFound)
}
