module github.com/appuio/seiso

go 1.16

require (
	github.com/hashicorp/go-version v1.3.0
	github.com/karrick/tparse/v2 v2.8.2
	github.com/knadh/koanf v1.3.2
	github.com/onsi/ginkgo v1.14.1 // indirect
	github.com/openshift/api v0.0.0-20210521075222-e273a339932a
	github.com/openshift/client-go v0.0.0-20210521082421-73d9475a9142
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.2.1
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	github.com/thoas/go-funk v0.9.1
	gopkg.in/src-d/go-git.v4 v4.13.1
	helm.sh/helm/v3 v3.7.2
	k8s.io/api v0.22.4
	k8s.io/apimachinery v0.22.4
	k8s.io/cli-runtime v0.22.4
	k8s.io/client-go v0.22.4
)

replace (
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d
	github.com/docker/docker => github.com/moby/moby v17.12.0-ce-rc1.0.20200618181300-9dc6525e6118+incompatible
)
