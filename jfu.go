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

// Package jfu provides backend support for the jQuery File Upload plugin.
package jfu

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bmizerany/mc"
	"github.com/jmcvetta/jfu/resize"
	"image"
	"image/png"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	// Register image handling libraries by importing them.
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

const (
	IMAGE_TYPES = "image/(gif|p?jpeg|(x-)?png)"
)

var (
	DefaultConfig = Config{
		MinFileSize:        1,       // bytes
		MaxFileSize:        5000000, // bytes
		AcceptFileTypes:    IMAGE_TYPES,
		ExpirationTime:     300,
		ThumbnailMaxWidth:  80,
		ThumbnailMaxHeight: 80,
	}
	imageRegex        = regexp.MustCompile(IMAGE_TYPES)
	FileNotFoundError = errors.New("File Not Found")
)

// Config is used to configure an UploadHandler.
type Config struct {
	MinFileSize        int    // bytes
	MaxFileSize        int    // bytes
	AcceptFileTypes    string // TODO: Change this to a Regexp object
	ExpirationTime     int    // seconds
	ThumbnailMaxWidth  int    // pixels
	ThumbnailMaxHeight int    // pixels
}

type DataStore interface {
	Create(*FileInfo, io.Reader) error           // Create a new file and return its key
	Get(key string) (FileInfo, io.Reader, error) // Get the file blob identified by key
	Delete(key string) error                     // Delete the file identified by key
}

// UploadHandler provides a functions to handle file upload and serve 
// thumbnails.
type UploadHandler struct {
	Prefix string     // URL prefix to serve
	Conf   *Config    // Configuration
	Store  *DataStore // Persistant storage for files
	Cache *mc.Conn // Cache for image thumbnails
}

// FileInfo describes a file that has been uploaded.
type FileInfo struct {
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
func http500(w http.ResponseWriter, err error) {
	if err != nil {
		msg := "500 Internal Server Error: " + err.Error()
		http.Error(w, msg, http.StatusInternalServerError)
		log.Panic(err)
	}
}

func escape(s string) string {
	return strings.Replace(url.QueryEscape(s), "+", "%20", -1)
}

func (h *UploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Remove h.Prefix from the URL path - based on http.StripPrefix()
	if !strings.HasPrefix(r.URL.Path, h.Prefix) {
		http.NotFound(w, r)
		return
	}
	r.URL.Path = r.URL.Path[len(h.Prefix):]
	params, err := url.ParseQuery(r.URL.RawQuery)
	http500(w, err)
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add(
		"Access-Control-Allow-Methods",
		// "OPTIONS, HEAD, GET, POST, PUT, DELETE",
		"GET, POST, PUT, DELETE",
	)
	switch r.Method {
	// case "OPTIONS":
	// case "HEAD":
	case "GET":
		h.get(w, r)
	case "POST":
		if len(params["_method"]) > 0 && params["_method"][0] == "DELETE" {
			h.delete(w, r)
			// http.Error(w, "POST-delete support not yet implmented", http.StatusNotImplemented)
		} else {
			h.post(w, r)
		}
	case "DELETE":
		h.delete(w, r)
	default:
		http.Error(w, "501 Not Implemented", http.StatusNotImplemented)
	}
}

func (h *UploadHandler) get(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		// Noop - not sure why JFU plugin sometimes calls this with no args
		return
	}
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) >= 2 && parts[1] == "thumbnails" {
		h.serveThumbnails(w, r)
		return
	}
	//
	//
	if len(parts) != 3 || parts[1] == "" {
		log.Printf("Invalid URL: '%v'", r.URL)
		http.Error(w, "404 Invalid URL", http.StatusNotFound)
		return
	}
	key := parts[1]
	log.Println("Get", key)
	fi, file, err := (*h.Store).Get(parts[1])
	if err == FileNotFoundError {
		log.Println("Not Found", key)
		http.NotFound(w, r)
		return
	}
	http500(w, err)
	log.Println("Found", key, fi.Type)
	//
	//
	w.Header().Add(
		"Cache-Control",
		fmt.Sprintf("public,max-age=%d", h.Conf.ExpirationTime),
	)
	if imageRegex.MatchString(fi.Type) {
		w.Header().Add("X-Content-Type-Options", "nosniff")
	} else {
		w.Header().Add("Content-Type", "application/octet-stream")
		w.Header().Add(
			"Content-Disposition:",
			fmt.Sprintf("attachment; filename=%s;", parts[2]),
		)
	}
	io.Copy(w, file)
	return
}

