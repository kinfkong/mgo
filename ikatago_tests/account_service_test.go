package ikatago_tests

import (
	"testing"
	"time"

	"github.com/globalsign/mgo/bson"
)

const (
	IKATAGO_USER_ACCOUNT_COLLECTION_ACCOUNT            = "user_accounts"
	IKATAGO_USER_VERIFICATION_CODES_COLLECTION_ACCOUNT = "user_verification_codes"
	DBNAME_TEST_ACCOUNT                                = "ikatago_test"
)

type Account struct {
	ID          bson.ObjectId `bson:"_id,omitempty"`
	Token       string        `bson:"token"`
	UpdatedAt   time.Time     `bson:"updatedAt"`
	LastLoginAt time.Time     `bson:"lastLoginAt"`
	CreatedAt   time.Time     `bson:"createdAt"`
}

type VerificationCode struct {
	ID        bson.ObjectId `bson:"_id,omitempty"`
	Code      string        `bson:"code"`
	Type      string        `bson:"type"`
	ExpiresAt time.Time     `bson:"expiresAt"`
	CreatedAt time.Time     `bson:"createdAt"`
}

// 1. Add unit tests the update({_id: xxx}, {$set: {...}}) is working correctly.
func TestUpdateSet(t *testing.T) {
	session := getSession(t)
	defer session.Close()

	c := session.DB(DBNAME_TEST_ACCOUNT).C(IKATAGO_USER_ACCOUNT_COLLECTION_ACCOUNT)
	defer c.DropCollection()

	account := &Account{ID: bson.NewObjectId(), Token: "old-token"}
	c.Insert(account)

	newToken := "new-token"
	now := time.Now().Truncate(time.Millisecond)
	err := c.Update(bson.M{"_id": account.ID}, bson.M{"$set": bson.M{
		"token":       newToken,
		"updatedAt":   now,
		"lastLoginAt": now,
	}})

	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	var updatedAccount Account
	c.FindId(account.ID).One(&updatedAccount)

	if updatedAccount.Token != newToken {
		t.Errorf("Expected token %s, got %s", newToken, updatedAccount.Token)
	}
	if !updatedAccount.UpdatedAt.Equal(now) {
		t.Errorf("Expected updatedAt %v, got %v", now, updatedAccount.UpdatedAt)
	}
	if !updatedAccount.LastLoginAt.Equal(now) {
		t.Errorf("Expected lastLoginAt %v, got %v", now, updatedAccount.LastLoginAt)
	}
}

// 2. Add unit tests to test: err := vc.Find(query).Sort("-createdAt").Limit(1).One(&code)
func TestFindSortLimitOne(t *testing.T) {
	session := getSession(t)
	defer session.Close()

	vc := session.DB(DBNAME_TEST_ACCOUNT).C(IKATAGO_USER_VERIFICATION_CODES_COLLECTION_ACCOUNT)
	defer vc.DropCollection()

	now := time.Now().Truncate(time.Millisecond)
	code1 := &VerificationCode{ID: bson.NewObjectId(), Code: "123", Type: "fast_login", ExpiresAt: now.Add(time.Hour), CreatedAt: now.Add(-time.Minute)}
	code2 := &VerificationCode{ID: bson.NewObjectId(), Code: "123", Type: "fast_login", ExpiresAt: now.Add(time.Hour), CreatedAt: now}
	vc.Insert(code1, code2)

	query := bson.M{
		"code": "123",
		"type": "fast_login",
		"expiresAt": bson.M{
			"$gte": now,
		},
	}
	var result VerificationCode
	err := vc.Find(query).Sort("-createdAt").Limit(1).One(&result)
	if err != nil {
		t.Fatalf("Find().Sort().Limit().One() failed: %v", err)
	}
	if result.ID != code2.ID {
		t.Errorf("Expected to get the latest code (%s), but got (%s).", code2.ID.Hex(), result.ID.Hex())
	}
}

// 3. add unit tests to test err = c.UpdateId(existingAccount.ID, existingAccount) updateId method
func TestUpdateId(t *testing.T) {
	session := getSession(t)
	defer session.Close()

	c := session.DB(DBNAME_TEST_ACCOUNT).C(IKATAGO_USER_ACCOUNT_COLLECTION_ACCOUNT)
	defer c.DropCollection()

	account := &Account{ID: bson.NewObjectId(), Token: "token1"}
	c.Insert(account)

	account.Token = "token2"
	err := c.UpdateId(account.ID, account)
	if err != nil {
		t.Fatalf("UpdateId failed: %v", err)
	}

	var updatedAccount Account
	c.FindId(account.ID).One(&updatedAccount)
	if updatedAccount.Token != "token2" {
		t.Errorf("Expected token 'token2', got '%s'", updatedAccount.Token)
	}
}

