/*
 * jQuery File Upload Plugin GAE Go Example 2.0
 * https://github.com/blueimp/jQuery-File-Upload
 *
 * Copyright 2011, Sebastian Tschan
 * https://blueimp.net
 *
 * Licensed under the MIT license:
 * http://www.opensource.org/licenses/MIT
 */

package jfup

import (
	//	"appengine"
	//	"appengine/blobstore"
	//	"appengine/memcache"
	//	"appengine/taskqueue"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

const (
	WEBSITE              = "http://blueimp.github.com/jQuery-File-Upload/"
	MIN_FILE_SIZE        = 1       // bytes
	MAX_FILE_SIZE        = 5000000 // bytes
	IMAGE_TYPES          = "image/(gif|p?jpeg|(x-)?png)"
	ACCEPT_FILE_TYPES    = IMAGE_TYPES
	EXPIRATION_TIME      = 300 // seconds
	THUMBNAIL_MAX_WIDTH  = 80
	THUMBNAIL_MAX_HEIGHT = THUMBNAIL_MAX_WIDTH
)

var (
	imageTypes      = regexp.MustCompile(IMAGE_TYPES)
	acceptFileTypes = regexp.MustCompile(ACCEPT_FILE_TYPES)
	defaultConfig   = Config{
		MinFileSize:        1,
		MaxFileSize:        2,
		AcceptFileTypes:    IMAGE_TYPES,
		ExpirationTime:     300,
		ThumbnailMaxWidth:  80,
		ThumbnailMaxHeight: 80,
	}
)

// Config is used to configure an UploadHandler.
type Config struct {
	MinFileSize        int    // bytes
	MaxFileSize        int    // bytes
	AcceptFileTypes    string // regular expression
	ExpirationTime     int    // seconds
	ThumbnailMaxWidth  int    // pixels
	ThumbnailMaxHeight int    // pixels
}

type cache interface {
	Add()
}

// UploadHandler provides a functions to handle file upload and serve 
// thumbnails.
type UploadHandler interface {
	HandleUpload(http.ResponseWriter, *http.Request)
	ServeThumbnail(http.ResponseWriter, *http.Request)
}

// http500 Raises an HTTP 500 Internal Server Error if err is non-nil
func http500(err error, w http.ResponseWriter) {
	if err != nil {
		msg := "500 Internal Server Error: " + err.Error()
		http.Error(w, msg, 500)
	}
}

func escape(s string) string {
	return strings.Replace(url.QueryEscape(s), "+", "%20", -1)
}
