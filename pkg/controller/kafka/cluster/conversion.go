/*
Copyright 2023 The Crossplane Authors.

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

package cluster

import (
	svcsdk "github.com/aws/aws-sdk-go/service/kafka"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/kafka/v1alpha1"
)

func generateManagedConnectivityInfo(current *svcsdk.ConnectivityInfo) *svcapitypes.CustomConnectivityInfo {
	if current == nil {
		return nil
	}

	res := &svcapitypes.CustomConnectivityInfo{}
	if current.PublicAccess != nil {
		res.PublicAccess = &svcapitypes.CustomPublicAccess{
			Type: current.PublicAccess.Type,
		}
	}
	if current.VpcConnectivity != nil {
		res.VPCConnectivity = &svcapitypes.VPCConnectivity{}
		if current.VpcConnectivity.ClientAuthentication != nil {
			res.VPCConnectivity.ClientAuthentication = &svcapitypes.VPCConnectivityClientAuthentication{}
			if current.VpcConnectivity.ClientAuthentication.Sasl != nil {
				res.VPCConnectivity.ClientAuthentication.SASL = &svcapitypes.VPCConnectivitySASL{}
				if current.VpcConnectivity.ClientAuthentication.Sasl.Iam != nil {
					res.VPCConnectivity.ClientAuthentication.SASL.IAM = &svcapitypes.VPCConnectivityIAM{
						Enabled: current.VpcConnectivity.ClientAuthentication.Sasl.Iam.Enabled,
					}
				}
				if current.VpcConnectivity.ClientAuthentication.Sasl.Scram != nil {
					res.VPCConnectivity.ClientAuthentication.SASL.SCRAM = &svcapitypes.VPCConnectivitySCRAM{
						Enabled: current.VpcConnectivity.ClientAuthentication.Sasl.Scram.Enabled,
					}
				}
			}
			if current.VpcConnectivity.ClientAuthentication.Tls != nil {
				res.VPCConnectivity.ClientAuthentication.TLS = &svcapitypes.VPCConnectivityTLS{
					Enabled: current.VpcConnectivity.ClientAuthentication.Tls.Enabled,
				}
			}
		}
	}
	return res
}

func generateAWSConnectivityInfo(current *svcapitypes.CustomConnectivityInfo) *svcsdk.ConnectivityInfo {
	if current == nil {
		return nil
	}

	res := &svcsdk.ConnectivityInfo{}
	if current.PublicAccess != nil {
		res.PublicAccess = &svcsdk.PublicAccess{
			Type: current.PublicAccess.Type,
		}
	}
	if current.VPCConnectivity != nil {
		res.VpcConnectivity = &svcsdk.VpcConnectivity{}
		if current.VPCConnectivity.ClientAuthentication != nil {
			res.VpcConnectivity.ClientAuthentication = &svcsdk.VpcConnectivityClientAuthentication{}
			if current.VPCConnectivity.ClientAuthentication.SASL != nil {
				res.VpcConnectivity.ClientAuthentication.Sasl = &svcsdk.VpcConnectivitySasl{}
				if current.VPCConnectivity.ClientAuthentication.SASL.IAM != nil {
					res.VpcConnectivity.ClientAuthentication.Sasl.Iam = &svcsdk.VpcConnectivityIam{
						Enabled: current.VPCConnectivity.ClientAuthentication.SASL.IAM.Enabled,
					}
				}
				if current.VPCConnectivity.ClientAuthentication.SASL.SCRAM != nil {
					res.VpcConnectivity.ClientAuthentication.Sasl.Scram = &svcsdk.VpcConnectivityScram{
						Enabled: current.VPCConnectivity.ClientAuthentication.SASL.SCRAM.Enabled,
					}
				}
			}
			if current.VPCConnectivity.ClientAuthentication.TLS != nil {
				res.VpcConnectivity.ClientAuthentication.Tls = &svcsdk.VpcConnectivityTls{
					Enabled: current.VPCConnectivity.ClientAuthentication.TLS.Enabled,
				}
			}
		}
	}
	return res
}
