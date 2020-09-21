package ecr

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ecr"
	"github.com/google/go-cmp/cmp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane/provider-aws/apis/ecr/v1alpha1"
	aws "github.com/crossplane/provider-aws/pkg/clients"
)

var (
	repositoryName  = "repo"
	testKey         = "key"
	testValue       = "value"
	createTime      = time.Now()
	tagMutability   = "MUTABLE"
	registryID      = "123"
	repositoryARN   = "arn"
	repositoryURI   = "testuri"
	imageScanConfig = v1alpha1.ImageScanningConfiguration{
		ScanOnPush: true,
	}
	awsImageScanConfig = ecr.ImageScanningConfiguration{
		ScanOnPush: &imageScanConfig.ScanOnPush,
	}
	ecrTag    = ecr.Tag{Key: &testKey, Value: &testValue}
	alpha1Tag = v1alpha1.Tag{Key: testKey, Value: testValue}
)

func TestGenerateRepositoryObservation(t *testing.T) {
	cases := map[string]struct {
		in  ecr.Repository
		out v1alpha1.RepositoryObservation
	}{
		"AllFilled": {
			in: ecr.Repository{
				CreatedAt:                  &createTime,
				ImageScanningConfiguration: &awsImageScanConfig,
				ImageTagMutability:         ecr.ImageTagMutability(tagMutability),
				RegistryId:                 aws.String(registryID),
				RepositoryName:             aws.String(repositoryName),
				RepositoryArn:              aws.String(repositoryARN),
				RepositoryUri:              aws.String(repositoryURI),
			},
			out: v1alpha1.RepositoryObservation{
				CreatedAt:      &metav1.Time{Time: createTime},
				RegistryID:     registryID,
				RepositoryName: repositoryName,
				RepositoryArn:  repositoryARN,
				RepositoryURI:  repositoryURI,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			r := GenerateRepositoryObservation(tc.in)
			if diff := cmp.Diff(tc.out, r); diff != "" {
				t.Errorf("GenerateRepositoryObservation(...): -want, +got:\n%s", diff)
			}
		})
	}
}

func TestIsRepositoryUpToDate(t *testing.T) {
	type args struct {
		ecrTags []ecr.Tag
		e       v1alpha1.RepositoryParameters
		repo    ecr.Repository
	}

	cases := map[string]struct {
		args args
		want bool
	}{
		"SameFields": {
			args: args{
				ecrTags: []ecr.Tag{ecrTag},
				e: v1alpha1.RepositoryParameters{
					ImageScanningConfiguration: &imageScanConfig,
					ImageTagMutability:         &tagMutability,
					Tags:                       []v1alpha1.Tag{alpha1Tag},
				},
				repo: ecr.Repository{
					ImageScanningConfiguration: &awsImageScanConfig,
					ImageTagMutability:         ecr.ImageTagMutabilityMutable,
				},
			},
			want: true,
		},
		"DifferentFields": {
			args: args{
				ecrTags: []ecr.Tag{},
				e: v1alpha1.RepositoryParameters{
					Tags: []v1alpha1.Tag{alpha1Tag},
				},
				repo: ecr.Repository{
					ImageScanningConfiguration: &awsImageScanConfig,
					ImageTagMutability:         ecr.ImageTagMutabilityMutable,
				},
			},
			want: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got, _ := IsRepositoryUpToDate(&tc.args.e, tc.args.ecrTags, &tc.args.repo)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateCreateRepositoryInput(t *testing.T) {
	type args struct {
		name string
		p    *v1alpha1.RepositoryParameters
	}

	cases := map[string]struct {
		args args
		want *ecr.CreateRepositoryInput
	}{
		"AllFields": {
			args: args{
				name: repositoryName,
				p: &v1alpha1.RepositoryParameters{
					Tags:                       []v1alpha1.Tag{alpha1Tag},
					ImageScanningConfiguration: &imageScanConfig,
					ImageTagMutability:         &tagMutability,
				},
			},
			want: &ecr.CreateRepositoryInput{
				RepositoryName:             &repositoryName,
				ImageTagMutability:         ecr.ImageTagMutabilityMutable,
				ImageScanningConfiguration: &awsImageScanConfig,
			},
		},
		"SomeFields": {
			args: args{
				name: repositoryName,
				p: &v1alpha1.RepositoryParameters{
					Tags:               []v1alpha1.Tag{alpha1Tag},
					ImageTagMutability: &tagMutability,
				},
			},
			want: &ecr.CreateRepositoryInput{
				RepositoryName:     &repositoryName,
				ImageTagMutability: ecr.ImageTagMutabilityMutable,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateCreateRepositoryInput(tc.args.name, tc.args.p)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestLateInitialize(t *testing.T) {
	cases := map[string]struct {
		parameters *v1alpha1.RepositoryParameters
		repository *ecr.Repository
		want       *v1alpha1.RepositoryParameters
	}{
		"AllOptionalFields": {
			parameters: &v1alpha1.RepositoryParameters{},
			repository: &ecr.Repository{
				ImageScanningConfiguration: &awsImageScanConfig,
				ImageTagMutability:         ecr.ImageTagMutabilityMutable,
			},
			want: &v1alpha1.RepositoryParameters{
				ImageScanningConfiguration: &imageScanConfig,
				ImageTagMutability:         &tagMutability,
			},
		},
		"SomeFieldsDontOverwrite": {
			parameters: &v1alpha1.RepositoryParameters{
				ImageScanningConfiguration: &imageScanConfig,
				ImageTagMutability:         &tagMutability,
				Tags:                       []v1alpha1.Tag{alpha1Tag},
			},
			repository: &ecr.Repository{
				ImageScanningConfiguration: &awsImageScanConfig,
				ImageTagMutability:         ecr.ImageTagMutabilityMutable,
			},
			want: &v1alpha1.RepositoryParameters{
				ImageScanningConfiguration: &imageScanConfig,
				ImageTagMutability:         &tagMutability,
				Tags:                       []v1alpha1.Tag{alpha1Tag},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			LateInitializeRepository(tc.parameters, tc.repository)
			if diff := cmp.Diff(tc.want, tc.parameters); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
