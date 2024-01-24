//go:build !ignore_autogenerated

/*
Copyright 2024.

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

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BgpPeerDiscovery) DeepCopyInto(out *BgpPeerDiscovery) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	in.Status.DeepCopyInto(&out.Status)
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BgpPeerDiscovery.
func (in *BgpPeerDiscovery) DeepCopy() *BgpPeerDiscovery {
	if in == nil {
		return nil
	}
	out := new(BgpPeerDiscovery)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *BgpPeerDiscovery) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BgpPeerDiscoveryList) DeepCopyInto(out *BgpPeerDiscoveryList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]BgpPeerDiscovery, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BgpPeerDiscoveryList.
func (in *BgpPeerDiscoveryList) DeepCopy() *BgpPeerDiscoveryList {
	if in == nil {
		return nil
	}
	out := new(BgpPeerDiscoveryList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *BgpPeerDiscoveryList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BgpPeerDiscoverySpec) DeepCopyInto(out *BgpPeerDiscoverySpec) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BgpPeerDiscoverySpec.
func (in *BgpPeerDiscoverySpec) DeepCopy() *BgpPeerDiscoverySpec {
	if in == nil {
		return nil
	}
	out := new(BgpPeerDiscoverySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *BgpPeerDiscoveryStatus) DeepCopyInto(out *BgpPeerDiscoveryStatus) {
	*out = *in
	if in.DiscoveredPeers != nil {
		in, out := &in.DiscoveredPeers, &out.DiscoveredPeers
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new BgpPeerDiscoveryStatus.
func (in *BgpPeerDiscoveryStatus) DeepCopy() *BgpPeerDiscoveryStatus {
	if in == nil {
		return nil
	}
	out := new(BgpPeerDiscoveryStatus)
	in.DeepCopyInto(out)
	return out
}
