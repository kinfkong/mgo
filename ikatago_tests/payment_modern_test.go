package ikatago_tests

import (
	"testing"

	mgo "github.com/globalsign/mgo"
)

func TestModernEnsureIndicesPayment(t *testing.T) {
	session := getModernSession(t)
	defer session.Close()

	c := session.DB(DBNAME_TEST_PAYMENT).C(IKATAGO_CLUSTER_USER_PAYMENTS_COLLECTION)
	defer c.DropCollection()

	indices := []mgo.Index{
		{Key: []string{"userId"}},
		{Key: []string{"userId", "-createdAt"}},
	}
	for _, index := range indices {
		if err := c.EnsureIndex(index); err != nil {
			t.Fatalf("EnsureIndex failed: %v", err)
		}
	}

	indexes, err := c.Indexes()
	if err != nil {
		t.Fatalf("Failed to get indexes: %v", err)
	}

	if len(indexes) != len(indices)+1 { // +1 for _id index
		t.Errorf("Expected %d indexes, got %d", len(indices)+1, len(indexes))
	}
}
