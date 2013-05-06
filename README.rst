Meet Khargo
===========


Khargo is an http file server with a GridFS (MongoDB) backend.  If you want to
store files in GridFS and serve them over http, Khargo can help you out.

Running
=======

Launch Khargo from the command line::

    $ ./khargo 
    2013/05/06 16:45:41 Connecting to mongodb://localhost:27017/test with strong consistency.
    2013/05/06 16:45:41 Listening on :8000

Khargo accepts four command line arguments

- -consistency.  Default: "strong".  mgo driver consistency mode.  One of eventual, monotonic, or strong. See http://godoc.org/labix.org/v2/mgo#Session.SetMode.

- -dburl.  Default: mongodb://localhost:27017/test. See http://godoc.org/labix.org/v2/mgo#Dial for other values.

- -max-age.  Default: 31557600 (one year).  Lifetime (in seconds) for setting Cache-Control and Expires headers.

- -port.  Default: 8000.  Port to listen on.

Khargo does not daemonize.

The Khargo repository includes a Procfile that will launch Khargo as a 'web'
proc.  It expects $PORT and $DBURL environment variables to be set.

Compiling
=========

Khargo is written in go.  The only requirement outside the Go standard library
is the mgo driver::

    go get labix.org/v2/mgo
    go build khargo.go

Logging
=======

Khargo emits logs to stdout.  Log lines look like this::

    2013/05/06 15:05:03 10.0.2.2 - GET - /myfile.txt - curl/7.21.4 (universal-apple-darwin11.0) libcurl/7.21.4 OpenSSL/0.9.8r zlib/1.2.5 -  - 200 - 14.024ms

That format is::

    date time ip - method - path - user agent - referer - status - response time

Uploading Files
===============

The 'mongo' command line client is usually also installed with a mongofiles_
utility that you can use to upload files to GridFS.  Here's an example of using
it to upload this README file, from one directory up::

    mongofiles put README.rst --local khargo/README.rst -h mymongoserver --db myfilesdb

The Name
========

I previously created a Python program called Khartoum_ that could serve GridFS
files over http.  Khargo is a rewrite in Go.

.. _mongofiles: http://docs.mongodb.org/manual/reference/mongofiles/
.. _Khartoum: https://bitbucket.org/btubbs/khartoum
