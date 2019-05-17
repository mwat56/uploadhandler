/*
   Copyright © 2019 M.Watermann, 10247 Berlin, Germany
                   All rights reserved
               EMail : <support@mwat.de>
*/

package uploadhandler

import "testing"

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
