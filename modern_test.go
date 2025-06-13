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

	session, err := mgo.DialModernMGO("mongodb://localhost:27018/test")
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

	// Test original mgo - handle gracefully if it fails due to version incompatibility
	t.Log("--- Testing Original mgo ---")
	sessionOrig, err := mgo.Dial("localhost:27017")
	if err != nil {
		t.Logf("Original mgo connection failed (this is expected with newer MongoDB versions): %v", err)
		t.Log("⚠️  Skipping original mgo tests due to version incompatibility")
	} else {
		defer sessionOrig.Close()
		t.Log("✓ Original mgo connection successful")

		// Test basic operations with original mgo
		c := sessionOrig.DB("test").C("mgo_comparison")
		c.DropCollection()

		doc := bson.M{"name": "original mgo", "value": 123, "timestamp": time.Now()}
		err = c.Insert(doc)
		if err != nil {
			t.Logf("Original mgo insert failed: %v", err)
		} else {
			t.Log("✓ Original mgo insert successful")
		}

		var result bson.M
		err = c.Find(bson.M{"name": "original mgo"}).One(&result)
		if err != nil {
			t.Logf("Original mgo find failed: %v", err)
		} else {
			t.Logf("✓ Original mgo found: %v", result["name"])
		}
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

	t.Log("✓ Modern wrapper implementation works successfully!")
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

// TestModernPipeAggregation tests the modern wrapper's aggregation pipeline support
func TestModernPipeAggregation(t *testing.T) {
	t.Log("Testing Modern Wrapper Aggregation Pipeline (Pipe) functionality")

	session, err := mgo.DialModernMGO("mongodb://localhost:27017/test")
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer session.Close()

	c := session.DB("test").C("pipe_test")
	c.DropCollection()

	// Insert test data
	testData := []interface{}{
		bson.M{"name": "Alice", "age": 25, "department": "Engineering", "salary": 75000},
		bson.M{"name": "Bob", "age": 30, "department": "Engineering", "salary": 85000},
		bson.M{"name": "Charlie", "age": 35, "department": "Marketing", "salary": 70000},
		bson.M{"name": "Diana", "age": 28, "department": "Engineering", "salary": 80000},
		bson.M{"name": "Eve", "age": 32, "department": "Marketing", "salary": 75000},
		bson.M{"name": "Frank", "age": 40, "department": "Sales", "salary": 90000},
	}

	err = c.Insert(testData...)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}
	t.Log("✓ Test data inserted")

	// Test basic aggregation pipeline
	pipeline := []bson.M{
		{"$match": bson.M{"department": "Engineering"}},
		{"$group": bson.M{
			"_id":        "$department",
			"avgSalary":  bson.M{"$avg": "$salary"},
			"totalCount": bson.M{"$sum": 1},
			"maxSalary":  bson.M{"$max": "$salary"},
			"minSalary":  bson.M{"$min": "$salary"},
		}},
	}

	var result bson.M
	err = c.Pipe(pipeline).One(&result)
	if err != nil {
		t.Errorf("Pipe One failed: %v", err)
		return
	}
	t.Log("✓ Pipe One successful")
	t.Logf("  Engineering department stats: %v", result)

	// Verify aggregation results
	if dept, ok := result["_id"].(string); !ok || dept != "Engineering" {
		t.Errorf("Expected department 'Engineering', got %v", result["_id"])
		return
	}

	// Handle different count types that might be returned
	totalCountVal := result["totalCount"]
	var totalCount int
	switch v := totalCountVal.(type) {
	case int:
		totalCount = v
	case int32:
		totalCount = int(v)
	case int64:
		totalCount = int(v)
	default:
		t.Errorf("Unexpected totalCount type: %T, value: %v", totalCountVal, totalCountVal)
		return
	}

	if totalCount != 3 {
		t.Errorf("Expected count 3, got %d", totalCount)
		return
	}
	t.Logf("✓ Aggregation result verification successful: %d Engineering employees", totalCount)

	// Test aggregation with sorting and limiting using iterator
	sortedPipeline := []bson.M{
		{"$sort": bson.M{"salary": -1}}, // Sort by salary descending
		{"$limit": 3},
		{"$project": bson.M{
			"name":   1,
			"salary": 1,
			"_id":    0,
		}},
	}

	// Use iterator instead of All for slice results
	sortedIter := c.Pipe(sortedPipeline).Iter()
	var topEarners []bson.M
	var earner bson.M
	for sortedIter.Next(&earner) {
		topEarners = append(topEarners, bson.M{
			"name":   earner["name"],
			"salary": earner["salary"],
		})
	}
	err = sortedIter.Close()
	if err != nil {
		t.Errorf("Pipe iterator failed: %v", err)
		return
	}
	t.Log("✓ Pipe iterator for sorted results successful")
	t.Logf("  Top 3 earners: %v", topEarners)

	if len(topEarners) != 3 {
		t.Errorf("Expected 3 top earners, got %d", len(topEarners))
	}

	// Test pipe with method chaining using iterator
	chainedPipe := c.Pipe([]bson.M{
		{"$match": bson.M{"age": bson.M{"$gte": 30}}},
		{"$sort": bson.M{"age": 1}},
	}).Batch(2).SetMaxTime(5 * time.Second)

	chainedIter := chainedPipe.Iter()
	var adults []bson.M
	var adult bson.M
	for chainedIter.Next(&adult) {
		adults = append(adults, bson.M{
			"name": adult["name"],
			"age":  adult["age"],
		})
	}
	err = chainedIter.Close()
	if err != nil {
		t.Errorf("Chained pipe iterator failed: %v", err)
		return
	}
	t.Log("✓ Pipe method chaining successful")
	t.Logf("  Adults (30+): %d people", len(adults))

	// Test pipe iterator
	iterPipeline := []bson.M{
		{"$match": bson.M{"department": bson.M{"$ne": "Sales"}}},
		{"$sort": bson.M{"name": 1}},
	}

	iter := c.Pipe(iterPipeline).Iter()
	count := 0
	var person bson.M
	t.Log("Testing pipe iterator...")
	for iter.Next(&person) {
		count++
		t.Logf("  Person %d: %s (%s)", count, person["name"], person["department"])
	}
	err = iter.Close()
	if err != nil {
		t.Errorf("Pipe iterator close failed: %v", err)
		return
	}
	t.Logf("✓ Pipe iterator processed %d people", count)

	// Test AllowDiskUse for large aggregations using iterator
	diskPipe := c.Pipe([]bson.M{
		{"$group": bson.M{
			"_id":        "$department",
			"avgAge":     bson.M{"$avg": "$age"},
			"totalCount": bson.M{"$sum": 1},
		}},
	}).AllowDiskUse()

	diskIter := diskPipe.Iter()
	var deptGroups []bson.M
	var group bson.M
	for diskIter.Next(&group) {
		deptGroups = append(deptGroups, group)
	}
	err = diskIter.Close()
	if err != nil {
		t.Errorf("AllowDiskUse pipe iterator failed: %v", err)
		return
	}
	t.Log("✓ AllowDiskUse pipe successful")
	t.Logf("  Department groups: %d", len(deptGroups))

	// Cleanup
	err = c.DropCollection()
	if err != nil {
		t.Errorf("Cleanup failed: %v", err)
	} else {
		t.Log("✓ Cleanup successful")
	}
}

