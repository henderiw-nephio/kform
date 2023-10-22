package schema

import (
	"strings"

	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/kube-openapi/pkg/validation/spec"
	"k8s.io/kube-openapi/pkg/validation/strfmt"
	"k8s.io/kube-openapi/pkg/validation/validate"
)

type SchemaValidator interface {
	SchemaCreateValidator
	ValidateUpdate(new, old interface{}) *validate.Result
}

type SchemaCreateValidator interface {
	Validate(value interface{}) *validate.Result
}

// basicSchemaValidator wraps a kube-openapi SchemaCreateValidator to
// support ValidateUpdate. It implements ValidateUpdate by simply validating
// the new value via kube-openapi, ignoring the old value.
type basicSchemaValidator struct {
	*validate.SchemaValidator
}

func (s basicSchemaValidator) ValidateUpdate(new, old interface{}) *validate.Result {
	return s.Validate(new)
}

// NewSchemaValidator creates an openapi schema validator for the given CRD validation.
//
// If feature `CRDValidationRatcheting` is disabled, this returns validator which
// validates all `Update`s and `Create`s as a `Create` - without considering old value.
//
// If feature `CRDValidationRatcheting` is enabled - the validator returned
// will support ratcheting unchanged correlatable fields across an update.
func NewSchemaValidator(customResourceValidation *apiext.JSONSchemaProps) (SchemaValidator, *spec.Schema, error) {
	// Convert CRD schema to openapi schema
	openapiSchema := &spec.Schema{}
	if customResourceValidation != nil {
		// TODO: replace with NewStructural(...).ToGoOpenAPI
		if err := ConvertJSONSchemaPropsWithPostProcess(customResourceValidation, openapiSchema, StripUnsupportedFormatsPostProcess); err != nil {
			return nil, nil, err
		}
	}

	return basicSchemaValidator{validate.NewSchemaValidator(openapiSchema, nil, "", strfmt.Default)}, openapiSchema, nil
}

// PostProcessFunc post-processes one node of a spec.Schema.
type PostProcessFunc func(*spec.Schema) error

