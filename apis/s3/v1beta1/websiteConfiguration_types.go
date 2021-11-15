/*
Copyright 2020 The Crossplane Authors.

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

package v1beta1

// WebsiteConfiguration specifies website configuration parameters for an Amazon S3 bucket.
type WebsiteConfiguration struct {
	// The name of the error document for the website.
	// +optional
	ErrorDocument *ErrorDocument `json:"errorDocument,omitempty"`

	// The name of the index document for the website.
	// +optional
	IndexDocument *IndexDocument `json:"indexDocument,omitempty"`

	// The redirect behavior for every request to this bucket's website endpoint.
	// If you specify this property, you can't specify any other property.
	// +optional
	RedirectAllRequestsTo *RedirectAllRequestsTo `json:"redirectAllRequestsTo,omitempty"`

	// Rules that define when a redirect is applied and the redirect behavior.
	// +optional
	RoutingRules []RoutingRule `json:"routingRules,omitempty"`
}

// ErrorDocument is the error information.
type ErrorDocument struct {
	// The object key name to use when a 4XX class error occurs.
	Key string `json:"key"`
}

// IndexDocument is container for the Suffix element.
type IndexDocument struct {
	// A suffix that is appended to a request that is for a directory on the website
	// endpoint (for example,if the suffix is index.html and you make a request
	// to samplebucket/images/ the data that is returned will be for the object
	// with the key name images/index.html) The suffix must not be empty and must
	// not include a slash character.
	Suffix string `json:"suffix"`
}

// RedirectAllRequestsTo specifies the redirect behavior of all requests to a
// website endpoint of an Amazon S3 bucket.
type RedirectAllRequestsTo struct {
	// Name of the host where requests are redirected.
	HostName string `json:"hostName"`

	// Protocol to use when redirecting requests. The default is the protocol that
	// is used in the original request.
	// +kubebuilder:validation:Enum=http;https
	Protocol string `json:"protocol,omitempty"`
}

// RoutingRule specifies the redirect behavior and when a redirect is applied.
type RoutingRule struct {
	// A container for describing a condition that must be met for the specified
	// redirect to apply. For example, 1. If request is for pages in the /docs folder,
	// redirect to the /documents folder. 2. If request results in HTTP error 4xx,
	// redirect request to another host where you might process the error.
	// +optional
	Condition *Condition `json:"condition,omitempty"`

	// Container for redirect information. You can redirect requests to another
	// host, to another page, or with another protocol. In the event of an error,
	// you can specify a different error code to return.
	Redirect Redirect `json:"redirect"`
}

// Condition is a container for describing a condition that must be met for the specified
// redirect to apply. For example, 1. If request is for pages in the /docs folder,
// redirect to the /documents folder. 2. If request results in HTTP error 4xx,
// redirect request to another host where you might process the error.
type Condition struct {
	// The HTTP error code when the redirect is applied. In the event of an error,
	// if the error code equals this value, then the specified redirect is applied.
	// Required when parent element Condition is specified and sibling KeyPrefixEquals
	// is not specified. If both are specified, then both must be true for the redirect
	// to be applied.
	HTTPErrorCodeReturnedEquals *string `json:"httpErrorCodeReturnedEquals,omitempty"`

	// The object key name prefix when the redirect is applied. For example, to
	// redirect requests for ExamplePage.html, the key prefix will be ExamplePage.html.
	// To redirect request for all pages with the prefix docs/, the key prefix will
	// be /docs, which identifies all objects in the docs/ folder. Required when
	// the parent element Condition is specified and sibling HttpErrorCodeReturnedEquals
	// is not specified. If both conditions are specified, both must be true for
	// the redirect to be applied.
	KeyPrefixEquals *string `json:"keyPrefixEquals,omitempty"`
}

// Redirect specifies how requests are redirected. In the event of an error, you can
// specify a different error code to return.
type Redirect struct {
	// The host name to use in the redirect request.
	// +optional
	HostName *string `json:"keyPrefixEquals,omitempty"`

	// The HTTP redirect code to use on the response. Not required if one of the
	// siblings is present.
	HTTPRedirectCode *string `json:"httpRedirectCode,omitempty"`

	// Protocol to use when redirecting requests. The default is the protocol that
	// is used in the original request.
	Protocol string `json:"protocol,omitempty"`

	// The object key prefix to use in the redirect request. For example, to redirect
	// requests for all pages with prefix docs/ (objects in the docs/ folder) to
	// documents/, you can set a condition block with KeyPrefixEquals set to docs/
	// and in the Redirect set ReplaceKeyPrefixWith to /documents. Not required
	// if one of the siblings is present. Can be present only if ReplaceKeyWith
	// is not provided.
	ReplaceKeyPrefixWith *string `json:"replaceKeyPrefixWith,omitempty"`

	// The specific object key to use in the redirect request. For example, redirect
	// request to error.html. Not required if one of the siblings is present. Can
	// be present only if ReplaceKeyPrefixWith is not provided.
	ReplaceKeyWith *string `json:"replaceKeyWith,omitempty"`
}
