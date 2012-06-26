/*
 * (c) 2012 Jason McVetta.  This is Free Software, released under the terms of
 * the GPL v3.  See http://www.gnu.org/copyleft/gpl.html for details.
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
	"github.com/bradfitz/gomemcache/memcache"
	"io"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

// An implementation of UploadHandler based on MongoDB
type mongoHandler struct {
	conf  Config           // UploadHandler configuration
	cache *memcache.Client // Memcache client (optional)
}

type mongoStore struct {
	col *mgo.Collection // Collection where file info will be stored
	fs  *mgo.GridFS     // GridFS where file blob will be stored
}

// Exists checks whether a blob identified by key exists in the data store
func (s *mongoStore) Exists(key string) (result bool, err error) {
	// blobKey := appengine.BlobKey(key)
	// bi, err := blobstore.Stat(appengine.NewContext(r), blobKey)
	q := s.fs.Find(bson.M{"_id": key})
	cnt, err := q.Count()
	if err != nil {
		return
	}
	if cnt > 0 {
		result = true
	} else {
		result = false
	}
	return
}

// Create stores a new file from  buffer r to disk and returns its key
func (s *mongoStore) Create(filetype string, data io.Reader) (key string, err error) {
	file, err := s.fs.Create("")
	if err != nil {
		return
	}
	defer file.Close()
	_, err = io.Copy(file, data)
	if err != nil {
		return
	}
	key = file.Id().(bson.ObjectId).Hex()
	file.SetContentType(filetype)
	return
}
