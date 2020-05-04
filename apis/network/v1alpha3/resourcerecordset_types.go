package v1alpha3

import (
	runtimev1alpha1 "github.com/crossplane/crossplane-runtime/apis/core/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DNSRecordType defines the valid DNS Record Types that can be used.
type DNSRecordType string

const (
	// DNSRecordTypeSoa represents DNS SOA record type.
	DNSRecordTypeSoa DNSRecordType = "SOA"

	// DNSRecordTypeA represents DNS A record type.
	DNSRecordTypeA DNSRecordType = "A"

	// DNSRecordTypeTxt represents DNS TXT record type.
	DNSRecordTypeTxt DNSRecordType = "TXT"

	// DNSRecordTypeNs represents DNS NS record type.
	DNSRecordTypeNs DNSRecordType = "NS"

	// DNSRecordTypeCname represents DNS CNAME record type.
	DNSRecordTypeCname DNSRecordType = "CNAME"

	// DNSRecordTypeMx represents DNS MX record type.
	DNSRecordTypeMx DNSRecordType = "MX"

	// DNSRecordTypeNaptr represents DNS NAPTR record type.
	DNSRecordTypeNaptr DNSRecordType = "NAPTR"

	// DNSRecordTypePtr represents DNS PTR record type.
	DNSRecordTypePtr DNSRecordType = "PTR"

	// DNSRecordTypeSrv represents DNS SRV record type.
	DNSRecordTypeSrv DNSRecordType = "SRV"

	// DNSRecordTypeSpf represents DNS SPF record type.
	DNSRecordTypeSpf DNSRecordType = "SPF"

	// DNSRecordTypeAaaa represents DNS AAAA record type.
	DNSRecordTypeAaaa DNSRecordType = "AAAA"

	// DNSRecordTypeCaa represents DNS CAA record type.
	DNSRecordTypeCaa DNSRecordType = "CAA"
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
// +kubebuilder:printcolumn:name="ARN",type="string",JSONPath=".status.atProvider.arn"
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster
type ResourceRecordSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ResourceRecordSetSpec   `json:"spec"`
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
	AtProvider                     ResourceRecordSetObservation `json:"atProvider"`
}

// ResourceRecord holds the DNS value to be used for the record.
type ResourceRecord struct {
	// The current or new DNS record value, not to exceed 4,000 characters. In the
	// case of a DELETE action, if the current value does not match the actual value,
	// an error is returned.
	Value *string `json:"Value"`
}

// ResourceRecordSetParameters define the desired state of an AWS Route53 Resource Record.
type ResourceRecordSetParameters struct {
	// Name of the record that you want to create, update, or delete.
	Name *string `json:"name"`

	// Type represents the DNS record type
	Type *string `json:"type"`

	// The resource record cache time to live (TTL), in seconds.
	TTL *int64 `json:"ttl,omitempty"`

	// ResourceRecord holds the information about the resource records to act upon.
	Records []string `json:"records"`

	// ZoneID of the HostedZone in which you want to CREATE, CHANGE, or DELETE the Resource Record.
	ZoneID *string `json:"zoneId,omitempty"`

	// ZoneIDRef references a VPC to and retrieves its vpcId
	// +optional
	ZoneIDRef *runtimev1alpha1.Reference `json:"zoneIdRef,omitempty"`

	// ZoneIDSelector selects a reference to a Zone to and retrieves its ZoneID
	// +optional
	ZoneIDSelector *runtimev1alpha1.Selector `json:"zoneIdSelector,omitempty"`
}

// ResourceRecordSetObservation keeps the state for the external resource.
type ResourceRecordSetObservation struct {
	// Name of the record that you want to create, update, or delete.
	Name *string `json:"Name"`

	Type *string `json:"Type"`

	// The resource record cache time to live (TTL), in seconds.
	TTL *int64 `json:"TTL"`

	// ResourceRecord holds the information about the resource records to act upon.
	ResourceRecords []*ResourceRecord `json:"ResourceRecords"`
}

// +kubebuilder:object:root=true

// ResourceRecordSetList contains a list of ResourceRecordSet.
type ResourceRecordSetList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []ResourceRecordSet `json:"items"`
}
