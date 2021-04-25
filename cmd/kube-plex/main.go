package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/munnerz/kube-plex/pkg/signals"
	batch "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
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

// pms namespace
var namespace = os.Getenv("KUBE_NAMESPACE")

// image for the plexmediaserver container containing the transcoder. This
// should be set to the same as the 'master' pms server
var pmsImage = os.Getenv("PMS_IMAGE")
var pmsInternalAddress = os.Getenv("PMS_INTERNAL_ADDRESS")

func main() {
	env := os.Environ()
	args := os.Args

	ctx := context.Background()

	cwd, err := os.Getwd()
	if err != nil {
		klog.Exitf("Error getting working directory: %s", err)
	}
	r := rewriter{pmsInternalAddress: pmsInternalAddress}
	job := generateJob(cwd, r.Env(env), r.Args(args))

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

	job, err = kubeClient.BatchV1().Jobs(namespace).Create(ctx, job, metav1.CreateOptions{})
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
	err = kubeClient.BatchV1().Jobs(namespace).Delete(ctx, job.Name, metav1.DeleteOptions{})
	if err != nil {
		klog.Exitf("Error cleaning up pod: %s", err)
	}
}

type rewriter struct {
	pmsInternalAddress string
}

// rewriteEnv rewrites environment variables to be passed to the transcoder
func (r rewriter) Env(in []string) []string {
	return in
}

// Args rewrites argument list to use kube-plex specific values
func (r rewriter) Args(args []string) []string {
	out := make([]string, len(args))
	copy(out, args)
	for i, v := range args {
		switch v {
		case "-progressurl", "-manifest_name", "-segment_list":
			out[i+1] = strings.Replace(out[i+1], "http://127.0.0.1:32400", r.pmsInternalAddress, 1)
		case "-loglevel", "-loglevel_plex":
			out[i+1] = "debug"
		}
	}
	return out
}

func generateJob(cwd string, env []string, args []string) *batch.Job {
	envVars := toCoreV1EnvVar(env)
	var ttl, backoff int32
	ttl = int32((24 * time.Hour).Seconds())
	backoff = 1
	return &batch.Job{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "pms-elastic-transcoder-",
		},
		Spec: batch.JobSpec{
			BackoffLimit:            &backoff,
			TTLSecondsAfterFinished: &ttl,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					NodeSelector: map[string]string{
						"beta.kubernetes.io/arch": "amd64",
					},
					RestartPolicy: corev1.RestartPolicyNever,
					Containers: []corev1.Container{
						{
							Name:       "plex",
							Command:    args,
							Image:      pmsImage,
							Env:        envVars,
							WorkingDir: cwd,
							VolumeMounts: []corev1.VolumeMount{
								{Name: "data", MountPath: "/data", ReadOnly: true},
								{Name: "transcode", MountPath: "/transcode"},
							},
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "data",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: dataPVC,
								},
							},
						},
						{
							Name: "transcode",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: transcodePVC,
								},
							},
						},
					},
				},
			},
		},
	}
}

func toCoreV1EnvVar(in []string) []corev1.EnvVar {
	out := make([]corev1.EnvVar, len(in))
	for i, v := range in {
		splitvar := strings.SplitN(v, "=", 2)
		out[i] = corev1.EnvVar{
			Name:  splitvar[0],
			Value: splitvar[1],
		}
	}
	return out
}

func waitForPodCompletion(ctx context.Context, cl kubernetes.Interface, job *batch.Job) error {
	for {
		job, err := cl.BatchV1().Jobs(job.Namespace).Get(ctx, job.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		switch {
		case job.Status.Failed > 0:
			return fmt.Errorf("pod %q failed", job.Name)
		case job.Status.Succeeded > 0:
			return nil
		}
		time.Sleep(1 * time.Second)
	}
}
