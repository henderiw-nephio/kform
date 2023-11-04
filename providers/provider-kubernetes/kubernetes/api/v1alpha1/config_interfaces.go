package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (r *ProviderConfigSpec) IsKindValid() bool {
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

// BuildProviderConfig returns a ProviderConfig from a meta Object and
// an ProviderConfig Spec
func BuildProviderConfig(meta metav1.ObjectMeta, spec ProviderConfigSpec) *ProviderConfig {
	return &ProviderConfig{
		TypeMeta: metav1.TypeMeta{
			APIVersion: APIVersion,
			Kind:       ProviderConfigKind,
		},
		ObjectMeta: meta,
		Spec:       spec,
	}
}
