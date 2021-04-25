package main

import (
	"context"
	"reflect"
	"testing"

	batch "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func Test_rewriter_Args(t *testing.T) {
	r := rewriter{
		pmsInternalAddress: "http://test-svc:32400",
	}
	tests := []struct {
		name      string
		args      []string
		want      []string
		ischanged bool
	}{
		{"unmodified args", []string{"-args", "arg1"}, []string{"-args", "arg1"}, false},
		{"modifies loglevel", []string{"-s", "1", "-loglevel", "info"}, []string{"-s", "1", "-loglevel", "debug"}, true},
		{"modifies plex loglevel", []string{"-s", "1", "-loglevel_plex", "info"}, []string{"-s", "1", "-loglevel_plex", "debug"}, true},
		{"server address is adjusted", []string{"-progressurl", "http://127.0.0.1:32400/", "-l", "i"}, []string{"-progressurl", "http://test-svc:32400/", "-l", "i"}, true},
		{"manifest url is adjusted", []string{"-manifest_name", "http://127.0.0.1:32400/manifest", "-l", "i"}, []string{"-manifest_name", "http://test-svc:32400/manifest", "-l", "i"}, true},
		{"segment list url is adjusted", []string{"-segment_list", "http://127.0.0.1:32400/segments/url", "-l", "i"}, []string{"-segment_list", "http://test-svc:32400/segments/url", "-l", "i"}, true},
		{"adjusts multiple arguments", []string{"-progressurl", "http://127.0.0.1:32400/", "-loglevel", "x"}, []string{"-progressurl", "http://test-svc:32400/", "-loglevel", "debug"}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orig := make([]string, len(tt.args))
			copy(orig, tt.args)
			if got := r.Args(tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("rewriter.Args() = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(orig, tt.args) {
				t.Errorf("rewriter.Args() modifies input, orig: %v, after: %v", orig, tt.args)
			}
		})
	}
}

func Test_toCoreV1EnvVar(t *testing.T) {
	tests := []struct {
		name string
		args []string
		want []corev1.EnvVar
	}{
		{"handles empty environment", []string{}, []corev1.EnvVar{}},
		{"splits single entries correctly", []string{"ENV=val"}, []corev1.EnvVar{{Name: "ENV", Value: "val"}}},
		{"splits multiple entries", []string{"ENV=val", "ENV2=val2"}, []corev1.EnvVar{{Name: "ENV", Value: "val"}, {Name: "ENV2", Value: "val2"}}},
		{"handles `=` in values", []string{"ENV=val=ue"}, []corev1.EnvVar{{Name: "ENV", Value: "val=ue"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := toCoreV1EnvVar(tt.args); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("toCoreV1EnvVar() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_waitForPodCompletion(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tests := []struct {
		name    string
		job     *batch.Job
		wantErr bool
	}{
		{"successful run", &batch.Job{ObjectMeta: metav1.ObjectMeta{Name: "job", Namespace: "plex"}, Status: batch.JobStatus{Succeeded: 1}}, false},
		{"failed job", &batch.Job{ObjectMeta: metav1.ObjectMeta{Name: "job", Namespace: "plex"}, Status: batch.JobStatus{Failed: 1}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl := fake.NewSimpleClientset(tt.job)
			if err := waitForPodCompletion(ctx, cl, tt.job); (err != nil) != tt.wantErr {
				t.Errorf("waitForPodCompletion() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
