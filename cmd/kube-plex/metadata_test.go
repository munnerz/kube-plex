package main

import (
	"context"
	"reflect"
	"testing"

	"github.com/go-test/deep"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func Test_pmsMetadata_FetchMetadata(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cpuQuantity, _ := resource.ParseQuantity("1")
	validPod := corev1.Pod{
		ObjectMeta: v1.ObjectMeta{
			Namespace: "plex", Name: "pms", UID: "123",
			Annotations: map[string]string{"kube-plex/pms-addr": "service:32400", "kube-plex/mounts": "/data"},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{Name: "plex", Image: "plex:test", VolumeMounts: []corev1.VolumeMount{{Name: "data", MountPath: "/data"}}}},
			Volumes:    []corev1.Volume{{Name: "data"}},
		},
		Status: corev1.PodStatus{
			InitContainerStatuses: []corev1.ContainerStatus{{Name: "kube-plex-init", Image: "kubeplex:latest", ImageID: "kubeplex@sha256:12345"}},
			ContainerStatuses:     []corev1.ContainerStatus{{Name: "plex", Image: "pms:latest", ImageID: "pms@sha256:12345"}},
		},
	}

	tests := []struct {
		name         string
		podname      string
		podnamespace string
		pod          corev1.Pod
		wantPms      PmsMetadata
		wantErr      bool
	}{
		{"fetches info from api", "pms", "plex", validPod, PmsMetadata{
			Name: "pms", Namespace: "plex", UID: "123", PmsImage: "pms@sha256:12345", KubePlexImage: "kubeplex@sha256:12345", PmsAddr: "service:32400",
			Mounts: []string{"/data"}, VolumeMounts: validPod.Spec.Containers[0].VolumeMounts, Volumes: validPod.Spec.Volumes}, false,
		},
		{"fails on missing podname", "", "plex", validPod, PmsMetadata{}, true},
		{"fails on missing namespace", "pms", "", validPod, PmsMetadata{}, true},
		{"fails gracefully on wrong pod name", "wrong", "plex", validPod, PmsMetadata{}, true},
		{"fails gracefully on wrong namespace", "pms", "wrong", validPod, PmsMetadata{}, true},
		{"plex container missing", "pms", "plex", corev1.Pod{ObjectMeta: validPod.ObjectMeta, Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "wrong", Image: "pms:own"}}, Volumes: []corev1.Volume{{Name: "data"}, {Name: "transcode"}}}}, PmsMetadata{}, true},
		{"plex data volume missing", "pms", "plex", corev1.Pod{ObjectMeta: validPod.ObjectMeta, Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "plex", Image: "pms:own"}}, Volumes: []corev1.Volume{{Name: "transcode"}}}}, PmsMetadata{}, true},
		{"plex default volumes", "pms", "plex",
			corev1.Pod{
				ObjectMeta: v1.ObjectMeta{Namespace: "plex", Name: "pms", UID: "123", Annotations: map[string]string{"kube-plex/pms-addr": "service:32400"}},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{Name: "plex", Image: "plex:test", VolumeMounts: []corev1.VolumeMount{{Name: "data", MountPath: "/data"}, {Name: "transcode", MountPath: "/transcode"}}}},
					Volumes:    []corev1.Volume{{Name: "transcode"}, {Name: "data"}}},
				Status: validPod.Status,
			},
			PmsMetadata{
				Name: "pms", Namespace: "plex", UID: "123", PmsImage: "pms@sha256:12345", KubePlexImage: "kubeplex@sha256:12345", PmsAddr: "service:32400",
				Mounts:       []string{"/transcode", "/data"},
				VolumeMounts: []corev1.VolumeMount{{Name: "transcode", MountPath: "/transcode"}, {Name: "data", MountPath: "/data"}},
				Volumes:      []corev1.Volume{{Name: "data"}, {Name: "transcode"}},
			},
			false},
		{"plex transcode volume missing", "pms", "plex", corev1.Pod{ObjectMeta: validPod.ObjectMeta, Spec: corev1.PodSpec{Containers: []corev1.Container{{Name: "plex", Image: "pms:own"}}, Volumes: []corev1.Volume{{Name: "data"}}}}, PmsMetadata{}, true},
		{"kube-plex debug set", "pms", "plex", corev1.Pod{
			ObjectMeta: v1.ObjectMeta{Namespace: "plex", Name: "pms", UID: "123", Annotations: map[string]string{"kube-plex/pms-addr": "a:32400", "kube-plex/loglevel": "debug", "kube-plex/mounts": ""}}, Spec: validPod.Spec, Status: validPod.Status},
			PmsMetadata{Name: "pms", Namespace: "plex", UID: "123", PmsImage: "pms@sha256:12345", KubePlexImage: "kubeplex@sha256:12345", KubePlexLevel: "debug", PmsAddr: "a:32400"},
			false,
		},
		{"renamed kube-plex container", "pms", "plex",
			corev1.Pod{ObjectMeta: v1.ObjectMeta{Namespace: "plex", Name: "pms", UID: "123", Annotations: map[string]string{"kube-plex/container-name": "kp-init", "kube-plex/pms-addr": "a:32400", "kube-plex/mounts": ""}}, Spec: validPod.Spec, Status: corev1.PodStatus{ContainerStatuses: validPod.Status.ContainerStatuses, InitContainerStatuses: []corev1.ContainerStatus{{Name: "kp-init", ImageID: "aaa@sha256:12345"}}}},
			PmsMetadata{Name: "pms", Namespace: "plex", UID: "123", PmsImage: "pms@sha256:12345", KubePlexImage: "aaa@sha256:12345", PmsAddr: "a:32400"},
			false,
		},
		{"renamed PMS container", "pms", "plex",
			corev1.Pod{ObjectMeta: v1.ObjectMeta{Namespace: "plex", Name: "pms", UID: "123", Annotations: map[string]string{"kube-plex/pms-container-name": "test", "kube-plex/pms-addr": "a:32400", "kube-plex/mounts": ""}}, Spec: validPod.Spec, Status: corev1.PodStatus{InitContainerStatuses: validPod.Status.InitContainerStatuses, ContainerStatuses: []corev1.ContainerStatus{{Name: "test", ImageID: "aaa@sha256:12345"}}}},
			PmsMetadata{Name: "pms", Namespace: "plex", UID: "123", PmsImage: "aaa@sha256:12345", KubePlexImage: "kubeplex@sha256:12345", PmsAddr: "a:32400"},
			false,
		},
		{"sets resource definitions", "pms", "plex",
			corev1.Pod{ObjectMeta: v1.ObjectMeta{Namespace: "plex", Name: "pms", UID: "123", Annotations: map[string]string{"kube-plex/pms-addr": "a:32400", "kube-plex/mounts": "", "kube-plex/resources-requests": "{\"cpu\": \"1\"}", "kube-plex/resources-limits": "{\"cpu\": \"1\"}"}}, Spec: validPod.Spec, Status: validPod.Status},
			PmsMetadata{Name: "pms", Namespace: "plex", UID: "123", PmsImage: "pms@sha256:12345", KubePlexImage: "kubeplex@sha256:12345", PmsAddr: "a:32400", ResourceRequests: corev1.ResourceList{corev1.ResourceCPU: cpuQuantity}, ResourceLimits: corev1.ResourceList{corev1.ResourceCPU: cpuQuantity}},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl := fake.NewSimpleClientset(&tt.pod)
			m, err := FetchMetadata(ctx, cl, tt.podname, tt.podnamespace)
			if (err != nil) != tt.wantErr {
				t.Errorf("pmsMetadata.FetchAPI() error = %v, wantErr %v", err, tt.wantErr)
			}
			// We don't want to check output state if error occurs
			if !tt.wantErr {
				if diff := deep.Equal(m, tt.wantPms); diff != nil {
					t.Errorf("pmsMetadata.FetchAPI() diff: %v", diff)
				}
			}
		})
	}
}

