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
	"bytes"
	"encoding/base64"
	"github.com/bradfitz/gomemcache/memcache"
	"github.com/jmcvetta/jqupload/resize"
	"image"
	"image/png"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
)

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

func (fi *FileInfo) ValidateType() (valid bool) {
	if acceptFileTypes.MatchString(fi.Type) {
		return true
	}
	fi.Error = "acceptFileTypes"
	return false
}

func (fi *FileInfo) ValidateSize() (valid bool) {
	if fi.Size < MIN_FILE_SIZE {
		fi.Error = "minFileSize"
	} else if fi.Size > MAX_FILE_SIZE {
		fi.Error = "maxFileSize"
	} else {
		return true
	}
	return false
}

func (fi *FileInfo) CreateUrls(r *http.Request, c appengine.Context) {
	u := &url.URL{
		Scheme: r.URL.Scheme,
		Host:   appengine.DefaultVersionHostname(c),
		Path:   "/",
	}
	uString := u.String()
	fi.Url = uString + escape(string(fi.Key)) + "/" +
		escape(string(fi.Name))
	fi.DeleteUrl = fi.Url
	fi.DeleteType = "DELETE"
	if fi.ThumbnailUrl != "" && -1 == strings.Index(
		r.Header.Get("Accept"),
		"application/json",
	) {
		fi.ThumbnailUrl = uString + "thumbnails/" +
			escape(string(fi.Key))
	}
}

func (fi *FileInfo) CreateThumbnail(r io.Reader, c appengine.Context) (data []byte, err error) {
	defer func() {
		if rec := recover(); rec != nil {
			log.Println(rec)
			// 1x1 pixel transparent GIf, bas64 encoded:
			s := "R0lGODlhAQABAIAAAP///////yH5BAEKAAEALAAAAAABAAEAAAICTAEAOw=="
			data, _ = base64.StdEncoding.DecodeString(s)
			fi.ThumbnailUrl = "data:image/gif;base64," + s
		}
		memcache.Add(c, &memcache.Item{
			Key:        string(fi.Key),
			Value:      data,
			Expiration: EXPIRATION_TIME,
		})
	}()
	img, _, err := image.Decode(r)
	check(err)
	if bounds := img.Bounds(); bounds.Dx() > THUMBNAIL_MAX_WIDTH ||
		bounds.Dy() > THUMBNAIL_MAX_HEIGHT {
		w, h := THUMBNAIL_MAX_WIDTH, THUMBNAIL_MAX_HEIGHT
		if bounds.Dx() > bounds.Dy() {
			h = bounds.Dy() * h / bounds.Dx()
		} else {
			w = bounds.Dx() * w / bounds.Dy()
		}
		img = resize.Resize(img, img.Bounds(), w, h)
	}
	var b bytes.Buffer
	err = png.Encode(&b, img)
	check(err)
	data = b.Bytes()
	fi.ThumbnailUrl = "data:image/png;base64," +
		base64.StdEncoding.EncodeToString(data)
	return
}

