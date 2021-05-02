package main

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

const (
	pmsUrl        = "kube-plex/pms-url"
	kubePlexImage = "kube-plex/image"
)

type pmsMetadata struct {
	// Fields fetched from pod info filesystem
	Name      string
	Namespace string

	// Fields fetched from kubernetes API
	Uuid     types.UID
	PmsImage string
	Volumes  []corev1.Volume

	// Kube-plex fields
	KubePlexImage string
	PmsURL        string
}

func FetchMetadata(ctx context.Context, cl kubernetes.Interface, name, namespace string) (pmsMetadata, error) {
	m := pmsMetadata{Name: name, Namespace: namespace}

	if m.Name == "" {
		return pmsMetadata{}, fmt.Errorf("pod name is empty")
	}

	if m.Namespace == "" {
		return pmsMetadata{}, fmt.Errorf("namespace is empty")
	}

	pod, err := cl.CoreV1().Pods(namespace).Get(ctx, name, v1.GetOptions{})
	if err != nil {
		return pmsMetadata{}, fmt.Errorf("unable to fetch Pod info: %v", err)
	}

	m.Uuid = pod.ObjectMeta.UID

	for _, c := range pod.Spec.Containers {
		if c.Name == "plex" {
			m.PmsImage = c.Image
			break
		}
	}

	if m.PmsImage == "" {
		return pmsMetadata{}, fmt.Errorf("could not find Plex container image, is there a container named `plex`?")
	}

	// Fetch data volumes from pod spec
	dv, err := getVolume(pod.Spec, "data")
	if err != nil {
		return pmsMetadata{}, fmt.Errorf("error when getting data volume: %v", err)
	}

	tv, err := getVolume(pod.Spec, "transcode")
	if err != nil {
		return pmsMetadata{}, fmt.Errorf("error when getting transcode volume: %v", err)
	}

	m.Volumes = []corev1.Volume{dv, tv}

	// Get PMS URL
	a := pod.GetAnnotations()
	u, ok := a[pmsUrl]
	if !ok {
		return pmsMetadata{}, fmt.Errorf("unable to determine plex service URL")
	}

	m.PmsURL = u

	// Get kube-plex image
	i, ok := a[kubePlexImage]
	if !ok {
		return pmsMetadata{}, fmt.Errorf("unable to determine kube-plex image")
	}

	m.KubePlexImage = i

	return m, nil
}

// getVolume returns a volume matching given name from podspec
func getVolume(podspec corev1.PodSpec, name string) (corev1.Volume, error) {
	for _, v := range podspec.Volumes {
		if v.Name == name {
			return v, nil
		}
	}
	return corev1.Volume{}, fmt.Errorf("volume %s not found", name)
}

func (p pmsMetadata) OwnerReference() (v1.OwnerReference, error) {
	if p.Uuid == "" {
		return v1.OwnerReference{}, fmt.Errorf("UUID is empty, has Fetch() been run?")
	}

	return v1.OwnerReference{
		APIVersion: "v1",
		Kind:       "Pod",
		Name:       p.Name,
		UID:        p.Uuid,
	}, nil
}

func (p pmsMetadata) LauncherCmd(args ...string) []string {
	a := []string{
		"/shared/transcode-launcher",
		fmt.Sprintf("--pms-url=%s", p.PmsURL),
		"--port=32400", "--",
	}
	return append(a, args...)
}
