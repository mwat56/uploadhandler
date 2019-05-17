package main

import (
	"log"
	"net/http"

	uploadhandler "github.com/mwat56/go-uploadhandler"
)

func testHandler(aWriter http.ResponseWriter, aRequest *http.Request) {
	page := `<!DOCTYPE html><html><title>Go upload test</title><body>
<form action="http://localhost:8080/up" method="post" enctype="multipart/form-data">
	<p><label for="uploadFile">Filename:</label>
	<input type="file" name="uploadFile" id="uploadFile"></p>
	<p><input type="submit" name="submit" value="Submit"></p>
</form></body></html>`
	aWriter.WriteHeader(200)
	aWriter.Write([]byte(page))
	// POST is handled by the UploadHandler
} // testHandler()

func main() {
	handler := uploadhandler.Wrap(http.HandlerFunc(testHandler), "./static", "uploadFile", "up", "/", 500*1024*1024, nil)

	if err := http.ListenAndServe("127.0.0.1:8080", handler); nil != err {
		log.Fatalf("Problem: %v", err)
	}
} // main()
