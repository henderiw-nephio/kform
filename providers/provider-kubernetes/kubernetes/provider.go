package kubernetes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/henderiw-nephio/kform/kform-sdk-go/pkg/diag"
	"github.com/henderiw-nephio/kform/kform-sdk-go/pkg/schema"
	"github.com/henderiw-nephio/kform/providers/provider-kubernetes/kubernetes/api/v1alpha1"
	"github.com/henderiw-nephio/kform/providers/provider-kubernetes/kubernetes/client/k8sclient"
	"github.com/henderiw-nephio/kform/providers/provider-kubernetes/kubernetes/client/pkgclient"
	"github.com/mitchellh/go-homedir"

	//apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"

	apimachineryschema "k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func Provider() *schema.Provider {
	p := &schema.Provider{
		//Schema:         provSchema,
		ResourceMap: map[string]*schema.Resource{
			"kubernetes_manifest": resourceKubernetesManifest(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"kubernetes_manifest": dataSourceKubernetesManifest(),
		},
		ListDataSourcesMap: map[string]*schema.Resource{
			"kubernetes_manifest": dataSourcesKubernetesManifest(),
		},
	}
	p.ConfigureContextFunc = func(ctx context.Context, d []byte) (any, diag.Diagnostics) {
		return providerConfigure(ctx, d, p.Version)
	}
	return p
}

/*
func (k kubeClientsets) MainClientset() (*kubernetes.Clientset, error) {
	if k.mainClientset != nil {
		return k.mainClientset, nil
	}

	if k.config != nil {
		kc, err := kubernetes.NewForConfig(k.config)
		if err != nil {
			return nil, fmt.Errorf("Failed to configure client: %s", err)
		}
		k.mainClientset = kc
	}
	return k.mainClientset, nil
}
*/

func providerConfigure(ctx context.Context, d []byte, version string) (any, diag.Diagnostics) {
	/*
		b, err := yaml.Marshal(d)
		if err != nil {
			return nil, diag.FromErr(err)
		}
	*/
	providerConfig := &v1alpha1.ProviderConfig{}
	if err := json.Unmarshal(d, providerConfig); err != nil {
		return nil, diag.FromErr(err)
	}

	if !providerConfig.Spec.IsKindValid() {
		return nil, diag.Errorf("invalid provider kind, got: %s, expected: %v", providerConfig.Kind, v1alpha1.ExpectedProviderKinds)
	}

	if providerConfig.Spec.Kind == v1alpha1.ProviderKindPackage {
		dir := "./out"
		if providerConfig.Spec.Directory != nil {
			dir = *providerConfig.Spec.Directory
		}

		c, err := pkgclient.New(pkgclient.Config{
			Dir:               dir,
			IgnoreAnnotations: []string{},
			IgnoreLabels:      []string{},
		})
		if err != nil {
			return nil, diag.FromErr(err)
		}
		return c, diag.Diagnostics{}
	}

	cfg, err := initializeConfiguration(ctx, providerConfig)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	if cfg == nil {
		// IMPORTANT: if the supplied configuration is incomplete or invalid
		///IMPORTANT: provider operations will fail or attempt to connect to localhost endpoints
		cfg = &rest.Config{}
	}
	cfg.UserAgent = fmt.Sprintf("K8sForm/%s", version)

	c, err := k8sclient.New(k8sclient.Config{
		RESTCOnfig:        cfg,
		IgnoreAnnotations: []string{},
		IgnoreLabels:      []string{},
	})
	if err != nil {
		return nil, diag.FromErr(err)
	}

	return c, diag.Diagnostics{}
}

