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

import (
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ArrayProperties define an Batch array job.
type ArrayProperties struct {
	// The size of the array job.
	Size *int64 `json:"size,omitempty"`
}

// ArrayPropertiesDetail defines the array properties of a job for observation.
type ArrayPropertiesDetail struct {
	// The job index within the array that's associated with this job. This parameter
	// is returned for array job children.
	Index *int64 `json:"index,omitempty"`

	// The size of the array job. This parameter is returned for parent array jobs.
	Size *int64 `json:"size,omitempty"`

	// A summary of the number of array job children in each available job status.
	// This parameter is returned for parent array jobs.
	StatusSummary map[string]*int64 `json:"statusSummary,omitempty"`
}

// NetworkInterface defines the elastic network interface for a multi-node parallel
// job node for observation.
type NetworkInterface struct {
	// The attachment ID for the network interface.
	AttachmentID *string `json:"attachmentId,omitempty"`

	// The private IPv6 address for the network interface.
	Ipv6Address *string `json:"ipv6Address,omitempty"`

	// The private IPv4 address for the network interface.
	PrivateIpv4Address *string `json:"privateIpv4Address,omitempty"`
}

// AttemptContainerDetail defines the details of a container that's part of a job attempt for observation
type AttemptContainerDetail struct {
	// The Amazon Resource Name (ARN) of the Amazon ECS container instance that
	// hosts the job attempt.
	ContainerInstanceArn *string `json:"containerInstanceArn,omitempty"`

	// The exit code for the job attempt. A non-zero exit code is considered a failure.
	ExitCode *int64 `json:"exitCode,omitempty"`

	// The name of the CloudWatch Logs log stream associated with the container.
	// The log group for Batch jobs is /aws/batch/job. Each container attempt receives
	// a log stream name when they reach the RUNNING status.
	LogStreamName *string `json:"logStreamName,omitempty"`

	// The network interfaces associated with the job attempt.
	NetworkInterfaces []*NetworkInterface `json:"networkInterfaces,omitempty"`

	// A short (255 max characters) human-readable string to provide additional
	// details about a running or stopped container.
	Reason *string `json:"reason,omitempty"`

	// The Amazon Resource Name (ARN) of the Amazon ECS task that's associated with
	// the job attempt. Each container attempt receives a task ARN when they reach
	// the STARTING status.
	TaskArn *string `json:"taskArn,omitempty"`
}

// AttemptDetail defines a job attempt for observation
type AttemptDetail struct {
	// Details about the container in this job attempt.
	Container *AttemptContainerDetail `json:"container,omitempty"`

	// The Unix timestamp (in milliseconds) for when the attempt was started (when
	// the attempt transitioned from the STARTING state to the RUNNING state).
	StartedAt *int64 `json:"startedAt,omitempty"`

	// A short, human-readable string to provide additional details about the current
	// status of the job attempt.
	StatusReason *string `json:"statusReason,omitempty"`

	// The Unix timestamp (in milliseconds) for when the attempt was stopped (when
	// the attempt transitioned from the RUNNING state to a terminal state, such
	// as SUCCEEDED or FAILED).
	StoppedAt *int64 `json:"stoppedAt,omitempty"`
}

// ContainerOverrides define the overrides that should be sent to a container.
type ContainerOverrides struct {
	// The command to send to the container that overrides the default command from
	// the Docker image or the job definition.
	Command []*string `json:"command,omitempty"`

	// The environment variables to send to the container. You can add new environment
	// variables, which are added to the container at launch, or you can override
	// the existing environment variables from the Docker image or the job definition.
	//
	// Environment variables must not start with AWS_BATCH; this naming convention
	// is reserved for variables that are set by the Batch service.
	Environment []*KeyValuePair `json:"environment,omitempty"`

	// The instance type to use for a multi-node parallel job.
	//
	// This parameter isn't applicable to single-node container jobs or jobs that
	// run on Fargate resources, and shouldn't be provided.
	InstanceType *string `json:"instanceType,omitempty"`

	// The type and amount of resources to assign to a container. This overrides
	// the settings in the job definition. The supported resources include GPU,
	// MEMORY, and VCPU.
	ResourceRequirements []*ResourceRequirement `json:"resourceRequirements,omitempty"`
}

// JobDependency defines an Batch job dependency.
type JobDependency struct {
	// The job ID of the Batch job associated with this dependency.
	//
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/batch/manualv1alpha1.Job
	// +crossplane:generate:reference:refFieldName=JobIDRef
	// +crossplane:generate:reference:selectorFieldName=JobIDSelector
	JobID *string `json:"jobId,omitempty"`

	// JobIDRef is a reference to an JobID.
	// +optional
	JobIDRef *xpv1.Reference `json:"jobIdRef,omitempty"`

	// JobIDSelector selects references to an JobID.
	// +optional
	JobIDSelector *xpv1.Selector `json:"jobIdSelector,omitempty"`

	// The type of the job dependency.
	// +kubebuilder:validation:Enum=N_TO_N;SEQUENTIAL
	Type *string `json:"type,omitempty"`
}

// NodePropertyOverride defines any node overrides to a job definition that's used in
// a SubmitJob API operation.
type NodePropertyOverride struct {
	// The overrides that should be sent to a node range.
	ContainerOverrides *ContainerOverrides `json:"containerOverrides,omitempty"`

	// The range of nodes, using node index values, that's used to override. A range
	// of 0:3 indicates nodes with index values of 0 through 3. If the starting
	// range value is omitted (:n), then 0 is used to start the range. If the ending
	// range value is omitted (n:), then the highest possible node index is used
	// to end the range.
	//
	// TargetNodes is a required field
	// +kubebuilder:validation:Required
	TargetNodes string `json:"targetNodes"`
}

// NodeOverrides define any node overrides to a job definition that's used in
// a SubmitJob API operation.
//
// This isn't applicable to jobs that are running on Fargate resources and shouldn't
// be provided; use containerOverrides instead.
type NodeOverrides struct {
	// The node property overrides for the job.
	NodePropertyOverrides []*NodePropertyOverride `json:"nodePropertyOverrides,omitempty"`

	// The number of nodes to use with a multi-node parallel job. This value overrides
	// the number of nodes that are specified in the job definition. To use this
	// override:
	//
	//    * There must be at least one node range in your job definition that has
	//    an open upper boundary (such as : or n:).
	//
	//    * The lower boundary of the node range specified in the job definition
	//    must be fewer than the number of nodes specified in the override.
	//
	//    * The main node index specified in the job definition must be fewer than
	//    the number of nodes specified in the override.
	NumNodes *int64 `json:"numNodes,omitempty"`
}

// JobParameters define the desired state of a Batch Job
type JobParameters struct {
	// Region is which region the Function will be created.
	// +kubebuilder:validation:Required
	Region string `json:"region"`

	// The array properties for the submitted job, such as the size of the array.
	// The array size can be between 2 and 10,000. If you specify array properties
	// for a job, it becomes an array job. For more information, see Array Jobs
	// (https://docs.aws.amazon.com/batch/latest/userguide/array_jobs.html) in the
	// Batch User Guide.
	ArrayProperties *ArrayProperties `json:"arrayProperties,omitempty"`

	// A list of container overrides in the JSON format that specify the name of
	// a container in the specified job definition and the overrides it should receive.
	// You can override the default command for a container, which is specified
	// in the job definition or the Docker image, with a command override. You can
	// also override existing environment variables on a container or add new environment
	// variables to it with an environment override.
	ContainerOverrides *ContainerOverrides `json:"containerOverrides,omitempty"`

	// A list of dependencies for the job. A job can depend upon a maximum of 20
	// jobs. You can specify a SEQUENTIAL type dependency without specifying a job
	// ID for array jobs so that each child array job completes sequentially, starting
	// at index 0. You can also specify an N_TO_N type dependency with a job ID
	// for array jobs. In that case, each index child of this job must wait for
	// the corresponding index child of each dependency to complete before it can
	// begin.
	DependsOn []*JobDependency `json:"dependsOn,omitempty"`

	// The job definition used by this job. This value can be one of name, name:revision,
	// or the Amazon Resource Name (ARN) for the job definition. If name is specified
	// without a revision then the latest active revision is used.
	//
	// JobDefinition is a required field
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/batch/manualv1alpha1.JobDefinition
	// +crossplane:generate:reference:refFieldName=JobDefinitionRef
	// +crossplane:generate:reference:selectorFieldName=JobDefinitionSelector
	JobDefinition string `json:"jobDefinition,omitempty"`

	// JobDefinitionRef is a reference to an JobDefinition.
	// +optional
	JobDefinitionRef *xpv1.Reference `json:"jobDefinitionRef,omitempty"`

	// JobDefinitionSelector selects references to an JobDefinition.
	// +optional
	JobDefinitionSelector *xpv1.Selector `json:"jobDefinitionSelector,omitempty"`

	// The job queue where the job is submitted. You can specify either the name
	// or the Amazon Resource Name (ARN) of the queue.
	//
	// JobQueue is a required field
	// +crossplane:generate:reference:type=github.com/crossplane-contrib/provider-aws/apis/batch/v1alpha1.JobQueue
	// +crossplane:generate:reference:refFieldName=JobQueueRef
	// +crossplane:generate:reference:selectorFieldName=JobQueueSelector
	JobQueue string `json:"jobQueue,omitempty"`

	// JobQueueRef is a reference to an JobQueue.
	// +optional
	JobQueueRef *xpv1.Reference `json:"jobQueueRef,omitempty"`

	// JobQueueSelector selects references to an JobQueue.
	// +optional
	JobQueueSelector *xpv1.Selector `json:"jobQueueSelector,omitempty"`

	// A list of node overrides in JSON format that specify the node range to target
	// and the container overrides for that node range.
	//
	// This parameter isn't applicable to jobs that are running on Fargate resources;
	// use containerOverrides instead.
	NodeOverrides *NodeOverrides `json:"nodeOverrides,omitempty"`

	// Additional parameters passed to the job that replace parameter substitution
	// placeholders that are set in the job definition. Parameters are specified
	// as a key and value pair mapping. Parameters in a SubmitJob request override
	// any corresponding parameter defaults from the job definition.
	Parameters map[string]*string `json:"parameters,omitempty"`

	// Specifies whether to propagate the tags from the job or job definition to
	// the corresponding Amazon ECS task. If no value is specified, the tags aren't
	// propagated. Tags can only be propagated to the tasks during task creation.
	// For tags with the same name, job tags are given priority over job definitions
	// tags. If the total number of combined tags from the job and job definition
	// is over 50, the job is moved to the FAILED state. When specified, this overrides
	// the tag propagation setting in the job definition.
	PropagateTags *bool `json:"propagateTags,omitempty"`

	// The retry strategy to use for failed jobs from this SubmitJob operation.
	// When a retry strategy is specified here, it overrides the retry strategy
	// defined in the job definition.
	RetryStrategy *RetryStrategy `json:"retryStrategy,omitempty"`

	// The tags that you apply to the job request to help you categorize and organize
	// your resources. Each tag consists of a key and an optional value. For more
	// information, see Tagging Amazon Web Services Resources (https://docs.aws.amazon.com/general/latest/gr/aws_tagging.html)
	// in Amazon Web Services General Reference.
	Tags map[string]*string `json:"tags,omitempty"`

	// The timeout configuration for this SubmitJob operation. You can specify a
	// timeout duration after which Batch terminates your jobs if they haven't finished.
	// If a job is terminated due to a timeout, it isn't retried. The minimum value
	// for the timeout is 60 seconds. This configuration overrides any timeout configuration
	// specified in the job definition. For array jobs, child jobs have the same
	// timeout configuration as the parent job. For more information, see Job Timeouts
	// (https://docs.aws.amazon.com/AmazonECS/latest/developerguide/job_timeouts.html)
	// in the Amazon Elastic Container Service Developer Guide.
	Timeout *JobTimeout `json:"timeout,omitempty"`
}

// A JobSpec defines the desired state of a Job.
type JobSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       JobParameters `json:"forProvider"`
}

