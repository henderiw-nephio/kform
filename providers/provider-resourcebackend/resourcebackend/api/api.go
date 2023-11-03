package api

//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// +kubebuilder:validation:Required
// +kubebuilder:validation:MaxLength=64
type ProviderAPI struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=api,mock
	// +kubebuilder:default:=api
	Kind ProviderKind `json:"kind" yaml:"kind"`

	Address string `json:"address,omitempty" yaml:"address,omitempty"`
}

type ProviderKind string

const (
	ProviderKindMock ProviderKind = "mock"
	ProviderKindAPI  ProviderKind = "api"
)

func (r *ProviderAPI) IsKindValid() bool {
	switch r.Kind {
	case ProviderKindMock:
		return true
	case ProviderKindAPI:
		return true
	default:
		return false
	}
}

var ExpectedProviderKinds = []string{string(ProviderKindMock), string(ProviderKindAPI)}
