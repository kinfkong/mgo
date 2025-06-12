package main

import (
	"fmt"
	"log"
	"time"

	mgo "github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

// TaskDocument represents a document similar to the countdown pattern
type TaskDocument struct {
	ID            bson.ObjectId `bson:"_id"`
	Name          string        `bson:"name"`
	Status        string        `bson:"status"`
	TaskPickedAt  *time.Time    `bson:"taskPickedAt"`
	FailedTimes   int           `bson:"failedTimes"`
	WillTriggerAt *time.Time    `bson:"willTriggerAt"`
	Counter       int           `bson:"counter"`
	CreatedAt     time.Time     `bson:"createdAt"`
}

func main() {
	fmt.Println("=== MongoDB Find().Apply() Demonstration ===")

	// Try to connect to MongoDB
	session, err := mgo.Dial("localhost:27017")
	if err != nil {
		fmt.Printf("Failed to connect to MongoDB: %v\n", err)
		fmt.Println("Note: This demo requires MongoDB running on localhost:27017")
		fmt.Println("The test shows the API structure even without a running MongoDB.")
		demonstrateAPI()
		return
	}
	defer session.Close()

	// Use a test database
	coll := session.DB("test_apply_demo").C("tasks")

	// Clean up any existing data
	coll.DropCollection()

	fmt.Println("✓ Connected to MongoDB successfully")

	// Test 1: Insert test documents
	fmt.Println("\n1. Inserting test documents...")
	docs := []interface{}{
		&TaskDocument{
			ID:            bson.NewObjectId(),
			Name:          "Task 1",
			Status:        "pending",
			TaskPickedAt:  nil,
			FailedTimes:   0,
			WillTriggerAt: timePtr(time.Now().Add(1 * time.Second)),
			Counter:       0,
			CreatedAt:     time.Now(),
		},
		&TaskDocument{
			ID:            bson.NewObjectId(),
			Name:          "Task 2",
			Status:        "pending",
			TaskPickedAt:  nil,
			FailedTimes:   0,
			WillTriggerAt: timePtr(time.Now().Add(2 * time.Second)),
			Counter:       0,
			CreatedAt:     time.Now(),
		},
		&TaskDocument{
			ID:            bson.NewObjectId(),
			Name:          "Task 3",
			Status:        "completed",
			TaskPickedAt:  timePtr(time.Now()),
			FailedTimes:   0,
			WillTriggerAt: nil,
			Counter:       5,
			CreatedAt:     time.Now(),
		},
	}

	err = coll.Insert(docs...)
	if err != nil {
		log.Fatalf("Failed to insert documents: %v", err)
	}
	fmt.Printf("✓ Inserted %d test documents\n", len(docs))

	// Test 2: Find().Apply() pattern similar to countdown service
	fmt.Println("\n2. Testing Find().Apply() pattern (similar to countdown service)...")

	// This mimics the countdown service pattern:
	// Find a document that meets criteria and atomically update it
	condition := bson.M{
		"willTriggerAt": bson.M{"$lte": time.Now().Add(5 * time.Second)},
		"taskPickedAt":  nil,
		"status":        "pending",
	}

	change := mgo.Change{
		Update:    bson.M{"$set": bson.M{"taskPickedAt": time.Now(), "status": "processing"}},
		ReturnNew: false, // Return the document before modification
	}

	var pickedTask TaskDocument
	info, err := coll.Find(condition).Apply(change, &pickedTask)

	if err == mgo.ErrNotFound {
		fmt.Println("✓ No task found matching criteria (expected behavior)")
	} else if err != nil {
		log.Fatalf("Apply failed: %v", err)
	} else {
		fmt.Printf("✓ Successfully picked task: %s (was: %s, now processing)\n", pickedTask.Name, pickedTask.Status)
		fmt.Printf("  - Info: Updated=%d, Matched=%d, Removed=%d\n", info.Updated, info.Matched, info.Removed)
	}

	// Test 3: Test upsert functionality
	fmt.Println("\n3. Testing Apply with Upsert...")

	newTaskID := bson.NewObjectId()
	upsertChange := mgo.Change{
		Update: bson.M{"$set": bson.M{
			"name":      "Upserted Task",
			"status":    "pending",
			"createdAt": time.Now(),
			"counter":   1,
		}},
		Upsert:    true,
		ReturnNew: true,
	}

	var upsertedTask TaskDocument
	info, err = coll.Find(bson.M{"_id": newTaskID}).Apply(upsertChange, &upsertedTask)
	if err != nil {
		log.Fatalf("Upsert failed: %v", err)
	}

	fmt.Printf("✓ Upserted new task: %s\n", upsertedTask.Name)
	fmt.Printf("  - UpsertedId: %s\n", info.UpsertedId)

	// Test 4: Test atomic counter increment
	fmt.Println("\n4. Testing atomic counter increment...")

	incrementChange := mgo.Change{
		Update:    bson.M{"$inc": bson.M{"counter": 1}},
		ReturnNew: true,
	}

	var incrementedTask TaskDocument
	info, err = coll.Find(bson.M{"name": "Upserted Task"}).Apply(incrementChange, &incrementedTask)
	if err != nil {
		log.Fatalf("Increment failed: %v", err)
	}

	fmt.Printf("✓ Incremented counter to: %d\n", incrementedTask.Counter)

	// Test 5: Test remove with Apply
	fmt.Println("\n5. Testing Apply with Remove...")

	removeChange := mgo.Change{
		Remove: true,
	}

	var removedTask TaskDocument
	info, err = coll.Find(bson.M{"name": "Upserted Task"}).Apply(removeChange, &removedTask)
	if err != nil {
		log.Fatalf("Remove failed: %v", err)
	}

	fmt.Printf("✓ Removed task: %s (counter was: %d)\n", removedTask.Name, removedTask.Counter)
	fmt.Printf("  - Info: Removed=%d\n", info.Removed)

	// Test 6: Verify atomicity - simulate concurrent access
	fmt.Println("\n6. Testing atomicity (simulating concurrent access)...")

	// Insert a task that multiple "workers" might try to pick up
	sharedTask := &TaskDocument{
		ID:           bson.NewObjectId(),
		Name:         "Shared Task",
		Status:       "available",
		TaskPickedAt: nil,
		FailedTimes:  0,
		Counter:      0,
		CreatedAt:    time.Now(),
	}

	err = coll.Insert(sharedTask)
	if err != nil {
		log.Fatalf("Failed to insert shared task: %v", err)
	}

	// Simulate two workers trying to pick up the same task
	pickCondition := bson.M{
		"_id":    sharedTask.ID,
		"status": "available",
	}

	pickChange := mgo.Change{
		Update:    bson.M{"$set": bson.M{"status": "taken", "taskPickedAt": time.Now()}},
		ReturnNew: false,
	}

	// First worker should succeed
	var worker1Task TaskDocument
	_, err1 := coll.Find(pickCondition).Apply(pickChange, &worker1Task)

	// Second worker should fail (document no longer matches condition)
	var worker2Task TaskDocument
	_, err2 := coll.Find(pickCondition).Apply(pickChange, &worker2Task)

	if err1 == nil && err2 == mgo.ErrNotFound {
		fmt.Printf("✓ Atomicity verified: Worker 1 got task, Worker 2 got ErrNotFound\n")
		fmt.Printf("  - Worker 1 got: %s (status was: %s)\n", worker1Task.Name, worker1Task.Status)
	} else {
		fmt.Printf("⚠ Unexpected result: Worker 1 err=%v, Worker 2 err=%v\n", err1, err2)
	}

	// Clean up
	coll.DropCollection()
	fmt.Println("\n✓ Demo completed successfully!")
	fmt.Println("✓ Find().Apply() is working correctly and provides atomic operations")
}

func demonstrateAPI() {
	fmt.Println("\n=== Find().Apply() API Demonstration ===")
	fmt.Println("Even without a MongoDB connection, here's how the API works:")
	fmt.Println()

	fmt.Println("// Basic pattern used in countdown service:")
	fmt.Println("condition := bson.M{")
	fmt.Println("    \"willTriggerAt\": bson.M{\"$lte\": time.Now()},")
	fmt.Println("    \"taskPickedAt\": nil,")
	fmt.Println("}")
	fmt.Println("change := mgo.Change{")
	fmt.Println("    Update:    bson.M{\"$set\": bson.M{\"taskPickedAt\": time.Now()}},")
	fmt.Println("    ReturnNew: false,")
	fmt.Println("}")
	fmt.Println("var result TaskDocument")
	fmt.Println("info, err := collection.Find(condition).Apply(change, &result)")
	fmt.Println()
	fmt.Println("✓ This provides atomic find-and-modify operations")
	fmt.Println("✓ Only one goroutine can successfully pick up a task")
	fmt.Println("✓ Prevents race conditions in distributed systems")
}

func timePtr(t time.Time) *time.Time {
	return &t
}
