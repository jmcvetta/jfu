/*
 * Copyright (c) 2012 Jason McVetta.  This is Free Software, released under the
 * terms of the AGPL v3.  See www.gnu.org/licenses/agpl-3.0.html for details.
 */
 
package jfu

import (
	"bytes"
	"github.com/jmcvetta/randutil"
	"labix.org/v2/mgo"
	// "labix.org/v2/mgo/bson"
	// "log"
	"testing"
)

// Initialize a randomly-named testing database
func InitMongo(t *testing.T) *mgo.Database {
	randStr, err := randutil.AlphaString(64)
	if err != nil { t.Fatal(err) }
	dbName := "testing_" + randStr
	conn, err := mgo.Dial("localhost")
	if err != nil { t.Fatal(err) }
	return conn.DB(dbName)
}

func TestRoundTrip(t *testing.T) {
	//
	// Initialize mongoStore instance
	//
	db := InitMongo(t)
	defer db.DropDatabase()
	data, err := randutil.AlphaString(512)
	if err != nil { t.Fatal(err) }
	ms := mongoStore{fs: db.GridFS("")}
	//
	// Create
	//
	b := bytes.NewBufferString(data)
	fi := FileInfo{
		Type: "text/plain",
	}
	err = ms.Create(fi, b)
	if err != nil {
		t.Fatal(err)
	}
	key = fi.Key
	//
	// Get
	//
	fi, r, err := ms.Get(key)
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
