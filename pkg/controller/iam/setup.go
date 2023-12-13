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

package iam

import (
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane-contrib/provider-aws/pkg/controller/iam/accesskey"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/iam/group"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/iam/grouppolicyattachment"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/iam/groupusermembership"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/iam/instanceprofile"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/iam/openidconnectprovider"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/iam/policy"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/iam/role"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/iam/rolepolicy"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/iam/rolepolicyattachment"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/iam/servicelinkedrole"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/iam/user"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/iam/userpolicyattachment"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/setup"
)

// Setup iam controllers.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	return setup.SetupControllers(
		mgr, o,
		accesskey.SetupAccessKey,
		group.SetupGroup,
		grouppolicyattachment.SetupGroupPolicyAttachment,
		groupusermembership.SetupGroupUserMembership,
		instanceprofile.SetupInstanceProfile,
		openidconnectprovider.SetupOpenIDConnectProvider,
		policy.SetupPolicy,
		role.SetupRole,
		rolepolicy.SetupRolePolicy,
		rolepolicyattachment.SetupRolePolicyAttachment,
		servicelinkedrole.SetupServiceLinkedRole,
		user.SetupUser,
		userpolicyattachment.SetupUserPolicyAttachment,
	)
}
