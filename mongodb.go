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
	"io"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

type mongoStore struct {
	col *mgo.Collection // Collection where file info will be stored
	fs  *mgo.GridFS     // GridFS where file blob will be stored
}

// Exists checks whether a blob identified by key exists in the data store
func (s mongoStore) Exists(key string) (result bool, err error) {
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

// ContentType returns the content type for the saved file indicated by key, 
// or returns a FileNotFoundError if no such file exists.
func (s mongoStore) ContentType(key string) (ct string, err error) {
	var file mgo.GridFile
	q := s.fs.Find(bson.M{"_id": key})
	err = q.One(&file)
	switch {
	case err == mgo.ErrNotFound:
		return ct, FileNotFoundError
	case err != nil:
		return
	}
	return file.ContentType(), nil
}

func (s mongoStore) Create(fi *FileInfo, r io.Reader) error {
	file, err := s.fs.Create(fi.Name)
	if err != nil {
		return err
	}
	defer file.Close()
	file.SetContentType(fi.Type)
	_, err = io.Copy(file, r)
	if err != nil {
		return err
	}
	fi.Key = file.Id().(bson.ObjectId).Hex()
	fi.Size = file.Size()
	return nil
}

func (s mongoStore) Get(key string) (fi FileInfo, r io.Reader, err error) {
	// blobKey := appengine.BlobKey(key)
	// bi, err := blobstore.Stat(appengine.NewContext(r), blobKey)
	// blobstore.Send(w, appengine.BlobKey(key))
	id := bson.ObjectIdHex(key)
	file, err := s.fs.OpenId(id)
	if err != nil {
		fi = FileInfo{
			Error: err.Error(),
		}
	} else {
		fi = FileInfo{
			Key: key,
			// Url
			// ThumbnailUrl
			Name: file.Name(),
			Type: file.ContentType(),
			Size: file.Size(),
			// Error
			// DeleteUrl
			// DeleteType
		}
	}
	r = file
	return
}

func NewMongoStore(gfs *mgo.GridFS) DataStore {
	ms := mongoStore{fs: gfs}
	return ms
}
