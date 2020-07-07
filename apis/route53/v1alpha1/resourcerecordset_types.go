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

package v1alpha1

import (
	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DNSRecordType defines the valid DNS Record Types that can be used.
type DNSRecordType string

const (
	// DNSRecordTypeSOA represents DNS SOA record type.
	DNSRecordTypeSOA DNSRecordType = "SOA"

	// DNSRecordTypeA represents DNS A record type.
	DNSRecordTypeA DNSRecordType = "A"

	// DNSRecordTypeTXT represents DNS TXT record type.
	DNSRecordTypeTXT DNSRecordType = "TXT"

	// DNSRecordTypeNS represents DNS NS record type.
	DNSRecordTypeNS DNSRecordType = "NS"

	// DNSRecordTypeCNAME represents DNS CNAME record type.
	DNSRecordTypeCNAME DNSRecordType = "CNAME"

	// DNSRecordTypeMX represents DNS MX record type.
	DNSRecordTypeMX DNSRecordType = "MX"

	// DNSRecordTypeNAPTR represents DNS NAPTR record type.
	DNSRecordTypeNAPTR DNSRecordType = "NAPTR"

	// DNSRecordTypePTR represents DNS PTR record type.
	DNSRecordTypePTR DNSRecordType = "PTR"

	// DNSRecordTypeSRV represents DNS SRV record type.
	DNSRecordTypeSRV DNSRecordType = "SRV"

	// DNSRecordTypeSPF represents DNS SPF record type.
	DNSRecordTypeSPF DNSRecordType = "SPF"

	// DNSRecordTypeAAAA represents DNS AAAA record type.
	DNSRecordTypeAAAA DNSRecordType = "AAAA"

	// DNSRecordTypeCAA represents DNS CAA record type.
	DNSRecordTypeCAA DNSRecordType = "CAA"
)

// ChangeAction defines the valid actions that can be performed on a ResourceRecordSet.
type ChangeAction string

const (
	// ChangeActionCreate represents a Resource Record CREATE operation.
	ChangeActionCreate ChangeAction = "CREATE"

	// ChangeActionDelete represents a Resource Record DELETE operation.
	ChangeActionDelete ChangeAction = "DELETE"

	// ChangeActionUpsert represents a Resource Record UPSERT operation.
	ChangeActionUpsert ChangeAction = "UPSERT"
)

// +kubebuilder:object:root=true

// ResourceRecordSet is a managed resource that represents an AWS Route53 Resource Record.
// +kubebuilder:printcolumn:name="TYPE",type="string",JSONPath=".spec.forProvider.type"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
type ResourceRecordSet struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ResourceRecordSetSpec `json:"spec"`
	// +optional
	Status ResourceRecordSetStatus `json:"status,omitempty"`
}

// ResourceRecordSetSpec defines the desired state of an AWS Route53 Resource Record.
type ResourceRecordSetSpec struct {
	runtimev1alpha1.ResourceSpec `json:",inline"`
	ForProvider                  ResourceRecordSetParameters `json:"forProvider"`
}

// ResourceRecordSetStatus represents the observed state of a ResourceRecordSet.
type ResourceRecordSetStatus struct {
	runtimev1alpha1.ResourceStatus `json:",inline"`
}

