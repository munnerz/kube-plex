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

func Test_jobDone(t *testing.T) {
	tests := []struct {
		name    string
		job     *batch.Job
		want    bool
		wantErr bool
	}{
		{"incomplete job", &batch.Job{Status: batch.JobStatus{Active: 1}}, false, false},
		{"successful job", &batch.Job{Status: batch.JobStatus{Succeeded: 1}}, true, false},
		{"failed job", &batch.Job{Status: batch.JobStatus{Failed: 1}}, true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := jobDone(tt.job)
			if (err != nil) != tt.wantErr {
				t.Errorf("jobDone() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("jobDone() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_podWatcher(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tests := []struct {
		name    string
		status  []batch.JobStatus
		wantErr bool
	}{
		{"running to success", []batch.JobStatus{{Active: 1}, {Active: 0, Succeeded: 1}}, false},
		{"running to failure", []batch.JobStatus{{Active: 1}, {Active: 0, Failed: 1}}, true},
		{"long running to success", []batch.JobStatus{{Active: 1}, {Active: 1}, {Active: 0, Succeeded: 1}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := &batch.Job{
				ObjectMeta: metav1.ObjectMeta{Name: "testjob", Namespace: "test"},
			}

			cl := fake.NewSimpleClientset(job)
			w, err := cl.BatchV1().Jobs(job.Namespace).Watch(ctx, metav1.SingleObject(job.ObjectMeta))

			done := make(chan bool)
			go func() {
				err = podWatcher(ctx, w)
				if (err != nil) != tt.wantErr {
					t.Errorf("podWatcher() error = %v, wantErr %v", err, tt.wantErr)
				}
				done <- true
			}()

			for _, s := range tt.status {
				job.Status = s
				cl.BatchV1().Jobs(job.Namespace).UpdateStatus(ctx, job, metav1.UpdateOptions{})
			}

			<-done
		})
	}

	t.Run("job deletion", func(t *testing.T) {
		job := &batch.Job{
			ObjectMeta: metav1.ObjectMeta{Name: "testjob", Namespace: "test"},
		}

		cl := fake.NewSimpleClientset(job)
		w, err := cl.BatchV1().Jobs(job.Namespace).Watch(ctx, metav1.SingleObject(job.ObjectMeta))

		done := make(chan bool)
		go func() {
			err = podWatcher(ctx, w)
			if err == nil {
				t.Errorf("podWatcher() returned success, expected error")
			}
			done <- true
		}()

		cl.BatchV1().Jobs(job.Namespace).Delete(ctx, job.Name, metav1.DeleteOptions{})
		<-done
	})

	t.Run("termination from context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		job := &batch.Job{
			ObjectMeta: metav1.ObjectMeta{Name: "testjob", Namespace: "test"},
		}

		cl := fake.NewSimpleClientset(job)
		w, err := cl.BatchV1().Jobs(job.Namespace).Watch(ctx, metav1.SingleObject(job.ObjectMeta))

		done := make(chan bool)
		go func() {
			err = podWatcher(ctx, w)
			if err == nil {
				t.Errorf("podWatcher() returned success, expected error")
			}
			done <- true
		}()

		cancel()
		<-done
	})

}