// JobObservation keeps the state for the external resource
type JobObservation struct {
	// The array properties of the job, if it is an array job.
	ArrayProperties *ArrayPropertiesDetail `json:"arrayProperties,omitempty"`

	// A list of job attempts associated with this job.
	Attempts []*AttemptDetail `json:"attempts,omitempty"`

	// The Unix timestamp (in milliseconds) for when the job was created. For non-array
	// jobs and parent array jobs, this is when the job entered the SUBMITTED state
	// (at the time SubmitJob was called). For array child jobs, this is when the
	// child job was spawned by its parent and entered the PENDING state.
	CreatedAt *int64 `json:"createdAt,omitempty"`

	// The Amazon Resource Name (ARN) of the job.
	JobArn *string `json:"jobArn,omitempty"`

	// The ID for the job.
	JobID *string `json:"jobId,omitempty"`

	// The Unix timestamp (in milliseconds) for when the job was started (when the
	// job transitioned from the STARTING state to the RUNNING state). This parameter
	// isn't provided for child jobs of array jobs or multi-node parallel jobs.
	StartedAt *int64 `json:"startedAt,omitempty"`

	// The current status for the job.
	//
	// If your jobs don't progress to STARTING, see Jobs Stuck in RUNNABLE Status
	// (https://docs.aws.amazon.com/batch/latest/userguide/troubleshooting.html#job_stuck_in_runnable)
	// in the troubleshooting section of the Batch User Guide.
	Status *string `json:"status,omitempty"`

	// A short, human-readable string to provide additional details about the current
	// status of the job.
	StatusReason *string `json:"statusReason,omitempty"`

	// The Unix timestamp (in milliseconds) for when the job was stopped (when the
	// job transitioned from the RUNNING state to a terminal state, such as SUCCEEDED
	// or FAILED).
	StoppedAt *int64 `json:"stoppedAt,omitempty"`
}

// A JobStatus represents the observed state of a Job.
type JobStatus struct {
	xpv1.ResourceStatus `json:",inline"`
	AtProvider          JobObservation `json:"atProvider,omitempty"`
}

// +kubebuilder:object:root=true

// A Job is a managed resource that represents an AWS Batch Job.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type Job struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   JobSpec   `json:"spec"`
	Status JobStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// JobList contains a list of Jobs
type JobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Job `json:"items"`
}
