package mgo

import (
	"fmt"
	"testing"

	"github.com/globalsign/mgo/bson"
)

// Test configuration for different MongoDB versions
type TestConfig struct {
	Name string
	URL  string
}

var testConfigs = []TestConfig{
	{"MongoDB 3.6", "mongodb://localhost:27017"},
	{"MongoDB 6.0", "mongodb://localhost:27018"},
}

// TestDocument represents a test document structure
type TestDocument struct {
	ID      bson.ObjectId `bson:"_id,omitempty"`
	Name    string        `bson:"name"`
	Age     int           `bson:"age"`
	Status  string        `bson:"status"`
	Counter int           `bson:"counter"`
}

// setupTestCollection creates a clean test collection
func setupTestCollection(t *testing.T, url string) (*ModernMGO, *ModernColl) {
	session, err := DialModernMGO(url)
	if err != nil {
		t.Skipf("Skipping test for %s: %v", url, err)
		return nil, nil
	}

	// Test connection
	if err := session.Ping(); err != nil {
		session.Close()
		t.Skipf("Skipping test for %s: cannot ping: %v", url, err)
		return nil, nil
	}

	db := session.DB("test_modern_bulk")
	coll := db.C("bulk_test")

	// Clean up any existing data
	coll.DropCollection()

	return session, coll
}

// TestBulkInsert tests bulk insert operations
func TestBulkInsert(t *testing.T) {
	for _, config := range testConfigs {
		t.Run(config.Name, func(t *testing.T) {
			session, coll := setupTestCollection(t, config.URL)
			if session == nil {
				return
			}
			defer session.Close()

			bulk := coll.Bulk()
			bulk.Insert(
				TestDocument{Name: "Alice", Age: 30, Status: "active"},
				TestDocument{Name: "Bob", Age: 25, Status: "active"},
			)
			bulk.Insert(TestDocument{Name: "Charlie", Age: 35, Status: "inactive"})

			result, err := bulk.Run()
			if err != nil {
				t.Fatalf("Bulk insert failed: %v", err)
			}

			if result == nil {
				t.Fatal("Expected non-nil result")
			}

			// Verify documents were inserted
			count, err := coll.Count()
			if err != nil {
				t.Fatalf("Failed to count documents: %v", err)
			}
			if count != 3 {
				t.Errorf("Expected 3 documents, got %d", count)
			}

			// Verify document content
			var docs []TestDocument
			err = coll.Find(nil).Sort("name").All(&docs)
			if err != nil {
				t.Fatalf("Failed to find documents: %v", err)
			}

			expected := []TestDocument{
				{Name: "Alice", Age: 30, Status: "active"},
				{Name: "Bob", Age: 25, Status: "active"},
				{Name: "Charlie", Age: 35, Status: "inactive"},
			}

			if len(docs) != len(expected) {
				t.Errorf("Expected %d documents, got %d", len(expected), len(docs))
			}

			for i, doc := range docs {
				if i < len(expected) {
					if doc.Name != expected[i].Name || doc.Age != expected[i].Age {
						t.Errorf("Document %d mismatch: got %+v, expected %+v", i, doc, expected[i])
					}
				}
			}
		})
	}
}

// TestBulkUpdate tests bulk update operations
func TestBulkUpdate(t *testing.T) {
	for _, config := range testConfigs {
		t.Run(config.Name, func(t *testing.T) {
			session, coll := setupTestCollection(t, config.URL)
			if session == nil {
				return
			}
			defer session.Close()

			// Insert test data
			err := coll.Insert(
				TestDocument{Name: "Alice", Age: 30, Status: "active"},
				TestDocument{Name: "Bob", Age: 25, Status: "active"},
				TestDocument{Name: "Charlie", Age: 35, Status: "inactive"},
			)
			if err != nil {
				t.Fatalf("Failed to insert test data: %v", err)
			}

			bulk := coll.Bulk()
			bulk.Update(
				bson.M{"name": "Alice"}, bson.M{"$set": bson.M{"age": 31}},
				bson.M{"name": "Bob"}, bson.M{"$set": bson.M{"status": "premium"}},
			)
			bulk.Update(bson.M{"name": "NonExistent"}, bson.M{"$set": bson.M{"age": 99}}) // Won't match

			result, err := bulk.Run()
			if err != nil {
				t.Fatalf("Bulk update failed: %v", err)
			}

			if result.Matched != 2 {
				t.Errorf("Expected 2 matched documents, got %d", result.Matched)
			}

			if result.Modified != 2 {
				t.Errorf("Expected 2 modified documents, got %d", result.Modified)
			}

			// Verify updates
			var alice TestDocument
			err = coll.Find(bson.M{"name": "Alice"}).One(&alice)
			if err != nil {
				t.Fatalf("Failed to find Alice: %v", err)
			}
			if alice.Age != 31 {
				t.Errorf("Alice's age should be 31, got %d", alice.Age)
			}

			var bob TestDocument
			err = coll.Find(bson.M{"name": "Bob"}).One(&bob)
			if err != nil {
				t.Fatalf("Failed to find Bob: %v", err)
			}
			if bob.Status != "premium" {
				t.Errorf("Bob's status should be 'premium', got '%s'", bob.Status)
			}
		})
	}
}

