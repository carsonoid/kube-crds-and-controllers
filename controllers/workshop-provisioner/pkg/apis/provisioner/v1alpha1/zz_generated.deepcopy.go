// +build !ignore_autogenerated

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

// This file was autogenerated by deepcopy-gen. Do not edit it manually!

package v1alpha1

import (
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	conversion "k8s.io/apimachinery/pkg/conversion"
	runtime "k8s.io/apimachinery/pkg/runtime"
	reflect "reflect"
)

func init() {
	SchemeBuilder.Register(RegisterDeepCopies)
}

// RegisterDeepCopies adds deep-copy functions to the given scheme. Public
// to allow building arbitrary schemes.
//
// Deprecated: deepcopy registration will go away when static deepcopy is fully implemented.
func RegisterDeepCopies(scheme *runtime.Scheme) error {
	return scheme.AddGeneratedDeepCopyFuncs(
		conversion.GeneratedDeepCopyFunc{Fn: func(in interface{}, out interface{}, c *conversion.Cloner) error {
			in.(*WorkshopAttendee).DeepCopyInto(out.(*WorkshopAttendee))
			return nil
		}, InType: reflect.TypeOf(&WorkshopAttendee{})},
		conversion.GeneratedDeepCopyFunc{Fn: func(in interface{}, out interface{}, c *conversion.Cloner) error {
			in.(*WorkshopAttendeeList).DeepCopyInto(out.(*WorkshopAttendeeList))
			return nil
		}, InType: reflect.TypeOf(&WorkshopAttendeeList{})},
		conversion.GeneratedDeepCopyFunc{Fn: func(in interface{}, out interface{}, c *conversion.Cloner) error {
			in.(*WorkshopAttendeeSpec).DeepCopyInto(out.(*WorkshopAttendeeSpec))
			return nil
		}, InType: reflect.TypeOf(&WorkshopAttendeeSpec{})},
		conversion.GeneratedDeepCopyFunc{Fn: func(in interface{}, out interface{}, c *conversion.Cloner) error {
			in.(*WorkshopAttendeeStatus).DeepCopyInto(out.(*WorkshopAttendeeStatus))
			return nil
		}, InType: reflect.TypeOf(&WorkshopAttendeeStatus{})},
	)
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WorkshopAttendee) DeepCopyInto(out *WorkshopAttendee) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WorkshopAttendee.
func (in *WorkshopAttendee) DeepCopy() *WorkshopAttendee {
	if in == nil {
		return nil
	}
	out := new(WorkshopAttendee)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *WorkshopAttendee) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	} else {
		return nil
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WorkshopAttendeeList) DeepCopyInto(out *WorkshopAttendeeList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]WorkshopAttendee, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WorkshopAttendeeList.
func (in *WorkshopAttendeeList) DeepCopy() *WorkshopAttendeeList {
	if in == nil {
		return nil
	}
	out := new(WorkshopAttendeeList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *WorkshopAttendeeList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	} else {
		return nil
	}
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WorkshopAttendeeSpec) DeepCopyInto(out *WorkshopAttendeeSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WorkshopAttendeeSpec.
func (in *WorkshopAttendeeSpec) DeepCopy() *WorkshopAttendeeSpec {
	if in == nil {
		return nil
	}
	out := new(WorkshopAttendeeSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *WorkshopAttendeeStatus) DeepCopyInto(out *WorkshopAttendeeStatus) {
	*out = *in
	if in.Children != nil {
		in, out := &in.Children, &out.Children
		*out = make(map[string]v1.Time, len(*in))
		for key, val := range *in {
			(*out)[key] = *val.DeepCopy()
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new WorkshopAttendeeStatus.
func (in *WorkshopAttendeeStatus) DeepCopy() *WorkshopAttendeeStatus {
	if in == nil {
		return nil
	}
	out := new(WorkshopAttendeeStatus)
	in.DeepCopyInto(out)
	return out
}
