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
	v1alpha1 "github.com/carsonoid/kube-crds-and-controllers/controllers/crd-configured/pkg/apis/podlabeler/v1alpha1"
	scheme "github.com/carsonoid/kube-crds-and-controllers/controllers/crd-configured/pkg/client/clientset/versioned/scheme"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	types "k8s.io/apimachinery/pkg/types"
	watch "k8s.io/apimachinery/pkg/watch"
	rest "k8s.io/client-go/rest"
)

// PodLabelConfigsGetter has a method to return a PodLabelConfigInterface.
// A group's client should implement this interface.
type PodLabelConfigsGetter interface {
	PodLabelConfigs(namespace string) PodLabelConfigInterface
}

// PodLabelConfigInterface has methods to work with PodLabelConfig resources.
type PodLabelConfigInterface interface {
	Create(*v1alpha1.PodLabelConfig) (*v1alpha1.PodLabelConfig, error)
	Update(*v1alpha1.PodLabelConfig) (*v1alpha1.PodLabelConfig, error)
	Delete(name string, options *v1.DeleteOptions) error
	DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error
	Get(name string, options v1.GetOptions) (*v1alpha1.PodLabelConfig, error)
	List(opts v1.ListOptions) (*v1alpha1.PodLabelConfigList, error)
	Watch(opts v1.ListOptions) (watch.Interface, error)
	Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.PodLabelConfig, err error)
	PodLabelConfigExpansion
}

// podLabelConfigs implements PodLabelConfigInterface
type podLabelConfigs struct {
	client rest.Interface
	ns     string
}

// newPodLabelConfigs returns a PodLabelConfigs
func newPodLabelConfigs(c *PodlabelerV1alpha1Client, namespace string) *podLabelConfigs {
	return &podLabelConfigs{
		client: c.RESTClient(),
		ns:     namespace,
	}
}

// Get takes name of the podLabelConfig, and returns the corresponding podLabelConfig object, and an error if there is any.
func (c *podLabelConfigs) Get(name string, options v1.GetOptions) (result *v1alpha1.PodLabelConfig, err error) {
	result = &v1alpha1.PodLabelConfig{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("podlabelconfigs").
		Name(name).
		VersionedParams(&options, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// List takes label and field selectors, and returns the list of PodLabelConfigs that match those selectors.
func (c *podLabelConfigs) List(opts v1.ListOptions) (result *v1alpha1.PodLabelConfigList, err error) {
	result = &v1alpha1.PodLabelConfigList{}
	err = c.client.Get().
		Namespace(c.ns).
		Resource("podlabelconfigs").
		VersionedParams(&opts, scheme.ParameterCodec).
		Do().
		Into(result)
	return
}

// Watch returns a watch.Interface that watches the requested podLabelConfigs.
func (c *podLabelConfigs) Watch(opts v1.ListOptions) (watch.Interface, error) {
	opts.Watch = true
	return c.client.Get().
		Namespace(c.ns).
		Resource("podlabelconfigs").
		VersionedParams(&opts, scheme.ParameterCodec).
		Watch()
}

// Create takes the representation of a podLabelConfig and creates it.  Returns the server's representation of the podLabelConfig, and an error, if there is any.
func (c *podLabelConfigs) Create(podLabelConfig *v1alpha1.PodLabelConfig) (result *v1alpha1.PodLabelConfig, err error) {
	result = &v1alpha1.PodLabelConfig{}
	err = c.client.Post().
		Namespace(c.ns).
		Resource("podlabelconfigs").
		Body(podLabelConfig).
		Do().
		Into(result)
	return
}

// Update takes the representation of a podLabelConfig and updates it. Returns the server's representation of the podLabelConfig, and an error, if there is any.
func (c *podLabelConfigs) Update(podLabelConfig *v1alpha1.PodLabelConfig) (result *v1alpha1.PodLabelConfig, err error) {
	result = &v1alpha1.PodLabelConfig{}
	err = c.client.Put().
		Namespace(c.ns).
		Resource("podlabelconfigs").
		Name(podLabelConfig.Name).
		Body(podLabelConfig).
		Do().
		Into(result)
	return
}

// Delete takes name of the podLabelConfig and deletes it. Returns an error if one occurs.
func (c *podLabelConfigs) Delete(name string, options *v1.DeleteOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("podlabelconfigs").
		Name(name).
		Body(options).
		Do().
		Error()
}

// DeleteCollection deletes a collection of objects.
func (c *podLabelConfigs) DeleteCollection(options *v1.DeleteOptions, listOptions v1.ListOptions) error {
	return c.client.Delete().
		Namespace(c.ns).
		Resource("podlabelconfigs").
		VersionedParams(&listOptions, scheme.ParameterCodec).
		Body(options).
		Do().
		Error()
}

// Patch applies the patch and returns the patched podLabelConfig.
func (c *podLabelConfigs) Patch(name string, pt types.PatchType, data []byte, subresources ...string) (result *v1alpha1.PodLabelConfig, err error) {
	result = &v1alpha1.PodLabelConfig{}
	err = c.client.Patch(pt).
		Namespace(c.ns).
		Resource("podlabelconfigs").
		SubResource(subresources...).
		Name(name).
		Body(data).
		Do().
		Into(result)
	return
}
