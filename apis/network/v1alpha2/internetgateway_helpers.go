/*
Copyright 2019 The Crossplane Authors.

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

package v1alpha2

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

// UpdateExternalStatus updates the external status object, given the observation
func (i *InternetGateway) UpdateExternalStatus(observation ec2.InternetGateway) {
	attachments := make([]InternetGatewayAttachment, len(observation.Attachments))
	for k, a := range observation.Attachments {
		attachments[k] = InternetGatewayAttachment{
			AttachmentStatus: string(a.State),
			VPCID:            aws.StringValue(a.VpcId),
		}
	}

	i.Status.InternetGatewayExternalStatus = InternetGatewayExternalStatus{
		InternetGatewayID: aws.StringValue(observation.InternetGatewayId),
		Attachments:       attachments,
	}
}
