package securitygrouprule

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	awsec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-aws/apis/ec2/manualv1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/clients/ec2"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	connectaws "github.com/crossplane-contrib/provider-aws/pkg/utils/connect/aws"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

const (
	errUnexpectedObject = "The managed resource is not an SecurityGroupRule resource"
	errDelete           = "failed to delete the SecurityGroupRule resource"
	ingressType         = "ingress"
	egressType          = "egress"
)

// SetupSecurityGroupRule adds a controller that reconciles SecurityGroupRules.
func SetupSecurityGroupRule(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(manualv1alpha1.SecurityGroupRuleKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient(), newClientFn: ec2.NewSecurityGroupRuleClient}),
		managed.WithReferenceResolver(managed.NewAPISimpleReferenceResolver(mgr.GetClient())),
		managed.WithConnectionPublishers(),
		managed.WithInitializers(),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(manualv1alpha1.SecurityGroupRuleGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&manualv1alpha1.SecurityGroupRule{}).
		Complete(r)
}

type connector struct {
	kube        client.Client
	newClientFn func(config aws.Config) ec2.SecurityGroupRuleClient
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*manualv1alpha1.SecurityGroupRule)
	if !ok {
		return nil, errors.New(errUnexpectedObject)
	}
	cfg, err := connectaws.GetConfig(ctx, c.kube, mg, pointer.StringValue(cr.Spec.ForProvider.Region))
	if err != nil {
		return nil, err
	}
	return &external{client: c.newClientFn(*cfg), kube: c.kube}, nil
}

type external struct {
	kube   client.Client
	client ec2.SecurityGroupRuleClient
}

func (e *external) Observe(ctx context.Context, mgd resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mgd.(*manualv1alpha1.SecurityGroupRule)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	externalName := meta.GetExternalName(cr)

	if externalName == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	existingSgrP, err := e.getExternalSgr(ctx, externalName)

	// If err is not nil, the sgr does not exist
	if err != nil {
		return managed.ExternalObservation{ //nolint:nilerr
			ResourceExists: false,
		}, nil
	}
	cr.Status.AtProvider = manualv1alpha1.SecurityGroupRuleObservation{
		SecurityGroupRuleID: &externalName,
	}
	// Check if the two sgr are in sync
	needsUpdate, _, _ := compareSgr(&cr.Spec.ForProvider, existingSgrP)
	cr.SetConditions(xpv1.Available())
	return managed.ExternalObservation{
		ResourceExists:   true,
		ResourceUpToDate: !needsUpdate,
	}, nil
}

func (e *external) Create(ctx context.Context, mgd resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mgd.(*manualv1alpha1.SecurityGroupRule)

	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}
	err := e.createSgr(ctx, cr)

	return managed.ExternalCreation{}, err
}

