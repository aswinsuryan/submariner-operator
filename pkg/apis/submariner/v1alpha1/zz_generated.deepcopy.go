// +build !ignore_autogenerated

/*
Copyright The Kubernetes Authors.

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

// Code generated by deepcopy-gen. DO NOT EDIT.

package v1alpha1

import (
	v1 "k8s.io/api/apps/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceDiscovery) DeepCopyInto(out *ServiceDiscovery) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceDiscovery.
func (in *ServiceDiscovery) DeepCopy() *ServiceDiscovery {
	if in == nil {
		return nil
	}
	out := new(ServiceDiscovery)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ServiceDiscovery) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceDiscoveryList) DeepCopyInto(out *ServiceDiscoveryList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]ServiceDiscovery, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceDiscoveryList.
func (in *ServiceDiscoveryList) DeepCopy() *ServiceDiscoveryList {
	if in == nil {
		return nil
	}
	out := new(ServiceDiscoveryList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *ServiceDiscoveryList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceDiscoverySpec) DeepCopyInto(out *ServiceDiscoverySpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceDiscoverySpec.
func (in *ServiceDiscoverySpec) DeepCopy() *ServiceDiscoverySpec {
	if in == nil {
		return nil
	}
	out := new(ServiceDiscoverySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *ServiceDiscoveryStatus) DeepCopyInto(out *ServiceDiscoveryStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new ServiceDiscoveryStatus.
func (in *ServiceDiscoveryStatus) DeepCopy() *ServiceDiscoveryStatus {
	if in == nil {
		return nil
	}
	out := new(ServiceDiscoveryStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Submariner) DeepCopyInto(out *Submariner) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Submariner.
func (in *Submariner) DeepCopy() *Submariner {
	if in == nil {
		return nil
	}
	out := new(Submariner)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Submariner) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SubmarinerList) DeepCopyInto(out *SubmarinerList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Submariner, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SubmarinerList.
func (in *SubmarinerList) DeepCopy() *SubmarinerList {
	if in == nil {
		return nil
	}
	out := new(SubmarinerList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *SubmarinerList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SubmarinerSpec) DeepCopyInto(out *SubmarinerSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SubmarinerSpec.
func (in *SubmarinerSpec) DeepCopy() *SubmarinerSpec {
	if in == nil {
		return nil
	}
	out := new(SubmarinerSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *SubmarinerStatus) DeepCopyInto(out *SubmarinerStatus) {
	*out = *in
	if in.EngineDaemonSetStatus != nil {
		in, out := &in.EngineDaemonSetStatus, &out.EngineDaemonSetStatus
		*out = new(v1.DaemonSetStatus)
		(*in).DeepCopyInto(*out)
	}
	if in.RouteAgentDaemonSetStatus != nil {
		in, out := &in.RouteAgentDaemonSetStatus, &out.RouteAgentDaemonSetStatus
		*out = new(v1.DaemonSetStatus)
		(*in).DeepCopyInto(*out)
	}
	if in.GlobalnetDaemonSetStatus != nil {
		in, out := &in.GlobalnetDaemonSetStatus, &out.GlobalnetDaemonSetStatus
		*out = new(v1.DaemonSetStatus)
		(*in).DeepCopyInto(*out)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new SubmarinerStatus.
func (in *SubmarinerStatus) DeepCopy() *SubmarinerStatus {
	if in == nil {
		return nil
	}
	out := new(SubmarinerStatus)
	in.DeepCopyInto(out)
	return out
}
