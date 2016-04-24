package common

import (
	"k8s.io/kubernetes/pkg/api"
)

type Config struct {
	LogFile string      `group:"config" namespace:"config"`
	Plex    *PlexConfig `group:"plex config" namespace:"plex"`

	Kubernetes *KubernetesConfig `group:"kubernetes executor" namespace:"kubernetes"`
}

type PlexConfig struct {
	URL string

	TranscodeDir string `yaml:"transcodeDir"`
	MediaDir     string `yaml:"mediaDir"`
}

type KubernetesConfig struct {
	ProxyURL    string `yaml:"proxyUrl"`
	Namespace   string
	PodBasename string `yaml:"podBasename"`
	Image       string

	TranscodeVolumeSource api.VolumeSource `yaml:"transcodeVolumeSource"`
	MediaVolumeSource     api.VolumeSource `yaml:"mediaVolumeSource"`
}
