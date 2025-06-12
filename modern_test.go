package mgo_test

import (
	"testing"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// TestModernWrapperMongoDB36 tests the modern wrapper against MongoDB 3.6 (localhost:27017)
func TestModernWrapperMongoDB36(t *testing.T) {
	t.Log("Testing Modern Wrapper against MongoDB 3.6 on localhost:27017")

	session, err := mgo.DialModernMGO("mongodb://localhost:27017/test")
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB 3.6: %v", err)
	}
	defer session.Close()

	// Test server info
	buildInfo, err := session.BuildInfo()
	if err != nil {
		t.Errorf("Failed to get build info: %v", err)
	} else {
		t.Logf("MongoDB version: %s", buildInfo.Version)
		t.Logf("Git version: %s", buildInfo.GitVersion)
	}

	// Test ping
	err = session.Ping()
	if err != nil {
		t.Errorf("Ping failed: %v", err)
	} else {
		t.Log("✓ Ping successful")
	}

	// Test basic CRUD operations
	testModernOperations(t, session, "test", "modern_test_36")
}

// TestModernWrapperMongoDB60 tests the modern wrapper against MongoDB 6.0 (localhost:27018)
func TestModernWrapperMongoDB60(t *testing.T) {
	t.Log("Testing Modern Wrapper against MongoDB 6.0 on localhost:27018")

	session, err := mgo.DialModernMGOWithTimeout("mongodb://localhost:27018/test", 10*time.Second)
	if err != nil {
		t.Logf("Failed to connect to MongoDB 6.0: %v", err)
		t.Log("This test will check if the modern wrapper can handle MongoDB 6.0")
		t.Skip("Skipping MongoDB 6.0 tests due to connection failure")
		return
	}
	defer session.Close()

	// Test server info
	buildInfo, err := session.BuildInfo()
	if err != nil {
		t.Errorf("Failed to get build info: %v", err)
	} else {
		t.Logf("MongoDB version: %s", buildInfo.Version)
		t.Logf("Git version: %s", buildInfo.GitVersion)

		// Check if this is really MongoDB 6.0+
		if len(buildInfo.VersionArray) > 0 && buildInfo.VersionArray[0] >= 6 {
			t.Log("✓ Successfully connected to MongoDB 6.0+ using modern wrapper!")
		}
	}

	// Test ping
	err = session.Ping()
	if err != nil {
		t.Errorf("Ping failed: %v", err)
	} else {
		t.Log("✓ Ping successful")
	}

	// Test basic CRUD operations
	testModernOperations(t, session, "test", "modern_test_60")
}

// TestCompareOriginalVsModern compares original mgo with modern wrapper on MongoDB 3.6
func TestCompareOriginalVsModern(t *testing.T) {
	t.Log("Comparing Original mgo vs Modern Wrapper on MongoDB 3.6")

	// Test original mgo
	t.Log("--- Testing Original mgo ---")
	sessionOrig, err := mgo.Dial("localhost:27017")
	if err != nil {
		t.Fatalf("Failed to connect with original mgo: %v", err)
	}
	defer sessionOrig.Close()

	// Test basic operations with original mgo
	c := sessionOrig.DB("test").C("mgo_comparison")
	c.DropCollection()

	doc := bson.M{"name": "original mgo", "value": 123, "timestamp": time.Now()}
	err = c.Insert(doc)
	if err != nil {
		t.Errorf("Original mgo insert failed: %v", err)
	} else {
		t.Log("✓ Original mgo insert successful")
	}

	var result bson.M
	err = c.Find(bson.M{"name": "original mgo"}).One(&result)
	if err != nil {
		t.Errorf("Original mgo find failed: %v", err)
	} else {
		t.Logf("✓ Original mgo found: %v", result["name"])
	}

	// Test modern wrapper
	t.Log("--- Testing Modern Wrapper ---")
	sessionModern, err := mgo.DialModernMGO("mongodb://localhost:27017/test")
	if err != nil {
		t.Fatalf("Failed to connect with modern wrapper: %v", err)
	}
	defer sessionModern.Close()

	// Test basic operations with modern wrapper
	cm := sessionModern.DB("test").C("modern_comparison")
	cm.DropCollection()

	docModern := bson.M{"name": "modern wrapper", "value": 456, "timestamp": time.Now()}
	err = cm.Insert(docModern)
	if err != nil {
		t.Errorf("Modern wrapper insert failed: %v", err)
	} else {
		t.Log("✓ Modern wrapper insert successful")
	}

	var resultModern bson.M
	err = cm.Find(bson.M{"name": "modern wrapper"}).One(&resultModern)
	if err != nil {
		t.Errorf("Modern wrapper find failed: %v", err)
	} else {
		t.Logf("✓ Modern wrapper found: %v", resultModern["name"])
	}

	t.Log("✓ Both implementations work successfully!")
}