func (e *external) createSgr(ctx context.Context, sgr *manualv1alpha1.SecurityGroupRule) error { //nolint: gocyclo
	if *sgr.Spec.ForProvider.Type == ingressType {
		providerValues := sgr.Spec.ForProvider
		input := &awsec2.AuthorizeSecurityGroupIngressInput{
			GroupId: providerValues.SecurityGroupID,
		}
		// To add an description to the sgr, we need to use the IpPermissions array
		if providerValues.CidrBlock != nil {
			input.IpPermissions = []awsec2types.IpPermission{{
				FromPort:   providerValues.FromPort,
				ToPort:     providerValues.ToPort,
				IpProtocol: providerValues.Protocol,
				IpRanges: []awsec2types.IpRange{{
					CidrIp: providerValues.CidrBlock,

					Description: providerValues.Description,
				}},
			}}
		}
		if providerValues.Ipv6CidrBlock != nil {
			input.IpPermissions = []awsec2types.IpPermission{{
				FromPort:   providerValues.FromPort,
				ToPort:     providerValues.ToPort,
				IpProtocol: providerValues.Protocol,
				Ipv6Ranges: []awsec2types.Ipv6Range{{
					CidrIpv6: providerValues.Ipv6CidrBlock,

					Description: providerValues.Description,
				}},
			}}
		}
		if providerValues.SourceSecurityGroupID != nil {
			input.IpPermissions = []awsec2types.IpPermission{{
				FromPort:   providerValues.FromPort,
				ToPort:     providerValues.ToPort,
				IpProtocol: providerValues.Protocol,
				UserIdGroupPairs: []awsec2types.UserIdGroupPair{{
					GroupId: providerValues.SourceSecurityGroupID,

					Description: providerValues.Description,
				}},
			}}
		}
		if providerValues.PrefixListID != nil {
			input.IpPermissions = []awsec2types.IpPermission{{
				FromPort:   providerValues.FromPort,
				ToPort:     providerValues.ToPort,
				IpProtocol: providerValues.Protocol,
				PrefixListIds: []awsec2types.PrefixListId{{
					Description:  providerValues.Description,
					PrefixListId: providerValues.PrefixListID,
				}},
			}}
		}
		result, err := e.client.AuthorizeSecurityGroupIngress(ctx, input)

		if err != nil {
			return err
		}

		if result != nil {
			if len(result.SecurityGroupRules) > 0 && result.SecurityGroupRules[0].SecurityGroupRuleId != nil {
				sgrID := result.SecurityGroupRules[0].SecurityGroupRuleId
				meta.SetExternalName(sgr, pointer.StringValue(sgrID))
			}
		}
	} else if *sgr.Spec.ForProvider.Type == egressType {
		providerValues := sgr.Spec.ForProvider
		input := &awsec2.AuthorizeSecurityGroupEgressInput{
			GroupId: providerValues.SecurityGroupID,
		}
		// To add an description to the sgr, we need to use the IpPermissions array
		if providerValues.CidrBlock != nil {
			input.IpPermissions = []awsec2types.IpPermission{{
				FromPort:   providerValues.FromPort,
				ToPort:     providerValues.ToPort,
				IpProtocol: providerValues.Protocol,
				IpRanges: []awsec2types.IpRange{{
					CidrIp: providerValues.CidrBlock,

					Description: providerValues.Description,
				}},
			}}
		}
		if providerValues.Ipv6CidrBlock != nil {
			input.IpPermissions = []awsec2types.IpPermission{{
				FromPort:   providerValues.FromPort,
				ToPort:     providerValues.ToPort,
				IpProtocol: providerValues.Protocol,
				Ipv6Ranges: []awsec2types.Ipv6Range{{
					CidrIpv6: providerValues.Ipv6CidrBlock,

					Description: providerValues.Description,
				}},
			}}
		}
		if providerValues.SourceSecurityGroupID != nil {
			input.IpPermissions = []awsec2types.IpPermission{{
				FromPort:   providerValues.FromPort,
				ToPort:     providerValues.ToPort,
				IpProtocol: providerValues.Protocol,
				UserIdGroupPairs: []awsec2types.UserIdGroupPair{{
					GroupId: providerValues.SourceSecurityGroupID,

					Description: providerValues.Description,
				}},
			}}
		}
		if providerValues.PrefixListID != nil {
			input.IpPermissions = []awsec2types.IpPermission{{
				FromPort:   providerValues.FromPort,
				ToPort:     providerValues.ToPort,
				IpProtocol: providerValues.Protocol,
				PrefixListIds: []awsec2types.PrefixListId{{
					Description:  providerValues.Description,
					PrefixListId: providerValues.PrefixListID,
				}},
			}}
		}
		result, err := e.client.AuthorizeSecurityGroupEgress(ctx, input)

		if err != nil {
			return err
		}

		if result != nil {
			if len(result.SecurityGroupRules) > 0 && result.SecurityGroupRules[0].SecurityGroupRuleId != nil {
				sgrID := result.SecurityGroupRules[0].SecurityGroupRuleId
				meta.SetExternalName(sgr, pointer.StringValue(sgrID))
			}
		}
	}
	return nil
}