// ConvertJSONSchemaPropsWithPostProcess converts the schema from apiextensions.JSONSchemaPropos to go-openapi/spec.Schema
// and run a post process step on each JSONSchemaProps node. postProcess is never called for nil schemas.
func ConvertJSONSchemaPropsWithPostProcess(in *apiext.JSONSchemaProps, out *spec.Schema, postProcess PostProcessFunc) error {
	if in == nil {
		return nil
	}

	out.ID = in.ID
	out.Schema = spec.SchemaURL(in.Schema)
	out.Description = in.Description
	if in.Type != "" {
		out.Type = spec.StringOrArray([]string{in.Type})
	}
	if in.XIntOrString {
		out.VendorExtensible.AddExtension("x-kubernetes-int-or-string", true)
		out.Type = spec.StringOrArray{"integer", "string"}
	}
	out.Nullable = in.Nullable
	out.Format = in.Format
	out.Title = in.Title
	out.Maximum = in.Maximum
	out.ExclusiveMaximum = in.ExclusiveMaximum
	out.Minimum = in.Minimum
	out.ExclusiveMinimum = in.ExclusiveMinimum
	out.MaxLength = in.MaxLength
	out.MinLength = in.MinLength
	out.Pattern = in.Pattern
	out.MaxItems = in.MaxItems
	out.MinItems = in.MinItems
	out.UniqueItems = in.UniqueItems
	out.MultipleOf = in.MultipleOf
	out.MaxProperties = in.MaxProperties
	out.MinProperties = in.MinProperties
	out.Required = in.Required

	if in.Default != nil {
		out.Default = *(in.Default)
	}
	if in.Example != nil {
		out.Example = *(in.Example)
	}

	if in.Enum != nil {
		out.Enum = make([]interface{}, len(in.Enum))
		for k, v := range in.Enum {
			out.Enum[k] = v
		}
	}

	if err := convertSliceOfJSONSchemaProps(&in.AllOf, &out.AllOf, postProcess); err != nil {
		return err
	}
	if err := convertSliceOfJSONSchemaProps(&in.OneOf, &out.OneOf, postProcess); err != nil {
		return err
	}
	if err := convertSliceOfJSONSchemaProps(&in.AnyOf, &out.AnyOf, postProcess); err != nil {
		return err
	}

	if in.Not != nil {
		in, out := &in.Not, &out.Not
		*out = new(spec.Schema)
		if err := ConvertJSONSchemaPropsWithPostProcess(*in, *out, postProcess); err != nil {
			return err
		}
	}

	var err error
	out.Properties, err = convertMapOfJSONSchemaProps(in.Properties, postProcess)
	if err != nil {
		return err
	}

	out.PatternProperties, err = convertMapOfJSONSchemaProps(in.PatternProperties, postProcess)
	if err != nil {
		return err
	}

	out.Definitions, err = convertMapOfJSONSchemaProps(in.Definitions, postProcess)
	if err != nil {
		return err
	}

	if in.Ref != nil {
		out.Ref, err = spec.NewRef(*in.Ref)
		if err != nil {
			return err
		}
	}

	if in.AdditionalProperties != nil {
		in, out := &in.AdditionalProperties, &out.AdditionalProperties
		*out = new(spec.SchemaOrBool)
		if err := convertJSONSchemaPropsorBool(*in, *out, postProcess); err != nil {
			return err
		}
	}

	if in.AdditionalItems != nil {
		in, out := &in.AdditionalItems, &out.AdditionalItems
		*out = new(spec.SchemaOrBool)
		if err := convertJSONSchemaPropsorBool(*in, *out, postProcess); err != nil {
			return err
		}
	}

	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = new(spec.SchemaOrArray)
		if err := convertJSONSchemaPropsOrArray(*in, *out, postProcess); err != nil {
			return err
		}
	}

	if in.Dependencies != nil {
		in, out := &in.Dependencies, &out.Dependencies
		*out = make(spec.Dependencies, len(*in))
		for key, val := range *in {
			newVal := new(spec.SchemaOrStringArray)
			if err := convertJSONSchemaPropsOrStringArray(&val, newVal, postProcess); err != nil {
				return err
			}
			(*out)[key] = *newVal
		}
	}

	if in.ExternalDocs != nil {
		out.ExternalDocs = &spec.ExternalDocumentation{}
		out.ExternalDocs.Description = in.ExternalDocs.Description
		out.ExternalDocs.URL = in.ExternalDocs.URL
	}

	if postProcess != nil {
		if err := postProcess(out); err != nil {
			return err
		}
	}

	if in.XPreserveUnknownFields != nil {
		out.VendorExtensible.AddExtension("x-kubernetes-preserve-unknown-fields", *in.XPreserveUnknownFields)
	}
	if in.XEmbeddedResource {
		out.VendorExtensible.AddExtension("x-kubernetes-embedded-resource", true)
	}
	if len(in.XListMapKeys) != 0 {
		out.VendorExtensible.AddExtension("x-kubernetes-list-map-keys", convertSliceToInterfaceSlice(in.XListMapKeys))
	}
	if in.XListType != nil {
		out.VendorExtensible.AddExtension("x-kubernetes-list-type", *in.XListType)
	}
	if in.XMapType != nil {
		out.VendorExtensible.AddExtension("x-kubernetes-map-type", *in.XMapType)
	}
	/*
		if len(in.XValidations) != 0 {
			var serializationValidationRules apiext.ValidationRules
			if err := apiext.Convert_apiextensions_ValidationRules_To_v1_ValidationRules(&in.XValidations, &serializationValidationRules, nil); err != nil {
				return err
			}
			out.VendorExtensible.AddExtension("x-kubernetes-validations", convertSliceToInterfaceSlice(serializationValidationRules))
		}
	*/
	return nil
}
func convertSliceToInterfaceSlice[T any](in []T) []interface{} {
	var res []interface{}
	for _, v := range in {
		res = append(res, v)
	}
	return res
}

func convertSliceOfJSONSchemaProps(in *[]apiext.JSONSchemaProps, out *[]spec.Schema, postProcess PostProcessFunc) error {
	if in != nil {
		for _, jsonSchemaProps := range *in {
			schema := spec.Schema{}
			if err := ConvertJSONSchemaPropsWithPostProcess(&jsonSchemaProps, &schema, postProcess); err != nil {
				return err
			}
			*out = append(*out, schema)
		}
	}
	return nil
}

func convertMapOfJSONSchemaProps(in map[string]apiext.JSONSchemaProps, postProcess PostProcessFunc) (map[string]spec.Schema, error) {
	if in == nil {
		return nil, nil
	}

	out := make(map[string]spec.Schema)
	for k, jsonSchemaProps := range in {
		schema := spec.Schema{}
		if err := ConvertJSONSchemaPropsWithPostProcess(&jsonSchemaProps, &schema, postProcess); err != nil {
			return nil, err
		}
		out[k] = schema
	}
	return out, nil
}

func convertJSONSchemaPropsOrArray(in *apiext.JSONSchemaPropsOrArray, out *spec.SchemaOrArray, postProcess PostProcessFunc) error {
	if in.Schema != nil {
		in, out := &in.Schema, &out.Schema
		*out = new(spec.Schema)
		if err := ConvertJSONSchemaPropsWithPostProcess(*in, *out, postProcess); err != nil {
			return err
		}
	}
	if in.JSONSchemas != nil {
		in, out := &in.JSONSchemas, &out.Schemas
		*out = make([]spec.Schema, len(*in))
		for i := range *in {
			if err := ConvertJSONSchemaPropsWithPostProcess(&(*in)[i], &(*out)[i], postProcess); err != nil {
				return err
			}
		}
	}
	return nil
}

