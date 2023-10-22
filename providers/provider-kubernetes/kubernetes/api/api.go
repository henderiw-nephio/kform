package api

//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +kubebuilder:validation:Required
// +kubebuilder:validation:MaxLength=64
type ProviderAPI struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=api,package
	// +kubebuilder:default:=api
	Kind ProviderKind `json:"kind" yaml:"kind"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MaxLength=64
	Directory *string `json:"directory,omitempty" yaml:"directory,omitempty"`
	// The hostname (in form of URI) of Kubernetes master.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MaxLength=64
	Host *string `json:"host,omitempty" yaml:"host,omitempty"`

	// The username to use for HTTP basic authentication when accessing the Kubernetes master endpoint.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MaxLength=64
	Username *string `json:"username,omitempty" yaml:"username,omitempty"`

	// The password to use for HTTP basic authentication when accessing the Kubernetes master endpoint.
	// The hostname (in form of URI) of Kubernetes master.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MaxLength=64
	Password *string `json:"password,omitempty" yaml:"password,omitempty"`

	// Insecure determines whether the server should be accessible without verifying the TLS certificate
	// +kubebuilder:default=false
	Insecure *bool `json:"insecure,omitempty" yaml:"insecure,omitempty"`

	// Server name passed to the server for SNI and is used in the client to check server certificates against
	// example: Some name
	TLSServerName *string `json:"tlsServerName,omitempty" yaml:"tlsServerName,omitempty"`

	// PEM-encoded client certificate for TLS authentication.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MaxLength=64
	ClientCertificate *string `json:"clientCertificate,omitempty" yaml:"clientCertificate,omitempty"`

	// PEM-encoded client certificate key for TLS authentication.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MaxLength=64
	ClientKey *string `json:"clientKey,omitempty" yaml:"clientKey,omitempty"`

	// PEM-encoded root certificates bundle for TLS authentication.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MaxLength=64
	ClusterCACertificate *string `json:"clusterCACertificate,omitempty" yaml:"clusterCACertificate,omitempty"`

	// ConfigPaths defines a list of paths to kube config files.
	ConfigPaths []string `json:"configPaths,omitempty" yaml:"configPaths,omitempty"`

	// ConfigPath defines the path to the kube config file.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MaxLength=64
	// +kubebuilder:default="~/.kube/config"
	ConfigPath *string `json:"configPath,omitempty" yaml:"configPath,omitempty"`

	// ConfigContext defines the context to be used in the kube config file.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MaxLength=64
	ConfigContext *string `json:"configContext,omitempty" yaml:"configContext,omitempty"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MaxLength=64
	ConfigContextAuthInfo *string `json:"configContextAuthInfo,omitempty" yaml:"configContextAuthInfo,omitempty"`

	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MaxLength=64
	ConfigContextCluster *string `json:"configContextCluster,omitempty" yaml:"configContextCluster,omitempty"`

	// Token to authenticate a service account.
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MaxLength=64
	Token *string `json:"token,omitempty" yaml:"token,omitempty"`

	// ProxyURL defines the URL of the proxy to be used for all API requests
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:MaxLength=64
	ProxyURL *string `json:"proxyURL,omitempty" yaml:"proxyURL,omitempty"`

	// Use the  local kubeconfig
	UseConfigFile *bool `json:"useConfigFile,omitempty" yaml:"useConfigFile,omitempty"`

	// Exec executes a command to get the authentication context
	Exec *ExecContext `json:"exec,omitempty" yaml:"exec,omitempty"`
}

type ExecContext struct {
	APIVersion string            `json:"apiVersion" yaml:"apiVersion"`
	Command    string            `json:"command" yaml:"command"`
	Env        map[string]string `json:"env,omitempty" yaml:"env,omitempty"`
	Args       []string          `json:"args,omitempty" yaml:"args,omitempty"`
}

type ProviderKind string

const (
	ProviderKindPackage ProviderKind = "package"
	ProviderKindAPI     ProviderKind = "api"
)

func (r *ProviderAPI) IsKindValid() bool {
	switch r.Kind {
	case ProviderKindPackage:
		return true
	case ProviderKindAPI:
		return true
	default:
		return false
	}
}

var ExpectedProviderKinds = []string{string(ProviderKindPackage), string(ProviderKindAPI)}
