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

package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Rabbitmq) DeepCopyInto(out *Rabbitmq) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Rabbitmq.
func (in *Rabbitmq) DeepCopy() *Rabbitmq {
	if in == nil {
		return nil
	}
	out := new(Rabbitmq)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *Rabbitmq) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RabbitmqAuth) DeepCopyInto(out *RabbitmqAuth) {
	*out = *in
	if in.Config != nil {
		in, out := &in.Config, &out.Config
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RabbitmqAuth.
func (in *RabbitmqAuth) DeepCopy() *RabbitmqAuth {
	if in == nil {
		return nil
	}
	out := new(RabbitmqAuth)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RabbitmqImage) DeepCopyInto(out *RabbitmqImage) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RabbitmqImage.
func (in *RabbitmqImage) DeepCopy() *RabbitmqImage {
	if in == nil {
		return nil
	}
	out := new(RabbitmqImage)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RabbitmqList) DeepCopyInto(out *RabbitmqList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]Rabbitmq, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RabbitmqList.
func (in *RabbitmqList) DeepCopy() *RabbitmqList {
	if in == nil {
		return nil
	}
	out := new(RabbitmqList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *RabbitmqList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RabbitmqManagementPlugin) DeepCopyInto(out *RabbitmqManagementPlugin) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RabbitmqManagementPlugin.
func (in *RabbitmqManagementPlugin) DeepCopy() *RabbitmqManagementPlugin {
	if in == nil {
		return nil
	}
	out := new(RabbitmqManagementPlugin)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RabbitmqPolicy) DeepCopyInto(out *RabbitmqPolicy) {
	*out = *in
	out.Definition = in.Definition
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RabbitmqPolicy.
func (in *RabbitmqPolicy) DeepCopy() *RabbitmqPolicy {
	if in == nil {
		return nil
	}
	out := new(RabbitmqPolicy)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RabbitmqPolicyDefinition) DeepCopyInto(out *RabbitmqPolicyDefinition) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RabbitmqPolicyDefinition.
func (in *RabbitmqPolicyDefinition) DeepCopy() *RabbitmqPolicyDefinition {
	if in == nil {
		return nil
	}
	out := new(RabbitmqPolicyDefinition)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RabbitmqSSL) DeepCopyInto(out *RabbitmqSSL) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RabbitmqSSL.
func (in *RabbitmqSSL) DeepCopy() *RabbitmqSSL {
	if in == nil {
		return nil
	}
	out := new(RabbitmqSSL)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RabbitmqSpec) DeepCopyInto(out *RabbitmqSpec) {
	*out = *in
	out.K8SImage = in.K8SImage
	out.RabbitmqSSL = in.RabbitmqSSL
	in.RabbitmqAuth.DeepCopyInto(&out.RabbitmqAuth)
	if in.K8SENV != nil {
		in, out := &in.K8SENV, &out.K8SENV
		*out = make([]corev1.EnvVar, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	if in.K8SLabels != nil {
		in, out := &in.K8SLabels, &out.K8SLabels
		*out = make([]metav1.LabelSelector, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	out.RabbitmqVolumeSize = in.RabbitmqVolumeSize.DeepCopy()
	if in.RabbitmqPodRequests != nil {
		in, out := &in.RabbitmqPodRequests, &out.RabbitmqPodRequests
		*out = make(corev1.ResourceList, len(*in))
		for key, val := range *in {
			(*out)[key] = val.DeepCopy()
		}
	}
	if in.RabbitmqPodLimits != nil {
		in, out := &in.RabbitmqPodLimits, &out.RabbitmqPodLimits
		*out = make(corev1.ResourceList, len(*in))
		for key, val := range *in {
			(*out)[key] = val.DeepCopy()
		}
	}
	if in.RabbitmqPolicies != nil {
		in, out := &in.RabbitmqPolicies, &out.RabbitmqPolicies
		*out = make([]RabbitmqPolicy, len(*in))
		copy(*out, *in)
	}
	if in.RabbitmqPlugins != nil {
		in, out := &in.RabbitmqPlugins, &out.RabbitmqPlugins
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
	if in.RabbitmqCredentials != nil {
		in, out := &in.RabbitmqCredentials, &out.RabbitmqCredentials
		*out = make(map[string][]byte, len(*in))
		for key, val := range *in {
			var outVal []byte
			if val == nil {
				(*out)[key] = nil
			} else {
				in, out := &val, &outVal
				*out = make([]byte, len(*in))
				copy(*out, *in)
			}
			(*out)[key] = outVal
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RabbitmqSpec.
func (in *RabbitmqSpec) DeepCopy() *RabbitmqSpec {
	if in == nil {
		return nil
	}
	out := new(RabbitmqSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *RabbitmqStatus) DeepCopyInto(out *RabbitmqStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new RabbitmqStatus.
func (in *RabbitmqStatus) DeepCopy() *RabbitmqStatus {
	if in == nil {
		return nil
	}
	out := new(RabbitmqStatus)
	in.DeepCopyInto(out)
	return out
}
