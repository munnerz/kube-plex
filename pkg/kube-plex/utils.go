package kubeplex

import (
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"strings"
	"os"
)

var pmsInternalAddress = os.Getenv("PMS_INTERNAL_ADDRESS")

func KubeClient() (cfg *rest.Config, clientset *kubernetes.Clientset, err error) {
	cfg, err = clientcmd.BuildConfigFromFlags("", "/home/user/.secrets/clusters/codesink/auth/kubeconfig")
	if err != nil {
		return
	}

	clientset, err = kubernetes.NewForConfig(cfg)
	return
}

func RewriteArgs(in []string) {
	for i, v := range in {
		switch v {
		case "-progressurl", "-manifest_name", "-segment_list":
			in[i+1] = strings.Replace(in[i+1], "http://127.0.0.1:32400", pmsInternalAddress, 1)
		case "-loglevel", "-loglevel_plex":
			in[i+1] = "debug"
		}
	}
}
