package jfu

import (
	"bytes"
	"github.com/jmcvetta/randutil"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"log"
	"testing"
)

// Initialize a randomly-named testing database
func InitMongo(t *testing.T) *mgo.Database {
	dbName := "testing_" + randutil.AlphaString(64)
	conn, err := mgo.Dial("localhost")
	if err != nil {
		t.Fatal(err)
	}
	return conn.DB(dbName)
}

func TestRoundTrip(t *testing.T) {
	//
	// Initialize mongoStore instance
	//
	db := InitMongo(t)
	defer db.DropDatabase()
	data = randutil.AlphaString(512)
	ms := mongoStore{fs: db.GridFS("")}
	//
	// Create
	//
	b := bytes.NewBufferString(data)
	key, err := ms.Create("text/plain", b)
	if err != nil {
		t.Fatal(err)
	}
	//
	// Exists
	//
	e, err := ms.Exists(key)
	if err != nil {
		t.Error(err)
	}
	if !e {
		msg := "Exists() returned false on the key for a file we just created."
		t.Error(msg)
	}
	//
	// Get
	//
	r, err := ms.Get(key)
	if err != nil {
		t.Error(err)
	}
	retr := string(r)
	if retr != data {
		msg := "Different data retrieved than was saved!\n"
		msg += "data: " + data + "\n"
		msg += "retr: " + retr + "\n"
		t.Error(msg)
	}
}
