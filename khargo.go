package main

import (
    "fmt"
    "io"
    "labix.org/v2/mgo"
	"net/http"
)


func check(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func main() {
    session, err := mgo.Dial("localhost")
    check(err)
    defer session.Close()

    // Provide strong consistency by default.  Consider making this
    // configurable.
    session.SetMode(mgo.Strong, true)

    db := session.DB("test")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        // FIXME: better logging here.
        fmt.Printf(r.URL.Path)
        filename := r.URL.Path[1:]

        file, err := db.GridFS("fs").Open(filename)
        if err != nil {
            // FIXME: only return a 404 if the error from gridfs was 'not
            // found'.  If something else goes wrong, return 500.
            w.WriteHeader(http.StatusNotFound)
            fmt.Fprintf(w, "%s not found", filename)
            return
        }

        // FIXME: serve gzipped response if:
        // - the request headers indicate the client can handle it.
        // - the mimetype is one that's not already compressed (text/plain,
        // html, css, json, javascript)
        io.Copy(w, file)
    })
	http.ListenAndServe(":8000", nil)
}