func (e *external) Delete(ctx context.Context, mgd resource.Managed) error {
	cr, ok := mgd.(*manualv1alpha1.SecurityGroupRule)
	if !ok {
		return errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(xpv1.Deleting())
	return e.deleteSgr(ctx, cr)
}
func (e *external) deleteSgr(ctx context.Context, sgr *manualv1alpha1.SecurityGroupRule) error {
	return e.deleteSgrForType(ctx, sgr, *sgr.Spec.ForProvider.Type)
}

func (e *external) deleteSgrForType(ctx context.Context, sgr *manualv1alpha1.SecurityGroupRule, sgrType string) error {
	// We cant use the type of the sgr, because in case of an update of the type property of
	// an existing sgr, we need to delete the actual sgr with the old type
	if sgrType == ingressType {
		_, err := e.client.RevokeSecurityGroupIngress(ctx, &awsec2.RevokeSecurityGroupIngressInput{
			SecurityGroupRuleIds: []string{meta.GetExternalName(sgr)},
			GroupId:              sgr.Spec.ForProvider.SecurityGroupID,
		})

		return errorutils.Wrap(resource.Ignore(ec2.IsCIDRNotFound, err), errDelete)
	} else if sgrType == egressType {
		_, err := e.client.RevokeSecurityGroupEgress(ctx, &awsec2.RevokeSecurityGroupEgressInput{
			SecurityGroupRuleIds: []string{meta.GetExternalName(sgr)},
			GroupId:              sgr.Spec.ForProvider.SecurityGroupID,
		})

		return errorutils.Wrap(resource.Ignore(ec2.IsCIDRNotFound, err), errDelete)
	}
	return nil
}

func getTypeForDeletion(sgr *manualv1alpha1.SecurityGroupRule, switchType bool) string {
	// If we need to delete a sgr due to a cnage of the type,
	// the sgr that needs to be deleted has the old, that is opposite, type
	returnType := *sgr.Spec.ForProvider.Type
	if switchType {
		if returnType == ingressType {
			return egressType
		}
		return ingressType
	}
	return returnType

}

func (e *external) Update(ctx context.Context, mgd resource.Managed) (managed.ExternalUpdate, error) {

	cr, ok := mgd.(*manualv1alpha1.SecurityGroupRule)

	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}
	existingSgr, err := e.getExternalSgr(ctx, meta.GetExternalName(cr))
	if err != nil {
		return managed.ExternalUpdate{}, err
	}
	needsUpdate, recreate, typechange := compareSgr(&cr.Spec.ForProvider, existingSgr)

	if needsUpdate {
		if recreate {
			sgrType := getTypeForDeletion(cr, typechange)
			err := e.deleteSgrForType(ctx, cr, sgrType)
			if err != nil {
				return managed.ExternalUpdate{}, err
			}

			// We return an error to fore a recreation, as we cant create a new sgr and update externalName here
			return managed.ExternalUpdate{}, errors.New("Update needs recreation")
		}
		return e.updateSgr(ctx, cr)
	}
	return managed.ExternalUpdate{}, nil
}