func initializeConfiguration(ctx context.Context, providerConfig *v1alpha1.ProviderConfig) (*rest.Config, error) {
	overrides := &clientcmd.ConfigOverrides{}
	loader := &clientcmd.ClientConfigLoadingRules{}

	configPaths := []string{}
	if providerConfig.Spec.ConfigPath != nil {
		configPaths = []string{*providerConfig.Spec.ConfigPath}
	} else if len(providerConfig.Spec.ConfigPaths) > 0 {
		configPaths = append(configPaths, providerConfig.Spec.ConfigPaths...)
	} else if v := os.Getenv("KUBE_CONFIG_PATHS"); v != "" {
		configPaths = filepath.SplitList(v)
	}

	if len(configPaths) > 0 && providerConfig.Spec.UseConfigFile != nil && *providerConfig.Spec.UseConfigFile {
		expandedPaths := []string{}
		for _, p := range configPaths {
			path, err := homedir.Expand(p)
			if err != nil {
				return nil, err
			}
			slog.Debug("using kubeconfig", "file", path)
			expandedPaths = append(expandedPaths, path)
		}

		if len(expandedPaths) == 1 {
			loader.ExplicitPath = expandedPaths[0]
		} else {
			loader.Precedence = expandedPaths
		}
		ctxSuffix := "; default context"

		if providerConfig.Spec.ConfigContext != nil ||
			providerConfig.Spec.ConfigContextAuthInfo != nil ||
			providerConfig.Spec.ConfigContextCluster != nil {
			ctxSuffix = "; overridden context"
			if providerConfig.Spec.ConfigContext != nil {
				overrides.CurrentContext = *providerConfig.Spec.ConfigContext
				ctxSuffix += fmt.Sprintf("; config ctx: %s", overrides.CurrentContext)
				slog.Debug("using custom current context", "context", overrides.CurrentContext)
			}
			overrides.Context = clientcmdapi.Context{}
			if providerConfig.Spec.ConfigContextAuthInfo != nil {
				overrides.Context.AuthInfo = *providerConfig.Spec.ConfigContextAuthInfo
				ctxSuffix += fmt.Sprintf("; auth_info: %s", overrides.Context.AuthInfo)
			}
			if providerConfig.Spec.ConfigContextCluster != nil {
				overrides.Context.Cluster = *providerConfig.Spec.ConfigContextCluster
				ctxSuffix += fmt.Sprintf("; cluster: %s", overrides.Context.Cluster)
			}
			slog.Debug("using overridden context", "context", overrides.Context)
		}
	}

	// Overriding with static configuration
	if providerConfig.Spec.Insecure != nil {
		overrides.ClusterInfo.InsecureSkipTLSVerify = *providerConfig.Spec.Insecure
	}
	if providerConfig.Spec.TLSServerName != nil {
		overrides.ClusterInfo.TLSServerName = *providerConfig.Spec.TLSServerName
	}
	if providerConfig.Spec.ClusterCACertificate != nil {
		overrides.ClusterInfo.CertificateAuthorityData = bytes.NewBufferString(*providerConfig.Spec.ClusterCACertificate).Bytes()
	}
	if providerConfig.Spec.ClientCertificate != nil {
		overrides.AuthInfo.ClientCertificateData = bytes.NewBufferString(*providerConfig.Spec.ClientCertificate).Bytes()
	}
	if providerConfig.Spec.Host != nil {
		// Server has to be the complete address of the kubernetes cluster (scheme://hostname:port), not just the hostname,
		// because `overrides` are processed too late to be taken into account by `defaultServerUrlFor()`.
		// This basically replicates what defaultServerUrlFor() does with config but for overrides,
		// see https://github.com/kubernetes/client-go/blob/v12.0.0/rest/url_utils.go#L85-L87
		hasCA := len(overrides.ClusterInfo.CertificateAuthorityData) != 0
		hasCert := len(overrides.AuthInfo.ClientCertificateData) != 0
		defaultTLS := hasCA || hasCert || overrides.ClusterInfo.InsecureSkipTLSVerify
		host, _, err := rest.DefaultServerURL(*providerConfig.Spec.Host, "", apimachineryschema.GroupVersion{}, defaultTLS)
		if err != nil {
			return nil, fmt.Errorf("failed to parse host: %s", err)
		}

		overrides.ClusterInfo.Server = host.String()
	}
	if providerConfig.Spec.Username != nil {
		overrides.AuthInfo.Username = *providerConfig.Spec.Username
	}
	if providerConfig.Spec.Password != nil {
		overrides.AuthInfo.Password = *providerConfig.Spec.Password
	}
	if providerConfig.Spec.ClientKey != nil {
		overrides.AuthInfo.ClientKeyData = bytes.NewBufferString(*providerConfig.Spec.ClientKey).Bytes()
	}
	if providerConfig.Spec.Token != nil {
		overrides.AuthInfo.Token = *providerConfig.Spec.Token
	}

	/*
		if providerConfig.Spec.Exec != nil {
			exec := &clientcmdapi.ExecConfig{
				APIVersion: providerConfig.Spec.Exec.APIVersion,
				Command:    providerConfig.Spec.Exec.Command,
				Args:       providerConfig.Spec.Exec.Args,
			}
			for k, v := range providerConfig.Spec.Exec.Env {
				exec.Env = append(exec.Env, clientcmdapi.ExecEnvVar{Name: k, Value: v})
			}
			overrides.AuthInfo.Exec = exec
		}
	*/

	if providerConfig.Spec.ProxyURL != nil {
		overrides.ClusterDefaults.ProxyURL = *providerConfig.Spec.ProxyURL
	}

	cc := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loader, overrides)
	cfg, err := cc.ClientConfig()
	if err != nil {
		slog.Warn("Invalid provider configuration was supplied. Provider operations likely to fail", "error", err)
		return nil, nil
	}

	return cfg, nil

}
