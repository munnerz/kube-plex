package main

import (
	"reflect"
	"testing"
)

func Test_rewriter_Args(t *testing.T) {
	r := rewriter{
		pmsInternalAddress: "http://test-svc:32400",
	}
	tests := []struct {
		name      string
		args      []string
		want      []string
		ischanged bool
	}{
		{"unmodified args", []string{"-args", "arg1"}, []string{"-args", "arg1"}, false},
		{"modifies loglevel", []string{"-s", "1", "-loglevel", "info"}, []string{"-s", "1", "-loglevel", "debug"}, true},
		{"modifies plex loglevel", []string{"-s", "1", "-loglevel_plex", "info"}, []string{"-s", "1", "-loglevel_plex", "debug"}, true},
		{"server address is adjusted", []string{"-progressurl", "http://127.0.0.1:32400/", "-l", "i"}, []string{"-progressurl", "http://test-svc:32400/", "-l", "i"}, true},
		{"manifest url is adjusted", []string{"-manifest_name", "http://127.0.0.1:32400/manifest", "-l", "i"}, []string{"-manifest_name", "http://test-svc:32400/manifest", "-l", "i"}, true},
		{"segment list url is adjusted", []string{"-segment_list", "http://127.0.0.1:32400/segments/url", "-l", "i"}, []string{"-segment_list", "http://test-svc:32400/segments/url", "-l", "i"}, true},
		{"adjusts multiple arguments", []string{"-progressurl", "http://127.0.0.1:32400/", "-loglevel", "x"}, []string{"-progressurl", "http://test-svc:32400/", "-loglevel", "debug"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orig := make([]string, len(tt.args))
			copy(orig, tt.args)
			if got := r.Args(tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("rewriter.Args() = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(orig, tt.args) {
				t.Errorf("rewriter.Args() modifies input, orig: %v, after: %v", orig, tt.args)
			}
		})
	}
}