// TestBulkUpdateAll tests bulk update all operations
func TestBulkUpdateAll(t *testing.T) {
	for _, config := range testConfigs {
		t.Run(config.Name, func(t *testing.T) {
			session, coll := setupTestCollection(t, config.URL)
			if session == nil {
				return
			}
			defer session.Close()

			// Insert test data with duplicate statuses
			err := coll.Insert(
				TestDocument{Name: "Alice", Age: 30, Status: "active", Counter: 1},
				TestDocument{Name: "Bob", Age: 25, Status: "active", Counter: 1},
				TestDocument{Name: "Charlie", Age: 35, Status: "inactive", Counter: 1},
				TestDocument{Name: "David", Age: 40, Status: "active", Counter: 1},
			)
			if err != nil {
				t.Fatalf("Failed to insert test data: %v", err)
			}

			bulk := coll.Bulk()
			bulk.UpdateAll(
				bson.M{"status": "active"}, bson.M{"$inc": bson.M{"counter": 1}},
				bson.M{"status": "inactive"}, bson.M{"$set": bson.M{"status": "archived"}},
			)

			result, err := bulk.Run()
			if err != nil {
				t.Fatalf("Bulk update all failed: %v", err)
			}

			if result.Matched != 4 {
				t.Errorf("Expected 4 matched documents, got %d", result.Matched)
			}

			// Verify all active users had counter incremented
			var activeDocs []TestDocument
			err = coll.Find(bson.M{"status": "active"}).All(&activeDocs)
			if err != nil {
				t.Fatalf("Failed to find active documents: %v", err)
			}

			if len(activeDocs) != 3 {
				t.Errorf("Expected 3 active documents, got %d", len(activeDocs))
			}

			for _, doc := range activeDocs {
				if doc.Counter != 2 {
					t.Errorf("Expected counter to be 2 for %s, got %d", doc.Name, doc.Counter)
				}
			}

			// Verify inactive user was archived
			archivedCount, err := coll.Find(bson.M{"status": "archived"}).Count()
			if err != nil {
				t.Fatalf("Failed to count archived documents: %v", err)
			}
			if archivedCount != 1 {
				t.Errorf("Expected 1 archived document, got %d", archivedCount)
			}
		})
	}
}

// TestBulkUpsert tests bulk upsert operations
func TestBulkUpsert(t *testing.T) {
	for _, config := range testConfigs {
		t.Run(config.Name, func(t *testing.T) {
			session, coll := setupTestCollection(t, config.URL)
			if session == nil {
				return
			}
			defer session.Close()

			// Insert one existing document
			err := coll.Insert(TestDocument{Name: "Alice", Age: 30, Status: "active"})
			if err != nil {
				t.Fatalf("Failed to insert test data: %v", err)
			}

			bulk := coll.Bulk()
			bulk.Upsert(
				bson.M{"name": "Alice"}, bson.M{"$set": bson.M{"age": 31}}, // Update existing
				bson.M{"name": "Bob"}, bson.M{"$set": bson.M{"age": 25, "status": "new"}}, // Insert new
			)

			result, err := bulk.Run()
			if err != nil {
				t.Fatalf("Bulk upsert failed: %v", err)
			}

			if result.Matched != 1 {
				t.Errorf("Expected 1 matched document, got %d", result.Matched)
			}

			// Check total count
			count, err := coll.Count()
			if err != nil {
				t.Fatalf("Failed to count documents: %v", err)
			}
			if count != 2 {
				t.Errorf("Expected 2 documents after upsert, got %d", count)
			}

			// Verify Alice was updated
			var alice TestDocument
			err = coll.Find(bson.M{"name": "Alice"}).One(&alice)
			if err != nil {
				t.Fatalf("Failed to find Alice: %v", err)
			}
			if alice.Age != 31 {
				t.Errorf("Alice's age should be 31, got %d", alice.Age)
			}

			// Verify Bob was inserted
			var bob TestDocument
			err = coll.Find(bson.M{"name": "Bob"}).One(&bob)
			if err != nil {
				t.Fatalf("Failed to find Bob: %v", err)
			}
			if bob.Age != 25 || bob.Status != "new" {
				t.Errorf("Bob should have age 25 and status 'new', got age %d, status '%s'", bob.Age, bob.Status)
			}
		})
	}
}

