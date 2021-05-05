package main

import (
	"archive/tar"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"testing/fstest"
)

func Test_codecPackage(t *testing.T) {
	tests := []struct {
		name     string
		fs       fstest.MapFS
		contents []string
	}{
		{"packages a directory", fstest.MapFS{"foo/bar": &fstest.MapFile{Data: []byte("bar")}}, []string{"foo", "foo/bar"}},
		{"handles empty diretory", fstest.MapFS{}, []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := codecServe{fs: tt.fs}
			ts := httptest.NewServer(http.HandlerFunc(c.codecPackage))
			defer ts.Close()

			res, err := http.Get(ts.URL)
			if err != nil {
				t.Fatalf("codecPackage() HTTP GET err = %v", err)
			}

			if res.StatusCode != http.StatusOK {
				t.Errorf("codecPackage() returned HTTP error: %v", res.Status)
			}

			if res.Header.Get("Content-Type") != "application/x-tar" {
				t.Errorf("codecPackage() returned incorrect content type: %v", res.Header.Get("Content-Type"))
			}

			tr := tar.NewReader(res.Body)
			defer res.Body.Close()

			files := []string{}
			for {
				hdr, err := tr.Next()
				if err == io.EOF {
					break
				}
				if err != nil {
					t.Fatalf("error reading tar stream: %v", err)
				}
				files = append(files, hdr.Name)
			}

			if !reflect.DeepEqual(tt.contents, files) {
				t.Errorf("codecPackage() contents differ, want = %v, got = %v", tt.contents, files)
			}

		})
	}
}
