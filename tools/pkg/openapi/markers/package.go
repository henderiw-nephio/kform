package markers

import "github.com/henderiw-nephio/kform/tools/pkg/markers"

func init() {
	AllDefinitions = append(AllDefinitions,
		must(markers.MakeDefinition("groupName", markers.DescribesPackage, "")).
			WithHelp(markers.SimpleHelp("CRD", "specifies the API group name for this package.")),

		must(markers.MakeDefinition("versionName", markers.DescribesPackage, "")).
			WithHelp(markers.SimpleHelp("CRD", "overrides the API group version for this package (defaults to the package name).")),

		must(markers.MakeDefinition("kubebuilder:validation:Optional", markers.DescribesPackage, struct{}{})).
			WithHelp(markers.SimpleHelp("CRD validation", "specifies that all fields in this package are optional by default.")),

		must(markers.MakeDefinition("kubebuilder:validation:Required", markers.DescribesPackage, struct{}{})).
			WithHelp(markers.SimpleHelp("CRD validation", "specifies that all fields in this package are required by default.")),

		must(markers.MakeDefinition("kubebuilder:skip", markers.DescribesPackage, struct{}{})).
			WithHelp(markers.SimpleHelp("CRD", "don't consider this package as an API version.")),
	)
}
