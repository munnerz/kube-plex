package main

import (
	"context"
	"io/fs"
	"reflect"
	"testing"
	"testing/fstest"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/fake"
)

func Test_getMetadata(t *testing.T) {
	tests := []struct {
		name    string
		fs      fs.FS
		want    pmsMetadata
		wantErr bool
	}{
		{"reads pms metadata",
			fstest.MapFS{"podname": {Data: []byte("testpod")}, "namespace": {Data: []byte("plex")}},
			pmsMetadata{Name: "testpod", Namespace: "plex"}, false},
		{"handles missing namespace data",
			fstest.MapFS{"podname": {Data: []byte("testpod")}}, pmsMetadata{}, true},
		{"handles empty namespace data",
			fstest.MapFS{"namespace": {}, "podname": {Data: []byte("testpod")}}, pmsMetadata{}, true},
		{"handles missing uuid data",
			fstest.MapFS{"namespace": {Data: []byte("plex")}}, pmsMetadata{}, true},
		{"handles empty uuid data",
			fstest.MapFS{"podname": {}, "namespace": {Data: []byte("plex")}}, pmsMetadata{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getMetadata(tt.fs)
			if (err != nil) != tt.wantErr {
				t.Errorf("getMetadata() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getMetadata() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_pmsMetadata_FetchAPI(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	validPod := corev1.Pod{
		ObjectMeta: v1.ObjectMeta{Namespace: "plex", Name: "pms", UID: "123"},
		Spec:       corev1.PodSpec{Containers: []corev1.Container{{Name: "plex", Image: "plex:test"}}},
	}
	rawPms := pmsMetadata{Name: "pms", Namespace: "plex"}

	type fields struct {
		Name      string
		Namespace string
		Uuid      types.UID
	}
	tests := []struct {
		name    string
		pod     corev1.Pod
		havePms pmsMetadata
		wantPms pmsMetadata
		wantErr bool
	}{
		{"fetches info from api", validPod, rawPms, pmsMetadata{Name: "pms", Namespace: "plex", Uuid: "123", PMSImage: "plex:test"}, false},
		{"plex container missing", corev1.Pod{ObjectMeta: validPod.ObjectMeta, Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "wrong", Image: "pms:own"}}}}, rawPms, rawPms, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.havePms
			cl := fake.NewSimpleClientset(&tt.pod)
			if err := p.FetchAPI(ctx, cl); (err != nil) != tt.wantErr {
				t.Errorf("pmsMetadata.FetchAPI() error = %v, wantErr %v", err, tt.wantErr)
			}
			// We don't want to check output state if error occurs
			if !tt.wantErr && !reflect.DeepEqual(p, tt.wantPms) {
				t.Errorf("pmsMetadata.FetchAPI() = %v, want %v", p, tt.wantPms)
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
