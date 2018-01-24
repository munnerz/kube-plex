/*
Copyright 2018 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	v1 "github.com/munnerz/kube-plex/pkg/apis/ptj/v1"
	scheme "github.com/munnerz/kube-plex/pkg/client/clientset/versioned/scheme"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// PlexTranscodeJobsGetter has a method to return a PlexTranscodeJobInterface.
// A group's client should implement this interface.
type PlexTranscodeJobsGetter interface {
	PlexTranscodeJobs(namespace string) PlexTranscodeJobInterface
}

// PlexTranscodeJobInterface has methods to work with PlexTranscodeJob resources.
type PlexTranscodeJobInterface interface {
	Create(*v1.PlexTranscodeJob) (*v1.PlexTranscodeJob, error)
	Update(*v1.PlexTranscodeJob) (*v1.PlexTranscodeJob, error)
	UpdateStatus(*v1.PlexTranscodeJob) (*v1.PlexTranscodeJob, error)
	Delete(name string, options *meta_v1.DeleteOptions) error
	DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error
	Get(name string, options meta_v1.GetOptions) (*v1.PlexTranscodeJob, error)
	List(opts meta_v1.ListOptions) (*v1.PlexTranscodeJobList, error)
	Watch(opts meta_v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.PlexTranscodeJob, err error)
	PlexTranscodeJobExpansion
}

// plexTranscodeJobs implements PlexTranscodeJobInterface
type plexTranscodeJobs struct {
	client rest.Interface
	ns     string
}

// newPlexTranscodeJobs returns a PlexTranscodeJobs
func newPlexTranscodeJobs(c *KubeplexV1Client, namespace string) *plexTranscodeJobs {
	return &plexTranscodeJobs{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the plexTranscodeJob, and returns the corresponding plexTranscodeJob object, and an error if there is any.
func (c *plexTranscodeJobs) Get(name string, options meta_v1.GetOptions) (result *v1.PlexTranscodeJob, err error) {
	result = &v1.PlexTranscodeJob{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("plextranscodejobs").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of PlexTranscodeJobs that match those selectors.
func (c *plexTranscodeJobs) List(opts meta_v1.ListOptions) (result *v1.PlexTranscodeJobList, err error) {
	result = &v1.PlexTranscodeJobList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("plextranscodejobs").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested plexTranscodeJobs.
func (c *plexTranscodeJobs) Watch(opts meta_v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("plextranscodejobs").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a plexTranscodeJob and creates it.  Returns the server's representation of the plexTranscodeJob, and an error, if there is any.
func (c *plexTranscodeJobs) Create(plexTranscodeJob *v1.PlexTranscodeJob) (result *v1.PlexTranscodeJob, err error) {
	result = &v1.PlexTranscodeJob{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("plextranscodejobs").
		Body(plexTranscodeJob).
		Do().
		Into(result)
	return
}

// Update takes the representation of a plexTranscodeJob and updates it. Returns the server's representation of the plexTranscodeJob, and an error, if there is any.
func (c *plexTranscodeJobs) Update(plexTranscodeJob *v1.PlexTranscodeJob) (result *v1.PlexTranscodeJob, err error) {
	result = &v1.PlexTranscodeJob{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("plextranscodejobs").
		Name(plexTranscodeJob.Name).
		Body(plexTranscodeJob).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *plexTranscodeJobs) UpdateStatus(plexTranscodeJob *v1.PlexTranscodeJob) (result *v1.PlexTranscodeJob, err error) {
	result = &v1.PlexTranscodeJob{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("plextranscodejobs").
		Name(plexTranscodeJob.Name).
		SubResource("status").
		Body(plexTranscodeJob).
		Do().
		Into(result)
	return
}

// Delete takes name of the plexTranscodeJob and deletes it. Returns an error if one occurs.
func (c *plexTranscodeJobs) Delete(name string, options *meta_v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("plextranscodejobs").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *plexTranscodeJobs) DeleteCollection(options *meta_v1.DeleteOptions, listOptions meta_v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("plextranscodejobs").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched plexTranscodeJob.
func (c *plexTranscodeJobs) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1.PlexTranscodeJob, err error) {
	result = &v1.PlexTranscodeJob{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("plextranscodejobs").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
