module github.com/crossplane/provider-aws

go 1.16

require (
	github.com/aws/aws-sdk-go v1.37.4
	github.com/aws/aws-sdk-go-v2 v0.23.0
	github.com/crossplane/crossplane-runtime v0.14.1-0.20210722005935-0b469fcc77cd
	github.com/crossplane/crossplane-tools v0.0.0-20210320162312-1baca298c527
	github.com/evanphx/json-patch v4.11.0+incompatible
	github.com/go-ini/ini v1.46.0
	github.com/google/go-cmp v0.5.5
	github.com/mitchellh/copystructure v1.0.0
	github.com/onsi/gomega v1.13.0
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.7.0
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	k8s.io/api v0.21.2
	k8s.io/apimachinery v0.21.2
	k8s.io/client-go v0.21.2
	sigs.k8s.io/controller-runtime v0.9.2
	sigs.k8s.io/controller-tools v0.4.0
)