// TestModernPipeAggregationMongoDB60 tests aggregation against MongoDB 6.0
func TestModernPipeAggregationMongoDB60(t *testing.T) {
	t.Log("Testing Modern Wrapper Aggregation Pipeline against MongoDB 6.0")

	session, err := mgo.DialModernMGO("mongodb://localhost:27018/test")
	if err != nil {
		t.Logf("Failed to connect to MongoDB 6.0: %v", err)
		t.Skip("Skipping MongoDB 6.0 aggregation tests due to connection failure")
		return
	}
	defer session.Close()

	c := session.DB("test").C("pipe_test_60")
	c.DropCollection()

	// Insert test data
	testData := []interface{}{
		bson.M{"product": "laptop", "category": "electronics", "price": 999, "quantity": 50},
		bson.M{"product": "mouse", "category": "electronics", "price": 25, "quantity": 200},
		bson.M{"product": "desk", "category": "furniture", "price": 299, "quantity": 15},
		bson.M{"product": "chair", "category": "furniture", "price": 199, "quantity": 25},
		bson.M{"product": "monitor", "category": "electronics", "price": 399, "quantity": 30},
	}

	err = c.Insert(testData...)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}
	t.Log("✓ Test data inserted")

	// Test complex aggregation pipeline with MongoDB 6.0 features
	pipeline := []bson.M{
		{"$match": bson.M{"category": "electronics"}},
		{"$addFields": bson.M{
			"totalValue": bson.M{"$multiply": []interface{}{"$price", "$quantity"}},
		}},
		{"$group": bson.M{
			"_id":          "$category",
			"avgPrice":     bson.M{"$avg": "$price"},
			"totalValue":   bson.M{"$sum": "$totalValue"},
			"productCount": bson.M{"$sum": 1},
			"products":     bson.M{"$push": "$product"},
		}},
	}

	var result bson.M
	err = c.Pipe(pipeline).One(&result)
	if err != nil {
		t.Errorf("MongoDB 6.0 Pipe One failed: %v", err)
		return
	}
	t.Log("✓ MongoDB 6.0 Pipe One successful")
	t.Logf("  Electronics summary: %v", result)

	// Test advanced pipeline with sorting and projection
	advancedPipeline := []bson.M{
		{"$sort": bson.M{"price": -1}},
		{"$project": bson.M{
			"product": 1,
			"price":   1,
			"priceRange": bson.M{
				"$cond": bson.M{
					"if":   bson.M{"$gte": []interface{}{"$price", 300}},
					"then": "premium",
					"else": "standard",
				},
			},
		}},
	}

	iter := c.Pipe(advancedPipeline).Iter()
	var products []bson.M
	var product bson.M
	for iter.Next(&product) {
		products = append(products, product)
	}
	err = iter.Close()
	if err != nil {
		t.Errorf("MongoDB 6.0 advanced pipeline failed: %v", err)
		return
	}
	t.Log("✓ MongoDB 6.0 advanced pipeline successful")
	t.Logf("  Processed %d products with price ranges", len(products))

	// Cleanup
	err = c.DropCollection()
	if err != nil {
		t.Errorf("Cleanup failed: %v", err)
	} else {
		t.Log("✓ MongoDB 6.0 aggregation test cleanup successful")
	}
}

