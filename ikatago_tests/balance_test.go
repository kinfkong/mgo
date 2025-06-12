package ikatago_tests

import (
	"testing"

	mgo "github.com/globalsign/mgo"
)

const (
	IKATAGO_CLUSTER_EARNINGS_CHECKPOINT_COLLECTION_BALANCE = "earnings_checkpoint"
	IKATAGO_CLUSTER_BALANCE_CHECKPOINT_COLLECTION_BALANCE  = "balance_checkpoint"
	IKATAGO_CLUSTER_CHECKPOINT_JOB_COLLECTION_BALANCE      = "checkpoint_job"
	DBNAME_TEST_BALANCE                                    = "ikatago_test"
)

func TestEnsureIndexesBalance(t *testing.T) {
	session := getSession(t)
	defer session.Close()

	// IKATAGO_CLUSTER_EARNINGS_CHECKPOINT_COLLECTION
	ecc := session.DB(DBNAME_TEST_BALANCE).C(IKATAGO_CLUSTER_EARNINGS_CHECKPOINT_COLLECTION_BALANCE)
	defer ecc.DropCollection()
	if err := ecc.EnsureIndex(mgo.Index{Key: []string{"userId"}, Unique: true}); err != nil {
		t.Fatal(err)
	}

	// IKATAGO_CLUSTER_BALANCE_CHECKPOINT_COLLECTION
	bcc := session.DB(DBNAME_TEST_BALANCE).C(IKATAGO_CLUSTER_BALANCE_CHECKPOINT_COLLECTION_BALANCE)
	defer bcc.DropCollection()
	if err := bcc.EnsureIndex(mgo.Index{Key: []string{"userId"}, Unique: true}); err != nil {
		t.Fatal(err)
	}

	// IKATAGO_CLUSTER_CHECKPOINT_JOB_COLLECTION
	cjc := session.DB(DBNAME_TEST_BALANCE).C(IKATAGO_CLUSTER_CHECKPOINT_JOB_COLLECTION_BALANCE)
	defer cjc.DropCollection()
	indices := []mgo.Index{
		{Key: []string{"userId", "jobType"}},
		{Key: []string{"userId", "jobType", "finished"}},
		{Key: []string{"userId", "jobType", "createdAt"}},
		{Key: []string{"finished"}},
	}
	for _, index := range indices {
		if err := cjc.EnsureIndex(index); err != nil {
			t.Fatalf("EnsureIndex failed: %v", err)
		}
	}

	indexes, err := cjc.Indexes()
	if err != nil {
		t.Fatal(err)
	}
	if len(indexes) != len(indices)+1 {
		t.Errorf("Expected %d indexes, got %d", len(indices)+1, len(indexes))
	}
}
