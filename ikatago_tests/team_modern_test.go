package ikatago_tests

import (
	"testing"

	mgo "github.com/globalsign/mgo"
)

func TestModernEnsureIndicesTeam(t *testing.T) {
	session := getModernSession(t)
	defer session.Close()

	tc := session.DB(DBNAME_TEST_TEAM).C(IKATAGO_CLUSTER_TEAM_COLLECTION_TEAM)
	defer tc.DropCollection()
	if err := tc.EnsureIndex(mgo.Index{Key: []string{"adminUserId"}}); err != nil {
		t.Fatal(err)
	}

	crc := session.DB(DBNAME_TEST_TEAM).C(IKATAGO_CLUSTER_TEAM_CHARGE_RECORD_COLLECTION_TEAM)
	defer crc.DropCollection()
	crcIndices := []mgo.Index{
		{Key: []string{"teamId"}},
		{Key: []string{"teamId", "-createdAt"}},
	}
	for _, index := range crcIndices {
		if err := crc.EnsureIndex(index); err != nil {
			t.Fatal(err)
		}
	}

	trc := session.DB(DBNAME_TEST_TEAM).C(IKATAGO_CLUSTER_TEAM_TRANSFER_RECORD_COLLECTION_TEAM)
	defer trc.DropCollection()
	trcIndices := []mgo.Index{
		{Key: []string{"teamId"}},
		{Key: []string{"teamId", "-createdAt"}},
	}
	for _, index := range trcIndices {
		if err := trc.EnsureIndex(index); err != nil {
			t.Fatal(err)
		}
	}
}
