/*
   Copyright © 2019, 2020 M.Watermann, 10247 Berlin, Germany
                   All rights reserved
               EMail : <support@mwat.de>
*/

package uploadhandler

//lint:file-ignore ST1017 - I prefer Yoda conditions

import (
	"path/filepath"
	"regexp"
	"testing"
)

func Test_urlPath(t *testing.T) {
	type args struct {
		aURL string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		// TODO: Add test cases.
		{" 1", args{"/path/to/file"}, "path"},
		{" 2", args{"/path"}, "path"},
		{" 3", args{"/"}, ""},
		{" 4", args{""}, ""},
		{" 5", args{"path/to/file"}, "path"},
		{" 6", args{"path"}, "path"},
		{" 7", args{"p-a.t_h"}, "p-a.t_h"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := urlPath(tt.args.aURL); got != tt.want {
				t.Errorf("urlPath() = %v, want %v", got, tt.want)
			}
		})
	}
} // Test_urlPath()

func TestTUploadHandler_newFilename(t *testing.T) {
	dir, _ := filepath.Abs(`./`)
	h1 := NewHandler(dir, `Upload`, 0)
	re := regexp.MustCompile(dir + `/[a-f0-9]+_.+\.[a-z]+$`)
	type args struct {
		aFilename  string
		aExtension string
	}
	tests := []struct {
		name   string
		fields *TUploadHandler
		args   args
	}{
		// TODO: Add test cases.
		// {" 0", h1, args{``, ``}},
		{" 1", h1, args{`profile.asc`, `.asc`}},
		{" 2", h1, args{`profi.jpg`, `.jpg`}},
		{" 3", h1, args{`x.mpg`, `.mpg`}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uh := tt.fields
			if got := uh.newFilename(tt.args.aFilename, tt.args.aExtension); !re.Match([]byte(got)) {
				t.Errorf("TUploadHandler.newFilename() = '%v'", got)
			}
		})
	}
} // TestTUploadHandler_newFilename()
