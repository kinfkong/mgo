package ikatago_tests

import (
	"testing"
	"time"

	mgo "github.com/globalsign/mgo"
	"github.com/globalsign/mgo/bson"
)

const (
	IKATAGO_USER_ACCOUNT_COLLECTION_ACC            = "user_accounts"
	IKATAGO_USER_VERIFICATION_CODES_COLLECTION_ACC = "user_verification_codes"
	IKATAGO_USER_ACTIVATE_CODE_COLLECTION_ACC      = "user_activate_code"
	IKATAGO_SOCKET_IO_TOKEN_COLLECTION_ACC         = "socket_io_token"
	DBNAME_TEST_ACCOUNT_ENSURE                     = "ikatago_test"
)

type AccountEnsure struct {
	ID                  bson.ObjectId `bson:"_id,omitempty"`
	Phone               string        `bson:"phone"`
	Email               string        `bson:"email"`
	Token               string        `bson:"token"`
	MembershipExpiresAt time.Time     `bson:"membershipExpiresAt"`
	ReferCode           string        `bson:"referCode"`
	CreatedAt           time.Time     `bson:"createdAt"`
}

type VerificationCodeEnsure struct {
	Phone string `bson:"phone"`
	Email string `bson:"email"`
	Type  string `bson:"type"`
}

type ActivateCodeEnsure struct {
	ActivateCode string `bson:"activateCode"`
	UserID       string `bson:"userId"`
}

type SocketIOTokenEnsure struct {
	Token string `bson:"token"`
}

func TestEnsureIndicesAccount(t *testing.T) {
	session := getSession(t)
	defer session.Close()

	// IKATAGO_USER_ACCOUNT_COLLECTION
	c := session.DB(DBNAME_TEST_ACCOUNT_ENSURE).C(IKATAGO_USER_ACCOUNT_COLLECTION_ACC)
	defer c.DropCollection()

	indices := []mgo.Index{
		{Key: []string{"phone"}},
		{Key: []string{"email"}},
		{Key: []string{"token"}, Unique: true},
		{Key: []string{"membershipExpiresAt"}},
		{Key: []string{"referCode"}},
		{Key: []string{"createdAt"}},
	}
	for _, index := range indices {
		if err := c.EnsureIndex(index); err != nil {
			t.Fatalf("EnsureIndex failed for collection %s: %v", IKATAGO_USER_ACCOUNT_COLLECTION_ACC, err)
		}
	}
	// Test token uniqueness
	err := c.Insert(&AccountEnsure{Token: "token1"}, &AccountEnsure{Token: "token1"})
	if !mgo.IsDup(err) {
		t.Errorf("Expected duplicate key error for token, but got %v", err)
	}

	// IKATAGO_USER_VERIFICATION_CODES_COLLECTION
	vc := session.DB(DBNAME_TEST_ACCOUNT_ENSURE).C(IKATAGO_USER_VERIFICATION_CODES_COLLECTION_ACC)
	defer vc.DropCollection()
	indices = []mgo.Index{
		{Key: []string{"phone", "type"}},
		{Key: []string{"email", "type"}},
	}
	for _, index := range indices {
		if err := vc.EnsureIndex(index); err != nil {
			t.Fatalf("EnsureIndex failed for collection %s: %v", IKATAGO_USER_VERIFICATION_CODES_COLLECTION_ACC, err)
		}
	}

	// IKATAGO_USER_ACTIVATE_CODE_COLLECTION
	ac := session.DB(DBNAME_TEST_ACCOUNT_ENSURE).C(IKATAGO_USER_ACTIVATE_CODE_COLLECTION_ACC)
	defer ac.DropCollection()
	indices = []mgo.Index{
		{Key: []string{"activateCode"}, Unique: true},
		{Key: []string{"userId"}},
	}
	for _, index := range indices {
		if err := ac.EnsureIndex(index); err != nil {
			t.Fatalf("EnsureIndex failed for collection %s: %v", IKATAGO_USER_ACTIVATE_CODE_COLLECTION_ACC, err)
		}
	}
	err = ac.Insert(&ActivateCodeEnsure{ActivateCode: "code1"}, &ActivateCodeEnsure{ActivateCode: "code1"})
	if !mgo.IsDup(err) {
		t.Errorf("Expected duplicate key error for activateCode, but got %v", err)
	}

	// IKATAGO_SOCKET_IO_TOKEN_COLLECTION
	st := session.DB(DBNAME_TEST_ACCOUNT_ENSURE).C(IKATAGO_SOCKET_IO_TOKEN_COLLECTION_ACC)
	defer st.DropCollection()
	index := mgo.Index{
		Key:    []string{"token"},
		Unique: true,
	}
	if err := st.EnsureIndex(index); err != nil {
		t.Fatal(err)
	}
	err = st.Insert(&SocketIOTokenEnsure{Token: "token1"}, &SocketIOTokenEnsure{Token: "token1"})
	if !mgo.IsDup(err) {
		t.Errorf("Expected duplicate key error for socket io token, but got %v", err)
	}
}
