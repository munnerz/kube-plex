package kubeplex

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ptjv1 "github.com/munnerz/kube-plex/pkg/apis/ptj/v1"
	"bytes"
	"os/exec"
)

func GeneratePlexTranscodeJob(args []string, env []string) ptjv1.PlexTranscodeJob {
    return ptjv1.PlexTranscodeJob{
        ObjectMeta: metav1.ObjectMeta{
            GenerateName: "plex-transcode-job-",
        },
        Spec: ptjv1.PlexTranscodeJobSpec{
            Args: args,
            Env: env,
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

func UpdatePlexTranscodeJob(ptj *ptjv1.PlexTranscodeJob, kubeClient *KubeClient) (*ptjv1.PlexTranscodeJob, error) {
	ptjs := kubeClient.KubeplexClient.KubeplexV1().PlexTranscodeJobs("kube-plex")
	return ptjs.Update(ptj)
}

func RunPlexTranscodeJob(ptj *ptjv1.PlexTranscodeJob) (ptjv1.PlexTranscodeJobState, string) {
	args := ptj.Spec.Args[1:len(ptj.Spec.Args)]
	cmd := ptj.Spec.Args[0]

	command := exec.Command(cmd, args...)

	var stderr bytes.Buffer
	command.Stderr = &stderr
	command.Env = ptj.Spec.Env

	if command.Run() != nil {
		return ptjv1.PlexTranscodeStateFailed, stderr.String()
	} else {
		return ptjv1.PlexTranscodeStateCompleted, ""
	}
}
