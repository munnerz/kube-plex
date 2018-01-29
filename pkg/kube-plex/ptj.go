package kubeplex

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ptjv1 "github.com/munnerz/kube-plex/pkg/apis/ptj/v1"
	"io"
	"os"
	"os/exec"
)

func GeneratePlexTranscodeJob(args []string, env []string, cwd string) ptjv1.PlexTranscodeJob {
    return ptjv1.PlexTranscodeJob{
        ObjectMeta: metav1.ObjectMeta{
            GenerateName: "plex-transcode-job-",
        },
        Spec: ptjv1.PlexTranscodeJobSpec{
            Args: args,
            Env: env,
            Cwd: cwd,
        },
        Status: ptjv1.PlexTranscodeJobStatus{
            State: ptjv1.PlexTranscodeStateCreated,
        },
    }
}

func GetPlexTranscodeJob(kubeClient *KubeClient, jobname string) (ptj *ptjv1.PlexTranscodeJob, err error) {
	ptjs := kubeClient.KubeplexClient.KubeplexV1().PlexTranscodeJobs(kubeClient.Namespace)

	return ptjs.Get(jobname, metav1.GetOptions{})
}

func CreatePlexTranscodeJob(ptj *ptjv1.PlexTranscodeJob, kubeClient *KubeClient) (*ptjv1.PlexTranscodeJob, error) {
	ptjs := kubeClient.KubeplexClient.KubeplexV1().PlexTranscodeJobs(kubeClient.Namespace)
	return ptjs.Create(ptj)
}

func UpdatePlexTranscodeJob(ptj *ptjv1.PlexTranscodeJob, kubeClient *KubeClient) (*ptjv1.PlexTranscodeJob, error) {
	ptjs := kubeClient.KubeplexClient.KubeplexV1().PlexTranscodeJobs(kubeClient.Namespace)
	return ptjs.Update(ptj)
}

func RunPlexTranscodeJob(ptj *ptjv1.PlexTranscodeJob) (ptjv1.PlexTranscodeJobState, string) {
	args := ptj.Spec.Args[1:len(ptj.Spec.Args)]
	cmd := ptj.Spec.Args[0]

	command := exec.Command(cmd, args...)

	command.Dir = ptj.Spec.Cwd

	stderr, err := command.StderrPipe()
	if err != nil {
		return ptjv1.PlexTranscodeStateFailed, err.Error()
	}

	stdout, err := command.StdoutPipe()
	if err != nil {
		return ptjv1.PlexTranscodeStateFailed, err.Error()
	}

	command.Env = ptj.Spec.Env

	err = command.Start()
	if err != nil {
		return ptjv1.PlexTranscodeStateFailed, err.Error()
	}

	go io.Copy(os.Stderr, stderr)
	go io.Copy(os.Stdout, stdout)

	err = command.Wait()
	if err != nil {
		return ptjv1.PlexTranscodeStateFailed, err.Error()
	}

	return ptjv1.PlexTranscodeStateCompleted, ""
}