func convertJSONSchemaPropsorBool(in *apiext.JSONSchemaPropsOrBool, out *spec.SchemaOrBool, postProcess PostProcessFunc) error {
	out.Allows = in.Allows
	if in.Schema != nil {
		in, out := &in.Schema, &out.Schema
		*out = new(spec.Schema)
		if err := ConvertJSONSchemaPropsWithPostProcess(*in, *out, postProcess); err != nil {
			return err
		}
	}
	return nil
}

func convertJSONSchemaPropsOrStringArray(in *apiext.JSONSchemaPropsOrStringArray, out *spec.SchemaOrStringArray, postProcess PostProcessFunc) error {
	out.Property = in.Property
	if in.Schema != nil {
		in, out := &in.Schema, &out.Schema
		*out = new(spec.Schema)
		if err := ConvertJSONSchemaPropsWithPostProcess(*in, *out, postProcess); err != nil {
			return err
		}
	}
	return nil
}

var supportedFormats = sets.NewString(
	"bsonobjectid", // bson object ID
	"uri",          // an URI as parsed by Golang net/url.ParseRequestURI
	"email",        // an email address as parsed by Golang net/mail.ParseAddress
	"hostname",     // a valid representation for an Internet host name, as defined by RFC 1034, section 3.1 [RFC1034].
	"ipv4",         // an IPv4 IP as parsed by Golang net.ParseIP
	"ipv6",         // an IPv6 IP as parsed by Golang net.ParseIP
	"cidr",         // a CIDR as parsed by Golang net.ParseCIDR
	"mac",          // a MAC address as parsed by Golang net.ParseMAC
	"uuid",         // an UUID that allows uppercase defined by the regex (?i)^[0-9a-f]{8}-?[0-9a-f]{4}-?[0-9a-f]{4}-?[0-9a-f]{4}-?[0-9a-f]{12}$
	"uuid3",        // an UUID3 that allows uppercase defined by the regex (?i)^[0-9a-f]{8}-?[0-9a-f]{4}-?3[0-9a-f]{3}-?[0-9a-f]{4}-?[0-9a-f]{12}$
	"uuid4",        // an UUID4 that allows uppercase defined by the regex (?i)^[0-9a-f]{8}-?[0-9a-f]{4}-?4[0-9a-f]{3}-?[89ab][0-9a-f]{3}-?[0-9a-f]{12}$
	"uuid5",        // an UUID6 that allows uppercase defined by the regex (?i)^[0-9a-f]{8}-?[0-9a-f]{4}-?5[0-9a-f]{3}-?[89ab][0-9a-f]{3}-?[0-9a-f]{12}$
	"isbn",         // an ISBN10 or ISBN13 number string like "0321751043" or "978-0321751041"
	"isbn10",       // an ISBN10 number string like "0321751043"
	"isbn13",       // an ISBN13 number string like "978-0321751041"
	"creditcard",   // a credit card number defined by the regex ^(?:4[0-9]{12}(?:[0-9]{3})?|5[1-5][0-9]{14}|6(?:011|5[0-9][0-9])[0-9]{12}|3[47][0-9]{13}|3(?:0[0-5]|[68][0-9])[0-9]{11}|(?:2131|1800|35\\d{3})\\d{11})$ with any non digit characters mixed in
	"ssn",          // a U.S. social security number following the regex ^\\d{3}[- ]?\\d{2}[- ]?\\d{4}$
	"hexcolor",     // an hexadecimal color code like "#FFFFFF", following the regex ^#?([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$
	"rgbcolor",     // an RGB color code like rgb like "rgb(255,255,2559"
	"byte",         // base64 encoded binary data
	"password",     // any kind of string
	"date",         // a date string like "2006-01-02" as defined by full-date in RFC3339
	"duration",     // a duration string like "22 ns" as parsed by Golang time.ParseDuration or compatible with Scala duration format
	"datetime",     // a date time string like "2014-12-15T19:30:20.000Z" as defined by date-time in RFC3339
)

// StripUnsupportedFormatsPostProcess sets unsupported formats to empty string.
func StripUnsupportedFormatsPostProcess(s *spec.Schema) error {
	if len(s.Format) == 0 {
		return nil
	}

	normalized := strings.Replace(s.Format, "-", "", -1) // go-openapi default format name normalization
	if !supportedFormats.Has(normalized) {
		s.Format = ""
	}

	return nil
}
