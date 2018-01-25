package kubeplex

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ptjv1 "github.com/munnerz/kube-plex/pkg/apis/ptj/v1"
	"os/exec"
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

func GetPlexTranscodeJob(kubeClient *KubeClient, jobname string) (ptj *ptjv1.PlexTranscodeJob, err error) {
	ptjs := kubeClient.KubeplexClient.KubeplexV1().PlexTranscodeJobs("kube-plex")

	return ptjs.Get(jobname, metav1.GetOptions{})
}

func CreatePlexTranscodeJob(ptj *ptjv1.PlexTranscodeJob, kubeClient *KubeClient) (*ptjv1.PlexTranscodeJob, error) {
	ptjs := kubeClient.KubeplexClient.KubeplexV1().PlexTranscodeJobs("kube-plex")
	return ptjs.Create(ptj)
}

func UpdatePlexTranscodeJobState(ptj *ptjv1.PlexTranscodeJob, state ptjv1.PlexTranscodeJobState, kubeClient *KubeClient) (*ptjv1.PlexTranscodeJob, error) {
	ptjs := kubeClient.KubeplexClient.KubeplexV1().PlexTranscodeJobs("kube-plex")

	ptj.Status.State = state
	return ptjs.Update(ptj)
}

func RunPlexTranscodeJob(ptj *ptjv1.PlexTranscodeJob) ptjv1.PlexTranscodeJobState {
	args := ptj.Spec.Args[1:len(ptj.Spec.Args)]
	cmd := ptj.Spec.Args[0]

	if exec.Command(cmd, args...).Run() != nil {
		return ptjv1.PlexTranscodeStateFailed
	} else {
		return ptjv1.PlexTranscodeStateCompleted
	}
}
