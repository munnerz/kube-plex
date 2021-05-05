package main

import (
	"archive/tar"
	"fmt"
	"io"
	"io/fs"
	"net"
	"net/http"
	"os"

	"k8s.io/klog/v2"
)

type codecServe struct {
	fs fs.FS
}

func (c codecServe) codecPackage(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-type", "application/x-tar")
	tw := tar.NewWriter(w)
	err := fs.WalkDir(c.fs, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("received error from WalkDir: %v", err)
		}

		// Don't add the parent directory to tar
		if path == "." {
			return nil
		}

		f, err := d.Info()
		if err != nil {
			return fmt.Errorf("file info read failed: %v", err)
		}

		hdr, err := tar.FileInfoHeader(f, f.Name())
		if err != nil {
			return fmt.Errorf("unable to create tar header for file %s: %v", f.Name(), err)
		}

		// Use relative path as the filename, since tar wants a parent directories for the entry
		hdr.Name = path

		if err := tw.WriteHeader(hdr); err != nil {
			return fmt.Errorf("tar header write failed: %v", err)
		}

		// for directories, we only need the header
		if d.IsDir() {
			return nil
		}

		fr, err := c.fs.Open(path)
		if err != nil {
			return fmt.Errorf("could not open file %s: %v", d.Name(), err)
		}
		defer fr.Close()

		if _, err := io.Copy(tw, fr); err != nil {
			return fmt.Errorf("writing to tar failed: %v", err)
		}
		return nil
	})
	if err != nil {
		klog.Errorf("failed to create tar package: %v", err)
	}
}

func startCodecServe(path string, l net.Listener) error {
	fp := os.DirFS(path)
	f := codecServe{fs: fp}
	http.HandleFunc("/", f.codecPackage)
	return http.Serve(l, nil)
}
