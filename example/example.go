package main

import (
	// "github.com/bradfitz/gomemcache/memcache"
	"github.com/bmizerany/mc"
	"github.com/jmcvetta/jfu"
	"labix.org/v2/mgo"
	"log"
	"net/http"
	"os"
)

func main() {
	log.SetFlags(log.Ltime | log.Lshortfile)
	//
	path := os.Getenv("PWD")
	port := os.Getenv("PORT")
	if port == "" {
		port = "5000"
	}
	url := "0.0.0.0:" + port
	//
	conn, err := mgo.Dial("localhost")
	if err != nil {
		log.Panic(err)
	}
	db := conn.DB("test_foobar")
	gfs := db.GridFS("test_foobar")
	store := jfu.NewMongoStore(gfs)
	//
	//client := memcache.New("localhost")
	client, err := mc.Dial("tcp", "localhost:11211")
	if err != nil {
		log.Panic(err)
	}
	//
	uh := jfu.UploadHandler{
		Prefix: "/jfu",
		Store:  &store,
		Conf:   &jfu.DefaultConfig,
		Cache:  client,
	}
	//
	http.Handle("/jfu", &uh)
	http.Handle("/jfu/", &uh)
	http.Handle("/", http.FileServer(http.Dir(path)))
	log.Println("Starting webserver on " + url + "...")
	http.ListenAndServe(url, nil)
}
