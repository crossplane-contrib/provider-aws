package zone

import (
	"errors"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/service/route53"
)

func Test_getRegion(t *testing.T) {
	tests := map[string]struct {
		name string
		want route53.VPCRegion
	}{
		"us-east-1": {
			name: "us-east-1",
			want: route53.VPCRegionUsEast1,
		},
		"us-east-2": {
			name: "us-east-2",
			want: route53.VPCRegionUsEast2,
		},
		"us-west-1": {
			name: "us-west-1",
			want: route53.VPCRegionUsWest1,
		},
		"us-west-2": {
			name: "us-west-2",
			want: route53.VPCRegionUsWest2,
		},
		"eu-west-1": {
			name: "eu-west-1",
			want: route53.VPCRegionEuWest1,
		},
		"eu-west-2": {
			name: "eu-west-2",
			want: route53.VPCRegionEuWest2,
		},
		"eu-west-3": {
			name: "eu-west-3",
			want: route53.VPCRegionEuWest3,
		},
		"eu-central-1": {
			name: "eu-central-1",
			want: route53.VPCRegionEuCentral1,
		},
		"ap-east-1": {
			name: "ap-east-1",
			want: route53.VPCRegionApEast1,
		},
		"me-south-1": {
			name: "me-south-1",
			want: route53.VPCRegionMeSouth1,
		},
		"ap-southeast-1": {
			name: "ap-southeast-1",
			want: route53.VPCRegionApSoutheast1,
		},
		"ap-southeast-2": {
			name: "ap-southeast-2",
			want: route53.VPCRegionApSoutheast2,
		},
		"ap-south-1": {
			name: "ap-south-1",
			want: route53.VPCRegionApSouth1,
		},
		"ap-northeast-1": {
			name: "ap-northeast-1",
			want: route53.VPCRegionApNortheast1,
		},
		"ap-northeast-2": {
			name: "ap-northeast-2",
			want: route53.VPCRegionApNortheast2,
		},
		"ap-northeast-3": {
			name: "ap-northeast-3",
			want: route53.VPCRegionApNortheast3,
		},
		"eu-north-1": {
			name: "eu-north-1",
			want: route53.VPCRegionEuNorth1,
		},
		"sa-east-1": {
			name: "sa-east-1",
			want: route53.VPCRegionSaEast1,
		},
		"ca-central-1": {
			name: "ca-central-1",
			want: route53.VPCRegionCaCentral1,
		},
		"cn-north-1": {
			name: "cn-north-1",
			want: route53.VPCRegionCnNorth1,
		},
		"blank-region": {
			name: "",
			want: "",
		},
		"wrong-region": {
			name: "us-il-9",
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getRegion(tt.name); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("getRegion() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsErrorNoSuchHostedZone(t *testing.T) {
	tests := map[string]struct {
		err  error
		want bool
	}{
		"validError": {
			err:  awserr.New(route53.ErrCodeNoSuchHostedZone, "The specified hosted zone does not exist.", errors.New(route53.ErrCodeNoSuchHostedZone)),
			want: true,
		},
		"invalidAwsError": {
			err:  awserr.New(route53.ErrCodeHostedZoneNotFound, "The specified HostedZone can't be found.", errors.New(route53.ErrCodeHostedZoneNotFound)),
			want: false,
		},
		"randomError": {
			err:  errors.New("the specified hosted zone does not exist"),
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.err.Error(), func(t *testing.T) {
			if got := IsErrorNoSuchHostedZone(tt.err); got != tt.want {
				t.Errorf("IsErrorNoSuchHostedZone() = %v, want %v", got, tt.want)
			}
		})
	}
}
