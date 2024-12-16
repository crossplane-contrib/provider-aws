/*
   Copyright 2025 The Crossplane Authors.
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

package webacl

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	svcsdk "github.com/aws/aws-sdk-go/service/wafv2"
	svcsdkapi "github.com/aws/aws-sdk-go/service/wafv2/wafv2iface"
	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
	"github.com/crossplane/crossplane-runtime/pkg/connection"
	"github.com/crossplane/crossplane-runtime/pkg/controller"
	"github.com/crossplane/crossplane-runtime/pkg/event"
	"github.com/crossplane/crossplane-runtime/pkg/meta"
	"github.com/crossplane/crossplane-runtime/pkg/reconciler/managed"
	"github.com/crossplane/crossplane-runtime/pkg/resource"
	cpresource "github.com/crossplane/crossplane-runtime/pkg/resource"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/pkg/errors"
	"k8s.io/utils/ptr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/crossplane-contrib/provider-aws/apis/v1alpha1"
	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/wafv2/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/features"
	connectaws "github.com/crossplane-contrib/provider-aws/pkg/utils/connect/aws"
	errorutils "github.com/crossplane-contrib/provider-aws/pkg/utils/errors"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/jsonpatch"
	custommanaged "github.com/crossplane-contrib/provider-aws/pkg/utils/reconciler/managed"
	tagutils "github.com/crossplane-contrib/provider-aws/pkg/utils/tags"
)

// SetupWebACL adds a controller that reconciles SetupWebAcl.
func SetupWebACL(mgr ctrl.Manager, o controller.Options) error {
	name := managed.ControllerName(svcapitypes.WebACLKind)

	cps := []managed.ConnectionPublisher{managed.NewAPISecretPublisher(mgr.GetClient(), mgr.GetScheme())}
	if o.Features.Enabled(features.EnableAlphaExternalSecretStores) {
		cps = append(cps, connection.NewDetailsManager(mgr.GetClient(), v1alpha1.StoreConfigGroupVersionKind))
	}

	reconcilerOpts := []managed.ReconcilerOption{
		managed.WithInitializers(managed.NewNameAsExternalName(mgr.GetClient())),
		managed.WithCriticalAnnotationUpdater(custommanaged.NewRetryingCriticalAnnotationUpdater(mgr.GetClient())),
		managed.WithExternalConnecter(&customConnector{kube: mgr.GetClient()}),
		managed.WithPollInterval(o.PollInterval),
		managed.WithLogger(o.Logger.WithValues("controller", name)),
		managed.WithRecorder(event.NewAPIRecorder(mgr.GetEventRecorderFor(name))),
		managed.WithConnectionPublishers(cps...),
	}

	if o.Features.Enabled(features.EnableAlphaManagementPolicies) {
		reconcilerOpts = append(reconcilerOpts, managed.WithManagementPolicies())
	}

	r := managed.NewReconciler(mgr,
		resource.ManagedKind(svcapitypes.WebACLGroupVersionKind),
		reconcilerOpts...)

	return ctrl.NewControllerManagedBy(mgr).
		Named(name).
		WithOptions(o.ForControllerRuntime()).
		WithEventFilter(resource.DesiredStateChanged()).
		For(&svcapitypes.WebACL{}).
		Complete(r)
}

type statementWithInfiniteRecursion interface {
	svcsdk.Statement | svcsdk.AndStatement | svcsdk.OrStatement | svcsdk.NotStatement
}

// customConnector is external connector with overridden Observe method due to ACK v0.38.1 doesn't correctly generate it.
type customConnector struct {
	kube client.Client
}

type customExternal struct {
	external
	cache *cache
}

type shared struct {
	cache  *cache
	client svcsdkapi.WAFV2API
}

type cache struct {
	tagListOutput []*svcsdk.Tag
}

func newCustomExternal(kube client.Client, client svcsdkapi.WAFV2API) *customExternal {
	shared := &shared{client: client, cache: &cache{}}
	e := &customExternal{
		external{
			kube:           kube,
			client:         client,
			preObserve:     nopPreObserve,
			isUpToDate:     shared.isUpToDate,
			lateInitialize: nopLateInitialize,
			postObserve:    nopPostObserve,
			preCreate:      preCreate,
			postCreate:     nopPostCreate,
			preUpdate:      shared.preUpdate,
			postUpdate:     nopPostUpdate,
			preDelete:      preDelete,
			postDelete:     nopPostDelete,
		},
		shared.cache,
	}
	return e
}

func (c *customConnector) Connect(ctx context.Context, mg cpresource.Managed) (managed.ExternalClient, error) {
	cr, ok := mg.(*svcapitypes.WebACL)
	if !ok {
		return nil, errors.New(errUnexpectedObject)
	}
	sess, err := connectaws.GetConfigV1(ctx, c.kube, mg, cr.Spec.ForProvider.Region)
	if err != nil {
		return nil, errors.Wrap(err, errCreateSession)
	}
	return newCustomExternal(c.kube, svcsdk.New(sess)), nil
}

func (e *customExternal) Observe(ctx context.Context, mg cpresource.Managed) (managed.ExternalObservation, error) { //nolint:gocyclo
	cr, ok := mg.(*svcapitypes.WebACL)
	if !ok {
		return managed.ExternalObservation{}, errors.New(errUnexpectedObject)
	}
	if meta.GetExternalName(cr) == "" {
		return managed.ExternalObservation{
			ResourceExists: false,
		}, nil
	}
	input := GenerateGetWebACLInput(cr)
	input.Name = aws.String(meta.GetExternalName(cr))
	listWebACLInput := svcsdk.ListWebACLsInput{
		Scope: cr.Spec.ForProvider.Scope,
	}
	ls, err := e.client.ListWebACLs(&listWebACLInput)
	if err != nil {
		return managed.ExternalObservation{}, err
	}
	for n, webACLSummary := range ls.WebACLs {
		if aws.StringValue(webACLSummary.Name) == meta.GetExternalName(cr) {
			input.Id = webACLSummary.Id
			break
		}
		if n == len(ls.WebACLs)-1 {
			return managed.ExternalObservation{
				ResourceExists: false,
			}, nil
		}
	}
	if err := e.preObserve(ctx, cr, input); err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, "pre-observe failed")
	}
	resp, err := e.client.GetWebACLWithContext(ctx, input)
	if err != nil {
		return managed.ExternalObservation{ResourceExists: false}, errorutils.Wrap(cpresource.Ignore(IsNotFound, err), errDescribe)
	}
	currentSpec := cr.Spec.ForProvider.DeepCopy()
	if err := e.lateInitialize(&cr.Spec.ForProvider, resp); err != nil {
		return managed.ExternalObservation{}, errors.Wrap(err, "late-init failed")
	}
	GenerateWebACL(resp).Status.AtProvider.DeepCopyInto(&cr.Status.AtProvider)
	upToDate := true
	diff := ""
	if !meta.WasDeleted(cr) { // There is no need to run isUpToDate if the resource is deleted
		upToDate, diff, err = e.isUpToDate(ctx, cr, resp)
		if err != nil {
			return managed.ExternalObservation{}, errors.Wrap(err, "isUpToDate check failed")
		}
	}
	cr.SetConditions(xpv1.Available())
	cr.Status.AtProvider = svcapitypes.WebACLObservation{
		ARN:       resp.WebACL.ARN,
		ID:        resp.WebACL.Id,
		LockToken: resp.LockToken,
	}
	return e.postObserve(ctx, cr, resp, managed.ExternalObservation{
		ResourceExists:          true,
		ResourceUpToDate:        upToDate,
		Diff:                    diff,
		ResourceLateInitialized: !cmp.Equal(&cr.Spec.ForProvider, currentSpec),
	}, nil)
}

func (s *shared) isUpToDate(_ context.Context, cr *svcapitypes.WebACL, resp *svcsdk.GetWebACLOutput) (upToDate bool, diff string, err error) {
	listTagOutput, err := s.client.ListTagsForResource(&svcsdk.ListTagsForResourceInput{ResourceARN: cr.Status.AtProvider.ARN})
	if err != nil {
		return false, "", err
	}
	s.cache.tagListOutput = listTagOutput.TagInfoForResource.TagList
	patch, err := createPatch(&cr.Spec.ForProvider, resp, listTagOutput.TagInfoForResource.TagList)
	if err != nil {
		return false, "", err
	}

	opts := []cmp.Option{
		cmpopts.EquateEmpty(),
		// Name and Scope are immutables
		cmpopts.IgnoreFields(svcapitypes.WebACLParameters{}, "Region", "Scope"),
	}
	diff = cmp.Diff(&svcapitypes.WebACLParameters{}, patch, opts...)
	if diff != "" {
		return false, "Found observed difference in wafv2 webacl " + diff, nil
	}
	return true, "", nil
}

func preCreate(_ context.Context, cr *svcapitypes.WebACL, input *svcsdk.CreateWebACLInput) error {
	input.Name = aws.String(meta.GetExternalName(cr))
	err := setInputRuleStatementsFromJSON(cr, input.Rules)
	if err != nil {
		return err
	}
	return nil
}

func (s *shared) preUpdate(_ context.Context, cr *svcapitypes.WebACL, input *svcsdk.UpdateWebACLInput) error {
	input.Name = aws.String(meta.GetExternalName(cr))
	err := setInputRuleStatementsFromJSON(cr, input.Rules)
	if err != nil {
		return err
	}

	desiredTags := map[string]*string{}
	observedTags := map[string]*string{}

	for _, tag := range cr.Spec.ForProvider.Tags {
		desiredTags[*tag.Key] = tag.Value
	}
	for _, tag := range s.cache.tagListOutput {
		// ignore system tags
		if strings.HasPrefix(*tag.Key, "aws:") {
			continue
		}
		observedTags[*tag.Key] = tag.Value
	}
	tagsToAdd, tagsToRemove := tagutils.DiffTagsMapPtr(desiredTags, observedTags)

	if len(tagsToAdd) > 0 {
		var inputTags []*svcsdk.Tag
		for k, v := range tagsToAdd {
			inputTags = append(inputTags, &svcsdk.Tag{Key: aws.String(k), Value: v})
		}
		_, err = s.client.TagResource(&svcsdk.TagResourceInput{ResourceARN: cr.Status.AtProvider.ARN, Tags: inputTags})
		if err != nil {
			return err
		}
	}

	if len(tagsToRemove) > 0 {
		_, err = s.client.UntagResource(&svcsdk.UntagResourceInput{ResourceARN: cr.Status.AtProvider.ARN, TagKeys: tagsToRemove})
		if err != nil {
			return err
		}
	}

	return nil
}

func preDelete(_ context.Context, cr *svcapitypes.WebACL, input *svcsdk.DeleteWebACLInput) (bool, error) {
	input.Name = aws.String(meta.GetExternalName(cr))
	input.LockToken = cr.Status.AtProvider.LockToken
	return false, nil
}

// statementFromJSONString convert back to sdk types the rule statements which were ignored in generator-config.yaml and handled by the controller as string(json)
func statementFromJSONString[S statementWithInfiniteRecursion](jsonPointer *string) (*S, error) {
	jsonString := ptr.Deref(jsonPointer, "")
	var statement S
	err := json.Unmarshal([]byte(jsonString), &statement)
	if err != nil {
		return nil, err
	}
	return &statement, nil
}

// statementToJSONString converts the statement which the controller handles as string to JSON string
func statementToJSONString[S statementWithInfiniteRecursion](statement S) (*string, error) {
	configBytes, err := json.Marshal(statement)
	if err != nil {
		return nil, err
	}
	configStr := string(configBytes)
	return &configStr, nil
}

// setInputRuleStatementsFromJSON sets the input for rule statements which were ignored in generator-config.yaml and handled as string(json)
func setInputRuleStatementsFromJSON(cr *svcapitypes.WebACL, rules []*svcsdk.Rule) (err error) { //nolint:gocyclo
	for i, rule := range cr.Spec.ForProvider.Rules {
		if rule.Statement.OrStatement != nil {
			rules[i].Statement.OrStatement, err = statementFromJSONString[svcsdk.OrStatement](rule.Statement.OrStatement)
			if err != nil {
				return err
			}
		}
		if rule.Statement.AndStatement != nil {
			rules[i].Statement.AndStatement, err = statementFromJSONString[svcsdk.AndStatement](rule.Statement.AndStatement)
			if err != nil {
				return err
			}
		}
		if rule.Statement.NotStatement != nil {
			rules[i].Statement.NotStatement, err = statementFromJSONString[svcsdk.NotStatement](rule.Statement.NotStatement)
			if err != nil {
				return err
			}
		}
		if rule.Statement.ManagedRuleGroupStatement != nil && rule.Statement.ManagedRuleGroupStatement.ScopeDownStatement != nil {
			rules[i].Statement.ManagedRuleGroupStatement.ScopeDownStatement, err = statementFromJSONString[svcsdk.Statement](rule.Statement.ManagedRuleGroupStatement.ScopeDownStatement)
			if err != nil {
				return err
			}
		}
		if rule.Statement.RateBasedStatement != nil && rule.Statement.RateBasedStatement.ScopeDownStatement != nil {
			rules[i].Statement.RateBasedStatement.ScopeDownStatement, err = statementFromJSONString[svcsdk.Statement](rule.Statement.RateBasedStatement.ScopeDownStatement)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// TODO(teeverr): find an easier way to ignore case insensitive fields, probably is it possible via cmp.FilterPath/cmp.FilterValues? I didn't get how

// changeCaseInsensitiveFields changes the case of two in case-insensitive fields in wafv2.WebACL(https://github.com/aws/aws-sdk-go/blob/main/service/wafv2/api.go#L26332)
// which is important because current configuration might have different case than the external configuration(external has lower case every time) and isUpToDate will return false for an equal configuration
// These two fields are SingleHeader.Name and SingleQueryArgument.Name from FieldToMatch. FieldToMatch is used in ByteMatchStatement, RegexMatchStatement, RegexPatternSetReferenceStatement
// SizeConstraintStatement, SQLIMatchStatement, XSSMatchStatement which in turn can be placed in AndStatement, OrStatement, NotStatement in any deeper level of nestedness.
// This function works with svcsdk(aws-sdk) and svcapitypes(provider-aws) types because they have very similar structure
func changeCaseInsensitiveFields(params any) { //nolint:gocyclo
	v := reflect.Indirect(reflect.ValueOf(params))
	for i := 0; i < v.NumField(); i++ {
		field := reflect.TypeOf(v.Interface()).Field(i)
		if !v.FieldByName(field.Name).IsZero() {
			switch field.Type {
			case reflect.TypeOf([]*svcapitypes.Rule{}):
				traverseStuctList(field, v)
			case reflect.TypeOf([]*svcsdk.Rule{}):
				traverseStuctList(field, v)
			case reflect.TypeOf(&svcapitypes.Statement{}):
				traverseStruct(field, v)
			case reflect.TypeOf(&svcsdk.Statement{}):
				traverseStruct(field, v)
			case reflect.TypeOf([]*svcapitypes.Statement{}):
				traverseStuctList(field, v)
			case reflect.TypeOf([]*svcsdk.Statement{}):
				traverseStuctList(field, v)
				// AndStatement, AndStatement, NotStatement in svcapitypes have type *string and ingored here
			case reflect.TypeOf(&svcsdk.AndStatement{}):
				traverseStruct(field, v)
			case reflect.TypeOf(&svcsdk.OrStatement{}):
				traverseStruct(field, v)
			case reflect.TypeOf(&svcsdk.NotStatement{}):
				traverseStruct(field, v)
			case reflect.TypeOf(&svcapitypes.ByteMatchStatement{}):
				traverseStruct(field, v)
			case reflect.TypeOf(&svcsdk.ByteMatchStatement{}):
				traverseStruct(field, v)
			case reflect.TypeOf(&svcapitypes.RegexMatchStatement{}):
				traverseStruct(field, v)
			case reflect.TypeOf(&svcsdk.RegexMatchStatement{}):
				traverseStruct(field, v)
			case reflect.TypeOf(&svcapitypes.RegexPatternSetReferenceStatement{}):
				traverseStruct(field, v)
			case reflect.TypeOf(&svcsdk.RegexPatternSetReferenceStatement{}):
				traverseStruct(field, v)
			case reflect.TypeOf(&svcapitypes.SizeConstraintStatement{}):
				traverseStruct(field, v)
			case reflect.TypeOf(&svcsdk.SizeConstraintStatement{}):
				traverseStruct(field, v)
			case reflect.TypeOf(&svcapitypes.SQLIMatchStatement{}):
				traverseStruct(field, v)
			case reflect.TypeOf(&svcsdk.SqliMatchStatement{}):
				traverseStruct(field, v)
			case reflect.TypeOf(&svcapitypes.XSSMatchStatement{}):
				traverseStruct(field, v)
			case reflect.TypeOf(&svcsdk.XssMatchStatement{}):
				traverseStruct(field, v)
			case reflect.TypeOf(&svcapitypes.FieldToMatch{}):
				traverseStruct(field, v)
			case reflect.TypeOf(&svcsdk.FieldToMatch{}):
				traverseStruct(field, v)
			case reflect.TypeOf(&svcapitypes.SingleHeader{}):
				setToLower(field, v)
			case reflect.TypeOf(&svcsdk.SingleHeader{}):
				setToLower(field, v)
			case reflect.TypeOf(&svcapitypes.SingleQueryArgument{}):
				setToLower(field, v)
			case reflect.TypeOf(&svcsdk.SingleQueryArgument{}):
				setToLower(field, v)
			}
		}
	}
}

func setToLower(field reflect.StructField, v reflect.Value) {
	caseInSensitiveName := v.FieldByName(field.Name).Elem().FieldByName("Name").Elem()
	if caseInSensitiveName.IsValid() && caseInSensitiveName.CanSet() {
		lowerCasedName := strings.ToLower(caseInSensitiveName.String())
		caseInSensitiveName.SetString(lowerCasedName)
	}
}

func traverseStuctList(field reflect.StructField, v reflect.Value) {
	for i := 0; i < v.FieldByName(field.Name).Len(); i++ {
		interfaceValue := v.FieldByName(field.Name).Index(i).Interface()
		changeCaseInsensitiveFields(interfaceValue)
	}
}

func traverseStruct(field reflect.StructField, v reflect.Value) {
	interfaceValue := v.FieldByName(field.Name).Interface()
	changeCaseInsensitiveFields(interfaceValue)
}

// GenerateWebACL returns WebACLParameters with a diff between the current and external configuration
func createPatch(currentParams *svcapitypes.WebACLParameters, resp *svcsdk.GetWebACLOutput, respTagList []*svcsdk.Tag) (*svcapitypes.WebACLParameters, error) { //nolint:gocyclo
	targetConfig := currentParams.DeepCopy()
	changeCaseInsensitiveFields(targetConfig)
	externalConfig := GenerateWebACL(resp).Spec.ForProvider
	patch := &svcapitypes.WebACLParameters{}
	err := addJsonifiedRuleStatements(resp.WebACL.Rules, externalConfig)
	if err != nil {
		return patch, err
	}

	for i, rule := range targetConfig.Rules {
		// Unmarshalling the JSON string to the struct and marshaling it back to JSON string to further comparison with marshaled JSON string from the response
		if rule.Statement.AndStatement != nil {
			sdkStatement, err := statementFromJSONString[svcsdk.AndStatement](targetConfig.Rules[i].Statement.AndStatement)
			if err != nil {
				return patch, err
			}
			// Change the case of the fields which are case-insensitive, so that the comparison is accurate.
			// It is convinient to do it here, as we have the statement in the struct form(which is originally a json string)
			// Marshal the struct back to JSON string, so that it can be compared with the JSON string from the response because the JSON string from the response is marshaled from the struct as well
			changeCaseInsensitiveFields(sdkStatement)
			targetConfig.Rules[i].Statement.AndStatement, err = statementToJSONString[svcsdk.AndStatement](*sdkStatement)
			if err != nil {
				return patch, err
			}
		}
		if rule.Statement.OrStatement != nil {
			sdkStatement, err := statementFromJSONString[svcsdk.OrStatement](targetConfig.Rules[i].Statement.OrStatement)
			if err != nil {
				return patch, err
			}
			changeCaseInsensitiveFields(sdkStatement)
			targetConfig.Rules[i].Statement.OrStatement, err = statementToJSONString[svcsdk.OrStatement](*sdkStatement)
			if err != nil {
				return patch, err
			}
		}
		if rule.Statement.NotStatement != nil {
			sdkStatement, err := statementFromJSONString[svcsdk.NotStatement](targetConfig.Rules[i].Statement.NotStatement)
			if err != nil {
				return patch, err
			}
			changeCaseInsensitiveFields(sdkStatement)
			targetConfig.Rules[i].Statement.NotStatement, err = statementToJSONString[svcsdk.NotStatement](*sdkStatement)
			if err != nil {
				return patch, err
			}
		}
		if rule.Statement.ManagedRuleGroupStatement != nil && rule.Statement.ManagedRuleGroupStatement.ScopeDownStatement != nil {
			sdkStatement, err := statementFromJSONString[svcsdk.Statement](targetConfig.Rules[i].Statement.ManagedRuleGroupStatement.ScopeDownStatement)
			if err != nil {
				return patch, err
			}
			changeCaseInsensitiveFields(sdkStatement)
			targetConfig.Rules[i].Statement.ManagedRuleGroupStatement.ScopeDownStatement, err = statementToJSONString[svcsdk.Statement](*sdkStatement)
			if err != nil {
				return patch, err
			}
		}
		if rule.Statement.RateBasedStatement != nil && rule.Statement.RateBasedStatement.ScopeDownStatement != nil {
			sdkStatement, err := statementFromJSONString[svcsdk.Statement](targetConfig.Rules[i].Statement.RateBasedStatement.ScopeDownStatement)
			if err != nil {
				return patch, err
			}
			changeCaseInsensitiveFields(sdkStatement)
			targetConfig.Rules[i].Statement.RateBasedStatement.ScopeDownStatement, err = statementToJSONString[svcsdk.Statement](*sdkStatement)
			if err != nil {
				return patch, err
			}
		}
	}
	for _, v := range respTagList {
		externalConfig.Tags = append(externalConfig.Tags, &svcapitypes.Tag{Key: v.Key, Value: v.Value})
	}
	jsonPatch, err := jsonpatch.CreateJSONPatch(externalConfig, targetConfig)
	if err != nil {
		return patch, err
	}
	if err := json.Unmarshal(jsonPatch, patch); err != nil {
		return patch, err
	}
	return patch, nil
}

// addJsonifiedRuleStatements adds the Jsonified rule statements to the externalConfig(other fields prepared by GenerateWebACL)
func addJsonifiedRuleStatements(resp []*svcsdk.Rule, externalConfig svcapitypes.WebACLParameters) error { //nolint:gocyclo
	for i, rule := range resp {
		if rule.Statement.AndStatement != nil {
			jsonStringStatement, err := statementToJSONString[svcsdk.AndStatement](*rule.Statement.AndStatement)
			if err != nil {
				return err
			}
			externalConfig.Rules[i].Statement.AndStatement = jsonStringStatement
		}
		if rule.Statement.OrStatement != nil {
			jsonStringStatement, err := statementToJSONString[svcsdk.OrStatement](*rule.Statement.OrStatement)
			if err != nil {
				return err
			}
			externalConfig.Rules[i].Statement.OrStatement = jsonStringStatement
		}
		if rule.Statement.NotStatement != nil {
			jsonStringStatement, err := statementToJSONString[svcsdk.NotStatement](*rule.Statement.NotStatement)
			if err != nil {
				return err
			}
			externalConfig.Rules[i].Statement.NotStatement = jsonStringStatement
		}
		if rule.Statement.ManagedRuleGroupStatement != nil && rule.Statement.ManagedRuleGroupStatement.ScopeDownStatement != nil {
			jsonStringStatement, err := statementToJSONString[svcsdk.Statement](*rule.Statement.ManagedRuleGroupStatement.ScopeDownStatement)
			if err != nil {
				return err
			}
			externalConfig.Rules[i].Statement.ManagedRuleGroupStatement.ScopeDownStatement = jsonStringStatement
		}
		if rule.Statement.RateBasedStatement != nil && rule.Statement.RateBasedStatement.ScopeDownStatement != nil {
			jsonStringStatement, err := statementToJSONString[svcsdk.Statement](*rule.Statement.RateBasedStatement.ScopeDownStatement)
			if err != nil {
				return err
			}
			externalConfig.Rules[i].Statement.RateBasedStatement.ScopeDownStatement = jsonStringStatement
		}
	}
	return nil
}
