package v1alpha1

import (
	"reflect"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ProviderKind string

const (
	ProviderKindMock ProviderKind = "mock"
	ProviderKindAPI  ProviderKind = "api"
)

type ProviderConfigSpec struct {
	// +kubebuilder:validation:Required
	// +kubebuilder:validation:Enum=api,mock
	// +kubebuilder:default:=api
	Kind ProviderKind `json:"kind" yaml:"kind"`

	Address string `json:"address,omitempty" yaml:"address,omitempty"`
}

type ProviderConfig struct {
	metav1.TypeMeta   `json:",inline" yaml:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty" yaml:"metadata,omitempty"`

	Spec ProviderConfigSpec `json:"spec,omitempty" yaml:"spec,omitempty"`
}

var (
	ProviderConfigKind = reflect.TypeOf(ProviderConfig{}).Name()
)
