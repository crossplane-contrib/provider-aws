package ecr

import (
	"sort"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ecr"
	ecrtypes "github.com/aws/aws-sdk-go-v2/service/ecr/types"
	"github.com/aws/smithy-go/document"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/crossplane-contrib/provider-aws/apis/ecr/v1beta1"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
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
	imageScanConfig = v1beta1.ImageScanningConfiguration{
		ScanOnPush: true,
	}
	awsImageScanConfig = ecrtypes.ImageScanningConfiguration{
		ScanOnPush: imageScanConfig.ScanOnPush,
	}
	ecrTag    = ecrtypes.Tag{Key: &testKey, Value: &testValue}
	alpha1Tag = v1beta1.Tag{Key: testKey, Value: testValue}
)

func TestGenerateRepositoryObservation(t *testing.T) {
	cases := map[string]struct {
		in  ecrtypes.Repository
		out v1beta1.RepositoryObservation
	}{
		"AllFilled": {
			in: ecrtypes.Repository{
				CreatedAt:                  &createTime,
				ImageScanningConfiguration: &awsImageScanConfig,
				ImageTagMutability:         ecrtypes.ImageTagMutability(tagMutability),
				RegistryId:                 pointer.ToOrNilIfZeroValue(registryID),
				RepositoryName:             pointer.ToOrNilIfZeroValue(repositoryName),
				RepositoryArn:              pointer.ToOrNilIfZeroValue(repositoryARN),
				RepositoryUri:              pointer.ToOrNilIfZeroValue(repositoryURI),
			},
			out: v1beta1.RepositoryObservation{
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
		ecrTags []ecrtypes.Tag
		e       v1beta1.RepositoryParameters
		repo    ecrtypes.Repository
	}

	cases := map[string]struct {
		args args
		want bool
	}{
		"SameFields": {
			args: args{
				ecrTags: []ecrtypes.Tag{ecrTag},
				e: v1beta1.RepositoryParameters{
					ImageScanningConfiguration: &imageScanConfig,
					ImageTagMutability:         &tagMutability,
					Tags:                       []v1beta1.Tag{alpha1Tag},
				},
				repo: ecrtypes.Repository{
					ImageScanningConfiguration: &awsImageScanConfig,
					ImageTagMutability:         ecrtypes.ImageTagMutabilityMutable,
				},
			},
			want: true,
		},
		"DifferentFields": {
			args: args{
				ecrTags: []ecrtypes.Tag{},
				e: v1beta1.RepositoryParameters{
					Tags: []v1beta1.Tag{alpha1Tag},
				},
				repo: ecrtypes.Repository{
					ImageScanningConfiguration: &awsImageScanConfig,
					ImageTagMutability:         ecrtypes.ImageTagMutabilityMutable,
				},
			},
			want: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := IsRepositoryUpToDate(&tc.args.e, tc.args.ecrTags, &tc.args.repo)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestGenerateCreateRepositoryInput(t *testing.T) {
	type args struct {
		name string
		p    *v1beta1.RepositoryParameters
	}

	cases := map[string]struct {
		args args
		want *ecr.CreateRepositoryInput
	}{
		"AllFields": {
			args: args{
				name: repositoryName,
				p: &v1beta1.RepositoryParameters{
					Tags:                       []v1beta1.Tag{alpha1Tag},
					ImageScanningConfiguration: &imageScanConfig,
					ImageTagMutability:         &tagMutability,
				},
			},
			want: &ecr.CreateRepositoryInput{
				RepositoryName:             &repositoryName,
				ImageTagMutability:         ecrtypes.ImageTagMutabilityMutable,
				ImageScanningConfiguration: &awsImageScanConfig,
			},
		},
		"SomeFields": {
			args: args{
				name: repositoryName,
				p: &v1beta1.RepositoryParameters{
					Tags:               []v1beta1.Tag{alpha1Tag},
					ImageTagMutability: &tagMutability,
				},
			},
			want: &ecr.CreateRepositoryInput{
				RepositoryName:     &repositoryName,
				ImageTagMutability: ecrtypes.ImageTagMutabilityMutable,
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := GenerateCreateRepositoryInput(tc.args.name, tc.args.p)
			if diff := cmp.Diff(tc.want, got, cmpopts.IgnoreTypes(document.NoSerde{})); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}

func TestLateInitialize(t *testing.T) {
	cases := map[string]struct {
		parameters *v1beta1.RepositoryParameters
		repository *ecrtypes.Repository
		want       *v1beta1.RepositoryParameters
	}{
		"AllOptionalFields": {
			parameters: &v1beta1.RepositoryParameters{},
			repository: &ecrtypes.Repository{
				ImageScanningConfiguration: &awsImageScanConfig,
				ImageTagMutability:         ecrtypes.ImageTagMutabilityMutable,
			},
			want: &v1beta1.RepositoryParameters{
				ImageScanningConfiguration: &imageScanConfig,
				ImageTagMutability:         &tagMutability,
			},
		},
		"SomeFieldsDontOverwrite": {
			parameters: &v1beta1.RepositoryParameters{
				ImageScanningConfiguration: &imageScanConfig,
				ImageTagMutability:         &tagMutability,
				Tags:                       []v1beta1.Tag{alpha1Tag},
			},
			repository: &ecrtypes.Repository{
				ImageScanningConfiguration: &awsImageScanConfig,
				ImageTagMutability:         ecrtypes.ImageTagMutabilityMutable,
			},
			want: &v1beta1.RepositoryParameters{
				ImageScanningConfiguration: &imageScanConfig,
				ImageTagMutability:         &tagMutability,
				Tags:                       []v1beta1.Tag{alpha1Tag},
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

func TestDiffTags(t *testing.T) {
	type args struct {
		local  []v1beta1.Tag
		remote []ecrtypes.Tag
	}
	type want struct {
		add    []ecrtypes.Tag
		remove []string
	}
	cases := map[string]struct {
		args args
		want want
	}{
		"AllNew": {
			args: args{
				local: []v1beta1.Tag{
					{Key: "key", Value: "val"},
				},
			},
			want: want{
				add: []ecrtypes.Tag{
					{Key: pointer.ToOrNilIfZeroValue("key"), Value: pointer.ToOrNilIfZeroValue("val")},
				},
			},
		},
		"SomeNew": {
			args: args{
				local: []v1beta1.Tag{
					{Key: "key", Value: "val"},
					{Key: "key1", Value: "val1"},
					{Key: "key2", Value: "val2"},
				},
				remote: []ecrtypes.Tag{
					{Key: pointer.ToOrNilIfZeroValue("key"), Value: pointer.ToOrNilIfZeroValue("val")},
				},
			},
			want: want{
				add: []ecrtypes.Tag{
					{Key: pointer.ToOrNilIfZeroValue("key1"), Value: pointer.ToOrNilIfZeroValue("val1")},
					{Key: pointer.ToOrNilIfZeroValue("key2"), Value: pointer.ToOrNilIfZeroValue("val2")},
				},
			},
		},
		"Update": {
			args: args{
				local: []v1beta1.Tag{
					{Key: "key", Value: "different"},
					{Key: "key1", Value: "val1"},
					{Key: "key2", Value: "val2"},
				},
				remote: []ecrtypes.Tag{
					{Key: pointer.ToOrNilIfZeroValue("key"), Value: pointer.ToOrNilIfZeroValue("val")},
					{Key: pointer.ToOrNilIfZeroValue("key1"), Value: pointer.ToOrNilIfZeroValue("val1")},
					{Key: pointer.ToOrNilIfZeroValue("key2"), Value: pointer.ToOrNilIfZeroValue("val2")},
				},
			},
			want: want{
				add: []ecrtypes.Tag{
					{Key: pointer.ToOrNilIfZeroValue("key"), Value: pointer.ToOrNilIfZeroValue("different")},
				},
				remove: []string{"key"},
			},
		},
		"RemoveAll": {
			args: args{
				remote: []ecrtypes.Tag{
					{Key: pointer.ToOrNilIfZeroValue("key"), Value: pointer.ToOrNilIfZeroValue("val")},
					{Key: pointer.ToOrNilIfZeroValue("key1"), Value: pointer.ToOrNilIfZeroValue("val1")},
					{Key: pointer.ToOrNilIfZeroValue("key2"), Value: pointer.ToOrNilIfZeroValue("val2")},
				},
			},
			want: want{
				remove: []string{"key", "key1", "key2"},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			tagCmp := cmpopts.SortSlices(func(i, j ecrtypes.Tag) bool {
				return pointer.StringValue(i.Key) < pointer.StringValue(j.Key)
			})
			add, remove := DiffTags(tc.args.local, tc.args.remote)
			if diff := cmp.Diff(tc.want.add, add, tagCmp, cmpopts.IgnoreTypes(document.NoSerde{})); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
			sort.Strings(tc.want.remove)
			sort.Strings(remove)
			if diff := cmp.Diff(tc.want.remove, remove, tagCmp); diff != "" {
				t.Errorf("r: -want, +got:\n%s", diff)
			}
		})
	}
}
