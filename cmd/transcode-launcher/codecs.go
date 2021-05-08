package main

import (
	"archive/tar"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"k8s.io/klog/v2"
)

func downloadCodecs(path, url string) error {
	err := os.MkdirAll(path, 0777)
	if err != nil {
		return fmt.Errorf("failed to create codec directory: %v", err)
	}

	res, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("error when fetching codec package: %v", err)
	}

	// ensure that res.Body is closed
	defer res.Body.Close()
	return unpackCodecs(path, res.Body)
}

func unpackCodecs(dest string, r io.Reader) error {
	klog.Infof("Unpacking codecs to: %s", dest)
	tr := tar.NewReader(r)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("invalid tar header: %v", err)
		}

		finfo := hdr.FileInfo()
		name := hdr.Name
		absName := filepath.Join(dest, name)
		if finfo.IsDir() {
			if err := os.MkdirAll(absName, finfo.Mode()); err != nil {
				return fmt.Errorf("failed to create directory %s: %v", name, err)
			}
			klog.V(1).Infof("Mkdir(%s)", name)
			continue
		}

		file, err := os.OpenFile(
			absName,
			os.O_RDWR|os.O_CREATE|os.O_TRUNC,
			finfo.Mode().Perm(),
		)
		if err != nil {
			return fmt.Errorf("error creating file %s: %v", name, err)
		}

		n, cErr := io.Copy(file, tr)
		if err := file.Close(); err != nil {
			return fmt.Errorf("error while closing file %s: %v", name, err)
		}
		if cErr != nil {
			return fmt.Errorf("error while writing to file %s: %v", name, err)
		}
		if n != finfo.Size() {
			return fmt.Errorf("wrote %d, size %d", n, finfo.Size())
		}
		klog.V(1).Infof("Write(%s)", name)
	}
	return nil
}
