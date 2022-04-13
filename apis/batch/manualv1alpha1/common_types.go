/*
Copyright 2022 The Crossplane Authors.

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

package manualv1alpha1

// Shared types for Job, JobDefinition

// KeyValuePair defines a key-value pair object.
type KeyValuePair struct {
	// The name of the key-value pair. For environment variables, this is the name
	// of the environment variable.
	Name *string `json:"name,omitempty"`

	// The value of the key-value pair. For environment variables, this is the value
	// of the environment variable.
	Value *string `json:"value,omitempty"`
}

// ResourceRequirement defines the type and amount of a resource to assign to a container. The supported
// resources include GPU, MEMORY, and VCPU.
type ResourceRequirement struct {
	// The type of resource to assign to a container. The supported resources include
	// GPU, MEMORY, and VCPU.
	//
	// Type is a required field
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=GPU;MEMORY;VCPU
	ResourceType string `json:"resourceType"` // renamed from Type bc json:"type_"

	// The quantity of the specified resource to reserve for the container. The
	// values vary based on the type specified.
	//
	// type="GPU"
	//
	// The number of physical GPUs to reserve for the container. The number of GPUs
	// reserved for all containers in a job shouldn't exceed the number of available
	// GPUs on the compute resource that the job is launched on.
	//
	// GPUs are not available for jobs that are running on Fargate resources.
	//
	// type="MEMORY"
	//
	// The memory hard limit (in MiB) present to the container. This parameter is
	// supported for jobs that are running on EC2 resources. If your container attempts
	// to exceed the memory specified, the container is terminated. This parameter
	// maps to Memory in the Create a container (https://docs.docker.com/engine/api/v1.23/#create-a-container)
	// section of the Docker Remote API (https://docs.docker.com/engine/api/v1.23/)
	// and the --memory option to docker run (https://docs.docker.com/engine/reference/run/).
	// You must specify at least 4 MiB of memory for a job. This is required but
	// can be specified in several places for multi-node parallel (MNP) jobs. It
	// must be specified for each node at least once. This parameter maps to Memory
	// in the Create a container (https://docs.docker.com/engine/api/v1.23/#create-a-container)
	// section of the Docker Remote API (https://docs.docker.com/engine/api/v1.23/)
	// and the --memory option to docker run (https://docs.docker.com/engine/reference/run/).
	//
	// If you're trying to maximize your resource utilization by providing your
	// jobs as much memory as possible for a particular instance type, see Memory
	// Management (https://docs.aws.amazon.com/batch/latest/userguide/memory-management.html)
	// in the Batch User Guide.
	//
	// For jobs that are running on Fargate resources, then value is the hard limit
	// (in MiB), and must match one of the supported values and the VCPU values
	// must be one of the values supported for that memory value.
	//
	// value = 512
	//
	// VCPU = 0.25
	//
	// value = 1024
	//
	// VCPU = 0.25 or 0.5
	//
	// value = 2048
	//
	// VCPU = 0.25, 0.5, or 1
	//
	// value = 3072
	//
	// VCPU = 0.5, or 1
	//
	// value = 4096
	//
	// VCPU = 0.5, 1, or 2
	//
	// value = 5120, 6144, or 7168
	//
	// VCPU = 1 or 2
	//
	// value = 8192
	//
	// VCPU = 1, 2, or 4
	//
	// value = 9216, 10240, 11264, 12288, 13312, 14336, 15360, or 16384
	//
	// VCPU = 2 or 4
	//
	// value = 17408, 18432, 19456, 20480, 21504, 22528, 23552, 24576, 25600, 26624,
	// 27648, 28672, 29696, or 30720
	//
	// VCPU = 4
	//
	// type="VCPU"
	//
	// The number of vCPUs reserved for the container. This parameter maps to CpuShares
	// in the Create a container (https://docs.docker.com/engine/api/v1.23/#create-a-container)
	// section of the Docker Remote API (https://docs.docker.com/engine/api/v1.23/)
	// and the --cpu-shares option to docker run (https://docs.docker.com/engine/reference/run/).
	// Each vCPU is equivalent to 1,024 CPU shares. For EC2 resources, you must
	// specify at least one vCPU. This is required but can be specified in several
	// places; it must be specified for each node at least once.
	//
	// For jobs that are running on Fargate resources, then value must match one
	// of the supported values and the MEMORY values must be one of the values supported
	// for that VCPU value. The supported values are 0.25, 0.5, 1, 2, and 4
	//
	// value = 0.25
	//
	// MEMORY = 512, 1024, or 2048
	//
	// value = 0.5
	//
	// MEMORY = 1024, 2048, 3072, or 4096
	//
	// value = 1
	//
	// MEMORY = 2048, 3072, 4096, 5120, 6144, 7168, or 8192
	//
	// value = 2
	//
	// MEMORY = 4096, 5120, 6144, 7168, 8192, 9216, 10240, 11264, 12288, 13312,
	// 14336, 15360, or 16384
	//
	// value = 4
	//
	// MEMORY = 8192, 9216, 10240, 11264, 12288, 13312, 14336, 15360, 16384, 17408,
	// 18432, 19456, 20480, 21504, 22528, 23552, 24576, 25600, 26624, 27648, 28672,
	// 29696, or 30720
	//
	// Value is a required field
	// +kubebuilder:validation:Required
	Value string `json:"value"`
}

// EvaluateOnExit specifies a set of conditions to be met, and an action to take (RETRY or
// EXIT) if all conditions are met.
type EvaluateOnExit struct {
	// Specifies the action to take if all of the specified conditions (onStatusReason,
	// onReason, and onExitCode) are met. The values aren't case sensitive.
	// (AWS gives lowercase back!)
	// Action is a required field
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=retry;exit
	Action string `json:"action"`

	// Contains a glob pattern to match against the decimal representation of the
	// ExitCode returned for a job. The pattern can be up to 512 characters in length.
	// It can contain only numbers, and can optionally end with an asterisk (*)
	// so that only the start of the string needs to be an exact match.
	OnExitCode *string `json:"onExitCode,omitempty"`

	// Contains a glob pattern to match against the Reason returned for a job. The
	// pattern can be up to 512 characters in length. It can contain letters, numbers,
	// periods (.), colons (:), and white space (including spaces and tabs). It
	// can optionally end with an asterisk (*) so that only the start of the string
	// needs to be an exact match.
	OnReason *string `json:"onReason,omitempty"`

	// Contains a glob pattern to match against the StatusReason returned for a
	// job. The pattern can be up to 512 characters in length. It can contain letters,
	// numbers, periods (.), colons (:), and white space (including spaces or tabs).
	// It can optionally end with an asterisk (*) so that only the start of the
	// string needs to be an exact match.
	OnStatusReason *string `json:"onStatusReason,omitempty"`
}

// RetryStrategy defines the retry strategy associated with a job. For more information, see Automated
// job retries (https://docs.aws.amazon.com/batch/latest/userguide/job_retries.html)
// in the Batch User Guide.
type RetryStrategy struct {
	// The number of times to move a job to the RUNNABLE status. You can specify
	// between 1 and 10 attempts. If the value of attempts is greater than one,
	// the job is retried on failure the same number of attempts as the value.
	Attempts *int64 `json:"attempts,omitempty"`

	// Array of up to 5 objects that specify conditions under which the job should
	// be retried or failed. If this parameter is specified, then the attempts parameter
	// must also be specified.
	EvaluateOnExit []*EvaluateOnExit `json:"evaluateOnExit,omitempty"`
}

// JobTimeout defines an object representing a job timeout configuration.
type JobTimeout struct {
	// The time duration in seconds (measured from the job attempt's startedAt timestamp)
	// after which Batch terminates your jobs if they have not finished. The minimum
	// value for the timeout is 60 seconds.
	AttemptDurationSeconds *int64 `json:"attemptDurationSeconds,omitempty"`
}
