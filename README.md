#

## structure

plugin:
- specific to go (needed for both the core and the provider)
- defines the structures for the kform plugin (grpc proto)
sdk:
- specific to go (needed for the provider)
- defines the provider structures


## registry commands

### build

crossplane build provider -> uses the name in the yaml file

crossplane.yaml

```yaml
apiVersion: meta.pkg.crossplane.io/v1
kind: Provider
metadata:
  name: provider-gcp
spec:
  crossplane:
    version: ">=v1.0.0"
  controller:
    image: crossplane/provider-gcp-controller:v0.14.0
```

### push

crossplane push provider xpkg.upbound.io/crossplane-contrib/provider-gcp:v0.22.0
crossplane push configuration xpkg.upbound.io/crossplane-contrib/my-org-infra:v0.1.0

### install

crossplane install provider xpkg.upbound.io/crossplane-contrib/provider-gcp:v0.22.0
crossplane install configuration xpkg.upbound.io/crossplane-contrib/my-org-infra:v0.1.0

```yaml
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: provider-gcp
spec:
  package: xpkg.upbound.io/crossplane-contrib/provider-gcp:v0.22.0
  packagePullPolicy: IfNotPresent
  revisionActivationPolicy: Automatic
  revisionHistoryLimit: 1
```

```yaml
apiVersion: pkg.crossplane.io/v1
kind: Configuration
metadata:
  name: my-org-infra
spec:
  package: xpkg.upbound.io/crossplane-contrib/my-org-infra:v0.1.0
  packagePullPolicy: IfNotPresent
  revisionActivationPolicy: Automatic
  revisionHistoryLimit: 1
```


brainstorm options:
- add a KRM file in the package (meta.pkg.xxx.io/v1)
    provides:
        - type of the package
        - name of the package
- generate the KRM file in the package (meta.pkg.xxx.io/v1) based on cli input
    


meta 
```yaml
apiVersion: meta.pkg.xxxx.io/v1
kind: Module or Provider
metadata:
  name: provider-gcp
spec:
  info:
    description:
    maintainers:
    - name: xx
      email: xxx
  refs:
  - apiVersion: xxx
    kind: xxx
    name: xxxx
```

kform package init PACKAGE-TYPE DIR [flags]
- type
- description
- icon
- maintaianers
- version

kform package build PACKAGE-TYPE DIR [flags]
kform package build module ./module-aws-vpc
kform package build provider ./provider-k8s

kform package push [PACKAGE-NAME]


module
- .kformignore
- kform.yaml
- README.md
- ... <configmap>

provider
- .kformignore
- kform.yaml
- README.md
- image
    - <images>
- schemas
    - provider
        - provider-schema.yaml
    - resources
        - <crd>.yaml
        - <core>.json ??????


# url for now

https://github.com/henderiw-nephio/kform/releases/download/v0.0.1/provider-kubernetes_0.0.1_darwin_amd64

europe-docker.pkg.dev/srlinux/eu.gcr.io/provider-xxxx

github.com/henderiw-nephio/kform/provider-xxxx

source: 
"hashicorp/aws"


manifest
- config
    digest
- []layers
    - digest




```shell
go run tools/cmd/kform/main.go pkg push ghcr.io/kform-providers/resourcebackend/resourcebackend:v0.0.1 ./build/provider-resourcebackend --releaser

go run tools/cmd/kform/main.go pkg pull ghcr.io/kform-providers/resourcebackend/resourcebackend:0.0.1 ./build/pull-test --kind provider   


go run tools/cmd/kform/main.go init examples
```