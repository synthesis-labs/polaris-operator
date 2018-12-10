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
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PolarisBuildPipeline) DeepCopyInto(out *PolarisBuildPipeline) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PolarisBuildPipeline.
func (in *PolarisBuildPipeline) DeepCopy() *PolarisBuildPipeline {
	if in == nil {
		return nil
	}
	out := new(PolarisBuildPipeline)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PolarisBuildPipeline) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PolarisBuildPipelineBuildSpec) DeepCopyInto(out *PolarisBuildPipelineBuildSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PolarisBuildPipelineBuildSpec.
func (in *PolarisBuildPipelineBuildSpec) DeepCopy() *PolarisBuildPipelineBuildSpec {
	if in == nil {
		return nil
	}
	out := new(PolarisBuildPipelineBuildSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PolarisBuildPipelineList) DeepCopyInto(out *PolarisBuildPipelineList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]PolarisBuildPipeline, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PolarisBuildPipelineList.
func (in *PolarisBuildPipelineList) DeepCopy() *PolarisBuildPipelineList {
	if in == nil {
		return nil
	}
	out := new(PolarisBuildPipelineList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PolarisBuildPipelineList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PolarisBuildPipelineSourceSpec) DeepCopyInto(out *PolarisBuildPipelineSourceSpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PolarisBuildPipelineSourceSpec.
func (in *PolarisBuildPipelineSourceSpec) DeepCopy() *PolarisBuildPipelineSourceSpec {
	if in == nil {
		return nil
	}
	out := new(PolarisBuildPipelineSourceSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PolarisBuildPipelineSpec) DeepCopyInto(out *PolarisBuildPipelineSpec) {
	*out = *in
	out.Source = in.Source
	if in.Builds != nil {
		in, out := &in.Builds, &out.Builds
		*out = make([]PolarisBuildPipelineBuildSpec, len(*in))
		copy(*out, *in)
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PolarisBuildPipelineSpec.
func (in *PolarisBuildPipelineSpec) DeepCopy() *PolarisBuildPipelineSpec {
	if in == nil {
		return nil
	}
	out := new(PolarisBuildPipelineSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PolarisBuildPipelineStatus) DeepCopyInto(out *PolarisBuildPipelineStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PolarisBuildPipelineStatus.
func (in *PolarisBuildPipelineStatus) DeepCopy() *PolarisBuildPipelineStatus {
	if in == nil {
		return nil
	}
	out := new(PolarisBuildPipelineStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PolarisContainerRegistry) DeepCopyInto(out *PolarisContainerRegistry) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PolarisContainerRegistry.
func (in *PolarisContainerRegistry) DeepCopy() *PolarisContainerRegistry {
	if in == nil {
		return nil
	}
	out := new(PolarisContainerRegistry)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PolarisContainerRegistry) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PolarisContainerRegistryList) DeepCopyInto(out *PolarisContainerRegistryList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]PolarisContainerRegistry, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PolarisContainerRegistryList.
func (in *PolarisContainerRegistryList) DeepCopy() *PolarisContainerRegistryList {
	if in == nil {
		return nil
	}
	out := new(PolarisContainerRegistryList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PolarisContainerRegistryList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PolarisContainerRegistrySpec) DeepCopyInto(out *PolarisContainerRegistrySpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PolarisContainerRegistrySpec.
func (in *PolarisContainerRegistrySpec) DeepCopy() *PolarisContainerRegistrySpec {
	if in == nil {
		return nil
	}
	out := new(PolarisContainerRegistrySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PolarisContainerRegistryStatus) DeepCopyInto(out *PolarisContainerRegistryStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PolarisContainerRegistryStatus.
func (in *PolarisContainerRegistryStatus) DeepCopy() *PolarisContainerRegistryStatus {
	if in == nil {
		return nil
	}
	out := new(PolarisContainerRegistryStatus)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PolarisSourceRepository) DeepCopyInto(out *PolarisSourceRepository) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	out.Spec = in.Spec
	out.Status = in.Status
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PolarisSourceRepository.
func (in *PolarisSourceRepository) DeepCopy() *PolarisSourceRepository {
	if in == nil {
		return nil
	}
	out := new(PolarisSourceRepository)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PolarisSourceRepository) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PolarisSourceRepositoryList) DeepCopyInto(out *PolarisSourceRepositoryList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	out.ListMeta = in.ListMeta
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]PolarisSourceRepository, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PolarisSourceRepositoryList.
func (in *PolarisSourceRepositoryList) DeepCopy() *PolarisSourceRepositoryList {
	if in == nil {
		return nil
	}
	out := new(PolarisSourceRepositoryList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *PolarisSourceRepositoryList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PolarisSourceRepositorySpec) DeepCopyInto(out *PolarisSourceRepositorySpec) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PolarisSourceRepositorySpec.
func (in *PolarisSourceRepositorySpec) DeepCopy() *PolarisSourceRepositorySpec {
	if in == nil {
		return nil
	}
	out := new(PolarisSourceRepositorySpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *PolarisSourceRepositoryStatus) DeepCopyInto(out *PolarisSourceRepositoryStatus) {
	*out = *in
	return
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new PolarisSourceRepositoryStatus.
func (in *PolarisSourceRepositoryStatus) DeepCopy() *PolarisSourceRepositoryStatus {
	if in == nil {
		return nil
	}
	out := new(PolarisSourceRepositoryStatus)
	in.DeepCopyInto(out)
	return out
}
