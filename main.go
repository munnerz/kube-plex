package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/munnerz/kube-plex/pkg/signals"
	"gopkg.in/alessio/shellescape.v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
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

// Whether or not to cleanup the generated pod (for debugging).
// Defaults to TRUE.
var cleanUpPod = os.Getenv("KUBE_PLEX_CLEANUP_POD")

// UID/GID settings, if configured
var plexUID = os.Getenv("PLEX_UID")
var plexGID = os.Getenv("PLEX_GID")

func main() {
	env := os.Environ()
	args := os.Args

	rewriteEnv(env)
	args = rewriteArgs(args)

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting working directory: %s", err)
	}
	log.Printf("Creating pod with transcode args: %s", args)
	pod := generatePod(cwd, env, args)

	cfg, err := clientcmd.BuildConfigFromFlags("", "")
	if err != nil {
		log.Fatalf("Error building kubeconfig: %s", err)
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		log.Fatalf("Error building kubernetes clientset: %s", err)
	}

	pod, err = kubeClient.CoreV1().Pods(namespace).Create(pod)
	if err != nil {
		log.Fatalf("Error creating pod: %s", err)
	} else {
		log.Printf("Pod '%s' created.", pod.Name)
	}

	stopCh := signals.SetupSignalHandler()
	waitFn := func() <-chan error {
		stopCh := make(chan error)
		go func() {
			stopCh <- waitForPodCompletion(kubeClient, pod)
		}()
		return stopCh
	}

	select {
	case err := <-waitFn():
		if err != nil {
			log.Printf("Error waiting for pod to complete: %s", err)
		}
	case <-stopCh:
		log.Printf("Exit requested.")
	}

	if cu, err := strconv.ParseBool(cleanUpPod); cu && err != nil {
		log.Printf("Cleaning up pod...")
		err = kubeClient.CoreV1().Pods(namespace).Delete(pod.Name, nil)
		if err != nil {
			log.Fatalf("Error cleaning up pod: %s", err)
		}
	}
}

// rewriteEnv rewrites environment variables to be passed to the transcoder
func rewriteEnv(in []string) {
	// no changes needed
}

func rewriteArgs(in []string) []string {
	for i, v := range in {
		switch v {
		case "-progressurl", "-manifest_name", "-segment_list":
			in[i+1] = strings.Replace(in[i+1], "http://127.0.0.1:32400", pmsInternalAddress, 1)
		case "-loglevel", "-loglevel_plex":
			in[i+1] = "debug"
		}
	}

	var args []string

	// If UID/GID is set, we have to modify the local groups before the transcode is called.
	if plexUID != "" {
		log.Printf("PLEX_UID set. Adding usermod command to transcode command.")
		args = append([]string{fmt.Sprintf("/usr/sbin/usermod -o -u %s plex", plexUID)}, args...)
	}
	if plexGID != "" {
		log.Printf("PLEX_GID set. Adding groupmod command to transcode command.")
		args = append([]string{fmt.Sprintf("/usr/sbin/groupmod -o -g %s plex", plexGID)}, args...)
	}

	// Change entrypoint to be bash if we need it for the permissions operations.
	if plexUID != "" || plexGID != "" {
		log.Printf("PLEX_UID || PLEX_GID set. Prefixing transcode command with bash.")

		//Replace the space in "Plex Transcoder" path
		in[0] = strings.Replace(in[0], " ", "\\ ", 1)

		var escaped_in []string
		escaped_in = append(escaped_in, in[0])
		for _, v := range in[1:] {
			escaped_in = append(escaped_in, shellescape.Quote(v))
		}

		args = append([]string{"/bin/bash", "-c"},
			fmt.Sprintf("%s && su plex -s /bin/bash -c %s",
				strings.Join(args, " && "),
				shellescape.Quote(strings.Join(escaped_in, " "))))
	} else {
		return in
	}
	return args
}

func generatePod(cwd string, env []string, args []string) *corev1.Pod {
	envVars := toCoreV1EnvVar(env)
	return &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "pms-elastic-transcoder-",
		},
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
						{
							Name:      "data",
							MountPath: "/data",
							ReadOnly:  true,
						},
						{
							Name:      "config",
							MountPath: "/config",
							ReadOnly:  true,
						},
						{
							Name:      "transcode",
							MountPath: "/transcode",
						},
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
					Name: "config",
					VolumeSource: corev1.VolumeSource{
						PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
							ClaimName: configPVC,
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

func waitForPodCompletion(cl kubernetes.Interface, pod *corev1.Pod) error {
	for {
		pod, err := cl.CoreV1().Pods(pod.Namespace).Get(pod.Name, metav1.GetOptions{})
		if err != nil {
			return err
		}

		switch pod.Status.Phase {
		case corev1.PodPending:
		case corev1.PodRunning:
		case corev1.PodUnknown:
			log.Printf("Warning: pod %q is in an unknown state", pod.Name)
		case corev1.PodFailed:
			return fmt.Errorf("pod %q failed", pod.Name)
		case corev1.PodSucceeded:
			return nil
		}
		time.Sleep(1 * time.Second)
	}
}
