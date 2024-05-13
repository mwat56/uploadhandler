# UploadHandler

[![Golang](https://img.shields.io/badge/Language-Go-green.svg)](https://golang.org/)
[![GoDoc](https://godoc.org/github.com/mwat56/uploadhandler?status.svg)](https://godoc.org/github.com/mwat56/uploadhandler/)
[![Go Report](https://goreportcard.com/badge/github.com/mwat56/uploadhandler)](https://goreportcard.com/report/github.com/mwat56/uploadhandler)
[![Issues](https://img.shields.io/github/issues/mwat56/uploadhandler.svg)](https://github.com/mwat56/uploadhandler/issues?q=is%3Aopen+is%3Aissue)
[![Size](https://img.shields.io/github/repo-size/mwat56/uploadhandler.svg)](https://github.com/mwat56/uploadhandler/)
[![Tag](https://img.shields.io/github/tag/mwat56/uploadhandler.svg)](https://github.com/mwat56/uploadhandler/tags)
[![License](https://img.shields.io/github/license/mwat56/uploadhandler.svg)](https://github.com/mwat56/uploadhandler/blob/main/LICENSE)
[![View examples](https://img.shields.io/badge/learn%20by-examples-0077b3.svg)](https://github.com/mwat56/uploadhandler/blob/main/cmd/demo.go)

- [UploadHandler](#uploadhandler)
	- [Purpose](#purpose)
	- [Installation](#installation)
	- [Usage](#usage)
	- [Libraries](#libraries)
	- [Licence](#licence)

----

## Purpose

Sometimes a web-server application needs a way to accept file uploads from the remote users.
This middleware package does just this: it accepts uploads (up to a certain size) and stores them in a configurable directory.

## Installation

You can use `Go` to install this package for you:

	go get -u github.com/mwat56/uploadhandler

## Usage

The main function to call is

    // Wrap returns a handler function that includes upload handling,
    // wrapping the given `aHandler` and calling it internally.
    //
    func Wrap(aHandler http.Handler,
        aDestDir, aFieldName, aUpURL, aNextURL string,
        aMaxSize int64, aPager errorhandler.TErrorPager) http.Handler {…}

While at first glance the number of arguments seems to be overwhelming they allow you to fully configure the package's behaviour when an uploaded file arrives.
Let's look at the arguments one by one:

* `aHandler` is the handler function which you're already using for your web-server.
It will continue to work as it used to before except that a certain URL (configured by `aUpURL`, see below) will be intercepted if something gets POSTed to it.
* `aDestDir` is the directory where the incoming file is finally stored after processing it.
* `aFieldName` is the name/ID of the form/field your web-page uses to accept the file-upload.
* `aUpURL` is the URL your web-page's `FORM` element POSTs its data to.
This URL will be intercepted (if it's accessed by the POST HTTP method) and its data will be processed.
* `aNextURL` is the URL the user gets forwarded to after the file upload was successfully processed.
* `aMaxSize` defines the max. accepted size of uploaded files; if the given value is smaller/equal to zero then a maximal filesize of 8 MB is used.
Files bigger than that value will be rejected.
Think carefully about which size will suit your actual needs.
* `aPager` is an optional provider of customised error pages (or `nil` if not needed). –
See [github.com/mwat56/errorhandler](https://github.com/mwat56/errorhandler) for details about that package.

Here is a very [simple example](https://github.com/mwat56/uploadhandler/blob/main/cmd/demo.go) using this package:

    func testHandler(aWriter http.ResponseWriter, aRequest *http.Request) {
        // the upload form to show
        page := `<!DOCTYPE html><html><head><title>Go Upload</title></head><body>
        <form action="/up" method="post" enctype="multipart/form-data">
        <p><label for="uploadFile">Filename:</label>
        <input type="file" name="uploadFile" id="uploadFile"></p>
        <p><input type="submit" name="submit" value="Submit"></p>
        </form></body></html>`

        // send it to the remote user:
        aWriter.WriteHeader(200)
        aWriter.Write([]byte(page))
        // POST is handled by the UploadHandler
    } // testHandler()

    func main() {
        // let the upload handler wrap our own page handler:
        handler := uploadhandler.Wrap(http.HandlerFunc(testHandler),
            "./static", "uploadFile", "up", "/", 10*1024*1024, nil)

        if err := http.ListenAndServe("127.0.0.1:8080", handler); nil != err {
            log.Fatalf("%s Problem: %v", os.Args[0], err)
        }
    } // main()

You'll probably store the required values for e.g. `aDestDir` and `aMaxSize` in some kind of config-file, reading them at start of your web-server, and passing them along to the final `Wrap(…)` call instead of hard-coding them like in the example above.
And the values of `aUpURL` and `aFieldName` must, obviously, correspond with those you're actually using in your own application's forms.

If you don't use [customised error pages](https://github.com/mwat56/errorhandler) you can pass `nil` for the `aPager` argument as done in the example.

So, to add the file-upload functionality to your web-server application all that's needed is a single `Wrap()` function call. That's it.

However, if for some reason you'd like to be a little more "hands on", you can use another function to get a `TUploadHandler` instance:

    // NewHandler returns a new `tUploadHandler` instance.
    func NewHandler(aDestDir, aFieldName string,
        aMaxSize int64) *TUploadHandler {…}

This function call needs only a subset of the arguments passed to the `Wrap(…)` function:

* `aDestDir`: the directory where the incoming file is finally stored after processing it.
* `aFieldName`: the name/ID of the form/field your web-page uses to accept the file-upload.
* `aMaxSize`: the max. accepted size of uploaded files; if the given value is smaller/equal to zero then a maximal filesize of 8 MB is used.

The `NewHandler()` function's result provides the method

    // ServeUpload handles the incoming file upload.
    func (uh *TUploadHandler) ServeUpload(aWriter http.ResponseWriter,
        aRequest *http.Request) (string, int) {…}

This method does the actual upload handling.
It returns a string (holding a possible error message) and an integer (holding the HTTP status code).
If the returned status code is `200` (i.e. everything's alright) then the string return value will be the name of the processed file.
In all other cases (i.e. result `status != 200`) the calling application can react to the return values as it sees fit.

You can use several `TUploadHandler` instances to serve different URLs and different destination directories etc.
Insofar calling `NewHandler()` and then `ServeUpload(…)` gives you more flexibility then simply calling `Wrap(…)`.
On the other hand, you could call `Wrap(…)` several times, wrapping one instance within the other and thus react to different URLs and form/fields etc. …

## Libraries

The following external libraries were used building `UploadHandler`:

* [ErrorHandler](https://github.com/mwat56/errorhandler/)

## Licence

        Copyright © 2019, 2024 M.Watermann, 10247 Berlin, Germany
                        All rights reserved
                    EMail : <support@mwat.de>

> This program is free software; you can redistribute it and/or modify it under the terms of the GNU General Public License as published by the Free Software Foundation; either version 3 of the License, or (at your option) any later version.
>
> This software is distributed in the hope that it will be useful, but WITHOUT ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.
>
> You should have received a copy of the GNU General Public License along with this program. If not, see the [GNU General Public License](http://www.gnu.org/licenses/gpl.html) for details.

----
