package common

// PasswordConstraintOption represents a setting for a password constraint.
// type PasswordConstraintOption string

// // PasswordConstraintOptionRequired values.
// const (
// 	PasswordConstraintOptionRequired  PasswordConstraintOption = "Required"
// 	PasswordConstraintOptionOptional  PasswordConstraintOption = "Optional"
// 	PasswordConstraintOptionForbidden PasswordConstraintOption = "Forbidden"
// )

// const (
// 	// PasswordDefaultMinLength is the default minimum length of a password.
// 	PasswordDefaultMinLength = 27
// )

// PasswordConstraints defines constraints and restriction for a password.
type PasswordConstraints struct {
	// // UpperCaseLetters specifies, whether the password should contain upper
	// // case characters (A-Z).
	// // +kubebuilder:validation:Enum=Required;Optional;Forbidden
	// // +kubebuilder:validation:Default=Required
	// UpperCaseLetters PasswordConstraintOption `json:"upperCaseLetters"`

	// // LowerCaseLetters specifies, whether the password should contain lower
	// // case characters (a-z).
	// // +kubebuilder:validation:Enum=Required;Optional;Forbidden
	// // +kubebuilder:validation:Default=Required
	// LowerCaseLetters PasswordConstraintOption `json:"lowerCaseLetters"`

	// // Digits specifies, whether the password should contain digits (0-9).
	// // +kubebuilder:validation:Enum=Required;Optional;Forbidden
	// // +kubebuilder:validation:Default=Required
	// Digits PasswordConstraintOption `json:"digits"`

	// // SpecialCharacters specifies, whether the password should contain
	// // special characters.
	// // +kubebuilder:validation:Enum=Required;Optional;Forbidden
	// // +kubebuilder:validation:Default=Required
	// SpecialCharacters PasswordConstraintOption `json:"specialCharacters"`

	// MinUpperCaseLetters specifies the minimal number of upper case letters.
	// +kubebuilder:validation:Minimum=0
	MinUpperCaseLetters int `json:"minUpperCaseLetters"`

	// MinLowerCaseLetters specifies the minimal number of lower case letters.
	// +kubebuilder:validation:Minimum=0
	MinLowerCaseLetters int `json:"minLowerCaseLetters"`

	// MinDigits specifies the minimal number of digits.
	// +kubebuilder:validation:Minimum=0
	MinDigits int `json:"minDigits"`

	// MinSpecialCharacters specifies the minimal number of special characters.
	// +kubebuilder:validation:Minimum=0
	MinSpecialCharacters int `json:"minSpecialCharacters"`

	// MinLength specifies the minimum length of the password.
	// +kubebuilder:validation:Minimum=1
	MinLength int `json:"minimumLength"`

	// MaxLength specifies the maximum length of the password.
	// +kubebuilder:validation:Minimum=1
	MaxLength int `json:"maximumLength"`
}
