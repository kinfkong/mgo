// modern_demo_test.go - Comprehensive tests for all mgo API methods

package mgo

import (
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/globalsign/mgo/bson"
)

// Test data structures
type TestUser struct {
	ID   bson.ObjectId `bson:"_id"`
	Name string        `bson:"name"`
	Age  int           `bson:"age"`
}

// TestSessionOperations tests session-level operations
func TestSessionOperations(t *testing.T) {
	// Test mgo.Dial() equivalent - DialModernMGO
	session, err := DialModernMGO("mongodb://localhost:27017/testdb")
	if err != nil {
		t.Skipf("MongoDB not available: %v", err)
	}
	defer session.Close()

	// Test session.Copy()
	sessionCopy := session.Copy()
	if sessionCopy == nil {
		t.Error("session.Copy() returned nil")
	}
	sessionCopy.Close()

	// Test session.Clone()
	sessionClone := session.Clone()
	if sessionClone == nil {
		t.Error("session.Clone() returned nil")
	}
	sessionClone.Close()

	// Test session.SetMode()
	session.SetMode(Monotonic, true)
	if session.Mode() != Monotonic {
		t.Error("session.SetMode() did not set mode correctly")
	}

	// Test session.DB()
	db := session.DB("testdb")
	if db == nil {
		t.Error("session.DB() returned nil")
	}

	// Test session.Ping()
	if err := session.Ping(); err != nil {
		t.Errorf("session.Ping() failed: %v", err)
	}

	// Test session.BuildInfo()
	buildInfo, err := session.BuildInfo()
	if err != nil {
		t.Errorf("session.BuildInfo() failed: %v", err)
	}
	if buildInfo.Version == "" {
		t.Error("BuildInfo.Version is empty")
	}
}

// TestDatabaseOperations tests database-level operations
func TestDatabaseOperations(t *testing.T) {
	session, err := DialModernMGO("mongodb://localhost:27017/testdb")
	if err != nil {
		t.Skipf("MongoDB not available: %v", err)
	}
	defer session.Close()

	db := session.DB("testdb")

	// Test db.C()
	coll := db.C("testcoll")
	if coll == nil {
		t.Error("db.C() returned nil")
	}

	// Test db.GridFS()
	gfs := db.GridFS("fs")
	if gfs == nil {
		t.Error("db.GridFS() returned nil")
	}
}

// TestCollectionOperations tests collection-level operations
func TestCollectionOperations(t *testing.T) {
	session, err := DialModernMGO("mongodb://localhost:27017/testdb")
	if err != nil {
		t.Skipf("MongoDB not available: %v", err)
	}
	defer session.Close()

	db := session.DB("testdb")
	coll := db.C("testcoll")

	// Clean up before testing
	coll.DropCollection()

	// Test c.Insert()
	user := TestUser{
		ID:   bson.NewObjectId(),
		Name: "John Doe",
		Age:  30,
	}
	if err := coll.Insert(user); err != nil {
		t.Errorf("c.Insert() failed: %v", err)
	}

	// Test c.Find()
	query := coll.Find(bson.M{"name": "John Doe"})
	if query == nil {
		t.Error("c.Find() returned nil")
	}

	// Test c.One() (via query)
	var foundUser TestUser
	if err := query.One(&foundUser); err != nil {
		t.Errorf("query.One() failed: %v", err)
	}
	if foundUser.Name != "John Doe" {
		t.Errorf("Expected name 'John Doe', got '%s'", foundUser.Name)
	}

	// Test c.Update()
	if err := coll.Update(bson.M{"_id": user.ID}, bson.M{"$set": bson.M{"age": 31}}); err != nil {
		t.Errorf("c.Update() failed: %v", err)
	}

	// Test c.FindId()
	foundUser = TestUser{}
	if err := coll.FindId(user.ID).One(&foundUser); err != nil {
		t.Errorf("c.FindId().One() failed: %v", err)
	}
	if foundUser.Age != 31 {
		t.Errorf("Expected age 31, got %d", foundUser.Age)
	}

	// Test c.Upsert()
	changeInfo, err := coll.Upsert(bson.M{"name": "Jane Doe"}, bson.M{"$set": bson.M{"name": "Jane Doe", "age": 25}})
	if err != nil {
		t.Errorf("c.Upsert() failed: %v", err)
	}
	if changeInfo == nil {
		t.Error("c.Upsert() returned nil changeInfo")
	}

	// Test c.Count()
	count, err := coll.Count()
	if err != nil {
		t.Errorf("c.Count() failed: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected count 2, got %d", count)
	}

	// Test c.All() (via query)
	var users []TestUser
	if err := coll.Find(bson.M{}).All(&users); err != nil {
		t.Errorf("query.All() failed: %v", err)
	}
	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}

	// Test c.Sort() (via query)
	users = []TestUser{}
	if err := coll.Find(bson.M{}).Sort("age").All(&users); err != nil {
		t.Errorf("query.Sort().All() failed: %v", err)
	}
	if len(users) != 2 {
		t.Errorf("Expected 2 users, got %d", len(users))
	}
	if users[0].Age > users[1].Age {
		t.Error("Sort() did not work correctly")
	}

	// Test c.EnsureIndex()
	index := Index{
		Key:    []string{"name"},
		Unique: true,
		Name:   "name_unique",
	}
	if err := coll.EnsureIndex(index); err != nil {
		t.Errorf("c.EnsureIndex() failed: %v", err)
	}

	// Test c.Remove()
	if err := coll.Remove(bson.M{"name": "Jane Doe"}); err != nil {
		t.Errorf("c.Remove() failed: %v", err)
	}

	// Test c.RemoveAll()
	changeInfo, err = coll.RemoveAll(bson.M{})
	if err != nil {
		t.Errorf("c.RemoveAll() failed: %v", err)
	}
	if changeInfo == nil || changeInfo.Removed == 0 {
		t.Error("c.RemoveAll() did not remove any documents")
	}
}

