package ikatago_tests

import (
	"testing"

	mgo "github.com/globalsign/mgo"
)

func TestModernEnsureIndicesUsage(t *testing.T) {
	session := getModernSession(t)
	defer session.Close()

	c := session.DB(DBNAME_TEST_USAGE).C(IKATAGO_CLUSTER_USER_USAGES_COLLECTION_USAGE)
	defer c.DropCollection()

	indices := []mgo.Index{
		{Key: []string{"nodeId"}},
		{Key: []string{"nodeOwnerUserId"}},
		{Key: []string{"nodename"}},
		{Key: []string{"connectUserId"}},
		{Key: []string{"connectUsername"}},
		{Key: []string{"connectUsername", "finished"}},
		{Key: []string{"connectUserId", "finished"}},
		{Key: []string{"finished"}},
		{Key: []string{"serialId"}},
		{Key: []string{"connectUserId", "serialId"}},
		{Key: []string{"connectUserId", "startedAt"}},
		{Key: []string{"nodeOwnerUserId", "serialId"}},
		{Key: []string{"-startedAt"}},
		{Key: []string{"-lastUpdatedAt"}},
		{Key: []string{"finished", "-endedAt", "-startedAt"}, Name: "usage_list_sort_index"},
		{Key: []string{"connectUserId", "finished", "-endedAt", "-startedAt"}, Name: "usage_list_user_sort_index"},
		{Key: []string{"connectUsername", "finished", "commandIds"}, Name: "usage_by_command_ids"},
		{Key: []string{"commandIds"}, Name: "usage_simple_command_ids"},
		{Key: []string{"connectUserId", "gameId"}},
	}
	for _, index := range indices {
		if err := c.EnsureIndex(index); err != nil {
			t.Fatalf("EnsureIndex failed for %+v: %v", index, err)
		}
	}

	dbIndexes, err := c.Indexes()
	if err != nil {
		t.Fatal(err)
	}

	if len(dbIndexes) != len(indices)+1 { // +1 for _id index
		t.Errorf("Expected %d indexes, got %d", len(indices)+1, len(dbIndexes))
	}
}
