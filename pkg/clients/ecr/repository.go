package ecr

import (
	"encoding/json"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane/provider-aws/apis/ecr/v1alpha1"
	awsclients "github.com/crossplane/provider-aws/pkg/clients"
)

const (
	// RepositoryAlreadyExists repository already exists
	RepositoryAlreadyExists = "RepositoryAlreadyExistsException"
	// LimitExceededException A service limit is exceeded
	LimitExceededException = "LimitExceededException"
	// RepositoryNotEmptyException ECR is not empty
	RepositoryNotEmptyException = "RepositoryNotEmptyException"
	// RepositoryNotFoundException ECR was not found
	RepositoryNotFoundException = "RepositoryNotFoundException"
)

// RepositoryClient is the external client used for ECR Custom Resource
type RepositoryClient interface {
	CreateRepositoryRequest(input *ecr.CreateRepositoryInput) ecr.CreateRepositoryRequest
	DescribeRepositoriesRequest(input *ecr.DescribeRepositoriesInput) ecr.DescribeRepositoriesRequest
	DeleteRepositoryRequest(input *ecr.DeleteRepositoryInput) ecr.DeleteRepositoryRequest
	ListTagsForResourceRequest(*ecr.ListTagsForResourceInput) ecr.ListTagsForResourceRequest
	TagResourceRequest(*ecr.TagResourceInput) ecr.TagResourceRequest
	PutImageTagMutabilityRequest(*ecr.PutImageTagMutabilityInput) ecr.PutImageTagMutabilityRequest
	PutImageScanningConfigurationRequest(*ecr.PutImageScanningConfigurationInput) ecr.PutImageScanningConfigurationRequest
	UntagResourceRequest(*ecr.UntagResourceInput) ecr.UntagResourceRequest
}

// GenerateRepositoryObservation is used to produce v1alpha1.RepositoryObservation from
// ecr.Repository
func GenerateRepositoryObservation(repo ecr.Repository) v1alpha1.RepositoryObservation {
	o := v1alpha1.RepositoryObservation{
		RegistryID:     aws.StringValue(repo.RegistryId),
		RepositoryArn:  aws.StringValue(repo.RepositoryArn),
		RepositoryName: aws.StringValue(repo.RepositoryName),
		RepositoryURI:  aws.StringValue(repo.RepositoryUri),
	}

	if repo.CreatedAt != nil {
		o.CreatedAt = &metav1.Time{Time: *repo.CreatedAt}
	}
	return o
}

// LateInitializeRepository fills the empty fields in *v1alpha1.RepositoryParameters with
// the values seen in ecr.Repository.
func LateInitializeRepository(in *v1alpha1.RepositoryParameters, r *ecr.Repository) { // nolint:gocyclo
	if r == nil {
		return
	}
	if r.ImageScanningConfiguration != nil && in.ImageScanningConfiguration == nil {
		scanConfig := v1alpha1.ImageScanningConfiguration{
			ScanOnPush: aws.BoolValue(r.ImageScanningConfiguration.ScanOnPush),
		}
		in.ImageScanningConfiguration = &scanConfig
	}
	in.ImageTagMutability = awsclients.LateInitializeStringPtr(in.ImageTagMutability, aws.String(string(r.ImageTagMutability)))
}

// CreatePatch creates a *v1alpha1.RepositoryParameters that has only the changed
// values between the target *v1alpha1.RepositoryParameters and the current
// *ecr.Repository.
func CreatePatch(in *ecr.Repository, target *v1alpha1.RepositoryParameters) (*v1alpha1.RepositoryParameters, error) {
	currentParams := &v1alpha1.RepositoryParameters{}
	LateInitializeRepository(currentParams, in)

	jsonPatch, err := awsclients.CreateJSONPatch(currentParams, target)
	if err != nil {
		return nil, err
	}
	patch := &v1alpha1.RepositoryParameters{}
	if err := json.Unmarshal(jsonPatch, patch); err != nil {
		return nil, err
	}
	return patch, nil
}

// IsRepositoryUpToDate checks whether there is a change in any of the modifiable fields.
func IsRepositoryUpToDate(e *v1alpha1.RepositoryParameters, tags []ecr.Tag, repo *ecr.Repository) bool {
	switch {
	case e.ImageScanningConfiguration != nil && repo.ImageScanningConfiguration != nil:
		if e.ImageScanningConfiguration.ScanOnPush != aws.BoolValue(repo.ImageScanningConfiguration.ScanOnPush) {
			return false
		}
	case e.ImageScanningConfiguration != nil && repo.ImageScanningConfiguration == nil:
		return false
	case e.ImageScanningConfiguration == nil && repo.ImageScanningConfiguration != nil:
		return false
	}
	return strings.EqualFold(aws.StringValue(e.ImageTagMutability), string(repo.ImageTagMutability)) &&
		CompareTags(e.Tags, tags)
}

// IsRepoNotFoundErr returns true if the error is because the item doesn't exist
func IsRepoNotFoundErr(err error) bool {
	if awsErr, ok := err.(awserr.Error); ok {
		if awsErr.Code() == RepositoryNotFoundException {
			return true
		}
	}
	return false
}

// GenerateCreateRepositoryInput Generates the CreateRepositoryInput from the RepositoryParameters
func GenerateCreateRepositoryInput(name string, params *v1alpha1.RepositoryParameters) *ecr.CreateRepositoryInput {
	c := &ecr.CreateRepositoryInput{
		RepositoryName:     awsclients.String(name),
		ImageTagMutability: ecr.ImageTagMutability(aws.StringValue(params.ImageTagMutability)),
	}
	if params.ImageScanningConfiguration != nil {
		scanConfig := ecr.ImageScanningConfiguration{
			ScanOnPush: awsclients.Bool(params.ImageScanningConfiguration.ScanOnPush),
		}
		c.ImageScanningConfiguration = &scanConfig
	}
	return c
}

// CompareTags compares arrays of v1alpha1.Tag and ecr.Tag
func CompareTags(tags []v1alpha1.Tag, ecrTags []ecr.Tag) bool {
	if len(tags) != len(ecrTags) {
		return false
	}

	SortTags(tags, ecrTags)

	for i, t := range tags {
		if t.Key != aws.StringValue(ecrTags[i].Key) || t.Value != aws.StringValue(ecrTags[i].Value) {
			return false
		}
	}

	return true
}

// SortTags sorts array of v1alpha1.Tag and ecr.Tag on 'Key'
func SortTags(tags []v1alpha1.Tag, ecrTags []ecr.Tag) {
	sort.Slice(tags, func(i, j int) bool {
		return tags[i].Key < tags[j].Key
	})

	sort.Slice(ecrTags, func(i, j int) bool {
		return *ecrTags[i].Key < *ecrTags[j].Key
	})
}

// DiffTags returns tags that should be added or removed.
func DiffTags(spec []v1alpha1.Tag, current []ecr.Tag) (addTags []ecr.Tag, remove []string) {
	addMap := make(map[string]string, len(spec))
	for _, t := range spec {
		addMap[t.Key] = t.Value
	}
	removeMap := map[string]struct{}{}
	for _, t := range current {
		if addMap[aws.StringValue(t.Key)] == aws.StringValue(t.Value) {
			delete(addMap, aws.StringValue(t.Key))
			continue
		}
		removeMap[aws.StringValue(t.Key)] = struct{}{}
	}
	for k, v := range addMap {
		addTags = append(addTags, ecr.Tag{Key: aws.String(k), Value: aws.String(v)})
	}
	for k := range removeMap {
		remove = append(remove, k)
	}
	return
}
