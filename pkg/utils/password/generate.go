package password

import (
	"github.com/go-passwd/randomstring"
	"k8s.io/utils/pointer"

	"github.com/crossplane-contrib/provider-aws/apis/common"
)

var defaultConstraints = common.PasswordConstraints{
	MinLength:            27, // for backwards compatibility with github.com/crossplane/crossplane-runtime/pkg/passwords
	MinUpperCaseLetters:  1,
	MinLowerCaseLetters:  1,
	MinDigits:            1,
	MinSpecialCharacters: 1,
	MaxLength:            1,
}

// GeneratePasswordFromConstraints generates a password based on the given
// password constraints.
func GeneratePasswordFromConstraints(constraints *common.PasswordConstraints) (string, error) {
	var c common.PasswordConstraints
	if constraints != nil {
		c = *constraints
	} else {
		c = defaultConstraints
	}

	opts := []any{
		randomstring.NewLength(uint(c.MaxLength)),
	}
	if c.MinUpperCaseLetters > 0 {
		opts = append(opts, randomstring.NewIncludeCharset(upperCaseLetters))
	}
	if c.MinLowerCaseLetters > 0 {
		opts = append(opts, randomstring.NewIncludeCharset(lowerCaseLetters))
	}
	if c.MinDigits > 0 {
		opts = append(opts, randomstring.NewIncludeCharset(digits))
	}
	if c.MinSpecialCharacters > 0 {
		opts = append(opts, randomstring.NewIncludeCharset(specialCharacters))
	}
	generator, err := randomstring.New(opts...)
	if err != nil {
		return "", err
	}
	res, err := generator.Generate()
	return pointer.StringDeref(res, ""), err
}
