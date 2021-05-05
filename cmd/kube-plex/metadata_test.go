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
		{"fetches info from api", "pms", "plex", validPod, pmsMetadata{Name: "pms", Namespace: "plex", UID: "123", PmsImage: "plex:test", KubePlexImage: "kubeplex:latest", PmsURL: "http://service:32400/", Volumes: []corev1.Volume{{Name: "data"}, {Name: "transcode"}}}, false},
		{"fails on missing podname", "", "plex", validPod, pmsMetadata{}, true},
		{"fails on missing namespace", "pms", "", validPod, pmsMetadata{}, true},
		{"fails gracefully on wrong pod name", "wrong", "plex", validPod, pmsMetadata{}, true},
		{"fails gracefully on wrong namespace", "pms", "wrong", validPod, pmsMetadata{}, true},
		{"plex container missing", "pms", "plex", corev1.Pod{ObjectMeta: validPod.ObjectMeta, Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "wrong", Image: "pms:own"}}, Volumes: []corev1.Volume{{Name: "data"}, {Name: "transcode"}}}}, pmsMetadata{}, true},
		{"plex data volume missing", "pms", "plex", corev1.Pod{ObjectMeta: validPod.ObjectMeta, Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "plex", Image: "pms:own"}}, Volumes: []corev1.Volume{{Name: "transcode"}}}}, pmsMetadata{}, true},
		{"plex transcode volume missing", "pms", "plex", corev1.Pod{ObjectMeta: validPod.ObjectMeta, Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "plex", Image: "pms:own"}}, Volumes: []corev1.Volume{{Name: "data"}}}}, pmsMetadata{}, true},
		{"plex service annotation missing", "pms", "plex", corev1.Pod{ObjectMeta: v1.ObjectMeta{Namespace: "plex", Name: "pms", UID: "123", Annotations: map[string]string{"kube-plex/image": "kp:latest"}}, Spec: validPod.Spec}, pmsMetadata{}, true},
		{"kube-plex image annotation missing", "pms", "plex", corev1.Pod{ObjectMeta: v1.ObjectMeta{Namespace: "plex", Name: "pms", UID: "123", Annotations: map[string]string{"kube-plex/pms-url": "http://p/"}}, Spec: validPod.Spec}, pmsMetadata{}, true},
		{"kube-plex debug set", "pms", "plex", corev1.Pod{ObjectMeta: v1.ObjectMeta{Namespace: "plex", Name: "pms", UID: "123", Annotations: map[string]string{"kube-plex/pms-url": "http://a/", "kube-plex/image": "kp:latest", "kube-plex/loglevel": "debug"}}, Spec: validPod.Spec}, pmsMetadata{Name: "pms", Namespace: "plex", UID: "123", PmsImage: "plex:test", KubePlexImage: "kp:latest", KubePlexLevel: "debug", PmsURL: "http://a/", Volumes: []corev1.Volume{{Name: "data"}, {Name: "transcode"}}}, false},
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
		{"success", pmsMetadata{Name: "testpod", Namespace: "plex", UID: "123"}, v1.OwnerReference{APIVersion: "v1", Kind: "Pod", Name: "testpod", UID: "123"}, false},
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

func Test_pmsMetadata_LauncherCmd(t *testing.T) {
	tests := []struct {
		name string
		p    pmsMetadata
		args []string
		want []string
	}{
		{"generates bare cmd", pmsMetadata{PmsURL: "http://a/"}, []string{"a"}, []string{"/shared/transcode-launcher", "--pms-url=http://a/", "--port=32400", "--", "a"}},
		{"generates codec server url", pmsMetadata{PmsURL: "http://a/", PodIP: "1.2.3.4", CodecPort: 1234}, []string{"a"}, []string{"/shared/transcode-launcher", "--pms-url=http://a/", "--port=32400", "--codec-server-url=http://1.2.3.4:1234/", "--codec-dir=/shared/codecs", "--", "a"}},
		{"generates debug flag", pmsMetadata{PmsURL: "http://a/", KubePlexLevel: "debug"}, []string{"a"}, []string{"/shared/transcode-launcher", "--pms-url=http://a/", "--port=32400", "--loglevel=debug", "--", "a"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.p
			if got := p.LauncherCmd(tt.args...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("pmsMetadata.LauncherCmd() = %v, want %v", got, tt.want)
			}
		})
	}
}
