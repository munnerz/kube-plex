package main

import (
	"flag"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"

	"github.com/munnerz/kube-plex/internal/ffmpeg"
	"k8s.io/klog/v2"
)

var (
	port        = flag.Int("port", 32400, "Port on which to listen for Plex traffic")
	pmsUrl      = flag.String("pms-url", os.Getenv("PMS_INTERNAL_ADDRESS"), "URL for the Plex Media Server instance")
	codecServer = flag.String("codec-server-url", os.Getenv("CODEC_SERVER"), "URL for codec server (kube-plex)")
	codecDir    = flag.String("codec-dir", os.Getenv("FFMPEG_EXTERNAL_LIBS"), "Directory to write codecs to, path will be created if doesn't exist")
	logLevel    = flag.String("loglevel", "", "Set the loglevel for transcoding process")
)

func main() {
	klog.InitFlags(nil)
	flag.Parse()

	klog.Info("Transcode launcher starting...")
	klog.Infof("Codec directory: %s", *codecDir)

	if *codecServer != "" && *codecDir != "" {
		klog.Infof("Codec server: %s", *codecServer)
		err := downloadCodecs(*codecDir, *codecServer)
		if err != nil {
			klog.Exitf("failed to download codecs: %v", err)
		}

		// write escaped codec directory to FFmpeg environmen
		// Optimally this should be modified in the command below, this is simpler
		eCodecDir := ffmpeg.Escape(*codecDir)
		klog.Infof("Updating environment, setting FFMPEG_EXTERNAL_LIBS to '%v'", eCodecDir)
		os.Setenv("FFMPEG_EXTERNAL_LIBS", eCodecDir)
	}

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

	cpath := a[0]
	cargs := []string{}
	if *logLevel != "" {
		klog.Infof("Setting debug level to %s on transcode process", *logLevel)
		cargs = append(cargs,
			"-loglevel", *logLevel,
			"-loglevel_plex", *logLevel,
		)
	}
	cargs = append(cargs, a[1:]...)

	klog.Infof("Transcode requested with command %v, args = %v", a[0], cargs)
	cmd := exec.Command(cpath, cargs...)
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
