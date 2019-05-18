/*
   Copyright © 2019 M.Watermann, 10247 Berlin, Germany
                   All rights reserved
               EMail : <support@mwat.de>
*/

package uploadhandler

/*
CREDITS for getting me started:
https://zupzup.org/go-http-file-upload-download/
*/

import (
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	errorhandler "github.com/mwat56/go-errorhandler"
)

type (
	// TUploadHandler embeds a `TErrorPager` which provides error page handling.
	TUploadHandler struct {
		dd string                   // destination directory
		ep errorhandler.TErrorPager // provider of customised error pages
		fn string                   // form field name
		ms int64                    // max. upload file size
	}
)

// `returnError()` sends a (possibly customized) error message
// to the remote user.
//
// `aWriter` writes the response to the remote user.
//
// `aData` is the original error text.
//
// `aStatus` is the number of the actual HTTP error.
func (uh *TUploadHandler) returnError(aWriter http.ResponseWriter,
	aData []byte, aStatus int) {
	if nil != uh.ep {
		if txt := uh.ep.GetErrorPage(aData, aStatus); 0 < len(txt) {
			aData = txt
		}
	}
	aWriter.WriteHeader(aStatus)
	aWriter.Write(aData)
} // returnError()

// ServeUpload handles the incoming file upload.
//
// The first return value will provide a short error message
// and the second return value is the HTTP status code.
// If that code is `200` (i.e. everything went well) then the
// message return value will hold the path/file name of the
// saved file.
//
// `aWriter` writes the response to the remote user.
//
// `aRequest` is the incoming upload request.
func (uh *TUploadHandler) ServeUpload(aWriter http.ResponseWriter,
	aRequest *http.Request) (string, int) {

	// first, check the file size
	aRequest.Body = http.MaxBytesReader(aWriter, aRequest.Body, uh.ms)
	if err := aRequest.ParseMultipartForm(uh.ms); nil != err {
		return "File too big", http.StatusBadRequest
	}

	// parse and validate file and post parameters
	inFile, fheader, err := aRequest.FormFile(uh.fn)
	if nil != err {
		return "Error retrieving file", http.StatusUnprocessableEntity
	}
	defer inFile.Close()

	fileType := ""
	if fileType, err = getFileContentType(inFile); nil != err {
		return "Invalid file", http.StatusUnprocessableEntity
	}
	if 0 == len(fileType) {
		fileType = "application/octet-stream"
	}

	fileExt := ""
	if fileEndings, err := mime.ExtensionsByType(fileType); nil == err {
		if nil == fileEndings {
			fileExt = ".bin"
		} else {
			fileExt = fileEndings[0]
		}
	} else {
		return "Can't read file type", http.StatusUnsupportedMediaType
	}

	switch fileExt {
	case ".asc":
		fileExt = ".txt"
		//TODO possibly re-map more extensions
	default:
		break
	}

	// preserve an (arbitrary) number of file extensions:
	switch strings.ToLower(filepath.Ext(fheader.Filename)) {
	case fileExt:
		fileExt = ""
	case ".amr", ".avi", ".bak", ".bibtex", ".bz2",
		".cfg", ".conf", ".css", ".csv",
		".db", ".deb", ".doc", ".docx", ".dia", ".epub", ".exe",
		".flv", ".gz", ".htm", ".html", ".ics", ".iso",
		".jar", ".jpeg", ".json", ".log", ".mp3",
		".odf", ".odg", ".odp", ".ods", ".odt", ".otf", ".oxt",
		".pas", ".php", ".pl", ".ppd", ".ppt", ".pptx",
		".mpg", ".rip", ".rpm", ".sh", ".spk", ".sql", ".sxg", ".sxw",
		".ttf", ".txt", ".vbox", ".vmdk", ".vcs",
		".wav", ".xhtml", ".xls", ".xpi", ".xsl":
		fileExt = ""
	}

	newPathFn := filepath.Join(uh.dd, fmt.Sprintf("%x_%s%s",
		time.Now().UnixNano(),
		strings.ReplaceAll(fheader.Filename, " ", "_"),
		fileExt))

	//TODO
	//FIXME use os.Rename() instead of copying the whole data
	//

	// copy file into the configured destination directory
	newFile, err := os.OpenFile(newPathFn, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0664)
	if nil != err {
		return "Can't open destination file", http.StatusInternalServerError
	}
	defer newFile.Close()

	if _, err := io.Copy(newFile, inFile); nil != err {
		return "Can't write destination file", http.StatusInsufficientStorage
	}

	return newPathFn, 200
} // ServeUpload()

/* * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * * */

// `getFileContentType()` returns the content type of `aFile`
//
// The given `aFile` is expected to implement both the `io.Reader`
// and the `io.Seeker` interfaces.
//
// `aFile` is the file the data of which is checked.
func getFileContentType(aFile multipart.File) (string, error) {
	// make sure to return to the start of file:
	defer aFile.Seek(0, io.SeekStart)

	fileBuf := make([]byte, 512)
	if len, err := aFile.Read(fileBuf); (nil != err) && (64 > len) {
		return "", err
	}
	contentType := http.DetectContentType(fileBuf)

	return contentType, nil
} // getFileContentType()

// NewHandler returns a new `tUploadHandler` instance.
//
// `aDestDir` is the directory to place the uploaded files.
//
// `aFieldName` the name/ID of the form/input holding the uploaded file.
//
// `aMaxSize` the max. accepted size of uploaded files.
func NewHandler(aDestDir, aFieldName string, aMaxSize int64) *TUploadHandler {
	result := TUploadHandler{
		fn: aFieldName,
		ms: aMaxSize,
	}
	if bd, err := filepath.Abs(aDestDir); nil == err {
		result.dd = bd
	} else {
		result.dd = aDestDir
	}

	return &result
} // NewHandler()

var (
	// RegEx to find initial/leading path
	pathRE = regexp.MustCompile(`^/?([\w\._-]+)?/?`)
)

// `urlPath()` returns the base-directory of `aURL`.
//
// Depending on the actual value of `aURL` the return value may
// be empty or filled; it won't hold a leading/trailing slash.
func urlPath(aURL string) string {
	if result, err := url.QueryUnescape(aURL); nil == err {
		aURL = result
	}
	matches := pathRE.FindStringSubmatch(aURL)
	if 1 < len(matches) {
		return matches[1]
	}

	return aURL
} // urlPath()

// Wrap returns a handler function that includes error page handling,
// wrapping the given `aHandler` and calling it internally.
//
// `aHandler` the previous handler responding to the HTTP request.
//
// `aDestDir` is the directory to place the uploaded files.
//
// `aFieldName` the name/ID of the form/input holding the uploaded file.
//
// `anUpURL` the URL uploads are POSTed to.
//
// `aNextURL` the URL to redirect the user after a asuccessful upload.
//
// `aMaxSize` the max. accepted size of uploaded files.
//
// `aPager` optional provider of error message pages (or `nil` if not needed).
func Wrap(aHandler http.Handler,
	aDestDir, aFieldName, anUpURL, aNextURL string,
	aMaxSize int64, aPager errorhandler.TErrorPager) http.Handler {
	uh := NewHandler(aDestDir, aFieldName, aMaxSize)
	uh.ep = aPager

	return http.HandlerFunc(
		func(aWriter http.ResponseWriter, aRequest *http.Request) {
			if ("POST" == aRequest.Method) && (urlPath(aRequest.URL.Path) == anUpURL) {
				txt, status := uh.ServeUpload(aWriter, aRequest)
				if 200 == status {
					http.Redirect(aWriter, aRequest, aNextURL, http.StatusSeeOther)
				} else {
					uh.returnError(aWriter, []byte(txt), status)
				}
				return
			}
			aHandler.ServeHTTP(aWriter, aRequest)
		})
} // Wrap()

/* EoF */
