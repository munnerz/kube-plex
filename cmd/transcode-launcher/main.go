package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"

	"k8s.io/klog/v2"
)

var (
	port   = flag.Int("port", 32400, "Port on which to listen for Plex traffic")
	pmsUrl = flag.String("pms-url", os.Getenv("PMS_INTERNAL_ADDRESS"), "URL for the Plex Media Server instance")
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	klog.Info("Transcode launcher starting...")

	if *pmsUrl == "" {
		klog.Exitf("No Plex address defined (pms-url flag)")
	}

	url, err := url.Parse(*pmsUrl)
	if err != nil {
		klog.Exitf("Unable to parse Plex url: %v", err)
	}

	klog.Infof("Creating reverse proxy on port %d to %s", *port, *pmsUrl)
	p := httputil.NewSingleHostReverseProxy(url)
	s := &http.Server{
		Addr:    fmt.Sprintf("localhost:%d", *port),
		Handler: p,
	}
	defer s.Close()

	srvErr := make(chan error)
	go func() { srvErr <- s.ListenAndServe() }()

	a := flag.Args()
	klog.Infof("Transcode requested with command: %v", a)
	cmd := exec.Command(a[0], a[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	cmdErr := make(chan error)
	go func() {
		klog.Info("Transcode output begins...")
		klog.Info("--------------------------------------------")
		cmdErr <- cmd.Run()
	}()

	select {
	case err := <-srvErr:
		klog.Exitf("reverse proxy exited with error: %v", err)
	case err := <-cmdErr:
		if err != nil {
			klog.Exitf("transcode failed with error: %v", err)
		}
	}
}
