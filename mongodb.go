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
	"io"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

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

func (s *mongoStore) Create(fi FileInfo, r io.Reader) (FileInfo, error) {
	file, err := s.fs.Create(fi.Name)
	if err != nil {
		return fi, err
	}
	defer file.Close()
	file.SetContentType(fi.Type)
	_, err = io.Copy(file, r)
	if err != nil {
		return fi, err
	}
	fi.Key = file.Id().(bson.ObjectId).Hex()
	return fi, nil
}

func (s *mongoStore) Get(key string) (r io.Reader, err error) {
	// blobKey := appengine.BlobKey(key)
	// bi, err := blobstore.Stat(appengine.NewContext(r), blobKey)
	// blobstore.Send(w, appengine.BlobKey(key))
	id := bson.ObjectIdHex(key)
	file, err := s.fs.OpenId(id)
	file.Name()
	return file, nil
}