func (e *external) updateSgr(ctx context.Context, cr *manualv1alpha1.SecurityGroupRule) (managed.ExternalUpdate, error) {
	externalName := meta.GetExternalName(cr)
	input := &awsec2.ModifySecurityGroupRulesInput{
		GroupId: cr.Spec.ForProvider.SecurityGroupID,
		SecurityGroupRules: []awsec2types.SecurityGroupRuleUpdate{{
			SecurityGroupRuleId: &externalName,
			SecurityGroupRule: &awsec2types.SecurityGroupRuleRequest{
				FromPort:    cr.Spec.ForProvider.FromPort,
				ToPort:      cr.Spec.ForProvider.ToPort,
				Description: cr.Spec.ForProvider.Description,
				IpProtocol:  cr.Spec.ForProvider.Protocol,
			},
		}},
	}
	if cr.Spec.ForProvider.CidrBlock != nil {
		input.SecurityGroupRules[0].SecurityGroupRule.CidrIpv4 = cr.Spec.ForProvider.CidrBlock
	}
	if cr.Spec.ForProvider.Ipv6CidrBlock != nil {
		input.SecurityGroupRules[0].SecurityGroupRule.CidrIpv6 = cr.Spec.ForProvider.Ipv6CidrBlock
	}
	if cr.Spec.ForProvider.SourceSecurityGroupID != nil {
		input.SecurityGroupRules[0].SecurityGroupRule.ReferencedGroupId = cr.Spec.ForProvider.SourceSecurityGroupID
	}
	if cr.Spec.ForProvider.PrefixListID != nil {
		input.SecurityGroupRules[0].SecurityGroupRule.PrefixListId = cr.Spec.ForProvider.PrefixListID
	}

	_, err := e.client.ModifySecurityGroupRules(ctx, input)
	return managed.ExternalUpdate{}, err
}

func compareSgr(desired *manualv1alpha1.SecurityGroupRuleParameters, actual *manualv1alpha1.SecurityGroupRuleParameters) (needsUpdate bool, recreate bool, typechange bool) {

	needsUpdate = false
	recreate = false
	typechange = false
	if pointer.StringValue(desired.CidrBlock) != pointer.StringValue(actual.CidrBlock) {
		needsUpdate = true
	}

	if pointer.StringValue(desired.Description) != pointer.StringValue(actual.Description) {
		needsUpdate = true
	}

	if pointer.Int32Value(desired.FromPort) != pointer.Int32Value(actual.FromPort) {
		needsUpdate = true
	}

	if pointer.Int32Value(desired.ToPort) != pointer.Int32Value(actual.ToPort) {
		needsUpdate = true
	}

	if pointer.StringValue(desired.Protocol) != pointer.StringValue(actual.Protocol) {
		needsUpdate = true
		recreate = true
	}

	if pointer.StringValue(desired.SourceSecurityGroupID) != pointer.StringValue(actual.SourceSecurityGroupID) {
		needsUpdate = true
	}

	if pointer.StringValue(desired.PrefixListID) != pointer.StringValue(actual.PrefixListID) {
		needsUpdate = true
	}

	if pointer.StringValue(desired.Type) != pointer.StringValue(actual.Type) {
		needsUpdate = true
		recreate = true
		typechange = true
	}

	return needsUpdate, recreate, typechange
}

func (e *external) getExternalSgr(ctx context.Context, externalName string) (*manualv1alpha1.SecurityGroupRuleParameters, error) {

	response, err := e.client.DescribeSecurityGroupRules(ctx, &awsec2.DescribeSecurityGroupRulesInput{
		SecurityGroupRuleIds: []string{externalName},
	})

	if err != nil {
		return nil, err
	}
	existingSgr := response.SecurityGroupRules[0]
	crType := ingressType
	if pointer.BoolValue(existingSgr.IsEgress) {
		crType = egressType
	}
	cr := &manualv1alpha1.SecurityGroupRuleParameters{
		Description:   existingSgr.Description,
		FromPort:      existingSgr.FromPort,
		ToPort:        existingSgr.ToPort,
		Type:          &crType,
		Protocol:      existingSgr.IpProtocol,
		CidrBlock:     existingSgr.CidrIpv4,
		Ipv6CidrBlock: existingSgr.CidrIpv6,
		PrefixListID:  existingSgr.PrefixListId,
	}
	if existingSgr.ReferencedGroupInfo != nil {
		cr.SourceSecurityGroupID = existingSgr.ReferencedGroupInfo.GroupId
	}

	return cr, nil
}