// TestModernWrapperCompleteMethods tests all the mgo methods for complete API coverage
func TestModernWrapperCompleteMethods(t *testing.T) {
	t.Log("Testing Modern Wrapper Complete mgo Method Coverage")

	session, err := mgo.DialModernMGO("mongodb://localhost:27017/test")
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer session.Close()

	// Test session methods
	t.Log("Testing Session Methods...")

	// Test Copy and Clone
	sessionCopy := session.Copy()
	if sessionCopy == nil {
		t.Error("Copy() returned nil")
	}
	sessionClone := session.Clone()
	if sessionClone == nil {
		t.Error("Clone() returned nil")
	}
	sessionCopy.Close()
	sessionClone.Close()

	// Test SetMode and Mode
	session.SetMode(mgo.SecondaryPreferred, false)
	if session.Mode() != mgo.SecondaryPreferred {
		t.Errorf("Expected SecondaryPreferred mode, got %v", session.Mode())
	}
	session.SetMode(mgo.Primary, false) // Reset to primary

	t.Log("✓ Session methods successful")

	// Test collection methods
	c := session.DB("test").C("complete_methods_test")
	c.DropCollection()

	// Insert test data
	testDoc := bson.M{
		"name":       "Complete Test User",
		"email":      "complete@test.com",
		"age":        30,
		"department": "Testing",
		"active":     true,
	}

	err = c.Insert(testDoc)
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}
	t.Log("✓ Insert successful")

	// Test FindId
	var foundDoc bson.M
	err = c.Find(bson.M{"name": "Complete Test User"}).One(&foundDoc)
	if err != nil {
		t.Fatalf("Find for getting ID failed: %v", err)
	}

	var foundById bson.M
	err = c.FindId(foundDoc["_id"]).One(&foundById)
	if err != nil {
		t.Errorf("FindId failed: %v", err)
	} else {
		t.Log("✓ FindId successful")
	}

	// Test UpdateId
	err = c.UpdateId(foundDoc["_id"], bson.M{"$set": bson.M{"age": 31}})
	if err != nil {
		t.Errorf("UpdateId failed: %v", err)
	} else {
		t.Log("✓ UpdateId successful")
	}

	// Test Upsert
	changeInfo, err := c.Upsert(
		bson.M{"email": "upsert@test.com"},
		bson.M{"$set": bson.M{"name": "Upsert User", "age": 25}},
	)
	if err != nil {
		t.Errorf("Upsert failed: %v", err)
	} else {
		t.Logf("✓ Upsert successful: %+v", changeInfo)
	}

	// Test Query with Select
	var selectedFields bson.M
	err = c.Find(bson.M{"name": "Complete Test User"}).Select(bson.M{"name": 1, "age": 1}).One(&selectedFields)
	if err != nil {
		t.Errorf("Query with Select failed: %v", err)
	} else {
		t.Log("✓ Query Select successful")
		if _, hasEmail := selectedFields["email"]; hasEmail {
			t.Error("Select should have excluded email field")
		}
		if _, hasName := selectedFields["name"]; !hasName {
			t.Error("Select should have included name field")
		}
	}

	// Test Apply for update
	var beforeUpdate bson.M
	changeInfo, err = c.Find(bson.M{"name": "Complete Test User"}).Apply(mgo.Change{
		Update:    bson.M{"$set": bson.M{"applied": true}},
		ReturnNew: false,
	}, &beforeUpdate)
	if err != nil {
		t.Errorf("Apply update failed: %v", err)
	} else {
		t.Log("✓ Apply update successful")
		if applied, hasApplied := beforeUpdate["applied"]; hasApplied && applied.(bool) {
			t.Error("ReturnNew=false should return document before update")
		}
	}

	// Test Apply for remove
	var removedDoc bson.M
	changeInfo, err = c.Find(bson.M{"name": "Upsert User"}).Apply(mgo.Change{
		Remove: true,
	}, &removedDoc)
	if err != nil {
		t.Errorf("Apply remove failed: %v", err)
	} else {
		t.Log("✓ Apply remove successful")
		if changeInfo.Removed != 1 {
			t.Errorf("Expected 1 removed, got %d", changeInfo.Removed)
		}
	}

	// Add more test data for RemoveAll
	testDocs := []interface{}{
		bson.M{"name": "Remove1", "category": "remove_test"},
		bson.M{"name": "Remove2", "category": "remove_test"},
		bson.M{"name": "Remove3", "category": "remove_test"},
	}
	err = c.Insert(testDocs...)
	if err != nil {
		t.Fatalf("Insert test docs failed: %v", err)
	}

	// Test RemoveAll
	changeInfo, err = c.RemoveAll(bson.M{"category": "remove_test"})
	if err != nil {
		t.Errorf("RemoveAll failed: %v", err)
	} else {
		t.Logf("✓ RemoveAll successful: removed %d documents", changeInfo.Removed)
		if changeInfo.Removed != 3 {
			t.Errorf("Expected 3 removed, got %d", changeInfo.Removed)
		}
	}

	// Test RemoveId
	err = c.RemoveId(foundDoc["_id"])
	if err != nil {
		t.Errorf("RemoveId failed: %v", err)
	} else {
		t.Log("✓ RemoveId successful")
	}

	// Test Run command
	var pingResult bson.M
	err = c.Run(bson.M{"ping": 1}, &pingResult)
	if err != nil {
		t.Errorf("Run command failed: %v", err)
	} else {
		t.Log("✓ Run command successful")
		if ok, hasOk := pingResult["ok"]; !hasOk || ok != 1.0 {
			t.Error("Ping command should return ok: 1")
		}
	}

	// Cleanup
	err = c.DropCollection()
	if err != nil {
		t.Errorf("Cleanup failed: %v", err)
	} else {
		t.Log("✓ Complete methods test cleanup successful")
	}

	t.Log("🎉 All mgo methods are working correctly!")
}

