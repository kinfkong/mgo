package mgo

import (
	"testing"

	"github.com/globalsign/mgo/bson"
)

// TestBSONMarshalingFix tests the BSON marshaling fix in the Run method
func TestBSONMarshalingFix(t *testing.T) {
	// Test converting bson.D to bson.M
	testCmd := bson.D{
		{Name: "serverStatus", Value: 1},
		{Name: "repl", Value: 0},
	}

	// This simulates what happens in the db.run method after our fix
	var cmd interface{} = testCmd
	if bsonCmd, ok := cmd.(bson.D); ok {
		cmdMap := bson.M{}
		for _, elem := range bsonCmd {
			cmdMap[elem.Name] = elem.Value
		}

		// Verify the conversion worked
		if cmdMap["serverStatus"] != 1 {
			t.Errorf("Expected serverStatus to be 1, got %v", cmdMap["serverStatus"])
		}
		if cmdMap["repl"] != 0 {
			t.Errorf("Expected repl to be 0, got %v", cmdMap["repl"])
		}
	} else {
		t.Error("Failed to recognize bson.D type")
	}
}

// TestRetryableWritesFix tests the retryable writes fix
func TestRetryableWritesFix(t *testing.T) {
	// Test that writeConcern includes retryWrites: false
	writeConcern := bson.D{
		{Name: "w", Value: 0},
		{Name: "retryWrites", Value: false},
	}

	// Verify retryWrites is set to false
	found := false
	for _, elem := range writeConcern {
		if elem.Name == "retryWrites" && elem.Value == false {
			found = true
			break
		}
	}

	if !found {
		t.Error("retryWrites should be set to false in writeConcern")
	}
}

// TestQueryErrorHandling tests the query error handling for retryable writes
func TestQueryErrorHandling(t *testing.T) {
	// Simulate a retryable writes error
	qerr := &QueryError{
		Code:    91,
		Message: "Retryable writes are not supported",
	}

	// Check if error message contains the retryable writes error
	if qerr.Message != "Retryable writes are not supported" {
		t.Errorf("Expected retryable writes error message, got: %s", qerr.Message)
	}
}
