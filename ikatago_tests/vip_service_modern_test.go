package ikatago_tests

import (
	"reflect"
	"testing"
	"time"

	mgo "github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// 1. should add unit test to test the sort, limit, one methods are working correctly
func TestModernSortLimitOne(t *testing.T) {
	session := getModernSession(t)
	defer session.Close()

	c := session.DB(DBNAME_TEST_VIP).C(IKATAGO_CLUSTER_VIP_STAT_DATA_COLLECTION)
	defer c.DropCollection()

	data1 := &VIPStatData{ID: bson.NewObjectId(), Date: "2024-01-01", UpdatedAt: time.Now().Add(-time.Hour)}
	data2 := &VIPStatData{ID: bson.NewObjectId(), Date: "2024-01-02", UpdatedAt: time.Now()}
	c.Insert(data1)
	c.Insert(data2)

	result := &VIPStatData{}
	err := c.Find(bson.M{}).Sort("-updatedAt").Limit(1).One(result)
	if err != nil {
		t.Fatalf("Find().Sort().Limit().One() failed: %v", err)
	}

	if result.Date != "2024-01-02" {
		t.Errorf("Expected date '2024-01-02', got '%s'", result.Date)
	}
}

// Test to verify that Sort("-updatedAt") sorts in descending order (reverse order)
func TestModernSortDescendingOrder(t *testing.T) {
	session := getModernSession(t)
	defer session.Close()

	c := session.DB(DBNAME_TEST_VIP).C(IKATAGO_CLUSTER_VIP_STAT_DATA_COLLECTION)
	defer c.DropCollection()

	// Create test data with different updatedAt times
	now := time.Now()
	data1 := &VIPStatData{ID: bson.NewObjectId(), Date: "2024-01-01", UpdatedAt: now.Add(-3 * time.Hour)} // oldest
	data2 := &VIPStatData{ID: bson.NewObjectId(), Date: "2024-01-02", UpdatedAt: now.Add(-2 * time.Hour)} // middle
	data3 := &VIPStatData{ID: bson.NewObjectId(), Date: "2024-01-03", UpdatedAt: now.Add(-1 * time.Hour)} // newest

	// Insert in random order
	c.Insert(data2)
	c.Insert(data1)
	c.Insert(data3)

	// Query with Sort("-updatedAt") to get descending order
	var results []VIPStatData
	err := c.Find(bson.M{}).Sort("-updatedAt").All(&results)
	if err != nil {
		t.Fatalf("Find().Sort('-updatedAt').All() failed: %v", err)
	}

	if len(results) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(results))
	}

	// Verify descending order: newest first, oldest last
	if results[0].Date != "2024-01-03" {
		t.Errorf("Expected first result to be '2024-01-03' (newest), got '%s'", results[0].Date)
	}
	if results[1].Date != "2024-01-02" {
		t.Errorf("Expected second result to be '2024-01-02' (middle), got '%s'", results[1].Date)
	}
	if results[2].Date != "2024-01-01" {
		t.Errorf("Expected third result to be '2024-01-01' (oldest), got '%s'", results[2].Date)
	}

	// Verify the actual timestamps are in descending order
	if !results[0].UpdatedAt.After(results[1].UpdatedAt) {
		t.Errorf("First result UpdatedAt (%v) should be after second result UpdatedAt (%v)",
			results[0].UpdatedAt, results[1].UpdatedAt)
	}
	if !results[1].UpdatedAt.After(results[2].UpdatedAt) {
		t.Errorf("Second result UpdatedAt (%v) should be after third result UpdatedAt (%v)",
			results[1].UpdatedAt, results[2].UpdatedAt)
	}
}

// 2. should add unit test to test the Find().Select().All() methods are working correctly.
func TestModernFindSelectAll(t *testing.T) {
	session := getModernSession(t)
	defer session.Close()

	uc := session.DB(DBNAME_TEST_VIP).C(IKATAGO_USER_ACCOUNT_COLLECTION)
	defer uc.DropCollection()

	now := time.Now()
	user1 := &UserAccount{ID: bson.NewObjectId(), MembershipExpiresAt: now, MembershipAutoRenew: true}
	user2 := &UserAccount{ID: bson.NewObjectId(), MembershipExpiresAt: now.Add(48 * time.Hour), MembershipAutoRenew: true}
	user3 := &UserAccount{ID: bson.NewObjectId(), MembershipExpiresAt: now, MembershipAutoRenew: false}

	uc.Insert(user1, user2, user3)

	autoRenewUsers := make([]UserAccount, 0)
	err := uc.Find(bson.M{
		"membershipExpiresAt": bson.M{
			"$lte": now.Add(24 * time.Hour),
			"$gte": now.Add(-24 * time.Hour),
		},
		"membershipAutoRenew": true,
	}).Select(bson.M{"_id": 1}).All(&autoRenewUsers)

	if err != nil {
		t.Fatalf("Find().Select().All() failed: %v", err)
	}

	if len(autoRenewUsers) != 1 {
		t.Fatalf("Expected 1 user, got %d", len(autoRenewUsers))
	}

	if autoRenewUsers[0].ID != user1.ID {
		t.Errorf("Expected user ID %s, got %s", user1.ID, autoRenewUsers[0].ID)
	}
	// Check that other fields are zeroed
	if !autoRenewUsers[0].MembershipExpiresAt.IsZero() {
		t.Errorf("Expected MembershipExpiresAt to be zero, but got %v", autoRenewUsers[0].MembershipExpiresAt)
	}
}