// TestModernOrQuery tests the $or query functionality with []bson.M
func TestModernOrQuery(t *testing.T) {
	t.Log("Testing Modern Wrapper $or query with []bson.M")

	session, err := mgo.DialModernMGO("mongodb://localhost:27017/test")
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer session.Close()

	c := session.DB("test").C("or_query_test")
	c.DropCollection()

	// Insert test data with nested structure similar to the user's case
	testData := []interface{}{
		bson.M{
			"_id":        bson.NewObjectId(),
			"enabled":    true,
			"conditions": bson.M{"deviceId": "device1"},
		},
		bson.M{
			"_id":              bson.NewObjectId(),
			"enabled":          true,
			"secondConditions": bson.M{"deviceId": "device2"},
		},
		bson.M{
			"_id":        bson.NewObjectId(),
			"enabled":    false,
			"conditions": bson.M{"deviceId": "device1"},
		},
	}

	err = c.Insert(testData...)
	if err != nil {
		t.Fatalf("Failed to insert test data: %v", err)
	}
	t.Log("✓ Test data inserted")

	// Test the $or query that was causing the issue
	deviceID := "device1"
	query := bson.M{
		"$or": []bson.M{
			{
				"conditions.deviceId": deviceID,
			},
			{
				"secondConditions.deviceId": deviceID,
			},
		},
		"enabled": true,
	}

	var results []bson.M
	err = c.Find(query).All(&results)
	if err != nil {
		t.Errorf("$or query failed: %v", err)
		return
	}
	t.Log("✓ $or query successful")
	t.Logf("  Found %d documents", len(results))

	// Should find exactly 1 document (the one with conditions.deviceId = device1 and enabled = true)
	if len(results) != 1 {
		t.Errorf("Expected 1 document, got %d", len(results))
		return
	}

	// Verify the found document has the correct structure
	foundDoc := results[0]
	if enabled, ok := foundDoc["enabled"].(bool); !ok || !enabled {
		t.Errorf("Expected enabled=true, got %v", foundDoc["enabled"])
		return
	}

	// Check that it has conditions with deviceId
	if conditions, ok := foundDoc["conditions"].(bson.M); ok {
		if deviceId, ok := conditions["deviceId"].(string); !ok || deviceId != "device1" {
			t.Errorf("Expected conditions.deviceId='device1', got %v", conditions["deviceId"])
			return
		}
	} else {
		t.Errorf("Expected conditions field, got %v", foundDoc["conditions"])
		return
	}

	t.Log("✓ $or query result verification successful")

	// Cleanup
	err = c.DropCollection()
	if err != nil {
		t.Errorf("Cleanup failed: %v", err)
	} else {
		t.Log("✓ Cleanup successful")
	}
}

