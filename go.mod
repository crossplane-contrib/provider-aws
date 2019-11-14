module github.com/crossplaneio/stack-aws

go 1.12

require (
	github.com/aws/aws-sdk-go-v2 v0.5.0
	github.com/crossplaneio/crossplane v0.4.0
	github.com/crossplaneio/crossplane-runtime v0.2.2
	github.com/crossplaneio/crossplane-tools v0.0.0-20191023215726-61fa1eff2a2e
	github.com/evanphx/json-patch v4.5.0+incompatible
	github.com/ghodss/yaml v1.0.0
	github.com/go-ini/ini v1.46.0
	github.com/google/go-cmp v0.3.1
	github.com/onsi/gomega v1.5.0
	github.com/pkg/errors v0.8.1
	github.com/stretchr/objx v0.2.0 // indirect
	github.com/stretchr/testify v1.4.0
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/ini.v1 v1.47.0 // indirect
	k8s.io/api v0.0.0-20190409021203-6e4e0e4f393b
	k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	sigs.k8s.io/controller-runtime v0.2.0
	sigs.k8s.io/controller-tools v0.2.1
)
