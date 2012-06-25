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

// Package jfup provides backend support for the jQuery File Upload Plugin.
package jfu

import (
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

const (
	IMAGE_TYPES          = "image/(gif|p?jpeg|(x-)?png)"
)

var (
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

func (c *Config) imageRegex() *regexp.Regexp {
	return regexp.MustCompile(IMAGE_TYPES)
}

func (c *Config) acceptRegex() *regexp.Regexp {
	return regexp.MustCompile(c.AcceptFileTypes)
}

// UploadHandler provides a functions to handle file upload and serve 
// thumbnails.
type UploadHandler interface {
	HandleUpload(http.ResponseWriter, *http.Request)
	ServeThumbnail(http.ResponseWriter, *http.Request)
}

// FileInfo describes a file that has been uploaded.
type FileInfo struct {
	// Key          appengine.BlobKey `json:"-"`
	Key          string `json:"-"`
	Url          string `json:"url,omitempty"`
	ThumbnailUrl string `json:"thumbnail_url,omitempty"`
	Name         string `json:"name"`
	Type         string `json:"type"`
	Size         int64  `json:"size"`
	Error        string `json:"error,omitempty"`
	DeleteUrl    string `json:"delete_url,omitempty"`
	DeleteType   string `json:"delete_type,omitempty"`
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
