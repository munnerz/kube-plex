package main

import (
	"log"
	"os"
	"strings"
	"time"

	plexv1alpha1 "gitlab.yvatechengineering.com/server/plex-operator/pkg/apis/plex/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/client-go/tools/cache"
)

// data pvc name
var dataPVC = os.Getenv("DATA_PVC")
var dataPVCSubpath = os.Getenv("DATA_PVC_SUBPATH")

// config pvc name
var configPVC = os.Getenv("CONFIG_PVC")
var configPVCSubpath = os.Getenv("CONFIG_PVC_SUBPATH")

// transcode pvc name
var transcodePVC = os.Getenv("TRANSCODE_PVC")
var transcodePVCSubpath = os.Getenv("TRANSCODE_PVC_SUBPATH")

// pms namespace
var namespace = os.Getenv("KUBE_NAMESPACE")
var podName = os.Getenv("POD_NAME")

// image for the plexmediaserver container containing the transcoder. This
// should be set to the same as the 'master' pms server
var pmsImage = os.Getenv("PMS_IMAGE")
var pmsInternalAddress = os.Getenv("PMS_INTERNAL_ADDRESS")

var instance = os.Getenv("KUBEPLEX_ENV")

func main() {
	if err != nil {
		log.Fatalf("Error getting working directory: %s", err)
	}

	cfg, err := clientcmd.BuildConfigFromFlags("", "")
	if err != nil {
		log.Fatalf("Error building kubeconfig: %s", err)
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		log.Fatalf("Error building kubernetes clientset: %s", err)
	}

	switch instance {
	case "executor":
		log.Println("Running executor.")
		err = runExecutor(kubeClient)
	case "worker":
		log.Println("Running worker.")
		err = runWorker(kubeClient)
	}

}

func runExecutor(kubeClient *kubernetes.Clientset) {
	env := os.Environ()
	args := os.Args

	rewriteEnv(env)
	rewriteArgs(args)
	cwd, err := os.Getwd()

	ptj := generatePlexTranscodeJob(cwd, env, args)
	err = kubeClient.Create(context.TODO(), ptj)
	if err != nil {
		log.Error(err, "Failed to create new PlexTranscodeJob", "PlexTranscodeJob.Namespace", ptj.Namespace, "PlexTranscodeJob.Name", ptj.Name)
	}
	// TODO watch and delete when goes to completed state
}

func runWorker(kubeClient *kubernetes.Clientset) {
	watchlist := cache.NewListWatchFromClient(
		kubeClient.CoreV1().RESTClient(),
		"plextranscodejobs",
		corev1.NamespaceAll,
		fields.Everything(),
	)
	_, controller := cache.NewInformer(
		watchlist,
		&plexv1alpha1.PlexTranscodeJob{},
		0, //Duration is int64
		cache.ResourceEventHandlerFuncs{
			AddFunc: func(obj interface{}) {
				log.Println("PlexTranscodeJob added: %s \n", obj)
			},
			DeleteFunc: func(obj interface{}) {
				log.Println("PlexTranscodeJob deleted: %s \n", obj)
			},
			UpdateFunc: func(oldObj, newObj interface{}) {
				log.Println("PlexTranscodeJob changed \n")

				// updated.Status.State, updated.Status.Error = runPlexTranscodeJob(updated)
			},
		},
	)
	stop := make(chan struct{})
	defer close(stop)
	go controller.Run(stop)
	for {
		time.Sleep(time.Second)
	}
}

// rewriteEnv rewrites environment variables to be passed to the transcoder
func rewriteEnv(in []string) {
	// no changes needed
}

func rewriteArgs(in []string) {
	for i, v := range in {
		switch v {
		case "-progressurl", "-manifest_name", "-segment_list":
			in[i+1] = strings.Replace(in[i+1], "http://127.0.0.1:32400", pmsInternalAddress, 1)
		case "-loglevel", "-loglevel_plex":
			in[i+1] = "debug"
		}
	}
}

func runPlexTranscodeJob(ptj *plexv1alpha1.PlexTranscodeJob) (plexv1alpha1.PlexTranscodeJobState, string) {
	args := ptj.Spec.Args[1:len(ptj.Spec.Args)]
	cmd := ptj.Spec.Args[0]

	command := exec.Command(cmd, args...)

	command.Dir = ptj.Spec.Cwd

	stderr, err := command.StderrPipe()
	if err != nil {
		return plexv1alpha1.PlexTranscodeStateFailed, err.Error()
	}

	stdout, err := command.StdoutPipe()
	if err != nil {
		return plexv1alpha1.PlexTranscodeStateFailed, err.Error()
	}

	command.Env = ptj.Spec.Env

	err = command.Start()
	if err != nil {
		return plexv1alpha1.PlexTranscodeStateFailed, err.Error()
	}

	go io.Copy(os.Stderr, stderr)
	go io.Copy(os.Stdout, stdout)

	err = command.Wait()
	if err != nil {
		return plexv1alpha1.PlexTranscodeStateFailed, err.Error()
	}

	return plexv1alpha1.PlexTranscodeStateCompleted, ""
}

func generatePlexTranscodeJob(cwd string, env []string, args []string) *plexv1alpha1.PlexTranscodeJob {
	return &plexv1alpha1.PlexTranscodeJob{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "plex-transcode-job-",
			Namespace: namespace,
		},
		Spec: plexv1alpha1.PlexTranscodeJobSpec{
			Args: args,
			Env: env,
			Cwd: cwd,
		},
		Status: plexv1alpha1.PlexTranscodeJobStatus{
			State: plexv1alpha1.PlexTranscodeStateCreated,
		},
	}
}
