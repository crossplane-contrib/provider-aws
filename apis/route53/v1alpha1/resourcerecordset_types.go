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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	xpv1 "github.com/crossplane/crossplane-runtime/apis/common/v1"
)

// ResourceRecordSetParameters define the desired state of an AWS Route53 Resource Record.
type ResourceRecordSetParameters struct {
	// Alias resource record sets only: Information about the AWS resource, such
	// as a CloudFront distribution or an Amazon S3 bucket, that you want to route
	// traffic to.
	//
	// If you're creating resource records sets for a private hosted zone, note
	// the following:
	//
	//    * You can't create an alias resource record set in a private hosted zone
	//    to route traffic to a CloudFront distribution.
	//
	//    * Creating geolocation alias resource record sets or latency alias resource
	//    record sets in a private hosted zone is unsupported.
	//
	//    * For information about creating failover resource record sets in a private
	//    hosted zone, see Configuring Failover in a Private Hosted Zone (https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/dns-failover-private-hosted-zones.html)
	//    in the Amazon Route 53 Developer Guide.
	// +optional
	AliasTarget *AliasTarget `json:"aliasTarget,omitempty"`

	// Failover resource record sets only: To configure failover, you add the Failover
	// element to two resource record sets. For one resource record set, you specify
	// PRIMARY as the value for Failover; for the other resource record set, you
	// specify SECONDARY. In addition, you include the HealthCheckId element and
	// specify the health check that you want Amazon Route 53 to perform for each
	// resource record set.
	//
	// Except where noted, the following failover behaviors assume that you have
	// included the HealthCheckId element in both resource record sets:
	//
	//    * When the primary resource record set is healthy, Route 53 responds to
	//    DNS queries with the applicable value from the primary resource record
	//    set regardless of the health of the secondary resource record set.
	//
	//    * When the primary resource record set is unhealthy and the secondary
	//    resource record set is healthy, Route 53 responds to DNS queries with
	//    the applicable value from the secondary resource record set.
	//
	//    * When the secondary resource record set is unhealthy, Route 53 responds
	//    to DNS queries with the applicable value from the primary resource record
	//    set regardless of the health of the primary resource record set.
	//
	//    * If you omit the HealthCheckId element for the secondary resource record
	//    set, and if the primary resource record set is unhealthy, Route 53 always
	//    responds to DNS queries with the applicable value from the secondary resource
	//    record set. This is true regardless of the health of the associated endpoint.
	//
	// You can't create non-failover resource record sets that have the same values
	// for the Name and Type elements as failover resource record sets.
	//
	// For failover alias resource record sets, you must also include the EvaluateTargetHealth
	// element and set the value to true.
	//
	// For more information about configuring failover for Route 53, see the following
	// topics in the Amazon Route 53 Developer Guide:
	//
	//    * Route 53 Health Checks and DNS Failover (https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/dns-failover.html)
	//
	//    * Configuring Failover in a Private Hosted Zone (https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/dns-failover-private-hosted-zones.html)
	// +optional
	Failover string `json:"failover,omitempty"`

	// Geolocation resource record sets only: A complex type that lets you control
	// how Amazon Route 53 responds to DNS queries based on the geographic origin
	// of the query. For example, if you want all queries from Africa to be routed
	// to a web server with an IP address of 192.0.2.111, create a resource record
	// set with a Type of A and a ContinentCode of AF.
	//
	// Although creating geolocation and geolocation alias resource record sets
	// in a private hosted zone is allowed, it's not supported.
	//
	// If you create separate resource record sets for overlapping geographic regions
	// (for example, one resource record set for a continent and one for a country
	// on the same continent), priority goes to the smallest geographic region.
	// This allows you to route most queries for a continent to one resource and
	// to route queries for a country on that continent to a different resource.
	//
	// You can't create two geolocation resource record sets that specify the same
	// geographic location.
	//
	// The value * in the CountryCode element matches all geographic locations that
	// aren't specified in other geolocation resource record sets that have the
	// same values for the Name and Type elements.
	//
	// Geolocation works by mapping IP addresses to locations. However, some IP
	// addresses aren't mapped to geographic locations, so even if you create geolocation
	// resource record sets that cover all seven continents, Route 53 will receive
	// some DNS queries from locations that it can't identify. We recommend that
	// you create a resource record set for which the value of CountryCode is *.
	// Two groups of queries are routed to the resource that you specify in this
	// record: queries that come from locations for which you haven't created geolocation
	// resource record sets and queries from IP addresses that aren't mapped to
	// a location. If you don't create a * resource record set, Route 53 returns
	// a "no answer" response for queries from those locations.
	//
	// You can't create non-geolocation resource record sets that have the same
	// values for the Name and Type elements as geolocation resource record sets.
	// +optional
	GeoLocation *GeoLocation `json:"geoLocation,omitempty"`

	// If you want Amazon Route 53 to return this resource record set in response
	// to a DNS query only when the status of a health check is healthy, include
	// the HealthCheckId element and specify the ID of the applicable health check.
	//
	// Route 53 determines whether a resource record set is healthy based on one
	// of the following:
	//
	//    * By periodically sending a request to the endpoint that is specified
	//    in the health check
	//
	//    * By aggregating the status of a specified group of health checks (calculated
	//    health checks)
	//
	//    * By determining the current state of a CloudWatch alarm (CloudWatch metric
	//    health checks)
	//
	// Route 53 doesn't check the health of the endpoint that is specified in the
	// resource record set, for example, the endpoint specified by the IP address
	// in the Value element. When you add a HealthCheckId element to a resource
	// record set, Route 53 checks the health of the endpoint that you specified
	// in the health check.
	//
	// For more information, see the following topics in the Amazon Route 53 Developer
	// Guide:
	//
	//    * How Amazon Route 53 Determines Whether an Endpoint Is Healthy (https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/dns-failover-determining-health-of-endpoints.html)
	//
	//    * Route 53 Health Checks and DNS Failover (https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/dns-failover.html)
	//
	//    * Configuring Failover in a Private Hosted Zone (https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/dns-failover-private-hosted-zones.html)
	//
	// When to Specify HealthCheckId
	//
	// Specifying a value for HealthCheckId is useful only when Route 53 is choosing
	// between two or more resource record sets to respond to a DNS query, and you
	// want Route 53 to base the choice in part on the status of a health check.
	// Configuring health checks makes sense only in the following configurations:
	//
	//    * Non-alias resource record sets: You're checking the health of a group
	//    of non-alias resource record sets that have the same routing policy, name,
	//    and type (such as multiple weighted records named www.example.com with
	//    a type of A) and you specify health check IDs for all the resource record
	//    sets. If the health check status for a resource record set is healthy,
	//    Route 53 includes the record among the records that it responds to DNS
	//    queries with. If the health check status for a resource record set is
	//    unhealthy, Route 53 stops responding to DNS queries using the value for
	//    that resource record set. If the health check status for all resource
	//    record sets in the group is unhealthy, Route 53 considers all resource
	//    record sets in the group healthy and responds to DNS queries accordingly.
	//
	//    * Alias resource record sets: You specify the following settings: You
	//    set EvaluateTargetHealth to true for an alias resource record set in a
	//    group of resource record sets that have the same routing policy, name,
	//    and type (such as multiple weighted records named www.example.com with
	//    a type of A). You configure the alias resource record set to route traffic
	//    to a non-alias resource record set in the same hosted zone. You specify
	//    a health check ID for the non-alias resource record set. If the health
	//    check status is healthy, Route 53 considers the alias resource record
	//    set to be healthy and includes the alias record among the records that
	//    it responds to DNS queries with. If the health check status is unhealthy,
	//    Route 53 stops responding to DNS queries using the alias resource record
	//    set. The alias resource record set can also route traffic to a group of
	//    non-alias resource record sets that have the same routing policy, name,
	//    and type. In that configuration, associate health checks with all of the
	//    resource record sets in the group of non-alias resource record sets.
	//
	// Geolocation Routing
	//
	// For geolocation resource record sets, if an endpoint is unhealthy, Route
	// 53 looks for a resource record set for the larger, associated geographic
	// region. For example, suppose you have resource record sets for a state in
	// the United States, for the entire United States, for North America, and a
	// resource record set that has * for CountryCode is *, which applies to all
	// locations. If the endpoint for the state resource record set is unhealthy,
	// Route 53 checks for healthy resource record sets in the following order until
	// it finds a resource record set for which the endpoint is healthy:
	//
	//    * The United States
	//
	//    * North America
	//
	//    * The default resource record set
	//
	// Specifying the Health Check Endpoint by Domain Name
	//
	// If your health checks specify the endpoint only by domain name, we recommend
	// that you create a separate health check for each endpoint. For example, create
	// a health check for each HTTP server that is serving content for www.example.com.
	// For the value of FullyQualifiedDomainName, specify the domain name of the
	// server (such as us-east-2-www.example.com), not the name of the resource
	// record sets (www.example.com).
	//
	// Health check results will be unpredictable if you do the following:
	//
	//    * Create a health check that has the same value for FullyQualifiedDomainName
	//    as the name of a resource record set.
	//
	//    * Associate that health check with the resource record set.
	// +optional
	HealthCheckID *string `json:"healthCheckId,omitempty"`

	// Multivalue answer resource record sets only: To route traffic approximately
	// randomly to multiple resources, such as web servers, create one multivalue
	// answer record for each resource and specify true for MultiValueAnswer. Note
	// the following:
	//
	//    * If you associate a health check with a multivalue answer resource record
	//    set, Amazon Route 53 responds to DNS queries with the corresponding IP
	//    address only when the health check is healthy.
	//
	//    * If you don't associate a health check with a multivalue answer record,
	//    Route 53 always considers the record to be healthy.
	//
	//    * Route 53 responds to DNS queries with up to eight healthy records; if
	//    you have eight or fewer healthy records, Route 53 responds to all DNS
	//    queries with all the healthy records.
	//
	//    * If you have more than eight healthy records, Route 53 responds to different
	//    DNS resolvers with different combinations of healthy records.
	//
	//    * When all records are unhealthy, Route 53 responds to DNS queries with
	//    up to eight unhealthy records.
	//
	//    * If a resource becomes unavailable after a resolver caches a response,
	//    client software typically tries another of the IP addresses in the response.
	//
	// You can't create multivalue answer alias records.
	// +optional
	MultiValueAnswer *bool `json:"multiValueAnswer,omitempty"`

	// Latency-based resource record sets only: The Amazon EC2 Region where you
	// created the resource that this resource record set refers to. The resource
	// typically is an AWS resource, such as an EC2 instance or an ELB load balancer,
	// and is referred to by an IP address or a DNS domain name, depending on the
	// record type.
	//
	// Although creating latency and latency alias resource record sets in a private
	// hosted zone is allowed, it's not supported.
	//
	// When Amazon Route 53 receives a DNS query for a domain name and type for
	// which you have created latency resource record sets, Route 53 selects the
	// latency resource record set that has the lowest latency between the end user
	// and the associated Amazon EC2 Region. Route 53 then returns the value that
	// is associated with the selected resource record set.
	//
	// Note the following:
	//
	//    * You can only specify one ResourceRecord per latency resource record
	//    set.
	//
	//    * You can only create one latency resource record set for each Amazon
	//    EC2 Region.
	//
	//    * You aren't required to create latency resource record sets for all Amazon
	//    EC2 Regions. Route 53 will choose the region with the best latency from
	//    among the regions that you create latency resource record sets for.
	//
	//    * You can't create non-latency resource record sets that have the same
	//    values for the Name and Type elements as latency resource record sets.
	// +optional
	Region string `json:"region,omitempty"`

	// Information about the resource records to act upon.
	//
	// If you're creating an alias resource record set, omit ResourceRecords.
	ResourceRecords []ResourceRecord `json:"resourceRecords,omitempty"`

	// Resource record sets that have a routing policy other than simple: An identifier
	// that differentiates among multiple resource record sets that have the same
	// combination of name and type, such as multiple weighted resource record sets
	// named acme.example.com that have a type of A. In a group of resource record
	// sets that have the same name and type, the value of SetIdentifier must be
	// unique for each resource record set.
	//
	// For information about routing policies, see Choosing a Routing Policy (https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/routing-policy.html)
	// in the Amazon Route 53 Developer Guide.
	// +optional
	SetIdentifier *string `json:"setIdentifier,omitempty"`

	// The resource record cache time to live (TTL), in seconds. Note the following:
	//
	//    * If you're creating or updating an alias resource record set, omit TTL.
	//    Amazon Route 53 uses the value of TTL for the alias target.
	//
	//    * If you're associating this resource record set with a health check (if
	//    you're adding a HealthCheckId element), we recommend that you specify
	//    a TTL of 60 seconds or less so clients respond quickly to changes in health
	//    status.
	//
	//    * All of the resource record sets in a group of weighted resource record
	//    sets must have the same value for TTL.
	//
	//    * If a group of weighted resource record sets includes one or more weighted
	//    alias resource record sets for which the alias target is an ELB load balancer,
	//    we recommend that you specify a TTL of 60 seconds for all of the non-alias
	//    weighted resource record sets that have the same name and type. Values
	//    other than 60 seconds (the TTL for load balancers) will change the effect
	//    of the values that you specify for Weight.
	// +optional
	TTL *int64 `json:"ttl,omitempty"`

	// When you create a traffic policy instance, Amazon Route 53 automatically
	// creates a resource record set. TrafficPolicyInstanceId is the ID of the traffic
	// policy instance that Route 53 created this resource record set for.
	//
	// To delete the resource record set that is associated with a traffic policy
	// instance, use DeleteTrafficPolicyInstance. Route 53 will delete the resource
	// record set automatically. If you delete the resource record set by using
	// ChangeResourceRecordSets, Route 53 doesn't automatically delete the traffic
	// policy instance, and you'll continue to be charged for it even though it's
	// no longer in use.
	// +optional
	TrafficPolicyInstanceID *string `json:"trafficPolicyInstanceId,omitempty"`

	// The DNS record type. For information about different record types and how
	// data is encoded for them, see Supported DNS Resource Record Types (https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/ResourceRecordTypes.html)
	// in the Amazon Route 53 Developer Guide.
	//
	// Valid values for basic resource record sets: A | AAAA | CAA | CNAME | MX
	// | NAPTR | NS | PTR | SOA | SPF | SRV | TXT
	//
	// Values for weighted, latency, geolocation, and failover resource record sets:
	// A | AAAA | CAA | CNAME | MX | NAPTR | PTR | SPF | SRV | TXT. When creating
	// a group of weighted, latency, geolocation, or failover resource record sets,
	// specify the same value for all of the resource record sets in the group.
	//
	// Valid values for multivalue answer resource record sets: A | AAAA | MX |
	// NAPTR | PTR | SPF | SRV | TXT
	//
	// SPF records were formerly used to verify the identity of the sender of email
	// messages. However, we no longer recommend that you create resource record
	// sets for which the value of Type is SPF. RFC 7208, Sender Policy Framework
	// (SPF) for Authorizing Use of Domains in Email, Version 1, has been updated
	// to say, "...[I]ts existence and mechanism defined in [RFC4408] have led to
	// some interoperability issues. Accordingly, its use is no longer appropriate
	// for SPF version 1; implementations are not to use it." In RFC 7208, see section
	// 14.1, The SPF DNS Record Type (http://tools.ietf.org/html/rfc7208#section-14.1).
	//
	// Values for alias resource record sets:
	//
	//    * Amazon API Gateway custom regional APIs and edge-optimized APIs: A
	//
	//    * CloudFront distributions: A If IPv6 is enabled for the distribution,
	//    create two resource record sets to route traffic to your distribution,
	//    one with a value of A and one with a value of AAAA.
	//
	//    * Amazon API Gateway environment that has a regionalized subdomain: A
	//
	//    * ELB load balancers: A | AAAA
	//
	//    * Amazon S3 buckets: A
	//
	//    * Amazon Virtual Private Cloud interface VPC endpoints A
	//
	//    * Another resource record set in this hosted zone: Specify the type of
	//    the resource record set that you're creating the alias for. All values
	//    are supported except NS and SOA. If you're creating an alias record that
	//    has the same name as the hosted zone (known as the zone apex), you can't
	//    route traffic to a record for which the value of Type is CNAME. This is
	//    because the alias record must have the same type as the record you're
	//    routing traffic to, and creating a CNAME record for the zone apex isn't
	//    supported even for an alias record.
	Type string `json:"type"`

	// Weighted resource record sets only: Among resource record sets that have
	// the same combination of DNS name and type, a value that determines the proportion
	// of DNS queries that Amazon Route 53 responds to using the current resource
	// record set. Route 53 calculates the sum of the weights for the resource record
	// sets that have the same combination of DNS name and type. Route 53 then responds
	// to queries based on the ratio of a resource's weight to the total. Note the
	// following:
	//
	//    * You must specify a value for the Weight element for every weighted resource
	//    record set.
	//
	//    * You can only specify one ResourceRecord per weighted resource record
	//    set.
	//
	//    * You can't create latency, failover, or geolocation resource record sets
	//    that have the same values for the Name and Type elements as weighted resource
	//    record sets.
	//
	//    * You can create a maximum of 100 weighted resource record sets that have
	//    the same values for the Name and Type elements.
	//
	//    * For weighted (but not weighted alias) resource record sets, if you set
	//    Weight to 0 for a resource record set, Route 53 never responds to queries
	//    with the applicable value for that resource record set. However, if you
	//    set Weight to 0 for all resource record sets that have the same combination
	//    of DNS name and type, traffic is routed to all resources with equal probability.
	//    The effect of setting Weight to 0 is different when you associate health
	//    checks with weighted resource record sets. For more information, see Options
	//    for Configuring Route 53 Active-Active and Active-Passive Failover (https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/dns-failover-configuring-options.html)
	//    in the Amazon Route 53 Developer Guide.
	// +optional
	Weight *int64 `json:"weight,omitempty"`

	// ZoneID is the ID of the hosted zone that contains the resource record sets
	// that you want to change.
	ZoneID *string `json:"zoneId,omitempty"`

	// ZoneIDRef references a Zone to retrieves its ZoneId
	// +optional
	ZoneIDRef *xpv1.Reference `json:"zoneIdRef,omitempty"`

	// ZoneIDSelector selects a reference to a Zone to retrieves its ZoneID
	// +optional
	ZoneIDSelector *xpv1.Selector `json:"zoneIdSelector,omitempty"`
}