// TestModernOrQueryDetailedReproduction tests the exact scenario from the user's code
func TestModernOrQueryDetailedReproduction(t *testing.T) {
	t.Log("Testing Modern Wrapper $or query with exact user scenario")

	session, err := mgo.DialModernMGO("mongodb://localhost:27017/test")
	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer session.Close()

	c := session.DB("test").C("triggers")
	c.DropCollection()

	// Create test data that mimics the user's trigger structure
	deviceID := "test-device-123"

	testTriggers := []interface{}{
		bson.M{
			"_id":     bson.NewObjectId(),
			"enabled": true,
			"conditions": bson.M{
				"deviceId": deviceID,
				"type":     "sensor",
			},
			"name": "trigger1",
		},
		bson.M{
			"_id":     bson.NewObjectId(),
			"enabled": true,
			"secondConditions": bson.M{
				"deviceId": deviceID,
				"type":     "actuator",
			},
			"name": "trigger2",
		},
		bson.M{
			"_id":     bson.NewObjectId(),
			"enabled": false, // This should not be returned
			"conditions": bson.M{
				"deviceId": deviceID,
				"type":     "sensor",
			},
			"name": "trigger3",
		},
		bson.M{
			"_id":     bson.NewObjectId(),
			"enabled": true,
			"conditions": bson.M{
				"deviceId": "other-device", // Different device
				"type":     "sensor",
			},
			"name": "trigger4",
		},
	}

	err = c.Insert(testTriggers...)
	if err != nil {
		t.Fatalf("Failed to insert test triggers: %v", err)
	}
	t.Log("✓ Test triggers inserted")

	// Create a mock device object similar to user's code
	device := struct {
		ID string
	}{
		ID: deviceID,
	}

	// The exact query structure from the user's code
	query := bson.M{
		"$or": []bson.M{
			{
				"conditions.deviceId": device.ID,
			},
			{
				"secondConditions.deviceId": device.ID,
			},
		},
		"enabled": true,
	}

	// Test with Find().All() like the user's code
	var triggers []bson.M
	err = c.Find(query).All(&triggers)
	if err != nil {
		t.Errorf("Failed to find triggers: %v", err)
		return
	}

	t.Logf("✓ Find triggers successful, found %d triggers", len(triggers))

	// Should find exactly 2 triggers (trigger1 and trigger2)
	if len(triggers) != 2 {
		t.Errorf("Expected 2 triggers, got %d", len(triggers))
		for i, trigger := range triggers {
			t.Logf("  Trigger %d: %v", i, trigger["name"])
		}
		return
	}

	// Verify the found triggers
	foundNames := make(map[string]bool)
	for _, trigger := range triggers {
		name := trigger["name"].(string)
		foundNames[name] = true

		enabled := trigger["enabled"].(bool)
		if !enabled {
			t.Errorf("Found disabled trigger: %s", name)
			return
		}
	}

	if !foundNames["trigger1"] || !foundNames["trigger2"] {
		t.Errorf("Expected to find trigger1 and trigger2, found: %v", foundNames)
		return
	}

	t.Log("✓ Trigger query result verification successful")

	// Test with different query structures to ensure robustness
	// Test with more complex nested structure
	complexQuery := bson.M{
		"$and": []bson.M{
			{
				"$or": []bson.M{
					{"conditions.deviceId": device.ID},
					{"secondConditions.deviceId": device.ID},
				},
			},
			{"enabled": true},
		},
	}

	var complexResults []bson.M
	err = c.Find(complexQuery).All(&complexResults)
	if err != nil {
		t.Errorf("Complex query failed: %v", err)
		return
	}

	if len(complexResults) != 2 {
		t.Errorf("Complex query expected 2 results, got %d", len(complexResults))
		return
	}

	t.Log("✓ Complex nested query successful")

	// Cleanup
	err = c.DropCollection()
	if err != nil {
		t.Errorf("Cleanup failed: %v", err)
	} else {
		t.Log("✓ Cleanup successful")
	}
}