// 4. add pipe unit tests
func TestAccountPipe(t *testing.T) {
	session := getSession(t)
	defer session.Close()

	c := session.DB(DBNAME_TEST_ACCOUNT).C(IKATAGO_USER_ACCOUNT_COLLECTION_ACCOUNT)
	defer c.DropCollection()

	loc, _ := time.LoadLocation("Asia/Shanghai") // corresponds to +08:00
	t1, _ := time.ParseInLocation("2006-01-02", "2024-01-01", loc)
	t2, _ := time.ParseInLocation("2006-01-02", "2024-01-01", loc)
	t3, _ := time.ParseInLocation("2006-01-02", "2024-01-02", loc)

	c.Insert(&Account{CreatedAt: t1}, &Account{CreatedAt: t2}, &Account{CreatedAt: t3})

	pipeline := []bson.M{
		{
			"$project": bson.M{
				"yearMonthDay": bson.M{
					"$dateToString": bson.M{"format": "%Y-%m-%d", "timezone": "+08:00", "date": "$createdAt"},
				},
			},
		},
		{
			"$group": bson.M{
				"_id": "$yearMonthDay",
				"totalNewUsers": bson.M{
					"$sum": 1,
				},
			},
		},
	}
	var resp []bson.M
	err := c.Pipe(pipeline).All(&resp)
	if err != nil {
		t.Fatalf("Pipe failed: %v", err)
	}

	if len(resp) != 2 {
		t.Fatalf("Expected 2 groups, got %d", len(resp))
	}

	for _, group := range resp {
		day := group["_id"]
		count := group["totalNewUsers"]
		if day == "2024-01-01" {
			if count != 2 {
				t.Errorf("Expected 2 users for 2024-01-01, got %d", count)
			}
		} else if day == "2024-01-02" {
			if count != 1 {
				t.Errorf("Expected 1 user for 2024-01-02, got %d", count)
			}
		} else {
			t.Errorf("Unexpected group: %s", day)
		}
	}
}

// 5. Add unit tests to test the removeId and removeAll
func TestRemove(t *testing.T) {
	session := getSession(t)
	defer session.Close()

	c := session.DB(DBNAME_TEST_ACCOUNT).C(IKATAGO_USER_ACCOUNT_COLLECTION_ACCOUNT)
	defer c.DropCollection()

	acc1 := &Account{ID: bson.NewObjectId(), Token: "1"}
	acc2 := &Account{ID: bson.NewObjectId(), Token: "2"}
	acc3 := &Account{ID: bson.NewObjectId(), Token: "3"}

	c.Insert(acc1, acc2, acc3)

	// Test RemoveId
	err := c.RemoveId(acc1.ID)
	if err != nil {
		t.Fatalf("RemoveId failed: %v", err)
	}
	count, _ := c.Count()
	if count != 2 {
		t.Errorf("Expected 2 docs after RemoveId, got %d", count)
	}

	// Test RemoveAll
	_, err = c.RemoveAll(bson.M{"token": "2"})
	if err != nil {
		t.Fatalf("RemoveAll failed: %v", err)
	}
	count, _ = c.Count()
	if count != 1 {
		t.Errorf("Expected 1 doc after RemoveAll, got %d", count)
	}

	var lastAccount Account
	c.Find(nil).One(&lastAccount)
	if lastAccount.Token != "3" {
		t.Errorf("Wrong document remaining")
	}
}

// from references/ikatago-service.md account/account.go
func TestEnsureIndicesAccountService(t *testing.T) {
	session := getSession(t)
	defer session.Close()
	c := session.DB(DBNAME_TEST_ACCOUNT).C(IKATAGO_USER_ACCOUNT_COLLECTION_ACCOUNT)
	defer c.DropCollection()

	if err := c.EnsureIndexKey("token"); err != nil {
		t.Fatal(err)
	}
	if err := c.EnsureIndexKey("phone"); err != nil {
		t.Fatal(err)
	}
	if err := c.EnsureIndexKey("auths.uid"); err != nil {
		t.Fatal(err)
	}
	if err := c.EnsureIndexKey("teamId"); err != nil {
		t.Fatal(err)
	}
	if err := c.EnsureIndexKey("inviterId"); err != nil {
		t.Fatal(err)
	}
}
