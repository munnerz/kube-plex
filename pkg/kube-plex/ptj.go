package kubeplex

import (
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    ptjv1 "github.com/munnerz/kube-plex/pkg/apis/ptj/v1"
)

func GeneratePlexTranscodeJob(args []string) ptjv1.PlexTranscodeJob {
    return ptjv1.PlexTranscodeJob{
        ObjectMeta: metav1.ObjectMeta{
            GenerateName: "plex-transcode-job-",
        },
        Spec: ptjv1.PlexTranscodeJobSpec{
            Args: args,
        },
        Status: ptjv1.PlexTranscodeJobStatus{
            State: ptjv1.PlexTranscodeStateCreated,
        },
    }
}

func (c Controller) UpdatePlexTranscodeJobState(jobname string, state ptjv1.PlexTranscodeJobState) (job *ptjv1.PlexTranscodeJob, err error) {
    ptjs := c.KubeplexClient.KubeplexV1().PlexTranscodeJobs("kube-plex")

    job, err = ptjs.Get(jobname, metav1.GetOptions{})
    if err != nil {
        return
    }

    job.Status.State = state
    job, err = ptjs.Update(job)
    return
}
