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

package user

import (
	svcsdk "github.com/aws/aws-sdk-go/service/transfer"
	"k8s.io/utils/ptr"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/transfer/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

func generateAPIStatusSSHPublicKeys(obj []*svcsdk.SshPublicKey) []*svcapitypes.SshPublicKey {
	if obj == nil {
		return nil
	}

	res := make([]*svcapitypes.SshPublicKey, len(obj))
	for i, key := range obj {
		if key == nil {
			continue
		}
		res[i] = &svcapitypes.SshPublicKey{
			DateImported:     pointer.TimeToMetaTime(key.DateImported),
			SshPublicKeyBody: key.SshPublicKeyBody,
			SshPublicKeyID:   key.SshPublicKeyId,
		}
	}
	return res
}

func generateAPISSHPublicKeys(obj []*svcsdk.SshPublicKey) []svcapitypes.SSHPublicKeySpec {
	if obj == nil {
		return nil
	}

	res := make([]svcapitypes.SSHPublicKeySpec, len(obj))
	for i, key := range obj {
		if key == nil {
			continue
		}
		res[i] = svcapitypes.SSHPublicKeySpec{
			Body: ptr.Deref(key.SshPublicKeyBody, ""),
		}
	}
	return res
}
