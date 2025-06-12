package mgo_test

import (
	"testing"
	"time"

	"github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// TestDocument represents a test document structure similar to action logs
type TestDocument struct {
	ID          bson.ObjectId `bson:"_id"`
	UserID      bson.ObjectId `bson:"userId"`
	Type        string        `bson:"type"`
	FamilyID    bson.ObjectId `bson:"familyId,omitempty"`
	TriggeredAt time.Time     `bson:"triggeredAt"`
	Message     string        `bson:"message"`
	Index       int           `bson:"index"` // For easy verification
}

// generateTestData creates test documents for skip/limit testing
func generateTestData(userID bson.ObjectId, count int) []interface{} {
	docs := make([]interface{}, count)
	baseTime := time.Now().Add(-time.Hour * 24) // Start from 24 hours ago

	for i := 0; i < count; i++ {
		docs[i] = TestDocument{
			ID:          bson.NewObjectId(),
			UserID:      userID,
			Type:        "test_action",
			FamilyID:    bson.NewObjectId(),
			TriggeredAt: baseTime.Add(time.Minute * time.Duration(i)), // Sequential timestamps
			Message:     "Test message " + string(rune('A'+i%26)),
			Index:       i,
		}
	}
	return docs
}

// testSkipLimitFunctionality tests skip/limit with given session
func testSkipLimitFunctionality(t *testing.T, session interface{}, testName string) {
	// Handle different session types separately for clarity
	switch s := session.(type) {
	case *mgo.Session:
		testSkipLimitWithOriginalMgo(t, s, testName)
	case *mgo.ModernMGO:
		testSkipLimitWithModernMgo(t, s, testName)
	default:
		t.Fatalf("Unknown session type for %s", testName)
	}
}

// testSkipLimitWithOriginalMgo tests skip/limit with original mgo.Session
func testSkipLimitWithOriginalMgo(t *testing.T, session *mgo.Session, testName string) {
	collection := session.DB("test_skip_limit").C("test_documents")
	runSkipLimitTests(t, &OriginalMgoWrapper{collection}, testName)
}

// testSkipLimitWithModernMgo tests skip/limit with modern mgo wrapper
func testSkipLimitWithModernMgo(t *testing.T, session *mgo.ModernMGO, testName string) {
	collection := session.DB("test_skip_limit").C("test_documents")
	runSkipLimitTests(t, &ModernMgoWrapper{collection}, testName)
}

// CollectionWrapper provides a common interface for both collection types
type CollectionWrapper interface {
	DropCollection() error
	Insert(docs ...interface{}) error
	Find(query interface{}) QueryWrapper
	Count() (int, error)
}

// QueryWrapper provides a common interface for both query types
type QueryWrapper interface {
	Sort(fields ...string) QueryWrapper
	Skip(n int) QueryWrapper
	Limit(n int) QueryWrapper
	All(result interface{}) error
}

// OriginalMgoWrapper wraps original mgo.Collection
type OriginalMgoWrapper struct {
	coll *mgo.Collection
}

func (w *OriginalMgoWrapper) DropCollection() error {
	return w.coll.DropCollection()
}

func (w *OriginalMgoWrapper) Insert(docs ...interface{}) error {
	return w.coll.Insert(docs...)
}

func (w *OriginalMgoWrapper) Find(query interface{}) QueryWrapper {
	return &OriginalQueryWrapper{w.coll.Find(query)}
}

func (w *OriginalMgoWrapper) Count() (int, error) {
	return w.coll.Count()
}

// OriginalQueryWrapper wraps original mgo.Query
type OriginalQueryWrapper struct {
	query *mgo.Query
}

func (w *OriginalQueryWrapper) Sort(fields ...string) QueryWrapper {
	return &OriginalQueryWrapper{w.query.Sort(fields...)}
}

func (w *OriginalQueryWrapper) Skip(n int) QueryWrapper {
	return &OriginalQueryWrapper{w.query.Skip(n)}
}

func (w *OriginalQueryWrapper) Limit(n int) QueryWrapper {
	return &OriginalQueryWrapper{w.query.Limit(n)}
}

func (w *OriginalQueryWrapper) All(result interface{}) error {
	return w.query.All(result)
}

// ModernMgoWrapper wraps modern mgo.ModernColl
type ModernMgoWrapper struct {
	coll *mgo.ModernColl
}

func (w *ModernMgoWrapper) DropCollection() error {
	return w.coll.DropCollection()
}

func (w *ModernMgoWrapper) Insert(docs ...interface{}) error {
	return w.coll.Insert(docs...)
}

func (w *ModernMgoWrapper) Find(query interface{}) QueryWrapper {
	return &ModernQueryWrapper{w.coll.Find(query)}
}

func (w *ModernMgoWrapper) Count() (int, error) {
	return w.coll.Count()
}

// ModernQueryWrapper wraps modern mgo.ModernQ
type ModernQueryWrapper struct {
	query *mgo.ModernQ
}

func (w *ModernQueryWrapper) Sort(fields ...string) QueryWrapper {
	return &ModernQueryWrapper{w.query.Sort(fields...)}
}

func (w *ModernQueryWrapper) Skip(n int) QueryWrapper {
	return &ModernQueryWrapper{w.query.Skip(n)}
}

func (w *ModernQueryWrapper) Limit(n int) QueryWrapper {
	return &ModernQueryWrapper{w.query.Limit(n)}
}

func (w *ModernQueryWrapper) All(result interface{}) error {
	return w.query.All(result)
}

// runSkipLimitTests runs the actual skip/limit tests using the wrapper interface
func runSkipLimitTests(t *testing.T, collection CollectionWrapper, testName string) {
	// Clean up
	err := collection.DropCollection()
	if err != nil {
		t.Logf("Warning: Could not drop collection for %s: %v", testName, err)
	}

	// Setup test data
	userID := bson.NewObjectId()
	testData := generateTestData(userID, 25) // 25 documents for thorough pagination testing

	t.Logf("[%s] Inserting %d test documents...", testName, len(testData))
	err = collection.Insert(testData...)
	if err != nil {
		t.Fatalf("[%s] Failed to insert test data: %v", testName, err)
	}

	// Verify total count
	totalCount, err := collection.Count()
	if err != nil {
		t.Fatalf("[%s] Failed to count documents: %v", testName, err)
	}
	if totalCount != 25 {
		t.Fatalf("[%s] Expected 25 documents, got %d", testName, totalCount)
	}
	t.Logf("[%s] ✓ Total documents verified: %d", testName, totalCount)

	// Test 1: Basic pagination (page 0, size 10)
	t.Logf("[%s] Testing basic pagination (page 0, size 10)...", testName)
	var page0Results []TestDocument
	pageSize := 10
	page := 0

	query := collection.Find(bson.M{"userId": userID, "type": "test_action"})
	err = query.Sort("-triggeredAt").Skip(pageSize * page).Limit(pageSize).All(&page0Results)
	if err != nil {
		t.Fatalf("[%s] Failed to execute page 0 query: %v", testName, err)
	}

	if len(page0Results) != pageSize {
		t.Errorf("[%s] Page 0: Expected %d results, got %d", testName, pageSize, len(page0Results))
	}

	// Verify sorting (should be newest first due to -triggeredAt)
	for i := 1; i < len(page0Results); i++ {
		if page0Results[i-1].TriggeredAt.Before(page0Results[i].TriggeredAt) {
			t.Errorf("[%s] Page 0: Sort order incorrect at index %d", testName, i)
		}
	}
	t.Logf("[%s] ✓ Page 0 results verified: %d items, properly sorted", testName, len(page0Results))

	// Test 2: Second page (page 1, size 10)
	t.Logf("[%s] Testing second page (page 1, size 10)...", testName)
	var page1Results []TestDocument
	page = 1

	query = collection.Find(bson.M{"userId": userID, "type": "test_action"})
	err = query.Sort("-triggeredAt").Skip(pageSize * page).Limit(pageSize).All(&page1Results)
	if err != nil {
		t.Fatalf("[%s] Failed to execute page 1 query: %v", testName, err)
	}

	if len(page1Results) != pageSize {
		t.Errorf("[%s] Page 1: Expected %d results, got %d", testName, pageSize, len(page1Results))
	}

	// Verify no overlap between pages
	for _, p0Item := range page0Results {
		for _, p1Item := range page1Results {
			if p0Item.ID == p1Item.ID {
				t.Errorf("[%s] Found duplicate item between page 0 and page 1: %s", testName, p0Item.ID.Hex())
			}
		}
	}
	t.Logf("[%s] ✓ Page 1 results verified: %d items, no overlap with page 0", testName, len(page1Results))

	// Test 3: Last partial page (page 2, size 10 - should have 5 items)
	t.Logf("[%s] Testing last partial page (page 2, size 10)...", testName)
	var page2Results []TestDocument
	page = 2

	query = collection.Find(bson.M{"userId": userID, "type": "test_action"})
	err = query.Sort("-triggeredAt").Skip(pageSize * page).Limit(pageSize).All(&page2Results)
	if err != nil {
		t.Fatalf("[%s] Failed to execute page 2 query: %v", testName, err)
	}

	expectedPage2Count := 5 // 25 total - 20 from first two pages = 5
	if len(page2Results) != expectedPage2Count {
		t.Errorf("[%s] Page 2: Expected %d results, got %d", testName, expectedPage2Count, len(page2Results))
	}
	t.Logf("[%s] ✓ Page 2 results verified: %d items (partial page)", testName, len(page2Results))

	// Test 4: Beyond available data (page 3, size 10 - should have 0 items)
	t.Logf("[%s] Testing beyond available data (page 3, size 10)...", testName)
	var page3Results []TestDocument
	page = 3

	query = collection.Find(bson.M{"userId": userID, "type": "test_action"})
	err = query.Sort("-triggeredAt").Skip(pageSize * page).Limit(pageSize).All(&page3Results)
	if err != nil {
		t.Fatalf("[%s] Failed to execute page 3 query: %v", testName, err)
	}

	if len(page3Results) != 0 {
		t.Errorf("[%s] Page 3: Expected 0 results, got %d", testName, len(page3Results))
	}
	t.Logf("[%s] ✓ Page 3 results verified: %d items (beyond data)", testName, len(page3Results))

	// Test 5: Different page size (page 0, size 7)
	t.Logf("[%s] Testing different page size (page 0, size 7)...", testName)
	var page0Size7Results []TestDocument
	pageSize = 7
	page = 0

	query = collection.Find(bson.M{"userId": userID, "type": "test_action"})
	err = query.Sort("-triggeredAt").Skip(pageSize * page).Limit(pageSize).All(&page0Size7Results)
	if err != nil {
		t.Fatalf("[%s] Failed to execute page 0 size 7 query: %v", testName, err)
	}

	if len(page0Size7Results) != pageSize {
		t.Errorf("[%s] Page 0 (size 7): Expected %d results, got %d", testName, pageSize, len(page0Size7Results))
	}
	t.Logf("[%s] ✓ Different page size results verified: %d items", testName, len(page0Size7Results))

	// Test 6: Large skip value (simulating deep pagination)
	t.Logf("[%s] Testing large skip value (skip 20, limit 5)...", testName)
	var largeSkipResults []TestDocument

	query = collection.Find(bson.M{"userId": userID, "type": "test_action"})
	err = query.Sort("-triggeredAt").Skip(20).Limit(5).All(&largeSkipResults)
	if err != nil {
		t.Fatalf("[%s] Failed to execute large skip query: %v", testName, err)
	}

	expectedLargeSkipCount := 5 // 25 total - 20 skipped = 5
	if len(largeSkipResults) != expectedLargeSkipCount {
		t.Errorf("[%s] Large skip: Expected %d results, got %d", testName, expectedLargeSkipCount, len(largeSkipResults))
	}
	t.Logf("[%s] ✓ Large skip results verified: %d items", testName, len(largeSkipResults))

	// Test 7: Skip with familyId filter (mimicking the actual action log scenario)
	t.Logf("[%s] Testing skip/limit with familyId filter...", testName)
	familyID := page0Results[0].FamilyID // Use familyID from first result
	var filteredResults []TestDocument

	query = collection.Find(bson.M{
		"userId":   userID,
		"type":     "test_action",
		"familyId": familyID,
	})
	err = query.Sort("-triggeredAt").Skip(0).Limit(5).All(&filteredResults)
	if err != nil {
		t.Fatalf("[%s] Failed to execute filtered query: %v", testName, err)
	}

	// Verify all results have the correct familyId
	for _, result := range filteredResults {
		if result.FamilyID != familyID {
			t.Errorf("[%s] Filtered query: Wrong familyId %s, expected %s", testName, result.FamilyID.Hex(), familyID.Hex())
		}
	}
	t.Logf("[%s] ✓ Filtered results verified: %d items with correct familyId", testName, len(filteredResults))

	// Cleanup
	err = collection.DropCollection()
	if err != nil {
		t.Logf("[%s] Warning: Could not clean up collection: %v", testName, err)
	} else {
		t.Logf("[%s] ✓ Cleanup successful", testName)
	}
}

// TestSkipLimitMongoDB36WithOriginalMgo tests skip/limit functionality on MongoDB 3.6 using original mgo.Dial
func TestSkipLimitMongoDB36WithOriginalMgo(t *testing.T) {
	t.Log("=== Testing Skip/Limit functionality on MongoDB 3.6 with original mgo.Dial ===")

	session, err := mgo.Dial("localhost:27017")
	if err != nil {
		t.Skipf("Failed to connect to MongoDB 3.6 on localhost:27017: %v", err)
	}
	defer session.Close()

	// Verify connection
	err = session.Ping()
	if err != nil {
		t.Fatalf("Failed to ping MongoDB 3.6: %v", err)
	}

	buildInfo, err := session.BuildInfo()
	if err != nil {
		t.Logf("Could not get build info: %v", err)
	} else {
		t.Logf("Connected to MongoDB version: %s", buildInfo.Version)
	}

	testSkipLimitFunctionality(t, session, "MongoDB 3.6 (original mgo)")
}

// TestSkipLimitMongoDB60WithModernMgo tests skip/limit functionality on MongoDB 6.0 using mgo.DialModernMGO
func TestSkipLimitMongoDB60WithModernMgo(t *testing.T) {
	t.Log("=== Testing Skip/Limit functionality on MongoDB 6.0 with mgo.DialModernMGO ===")

	session, err := mgo.DialModernMGO("mongodb://localhost:27018/test")
	if err != nil {
		t.Skipf("Failed to connect to MongoDB 6.0 on localhost:27018: %v", err)
	}
	defer session.Close()

	// Verify connection
	err = session.Ping()
	if err != nil {
		t.Fatalf("Failed to ping MongoDB 6.0: %v", err)
	}

	buildInfo, err := session.BuildInfo()
	if err != nil {
		t.Logf("Could not get build info: %v", err)
	} else {
		t.Logf("Connected to MongoDB version: %s", buildInfo.Version)
		if len(buildInfo.VersionArray) > 0 {
			t.Logf("Major version: %d", buildInfo.VersionArray[0])
		}
	}

	testSkipLimitFunctionality(t, session, "MongoDB 6.0 (modern mgo)")
}

// TestSkipLimitMongoDB36WithModernMgo tests skip/limit functionality on MongoDB 3.6 using mgo.DialModernMGO
func TestSkipLimitMongoDB36WithModernMgo(t *testing.T) {
	t.Log("=== Testing Skip/Limit functionality on MongoDB 3.6 with mgo.DialModernMGO ===")

	session, err := mgo.DialModernMGO("mongodb://localhost:27017/test")
	if err != nil {
		t.Skipf("Failed to connect to MongoDB 3.6 on localhost:27017 with modern mgo: %v", err)
	}
	defer session.Close()

	// Verify connection
	err = session.Ping()
	if err != nil {
		t.Fatalf("Failed to ping MongoDB 3.6 with modern mgo: %v", err)
	}

	buildInfo, err := session.BuildInfo()
	if err != nil {
		t.Logf("Could not get build info: %v", err)
	} else {
		t.Logf("Connected to MongoDB version: %s", buildInfo.Version)
	}

	testSkipLimitFunctionality(t, session, "MongoDB 3.6 (modern mgo)")
}

// TestSkipLimitComprehensive runs all skip/limit tests with performance timing
func TestSkipLimitComprehensive(t *testing.T) {
	t.Log("=== Comprehensive Skip/Limit Testing ===")

	tests := []struct {
		name string
		test func(*testing.T)
	}{
		{"MongoDB 3.6 with Original mgo.Dial", TestSkipLimitMongoDB36WithOriginalMgo},
		{"MongoDB 3.6 with mgo.DialModernMGO", TestSkipLimitMongoDB36WithModernMgo},
		{"MongoDB 6.0 with mgo.DialModernMGO", TestSkipLimitMongoDB60WithModernMgo},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			start := time.Now()
			test.test(t)
			duration := time.Since(start)
			t.Logf("Test completed in %v", duration)
		})
	}
}
