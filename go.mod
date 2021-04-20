module github.com/appuio/seiso

go 1.16

require (
	github.com/etdub/goparsetime v0.0.0-20160315173935-ea17b0ac3318 // indirect
	github.com/hashicorp/go-version v1.3.0
	github.com/karrick/tparse v2.4.2+incompatible
	github.com/knadh/koanf v0.16.0
	github.com/openshift/api v0.0.0-20210202165416-a9e731090f5e
	github.com/openshift/client-go v0.0.0-20210112165513-ebc401615f47
	github.com/sirupsen/logrus v1.8.1
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	github.com/thoas/go-funk v0.8.0
	gopkg.in/src-d/go-git.v4 v4.13.1
	helm.sh/helm/v3 v3.5.4
	k8s.io/api v0.20.4
	k8s.io/apimachinery v0.20.4
	k8s.io/cli-runtime v0.20.4
	k8s.io/client-go v0.20.4
)

replace (
	github.com/docker/distribution => github.com/docker/distribution v0.0.0-20191216044856-a8371794149d
	github.com/docker/docker => github.com/moby/moby v17.12.0-ce-rc1.0.20200618181300-9dc6525e6118+incompatible
)
