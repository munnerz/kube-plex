package main

import (
	"context"
	"fmt"
	"io"
	"net"

	"k8s.io/klog/v2"
)

// Convenience wrapper for listening on a given port and launcing dialAndCopy() for every
// incoming connection
func copyListener(ctx context.Context, listenAddr, serverAddr string) error {
	l, err := net.Listen("tcp", listenAddr)
	if err != nil {
		return fmt.Errorf("failed to listen on %s: %v", listenAddr, err)
	}
	defer l.Close()

	for {
		cConn, err := l.Accept()
		if err != nil {
			return fmt.Errorf("Accept() returned an error: %v", err)
		}
		go dialAndCopy(ctx, cConn, serverAddr)
	}
}

// dialAndCopy is a naive tunnel between 2 connections. It copies input and output between
// The client and server. Any errors will close the connection
func dialAndCopy(ctx context.Context, cConn net.Conn, addr string) {
	// Close client connection once we are done
	defer cConn.Close()

	var d net.Dialer
	sConn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		klog.Exitf("Dial() failed: %v", err)
	}
	defer sConn.Close()
	inErr := make(chan error, 1)
	outErr := make(chan error, 1)

	go func() {
		_, err := io.Copy(sConn, cConn)
		inErr <- err
	}()

	go func() {
		_, err := io.Copy(cConn, sConn)
		outErr <- err
	}()

	select {
	case err = <-inErr:
		if err != nil {
			klog.Infof("error while reading from client: %v", err)
		}
	case err = <-outErr:
		if err != nil {
			klog.Infof("error while reading from server: %v", err)
		}
	case <-ctx.Done():
		klog.Infof("context done")
	}
}
