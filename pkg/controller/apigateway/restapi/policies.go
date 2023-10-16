package restapi

import (
	"encoding/json"
	"regexp"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/crossplane/crossplane-runtime/pkg/errors"
	"github.com/google/go-cmp/cmp"

	apigwclient "github.com/crossplane-contrib/provider-aws/pkg/clients/apigateway"
)

func normalizePolicy(p *string) (*string, error) {
	if p == nil {
		return p, nil
	}

	mappedPol, err := policyStringToMap(p)
	if err != nil {
		return nil, errors.Wrap(err, "cannot conv to normalize")
	}

	return policyMapToString(mappedPol)
}

type wrapper struct {
	Data string `json:"data,omitempty"`
}

func policyEscapedStringToMap(p *string) (map[string]interface{}, error) {
	if p == nil {
		return nil, nil
	}

	var wrapper wrapper
	val := []byte("{\"data\":" + `"` + *p + `"` + "}")
	if err := json.Unmarshal(val, &wrapper); err != nil {
		return nil, errors.Wrap(err, "cannot unmarshal policy")
	}

	var pol map[string]interface{}
	if err := json.Unmarshal([]byte(wrapper.Data), &pol); err != nil {
		return nil, errors.Wrap(err, "cannot unmarshal policy")
	}

	return pol, nil
}

func policyStringToMap(p *string) (map[string]interface{}, error) {
	if p == nil {
		return nil, nil
	}
	var pol map[string]interface{}
	err := json.Unmarshal([]byte(*p), &pol)
	if err != nil {
		return nil, err
	}
	return pol, nil
}

func policyMapToString(p map[string]interface{}) (*string, error) {
	if p == nil {
		return nil, nil
	}
	parsed, err := json.Marshal(p)
	if err != nil {
		return nil, errors.Wrap(err, "cannot marshal policy")
	}

	return aws.String(string(parsed)), nil
}

func policiesAreKindOfTheSame(a map[string]interface{}, b map[string]interface{}) (bool, error) {
	if a != nil && b != nil {
		polPatch, err := apigwclient.GetJSONPatch(a, b)
		if err != nil {
			return false, errors.Wrap(err, "cannot compute jsonpatch")
		}

		for _, p := range polPatch {
			re := regexp.MustCompile(`/Statement/(\d+)/Resource`)
			if p.Operation != "replace" || !re.MatchString(p.Path) {
				return false, nil
			}

			index, _ := strconv.Atoi(re.FindString(p.Path))
			if a["Statement"].([]interface{})[index].(map[string]interface{})["Resource"] != "execute-api:/*/*/*" {
				return false, nil
			}
		}

		return true, nil
	}

	return cmp.Equal(a, b), nil
}
