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
	v1alpha1 "github.com/carsonoid/kube-crds-and-controllers/controllers/workshop-provisioner/pkg/apis/provisioner/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakeWorkshopAttendees implements WorkshopAttendeeInterface
type FakeWorkshopAttendees struct {
	Fake *FakeProvisionerV1alpha1
}

var workshopattendeesResource = schema.GroupVersionResource{Group: "provisioner.k8s.carsonoid.net", Version: "v1alpha1", Resource: "workshopattendees"}

var workshopattendeesKind = schema.GroupVersionKind{Group: "provisioner.k8s.carsonoid.net", Version: "v1alpha1", Kind: "WorkshopAttendee"}

// Get takes name of the workshopAttendee, and returns the corresponding workshopAttendee object, and an error if there is any.
func (c *FakeWorkshopAttendees) Get(name string, options v1.GetOptions) (result *v1alpha1.WorkshopAttendee, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootGetAction(workshopattendeesResource, name), &v1alpha1.WorkshopAttendee{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.WorkshopAttendee), err
}

// List takes label and field selectors, and returns the list of WorkshopAttendees that match those selectors.
func (c *FakeWorkshopAttendees) List(opts v1.ListOptions) (result *v1alpha1.WorkshopAttendeeList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootListAction(workshopattendeesResource, workshopattendeesKind, opts), &v1alpha1.WorkshopAttendeeList{})
	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.WorkshopAttendeeList{}
	for _, item := range obj.(*v1alpha1.WorkshopAttendeeList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested workshopAttendees.
func (c *FakeWorkshopAttendees) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewRootWatchAction(workshopattendeesResource, opts))
}

// Create takes the representation of a workshopAttendee and creates it.  Returns the server's representation of the workshopAttendee, and an error, if there is any.
func (c *FakeWorkshopAttendees) Create(workshopAttendee *v1alpha1.WorkshopAttendee) (result *v1alpha1.WorkshopAttendee, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootCreateAction(workshopattendeesResource, workshopAttendee), &v1alpha1.WorkshopAttendee{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.WorkshopAttendee), err
}

// Update takes the representation of a workshopAttendee and updates it. Returns the server's representation of the workshopAttendee, and an error, if there is any.
func (c *FakeWorkshopAttendees) Update(workshopAttendee *v1alpha1.WorkshopAttendee) (result *v1alpha1.WorkshopAttendee, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateAction(workshopattendeesResource, workshopAttendee), &v1alpha1.WorkshopAttendee{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.WorkshopAttendee), err
}

// UpdateStatus was generated because the type contains a Status member.
// Add a +genclient:noStatus comment above the type to avoid generating UpdateStatus().
func (c *FakeWorkshopAttendees) UpdateStatus(workshopAttendee *v1alpha1.WorkshopAttendee) (*v1alpha1.WorkshopAttendee, error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootUpdateSubresourceAction(workshopattendeesResource, "status", workshopAttendee), &v1alpha1.WorkshopAttendee{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.WorkshopAttendee), err
}

// Delete takes name of the workshopAttendee and deletes it. Returns an error if one occurs.
func (c *FakeWorkshopAttendees) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewRootDeleteAction(workshopattendeesResource, name), &v1alpha1.WorkshopAttendee{})
	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakeWorkshopAttendees) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewRootDeleteCollectionAction(workshopattendeesResource, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.WorkshopAttendeeList{})
	return err
}

// Patch applies the patch and returns the patched workshopAttendee.
func (c *FakeWorkshopAttendees) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.WorkshopAttendee, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewRootPatchSubresourceAction(workshopattendeesResource, name, data, subresources...), &v1alpha1.WorkshopAttendee{})
	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.WorkshopAttendee), err
}
