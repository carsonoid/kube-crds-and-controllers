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
	v1alpha1 "github.com/carsonoid/kube-crds-and-controllers/controllers/crd-configured/pkg/apis/podlabeler/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	schema "k8s.io/apimachinery/pkg/runtime/schema"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	testing "k8s.io/client-go/testing"
)

// FakePodLabelConfigs implements PodLabelConfigInterface
type FakePodLabelConfigs struct {
	Fake *FakePodlabelerV1alpha1
	ns   string
}

var podlabelconfigsResource = schema.GroupVersionResource{Group: "podlabeler.k8s.carsonoid.net", Version: "v1alpha1", Resource: "podlabelconfigs"}

var podlabelconfigsKind = schema.GroupVersionKind{Group: "podlabeler.k8s.carsonoid.net", Version: "v1alpha1", Kind: "PodLabelConfig"}

// Get takes name of the podLabelConfig, and returns the corresponding podLabelConfig object, and an error if there is any.
func (c *FakePodLabelConfigs) Get(name string, options v1.GetOptions) (result *v1alpha1.PodLabelConfig, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewGetAction(podlabelconfigsResource, c.ns, name), &v1alpha1.PodLabelConfig{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PodLabelConfig), err
}

// List takes label and field selectors, and returns the list of PodLabelConfigs that match those selectors.
func (c *FakePodLabelConfigs) List(opts v1.ListOptions) (result *v1alpha1.PodLabelConfigList, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewListAction(podlabelconfigsResource, podlabelconfigsKind, c.ns, opts), &v1alpha1.PodLabelConfigList{})

	if obj == nil {
		return nil, err
	}

	label, _, _ := testing.ExtractFromListOptions(opts)
	if label == nil {
		label = labels.Everything()
	}
	list := &v1alpha1.PodLabelConfigList{}
	for _, item := range obj.(*v1alpha1.PodLabelConfigList).Items {
		if label.Matches(labels.Set(item.Labels)) {
			list.Items = append(list.Items, item)
		}
	}
	return list, err
}

// Watch returns a watch.Interface that watches the requested podLabelConfigs.
func (c *FakePodLabelConfigs) Watch(opts v1.ListOptions) (watch.Interface, error) {
	return c.Fake.
		InvokesWatch(testing.NewWatchAction(podlabelconfigsResource, c.ns, opts))

}

// Create takes the representation of a podLabelConfig and creates it.  Returns the server's representation of the podLabelConfig, and an error, if there is any.
func (c *FakePodLabelConfigs) Create(podLabelConfig *v1alpha1.PodLabelConfig) (result *v1alpha1.PodLabelConfig, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewCreateAction(podlabelconfigsResource, c.ns, podLabelConfig), &v1alpha1.PodLabelConfig{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PodLabelConfig), err
}

// Update takes the representation of a podLabelConfig and updates it. Returns the server's representation of the podLabelConfig, and an error, if there is any.
func (c *FakePodLabelConfigs) Update(podLabelConfig *v1alpha1.PodLabelConfig) (result *v1alpha1.PodLabelConfig, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewUpdateAction(podlabelconfigsResource, c.ns, podLabelConfig), &v1alpha1.PodLabelConfig{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PodLabelConfig), err
}

// Delete takes name of the podLabelConfig and deletes it. Returns an error if one occurs.
func (c *FakePodLabelConfigs) Delete(name string, options *v1.DeleteOptions) error {
	_, err := c.Fake.
		Invokes(testing.NewDeleteAction(podlabelconfigsResource, c.ns, name), &v1alpha1.PodLabelConfig{})

	return err
}

// DeleteCollection deletes a collection of objects.
func (c *FakePodLabelConfigs) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	action := testing.NewDeleteCollectionAction(podlabelconfigsResource, c.ns, listOptions)

	_, err := c.Fake.Invokes(action, &v1alpha1.PodLabelConfigList{})
	return err
}

// Patch applies the patch and returns the patched podLabelConfig.
func (c *FakePodLabelConfigs) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.PodLabelConfig, err error) {
	obj, err := c.Fake.
		Invokes(testing.NewPatchSubresourceAction(podlabelconfigsResource, c.ns, name, data, subresources...), &v1alpha1.PodLabelConfig{})

	if obj == nil {
		return nil, err
	}
	return obj.(*v1alpha1.PodLabelConfig), err
}