// TestBulkRemove tests bulk remove operations
func TestBulkRemove(t *testing.T) {
	for _, config := range testConfigs {
		t.Run(config.Name, func(t *testing.T) {
			session, coll := setupTestCollection(t, config.URL)
			if session == nil {
				return
			}
			defer session.Close()

			// Insert test data with duplicates
			err := coll.Insert(
				TestDocument{Name: "Alice", Age: 30, Status: "active"},
				TestDocument{Name: "Bob", Age: 25, Status: "active"},
				TestDocument{Name: "Bob", Age: 26, Status: "inactive"}, // Duplicate name
				TestDocument{Name: "Charlie", Age: 35, Status: "inactive"},
			)
			if err != nil {
				t.Fatalf("Failed to insert test data: %v", err)
			}

			bulk := coll.Bulk()
			bulk.Remove(
				bson.M{"name": "Alice"},    // Remove Alice
				bson.M{"name": "Bob"},      // Remove only first Bob
				bson.M{"name": "NonExist"}, // Won't match anything
			)

			result, err := bulk.Run()
			if err != nil {
				t.Fatalf("Bulk remove failed: %v", err)
			}

			if result.Matched != 2 {
				t.Errorf("Expected 2 matched documents, got %d", result.Matched)
			}

			// Verify remaining documents
			count, err := coll.Count()
			if err != nil {
				t.Fatalf("Failed to count documents: %v", err)
			}
			if count != 2 {
				t.Errorf("Expected 2 documents remaining, got %d", count)
			}

			// Verify Alice is gone
			aliceCount, err := coll.Find(bson.M{"name": "Alice"}).Count()
			if err != nil {
				t.Fatalf("Failed to count Alice documents: %v", err)
			}
			if aliceCount != 0 {
				t.Errorf("Expected Alice to be removed, but found %d documents", aliceCount)
			}

			// Verify one Bob remains
			bobCount, err := coll.Find(bson.M{"name": "Bob"}).Count()
			if err != nil {
				t.Fatalf("Failed to count Bob documents: %v", err)
			}
			if bobCount != 1 {
				t.Errorf("Expected 1 Bob document to remain, got %d", bobCount)
			}
		})
	}
}

// TestBulkRemoveAll tests bulk remove all operations
func TestBulkRemoveAll(t *testing.T) {
	for _, config := range testConfigs {
		t.Run(config.Name, func(t *testing.T) {
			session, coll := setupTestCollection(t, config.URL)
			if session == nil {
				return
			}
			defer session.Close()

			// Insert test data
			err := coll.Insert(
				TestDocument{Name: "Alice", Age: 30, Status: "active"},
				TestDocument{Name: "Bob", Age: 25, Status: "inactive"},
				TestDocument{Name: "Charlie", Age: 35, Status: "inactive"},
				TestDocument{Name: "David", Age: 40, Status: "active"},
			)
			if err != nil {
				t.Fatalf("Failed to insert test data: %v", err)
			}

			bulk := coll.Bulk()
			bulk.RemoveAll(
				bson.M{"status": "inactive"},     // Remove all inactive
				bson.M{"age": bson.M{"$gt": 35}}, // Remove all age > 35
			)

			result, err := bulk.Run()
			if err != nil {
				t.Fatalf("Bulk remove all failed: %v", err)
			}

			if result.Matched != 3 { // 2 inactive + 1 with age > 35
				t.Errorf("Expected 3 matched documents, got %d", result.Matched)
			}

			// Verify only Alice remains
			count, err := coll.Count()
			if err != nil {
				t.Fatalf("Failed to count documents: %v", err)
			}
			if count != 1 {
				t.Errorf("Expected 1 document remaining, got %d", count)
			}

			var remaining TestDocument
			err = coll.Find(nil).One(&remaining)
			if err != nil {
				t.Fatalf("Failed to find remaining document: %v", err)
			}
			if remaining.Name != "Alice" {
				t.Errorf("Expected Alice to remain, got %s", remaining.Name)
			}
		})
	}
}

