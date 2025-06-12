package ikatago_tests

import (
	"testing"

	mgo "github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

func TestModernInQuery(t *testing.T) {
	session := getModernSession(t)
	defer session.Close()

	c := session.DB(DBNAME_TEST_USAGE_SERVICE).C(IKATAGO_CLUSTER_USER_USAGES_COLLECTION_USAGE_SERVICE)
	defer c.DropCollection()

	c.Insert(&UsageServiceUsage{Finished: false, CommandIDs: []string{"a", "b"}})
	c.Insert(&UsageServiceUsage{Finished: false, CommandIDs: []string{"c", "d"}})
	c.Insert(&UsageServiceUsage{Finished: true, CommandIDs: []string{"a"}})

	query := bson.M{
		"finished": false,
		"commandIds": bson.M{
			"$in": []string{"c", "e"},
		},
	}
	var results []UsageServiceUsage
	err := c.Find(query).All(&results)
	if err != nil {
		t.Fatal(err)
	}

	if len(results) != 1 {
		t.Fatalf("Expected 1 result, got %d", len(results))
	}
	if results[0].CommandIDs[0] != "c" {
		t.Errorf("Expected command id 'c', got %v", results[0].CommandIDs)
	}
}

func TestModernIndexName(t *testing.T) {
	session := getModernSession(t)
	defer session.Close()

	c := session.DB(DBNAME_TEST_USAGE_SERVICE).C(IKATAGO_CLUSTER_USER_USAGES_COLLECTION_USAGE_SERVICE)
	defer c.DropCollection()

	index := mgo.Index{
		Key:  []string{"connectUserId", "finished", "-endedAt", "-startedAt"},
		Name: "usage_list_user_sort_index",
	}
	if err := c.EnsureIndex(index); err != nil {
		t.Fatal(err)
	}

	indexes, err := c.Indexes()
	if err != nil {
		t.Fatal(err)
	}

	var found bool
	for _, idx := range indexes {
		if idx.Name == "usage_list_user_sort_index" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Expected index with name 'usage_list_user_sort_index', but not found")
	}
}

func TestModernLimitAll(t *testing.T) {
	session := getModernSession(t)
	defer session.Close()

	c := session.DB(DBNAME_TEST_USAGE_SERVICE).C(IKATAGO_CLUSTER_USER_USAGES_COLLECTION_USAGE_SERVICE)
	defer c.DropCollection()

	for i := 0; i < 5; i++ {
		c.Insert(&UsageServiceUsage{})
	}

	var results []UsageServiceUsage
	err := c.Find(nil).Limit(3).All(&results)
	if err != nil {
		t.Fatal(err)
	}

	if len(results) != 3 {
		t.Fatalf("Expected 3 results, got %d", len(results))
	}
}

func TestModernUpdateIdWithInc(t *testing.T) {
	session := getModernSession(t)
	defer session.Close()
	c := session.DB(DBNAME_TEST_USAGE_SERVICE).C(IKATAGO_CLUSTER_USER_USAGES_COLLECTION_USAGE_SERVICE)
	defer c.DropCollection()

	usage := &UsageServiceUsage{ID: bson.NewObjectId(), Counter: 5}
	c.Insert(usage)

	err := c.UpdateId(usage.ID, bson.M{"$inc": bson.M{"counter": 1}})
	if err != nil {
		t.Fatalf("UpdateId with $inc failed: %v", err)
	}

	var updatedUsage UsageServiceUsage
	c.FindId(usage.ID).One(&updatedUsage)
	if updatedUsage.Counter != 6 {
		t.Errorf("Expected counter to be 6, got %d", updatedUsage.Counter)
	}
}
