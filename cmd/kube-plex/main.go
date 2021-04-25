package main

import (
	"context"
	"os"
	"path/filepath"

	"github.com/munnerz/kube-plex/pkg/signals"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/klog/v2"
)

// data pvc name
var dataPVC = os.Getenv("DATA_PVC")

// config pvc name
var configPVC = os.Getenv("CONFIG_PVC")

// transcode pvc name
var transcodePVC = os.Getenv("TRANSCODE_PVC")

// image for the plexmediaserver container containing the transcoder. This
// should be set to the same as the 'master' pms server
var pmsInternalAddress = os.Getenv("PMS_INTERNAL_ADDRESS")

func main() {
	env := os.Environ()
	args := os.Args

	ctx := context.Background()

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
		klog.Exitf("Error building kubernetes clientset: %s", err)
	}
	m, err := getMetadata(os.DirFS("/pod-info"))
	if err != nil {
		klog.Exitf("Failed to read Plex Mediaserver metadata: %v", err)
	}

	err = m.FetchAPI(ctx, kubeClient)
	if err != nil {
		klog.Exitf("Error when fetching PMS pod metadata")
	}

	cwd, err := os.Getwd()
	if err != nil {
		klog.Exitf("Error getting working directory: %s", err)
	}
	r := rewriter{pmsInternalAddress: pmsInternalAddress}
	job, err := generateJob(cwd, m, r.Env(env), r.Args(args))
	if err != nil {
		klog.Exitf("Error while generating Job: %v", err)
	}

	job, err = kubeClient.BatchV1().Jobs(job.Namespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		klog.Exitf("Error creating pod: %s", err)
	}

	stopCh := signals.SetupSignalHandler()
	waitFn := func() <-chan error {
		stopCh := make(chan error)
		go func() {
			stopCh <- waitForPodCompletion(ctx, kubeClient, job)
		}()
		return stopCh
	}

	select {
	case err := <-waitFn():
		if err != nil {
			klog.Infof("Error waiting for pod to complete: %s", err)
		}
	case <-stopCh:
		klog.Infof("Exit requested.")
	}

	klog.Infof("Cleaning up pod...")
	err = kubeClient.BatchV1().Jobs(job.Namespace).Delete(ctx, job.Name, metav1.DeleteOptions{})
	if err != nil {
		klog.Exitf("Error cleaning up pod: %s", err)
	}
}