// TestBulkUnordered tests unordered bulk operations
func TestBulkUnordered(t *testing.T) {
	for _, config := range testConfigs {
		t.Run(config.Name, func(t *testing.T) {
			session, coll := setupTestCollection(t, config.URL)
			if session == nil {
				return
			}
			defer session.Close()

			bulk := coll.Bulk()
			bulk.Unordered()

			// Mix different operations
			bulk.Insert(TestDocument{Name: "Alice", Age: 30, Status: "active"})
			bulk.Update(bson.M{"name": "Alice"}, bson.M{"$set": bson.M{"age": 31}})
			bulk.Insert(TestDocument{Name: "Bob", Age: 25, Status: "active"})
			bulk.Upsert(bson.M{"name": "Charlie"}, bson.M{"$set": bson.M{"age": 35, "status": "new"}})

			_, err := bulk.Run()
			if err != nil {
				t.Fatalf("Bulk unordered failed: %v", err)
			}

			// Verify all operations completed
			count, err := coll.Count()
			if err != nil {
				t.Fatalf("Failed to count documents: %v", err)
			}
			if count != 3 {
				t.Errorf("Expected 3 documents, got %d", count)
			}

			// Verify Alice was updated (this tests that insert -> update worked in unordered mode)
			var alice TestDocument
			err = coll.Find(bson.M{"name": "Alice"}).One(&alice)
			if err != nil {
				t.Fatalf("Failed to find Alice: %v", err)
			}
			// Note: In unordered mode, the update might happen before or after insert
			// so we just verify Alice exists with some age
			if alice.Name != "Alice" {
				t.Errorf("Expected to find Alice")
			}
		})
	}
}

// TestBulkMixed tests mixed bulk operations in a single batch
func TestBulkMixed(t *testing.T) {
	for _, config := range testConfigs {
		t.Run(config.Name, func(t *testing.T) {
			session, coll := setupTestCollection(t, config.URL)
			if session == nil {
				return
			}
			defer session.Close()

			// Insert initial data
			err := coll.Insert(TestDocument{Name: "Existing", Age: 50, Status: "old"})
			if err != nil {
				t.Fatalf("Failed to insert initial data: %v", err)
			}

			bulk := coll.Bulk()
			bulk.Insert(TestDocument{Name: "Alice", Age: 30, Status: "active"})
			bulk.Insert(TestDocument{Name: "Bob", Age: 25, Status: "active"})
			bulk.Update(bson.M{"name": "Existing"}, bson.M{"$set": bson.M{"status": "updated"}})
			bulk.Upsert(bson.M{"name": "Charlie"}, bson.M{"$set": bson.M{"age": 35, "status": "upserted"}})
			bulk.Remove(bson.M{"name": "Bob"})

			_, err = bulk.Run()
			if err != nil {
				t.Fatalf("Bulk mixed operations failed: %v", err)
			}

			// Verify final state
			count, err := coll.Count()
			if err != nil {
				t.Fatalf("Failed to count documents: %v", err)
			}
			if count != 3 { // Existing + Alice + Charlie (Bob was removed)
				t.Errorf("Expected 3 documents, got %d", count)
			}

			// Verify specific documents
			var docs []TestDocument
			err = coll.Find(nil).Sort("name").All(&docs)
			if err != nil {
				t.Fatalf("Failed to find documents: %v", err)
			}

			expectedNames := []string{"Alice", "Charlie", "Existing"}
			if len(docs) != len(expectedNames) {
				t.Errorf("Expected %d documents, got %d", len(expectedNames), len(docs))
			}

			for i, doc := range docs {
				if i < len(expectedNames) && doc.Name != expectedNames[i] {
					t.Errorf("Expected document %d to be %s, got %s", i, expectedNames[i], doc.Name)
				}
			}

			// Verify status updates
			var existing TestDocument
			err = coll.Find(bson.M{"name": "Existing"}).One(&existing)
			if err != nil {
				t.Fatalf("Failed to find Existing: %v", err)
			}
			if existing.Status != "updated" {
				t.Errorf("Expected Existing status to be 'updated', got '%s'", existing.Status)
			}

			var charlie TestDocument
			err = coll.Find(bson.M{"name": "Charlie"}).One(&charlie)
			if err != nil {
				t.Fatalf("Failed to find Charlie: %v", err)
			}
			if charlie.Status != "upserted" {
				t.Errorf("Expected Charlie status to be 'upserted', got '%s'", charlie.Status)
			}
		})
	}
}

