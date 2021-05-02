package main

import (
	"context"
	"reflect"
	"testing"

	"github.com/go-test/deep"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func Test_pmsMetadata_FetchMetadata(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	validPod := corev1.Pod{
		ObjectMeta: v1.ObjectMeta{
			Namespace: "plex", Name: "pms", UID: "123",
			Annotations: map[string]string{"kube-plex/pms-url": "http://service:32400/", "kube-plex/image": "kubeplex:latest"},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{Name: "plex", Image: "plex:test"}},
			Volumes:    []corev1.Volume{{Name: "data"}, {Name: "transcode"}},
		},
	}

	tests := []struct {
		name         string
		podname      string
		podnamespace string
		pod          corev1.Pod
		wantPms      pmsMetadata
		wantErr      bool
	}{
		{"fetches info from api", "pms", "plex", validPod, pmsMetadata{Name: "pms", Namespace: "plex", Uuid: "123", PmsImage: "plex:test", KubePlexImage: "kubeplex:latest", PmsURL: "http://service:32400/", Volumes: []corev1.Volume{{Name: "data"}, {Name: "transcode"}}}, false},
		{"fails on missing podname", "", "plex", validPod, pmsMetadata{}, true},
		{"fails on missing namespace", "pms", "", validPod, pmsMetadata{}, true},
		{"fails gracefully on wrong pod name", "wrong", "plex", validPod, pmsMetadata{}, true},
		{"fails gracefully on wrong namespace", "pms", "wrong", validPod, pmsMetadata{}, true},
		{"plex container missing", "pms", "plex", corev1.Pod{ObjectMeta: validPod.ObjectMeta, Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "wrong", Image: "pms:own"}}, Volumes: []corev1.Volume{{Name: "data"}, {Name: "transcode"}}}}, pmsMetadata{}, true},
		{"plex data volume missing", "pms", "plex", corev1.Pod{ObjectMeta: validPod.ObjectMeta, Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "plex", Image: "pms:own"}}, Volumes: []corev1.Volume{{Name: "transcode"}}}}, pmsMetadata{}, true},
		{"plex transcode volume missing", "pms", "plex", corev1.Pod{ObjectMeta: validPod.ObjectMeta, Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "plex", Image: "pms:own"}}, Volumes: []corev1.Volume{{Name: "data"}}}}, pmsMetadata{}, true},
		{"plex service annotation missing", "pms", "plex", corev1.Pod{ObjectMeta: v1.ObjectMeta{Namespace: "plex", Name: "pms", UID: "123", Annotations: map[string]string{"kube-plex/image": "kp:latest"}}, Spec: validPod.Spec}, pmsMetadata{}, true},
		{"kube-plex image annotation missing", "pms", "plex", corev1.Pod{ObjectMeta: v1.ObjectMeta{Namespace: "plex", Name: "pms", UID: "123", Annotations: map[string]string{"kube-plex/pms-url": "http://p/"}}, Spec: validPod.Spec}, pmsMetadata{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl := fake.NewSimpleClientset(&tt.pod)
			m, err := FetchMetadata(ctx, cl, tt.podname, tt.podnamespace)
			if (err != nil) != tt.wantErr {
				t.Errorf("pmsMetadata.FetchAPI() error = %v, wantErr %v", err, tt.wantErr)
			}
			// We don't want to check output state if error occurs
			if !tt.wantErr {
				if diff := deep.Equal(m, tt.wantPms); diff != nil {
					t.Errorf("pmsMetadata.FetchAPI() diff: %v", diff)
				}
			}
		})
	}
}

func Test_pmsMetadata_OwnerReference(t *testing.T) {
	tests := []struct {
		name    string
		obj     pmsMetadata
		want    v1.OwnerReference
		wantErr bool
	}{
		{"success", pmsMetadata{Name: "testpod", Namespace: "plex", Uuid: "123"}, v1.OwnerReference{Kind: "Pod", Name: "testpod", UID: "123"}, false},
		{"missing uuid", pmsMetadata{Name: "testpod", Namespace: "plex"}, v1.OwnerReference{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.obj
			got, err := p.OwnerReference()
			if (err != nil) != tt.wantErr {
				t.Errorf("pmsMetadata.OwnerReference() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("pmsMetadata.OwnerReference() = %v, want %v", got, tt.want)
			}
		})
	}
}
