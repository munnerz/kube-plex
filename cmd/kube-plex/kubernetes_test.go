package main

import (
	"context"
	"reflect"
	"testing"

	"github.com/go-test/deep"
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

func Test_filterPodEnv(t *testing.T) {
	tests := []struct {
		name string
		in   []corev1.EnvVar
		want []corev1.EnvVar
	}{
		{"filters pod name", []corev1.EnvVar{{Name: "POD_NAME", Value: "pms"}, {Name: "SHELL", Value: "/bin/false"}}, []corev1.EnvVar{{Name: "SHELL", Value: "/bin/false"}}},
		{"filters pod namespace", []corev1.EnvVar{{Name: "SHELL", Value: "/bin/false"}, {Name: "POD_NAMESPACE", Value: "pms"}}, []corev1.EnvVar{{Name: "SHELL", Value: "/bin/false"}}},
		{"filters multiple elements", []corev1.EnvVar{{Name: "POD_NAME", Value: "pms"}, {Name: "SHELL", Value: "/bin/false"}, {Name: "POD_NAMESPACE", Value: "pms"}}, []corev1.EnvVar{{Name: "SHELL", Value: "/bin/false"}}},
		{"nothing to filter", []corev1.EnvVar{{Name: "SHELL", Value: "/bin/false"}}, []corev1.EnvVar{{Name: "SHELL", Value: "/bin/false"}}},
		{"empty vars", []corev1.EnvVar{}, []corev1.EnvVar{}},
		{"filter FFmpeg escaping", []corev1.EnvVar{{Name: "FFMPEG_EXTERNAL_LIBS", Value: "/path\\ to/codec"}}, []corev1.EnvVar{{Name: "FFMPEG_EXTERNAL_LIBS", Value: "/path to/codec"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filterPodEnv(tt.in)
			if diff := deep.Equal(got, tt.want); diff != nil {
				t.Errorf("filterPodEnv() diff = %v", diff)
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

func Test_generateJob(t *testing.T) {
	md := PmsMetadata{
		Name:          "pms",
		Namespace:     "plex",
		UID:           "abc123",
		PmsImage:      "pms:latest",
		PmsAddr:       "kubeplex:32400",
		KubePlexImage: "kubeplex:latest",
		Volumes: []corev1.Volume{
			{Name: "data", VolumeSource: corev1.VolumeSource{PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: "datapvc"}}},
			{Name: "transcode", VolumeSource: corev1.VolumeSource{PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: "transcodepvc"}}},
		},
		VolumeMounts: []corev1.VolumeMount{{Name: "data", MountPath: "/data"}, {Name: "transcode", MountPath: "/transcode"}},
	}
	e := []string{"FOO=bar", "BAR=oof"}
	a := []string{"a", "b", "c"}
	cwd := "/rundir"
	got, err := generateJob(cwd, md, e, a)
	if err != nil {
		t.Fatalf("generateJob() returned error, err=%v", err)
	}
	var backoff int32 = 1
	var ttl int32 = 86400
	want := &batch.Job{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName:    "pms-elastic-transcoder-",
			Namespace:       "plex",
			OwnerReferences: []metav1.OwnerReference{{APIVersion: "v1", UID: "abc123", Name: "pms", Kind: "Pod"}},
		},
		Spec: batch.JobSpec{
			BackoffLimit:            &backoff,
			TTLSecondsAfterFinished: &ttl,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					InitContainers: []corev1.Container{{
						Name:         "kube-plex-init",
						Image:        "kubeplex:latest",
						Command:      []string{"cp", "/transcode-launcher", "/shared/transcode-launcher"},
						VolumeMounts: []corev1.VolumeMount{{Name: "shared", MountPath: "/shared", ReadOnly: false}},
					}},
					NodeSelector:  map[string]string{"beta.kubernetes.io/arch": "amd64"},
					RestartPolicy: corev1.RestartPolicyNever,
					Containers: []corev1.Container{{
						Name:    "plex",
						Command: []string{"/shared/transcode-launcher", "--pms-addr=kubeplex:32400", "--listen=:32400", "--", "a", "b", "c"},
						Image:   "pms:latest",
						Env: []corev1.EnvVar{
							{Name: "FOO", Value: "bar"},
							{Name: "BAR", Value: "oof"},
						},
						WorkingDir: "/rundir",
						VolumeMounts: []corev1.VolumeMount{
							{Name: "shared", MountPath: "/shared", ReadOnly: false},
							{Name: "data", MountPath: "/data", ReadOnly: false},
							{Name: "transcode", MountPath: "/transcode", ReadOnly: false},
						},
					}},
					Volumes: []corev1.Volume{
						{Name: "shared", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}},
						{Name: "data", VolumeSource: corev1.VolumeSource{PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: "datapvc"}}},
						{Name: "transcode", VolumeSource: corev1.VolumeSource{PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{ClaimName: "transcodepvc"}}},
					},
				},
			},
		},
	}
	if diff := deep.Equal(want, got); diff != nil {
		t.Errorf("generateJob() output differs, diff: %v", diff)
	}
}