// AliasTarget : Alias resource record sets only. Information about the AWS
// resource, such as a CloudFront distribution or an Amazon S3 bucket, that you
// want to route traffic to.
//
// When creating resource record sets for a private hosted zone, note the following:
//
//   - Creating geolocation alias resource record sets or latency alias resource
//     record sets in a private hosted zone is unsupported.
//
//   - For information about creating failover resource record sets in a private
//     hosted zone, see Configuring Failover in a Private Hosted Zone (https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/dns-failover-private-hosted-zones.html).
type AliasTarget struct {

	// Alias resource record sets only: The value that you specify depends on where
	// you want to route queries:
	//
	// Amazon API Gateway custom regional APIs and edge-optimized APIs
	//
	// Specify the applicable domain name for your API. You can get the applicable
	// value using the AWS CLI command get-domain-names (https://docs.aws.amazon.com/cli/latest/reference/apigateway/get-domain-names.html):
	//
	//    * For regional APIs, specify the value of regionalDomainName.
	//
	//    * For edge-optimized APIs, specify the value of distributionDomainName.
	//    This is the name of the associated CloudFront distribution, such as da1b2c3d4e5.cloudfront.net.
	//
	// The name of the record that you're creating must match a custom domain name
	// for your API, such as api.example.com.
	//
	// Amazon Virtual Private Cloud interface VPC endpoint
	//
	// Enter the API endpoint for the interface endpoint, such as vpce-123456789abcdef01-example-us-east-1a.elasticloadbalancing.us-east-1.vpce.amazonaws.com.
	// For edge-optimized APIs, this is the domain name for the corresponding CloudFront
	// distribution. You can get the value of DnsName using the AWS CLI command
	// describe-vpc-endpoints (https://docs.aws.amazon.com/cli/latest/reference/ec2/describe-vpc-endpoints.html).
	//
	// CloudFront distribution
	//
	// Specify the domain name that CloudFront assigned when you created your distribution.
	//
	// Your CloudFront distribution must include an alternate domain name that matches
	// the name of the resource record set. For example, if the name of the resource
	// record set is acme.example.com, your CloudFront distribution must include
	// acme.example.com as one of the alternate domain names. For more information,
	// see Using Alternate Domain Names (CNAMEs) (https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/CNAMEs.html)
	// in the Amazon CloudFront Developer Guide.
	//
	// You can't create a resource record set in a private hosted zone to route
	// traffic to a CloudFront distribution.
	//
	// For failover alias records, you can't specify a CloudFront distribution for
	// both the primary and secondary records. A distribution must include an alternate
	// domain name that matches the name of the record. However, the primary and
	// secondary records have the same name, and you can't include the same alternate
	// domain name in more than one distribution.
	//
	// Elastic Beanstalk environment
	//
	// If the domain name for your Elastic Beanstalk environment includes the region
	// that you deployed the environment in, you can create an alias record that
	// routes traffic to the environment. For example, the domain name my-environment.us-west-2.elasticbeanstalk.com
	// is a regionalized domain name.
	//
	// For environments that were created before early 2016, the domain name doesn't
	// include the region. To route traffic to these environments, you must create
	// a CNAME record instead of an alias record. Note that you can't create a CNAME
	// record for the root domain name. For example, if your domain name is example.com,
	// you can create a record that routes traffic for acme.example.com to your
	// Elastic Beanstalk environment, but you can't create a record that routes
	// traffic for example.com to your Elastic Beanstalk environment.
	//
	// For Elastic Beanstalk environments that have regionalized subdomains, specify
	// the CNAME attribute for the environment. You can use the following methods
	// to get the value of the CNAME attribute:
	//
	//    * AWS Management Console: For information about how to get the value by
	//    using the console, see Using Custom Domains with AWS Elastic Beanstalk
	//    (https://docs.aws.amazon.com/elasticbeanstalk/latest/dg/customdomains.html)
	//    in the AWS Elastic Beanstalk Developer Guide.
	//
	//    * Elastic Beanstalk API: Use the DescribeEnvironments action to get the
	//    value of the CNAME attribute. For more information, see DescribeEnvironments
	//    (https://docs.aws.amazon.com/elasticbeanstalk/latest/api/API_DescribeEnvironments.html)
	//    in the AWS Elastic Beanstalk API Reference.
	//
	//    * AWS CLI: Use the describe-environments command to get the value of the
	//    CNAME attribute. For more information, see describe-environments (https://docs.aws.amazon.com/cli/latest/reference/elasticbeanstalk/describe-environments.html)
	//    in the AWS CLI Command Reference.
	//
	// ELB load balancer
	//
	// Specify the DNS name that is associated with the load balancer. Get the DNS
	// name by using the AWS Management Console, the ELB API, or the AWS CLI.
	//
	//    * AWS Management Console: Go to the EC2 page, choose Load Balancers in
	//    the navigation pane, choose the load balancer, choose the Description
	//    tab, and get the value of the DNS name field. If you're routing traffic
	//    to a Classic Load Balancer, get the value that begins with dualstack.
	//    If you're routing traffic to another type of load balancer, get the value
	//    that applies to the record type, A or AAAA.
	//
	//    * Elastic Load Balancing API: Use DescribeLoadBalancers to get the value
	//    of DNSName. For more information, see the applicable guide: Classic Load
	//    Balancers: DescribeLoadBalancers (https://docs.aws.amazon.com/elasticloadbalancing/2012-06-01/APIReference/API_DescribeLoadBalancers.html)
	//    Application and Network Load Balancers: DescribeLoadBalancers (https://docs.aws.amazon.com/elasticloadbalancing/latest/APIReference/API_DescribeLoadBalancers.html)
	//
	//    * AWS CLI: Use describe-load-balancers to get the value of DNSName. For
	//    more information, see the applicable guide: Classic Load Balancers: describe-load-balancers
	//    (http://docs.aws.amazon.com/cli/latest/reference/elb/describe-load-balancers.html)
	//    Application and Network Load Balancers: describe-load-balancers (http://docs.aws.amazon.com/cli/latest/reference/elbv2/describe-load-balancers.html)
	//
	// AWS Global Accelerator accelerator
	//
	// Specify the DNS name for your accelerator:
	//
	//    * Global Accelerator API: To get the DNS name, use DescribeAccelerator
	//    (https://docs.aws.amazon.com/global-accelerator/latest/api/API_DescribeAccelerator.html).
	//
	//    * AWS CLI: To get the DNS name, use describe-accelerator (https://docs.aws.amazon.com/cli/latest/reference/globalaccelerator/describe-accelerator.html).
	//
	// Amazon S3 bucket that is configured as a static website
	//
	// Specify the domain name of the Amazon S3 website endpoint that you created
	// the bucket in, for example, s3-website.us-east-2.amazonaws.com. For more
	// information about valid values, see the table Amazon S3 Website Endpoints
	// (https://docs.aws.amazon.com/general/latest/gr/s3.html#s3_website_region_endpoints)
	// in the Amazon Web Services General Reference. For more information about
	// using S3 buckets for websites, see Getting Started with Amazon Route 53 (https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/getting-started.html)
	// in the Amazon Route 53 Developer Guide.
	//
	// Another Route 53 resource record set
	//
	// Specify the value of the Name element for a resource record set in the current
	// hosted zone.
	//
	// If you're creating an alias record that has the same name as the hosted zone
	// (known as the zone apex), you can't specify the domain name for a record
	// for which the value of Type is CNAME. This is because the alias record must
	// have the same type as the record that you're routing traffic to, and creating
	// a CNAME record for the zone apex isn't supported even for an alias record.
	DNSName string `json:"dnsName"`

	// Applies only to alias, failover alias, geolocation alias, latency alias,
	// and weighted alias resource record sets: When EvaluateTargetHealth is true,
	// an alias resource record set inherits the health of the referenced AWS resource,
	// such as an ELB load balancer or another resource record set in the hosted
	// zone.
	//
	// Note the following:
	//
	// CloudFront distributions
	//
	// You can't set EvaluateTargetHealth to true when the alias target is a CloudFront
	// distribution.
	//
	// Elastic Beanstalk environments that have regionalized subdomains
	//
	// If you specify an Elastic Beanstalk environment in DNSName and the environment
	// contains an ELB load balancer, Elastic Load Balancing routes queries only
	// to the healthy Amazon EC2 instances that are registered with the load balancer.
	// (An environment automatically contains an ELB load balancer if it includes
	// more than one Amazon EC2 instance.) If you set EvaluateTargetHealth to true
	// and either no Amazon EC2 instances are healthy or the load balancer itself
	// is unhealthy, Route 53 routes queries to other available resources that are
	// healthy, if any.
	//
	// If the environment contains a single Amazon EC2 instance, there are no special
	// requirements.
	//
	// ELB load balancers
	//
	// Health checking behavior depends on the type of load balancer:
	//
	//    * Classic Load Balancers: If you specify an ELB Classic Load Balancer
	//    in DNSName, Elastic Load Balancing routes queries only to the healthy
	//    Amazon EC2 instances that are registered with the load balancer. If you
	//    set EvaluateTargetHealth to true and either no EC2 instances are healthy
	//    or the load balancer itself is unhealthy, Route 53 routes queries to other
	//    resources.
	//
	//    * Application and Network Load Balancers: If you specify an ELB Application
	//    or Network Load Balancer and you set EvaluateTargetHealth to true, Route
	//    53 routes queries to the load balancer based on the health of the target
	//    groups that are associated with the load balancer: For an Application
	//    or Network Load Balancer to be considered healthy, every target group
	//    that contains targets must contain at least one healthy target. If any
	//    target group contains only unhealthy targets, the load balancer is considered
	//    unhealthy, and Route 53 routes queries to other resources. A target group
	//    that has no registered targets is considered unhealthy.
	//
	// When you create a load balancer, you configure settings for Elastic Load
	// Balancing health checks; they're not Route 53 health checks, but they perform
	// a similar function. Do not create Route 53 health checks for the EC2 instances
	// that you register with an ELB load balancer.
	//
	// S3 buckets
	//
	// There are no special requirements for setting EvaluateTargetHealth to true
	// when the alias target is an S3 bucket.
	//
	// Other records in the same hosted zone
	//
	// If the AWS resource that you specify in DNSName is a record or a group of
	// records (for example, a group of weighted records) but is not another alias
	// record, we recommend that you associate a health check with all of the records
	// in the alias target. For more information, see What Happens When You Omit
	// Health Checks? (https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/dns-failover-complex-configs.html#dns-failover-complex-configs-hc-omitting)
	// in the Amazon Route 53 Developer Guide.
	//
	// For more information and examples, see Amazon Route 53 Health Checks and
	// DNS Failover (https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/dns-failover.html)
	// in the Amazon Route 53 Developer Guide.
	EvaluateTargetHealth bool `json:"evaluateTargetHealth"`

	// Alias resource records sets only: The value used depends on where you want
	// to route traffic:
	//
	// Amazon API Gateway custom regional APIs and edge-optimized APIs
	//
	// Specify the hosted zone ID for your API. You can get the applicable value
	// using the AWS CLI command get-domain-names (https://docs.aws.amazon.com/cli/latest/reference/apigateway/get-domain-names.html):
	//
	//    * For regional APIs, specify the value of regionalHostedZoneId.
	//
	//    * For edge-optimized APIs, specify the value of distributionHostedZoneId.
	//
	// Amazon Virtual Private Cloud interface VPC endpoint
	//
	// Specify the hosted zone ID for your interface endpoint. You can get the value
	// of HostedZoneId using the AWS CLI command describe-vpc-endpoints (https://docs.aws.amazon.com/cli/latest/reference/ec2/describe-vpc-endpoints.html).
	//
	// CloudFront distribution
	//
	// Specify Z2FDTNDATAQYW2.
	//
	// Alias resource record sets for CloudFront can't be created in a private zone.
	//
	// Elastic Beanstalk environment
	//
	// Specify the hosted zone ID for the region that you created the environment
	// in. The environment must have a regionalized subdomain. For a list of regions
	// and the corresponding hosted zone IDs, see AWS Elastic Beanstalk (https://docs.aws.amazon.com/general/latest/gr/rande.html#elasticbeanstalk_region)
	// in the "AWS Service Endpoints" chapter of the Amazon Web Services General
	// Reference.
	//
	// ELB load balancer
	//
	// Specify the value of the hosted zone ID for the load balancer. Use the following
	// methods to get the hosted zone ID:
	//
	//    * Service Endpoints (https://docs.aws.amazon.com/general/latest/gr/elb.html)
	//    table in the "Elastic Load Balancing Endpoints and Quotas" topic in the
	//    Amazon Web Services General Reference: Use the value that corresponds
	//    with the region that you created your load balancer in. Note that there
	//    are separate columns for Application and Classic Load Balancers and for
	//    Network Load Balancers.
	//
	//    * AWS Management Console: Go to the Amazon EC2 page, choose Load Balancers
	//    in the navigation pane, select the load balancer, and get the value of
	//    the Hosted zone field on the Description tab.
	//
	//    * Elastic Load Balancing API: Use DescribeLoadBalancers to get the applicable
	//    value. For more information, see the applicable guide: Classic Load Balancers:
	//    Use DescribeLoadBalancers (https://docs.aws.amazon.com/elasticloadbalancing/2012-06-01/APIReference/API_DescribeLoadBalancers.html)
	//    to get the value of CanonicalHostedZoneNameId. Application and Network
	//    Load Balancers: Use DescribeLoadBalancers (https://docs.aws.amazon.com/elasticloadbalancing/latest/APIReference/API_DescribeLoadBalancers.html)
	//    to get the value of CanonicalHostedZoneId.
	//
	//    * AWS CLI: Use describe-load-balancers to get the applicable value. For
	//    more information, see the applicable guide: Classic Load Balancers: Use
	//    describe-load-balancers (http://docs.aws.amazon.com/cli/latest/reference/elb/describe-load-balancers.html)
	//    to get the value of CanonicalHostedZoneNameId. Application and Network
	//    Load Balancers: Use describe-load-balancers (http://docs.aws.amazon.com/cli/latest/reference/elbv2/describe-load-balancers.html)
	//    to get the value of CanonicalHostedZoneId.
	//
	// AWS Global Accelerator accelerator
	//
	// Specify Z2BJ6XQ5FK7U4H.
	//
	// An Amazon S3 bucket configured as a static website
	//
	// Specify the hosted zone ID for the region that you created the bucket in.
	// For more information about valid values, see the table Amazon S3 Website
	// Endpoints (https://docs.aws.amazon.com/general/latest/gr/s3.html#s3_website_region_endpoints)
	// in the Amazon Web Services General Reference.
	//
	// Another Route 53 resource record set in your hosted zone
	//
	// Specify the hosted zone ID of your hosted zone. (An alias resource record
	// set can't reference a resource record set in a different hosted zone.)
	HostedZoneID string `json:"hostedZoneId"`
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
	// Constraint: Specifying ContinentCode with either CountryCode or SubdivisionCode
	// returns an InvalidInput error.
	// +optional
	ContinentCode *string `json:"continentCode,omitempty"`

	// For geolocation resource record sets, the two-letter code for a country.
	//
	// Amazon Route 53 uses the two-letter country codes that are specified in ISO
	// standard 3166-1 alpha-2 (https://en.wikipedia.org/wiki/ISO_3166-1_alpha-2).
	// +optional
	CountryCode *string `json:"countryCode,omitempty"`

	// For geolocation resource record sets, the two-letter code for a state of
	// the United States. Route 53 doesn't support any other values for SubdivisionCode.
	// For a list of state abbreviations, see Appendix B: Two–Letter State and
	// Possession Abbreviations (https://pe.usps.com/text/pub28/28apb.htm) on the
	// United States Postal Service website.
	//
	// If you specify subdivision code, you must also specify US for CountryCode.
	// +optional
	SubdivisionCode *string `json:"subdivisionCode,omitempty"`
}

