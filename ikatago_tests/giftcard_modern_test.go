package ikatago_tests

import (
	"testing"

	mgo "github.com/globalsign/mgo"
)

func TestModernEnsureIndicesGiftcard(t *testing.T) {
	session := getModernSession(t)
	defer session.Close()

	c := session.DB(DBNAME_TEST_GIFTCARD).C(IKATAGO_CLUSTER_GIFTCARD_COLLECTION_GIFTCARD)
	defer c.DropCollection()

	indices := []mgo.Index{
		{Key: []string{"giftCardCode"}, Unique: true},
		{Key: []string{"createUserId", "-createdAt"}},
	}
	for _, index := range indices {
		if err := c.EnsureIndex(index); err != nil {
			t.Fatal(err)
		}
	}
}
