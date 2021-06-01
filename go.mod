module github.com/crossplane/provider-aws

go 1.16

require (
	github.com/aws/aws-sdk-go v1.37.4
	github.com/aws/aws-sdk-go-v2 v0.23.0
	github.com/crossplane/crossplane-runtime v0.13.1-0.20210531122928-ded177829557
	github.com/crossplane/crossplane-tools v0.0.0-20210320162312-1baca298c527
	github.com/evanphx/json-patch v4.9.0+incompatible
	github.com/go-ini/ini v1.46.0
	github.com/google/go-cmp v0.5.2
	github.com/kr/text v0.2.0 // indirect
	github.com/mitchellh/copystructure v1.0.0
	github.com/niemeyer/pretty v0.0.0-20200227124842-a10e7caefd8e // indirect
	github.com/onsi/gomega v1.10.2
	github.com/pkg/errors v0.9.1
	github.com/stretchr/testify v1.6.1
	golang.org/x/tools v0.1.0 // indirect
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/check.v1 v1.0.0-20200227125254-8fa46927fb4f // indirect
	k8s.io/api v0.20.1
	k8s.io/apimachinery v0.20.1
	k8s.io/client-go v0.20.1
	sigs.k8s.io/controller-runtime v0.8.0
	sigs.k8s.io/controller-tools v0.4.0
)
