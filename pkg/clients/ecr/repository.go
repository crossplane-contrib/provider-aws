package ecr

import (
	"context"
	"encoding/json"
	"errors"
	"sort"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecr"
	ecrtypes "github.com/aws/aws-sdk-go-v2/service/ecr/types"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane-contrib/provider-aws/apis/ecr/v1beta1"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/jsonpatch"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
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
	CreateRepository(ctx context.Context, input *ecr.CreateRepositoryInput, opts ...func(*ecr.Options)) (*ecr.CreateRepositoryOutput, error)
	DescribeRepositories(ctx context.Context, input *ecr.DescribeRepositoriesInput, opts ...func(*ecr.Options)) (*ecr.DescribeRepositoriesOutput, error)
	DeleteRepository(ctx context.Context, input *ecr.DeleteRepositoryInput, opts ...func(*ecr.Options)) (*ecr.DeleteRepositoryOutput, error)
	ListTagsForResource(ctx context.Context, input *ecr.ListTagsForResourceInput, opts ...func(*ecr.Options)) (*ecr.ListTagsForResourceOutput, error)
	TagResource(ctx context.Context, input *ecr.TagResourceInput, opts ...func(*ecr.Options)) (*ecr.TagResourceOutput, error)
	PutImageTagMutability(ctx context.Context, input *ecr.PutImageTagMutabilityInput, opts ...func(*ecr.Options)) (*ecr.PutImageTagMutabilityOutput, error)
	PutImageScanningConfiguration(ctx context.Context, input *ecr.PutImageScanningConfigurationInput, opts ...func(*ecr.Options)) (*ecr.PutImageScanningConfigurationOutput, error)
	UntagResource(ctx context.Context, input *ecr.UntagResourceInput, opts ...func(*ecr.Options)) (*ecr.UntagResourceOutput, error)
}

// GenerateRepositoryObservation is used to produce v1alpha1.RepositoryObservation from
// ecr.Repository
func GenerateRepositoryObservation(repo ecrtypes.Repository) v1beta1.RepositoryObservation {
	o := v1beta1.RepositoryObservation{
		RegistryID:     pointer.StringValue(repo.RegistryId),
		RepositoryArn:  pointer.StringValue(repo.RepositoryArn),
		RepositoryName: pointer.StringValue(repo.RepositoryName),
		RepositoryURI:  pointer.StringValue(repo.RepositoryUri),
	}

	if repo.CreatedAt != nil {
		o.CreatedAt = &metav1.Time{Time: *repo.CreatedAt}
	}
	return o
}

// LateInitializeRepository fills the empty fields in *v1alpha1.RepositoryParameters with
// the values seen in ecr.Repository.
func LateInitializeRepository(in *v1beta1.RepositoryParameters, r *ecrtypes.Repository) {
	if r == nil {
		return
	}
	if r.ImageScanningConfiguration != nil && in.ImageScanningConfiguration == nil {
		scanConfig := v1beta1.ImageScanningConfiguration{
			ScanOnPush: r.ImageScanningConfiguration.ScanOnPush,
		}
		in.ImageScanningConfiguration = &scanConfig
	}
	in.ImageTagMutability = pointer.LateInitialize(in.ImageTagMutability, pointer.ToOrNilIfZeroValue(string(r.ImageTagMutability)))
}

// CreatePatch creates a *v1alpha1.RepositoryParameters that has only the changed
// values between the target *v1alpha1.RepositoryParameters and the current
// *ecr.Repository.
func CreatePatch(in *ecrtypes.Repository, target *v1beta1.RepositoryParameters) (*v1beta1.RepositoryParameters, error) {
	currentParams := &v1beta1.RepositoryParameters{}
	LateInitializeRepository(currentParams, in)

	jsonPatch, err := jsonpatch.CreateJSONPatch(currentParams, target)
	if err != nil {
		return nil, err
	}
	patch := &v1beta1.RepositoryParameters{}
	if err := json.Unmarshal(jsonPatch, patch); err != nil {
		return nil, err
	}
	return patch, nil
}

// IsRepositoryUpToDate checks whether there is a change in any of the modifiable fields.
func IsRepositoryUpToDate(e *v1beta1.RepositoryParameters, tags []ecrtypes.Tag, repo *ecrtypes.Repository) bool {
	switch {
	case e.ImageScanningConfiguration != nil && repo.ImageScanningConfiguration != nil:
		if e.ImageScanningConfiguration.ScanOnPush != repo.ImageScanningConfiguration.ScanOnPush {
			return false
		}
	case e.ImageScanningConfiguration != nil && repo.ImageScanningConfiguration == nil:
		return false
	case e.ImageScanningConfiguration == nil && repo.ImageScanningConfiguration != nil:
		return false
	}
	return strings.EqualFold(pointer.StringValue(e.ImageTagMutability), string(repo.ImageTagMutability)) &&
		CompareTags(e.Tags, tags)
}

// IsRepoNotFoundErr returns true if the error is because the item doesn't exist
func IsRepoNotFoundErr(err error) bool {
	var notFoundError *ecrtypes.RepositoryNotFoundException
	return errors.As(err, &notFoundError)
}

// GenerateCreateRepositoryInput Generates the CreateRepositoryInput from the RepositoryParameters
func GenerateCreateRepositoryInput(name string, params *v1beta1.RepositoryParameters) *ecr.CreateRepositoryInput {
	c := &ecr.CreateRepositoryInput{
		RepositoryName:     pointer.ToOrNilIfZeroValue(name),
		ImageTagMutability: ecrtypes.ImageTagMutability(pointer.StringValue(params.ImageTagMutability)),
	}
	if params.ImageScanningConfiguration != nil {
		scanConfig := ecrtypes.ImageScanningConfiguration{
			ScanOnPush: params.ImageScanningConfiguration.ScanOnPush,
		}
		c.ImageScanningConfiguration = &scanConfig
	}
	return c
}

// CompareTags compares arrays of v1alpha1.Tag and ecrtypes.Tag
func CompareTags(tags []v1beta1.Tag, ecrTags []ecrtypes.Tag) bool {
	if len(tags) != len(ecrTags) {
		return false
	}

	SortTags(tags, ecrTags)

	for i, t := range tags {
		if t.Key != aws.ToString(ecrTags[i].Key) || t.Value != aws.ToString(ecrTags[i].Value) {
			return false
		}
	}

	return true
}

// SortTags sorts array of v1alpha1.Tag and ecrtypes.Tag on 'Key'
func SortTags(tags []v1beta1.Tag, ecrTags []ecrtypes.Tag) {
	sort.Slice(tags, func(i, j int) bool {
		return tags[i].Key < tags[j].Key
	})

	sort.Slice(ecrTags, func(i, j int) bool {
		return *ecrTags[i].Key < *ecrTags[j].Key
	})
}

// DiffTags returns tags that should be added or removed.
func DiffTags(spec []v1beta1.Tag, current []ecrtypes.Tag) (addTags []ecrtypes.Tag, remove []string) {
	addMap := make(map[string]string, len(spec))
	for _, t := range spec {
		addMap[t.Key] = t.Value
	}
	removeMap := map[string]struct{}{}
	for _, t := range current {
		if addMap[aws.ToString(t.Key)] == aws.ToString(t.Value) {
			delete(addMap, aws.ToString(t.Key))
			continue
		}
		removeMap[aws.ToString(t.Key)] = struct{}{}
	}
	for k, v := range addMap {
		addTags = append(addTags, ecrtypes.Tag{Key: aws.String(k), Value: aws.String(v)})
	}
	for k := range removeMap {
		remove = append(remove, k)
	}
	return
}
