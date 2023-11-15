package v1alpha1

import (
	"github.com/apparentlymart/go-versions/versions"
	"github.com/henderiw-nephio/kform/tools/pkg/syntax/address"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// BuildVXLANClaim returns a VXLANClaim from a client Object a crName and
// an VXLANClaim Spec/Status
func BuildKptFile(meta metav1.ObjectMeta, spec KformFileSpec) *KformFile {
	return &KformFile{
		TypeMeta: metav1.TypeMeta{
			APIVersion: APIVersion,
			Kind:       KformFileKind,
		},
		ObjectMeta: meta,
		Spec:       spec,
	}
}

func (r Provider) Validate() error {
	if _, _, err := address.ParseSource(r.Source); err != nil {
		return err
	}
	if _, err := versions.MeetingConstraintsStringRuby(r.Version); err != nil {
		return err
	}
	return nil
}
