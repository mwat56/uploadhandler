/*
   Copyright © 2019, 2022 M.Watermann, 10247 Berlin, Germany
                   All rights reserved
               EMail : <support@mwat.de>
*/
package uploadhandler

//lint:file-ignore ST1017 - I prefer Yoda conditions

/*
 * CREDITS for getting me started:
 * https://zupzup.org/go-http-file-upload-download/
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

	"github.com/mwat56/errorhandler"
)

type (
	// TUploadHandler embeds a `TErrorPager` which provides
	// error page handling.
	TUploadHandler struct {
		dd string                   // destination directory
		ep errorhandler.TErrorPager // provider of customised error pages
		fn string                   // form field name
		ms int64                    // max. upload file size
	}
)

// `newFilename()` checks the filename extension of `aFilename` and
// returns a filename for the uploaded image.
//
//	`aFilename` The complete filename received with the upload request.
//	`aExtension` The filename extension based on the determined file type.
func (uh *TUploadHandler) newFilename(aFilename, aExtension string) string {
	switch aExtension {
	case ".asc":
		aExtension = ".txt"
	case ".jpg":
		aExtension = ".jpeg"
	case ".mpg":
		aExtension = ".mpeg"
		//XXX possibly re-map more extensions
	default:
		break
	}

	// Preserve an (arbitrary) number of file extensions:
	switch strings.ToLower(filepath.Ext(aFilename)) {
	case aExtension:
		aExtension = ""
	case ".amr", ".arj", ".avi", ".azw3",
		".bak", ".bibtex", ".bz2",
		".cfg", ".com", ".conf", ".css", ".csv",
		".db", ".deb", ".doc", ".docx", ".dia",
		".epub", ".exe", ".flv",
		".gif", ".gz", ".htm", ".html",
		".ics", ".iso", ".jar", ".jpeg", ".json",
		".log", ".md", ".mobi", ".mp3", ".mp4", ".mpeg",
		".odf", ".odg", ".odp", ".ods", ".odt", ".otf", ".oxt",
		".pas", ".pdf", ".php", ".pl", ".ppd", ".ppt", ".pptx",
		".rip", ".rpm",
		".sh", ".shtml", ".spk", ".sql", ".sxg", ".sxw",
		".ttf", ".txt", ".vbox", ".vmdk", ".vcs",
		".wav", ".xhtml", ".xls", ".xpi", ".xsl", ".zip":
		aExtension = ""
	}

	if 0 < len(aExtension) {
		// If we change the given filename extension remove it:
		if i := strings.LastIndexByte(aFilename, '.'); 0 < i {
			aFilename = aFilename[:i]
		}
	}

	return filepath.Join(uh.dd,
		fmt.Sprintf(`%x_%s%s`,
			time.Now().UnixNano(),
			strings.ReplaceAll(aFilename, ` `, `_`),
			aExtension))
} // newFilename()

// `returnError()` sends a (possibly customised) error message
// to the remote user.
//
//	`aWriter` writes the response to the remote user.
//	`aData` is the original error text.
//	`aStatus` is the code number of the actual HTTP error.
func (uh *TUploadHandler) returnError(aWriter http.ResponseWriter,
	aData []byte, aStatus int) {
	if nil != uh.ep {
		if txt := uh.ep.GetErrorPage(aData, aStatus); 0 < len(txt) {
			aData = txt
		}
	}
	aWriter.WriteHeader(aStatus)
	_, _ = aWriter.Write(aData)
} // returnError()

// ServeUpload handles the incoming file upload.
//
// The first return value (`rCause`) will provide a short error message
// and the second return value (`rCode`) the HTTP status code.
// If that code is `200` (i.e. everything went well) then the
// message return value (`rCause`) will hold the path/file name of
// the saved file.
//
//	`aWriter` writes the response to the remote user.
//	`aRequest` is the incoming upload request.
func (uh *TUploadHandler) ServeUpload(aWriter http.ResponseWriter,
	aRequest *http.Request) (rCause string, rCode int) {

	// first, check the file size
	aRequest.Body = http.MaxBytesReader(aWriter, aRequest.Body, uh.ms)
	if err := aRequest.ParseMultipartForm(uh.ms); nil != err {
		return "File too big", http.StatusBadRequest
	}

	// parse and validate file and post parameters
	inFile, fHeader, err := aRequest.FormFile(uh.fn)
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
	if fileEndings, err2 := mime.ExtensionsByType(fileType); nil == err2 {
		if nil == fileEndings {
			fileExt = ".bin"
		} else {
			fileExt = fileEndings[0]
		}
	} else {
		return "Can't read file type", http.StatusUnsupportedMediaType
	}

	// construct a filename for the uploaded file
	newPathFn := uh.newFilename(fHeader.Filename, fileExt)

	// copy file into the configured destination directory
	newFile, err := os.OpenFile(newPathFn, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0640) // #nosec G302
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

// `getFileContentType()` returns the content type of `aFile`.
//
// The given `aFile` is expected to implement both the `io.Reader`
// and the `io.Seeker` interfaces.
//
//	`aFile` is the file the data of which is checked.
func getFileContentType(aFile multipart.File) (string, error) {
	// make sure to return to the start of file:
	defer func() {
		_, _ = aFile.Seek(0, io.SeekStart)
	}()

	fileBuf := make([]byte, 512)
	if bLen, err := aFile.Read(fileBuf); (nil != err) && (64 > bLen) {
		return "", err
	}
	contentType := http.DetectContentType(fileBuf)

	return contentType, nil
} // getFileContentType()

// NewHandler returns a new `TUploadHandler` instance.
//
// If the `aMaxSize` value is smaller/equal to zero a maximal filesize
// of 8 MB is used.
//
//	`aDestDir` is the directory to place the uploaded files.
//	`aFieldName` the name/ID of the form/input holding the uploaded file.
//	`aMaxSize` the max. accepted size of uploaded files.
func NewHandler(aDestDir, aFieldName string, aMaxSize int64) (rHandler *TUploadHandler) {
	rHandler = &TUploadHandler{fn: aFieldName}
	if 0 < aMaxSize {
		rHandler.ms = aMaxSize
	} else {
		rHandler.ms = int64(1 << 23)
	}
	if absDir, err := filepath.Abs(aDestDir); nil == err {
		rHandler.dd = absDir
	} else {
		rHandler.dd = aDestDir
	}

	return
} // NewHandler()

var (
	// RegEx to find initial/leading path
	uhPathRE = regexp.MustCompile(`^/?([\w\._-]+)?/?`)
)

// `urlPath()` returns the base-directory of `aURL`.
//
// Depending on the actual value of `aURL` the return value may
// be empty or filled; it won't hold a leading/trailing slash.
func urlPath(aURL string) string {
	if result, err := url.QueryUnescape(aURL); nil == err {
		aURL = result
	}
	matches := uhPathRE.FindStringSubmatch(aURL)
	if 1 < len(matches) {
		return matches[1]
	}

	return aURL
} // urlPath()

// Wrap returns a handler function that includes upload handling,
// wrapping the given `aHandler` and calling it internally.
//
//	`aHandler` The previous handler responding to the HTTP request.
//	`aDestDir` Is the directory to place the uploaded files.
//	`aFieldName` The name/ID of the form/input field holding the uploaded file.
//	`aUpURL` The URL uploads are POSTed to.
//	`aNextURL` The URL to redirect the user after a successful upload.
//	`aMaxSize` The max. accepted size of uploaded files; if the given
//	value is smaller/equal to zero then a maximal filesize of 8 MB is used.
//	`aPager` Optional provider of customised error message pages
// (or `nil` if not needed).
func Wrap(aHandler http.Handler,
	aDestDir, aFieldName, aUpURL, aNextURL string,
	aMaxSize int64, aPager errorhandler.TErrorPager) http.Handler {
	uh := NewHandler(aDestDir, aFieldName, aMaxSize)
	uh.ep = aPager

	return http.HandlerFunc(
		func(aWriter http.ResponseWriter, aRequest *http.Request) {
			if ("POST" == aRequest.Method) && (urlPath(aRequest.URL.Path) == aUpURL) {
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
