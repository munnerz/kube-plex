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

func main() {
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

	podName := os.Getenv("POD_NAME")
	podNamespace := os.Getenv("POD_NAMESPACE")

	m, err := FetchMetadata(ctx, kubeClient, podName, podNamespace)
	if err != nil {
		klog.Exitf("Error when fetching PMS pod metadata")
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
