package v1beta1

// PaymentConfiguration specifies who pays for the download and request fees.
type PaymentConfiguration struct {
	// Payer is a required field, detailing who pays
	// Valid values are "Requester" and "BucketOwner"
	Payer string `json:"payer"`
}
