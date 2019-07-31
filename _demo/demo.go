package main

import (
	"log"
	"net/http"
	"os"

	"github.com/mwat56/uploadhandler"
)

// `testHandler()` is the demo page handler
//
// `aWriter` writes the response to the remote user.
//
// `aRequest` is the remote user's incoming request.
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
	_, _ = aWriter.Write([]byte(page))
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