// uploadFile handles the upload of a single file from a multipart form.  
func (h *UploadHandler) uploadFile(w http.ResponseWriter, p *multipart.Part) (fi *FileInfo) {
	fi = &FileInfo{
		Name: p.FileName(),
		Type: p.Header.Get("Content-Type"),
	}
	//
	// Validate file type
	//
	re := regexp.MustCompile(h.Conf.AcceptFileTypes) // It's inefficient to recompile the regex every time.  
	if re.MatchString(fi.Type) == false {
		fi.Error = "acceptFileTypes"
		return
	}
	isImage := imageRegex.MatchString(fi.Type)
	//
	// Copy into buffers for save and thumbnail generation
	//
	// Max + 1 for LimitedReader size, so we can detect below if file size is 
	// greater than max.
	lr := &io.LimitedReader{R: p, N: int64(h.Conf.MaxFileSize + 1)}
	var bSave bytes.Buffer  // Buffer to be saved
	var bThumb bytes.Buffer // Buffer to be thumbnailed
	var wr io.Writer
	if isImage {
		wr = io.MultiWriter(&bSave, &bThumb)
	} else {
		wr = &bSave
	}
	_, err := io.Copy(wr, lr)
	http500(w, err)
	//
	// Validate file size
	// 
	size := bSave.Len()
	if size < h.Conf.MinFileSize {
		log.Println("File failed validation: too small.", size, h.Conf.MinFileSize)
		fi.Error = "minFileSize"
		return
	} else if size > h.Conf.MaxFileSize {
		log.Println("File failed validation: too large.", size, h.Conf.MaxFileSize)
		fi.Error = "maxFileSize"
		return
	}
	//
	// Save to data store
	//
	err = (*h.Store).Create(fi, &bSave)
	http500(w, err)
	log.Println("Create", size)
	//
	// Set URLs in FileInfo
	//
	u := &url.URL{
		Path: h.Prefix + "/",
	}
	uString := u.String()
	fi.Url = uString + escape(string(fi.Key)) + "/" +
		escape(string(fi.Name))
	fi.DeleteUrl = fi.Url
	fi.DeleteType = "DELETE"
	fi.ThumbnailUrl = uString + "thumbnails/" + escape(string(fi.Key))
	//
	// Create thumbnail
	//
	if isImage && size > 0 {
		_, err = h.CreateThumbnail(fi, &bThumb)
		if err != nil {
			log.Println("Error creating thumbnail:", err)
		}
		// If we wanted to save thumbnails to peristent storage, this would be 
		// a good spot to do it.
	}
	return
}

func getFormValue(p *multipart.Part) string {
	var b bytes.Buffer
	io.CopyN(&b, p, int64(1<<20)) // Copy max: 1 MiB
	return b.String()
}

func (h *UploadHandler) post(w http.ResponseWriter, r *http.Request) {
	//
	// We may potentially handle multiple  uploads
	//
	fileInfos := make([]*FileInfo, 0)
	mr, err := r.MultipartReader()
	http500(w, err)
	r.Form, err = url.ParseQuery(r.URL.RawQuery)
	http500(w, err)
	for {
		// var part *multipart.Part
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		http500(w, err)
		if name := part.FormName(); name != "" {
			if part.FileName() != "" {
				fileInfos = append(fileInfos, h.uploadFile(w, part))
			} else {
				r.Form[name] = append(r.Form[name], getFormValue(part))
			}
		}
	}
	js, err := json.Marshal(fileInfos)
	http500(w, err)
	//
	//
	//
	if redirect := r.FormValue("redirect"); redirect != "" {
		http.Redirect(w, r, fmt.Sprintf(
			redirect,
			escape(string(js)),
		), http.StatusFound)
		return
	}
	jsonType := "application/json"
	if strings.Index(r.Header.Get("Accept"), jsonType) != -1 {
		w.Header().Set("Content-Type", jsonType)
	}
	fmt.Fprintln(w, string(js))
}

func (h *UploadHandler) delete(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) != 3 || parts[1] == "" {
		log.Println("Invalid URL:", r.URL)
		http.Error(w, "404 Invalid URL", http.StatusNotFound)
		return
	}
	key := parts[1]
	log.Println("Delete", key)
	err := (*h.Store).Delete(key)
	http500(w, err)
	return
}

func (h *UploadHandler) serveThumbnails(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) == 3 {
		if key := parts[2]; key != "" {
			var data []byte
			val, _, _, err := h.Cache.Get(key)
			if err == nil {
				data = []byte(val)
			} else {
				if fi, r, err := (*h.Store).Get(key); err == nil {
					_, err = h.CreateThumbnail(&fi, r)
				}
			}
			if err == nil && len(data) > 3 {
				w.Header().Add(
					"Cache-Control",
					fmt.Sprintf("public,max-age=%d", h.Conf.ExpirationTime),
				)
				contentType := "image/png"
				if string(data[:3]) == "GIF" {
					contentType = "image/gif"
				} else if string(data[1:4]) != "PNG" {
					contentType = "image/jpeg"
				}
				w.Header().Set("Content-Type", contentType)
				fmt.Fprintln(w, string(data))
				return
			}
		}
	}
	http.Error(w, "404 Not Found", http.StatusNotFound)
}

// CreateThumbnail generates a thumbnail and adds it to the cache.
func (h *UploadHandler) CreateThumbnail(fi *FileInfo, r io.Reader) (data []byte, err error) {
	defer func() {
		if rec := recover(); rec != nil {
			log.Println(rec)
			// 1x1 pixel transparent GIf, bas64 encoded:
			s := "R0lGODlhAQABAIAAAP///////yH5BAEKAAEALAAAAAABAAEAAAICTAEAOw=="
			data, _ = base64.StdEncoding.DecodeString(s)
			fi.ThumbnailUrl = "data:image/gif;base64," + s
		}
		h.Cache.Set(fi.Key, string(data), 0, 0, h.Conf.ExpirationTime)
	}()
	img, _, err := image.Decode(r)
	if err != nil {
		panic(err)
	}
	if bounds := img.Bounds(); bounds.Dx() > h.Conf.ThumbnailMaxWidth ||
		bounds.Dy() > h.Conf.ThumbnailMaxHeight {
		w, h := h.Conf.ThumbnailMaxWidth, h.Conf.ThumbnailMaxHeight
		if bounds.Dx() > bounds.Dy() {
			h = bounds.Dy() * h / bounds.Dx()
		} else {
			w = bounds.Dx() * w / bounds.Dy()
		}
		img = resize.Resize(img, img.Bounds(), w, h)
	}
	var b bytes.Buffer
	err = png.Encode(&b, img)
	if err != nil {
		panic(err)
	}
	data = b.Bytes()
	fi.ThumbnailUrl = "data:image/png;base64," +
		base64.StdEncoding.EncodeToString(data)
	return
}
