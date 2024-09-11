package configuration

import (
	"encoding/base64"
	"fmt"
	"strconv"

	svcsdk "github.com/aws/aws-sdk-go/service/mq"
	cperrors "github.com/crossplane/crossplane-runtime/pkg/errors"
	"github.com/crossplane/crossplane-runtime/pkg/meta"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/mq/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

func generateDescribeConfigurationRevisionInput(id *string, revision *int64) *svcsdk.DescribeConfigurationRevisionInput {
	currRevision := strconv.FormatInt(pointer.Int64Value(revision), 10)
	res := &svcsdk.DescribeConfigurationRevisionInput{
		ConfigurationId:       id,
		ConfigurationRevision: &currRevision,
	}

	return res
}

func generateUpdateConfigurationRequest(cr *svcapitypes.Configuration) *svcsdk.UpdateConfigurationRequest {
	res := &svcsdk.UpdateConfigurationRequest{
		ConfigurationId: pointer.ToOrNilIfZeroValue(meta.GetExternalName(cr)),
		Data:            pointer.ToOrNilIfZeroValue(base64.StdEncoding.EncodeToString([]byte(pointer.StringValue(cr.Spec.ForProvider.Data)))),
		Description:     cr.Spec.ForProvider.Description,
	}
	return res
}

const (
	ErrCannotAddTags         = "cannot add tags"
	ErrCannotRemoveTags      = "cannot remove tags"
	ErrSanitizedConfig       = "The desired configuration has been sanitized, please adjust the data field accordingly:\n"
	ErrUnknownSanitization   = "An unknown sanitization reason occurred."
	ErrDisallowedElement     = "The element '%s' was removed because it is disallowed.\n"
	ErrDisallowedAttribute   = "The attribute '%s' of element '%s' was removed because it is disallowed.\n"
	ErrInvalidAttributeValue = "The attribute '%s' of element '%s' was removed because it had an invalid value.\n"
)

func handleSanitizationWarnings(warnings []*svcsdk.SanitizationWarning) error {
	message := ErrSanitizedConfig
	for _, w := range warnings {
		reason := pointer.StringValue(w.Reason)
		switch svcapitypes.SanitizationWarningReason(reason) {
		case svcapitypes.SanitizationWarningReason_DISALLOWED_ELEMENT_REMOVED:
			message += fmt.Sprintf(ErrDisallowedElement, pointer.StringValue(w.ElementName))
		case svcapitypes.SanitizationWarningReason_DISALLOWED_ATTRIBUTE_REMOVED:
			message += fmt.Sprintf(ErrDisallowedAttribute, pointer.StringValue(w.AttributeName), pointer.StringValue(w.ElementName))
		case svcapitypes.SanitizationWarningReason_INVALID_ATTRIBUTE_VALUE_REMOVED:
			message += fmt.Sprintf(ErrInvalidAttributeValue, pointer.StringValue(w.AttributeName), pointer.StringValue(w.ElementName))
		default:
			message += ErrUnknownSanitization
		}
	}
	return cperrors.New(message)
}