// TestGridFSOperations tests GridFS operations
func TestGridFSOperations(t *testing.T) {
	session, err := DialModernMGO("mongodb://localhost:27017/testdb")
	if err != nil {
		t.Skipf("MongoDB not available: %v", err)
	}
	defer session.Close()

	db := session.DB("testdb")
	gfs := db.GridFS("fs")

	// Clean up any existing test files
	err = gfs.Remove("testfile.txt")
	if err != nil && err != ErrNotFound {
		t.Logf("Warning: failed to clean up previous test files: %v", err)
	}
	err = gfs.Remove("newname.txt")
	if err != nil && err != ErrNotFound {
		t.Logf("Warning: failed to clean up previous test files: %v", err)
	}

	// Also try to remove any files with custom IDs from previous test runs
	customId := fmt.Sprintf("custom-id-%d", time.Now().UnixNano())
	gfs.RemoveId(customId) // Ignore error if it doesn't exist

	// Test gfs.Create()
	file, err := gfs.Create("testfile.txt")
	if err != nil {
		t.Errorf("gfs.Create() failed: %v", err)
		return
	}

	// Test file.Write()
	testData := []byte("Hello, GridFS!")
	n, err := file.Write(testData)
	if err != nil {
		t.Errorf("file.Write() failed: %v", err)
		return
	}
	if n != len(testData) {
		t.Errorf("Expected to write %d bytes, wrote %d", len(testData), n)
		return
	}

	// Test file getters and setters
	file.SetId(customId) // Use unique ID
	if file.Id() != customId {
		t.Error("file.SetId()/Id() failed")
		return
	}

	file.SetName("newname.txt")
	if file.Name() != "newname.txt" {
		t.Error("file.SetName()/Name() failed")
		return
	}

	file.SetContentType("text/plain")
	if file.ContentType() != "text/plain" {
		t.Error("file.SetContentType()/ContentType() failed")
		return
	}

	file.SetChunkSize(1024)

	metadata := bson.M{"author": "test"}
	file.SetMeta(metadata)

	var retrievedMeta bson.M
	if err := file.GetMeta(&retrievedMeta); err != nil {
		t.Errorf("file.GetMeta() failed: %v", err)
		return
	}

	file.SetUploadDate(time.Now())

	// Test file.Close()
	if err := file.Close(); err != nil {
		t.Errorf("file.Close() failed: %v", err)
		return
	}

	// Test gfs.Open()
	readFile, err := gfs.Open("newname.txt")
	if err != nil {
		t.Errorf("gfs.Open() failed: %v", err)
		return
	}

	// Test file.Read()
	readData := make([]byte, len(testData))
	n, err = readFile.Read(readData)
	if err != nil && err != io.EOF {
		t.Errorf("file.Read() failed: %v", err)
		return
	}
	if string(readData[:n]) != string(testData) {
		t.Errorf("Expected to read '%s', got '%s'", string(testData), string(readData[:n]))
		return
	}

	// Test file properties
	if readFile.Size() != int64(len(testData)) {
		t.Errorf("Expected file size %d, got %d", len(testData), readFile.Size())
		return
	}

	// Clean up test files
	err = gfs.Remove("newname.txt")
	if err != nil {
		t.Logf("Warning: failed to clean up test file: %v", err)
	}
}

// TestDataStructures tests data structures and constants
func TestDataStructures(t *testing.T) {
	// Test mgo.Index{}
	index := Index{
		Key:         []string{"name", "-age"},
		Unique:      true,
		Background:  true,
		Sparse:      true,
		ExpireAfter: 24 * time.Hour,
		Name:        "test_index",
	}
	if len(index.Key) != 2 {
		t.Error("Index.Key not set correctly")
	}

	// Test mgo.ErrNotFound
	if ErrNotFound == nil {
		t.Error("ErrNotFound is nil")
	}
	if ErrNotFound.Error() != "not found" {
		t.Errorf("Expected 'not found', got '%s'", ErrNotFound.Error())
	}

	// Test mgo.Monotonic constant
	if Monotonic != 1 {
		t.Errorf("Expected Monotonic=1, got %d", Monotonic)
	}
}
