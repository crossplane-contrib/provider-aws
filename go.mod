module github.com/crossplane/provider-aws

go 1.13

require (
	github.com/aws/aws-sdk-go-v2 v0.23.0
	github.com/crossplane/crossplane v0.13.0
	github.com/crossplane/crossplane-runtime v0.10.0
	github.com/crossplane/crossplane-tools v0.0.0-20201007233256-88b291e145bb
	github.com/evanphx/json-patch v4.5.0+incompatible
	github.com/go-ini/ini v1.46.0
	github.com/google/go-cmp v0.5.0
	github.com/gopherjs/gopherjs v0.0.0-20180825215210-0210a2f0f73c // indirect
	github.com/jtolds/gls v4.2.1+incompatible // indirect
	github.com/mitchellh/copystructure v1.0.0
	github.com/onsi/gomega v1.10.1
	github.com/pkg/errors v0.9.1
	github.com/smartystreets/assertions v0.0.0-20180820201707-7c9eb446e3cf // indirect
	github.com/smartystreets/goconvey v0.0.0-20180222194500-ef6db91d284a // indirect
	github.com/stretchr/testify v1.5.1
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/ini.v1 v1.47.0 // indirect
	k8s.io/api v0.18.8
	k8s.io/apimachinery v0.18.8
	k8s.io/client-go v0.18.8
	sigs.k8s.io/controller-runtime v0.6.2
	sigs.k8s.io/controller-tools v0.2.4
)
