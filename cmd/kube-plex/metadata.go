package main

import (
	"context"
	"fmt"
	"sort"
	"strings"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
)

const (
	pmsURL            = "kube-plex/pms-addr"
	pmsContainer      = "kube-plex/pms-container-name"
	pmsMounts         = "kube-plex/mounts"
	kubePlexLevel     = "kube-plex/loglevel"
	kubePlexContainer = "kube-plex/container-name"
)

// PmsMetadata describes a Plex Media Server instance running in kubernetes.
type PmsMetadata struct {
	Name          string               // Pod Name
	Namespace     string               // Pod Namespace
	UID           types.UID            // Pod UID
	PodIP         string               // Pod IP address
	Mounts        []string             // List of mounts (paths) to copy to transcoder
	VolumeMounts  []corev1.VolumeMount // kube-plex volume mounts
	Volumes       []corev1.Volume      // kube-plex needed volumes
	KubePlexImage string               // container image for kube-plex
	KubePlexLevel string               // loglevel of kubeplex processes
	CodecPort     int                  // port on which the codec service runs
	PmsImage      string               // container image used by Plex Media Server
	PmsAddr       string               // URL for Plex Media Server
}

// FetchMetadata fetches and populates a metadata object based on the current environment
func FetchMetadata(ctx context.Context, cl kubernetes.Interface, name, namespace string) (PmsMetadata, error) {
	if name == "" {
		return PmsMetadata{}, fmt.Errorf("pod name is empty")
	}

	if namespace == "" {
		return PmsMetadata{}, fmt.Errorf("namespace is empty")
	}

	pod, err := cl.CoreV1().Pods(namespace).Get(ctx, name, v1.GetOptions{})
	if err != nil {
		return PmsMetadata{}, fmt.Errorf("unable to fetch Pod info: %v", err)
	}

	m := PmsMetadata{
		Name:      pod.GetName(),
		Namespace: pod.GetNamespace(),
		UID:       pod.GetUID(),
		PodIP:     pod.Status.PodIP,
	}

	// Get PMS URL
	a := pod.GetAnnotations()
	u, ok := a[pmsURL]
	if !ok {
		return PmsMetadata{}, fmt.Errorf("unable to determine plex service URL")
	}
	m.PmsAddr = u

	// Get debugging status
	d := a[kubePlexLevel]
	// TODO: It would be nice to enforce all valid options here
	m.KubePlexLevel = d

	// Plex media server container image
	pmsimage, pmsname, err := getContainerImage(pmsContainer, "plex", pod, pod.Status.ContainerStatuses)
	if err != nil {
		return PmsMetadata{}, fmt.Errorf("unable to determine Plex Media server image (set container name with '%s' annotation): %v", pmsContainer, err)
	}
	m.PmsImage = pmsimage

	// Kube-Plex container image
	kpimage, _, err := getContainerImage(kubePlexContainer, "kube-plex-init", pod, pod.Status.InitContainerStatuses)
	if err != nil {
		return PmsMetadata{}, fmt.Errorf("unable to determine kube-plex image (set init-container name with '%s' annotation): %v", kubePlexContainer, err)
	}
	m.KubePlexImage = kpimage

	// mounts to copy over
	mlist, ok := a[pmsMounts]
	if !ok {
		// default value, matches the old behaviour
		mlist = "/transcode,/data"
	}
	if mlist != "" {
		m.Mounts = strings.Split(mlist, ",")
	}

	v, vm, err := getVolumesAndMounts(m.Mounts, pod, pmsname)
	if err != nil {
		return PmsMetadata{}, fmt.Errorf("failed to get volumes and mounts: %v", err)
	}
	m.VolumeMounts = vm
	m.Volumes = v

	return m, nil
}

// getContainerImage from pod status based on the annotation given
func getContainerImage(annotation, defname string, pod *corev1.Pod, status []corev1.ContainerStatus) (string, string, error) {
	a := pod.GetAnnotations()
	name, ok := a[annotation]
	if !ok {
		name = defname
	}
	for _, c := range status {
		if c.Name == name {
			return c.ImageID, name, nil
		}
	}
	return "", "", fmt.Errorf("no containers found by name %s", name)
}

// getVolumesAndMounts for given directories in the pod
func getVolumesAndMounts(dirs []string, pod *corev1.Pod, name string) ([]corev1.Volume, []corev1.VolumeMount, error) {
	if len(dirs) == 0 {
		return nil, nil, nil
	}

	var c *corev1.Container
	for _, d := range pod.Spec.Containers {
		if d.Name == name {
			c = &d
		}
	}
	if c == nil {
		return nil, nil, fmt.Errorf("container %s not found in pod", name)
	}

	mounts := map[string]corev1.VolumeMount{}
	for _, m := range c.VolumeMounts {
		mounts[m.MountPath] = m
	}

	volumes := map[string]corev1.Volume{}
	for _, v := range pod.Spec.Volumes {
		volumes[v.Name] = v
	}

	// since a single volume can be mounted multiple times, deduplicate volumes
	var vm []corev1.VolumeMount
	vtmp := map[string]corev1.Volume{}
	for _, d := range dirs {
		m, ok := mounts[d]
		if !ok {
			return nil, nil, fmt.Errorf("no volume mount defined for '%s'", d)
		}
		vm = append(vm, m)

		v, ok := volumes[m.Name]
		if !ok {
			return nil, nil, fmt.Errorf("no volume definition found for volume '%s'", m.Name)
		}
		vtmp[m.Name] = v
	}

	// Finally collect the volumes in a slice for returning (in a predictable order for testing)
	vkeys := make([]string, 0, len(vtmp))
	for k := range vtmp {
		vkeys = append(vkeys, k)
	}
	sort.Strings(vkeys)

	var v []corev1.Volume
	for _, k := range vkeys {
		v = append(v, vtmp[k])
	}

	return v, vm, nil
}

// OwnerReference creates an owner reference that can be used to trigger cleanup
// when this PMS instance is deleted
func (p PmsMetadata) OwnerReference() (v1.OwnerReference, error) {
	if p.UID == "" {
		return v1.OwnerReference{}, fmt.Errorf("UUID is empty, has Fetch() been run?")
	}

	return v1.OwnerReference{
		APIVersion: "v1",
		Kind:       "Pod",
		Name:       p.Name,
		UID:        p.UID,
	}, nil
}

// LauncherCmd returns a valid launcher command for this transcode operation
func (p PmsMetadata) LauncherCmd(args ...string) []string {
	a := []string{
		"/shared/transcode-launcher",
		fmt.Sprintf("--pms-addr=%s", p.PmsAddr),
		"--listen=:32400",
	}
	if p.CodecPort != 0 {
		a = append(a,
			fmt.Sprintf("--codec-server-url=http://%s:%d/", p.PodIP, p.CodecPort),
			"--codec-dir=/shared/codecs/",
		)
	}
	if p.KubePlexLevel != "" {
		a = append(a, fmt.Sprintf("--loglevel=%s", p.KubePlexLevel))
	}
	a = append(a, "--")
	return append(a, args...)
}
