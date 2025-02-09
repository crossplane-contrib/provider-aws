package taskdefinitionfamily

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	awsecs "github.com/aws/aws-sdk-go/service/ecs"
	awsecsiface "github.com/aws/aws-sdk-go/service/ecs/ecsiface"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"github.com/pkg/errors"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	ecs "github.com/crossplane-contrib/provider-aws/apis/ecs/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	tdfclient "github.com/crossplane-contrib/provider-aws/pkg/clients/ecs/taskdefinitionfamily"
	"github.com/crossplane-contrib/provider-aws/pkg/controller/ecs/taskdefinition"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	connectaws "github.com/crossplane-contrib/provider-aws/pkg/utils/connect/aws"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
)

const (
	errUnexpectedObject = "managed resource is not a TaskDefinitionFamily resource"
	errCreateSession    = "cannot create a new session"
	errCreate           = "cannot create TaskDefinition in AWS"
	errUpdate           = "cannot update TaskDefinition in AWS"
	errDescribe         = "failed to describe TaskDefinition"
	errDelete           = "failed to delete TaskDefinition"
)

// SetupTaskDefinitionFamily adds a controller that reconciles a TaskDefinitionFamily.
func SetupTaskDefinitionFamily(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(ecs.TaskDefinitionFamilyGroupKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&connector{kube: mgr.GetClient()}),
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
		resource.ManagedKind(ecs.TaskDefinitionFamilyGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&ecs.TaskDefinitionFamily{}).
		Complete(r)
}

type connector struct {
	kube client.Client
}

type external struct {
	client awsecsiface.ECSAPI
	kube   client.Client
}

func (c *connector) Connect(ctx context.Context, mg resource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*ecs.TaskDefinitionFamily)
	if !ok {
		return nil, errors.New(errUnexpectedObject)
	}
	sess, err := connectaws.GetConfigV1(ctx, c.kube, mg, cr.Spec.ForProvider.Region)
	if err != nil {
		return nil, errors.Wrap(err, errCreateSession)
	}
	return &external{
		kube:   c.kube,
		client: awsecs.New(sess),
	}, nil
}

func (e *external) Observe(ctx context.Context, mg resource.Managed) (managed.ExternalObservation, error) {
	cr, ok := mg.(*ecs.TaskDefinitionFamily)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}

	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}

	input := taskdefinition.GenerateDescribeTaskDefinitionInput(GenerateTaskDefinition(cr))
	input.SetTaskDefinition(meta.GetExternalName(cr))
	if err := input.Validate(); err != nil {
		return managed.ExternalObservation{}, err
	}
	input.Include = []*string{ptr.To("TAGS")}

	resp, err := e.client.DescribeTaskDefinitionWithContext(ctx, input)
	if err != nil {
		return managed.ExternalObservation{ResourceExists: false}, errorutils.Wrap(resource.Ignore(isNotFound, err), errDescribe)
	}
	currentSpec := cr.Spec.ForProvider.DeepCopy()
	tdfclient.LateInitialize(&cr.Spec.ForProvider, resp)

	GenerateTaskDefinitionFamily(taskdefinition.GenerateTaskDefinition(resp)).
		Status.AtProvider.DeepCopyInto(&cr.Status.AtProvider)

	resourceExists := true
	if aws.StringValue(cr.Status.AtProvider.TaskDefinition.Status) == awsecs.TaskDefinitionStatusActive {
		cr.SetConditions(xpv1.Available())
	}
	if aws.StringValue(cr.Status.AtProvider.TaskDefinition.Status) == awsecs.TaskDefinitionStatusInactive {
		// Deleted task definitions can still be described in the API and show up with an INACTIVE status.
		resourceExists = false
		cr.SetConditions(xpv1.Unavailable())
	}

	isUpToDate, diff := tdfclient.IsUpToDate(cr, resp)

	return managed.ExternalObservation{
		ResourceExists:          resourceExists,
		ResourceUpToDate:        isUpToDate,
		Diff:                    diff,
		ResourceLateInitialized: !cmp.Equal(cr.Spec.ForProvider, *currentSpec),
	}, nil
}

