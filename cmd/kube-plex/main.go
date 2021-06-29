package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"regexp"

	"github.com/munnerz/kube-plex/internal/ffmpeg"
	"github.com/munnerz/kube-plex/internal/logger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

func main() {
	ctx := context.Background()

	// Set up SIGKILL protection
	ctx = protectSigKill(ctx)

	// Set up logging. We can assume that Plex is running on localhost.. that's
	// what the Plex Transcoder is built to expect too.
	l, _ := logger.NewPlexLogger("KubePlex", os.Getenv("X_PLEX_TOKEN"), "http://127.0.0.1:32400/")
	klog.SetLogger(l)

	if needBypass(os.Args) {
		klog.Info("Bypassing kube-plex and launching original binary")
		bypassKubePlex(ctx)
		os.Exit(0)
	}

	// Main program start
	codecPath := ffmpeg.Unescape(os.Getenv("FFMPEG_EXTERNAL_LIBS"))
	var codecPort int
	if codecPath != "" {
		l, err := net.Listen("tcp", ":0")
		if err != nil {
			klog.Exitf("Failed to listen on any ports: %v", err)
		}
		codecPort = l.Addr().(*net.TCPAddr).Port
		go func() {
			err := startCodecServe(codecPath, l)
			if err != nil {
				klog.Errorf("Error from startCodecServe(): %v", err)
			}
		}()
		klog.Infof("Codec server listening on port %d", codecPort)
	}

	cfg, err := rest.InClusterConfig()
	if err != nil {
		// fallback to local config for development
		kubeconfig := filepath.Join("~", ".kube", "config")
		if ke := os.Getenv("KUBECONFIG"); len(ke) > 0 {
			kubeconfig = ke
		}
		cfg, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			klog.Exitf("Error building kubeconfig: %s", err)
		}
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		klog.Exitf("Error building Kubernetes clientset: %s", err)
	}

	podName := os.Getenv("POD_NAME")
	podNamespace := os.Getenv("POD_NAMESPACE")

	m, err := FetchMetadata(ctx, kubeClient, podName, podNamespace)
	if err != nil {
		klog.Exitf("Error when fetching PMS pod metadata: %v", err)
	}

	// Write codecPort to pmsMetadata
	if codecPort != 0 {
		m.CodecPort = codecPort
	}

	cwd, err := os.Getwd()
	if err != nil {
		klog.Exitf("Error getting working directory: %s", err)
	}

	env := os.Environ()
	args := os.Args

	job, err := generateJob(cwd, m, env, args)
	if err != nil {
		klog.Exitf("Error while generating Job: %v", err)
	}

	klog.Infof("Starting transcode job")

	job, err = kubeClient.BatchV1().Jobs(job.Namespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		klog.Exitf("Error creating pod: %s", err)
	}

	// Set up job deletion
	defer func() {
		// start new context for cleanup since old one should already be done
		ctx := context.Background()
		klog.Infof("Cleaning up pod...")
		bg := metav1.DeletePropagationBackground
		err = kubeClient.BatchV1().Jobs(job.Namespace).Delete(ctx, job.Name, metav1.DeleteOptions{PropagationPolicy: &bg})
		if err != nil {
			klog.Exitf("Error cleaning up pod: %s", err)
		}
	}()

	klog.Infof("Transcoder launched as job/%s (namespace: %s)", job.Name, job.Namespace)

	ctx, stop := signal.NotifyContext(ctx, shutdownSignals...)
	defer stop()

	waitCh := make(chan error)
	go func() {
		waitCh <- waitForPodCompletion(ctx, kubeClient, job)
	}()

	select {
	case err := <-waitCh:
		if err != nil {
			klog.Infof("Error waiting for pod to complete: %s", err)
		}
	case <-ctx.Done():
		if ctx.Err() != nil {
			klog.Infof("Context terminated with error:", ctx.Err())
		}
	}
	stop()
}

// Checks if bypass is needed
func needBypass(args []string) bool {
	badArg, _ := regexp.Compile("^(e?ac3|truehd|mlp)_eae$")
	for _, a := range args {
		if badArg.Match([]byte(a)) {
			return true
		}
	}
	return false
}

// re-execute original transcoder
func bypassKubePlex(ctx context.Context) {
	args := os.Args
	tc := args[0]
	tc = tc + ".orig"

	// Setup original process
	cmd := exec.CommandContext(ctx, tc, args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	ecode := 0
	if err != nil {
		if cerr, ok := err.(*exec.ExitError); ok {
			ecode = cerr.ExitCode()
		}
		fmt.Printf("Error while starting original binary: %v\n", err)
	}

	// Quit once subprocess is done
	os.Exit(ecode)
}
