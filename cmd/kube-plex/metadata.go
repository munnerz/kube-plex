package main

import (
	"context"
	"fmt"
	"io/fs"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

type pmsMetadata struct {
	// Fields fetched from pod info filesystem
	Name      string
	Namespace string

	// Fields fetched from kubernetes API
	Uuid     types.UID
	PMSImage string
	Volumes  []corev1.Volume
}

func getMetadata(m fs.FS) (pmsMetadata, error) {
	name, err := fs.ReadFile(m, "podname")
	if err != nil {
		return pmsMetadata{}, fmt.Errorf("unable to read pod name: %v", err)
	}

	if len(name) == 0 {
		return pmsMetadata{}, fmt.Errorf("pod name is empty")
	}

	ns, err := fs.ReadFile(m, "namespace")
	if err != nil {
		return pmsMetadata{}, fmt.Errorf("unable to read namespace: %v", err)
	}

	if len(ns) == 0 {
		return pmsMetadata{}, fmt.Errorf("namespace is empty")
	}

	return pmsMetadata{
		Name:      string(name),
		Namespace: string(ns),
	}, nil
}

func (p *pmsMetadata) FetchAPI(ctx context.Context, cl kubernetes.Interface) error {
	pod, err := cl.CoreV1().Pods(p.Namespace).Get(ctx, p.Name, v1.GetOptions{})
	if err != nil {
		return fmt.Errorf("unable to fetch Pod info: %v", err)
	}

	p.Uuid = pod.ObjectMeta.UID

	for _, c := range pod.Spec.Containers {
		if c.Name == "plex" {
			p.PMSImage = c.Image
			break
		}
	}

	if p.PMSImage == "" {
		return fmt.Errorf("could not find Plex container image, is there a container named `plex`?")
	}

	// Fetch data volumes from pod spec
	dv, err := getVolume(pod.Spec, "data")
	if err != nil {
		return fmt.Errorf("error when getting data volume: %v", err)
	}

	tv, err := getVolume(pod.Spec, "transcode")
	if err != nil {
		return fmt.Errorf("error when getting transcode volume: %v", err)
	}

	p.Volumes = []corev1.Volume{dv, tv}

	return nil
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
		Kind: "Pod",
		Name: p.Name,
		UID:  p.Uuid,
	}, nil
}
