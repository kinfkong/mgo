package ikatago_tests

import (
	"testing"
	"time"

	"github.com/globalsign/mgo/bson"
)

const (
	IKATAGO_CLUSTER_CHECKPOINT_JOB_COLLECTION = "checkpoint_job"
	DBNAME_TEST_CHECKPOINT                    = "ikatago_test"
)

type CheckpointJob struct {
	ID        bson.ObjectId `bson:"_id,omitempty"`
	Finished  bool          `bson:"finished"`
	TryCount  int           `bson:"tryCount"`
	UpdatedAt time.Time     `bson:"updatedAt"`
}

func TestCheckpointLimitAll(t *testing.T) {
	session := getSession(t)
	defer session.Close()

	c := session.DB(DBNAME_TEST_CHECKPOINT).C(IKATAGO_CLUSTER_CHECKPOINT_JOB_COLLECTION)
	defer c.DropCollection()

	for i := 0; i < 200; i++ {
		c.Insert(&CheckpointJob{Finished: false})
	}
	for i := 0; i < 50; i++ {
		c.Insert(&CheckpointJob{Finished: true})
	}

	var jobs []CheckpointJob
	err := c.Find(bson.M{"finished": false}).Limit(100).All(&jobs)
	if err != nil {
		t.Fatalf("Find().Limit().All() failed: %v", err)
	}

	if len(jobs) != 100 {
		t.Errorf("Expected 100 jobs, got %d", len(jobs))
	}
}

func TestCheckpointUpdateIdWithInc(t *testing.T) {
	session := getSession(t)
	defer session.Close()

	c := session.DB(DBNAME_TEST_CHECKPOINT).C(IKATAGO_CLUSTER_CHECKPOINT_JOB_COLLECTION)
	defer c.DropCollection()

	job := &CheckpointJob{ID: bson.NewObjectId(), TryCount: 0}
	c.Insert(job)

	now := time.Now().Truncate(time.Millisecond)
	err := c.UpdateId(job.ID, bson.M{
		"$set": bson.M{
			"updatedAt": now,
		},
		"$inc": bson.M{
			"tryCount": 1,
		},
	})
	if err != nil {
		t.Fatalf("UpdateId with $inc failed: %v", err)
	}

	var updatedJob CheckpointJob
	c.FindId(job.ID).One(&updatedJob)

	if updatedJob.TryCount != 1 {
		t.Errorf("Expected tryCount to be 1, got %d", updatedJob.TryCount)
	}
	if !updatedJob.UpdatedAt.Equal(now) {
		t.Errorf("Expected updatedAt to be %v, got %v", now, updatedJob.UpdatedAt)
	}
}
