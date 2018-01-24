package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type PlexTranscodeJob struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`

	Spec   PlexTranscodeJobSpec   `json:"spec"`
	Status PlexTranscodeJobStatus `json:"status,omitempty"`
}

// The PlexTranscodeJobSpec used to describe a Plex Transcode
type PlexTranscodeJobSpec struct {
	// An array of arguments to pass to the real plex transcode binary
	Args []string
}

type PlexTranscodeJobStatus struct {
	// Name of the transcoder pod assigned the transcode job
	Transcoder string
	// The state of the job, one of: CREATED ASSIGNED STARTED FAILED COMPLETED
	State PlexTranscodeJobState
}

type PlexTranscodeJobState string

const (
	PlexTranscodeStateCreated PlexTranscodeJobState = "CREATED"
	PlexTranscodeStateAssigned PlexTranscodeJobState = "ASSIGNED"
	PlexTranscodeStateStarted PlexTranscodeJobState = "STARTED"
	PlexTranscodeStateFailed PlexTranscodeJobState = "FAILED"
	PlexTranscodeStateCompleted PlexTranscodeJobState = "COMPLETED"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

type PlexTranscodeJobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items []PlexTranscodeJob `json:"items"`
}
