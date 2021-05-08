package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"

	"github.com/munnerz/kube-plex/internal/ffmpeg"
	"github.com/munnerz/kube-plex/internal/logger"
	"k8s.io/klog/v2"
)

var (
	listenAddr  = flag.String("listen", ":32400", "Address on which to listen for Plex traffic")
	pmsAddr     = flag.String("pms-addr", "", "Address for the Plex Media Server instance (for example: '10.1.2.3:32400')")
	codecServer = flag.String("codec-server-url", os.Getenv("CODEC_SERVER"), "URL for codec server (kube-plex)")
	codecDir    = flag.String("codec-dir", os.Getenv("FFMPEG_EXTERNAL_LIBS"), "Directory to write codecs to, path will be created if doesn't exist")
	logLevel    = flag.String("loglevel", "", "Set the loglevel for transcoding process")
)

func main() {
	rcode := launch()
	os.Exit(rcode)
}

func launch() int {
	klog.InitFlags(nil)
	flag.Parse()

	// Set up logging.
	l, _ := logger.NewPlexLogger("KubePlexProxy", os.Getenv("X_PLEX_TOKEN"), fmt.Sprintf("http://%s/", *pmsAddr))
	klog.SetLogger(l)

	// Main launcher start
	klog.Info("Transcode launcher starting...")
	klog.Infof("Codec directory: %s", *codecDir)

	ctx := context.Background()

	if *codecServer != "" && *codecDir != "" {
		klog.Infof("Codec server: %s", *codecServer)
		err := downloadCodecs(*codecDir, *codecServer)
		if err != nil {
			klog.ErrorS(err, "failed to download codecs")
			return 1
		}

		// write escaped codec directory to FFmpeg environmen
		// Optimally this should be modified in the command below, this is simpler
		eCodecDir := ffmpeg.Escape(*codecDir)
		klog.Infof("Updating environment, setting FFMPEG_EXTERNAL_LIBS to '%v'", eCodecDir)
		os.Setenv("FFMPEG_EXTERNAL_LIBS", eCodecDir)
	}

	if *pmsAddr == "" {
		klog.Error("No Plex address defined (pms-url flag)")
		return 1
	}

	klog.Infof("Creating tunnel server on port %s to %s", *listenAddr, *pmsAddr)
	srvErr := make(chan error)
	go func() { srvErr <- copyListener(ctx, *listenAddr, *pmsAddr) }()

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
		klog.Info("Transcode begins...")
		cmdErr <- cmd.Run()
	}()

	select {
	case err := <-srvErr:
		klog.ErrorS(err, "reverse proxy exited with error")
		return 1
	case err := <-cmdErr:
		if err != nil {
			klog.ErrorS(err, "transcode failed")
			return 1
		}
	}
	return 0
}
