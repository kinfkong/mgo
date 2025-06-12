package ikatago_tests

import (
	"os"
	"testing"

	mgo "github.com/globalsign/mgo"
)

func getSession(t *testing.T) *mgo.Session {
	if os.Getenv("MODERN_MGO") == "1" {
		t.Skip("skipping original mgo tests for modern mgo")
	}
	session, err := mgo.Dial("localhost:27017")
	if err != nil {
		t.Fatalf("failed to dial mongo db (27017): %v", err)
	}
	session.SetMode(mgo.Monotonic, true)
	return session
}

func getModernSession(t *testing.T) *mgo.ModernMGO {
	if os.Getenv("MODERN_MGO") != "1" {
		t.Skip("skipping modern mgo tests for original mgo")
	}
	session, err := mgo.DialModernMGO("mongodb://localhost:27018/test")
	if err != nil {
		t.Fatalf("failed to dial modern mongo db (27018): %v", err)
	}
	session.SetMode(mgo.Monotonic, true)
	return session
}
