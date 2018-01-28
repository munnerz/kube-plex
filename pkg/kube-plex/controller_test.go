package kubeplex

import (
	"testing"
	fakekubeplex "github.com/munnerz/kube-plex/pkg/client/clientset/versioned/fake"
	fakekubernetes "k8s.io/client-go/kubernetes/fake"
)

func TestController(t *testing.T) {
	ptj := GeneratePlexTranscodeJob([]string{"/hello", "world"}, []string{})

	kc := KubeClient{
		Clientset: fakekubernetes.NewSimpleClientset(),
		KubeplexClient: fakekubeplex.NewSimpleClientset(&ptj),
	}

	NewController(&kc)
}