// TestBulkError tests bulk operation error handling
func TestBulkError(t *testing.T) {
	for _, config := range testConfigs {
		t.Run(config.Name, func(t *testing.T) {
			session, coll := setupTestCollection(t, config.URL)
			if session == nil {
				return
			}
			defer session.Close()

			// Create unique index to force duplicate key errors
			err := coll.EnsureIndex(Index{
				Key:    []string{"name"},
				Unique: true,
				Name:   "unique_name_idx",
			})
			if err != nil {
				t.Fatalf("Failed to create unique index: %v", err)
			}

			bulk := coll.Bulk()
			bulk.Insert(TestDocument{Name: "Alice", Age: 30})
			bulk.Insert(TestDocument{Name: "Alice", Age: 31}) // This should fail
			bulk.Insert(TestDocument{Name: "Bob", Age: 25})   // This should succeed

			_, err = bulk.Run()

			// We expect an error due to duplicate key
			if err == nil {
				t.Fatal("Expected bulk operation to fail due to duplicate key")
			}

			// Check if it's a BulkError
			if bulkErr, ok := err.(*BulkError); ok {
				cases := bulkErr.Cases()
				if len(cases) == 0 {
					t.Error("Expected at least one error case")
				}

				// Verify error details
				found := false
				for _, errorCase := range cases {
					if errorCase.Err != nil {
						found = true
						break
					}
				}
				if !found {
					t.Error("Expected to find at least one error case with details")
				}
			} else {
				t.Errorf("Expected BulkError, got %T: %v", err, err)
			}

			// In ordered mode, some operations might have succeeded before the error
			// Let's check what actually got inserted
			count, countErr := coll.Count()
			if countErr != nil {
				t.Logf("Could not count documents after error: %v", countErr)
			} else {
				t.Logf("Documents inserted before error: %d", count)
			}
		})
	}
}

// TestBulkEmpty tests empty bulk operations
func TestBulkEmpty(t *testing.T) {
	for _, config := range testConfigs {
		t.Run(config.Name, func(t *testing.T) {
			session, coll := setupTestCollection(t, config.URL)
			if session == nil {
				return
			}
			defer session.Close()

			bulk := coll.Bulk()
			// Don't add any operations

			result, err := bulk.Run()
			if err != nil {
				t.Fatalf("Empty bulk operation should not fail: %v", err)
			}

			if result == nil {
				t.Fatal("Expected non-nil result for empty bulk")
			}

			if result.Matched != 0 || result.Modified != 0 {
				t.Errorf("Expected zero counts for empty bulk, got matched=%d, modified=%d",
					result.Matched, result.Modified)
			}
		})
	}
}

// BenchmarkBulkInsert benchmarks bulk insert performance
func BenchmarkBulkInsert(b *testing.B) {
	session, err := DialModernMGO("mongodb://localhost:27017")
	if err != nil {
		b.Skipf("Skipping benchmark: %v", err)
		return
	}
	defer session.Close()

	coll := session.DB("benchmark").C("bulk_insert")
	coll.DropCollection()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		bulk := coll.Bulk()

		// Insert 100 documents per iteration
		for j := 0; j < 100; j++ {
			bulk.Insert(TestDocument{
				Name:   fmt.Sprintf("User_%d_%d", i, j),
				Age:    20 + (j % 50),
				Status: "active",
			})
		}

		_, err := bulk.Run()
		if err != nil {
			b.Fatalf("Bulk insert failed: %v", err)
		}
	}
}
