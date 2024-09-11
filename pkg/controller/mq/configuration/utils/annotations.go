package utils

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/crossplane/crossplane-runtime/pkg/meta"

	svcapitypes "github.com/crossplane-contrib/provider-aws/apis/mq/v1alpha1"
	"github.com/crossplane-contrib/provider-aws/pkg/utils/pointer"
)

func SetLatestUnsanitizedConfiguration(cr *svcapitypes.Configuration) {
	latestDesiredConfig := pointer.StringValue(cr.Spec.ForProvider.Data)
	meta.AddAnnotations(cr, map[string]string{svcapitypes.LatestUnsanitizedConfiguration: hashText(latestDesiredConfig)})
}

func GetLatestUnsanitizedConfiguration(o *svcapitypes.Configuration) string {
	return o.GetAnnotations()[svcapitypes.LatestUnsanitizedConfiguration]
}

func hashText(s string) string {
	hash := sha256.Sum256([]byte(s))
	return hex.EncodeToString(hash[:])
}

func HasBeenSanitized(o *svcapitypes.Configuration) bool {
	return GetLatestUnsanitizedConfiguration(o) != ""
}

func HasBeenUpdatedPostSanitization(o *svcapitypes.Configuration) bool {
	latestDesiredConfig := pointer.StringValue(o.Spec.ForProvider.Data)
	return GetLatestUnsanitizedConfiguration(o) != hashText(latestDesiredConfig)
}
