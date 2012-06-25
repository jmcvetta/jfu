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
	"labix.org/v2/mgo"
	// "labix.org/v2/mgo/bson"
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

func (s *mongoStore) Exists(key string) bool {
	// blobKey := appengine.BlobKey(key)
	// bi, err := blobstore.Stat(appengine.NewContext(r), blobKey)
	s.fs.Find()
}
