package v1alpha3

import (
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/onsi/gomega"

	aws "github.com/crossplane/provider-aws/pkg/clients"
)

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}

func Test_VPC_BuildExternalStatusFromObservation(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	r := VPC{}

	r.UpdateExternalStatus(ec2.Vpc{})

	g.Expect(r.Status.VPCExternalStatus).ToNot(gomega.BeNil())
}

func Test_RouteTable_BuildExternalStatusFromObservation(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	r := RouteTable{}

	r.UpdateExternalStatus(ec2.RouteTable{
		Routes: []ec2.Route{
			{
				GatewayId: aws.String("some gateway id"),
			},
		},
		Associations: []ec2.RouteTableAssociation{
			{
				RouteTableId: aws.String("some id"),
			},
		},
	})

	g.Expect(r.Status.RouteTableExternalStatus).ToNot(gomega.BeNil())
}

func Test_SecurityGroup_BuildExternalStatusFromObservation(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	r := SecurityGroup{}

	r.UpdateExternalStatus(ec2.SecurityGroup{
		GroupId: aws.String("some gorup id"),
	})

	g.Expect(r.Status.SecurityGroupExternalStatus).ToNot(gomega.BeNil())
}

func Test_InternetGateway_BuildEC2Permissions(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	r := InternetGateway{}
	r.UpdateExternalStatus(ec2.InternetGateway{
		Attachments: []ec2.InternetGatewayAttachment{
			{VpcId: aws.String("arbitrary vpcid")},
		},
	})

	g.Expect(len(r.Status.InternetGatewayExternalStatus.Attachments)).To(gomega.Equal(1))
}

func Test_Subnet_BuildEC2Permissions(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	r := Subnet{}
	r.UpdateExternalStatus(ec2.Subnet{
		VpcId: aws.String("arbitrary vpcid"),
	})

	g.Expect(r.Status.SubnetExternalStatus).ToNot(gomega.BeNil())
}

func Test_Common_BuildFromEC2Tags(t *testing.T) {
	g := gomega.NewGomegaWithT(t)

	ec2tags := []ec2.Tag{
		{}, {},
	}

	res := BuildFromEC2Tags(ec2tags)

	g.Expect(len(res)).To(gomega.Equal(2))
}

func Test_Zone_ZoneObservationUpdate(t *testing.T) {
	g := gomega.NewGomegaWithT(t)
	r := ZoneObservation{}
	id := "/hostedzone/XXXXXXXXXXXXXXXXXXX"
	location := "https://route53.amazonaws.com/2013-04-01/hostedzone/XXXXXXXXXXXXXXXXXXX"
	var rrCount int64 = 2

	r.Update(&route53.CreateHostedZoneOutput{
		HostedZone: &route53.HostedZone{

			Id:                     &id,
			ResourceRecordSetCount: &rrCount,
		},
		Location: &location,
	})

	g.Expect(r.ID).ToNot(gomega.BeNil())
	g.Expect(r.ResourceRecordCount).ToNot(gomega.BeNil())
	g.Expect(r.Location).ToNot(gomega.BeNil())
}
