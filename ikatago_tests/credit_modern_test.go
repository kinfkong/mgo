package ikatago_tests

import (
	"reflect"
	"testing"

	mgo "github.com/globalsign/mgo"
)

func TestModernEnsureIndexCredit(t *testing.T) {
	session := getModernSession(t)
	defer session.Close()

	c := session.DB(DBNAME_TEST_CREDIT).C(IKATAGO_CLUSTER_USER_CREDITS_COLLECTION_CREDIT)
	defer c.DropCollection()

	index := mgo.Index{
		Key: []string{"userId", "-createdAt"},
	}

	err := c.EnsureIndex(index)
	if err != nil {
		t.Fatalf("EnsureIndex failed: %v", err)
	}

	indexes, err := c.Indexes()
	if err != nil {
		t.Fatalf("Failed to get indexes: %v", err)
	}

	var found bool
	for _, idx := range indexes {
		if idx.Name == "userId_1_createdAt_-1" {
			found = true
			// Check keys if name matches
			if len(idx.Key) != 2 || idx.Key[0] != "userId" || idx.Key[1] != "-createdAt" {
				t.Errorf("Index 'userId_1_createdAt_-1' has wrong keys: %v", idx.Key)
			}
		}
	}

	if !found {
		t.Errorf("Index 'userId_1_createdAt_-1' not found")
	}
}

// from references/ikatago-service.md credit/credit.go
func TestModernEnsureIndicesCredit(t *testing.T) {
	session := getModernSession(t)
	defer session.Close()

	c := session.DB(DBNAME_TEST_CREDIT).C(IKATAGO_CLUSTER_USER_CREDITS_COLLECTION_CREDIT)
	defer c.DropCollection()

	index := mgo.Index{
		Key: []string{"userId"},
	}

	if err := c.EnsureIndex(index); err != nil {
		t.Fatal(err)
	}

	index = mgo.Index{
		Key: []string{"userId", "creditType"},
	}

	if err := c.EnsureIndex(index); err != nil {
		t.Fatal(err)
	}

	index = mgo.Index{
		Key: []string{"creditType"},
	}

	if err := c.EnsureIndex(index); err != nil {
		t.Fatal(err)
	}
	index = mgo.Index{
		Key: []string{"connectUserId"},
	}

	if err := c.EnsureIndex(index); err != nil {
		t.Fatal(err)
	}
	index = mgo.Index{
		Key: []string{"userId", "-createdAt"},
	}

	if err := c.EnsureIndex(index); err != nil {
		t.Fatal(err)
	}

	indexes, err := c.Indexes()
	if err != nil {
		t.Fatalf("Failed to get indexes: %v", err)
	}

	expectedIndexes := map[string][]string{
		"_id_":                  {"_id"},
		"userId_1":              {"userId"},
		"userId_1_creditType_1": {"userId", "creditType"},
		"creditType_1":          {"creditType"},
		"connectUserId_1":       {"connectUserId"},
		"userId_1_createdAt_-1": {"userId", "-createdAt"},
	}

	if len(indexes) != len(expectedIndexes) {
		t.Errorf("Expected %d indexes, but got %d", len(expectedIndexes), len(indexes))
		for _, idx := range indexes {
			t.Logf("Found index: %s, %v", idx.Name, idx.Key)
		}
	}

	for _, idx := range indexes {
		expectedKey, ok := expectedIndexes[idx.Name]
		if !ok {
			t.Errorf("Unexpected index found: %s", idx.Name)
			continue
		}

		if !reflect.DeepEqual(idx.Key, expectedKey) {
			t.Errorf("Index %s has wrong key. Expected %v, got %v", idx.Name, expectedKey, idx.Key)
		}
	}
}