// testModernOperations performs comprehensive CRUD tests
func testModernOperations(t *testing.T, session *mgo.ModernMGO, dbName, collName string) {
	c := session.DB(dbName).C(collName)

	// Clean up before starting
	c.DropCollection()

	// Test Insert
	doc := bson.M{
		"name":     "modern test user",
		"email":    "modern@example.com",
		"age":      35,
		"active":   true,
		"tags":     []string{"modern", "mongodb", "testing"},
		"metadata": bson.M{"source": "modern wrapper", "timestamp": time.Now()},
	}

	err := c.Insert(doc)
	if err != nil {
		t.Errorf("Insert failed: %v", err)
		return
	}
	t.Log("✓ Insert successful")

	// Test Find One
	var result bson.M
	err = c.Find(bson.M{"name": "modern test user"}).One(&result)
	if err != nil {
		t.Errorf("Find failed: %v", err)
		return
	}
	t.Log("✓ Find successful")
	t.Logf("  Found document with _id: %v", result["_id"])

	// Test Count
	count, err := c.Find(bson.M{"active": true}).Count()
	if err != nil {
		t.Errorf("Count failed: %v", err)
		return
	}
	t.Logf("✓ Count successful: %d documents", count)
	if count != 1 {
		t.Errorf("Expected 1 document, got %d", count)
	}

	// Test Update
	err = c.Update(bson.M{"name": "modern test user"}, bson.M{"$set": bson.M{"age": 36, "updated": true}})
	if err != nil {
		t.Errorf("Update failed: %v", err)
		return
	}
	t.Log("✓ Update successful")

	// Verify update
	var updated bson.M
	err = c.Find(bson.M{"name": "modern test user"}).One(&updated)
	if err != nil {
		t.Errorf("Find after update failed: %v", err)
		return
	}
	// Handle different integer types that might be returned
	ageVal := updated["age"]
	var ageInt int
	switch v := ageVal.(type) {
	case int:
		ageInt = v
	case int32:
		ageInt = int(v)
	case int64:
		ageInt = int(v)
	default:
		t.Errorf("Update verification failed: age = %v (type: %T)", ageVal, ageVal)
		return
	}

	if ageInt != 36 {
		t.Errorf("Update verification failed: expected age 36, got %d", ageInt)
	} else {
		t.Log("✓ Update verification successful")
	}

	// Test Index creation
	index := mgo.Index{
		Key:        []string{"email"},
		Unique:     true,
		Background: true,
		Name:       "email_unique_modern",
	}
	err = c.EnsureIndex(index)
	if err != nil {
		t.Errorf("Index creation failed: %v", err)
		return
	}
	t.Log("✓ Index creation successful")

	// Test Multiple documents
	docs := []interface{}{
		bson.M{"name": "modern_user1", "email": "user1@modern.com", "age": 25},
		bson.M{"name": "modern_user2", "email": "user2@modern.com", "age": 35},
		bson.M{"name": "modern_user3", "email": "user3@modern.com", "age": 28},
	}
	err = c.Insert(docs...)
	if err != nil {
		t.Errorf("Multiple insert failed: %v", err)
		return
	}
	t.Log("✓ Multiple insert successful")

	// Test query with sorting and limiting using iterator
	query := c.Find(bson.M{}).Sort("age").Limit(2)
	sortedIter := query.Iter()

	var userCount int
	var user bson.M
	t.Log("Testing sorted query with iterator...")
	for sortedIter.Next(&user) {
		userCount++
		t.Logf("  User %d: %v (age: %v)", userCount, user["name"], user["age"])
	}
	err = sortedIter.Close()
	if err != nil {
		t.Errorf("Sorted query iterator failed: %v", err)
		return
	}
	if userCount != 2 {
		t.Errorf("Expected 2 users, got %d", userCount)
	} else {
		t.Log("✓ Sorted query successful")
	}

	// Test iterator
	t.Log("Testing iterator...")
	iter := c.Find(bson.M{}).Sort("name").Iter()
	count = 0
	var iterResult bson.M
	for iter.Next(&iterResult) {
		count++
		t.Logf("  Iterator result %d: %v", count, iterResult["name"])
	}
	err = iter.Close()
	if err != nil {
		t.Errorf("Iterator close failed: %v", err)
		return
	}
	t.Logf("✓ Iterator processed %d documents", count)

	// Test Remove
	err = c.Remove(bson.M{"name": "modern_user1"})
	if err != nil {
		t.Errorf("Remove failed: %v", err)
		return
	}
	t.Log("✓ Remove successful")

	// Clean up
	err = c.DropCollection()
	if err != nil {
		t.Errorf("Cleanup failed: %v", err)
		return
	}
	t.Log("✓ Cleanup successful")
}
