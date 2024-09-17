package taskdefinition

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/ecs"
	"github.com/google/go-cmp/cmp"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/ecs/v1alpha1"
)

func TestConvertVolumes(t *testing.T) {
	cases := map[string]struct {
		reason string
		cr     *svcapitypes.TaskDefinition
		want   []*svcsdk.Volume
	}{
		"TestEmptyVolumeSpec": {
			"When passed empty volumes generate an empty slice",
			&svcapitypes.TaskDefinition{
				Spec: svcapitypes.TaskDefinitionSpec{
					ForProvider: svcapitypes.TaskDefinitionParameters{
						CustomTaskDefinitionParameters: svcapitypes.CustomTaskDefinitionParameters{},
					},
				},
			},
			[]*svcsdk.Volume{},
		},
		"TestMultipleVolumeSpec": {
			"When passed multiple volumes generate multiple volume types",
			&svcapitypes.TaskDefinition{
				Spec: svcapitypes.TaskDefinitionSpec{
					ForProvider: svcapitypes.TaskDefinitionParameters{
						CustomTaskDefinitionParameters: svcapitypes.CustomTaskDefinitionParameters{
							Volumes: []*svcapitypes.CustomVolume{
								{
									Name: aws.String("test1"),
								},
								{
									Name: aws.String("test2"),
								},
							},
						},
					},
				},
			},
			[]*svcsdk.Volume{
				{
					Name: aws.String("test1"),
				},
				{
					Name: aws.String("test2"),
				},
			},
		},
		"TestDockerVolumeConfiguration": {
			"When passed a volume with dockerVolumeConfiguration, generate a volume type with DockerVolumeConfiguration defined",
			&svcapitypes.TaskDefinition{
				Spec: svcapitypes.TaskDefinitionSpec{
					ForProvider: svcapitypes.TaskDefinitionParameters{
						CustomTaskDefinitionParameters: svcapitypes.CustomTaskDefinitionParameters{
							Volumes: []*svcapitypes.CustomVolume{
								{
									Name: aws.String("Test"),
									DockerVolumeConfiguration: &svcapitypes.DockerVolumeConfiguration{
										Autoprovision: aws.Bool(true),
										Driver:        aws.String("test"),
										DriverOpts: map[string]*string{
											"type":   aws.String("tmpfs"),
											"device": aws.String("tmpfs"),
										},
										Labels: map[string]*string{
											"test": aws.String("yes"),
										},
										Scope: aws.String("task"),
									},
								},
							},
						},
					},
				},
			},
			[]*svcsdk.Volume{
				{
					Name: aws.String("Test"),
					DockerVolumeConfiguration: &svcsdk.DockerVolumeConfiguration{
						Autoprovision: aws.Bool(true),
						Driver:        aws.String("test"),
						DriverOpts: map[string]*string{
							"type":   aws.String("tmpfs"),
							"device": aws.String("tmpfs"),
						},
						Labels: map[string]*string{
							"test": aws.String("yes"),
						},
						Scope: aws.String("task"),
					},
				},
			},
		},
		"TestEFSVolumeConfiguration": {
			"When passed a volume with efsVolumeConfiguration, generate a volume type with EFSVolumeConfiguration defined",
			&svcapitypes.TaskDefinition{
				Spec: svcapitypes.TaskDefinitionSpec{
					ForProvider: svcapitypes.TaskDefinitionParameters{
						CustomTaskDefinitionParameters: svcapitypes.CustomTaskDefinitionParameters{
							Volumes: []*svcapitypes.CustomVolume{
								{
									Name: aws.String("test"),
									EFSVolumeConfiguration: &svcapitypes.CustomEFSVolumeConfiguration{
										AuthorizationConfig: &svcapitypes.CustomEFSAuthorizationConfig{
											AccessPointID: aws.String("test-access"),
											IAM:           aws.String("DISABLED"),
										},
										FileSystemID:          aws.String("test-filesystem"),
										RootDirectory:         aws.String("/"),
										TransitEncryption:     aws.String("ENABLED"),
										TransitEncryptionPort: aws.Int64(443),
									},
								},
							},
						},
					},
				},
			},
			[]*svcsdk.Volume{
				{
					Name: aws.String("test"),
					EfsVolumeConfiguration: &svcsdk.EFSVolumeConfiguration{
						AuthorizationConfig: &svcsdk.EFSAuthorizationConfig{
							AccessPointId: aws.String("test-access"),
							Iam:           aws.String("DISABLED"),
						},
						FileSystemId:          aws.String("test-filesystem"),
						RootDirectory:         aws.String("/"),
						TransitEncryption:     aws.String("ENABLED"),
						TransitEncryptionPort: aws.Int64(443),
					},
				},
			},
		},
		"TestFsxWindowsFileServerVolumeConfiguration": {
			"When passed a volume with FsxWindowsFileServerVolumeConfiguration, generate a volume type with FsxWindowsFileServerVolumeConfiguration defined",
			&svcapitypes.TaskDefinition{
				Spec: svcapitypes.TaskDefinitionSpec{
					ForProvider: svcapitypes.TaskDefinitionParameters{
						CustomTaskDefinitionParameters: svcapitypes.CustomTaskDefinitionParameters{
							Volumes: []*svcapitypes.CustomVolume{
								{
									Name: aws.String("test"),
									FsxWindowsFileServerVolumeConfiguration: &svcapitypes.FSxWindowsFileServerVolumeConfiguration{
										AuthorizationConfig: &svcapitypes.FSxWindowsFileServerAuthorizationConfig{
											CredentialsParameter: aws.String("test"),
											Domain:               aws.String("example.com"),
										},
										FileSystemID:  aws.String("test-filesystem"),
										RootDirectory: aws.String("/"),
									},
								},
							},
						},
					},
				},
			},
			[]*svcsdk.Volume{
				{
					Name: aws.String("test"),
					FsxWindowsFileServerVolumeConfiguration: &svcsdk.FSxWindowsFileServerVolumeConfiguration{
						AuthorizationConfig: &svcsdk.FSxWindowsFileServerAuthorizationConfig{
							CredentialsParameter: aws.String("test"),
							Domain:               aws.String("example.com"),
						},
						FileSystemId:  aws.String("test-filesystem"),
						RootDirectory: aws.String("/"),
					},
				},
			},
		},
		"TestHostVolumeProperties": {
			"When passed a volume with Host field, generate a volume type with HostVolumeProperties",
			&svcapitypes.TaskDefinition{
				Spec: svcapitypes.TaskDefinitionSpec{
					ForProvider: svcapitypes.TaskDefinitionParameters{
						CustomTaskDefinitionParameters: svcapitypes.CustomTaskDefinitionParameters{
							Volumes: []*svcapitypes.CustomVolume{
								{
									Name: aws.String("test"),
									Host: &svcapitypes.HostVolumeProperties{
										SourcePath: aws.String("/foo"),
									},
								},
							},
						},
					},
				},
			},
			[]*svcsdk.Volume{
				{
					Name: aws.String("test"),
					Host: &svcsdk.HostVolumeProperties{
						SourcePath: aws.String("/foo"),
					},
				},
			},
		},
	}

	for name, tc := range cases {

		t.Run(name, func(t *testing.T) {
			got := GenerateVolumes(tc.cr)
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("%s\nExample(...): -want, +got:\n%s", tc.reason, diff)
			}
		})
	}
}
