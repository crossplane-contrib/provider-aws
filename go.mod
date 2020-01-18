module github.com/crossplaneio/stack-aws

go 1.13

require (
	github.com/aws/aws-sdk-go-v2 v0.5.0
	github.com/crossplaneio/crossplane v0.7.0
	github.com/crossplaneio/crossplane-runtime v0.4.0
	github.com/crossplaneio/crossplane-tools v0.0.0-20191220202319-9033bd8a02ce
	github.com/evanphx/json-patch v4.5.0+incompatible
	github.com/ghodss/yaml v1.0.0
	github.com/go-ini/ini v1.46.0
	github.com/google/go-cmp v0.3.1
	github.com/onsi/gomega v1.7.0
	github.com/pkg/errors v0.8.1
	github.com/stretchr/testify v1.4.0
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/ini.v1 v1.47.0 // indirect
	k8s.io/api v0.17.0
	k8s.io/apimachinery v0.17.0
	k8s.io/client-go v0.17.0
	sigs.k8s.io/controller-runtime v0.4.0
	sigs.k8s.io/controller-tools v0.2.4
)