func Test_pmsMetadata_OwnerReference(t *testing.T) {
	tests := []struct {
		name    string
		obj     PmsMetadata
		want    v1.OwnerReference
		wantErr bool
	}{
		{"success", PmsMetadata{Name: "testpod", Namespace: "plex", UID: "123"}, v1.OwnerReference{APIVersion: "v1", Kind: "Pod", Name: "testpod", UID: "123"}, false},
		{"missing uuid", PmsMetadata{Name: "testpod", Namespace: "plex"}, v1.OwnerReference{}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.obj
			got, err := p.OwnerReference()
			if (err != nil) != tt.wantErr {
				t.Errorf("pmsMetadata.OwnerReference() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("pmsMetadata.OwnerReference() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_pmsMetadata_LauncherCmd(t *testing.T) {
	tests := []struct {
		name string
		p    PmsMetadata
		args []string
		want []string
	}{
		{"generates bare cmd", PmsMetadata{PmsAddr: "a:32400"}, []string{"a"}, []string{"/shared/transcode-launcher", "--pms-addr=a:32400", "--listen=:32400", "--", "a"}},
		{"generates codec server url", PmsMetadata{PmsAddr: "a:32400", PodIP: "1.2.3.4", CodecPort: 1234}, []string{"a"}, []string{"/shared/transcode-launcher", "--pms-addr=a:32400", "--listen=:32400", "--codec-server-url=http://1.2.3.4:1234/", "--codec-dir=/shared/codecs/", "--", "a"}},
		{"generates debug flag", PmsMetadata{PmsAddr: "a:32400", KubePlexLevel: "debug"}, []string{"a"}, []string{"/shared/transcode-launcher", "--pms-addr=a:32400", "--listen=:32400", "--loglevel=debug", "--", "a"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.p
			if got := p.LauncherCmd(tt.args...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("pmsMetadata.LauncherCmd() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getVolumesAndMounts(t *testing.T) {
	testPod := corev1.Pod{
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Name: "container",
				VolumeMounts: []corev1.VolumeMount{
					{Name: "missing", MountPath: "/missing"},
					{Name: "data", MountPath: "/data"},
					{Name: "data", MountPath: "/data1", SubPath: "s1"},
					{Name: "data", MountPath: "/data2", SubPath: "s2"},
					{Name: "transcode", MountPath: "/transcode"},
					{Name: "full", MountPath: "/full", ReadOnly: true},
				},
			}},
			Volumes: []corev1.Volume{{Name: "data"}, {Name: "transcode"}, {Name: "full", VolumeSource: corev1.VolumeSource{EmptyDir: &corev1.EmptyDirVolumeSource{}}}},
		},
	}
	type args struct {
		dirs []string
		name string
	}
	tests := []struct {
		name       string
		args       args
		wantVolume []corev1.Volume
		wantMount  []corev1.VolumeMount
		wantErr    bool
	}{
		{"default kube-plex mounts",
			args{dirs: []string{"/data", "/transcode"}, name: "container"},
			[]corev1.Volume{{Name: "data"}, {Name: "transcode"}},
			[]corev1.VolumeMount{{Name: "data", MountPath: "/data"}, {Name: "transcode", MountPath: "/transcode"}},
			false,
		},
		{"deduplicate volumes",
			args{dirs: []string{"/data1", "/data2"}, name: "container"},
			[]corev1.Volume{{Name: "data"}},
			[]corev1.VolumeMount{{Name: "data", MountPath: "/data1", SubPath: "s1"}, {Name: "data", MountPath: "/data2", SubPath: "s2"}},
			false,
		},
		{"errors on invalid container", args{dirs: []string{"/data"}, name: "fail"}, nil, nil, true},
		{"errors on invalid path", args{dirs: []string{"/data", "/fail"}, name: "fail"}, nil, nil, true},
		// Test this even if it's an invalid case due to validation in Kubernetes
		{"errors on missing volume", args{dirs: []string{"/missing"}, name: "container"}, nil, nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			volumes, mounts, err := getVolumesAndMounts(tt.args.dirs, &testPod, tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("getVolumesAndMounts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := deep.Equal(volumes, tt.wantVolume); diff != nil {
				t.Errorf("getVolumesAndMounts() volumes don't match diff = %v", diff)
			}
			if diff := deep.Equal(mounts, tt.wantMount); diff != nil {
				t.Errorf("getVolumesAndMounts() mounts don't match diff = %v", diff)
			}
		})
	}
}

func Test_parseResources(t *testing.T) {
	cpuMilli, _ := resource.ParseQuantity("100m")
	qOne, _ := resource.ParseQuantity("1")
	tests := []struct {
		name    string
		t       string
		want    corev1.ResourceList
		wantErr bool
	}{
		{"empty input returns nil", "", nil, false},
		{"handles json input", "{\"cpu\": \"100m\"}", corev1.ResourceList{corev1.ResourceCPU: cpuMilli}, false},
		{"handles arbitrary fields", "{\"intel.gpu\": 1}", corev1.ResourceList{"intel.gpu": qOne}, false},
		{"returns error on broken json", "{\"", nil, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseResourcesJSON(tt.t)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseResources() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseResources() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPmsMetadata_ResourceRequirements(t *testing.T) {
	cpuMilli, _ := resource.ParseQuantity("100m")
	type fields struct {
		ResourceRequests corev1.ResourceList
		ResourceLimits   corev1.ResourceList
	}
	tests := []struct {
		name   string
		fields fields
		want   corev1.ResourceRequirements
	}{
		{"empty object when no requests exist", fields{}, corev1.ResourceRequirements{}},
		{"limit only", fields{ResourceRequests: corev1.ResourceList{corev1.ResourceCPU: cpuMilli}}, corev1.ResourceRequirements{Requests: corev1.ResourceList{corev1.ResourceCPU: cpuMilli}}},
		{"request only", fields{ResourceLimits: corev1.ResourceList{corev1.ResourceCPU: cpuMilli}}, corev1.ResourceRequirements{Limits: corev1.ResourceList{corev1.ResourceCPU: cpuMilli}}},
		{"both requests and limits",
			fields{ResourceLimits: corev1.ResourceList{corev1.ResourceCPU: cpuMilli}, ResourceRequests: corev1.ResourceList{corev1.ResourceCPU: cpuMilli}},
			corev1.ResourceRequirements{Limits: corev1.ResourceList{corev1.ResourceCPU: cpuMilli}, Requests: corev1.ResourceList{corev1.ResourceCPU: cpuMilli}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := PmsMetadata{
				ResourceRequests: tt.fields.ResourceRequests,
				ResourceLimits:   tt.fields.ResourceLimits,
			}
			if got := m.ResourceRequirements(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PmsMetadata.ResourceRequirements() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_getContainerImage(t *testing.T) {
	type args struct {
		annotation string
		defname    string
		pod        *corev1.Pod
		status     []corev1.ContainerStatus
	}
	tests := []struct {
		name      string
		args      args
		wantImage string
		wantName  string
		wantErr   bool
	}{
		{"docker pullable", args{defname: "kube-plex", pod: &corev1.Pod{}, status: []corev1.ContainerStatus{corev1.ContainerStatus{Name: "kube-plex", ImageID: "docker-pullable://a/b@sha256:abc"}}}, "a/b@sha256:abc", "kube-plex", false},
		{"containerd image", args{defname: "kube-plex", pod: &corev1.Pod{}, status: []corev1.ContainerStatus{corev1.ContainerStatus{Name: "kube-plex", ImageID: "a/b@sha256:abc"}}}, "a/b@sha256:abc", "kube-plex", false},
		{"image in annotation", args{annotation: "a", defname: "none", pod: &corev1.Pod{ObjectMeta: v1.ObjectMeta{Annotations: map[string]string{"a": "kubeplex"}}}, status: []corev1.ContainerStatus{corev1.ContainerStatus{Name: "kubeplex", ImageID: "a/b@sha256:abc"}}}, "a/b@sha256:abc", "kubeplex", false},
		{"name mismatch", args{defname: "kube-plex", pod: &corev1.Pod{}, status: []corev1.ContainerStatus{corev1.ContainerStatus{Name: "kubeplex", ImageID: "a/b@sha256:abc"}}}, "", "", true},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotImage, gotName, err := getContainerImage(tt.args.annotation, tt.args.defname, tt.args.pod, tt.args.status)
			if (err != nil) != tt.wantErr {
				t.Errorf("getContainerImage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotImage != tt.wantImage {
				t.Errorf("getContainerImage() got image = %v, want %v", gotImage, tt.wantImage)
			}
			if gotName != tt.wantName {
				t.Errorf("getContainerImage() got name = %v, want %v", gotName, tt.wantName)
			}
		})
	}
}