func (e *external) Create(ctx context.Context, mg resource.Managed) (managed.ExternalCreation, error) {
	cr, ok := mg.(*ecs.TaskDefinitionFamily)
	if !ok {
		return managed.ExternalCreation{}, errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(xpv1.Creating())

	input := taskdefinition.GenerateRegisterTaskDefinitionInput(GenerateTaskDefinition(cr))
	input.ExecutionRoleArn = cr.Spec.ForProvider.ExecutionRoleARN
	input.TaskRoleArn = cr.Spec.ForProvider.TaskRoleARN
	input.Volumes = taskdefinition.GenerateVolumes(GenerateTaskDefinition(cr))

	resp, err := e.client.RegisterTaskDefinitionWithContext(ctx, input)
	if err != nil {
		return managed.ExternalCreation{}, errorutils.Wrap(err, errCreate)
	}

	meta.SetExternalName(cr, aws.StringValue(stripRevision(resp.TaskDefinition.TaskDefinitionArn)))

	return managed.ExternalCreation{}, nil
}

func (e *external) Update(ctx context.Context, mg resource.Managed) (managed.ExternalUpdate, error) {
	cr, ok := mg.(*ecs.TaskDefinitionFamily)
	if !ok {
		return managed.ExternalUpdate{}, errors.New(errUnexpectedObject)
	}

	createInput := taskdefinition.GenerateRegisterTaskDefinitionInput(GenerateTaskDefinition(cr))
	createInput.ExecutionRoleArn = cr.Spec.ForProvider.ExecutionRoleARN
	createInput.TaskRoleArn = cr.Spec.ForProvider.TaskRoleARN
	createInput.Volumes = taskdefinition.GenerateVolumes(GenerateTaskDefinition(cr))

	_, err := e.client.RegisterTaskDefinitionWithContext(ctx, createInput)
	if err != nil {
		return managed.ExternalUpdate{}, errorutils.Wrap(err, errCreate)
	}

	if cr.Status.AtProvider.TaskDefinition != nil && cr.Status.AtProvider.TaskDefinition.TaskDefinitionARN != nil {
		deleteInput := &awsecs.DeregisterTaskDefinitionInput{
			TaskDefinition: cr.Status.AtProvider.TaskDefinition.TaskDefinitionARN,
		}
		_, err = e.client.DeregisterTaskDefinitionWithContext(ctx, deleteInput)
		if err != nil {
			return managed.ExternalUpdate{}, errorutils.Wrap(err, errDelete)
		}
	}

	return managed.ExternalUpdate{}, nil
}

func (e *external) Delete(ctx context.Context, mg resource.Managed) (managed.ExternalDelete, error) {
	cr, ok := mg.(*ecs.TaskDefinitionFamily)
	if !ok {
		return managed.ExternalDelete{}, errors.New(errUnexpectedObject)
	}

	cr.Status.SetConditions(xpv1.Deleting())

	input := taskdefinition.GenerateDeregisterTaskDefinitionInput(GenerateTaskDefinition(cr))
	input.SetTaskDefinition(*cr.Status.AtProvider.TaskDefinition.TaskDefinitionARN)
	_, err := e.client.DeregisterTaskDefinitionWithContext(ctx, input)

	return managed.ExternalDelete{}, errorutils.Wrap(resource.Ignore(taskdefinition.IsNotFound, err), errDelete)
}

func (e *external) Disconnect(ctx context.Context) error {
	// Unimplemented, required by newer versions of crossplane-runtime
	return nil
}

// Strips the revision of a TaskDefinition ARN.
func stripRevision(arn *string) *string {
	if arn != nil {
		if idx := strings.LastIndex(*arn, ":"); idx != -1 {
			if _, err := fmt.Sscanf((*arn)[idx:], ":%d", new(int)); err == nil {
				return ptr.To((*arn)[:idx])
			}
		}
		return arn
	}
	return nil
}

// IsNotFound returns whether the given error is of type NotFound or not.
func isNotFound(err error) bool {
	awsErr, ok := err.(awserr.Error) //nolint:errorlint
	// There is no specific error for a 404 Not Found, it always returns 400.
	return ok && awsErr.Code() == "ClientException"
}
