module github.com/crossplane-contrib/provider-aws

go 1.18

require (
	github.com/aws/aws-sdk-go v1.44.174
	github.com/aws/aws-sdk-go-v2 v1.16.16
	github.com/aws/aws-sdk-go-v2/config v1.11.1
	github.com/aws/aws-sdk-go-v2/credentials v1.6.5
	github.com/aws/aws-sdk-go-v2/service/acm v1.10.0
	github.com/aws/aws-sdk-go-v2/service/acmpca v1.12.0
	github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider v1.17.3
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.26.0
	github.com/aws/aws-sdk-go-v2/service/ecr v1.12.0
	github.com/aws/aws-sdk-go-v2/service/eks v1.22.1
	github.com/aws/aws-sdk-go-v2/service/elasticache v1.16.0
	github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing v1.10.0
	github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2 v1.10.0
	github.com/aws/aws-sdk-go-v2/service/iam v1.14.0
	github.com/aws/aws-sdk-go-v2/service/lambda v1.21.1
	github.com/aws/aws-sdk-go-v2/service/rds v1.14.0
	github.com/aws/aws-sdk-go-v2/service/redshift v1.17.0
	github.com/aws/aws-sdk-go-v2/service/route53 v1.15.0
	github.com/aws/aws-sdk-go-v2/service/route53resolver v1.10.2
	github.com/aws/aws-sdk-go-v2/service/s3 v1.22.0
	github.com/aws/aws-sdk-go-v2/service/sns v1.13.0
	github.com/aws/aws-sdk-go-v2/service/sqs v1.14.0
	github.com/aws/aws-sdk-go-v2/service/sts v1.12.0
	github.com/aws/smithy-go v1.13.3
	github.com/barkimedes/go-deepcopy v0.0.0-20220514131651-17c30cfc62df
	github.com/crossplane/crossplane-runtime v0.19.0
	github.com/crossplane/crossplane-tools v0.0.0-20220310165030-1f43fc12793e
	github.com/evanphx/json-patch v4.12.0+incompatible
	github.com/go-ini/ini v1.67.0
	github.com/golang/mock v1.5.0
	github.com/google/go-cmp v0.5.9
	github.com/mattbaird/jsonpatch v0.0.0-20200820163806-098863c1fc24
	github.com/mitchellh/copystructure v1.0.0
	github.com/onsi/gomega v1.24.2
	github.com/pkg/errors v0.9.1
	go.uber.org/zap v1.24.0
	golang.org/x/net v0.7.0
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	k8s.io/api v0.26.1
	k8s.io/apimachinery v0.26.1
	k8s.io/client-go v0.26.1
	k8s.io/utils v0.0.0-20221128185143-99ec85e7a448
	sigs.k8s.io/controller-runtime v0.14.1
	sigs.k8s.io/controller-tools v0.11.1
	sigs.k8s.io/yaml v1.3.0
)

require (
	github.com/alecthomas/template v0.0.0-20190718012654-fb15b899a751 // indirect
	github.com/alecthomas/units v0.0.0-20210912230133-d1bdfacee922 // indirect
	github.com/armon/go-metrics v0.3.9 // indirect
	github.com/armon/go-radix v1.0.0 // indirect
	github.com/aws/aws-sdk-go-v2/aws/protocol/eventstream v1.0.0 // indirect
	github.com/aws/aws-sdk-go-v2/feature/ec2/imds v1.8.2 // indirect
	github.com/aws/aws-sdk-go-v2/internal/configsources v1.1.23 // indirect
	github.com/aws/aws-sdk-go-v2/internal/endpoints/v2 v2.4.17 // indirect
	github.com/aws/aws-sdk-go-v2/internal/ini v1.3.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/accept-encoding v1.5.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/presigned-url v1.5.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/internal/s3shared v1.9.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/sso v1.7.0 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cenkalti/backoff/v3 v3.0.0 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/dave/jennifer v1.4.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/emicklei/go-restful/v3 v3.9.0 // indirect
	github.com/evanphx/json-patch/v5 v5.6.0 // indirect
	github.com/fatih/color v1.13.0 // indirect
	github.com/fsnotify/fsnotify v1.6.0 // indirect
	github.com/go-logr/logr v1.2.3 // indirect
	github.com/go-logr/zapr v1.2.3 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.20.0 // indirect
	github.com/go-openapi/swag v0.19.14 // indirect
	github.com/gobuffalo/flect v0.3.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/google/gnostic v0.5.7-v3refs // indirect
	github.com/google/gofuzz v1.1.0 // indirect
	github.com/google/uuid v1.1.2 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-hclog v0.16.2 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-plugin v1.4.3 // indirect
	github.com/hashicorp/go-retryablehttp v0.6.6 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/go-secure-stdlib/mlock v0.1.1 // indirect
	github.com/hashicorp/go-secure-stdlib/parseutil v0.1.1 // indirect
	github.com/hashicorp/go-secure-stdlib/strutil v0.1.1 // indirect
	github.com/hashicorp/go-sockaddr v1.0.2 // indirect
	github.com/hashicorp/go-uuid v1.0.2 // indirect
	github.com/hashicorp/go-version v1.2.0 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/hashicorp/vault/api v1.5.0 // indirect
	github.com/hashicorp/vault/sdk v0.4.1 // indirect
	github.com/hashicorp/yamux v0.0.0-20180604194846-3520598351bb // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/inconshreveable/mousetrap v1.0.1 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/mailru/easyjson v0.7.6 // indirect
	github.com/mattn/go-colorable v0.1.9 // indirect
	github.com/mattn/go-isatty v0.0.16 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/go-testing-interface v1.0.0 // indirect
	github.com/mitchellh/mapstructure v1.4.2 // indirect
	github.com/mitchellh/reflectwalk v1.0.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/oklog/run v1.0.0 // indirect
	github.com/pierrec/lz4 v2.5.2+incompatible // indirect
	github.com/prometheus/client_golang v1.14.0 // indirect
	github.com/prometheus/client_model v0.3.0 // indirect
	github.com/prometheus/common v0.37.0 // indirect
	github.com/prometheus/procfs v0.8.0 // indirect
	github.com/ryanuber/go-glob v1.0.0 // indirect
	github.com/spf13/afero v1.8.0 // indirect
	github.com/spf13/cobra v1.6.1 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
	go.uber.org/atomic v1.9.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/crypto v0.1.0 // indirect
	golang.org/x/mod v0.8.0 // indirect
	golang.org/x/oauth2 v0.0.0-20220223155221-ee480838109b // indirect
	golang.org/x/sys v0.5.0 // indirect
	golang.org/x/term v0.5.0 // indirect
	golang.org/x/text v0.7.0 // indirect
	golang.org/x/time v0.3.0 // indirect
	golang.org/x/tools v0.6.0 // indirect
	gomodules.xyz/jsonpatch/v2 v2.2.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20220502173005-c8bf987b8c21 // indirect
	google.golang.org/grpc v1.49.0 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/square/go-jose.v2 v2.5.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	k8s.io/apiextensions-apiserver v0.26.1 // indirect
	k8s.io/component-base v0.26.1 // indirect
	k8s.io/klog/v2 v2.80.1 // indirect
	k8s.io/kube-openapi v0.0.0-20221012153701-172d655c2280 // indirect
	sigs.k8s.io/json v0.0.0-20220713155537-f223a00ba0e2 // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.3 // indirect
)
