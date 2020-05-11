package zone

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/awserr"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/route53iface"

	"github.com/crossplane/provider-aws/apis/network/v1alpha3"
)

// Client defines Route53 Client operations
type Client interface {
	CreateZoneRequest(cr *v1alpha3.Zone) route53.CreateHostedZoneRequest
	DeleteZoneRequest(id *string) route53.DeleteHostedZoneRequest
	GetZoneRequest(id *string) route53.GetHostedZoneRequest
	UpdateZoneRequest(id, comment *string) route53.UpdateHostedZoneCommentRequest
}

type zoneClient struct {
	zone route53iface.ClientAPI
}

// NewClient creates new AWS Client with provided AWS Configurations/Credentials
func NewClient(config *aws.Config) Client {
	return &zoneClient{zone: route53.New(*config)}
}

// GetZoneRequest returns a route53 GetHostedZoneRequest using which a route53
// Hosted Zone can be fetched and checked for existence.
func (c *zoneClient) GetZoneRequest(id *string) route53.GetHostedZoneRequest {
	return c.zone.GetHostedZoneRequest(&route53.GetHostedZoneInput{
		Id: id,
	})
}

// CreateZoneRequest returns a route53 CreateHostedZoneRequest using which a route53
// Hosted Zone can be created.
func (c *zoneClient) CreateZoneRequest(cr *v1alpha3.Zone) route53.CreateHostedZoneRequest {

	reqInput := &route53.CreateHostedZoneInput{
		CallerReference: cr.Spec.ForProvider.CallerReference,
		Name:            cr.Spec.ForProvider.Name,
		HostedZoneConfig: &route53.HostedZoneConfig{
			PrivateZone: cr.Spec.ForProvider.PrivateZone,
			Comment:     cr.Spec.ForProvider.Comment,
		},
	}

	if *cr.Spec.ForProvider.PrivateZone {
		reqInput.HostedZoneConfig.PrivateZone = cr.Spec.ForProvider.PrivateZone

		if cr.Spec.ForProvider.VPCId != nil && cr.Spec.ForProvider.VPCRegion != nil {
			reqInput.VPC = &route53.VPC{
				VPCId:     cr.Spec.ForProvider.VPCId,
				VPCRegion: getRegion(*cr.Spec.ForProvider.VPCRegion),
			}
		}
	}

	return c.zone.CreateHostedZoneRequest(reqInput)
}

// UpdateZoneRequest returns a route53 UpdateHostedZoneRequest using which a route53
// Hosted Zone can be updated.
func (c *zoneClient) UpdateZoneRequest(id, comment *string) route53.UpdateHostedZoneCommentRequest {
	return c.zone.UpdateHostedZoneCommentRequest(&route53.UpdateHostedZoneCommentInput{Comment: comment, Id: id})
}

// DeleteZoneRequest returns a route53 DeleteHostedZoneRequest using which a route53
// Hosted Zone can be deleted.
func (c *zoneClient) DeleteZoneRequest(id *string) route53.DeleteHostedZoneRequest {
	return c.zone.DeleteHostedZoneRequest(&route53.DeleteHostedZoneInput{
		Id: id,
	})
}

func getRegion(r string) route53.VPCRegion {
	awsRegionMap := map[string]route53.VPCRegion{
		"us-east-1":      route53.VPCRegionUsEast1,
		"us-east-2":      route53.VPCRegionUsEast2,
		"us-west-1":      route53.VPCRegionUsWest1,
		"us-west-2":      route53.VPCRegionUsWest2,
		"eu-west-1":      route53.VPCRegionEuWest1,
		"eu-west-2":      route53.VPCRegionEuWest2,
		"eu-west-3":      route53.VPCRegionEuWest3,
		"eu-central-1":   route53.VPCRegionEuCentral1,
		"ap-east-1":      route53.VPCRegionApEast1,
		"me-south-1":     route53.VPCRegionMeSouth1,
		"ap-southeast-1": route53.VPCRegionApSoutheast1,
		"ap-southeast-2": route53.VPCRegionApSoutheast2,
		"ap-south-1":     route53.VPCRegionApSouth1,
		"ap-northeast-1": route53.VPCRegionApNortheast1,
		"ap-northeast-2": route53.VPCRegionApNortheast2,
		"ap-northeast-3": route53.VPCRegionApNortheast3,
		"eu-north-1":     route53.VPCRegionEuNorth1,
		"sa-east-1":      route53.VPCRegionSaEast1,
		"ca-central-1":   route53.VPCRegionCaCentral1,
		"cn-north-1":     route53.VPCRegionCnNorth1,
	}

	return awsRegionMap[r]

}

// IsErrorNoSuchHostedZone returns true if the error code indicates that the requested Zone was not found
func IsErrorNoSuchHostedZone(err error) bool {
	if zoneErr, ok := err.(awserr.Error); ok && zoneErr.Code() == route53.ErrCodeNoSuchHostedZone {
		return true
	}
	return false
}
