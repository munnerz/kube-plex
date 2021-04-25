package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	batch "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
	"k8s.io/klog/v2"
)

func generateJob(cwd string, env []string, args []string) *batch.Job {
	envVars := toCoreV1EnvVar(env)
	var ttl, backoff int32
	ttl = int32((24 * time.Hour).Seconds())
	backoff = 1
	return &batch.Job{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "pms-elastic-transcoder-",
		},
		Spec: batch.JobSpec{
			BackoffLimit:            &backoff,
			TTLSecondsAfterFinished: &ttl,
			Template: corev1.PodTemplateSpec{
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
								{Name: "data", MountPath: "/data", ReadOnly: true},
								{Name: "transcode", MountPath: "/transcode"},
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
							Name: "transcode",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: transcodePVC,
								},
							},
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

func waitForPodCompletion(ctx context.Context, cl kubernetes.Interface, job *batch.Job) error {
	w, err := cl.BatchV1().Jobs(job.Namespace).Watch(ctx, metav1.SingleObject(job.ObjectMeta))
	if err != nil {
		return fmt.Errorf("failed to watch job: %v", err)
	}
	defer w.Stop()

	// Check job state once before starting wait
	j, err := cl.BatchV1().Jobs(job.Namespace).Get(ctx, job.Name, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("unable to fetch job information for checking: %v", err)
	}

	if done, err := jobDone(j); done {
		return err
	}

	return podWatcher(ctx, w)
}

func podWatcher(ctx context.Context, w watch.Interface) error {
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("context cancelled: %v", ctx.Err())
		case r := <-w.ResultChan():
			switch r.Type {
			case watch.Added:
			case watch.Modified:
				j := r.Object.(*batch.Job)

				klog.V(2).Info("received an update")
				if done, err := jobDone(j); done {
					return err
				}
			case watch.Deleted:
				j := r.Object.(*batch.Job)
				klog.Error("Job %s deleted while waiting for it to complete!", j.Name)
				return fmt.Errorf("job %s deleted unexpectedly", j.Name)
			}
		}
	}
}

func jobDone(job *batch.Job) (bool, error) {
	switch {
	case job.Status.Failed > 0:
		return true, fmt.Errorf("job %q failed", job.Name)
	case job.Status.Succeeded > 0:
		return true, nil
	}
	return false, nil
}