// Test to verify that Select(bson.M{"_id": 1}) only returns the _id field
func TestModernSelectOnlyId(t *testing.T) {
	session := getModernSession(t)
	defer session.Close()

	c := session.DB(DBNAME_TEST_VIP).C(IKATAGO_CLUSTER_VIP_STAT_DATA_COLLECTION)
	defer c.DropCollection()

	// Create test data with multiple fields
	now := time.Now()
	testData := &VIPStatData{
		ID:        bson.NewObjectId(),
		Date:      "2024-01-01",
		UpdatedAt: now,
	}

	err := c.Insert(testData)
	if err != nil {
		t.Fatalf("Insert failed: %v", err)
	}

	// Query with Select to only get _id field
	var results []VIPStatData
	err = c.Find(bson.M{}).Select(bson.M{"_id": 1}).All(&results)
	if err != nil {
		t.Fatalf("Find().Select().All() failed: %v", err)
	}

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}

	result := results[0]

	// Verify that _id is populated
	if result.ID.Hex() == "" {
		t.Errorf("Expected _id to be populated, but it's empty")
	}
	if result.ID != testData.ID {
		t.Errorf("Expected _id to be %s, got %s", testData.ID.Hex(), result.ID.Hex())
	}

	// Verify that other fields are zero/empty (not selected)
	if result.Date != "" {
		t.Errorf("Expected Date to be empty (not selected), but got '%s'", result.Date)
	}
	if !result.UpdatedAt.IsZero() {
		t.Errorf("Expected UpdatedAt to be zero (not selected), but got %v", result.UpdatedAt)
	}
}

// 3. should add unit tests to test the pipe methods with similar parameters in the following code are working correctly.
func TestModernPipe(t *testing.T) {
	session := getModernSession(t)
	defer session.Close()

	uc := session.DB(DBNAME_TEST_VIP).C(IKATAGO_CLUSTER_USER_USAGES_COLLECTION)
	defer uc.DropCollection()

	usages := []Usage{
		{SerialID: 1, Finished: true, VIP: true, VirtualTotalCost: 10},
		{SerialID: 2, Finished: true, VIP: true, VirtualTotalCost: 20},
		{SerialID: 3, Finished: false, VIP: true, VirtualTotalCost: 30},
		{SerialID: 4, Finished: true, VIP: false, VirtualTotalCost: 40},
		{SerialID: 5, Finished: true, VIP: true, VirtualTotalCost: 50},
	}
	for _, u := range usages {
		uc.Insert(u)
	}

	// Test pipeline 1
	pipeline1 := []bson.M{
		{
			"$match": bson.M{
				"serialId": bson.M{
					"$gt": 0,
					"$lt": 4,
				},
				"finished": true,
				"vip":      true,
			},
		},
		{
			"$group": bson.M{
				"_id": nil,
				"totalUsedComsumption": bson.M{
					"$sum": "$virtualTotalCost",
				},
				"maxSerialId": bson.M{
					"$max": "$serialId",
				},
			},
		},
	}
	resp := []bson.M{}
	err := uc.Pipe(pipeline1).All(&resp)
	if err != nil {
		t.Fatalf("Pipe 1 failed: %v", err)
	}

	if len(resp) != 1 {
		t.Fatalf("Expected 1 result from pipe 1, got %d", len(resp))
	}
	if resp[0]["totalUsedComsumption"] != float64(30) {
		t.Errorf("Expected totalUsedComsumption 30, got %v", resp[0]["totalUsedComsumption"])
	}
	if resp[0]["maxSerialId"] != int64(2) {
		t.Errorf("Expected maxSerialId 2, got %v", resp[0]["maxSerialId"])
	}

	// Test pipeline 2
	pipeline2 := []bson.M{
		{
			"$match": bson.M{
				"serialId": bson.M{
					"$gt": 2,
				},
				"vip": true,
			},
		},
		{
			"$group": bson.M{
				"_id": nil,
				"totalUsedComsumption": bson.M{
					"$sum": "$virtualTotalCost",
				},
			},
		},
	}
	resp2 := []bson.M{}
	err = uc.Pipe(pipeline2).All(&resp2)
	if err != nil {
		t.Fatalf("Pipe 2 failed: %v", err)
	}
	if len(resp2) != 1 {
		t.Fatalf("Expected 1 result from pipe 2, got %d", len(resp2))
	}
	if resp2[0]["totalUsedComsumption"] != float64(80) {
		t.Errorf("Expected totalUsedComsumption 80, got %v", resp2[0]["totalUsedComsumption"])
	}
}