// ResourceRecord holds the DNS value to be used for the record.
type ResourceRecord struct {
	// The current or new DNS record value, not to exceed 4,000 characters. In the
	// case of a DELETE action, if the current value does not match the actual value,
	// an error is returned. For descriptions about how to format Value for different
	// record types, see Supported DNS Resource Record Types (https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/ResourceRecordTypes.html)
	// in the Amazon Route 53 Developer Guide.
	//
	// You can specify more than one value for all record types except CNAME and
	// SOA.
	//
	// If you're creating an alias resource record set, omit Value.
	Value string `json:"value"`
}

// +kubebuilder:object:root=true

// ResourceRecordSet is a managed resource that represents an AWS Route53 Resource Record.
// +kubebuilder:printcolumn:name="READY",type="string",JSONPath=".status.conditions[?(@.type=='Ready')].status"
// +kubebuilder:printcolumn:name="SYNCED",type="string",JSONPath=".status.conditions[?(@.type=='Synced')].status"
// +kubebuilder:printcolumn:name="EXTERNAL-NAME",type="string",JSONPath=".metadata.annotations.crossplane\\.io/external-name"
// +kubebuilder:printcolumn:name="TYPE",type="string",JSONPath=".spec.forProvider.type"
// +kubebuilder:printcolumn:name="AGE",type="date",JSONPath=".metadata.creationTimestamp"
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster,categories={crossplane,managed,aws}
type ResourceRecordSet struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ResourceRecordSetSpec   `json:"spec"`
	Status ResourceRecordSetStatus `json:"status,omitempty"`
}

// ResourceRecordSetSpec defines the desired state of an AWS Route53 Resource Record.
type ResourceRecordSetSpec struct {
	xpv1.ResourceSpec `json:",inline"`
	ForProvider       ResourceRecordSetParameters `json:"forProvider"`
}

// ResourceRecordSetStatus represents the observed state of a ResourceRecordSet.
type ResourceRecordSetStatus struct {
	xpv1.ResourceStatus `json:",inline"`
}

// +kubebuilder:object:root=true

// ResourceRecordSetList contains a list of ResourceRecordSet.
type ResourceRecordSetList struct {
	metav1.TypeMeta `json:",inline"`
	// +optional
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []ResourceRecordSet `json:"items"`
}
