package v1alpha1

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

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