// ResourceRecordSetParameters define the desired state of an AWS Route53 Resource Record.
type ResourceRecordSetParameters struct {
	// AliasTarget holds information about the AWS resource, such
	// as a CloudFront distribution or an Amazon S3 bucket, that you want to route
	// traffic to.
	// +optional
	AliasTarget *AliasTarget `json:"aliasTarget,omitempty"`

	// Name of the record that you want to create, update, or delete.
	Name *string `json:"name"`

	// Failover let you add the Failover element to two resource record sets.
	// For one resource record set, you specify PRIMARY as the value for Failover;
	// for the other resource record set, you specify SECONDARY.
	// In addition, you include the HealthCheckId element and
	// specify the health check that you want Amazon Route 53 to perform for each
	// resource record set.
	// +optional
	Failover string `json:"failover,omitempty"`

	// GeoLocation lets you control how Amazon Route 53 responds to DNS queries
	// based on the geographic origin of the query.
	// +optional
	GeoLocation *GeoLocation `json:"geoLocation,omitempty"`

	// HealthCheckID let you return this resource record set in response
	// to a DNS query only when the status of a health check is healthy, include
	// the HealthCheckId element and specify the ID of the applicable health check.
	// +optional
	HealthCheckID *string `json:"healthCheckId,omitempty"`

	// Multivalue answer resource record sets only: To route traffic approximately
	// randomly to multiple resources, such as web servers, create one multivalue
	// answer record for each resource and specify true for MultiValueAnswer. Note
	// +optional
	MultiValueAnswer *bool `json:"multiValueAnswer,omitempty"`

	// Region is the Amazon EC2 Region where you created the resource that this
	// resource record set refers to. The resource typically is an AWS resource,
	// such as an EC2 instance or an ELB load balancer, and is referred to by
	// an IP address or a DNS domain name, depending on the record type.
	// +optional
	Region string `json:"region,omitempty"`

	// ResourceRecord holds the information about the resource records to act upon.
	ResourceRecords []ResourceRecord `json:"resourceRecords"`

	// TrafficPolicyInstanceId is the ID of the traffic
	// policy instance that Route 53 created this resource record set for.
	// +optional
	TrafficPolicyInstanceID *string `json:"trafficPolicyInstanceId,omitempty"`

	// SetIdentifier helps you differentiates among multiple resource record sets that have the same
	// combination of name and type, such as multiple weighted resource record sets
	// named acme.example.com that have a type of A.
	// +optional
	SetIdentifier *string `json:"setIdentifier,omitempty"`

	// Type represents the DNS record type
	Type *string `json:"type"`

	// The resource record cache time to live (TTL), in seconds.
	// +optional
	TTL *int64 `json:"ttl,omitempty"`

	// Weight determines the proportion of DNS queries that Amazon Route 53
	// responds to using the current resource record set.
	// +optional
	Weight *int64 `json:"weight,omitempty"`

	// ZoneID of the HostedZone in which you want to CREATE, CHANGE, or DELETE the Resource Record.
	// +optional
	ZoneID *string `json:"zoneId,omitempty"`

	// ZoneIDRef references a Zone to retrieves its ZoneId
	// +optional
	ZoneIDRef *runtimev1alpha1.Reference `json:"zoneIdRef,omitempty"`

	// ZoneIDSelector selects a reference to a Zone to retrieves its ZoneID
	// +optional
	ZoneIDSelector *runtimev1alpha1.Selector `json:"zoneIdSelector,omitempty"`
}

// AliasTarget holds information about the AWS resource, such
// as a CloudFront distribution or an Amazon S3 bucket, that you want to route
// traffic to.
type AliasTarget struct {

	// DNSName is the value that you specify depends on where you want to route queries
	// No omitempty
	DNSName *string `json:"dnsName,omitempty"`

	// EvaluateTargetHealth let you inherit the health of the referenced AWS resource,
	// such as an ELB load balancer or another resource record set in the hosted
	// zone.
	// +optional
	EvaluateTargetHealth *bool `json:"evaluateTargetHealth,omitempty"`

	// HostedZoneId of the AWS service where you want to route your traffic.
	// Note: These are pre determined Hosted Zone Ids that is provided by AWS and is different for each service and each region.
	HostedZoneID *string `json:"hostedZoneId,omitempty"`
}

// GeoLocation lets you control how Amazon Route 53 responds to DNS queries
// based on the geographic origin of the query.
type GeoLocation struct {

	// ContinentCode is the two-letter code for the continent.
	// Amazon Route 53 supports the following continent codes:
	//    * AF: Africa
	//    * AN: Antarctica
	//    * AS: Asia
	//    * EU: Europe
	//    * OC: Oceania
	//    * NA: North America
	//    * SA: South America
	// +optional
	ContinentCode *string `json:"continentCode,omitempty"`

	// CountryCode is the two-letter code for a country.
	// Amazon Route 53 uses the two-letter country codes that are specified in ISO
	// standard 3166-1 alpha-2 (https://en.wikipedia.org/wiki/ISO_3166-1_alpha-2).
	// +optional
	CountryCode *string `json:"countryCode,omitempty"`

	// SubdivisionCode is the two-letter code for a state of the United States.
	// +optional
	SubdivisionCode *string `json:"subdivisionCode,omitempty"`
}

// ResourceRecord holds the DNS value to be used for the record.
type ResourceRecord struct {
	// Value represents the current or new DNS record value(max 4,000 characters).
	// In the case of a DELETE action, if the current value does not match the actual value,
	// an error is returned.
	Value *string `json:"value"`
}

// +kubebuilder:object:root=true

// ResourceRecordSetList contains a list of ResourceRecordSet.
type ResourceRecordSetList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []ResourceRecordSet `json:"items"`
}
