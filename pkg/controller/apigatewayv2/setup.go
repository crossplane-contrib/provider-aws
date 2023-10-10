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

package apigatewayv2

import (
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/crossplane-contrib/provider-aws/pkg/controller/apigatewayv2/api"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/apigatewayv2/apimapping"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/apigatewayv2/authorizer"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/apigatewayv2/deployment"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/apigatewayv2/domainname"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/apigatewayv2/integration"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/apigatewayv2/integrationresponse"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/apigatewayv2/model"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/apigatewayv2/route"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/apigatewayv2/routeresponse"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/apigatewayv2/stage"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/apigatewayv2/vpclink"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/setup"
)

// Setup apigatewayv2 controllers.
func Setup(mgr ctrl.Manager, o controller.Options) error {
	return setup.SetupControllers(
		mgr, o,
		api.SetupAPI,
		apimapping.SetupAPIMapping,
		authorizer.SetupAuthorizer,
		deployment.SetupDeployment,
		domainname.SetupDomainName,
		integration.SetupIntegration,
		integrationresponse.SetupIntegrationResponse,
		model.SetupModel,
		route.SetupRoute,
		routeresponse.SetupRouteResponse,
		stage.SetupStage,
		vpclink.SetupVPCLink,
	)
}
