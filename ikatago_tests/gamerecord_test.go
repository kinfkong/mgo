package ikatago_tests

import (
	"testing"
)

const (
	IKATAGO_GAMERECORDS_COLLECTION_GAMERECORD = "game_records"
	DBNAME_TEST_GAMERECORD                    = "ikatago_test"
)

func TestEnsureIndicesGameRecord(t *testing.T) {
	session := getSession(t)
	defer session.Close()
	c := session.DB(DBNAME_TEST_GAMERECORD).C(IKATAGO_GAMERECORDS_COLLECTION_GAMERECORD)
	defer c.DropCollection()

	if err := c.EnsureIndexKey("userId", "recordType", "-createdAt"); err != nil {
		t.Fatal(err)
	}

	if err := c.EnsureIndexKey("gameId"); err != nil {
		t.Fatal(err)
	}
	if err := c.EnsureIndexKey("ownerId"); err != nil {
		t.Fatal(err)
	}
	if err := c.EnsureIndexKey("sgfId"); err != nil {
		t.Fatal(err)
	}
	if err := c.EnsureIndexKey("recordType", "used"); err != nil {
		t.Fatal(err)
	}
	if err := c.EnsureIndexKey("recordType", "share"); err != nil {
		t.Fatal(err)
	}
	if err := c.EnsureIndexKey("recordType", "status"); err != nil {
		t.Fatal(err)
	}
}
