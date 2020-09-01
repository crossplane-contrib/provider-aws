package v1beta1

// CORSConfiguration describes the cross-origin access configuration for objects
// in an Amazon S3 bucket. For more information, see Enabling Cross-Origin Resource Sharing
// (https://docs.aws.amazon.com/AmazonS3/latest/dev/cors.html) in the Amazon
// Simple Storage Service Developer Guide.
type CORSConfiguration struct {
	// A set of origins and methods (cross-origin access that you want to allow).
	// You can add up to 100 rules to the configuration.
	CORSRules []CORSRule `json:"corsRules"`
}

// CORSRule specifies a cross-origin access rule for an Amazon S3 bucket.
type CORSRule struct {
	// Headers that are specified in the Access-Control-Request-Headers header.
	// These headers are allowed in a preflight OPTIONS request. In response to
	// any preflight OPTIONS request, Amazon S3 returns any requested headers that
	// are allowed.
	// +optional
	AllowedHeaders []string `json:"allowedHeaders,omitempty"`

	// An HTTP method that you allow the origin to execute. Valid values are GET,
	// PUT, HEAD, POST, and DELETE.
	AllowedMethods []string `json:"allowedMethods"`

	// One or more origins you want customers to be able to access the bucket from.
	AllowedOrigins []string `json:"allowedOrigins"`

	// One or more headers in the response that you want customers to be able to
	// access from their applications (for example, from a JavaScript XMLHttpRequest
	// object).
	// +optional
	ExposeHeaders []string `json:"exposeHeaders,omitempty"`

	// The time in seconds that your browser is to cache the preflight response
	// for the specified resource.
	// +optional
	MaxAgeSeconds *int64 `json:"maxAgeSeconds,omitempty"`
}
