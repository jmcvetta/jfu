/*
 * Copyright (c) 2012 Jason McVetta.  This is Free Software, released under the
 * terms of the WTFPL v2.  It comes without any warranty, express or implied.
 * See http://sam.zoy.org/wtfpl/COPYING for more details.
 * 
 *
 * Derived from: 
 *
 * jQuery File Upload Plugin GAE Go Example 2.0
 * https://github.com/blueimp/jQuery-File-Upload
 *
 * Copyright 2011, Sebastian Tschan
 * https://blueimp.net
 *
 * Original software by Tschan licensed under the MIT license:
 * http://www.opensource.org/licenses/MIT
 */

package jfu

import (
	"bytes"
	"github.com/jmcvetta/randutil"
	"labix.org/v2/mgo"
	// "labix.org/v2/mgo/bson"
	// "log"
	"io"
	"testing"
)

// Initialize a randomly-named testing database
func InitMongo(t *testing.T) *mgo.Database {
	randStr, err := randutil.AlphaString(32)
	if err != nil {
		t.Fatal(err)
	}
	dbName := "testing_" + randStr
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
	data, err := randutil.AlphaString(512)
	if err != nil {
		t.Fatal(err)
	}
	ms := mongoStore{gfs: db.GridFS("")}
	//
	// Create
	//
	b := bytes.NewBufferString(data)
	fi := FileInfo{
		Type: "text/plain",
	}
	err = ms.Create(&fi, b)
	if err != nil {
		t.Fatal(err)
	}
	key := fi.Key
	t.Log("key:", key)
	//
	// Get
	//
	fi, r, err := ms.Get(key)
	if err != nil {
		t.Error(err)
	}
	rbuf := new(bytes.Buffer)
	io.Copy(rbuf, r)
	retr := rbuf.String()
	if retr != data {
		msg := "Different data retrieved than was saved!\n"
		msg += "data: " + data + "\n"
		msg += "retr: " + retr + "\n"
		t.Error(msg)
	}
	//
	// Delete
	//
	t.Log(key)
	err = ms.Delete(key)
	if err != nil {
		t.Fatal(err)
	}
}