// 4. should add unit tests to test the upsert methods are working correctly.
func TestModernUpsert(t *testing.T) {
	session := getModernSession(t)
	defer session.Close()

	vipc := session.DB(DBNAME_TEST_VIP).C(IKATAGO_CLUSTER_VIP_STAT_DATA_COLLECTION)
	defer vipc.DropCollection()

	// Insert
	newData := &VIPStatData{Date: "2024-01-01", UpdatedAt: time.Now()}
	changeInfo, err := vipc.Upsert(bson.M{"date": newData.Date}, newData)
	if err != nil {
		t.Fatalf("Upsert failed: %v", err)
	}
	if changeInfo.UpsertedId == nil {
		t.Errorf("Expected an upserted ID, got nil")
	}
	if changeInfo.Updated != 0 {
		t.Errorf("Expected 0 updated docs, got %d", changeInfo.Updated)
	}

	// Update
	newData.UpdatedAt = time.Now().Add(time.Hour)
	changeInfo, err = vipc.Upsert(bson.M{"date": newData.Date}, newData)
	if err != nil {
		t.Fatalf("Upsert failed: %v", err)
	}
	if changeInfo.UpsertedId != nil {
		t.Errorf("Expected no upserted ID on update, got %v", changeInfo.UpsertedId)
	}
	if changeInfo.Updated != 1 {
		t.Errorf("Expected 1 updated doc, got %d", changeInfo.Updated)
	}

	count, _ := vipc.Count()
	if count != 1 {
		t.Errorf("Expected 1 document in collection, got %d", count)
	}
}

// 5. should add unit tests to test the count method is working correctly.
func TestModernCount(t *testing.T) {
	session := getModernSession(t)
	defer session.Close()

	uc := session.DB(DBNAME_TEST_VIP).C(IKATAGO_CLUSTER_USER_USAGES_COLLECTION)
	defer uc.DropCollection()

	usages := []Usage{
		{Finished: false, VIP: true},
		{Finished: false, VIP: true},
		{Finished: true, VIP: true},
		{Finished: false, VIP: false},
	}
	for _, u := range usages {
		uc.Insert(u)
	}

	count, err := uc.Find(bson.M{
		"vip":      true,
		"finished": false,
	}).Count()

	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}

	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}
}

// From /Users/jinggangwang/gochess/ikatago-service/vip/vip.go
func TestModernEnsureIndicesVIP(t *testing.T) {
	session := getModernSession(t)
	defer session.Close()

	c := session.DB(DBNAME_TEST_VIP).C(IKATAGO_CLUSTER_VIP_STAT_DATA_COLLECTION)
	defer c.DropCollection()

	index := mgo.Index{
		Key:    []string{"date"},
		Unique: true,
	}
	if err := c.EnsureIndex(index); err != nil {
		t.Fatal(err)
	}

	// check uniqueness
	err := c.Insert(&VIPStatData{Date: "2024-01-01"})
	if err != nil {
		t.Fatalf("first insert failed: %v", err)
	}
	err = c.Insert(&VIPStatData{Date: "2024-01-01"})
	if err == nil {
		t.Fatalf("expected duplicate key error, but got nil")
	} else if _, ok := err.(mongo.WriteException); !ok {
		t.Fatalf("expected mongo.WriteException, but got %T", err)
	}

	index = mgo.Index{
		Key: []string{"-updatedAt"},
	}

	if err := c.EnsureIndex(index); err != nil {
		t.Fatal(err)
	}

	indexes, err := c.Indexes()
	if err != nil {
		t.Fatalf("Failed to get indexes: %v", err)
	}

	expectedIndexes := map[string]mgo.Index{
		"_id_":         {Key: []string{"_id"}},
		"date_1":       {Key: []string{"date"}, Unique: true},
		"updatedAt_-1": {Key: []string{"-updatedAt"}},
	}

	if len(indexes) != len(expectedIndexes) {
		t.Errorf("Expected %d indexes, but got %d", len(expectedIndexes), len(indexes))
	}

	for _, idx := range indexes {
		expectedIdx, ok := expectedIndexes[idx.Name]
		if !ok {
			t.Errorf("Unexpected index found: %s", idx.Name)
			continue
		}
		if !reflect.DeepEqual(idx.Key, expectedIdx.Key) {
			t.Errorf("Index %s has wrong key. Expected %v, got %v", idx.Name, expectedIdx.Key, idx.Key)
		}
		if idx.Unique != expectedIdx.Unique {
			t.Errorf("Index %s has wrong unique property. Expected %v, got %v", idx.Name, expectedIdx.Unique, idx.Unique)
		}
	}
}
