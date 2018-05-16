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

package v1alpha1

import (
	v1alpha1 "github.com/carsonoid/kube-crds-and-controllers/controllers/workshop-provisioner/pkg/apis/provisioner/v1alpha1"
	scheme "github.com/carsonoid/kube-crds-and-controllers/controllers/workshop-provisioner/pkg/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// WorkshopAttendeesGetter has a method to return a WorkshopAttendeeInterface.
// A group's client should implement this interface.
type WorkshopAttendeesGetter interface {
	WorkshopAttendees() WorkshopAttendeeInterface
}

// WorkshopAttendeeInterface has methods to work with WorkshopAttendee resources.
type WorkshopAttendeeInterface interface {
	Create(*v1alpha1.WorkshopAttendee) (*v1alpha1.WorkshopAttendee, error)
	Update(*v1alpha1.WorkshopAttendee) (*v1alpha1.WorkshopAttendee, error)
	UpdateStatus(*v1alpha1.WorkshopAttendee) (*v1alpha1.WorkshopAttendee, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.WorkshopAttendee, error)
	List(opts v1.ListOptions) (*v1alpha1.WorkshopAttendeeList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.WorkshopAttendee, err error)
	WorkshopAttendeeExpansion
}

// workshopAttendees implements WorkshopAttendeeInterface
type workshopAttendees struct {
	client rest.Interface
}

// newWorkshopAttendees returns a WorkshopAttendees
func newWorkshopAttendees(c *ProvisionerV1alpha1Client) *workshopAttendees {
	return &workshopAttendees{
		client: c.RESTClient(),
	}
}

// Get takes name of the workshopAttendee, and returns the corresponding workshopAttendee object, and an error if there is any.
func (c *workshopAttendees) Get(name string, options v1.GetOptions) (result *v1alpha1.WorkshopAttendee, err error) {
	result = &v1alpha1.WorkshopAttendee{}
	err = c.client.Get().
		Resource("workshopattendees").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of WorkshopAttendees that match those selectors.
func (c *workshopAttendees) List(opts v1.ListOptions) (result *v1alpha1.WorkshopAttendeeList, err error) {
	result = &v1alpha1.WorkshopAttendeeList{}
	err = c.client.Get().
		Resource("workshopattendees").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested workshopAttendees.
func (c *workshopAttendees) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Resource("workshopattendees").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a workshopAttendee and creates it.  Returns the server's representation of the workshopAttendee, and an error, if there is any.
func (c *workshopAttendees) Create(workshopAttendee *v1alpha1.WorkshopAttendee) (result *v1alpha1.WorkshopAttendee, err error) {
	result = &v1alpha1.WorkshopAttendee{}
	err = c.client.Post().
		Resource("workshopattendees").
		Body(workshopAttendee).
		Do().
		Into(result)
	return
}

// Update takes the representation of a workshopAttendee and updates it. Returns the server's representation of the workshopAttendee, and an error, if there is any.
func (c *workshopAttendees) Update(workshopAttendee *v1alpha1.WorkshopAttendee) (result *v1alpha1.WorkshopAttendee, err error) {
	result = &v1alpha1.WorkshopAttendee{}
	err = c.client.Put().
		Resource("workshopattendees").
		Name(workshopAttendee.Name).
		Body(workshopAttendee).
		Do().
		Into(result)
	return
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().

func (c *workshopAttendees) UpdateStatus(workshopAttendee *v1alpha1.WorkshopAttendee) (result *v1alpha1.WorkshopAttendee, err error) {
	result = &v1alpha1.WorkshopAttendee{}
	err = c.client.Put().
		Resource("workshopattendees").
		Name(workshopAttendee.Name).
		SubResource("status").
		Body(workshopAttendee).
		Do().
		Into(result)
	return
}

// Delete takes name of the workshopAttendee and deletes it. Returns an error if one occurs.
func (c *workshopAttendees) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Resource("workshopattendees").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *workshopAttendees) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Resource("workshopattendees").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched workshopAttendee.
func (c *workshopAttendees) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.WorkshopAttendee, err error) {
	result = &v1alpha1.WorkshopAttendee{}
	err = c.client.Patch(pt).
		Resource("workshopattendees").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
