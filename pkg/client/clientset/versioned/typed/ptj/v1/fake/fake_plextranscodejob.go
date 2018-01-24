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

package fake

import (
	ptj_v1 "github.com/munnerz/kube-plex/pkg/apis/ptj/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakePlexTranscodeJobs implements PlexTranscodeJobInterface
type FakePlexTranscodeJobs struct {
	Fake *FakePlextranscodejobsV1
	ns   string
}

var plextranscodejobsResource = schema.GroupVersionResource{Group: "plextranscodejobs.kube-plex.munnerz.github.com", Version: "v1", Resource: "plextranscodejobs"}

var plextranscodejobsKind = schema.GroupVersionKind{Group: "plextranscodejobs.kube-plex.munnerz.github.com", Version: "v1", Kind: "PlexTranscodeJob"}

// Get takes name of the plexTranscodeJob, and returns the corresponding plexTranscodeJob object, and an error if there is any.
func (c *FakePlexTranscodeJobs) Get(name string, options v1.GetOptions) (result *ptj_v1.PlexTranscodeJob, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(plextranscodejobsResource, c.ns, name), &ptj_v1.PlexTranscodeJob{})

	if obj == nil {
		return nil, err
	}
	return obj.(*ptj_v1.PlexTranscodeJob), err
}

// List takes label and field selectors, and returns the list of PlexTranscodeJobs that match those selectors.
func (c *FakePlexTranscodeJobs) List(opts v1.ListOptions) (result *ptj_v1.PlexTranscodeJobList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(plextranscodejobsResource, plextranscodejobsKind, c.ns, opts), &ptj_v1.PlexTranscodeJobList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &ptj_v1.PlexTranscodeJobList{}
	for _, item := range obj.(*ptj_v1.PlexTranscodeJobList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested plexTranscodeJobs.
func (c *FakePlexTranscodeJobs) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(plextranscodejobsResource, c.ns, opts))

}

// Create takes the representation of a plexTranscodeJob and creates it.  Returns the server's representation of the plexTranscodeJob, and an error, if there is any.
func (c *FakePlexTranscodeJobs) Create(plexTranscodeJob *ptj_v1.PlexTranscodeJob) (result *ptj_v1.PlexTranscodeJob, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(plextranscodejobsResource, c.ns, plexTranscodeJob), &ptj_v1.PlexTranscodeJob{})

	if obj == nil {
		return nil, err
	}
	return obj.(*ptj_v1.PlexTranscodeJob), err
}

// Update takes the representation of a plexTranscodeJob and updates it. Returns the server's representation of the plexTranscodeJob, and an error, if there is any.
func (c *FakePlexTranscodeJobs) Update(plexTranscodeJob *ptj_v1.PlexTranscodeJob) (result *ptj_v1.PlexTranscodeJob, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(plextranscodejobsResource, c.ns, plexTranscodeJob), &ptj_v1.PlexTranscodeJob{})

	if obj == nil {
		return nil, err
	}
	return obj.(*ptj_v1.PlexTranscodeJob), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakePlexTranscodeJobs) UpdateStatus(plexTranscodeJob *ptj_v1.PlexTranscodeJob) (*ptj_v1.PlexTranscodeJob, error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateSubresourceAction(plextranscodejobsResource, "status", c.ns, plexTranscodeJob), &ptj_v1.PlexTranscodeJob{})

	if obj == nil {
		return nil, err
	}
	return obj.(*ptj_v1.PlexTranscodeJob), err
}

// Delete takes name of the plexTranscodeJob and deletes it. Returns an error if one occurs.
func (c *FakePlexTranscodeJobs) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(plextranscodejobsResource, c.ns, name), &ptj_v1.PlexTranscodeJob{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakePlexTranscodeJobs) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(plextranscodejobsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &ptj_v1.PlexTranscodeJobList{})
	return err
}

// Patch applies the patch and returns the patched plexTranscodeJob.
func (c *FakePlexTranscodeJobs) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *ptj_v1.PlexTranscodeJob, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(plextranscodejobsResource, c.ns, name, data, subresources...), &ptj_v1.PlexTranscodeJob{})

	if obj == nil {
		return nil, err
	}
	return obj.(*ptj_v1.PlexTranscodeJob), err
}
