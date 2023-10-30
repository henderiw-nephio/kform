# providers with aliases

Provider configurations are global resource. When using with modules special considerations should be taken into account, since  a resource can only be associated with 1 provider

provider blockKind can only come from the root module

## required providers

Specifies a version constraint to a provider (uses only root)
uses only raw provider syntax references


```yaml
apiVersion: meta.pkg.kform.io/v1alpha1
kind: KformFile
metadata:
  name: interface
spec:
  requiredProviders:
    kubernetes:
      source: .terraform/providers/kubernetes
```

## provider

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: providers
data:
  spec:
  - provider:
      kubernetes: 
        config: {}
  - provider:
      kubernetes_cluster01: ## provider alias syntax <provider>_<alias>
        config:
          kubeConfig: cluster01
```

## resource

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: main
  annotations:
    kform/config: true
data:
  spec:
  - resource:
      kubernetes_ipclaim:
        ipv4:
          attributes:
            provider: kubernetes.cluster01
            schema:
              apiVersion: ipam.resource.nephio.org/v1alpha1
              kind: IPClaim
          instances: {}
```