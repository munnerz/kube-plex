package kubeplex

import (
	"testing"
	fakeclientset "github.com/munnerz/kube-plex/pkg/client/clientset/versioned/fake"
	fakekubernetes "k8s.io/client-go/kubernetes/fake"
)

func TestController(t *testing.T) {
	ptj := GeneratePlexTranscodeJob([]string{"/hello", "world"})

	kc := KubeClient{
		Clientset: fakeclientset.NewSimpleClientset(ptj),
		KubeplexClient: fakekubernetes.NewSimpleClientset(),
	}

	NewController(&kc)
}
