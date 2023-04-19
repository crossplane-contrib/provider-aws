package password

import (
	"github.com/go-passwd/validator"

	"github.com/crossplane-contrib/provider-aws/apis/common"
)

const (
	upperCaseLetters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	lowerCaseLetters = "abcdefghijklmnopqrstuvwxyz"
	digits           = "0123456789"
	// `/`, `"`, or `@` are forbidden per default by AWS RDS.
	specialCharacters = "!#$%^&*"
)

// ValidatePassword based on constraints. Returns an error if the password if
// password is invalid.
// If constraints is nil, no checks are performed and nil is returned.
func ValidatePassword(constraints *common.PasswordConstraints, password string) error {
	if constraints == nil {
		return nil
	}
	passwordValidator := validator.New(
		validator.MinLength(constraints.MinLength, nil),
		validator.MaxLength(constraints.MaxLength, nil),
		validator.ContainsAtLeast(upperCaseLetters, constraints.MinUpperCaseLetters, nil),
		validator.ContainsAtLeast(lowerCaseLetters, constraints.MinLowerCaseLetters, nil),
		validator.ContainsAtLeast(digits, constraints.MinDigits, nil),
		validator.ContainsAtLeast(specialCharacters, constraints.MinSpecialCharacters, nil),
	)
	return passwordValidator.Validate(password)
}
