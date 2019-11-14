package main

import (
	"compress/gzip"
	"flag"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"labix.org/v2/mgo"
)

func main() {
	dburl := flag.String("dburl", "mongodb://localhost:27017/test",
		"MongoDB URL. See http://godoc.org/labix.org/v2/mgo#Dial")

	port := flag.String("port", "8000", "Port to listen on.")

	mode := flag.String("consistency", "strong", "mgo driver consistency "+
		"mode.  One of eventual, monotonic, or strong. "+
		"See http://godoc.org/labix.org/v2/mgo#Session.SetMode.")

	corsHeader := flag.String("allow-origin", "*",
		"value for Access-Control-Allow-Origin header")

	maxAge := flag.Int("max-age", 31557600, "Lifetime (in seconds) for "+
		"setting Cache-Control and Expires headers.  Defaults to one year.")

	flag.Parse()
	log.Printf("Connecting to %s with %s consistency.\n", *dburl, *mode)
	session, err := mgo.Dial(*dburl)
	check(err)
	defer session.Close()

	// It would be nice to do this with a map or a function, but mgo.mode is
	// not a type we can use :\
	switch {
	case *mode == "eventual":
		session.SetMode(mgo.Eventual, true)
	case *mode == "monotonic":
		session.SetMode(mgo.Monotonic, true)
	case *mode == "strong":
		session.SetMode(mgo.Strong, true)
	case true:
		panic(fmt.Sprintf("Invalid consistency mode %s.  Must be eventual, "+
			"monotonic, or strong.", *mode))
	}

	db := session.DB("") // will use DB specified in dburl
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		handler(w, r, db, *maxAge, *corsHeader)
	})

	log.Printf("Listening on :%s\n", *port)
	http.ListenAndServe(fmt.Sprintf(":%s", *port), nil)
}

func handler(w http.ResponseWriter, r *http.Request, db *mgo.Database, maxAge int, corsHeader string) {
	start := time.Now()
	var status = new(int)
	*status = 200
	defer logRequest(r, status, start)

	if r.Method != "GET" && r.Method != "HEAD" {
		*status = http.StatusMethodNotAllowed
		w.WriteHeader(*status)
		fmt.Fprintf(w, "%s Method Not Allowed\n", r.Method)
		return
	}

	filename := r.URL.Path[1:]
	file, err := db.GridFS("fs").Open(filename)

	// Only return a 404 if the error from gridfs was 'not found'.  If
	// something else goes wrong, return 500.
	if err != nil {
		if err == mgo.ErrNotFound {
			*status = http.StatusNotFound
			w.WriteHeader(*status)
			fmt.Fprintf(w, "%s Not Found\n", filename)
			return
		}
		fmt.Printf("[%s]: %v\n", filename, err)
		*status = http.StatusInternalServerError
		w.WriteHeader(*status)
		fmt.Fprintf(w, "Internal Server Error\n")
		return
	}
	defer file.Close()

	// Set CORS header
	w.Header().Set("Access-Control-Allow-Origin", corsHeader)

	// Set expiry headers
	w.Header().Set("Cache-Control", fmt.Sprintf("max-age=%d", maxAge))
	expiration := time.Now().Add(time.Duration(maxAge) * time.Second)
	w.Header().Set("Expires", expiration.Format(time.RFC1123))

	ctype := getMimeType(file)
	w.Header().Set("Content-Type", ctype)

	// Compress text mimetypes, but don't bother with already-compressed things
	// like jpg, png, tar.gz, etc.
	if shouldCompress(r.Header.Get("Accept-Encoding"), ctype) {
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		defer gz.Close()
		if r.Method == "GET" {
			io.Copy(gz, file)
		}
	} else {
		w.Header().Set("Content-Length", strconv.FormatInt(file.Size(), 10))
		w.Header().Set("ETag", file.MD5())
		if r.Method == "GET" {
			io.Copy(w, file)
		}
	}
}

func logRequest(r *http.Request, status *int, start time.Time) {
	logfields := []string{
		strings.Split(r.RemoteAddr, ":")[0],
		r.Method,
		r.URL.Path,
		r.UserAgent(),
		r.Referer(),
		strconv.Itoa(*status),
		time.Since(start).String(),
	}
	log.Printf(strings.Join(logfields, " - "))
}

func check(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func getMimeType(file *mgo.GridFile) string {
	// Honor the gridfs content type if set.  Otherwise guess by extension.
	ctype := file.ContentType()
	if ctype == "" {
		ctype := mime.TypeByExtension(filepath.Ext(file.Name()))
		if ctype == "" {
			return "application/octet-stream"
		}
		return ctype
	}
	return ctype
}

// As of May 2013 Go doesn't have a real Set type, so fake it with a map where
// we ignore the values.
var gzippableTypes = map[string]bool{
	"text/plain":                true,
	"text/plain; charset=utf-8": true,
	"text/html":                 true,
	"application/javascript":    true,
	"text/css":                  true,
	"text/css; charset=utf-8":   true,
	"application/json":          true,
	"application/xml":           true,
}

func shouldCompress(encodingHeader string, mimetype string) bool {
	_, goodtype := gzippableTypes[mimetype]
	return strings.Contains(encodingHeader, "gzip") && goodtype
}
